package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/routes"
	"gogator/internal/sites"

	_ "modernc.org/sqlite"
)

const DefaultPath = "gogator.sqlite"

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if strings.TrimSpace(path) == "" {
		path = DefaultPath
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Init() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS schema_meta (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS raw_points (
			id INTEGER PRIMARY KEY,
			row_hash TEXT NOT NULL UNIQUE,
			source_file TEXT NOT NULL,
			raw_row INTEGER NOT NULL,
			raw_dt TEXT NOT NULL,
			normalised_time TEXT NOT NULL,
			lat REAL NOT NULL,
			lng REAL NOT NULL,
			altitude REAL,
			angle REAL,
			speed_kph REAL,
			params_raw TEXT,
			params_json TEXT,
			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_raw_points_time ON raw_points(normalised_time)`,
		`CREATE TABLE IF NOT EXISTS sites (
			id INTEGER PRIMARY KEY,
			site TEXT NOT NULL UNIQUE,
			real_address TEXT,
			lat REAL NOT NULL,
			lng REAL NOT NULL,
			range_m REAL,
			min_destination_minutes REAL,
			type TEXT,
			important INTEGER NOT NULL DEFAULT 1,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS routes (
			id INTEGER PRIMARY KEY,
			route_name TEXT NOT NULL,
			from_site TEXT NOT NULL,
			to_site TEXT NOT NULL,
			expected_distance_min_km REAL,
			expected_distance_max_km REAL,
			expected_duration_min_min REAL,
			expected_duration_max_min REAL,
			confidence_boost TEXT,
			auto_merge_gap_min REAL,
			notes TEXT,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(from_site, to_site)
		)`,
		`CREATE TABLE IF NOT EXISTS processing_runs (
			id INTEGER PRIMARY KEY,
			started_at TEXT NOT NULL,
			completed_at TEXT,
			config_json TEXT,
			timezone TEXT,
			raw_time_correction_hours REAL,
			input_files TEXT,
			notes TEXT
		)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.Exec(stmt); err != nil {
			return err
		}
	}
	_, err := s.db.Exec(`INSERT INTO schema_meta(key, value) VALUES('version', '1') ON CONFLICT(key) DO UPDATE SET value=excluded.value`)
	return err
}

func (s *Store) ImportRaw(path string, loc *time.Location, cfg config.Config) (inserted, skipped int, err error) {
	points, err := gps.ReadRawCSV(path, loc, cfg)
	if err != nil {
		return 0, 0, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO raw_points(row_hash, source_file, raw_row, raw_dt, normalised_time, lat, lng, altitude, angle, speed_kph, params_raw, params_json) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return 0, 0, err
	}
	defer stmt.Close()
	for _, p := range points {
		paramsJSON, _ := json.Marshal(p.Params)
		h := rawPointHash(p)
		res, execErr := stmt.Exec(h, p.SourceFile, p.RawRow, p.RawDT, p.Time.Format(time.RFC3339Nano), p.Lat, p.Lng, p.Altitude, p.Angle, p.SpeedKPH, p.ParamsRaw, string(paramsJSON))
		if execErr != nil {
			return inserted, skipped, execErr
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			skipped++
		} else {
			inserted++
		}
	}
	if err = tx.Commit(); err != nil {
		return inserted, skipped, err
	}
	return inserted, skipped, nil
}

func (s *Store) ImportSites(path string, cfg config.Config) (int, error) {
	list, err := sites.Load(path, cfg)
	if err != nil {
		return 0, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(`INSERT INTO sites(site, real_address, lat, lng, range_m, min_destination_minutes, type, important, updated_at) VALUES(?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP) ON CONFLICT(site) DO UPDATE SET real_address=excluded.real_address, lat=excluded.lat, lng=excluded.lng, range_m=excluded.range_m, min_destination_minutes=excluded.min_destination_minutes, type=excluded.type, important=excluded.important, updated_at=CURRENT_TIMESTAMP`)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()
	for _, site := range list {
		important := 0
		if site.Important {
			important = 1
		}
		if _, err := stmt.Exec(site.Name, site.Address, site.Lat, site.Lng, site.RadiusM, site.MinDestinationMinutes, site.SiteType, important); err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(list), nil
}

func (s *Store) ExportSites(path string) error {
	rows, err := s.db.Query(`SELECT site, real_address, lat, lng, range_m, min_destination_minutes, type, important FROM sites ORDER BY site`)
	if err != nil {
		return err
	}
	defer rows.Close()
	return writeDelimited(path, []string{"Site", "Real Address", "GPS", "Range", "Min Destination Minutes", "Type", "Important"}, func(w *csv.Writer) error {
		for rows.Next() {
			var site, addr, typ string
			var lat, lng, radius, minDest float64
			var important int
			if err := rows.Scan(&site, &addr, &lat, &lng, &radius, &minDest, &typ, &important); err != nil {
				return err
			}
			imp := "no"
			if important != 0 {
				imp = "yes"
			}
			if err := w.Write([]string{site, addr, fmt.Sprintf("%.8f,%.8f", lat, lng), fmtFloat(radius), fmtFloat(minDest), typ, imp}); err != nil {
				return err
			}
		}
		return rows.Err()
	})
}

func (s *Store) ImportRoutes(path string) (int, error) {
	list, err := routes.Load(path)
	if err != nil {
		return 0, err
	}
	tx, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	stmt, err := tx.Prepare(`INSERT INTO routes(route_name, from_site, to_site, expected_distance_min_km, expected_distance_max_km, expected_duration_min_min, expected_duration_max_min, confidence_boost, auto_merge_gap_min, notes, updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP) ON CONFLICT(from_site, to_site) DO UPDATE SET route_name=excluded.route_name, expected_distance_min_km=excluded.expected_distance_min_km, expected_distance_max_km=excluded.expected_distance_max_km, expected_duration_min_min=excluded.expected_duration_min_min, expected_duration_max_min=excluded.expected_duration_max_min, confidence_boost=excluded.confidence_boost, auto_merge_gap_min=excluded.auto_merge_gap_min, notes=excluded.notes, updated_at=CURRENT_TIMESTAMP`)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	defer stmt.Close()
	for _, route := range list {
		if _, err := stmt.Exec(route.Name, route.FromSite, route.ToSite, route.DistanceMinKM, route.DistanceMaxKM, route.DurationMinMin, route.DurationMaxMin, route.ConfidenceBoost, route.AutoMergeGapMin, route.Notes); err != nil {
			tx.Rollback()
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(list), nil
}

func (s *Store) ExportRoutes(path string) error {
	rows, err := s.db.Query(`SELECT route_name, from_site, to_site, expected_distance_min_km, expected_distance_max_km, expected_duration_min_min, expected_duration_max_min, confidence_boost, auto_merge_gap_min, notes FROM routes ORDER BY from_site, to_site`)
	if err != nil {
		return err
	}
	defer rows.Close()
	return writeDelimited(path, routes.RouteHeaders(), func(w *csv.Writer) error {
		for rows.Next() {
			var r routes.Route
			if err := rows.Scan(&r.Name, &r.FromSite, &r.ToSite, &r.DistanceMinKM, &r.DistanceMaxKM, &r.DurationMinMin, &r.DurationMaxMin, &r.ConfidenceBoost, &r.AutoMergeGapMin, &r.Notes); err != nil {
				return err
			}
			if err := w.Write(routes.RouteRow(r)); err != nil {
				return err
			}
		}
		return rows.Err()
	})
}

func rawPointHash(p gps.RawPoint) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%.8f|%.8f|%.2f|%.2f|%.2f|%s", strings.TrimSpace(p.RawDT), p.Lat, p.Lng, p.Altitude, p.Angle, p.SpeedKPH, strings.TrimSpace(p.ParamsRaw))
	return hex.EncodeToString(h.Sum(nil))
}

func writeDelimited(path string, headers []string, writeRows func(*csv.Writer) error) error {
	var f *os.File
	var err error
	if strings.TrimSpace(path) == "" || path == "-" {
		f = os.Stdout
	} else {
		f, err = os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
	}
	w := csv.NewWriter(f)
	w.Comma = '\t'
	if err := w.Write(headers); err != nil {
		return err
	}
	if err := writeRows(w); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func fmtFloat(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%.2f", v)
}
