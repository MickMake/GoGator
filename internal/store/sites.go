package store

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SiteRecord struct {
	Name                  string
	Address               string
	Lat                   float64
	Lng                   float64
	RangeM                float64
	MinDestinationMinutes float64
	Type                  string
	Important             bool
	Notes                 string
}

var siteHeader = []string{"Site", "Address", "GPS", "Range", "Min Destination Minutes", "Type", "Important", "Notes"}

func ImportSites(path string) (int, error) {
	db, err := Open(DefaultPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	sites, err := parseSitesFile(path)
	if err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	for _, s := range sites {
		if err := upsertSiteTx(tx, s); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(sites), nil
}

func ExportSites(path string) error {
	db, err := Open(DefaultPath)
	if err != nil {
		return err
	}
	defer db.Close()
	rows, err := db.Query(`SELECT name,address,lat,lng,range_m,min_destination_minutes,type,important,COALESCE(notes,'') FROM sites ORDER BY name`)
	if err != nil {
		return err
	}
	defer rows.Close()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	w.Comma = '\t'
	if err := w.Write(siteHeader); err != nil {
		return err
	}
	for rows.Next() {
		var s SiteRecord
		if err := rows.Scan(&s.Name, &s.Address, &s.Lat, &s.Lng, &s.RangeM, &s.MinDestinationMinutes, &s.Type, &s.Important, &s.Notes); err != nil {
			return err
		}
		rec := []string{s.Name, s.Address, fmt.Sprintf("%.8f,%.8f", s.Lat, s.Lng), trimFloat(s.RangeM), trimFloat(s.MinDestinationMinutes), s.Type, boolText(s.Important), s.Notes}
		if err := w.Write(rec); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	w.Flush()
	return w.Error()
}

func UpsertSite(s SiteRecord) error {
	db, err := Open(DefaultPath)
	if err != nil {
		return err
	}
	defer db.Close()
	return upsertSiteTx(db, s)
}

func DeleteSite(name string, anyway bool) error {
	db, err := Open(DefaultPath)
	if err != nil {
		return err
	}
	defer db.Close()
	var id int64
	if err := db.QueryRow(`SELECT id FROM sites WHERE name=?`, name).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("site not found: %s", name)
		}
		return err
	}
	if !anyway {
		if err := referencedCheck(db, id); err != nil {
			return err
		}
	}
	if _, err := db.Exec(`DELETE FROM sites WHERE id=?`, id); err != nil {
		return fmt.Errorf("delete site blocked: %w", err)
	}
	return nil
}

func referencedCheck(db *sql.DB, siteID int64) error {
	var n int
	_ = db.QueryRow(`SELECT count(*) FROM routes WHERE from_site_id=? OR to_site_id=?`, siteID, siteID).Scan(&n)
	if n > 0 {
		return fmt.Errorf("site is referenced by %d route(s)", n)
	}
	_ = db.QueryRow(`SELECT count(*) FROM trips WHERE departure_site_id=? OR destination_site_id=?`, siteID, siteID).Scan(&n)
	if n > 0 {
		return fmt.Errorf("site is referenced by %d trip(s)", n)
	}
	return nil
}

func upsertSiteTx(exec interface {
	Exec(string, ...any) (sql.Result, error)
}, s SiteRecord) error {
	_, err := exec.Exec(`INSERT INTO sites(name,address,lat,lng,range_m,min_destination_minutes,type,important,notes,updated_at)
VALUES(?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
ON CONFLICT(name) DO UPDATE SET address=excluded.address,lat=excluded.lat,lng=excluded.lng,range_m=excluded.range_m,min_destination_minutes=excluded.min_destination_minutes,type=excluded.type,important=excluded.important,notes=excluded.notes,updated_at=CURRENT_TIMESTAMP`,
		s.Name, s.Address, s.Lat, s.Lng, s.RangeM, s.MinDestinationMinutes, s.Type, boolInt(s.Important), s.Notes)
	return err
}

func parseSitesFile(path string) ([]SiteRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	delim := ','
	for _, raw := range bytes.Split(data, []byte("\n")) {
		line := strings.TrimSpace(strings.TrimPrefix(string(raw), "\ufeff"))
		if line == "" {
			continue
		}
		if strings.Contains(line, "\t") {
			delim = '\t'
		}
		break
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = delim
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = true
	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(rows) < 1 {
		return nil, fmt.Errorf("missing header row")
	}
	idx := map[string]int{}
	for i, h := range rows[0] {
		idx[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")))] = i
	}
	if _, ok := idx["site"]; !ok {
		return nil, fmt.Errorf("missing required column: Site")
	}
	if _, ok := idx["gps"]; !ok {
		return nil, fmt.Errorf("missing required column: GPS")
	}
	var out []SiteRecord
	for ln, row := range rows[1:] {
		if len(strings.TrimSpace(strings.Join(row, ""))) == 0 {
			continue
		}
		get := func(k string) string {
			if i, ok := idx[k]; ok && i < len(row) {
				return strings.TrimSpace(row[i])
			}
			return ""
		}
		name := get("site")
		if name == "" {
			return nil, fmt.Errorf("row %d: missing site name", ln+2)
		}
		gpsValue := get("gps")
		if gpsValue == "" {
			if latText := get("lat"); latText != "" {
				if lngText := get("lng"); lngText != "" {
					gpsValue = latText + "," + lngText
				}
			}
		}
		if gpsValue == "" {
			if gpsIdx, ok := idx["gps"]; ok && gpsIdx+1 < len(row) {
				gpsValue = strings.TrimSpace(row[gpsIdx]) + "," + strings.TrimSpace(row[gpsIdx+1])
			}
		}
		lat, lng, ok := parseGPS(gpsValue)
		if !ok {
			return nil, fmt.Errorf("row %d: invalid gps", ln+2)
		}
		rng, err := parseOptionalFloat(get("range"))
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid range", ln+2)
		}
		dwell, err := parseOptionalFloat(get("min destination minutes"))
		if err != nil {
			return nil, fmt.Errorf("row %d: invalid min destination minutes", ln+2)
		}
		out = append(out, SiteRecord{Name: name, Address: get("address"), Lat: lat, Lng: lng, RangeM: rng, MinDestinationMinutes: dwell, Type: get("type"), Important: parseBool(get("important"), true), Notes: get("notes")})
	}
	return out, nil
}
func parseGPS(s string) (float64, float64, bool) {
	parts := strings.Split(strings.TrimSpace(s), ",")
	if len(parts) < 2 {
		return 0, 0, false
	}
	lat, e1 := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, e2 := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	return lat, lng, e1 == nil && e2 == nil
}
func parseBool(s string, def bool) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return def
	}
	return s == "1" || s == "true" || s == "yes" || s == "y"
}
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
func boolText(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
func trimFloat(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }
func parseOptionalFloat(s string) (float64, error) {
	if strings.TrimSpace(s) == "" {
		return 0, nil
	}
	return strconv.ParseFloat(strings.TrimSpace(s), 64)
}
