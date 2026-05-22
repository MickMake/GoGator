package store

import (
	"database/sql"
	"fmt"
	"os"

	_ "gogator/internal/storage/sqlite"
)

const DefaultPath = "gogator.sqlite"

type Counts struct {
	GPSPoints      int64
	Sites          int64
	Routes         int64
	ProcessingRuns int64
	Trips          int64
	Issues         int64
}

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Init(path string) error {
	db, err := Open(path)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func SQLiteVersion(db *sql.DB) (string, error) {
	var version string
	if err := db.QueryRow(`select sqlite_version()`).Scan(&version); err != nil {
		return "", err
	}
	return version, nil
}

func Status(path string) (Counts, string, error) {
	db, err := Open(path)
	if err != nil {
		return Counts{}, "", err
	}
	defer db.Close()

	var c Counts
	if err := queryCount(db, "gps_points", &c.GPSPoints); err != nil {
		return Counts{}, "", err
	}
	if err := queryCount(db, "sites", &c.Sites); err != nil {
		return Counts{}, "", err
	}
	if err := queryCount(db, "routes", &c.Routes); err != nil {
		return Counts{}, "", err
	}
	if err := queryCount(db, "processing_runs", &c.ProcessingRuns); err != nil {
		return Counts{}, "", err
	}
	if err := queryCount(db, "trips", &c.Trips); err != nil {
		return Counts{}, "", err
	}
	if err := queryCount(db, "issues", &c.Issues); err != nil {
		return Counts{}, "", err
	}

	version, err := SQLiteVersion(db)
	if err != nil {
		return Counts{}, "", err
	}
	return c, version, nil
}

func queryCount(db *sql.DB, table string, out *int64) error {
	q := fmt.Sprintf("select count(*) from %s", table)
	if err := db.QueryRow(q).Scan(out); err != nil {
		return fmt.Errorf("count %s: %w", table, err)
	}
	return nil
}

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,

	`CREATE TABLE IF NOT EXISTS gps_points (
    id INTEGER PRIMARY KEY,
    point_hash TEXT NOT NULL UNIQUE,
    first_source_file TEXT,
    first_raw_row INTEGER,
    last_source_file TEXT,
    last_raw_row INTEGER,
    seen_count INTEGER NOT NULL DEFAULT 1,
    raw_dt TEXT NOT NULL,
    normalised_time TEXT NOT NULL,
    lat REAL NOT NULL,
    lng REAL NOT NULL,
    altitude REAL,
    angle REAL,
    speed_kph REAL,
    params_raw TEXT,
    params_json TEXT,
    imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,

	`CREATE INDEX IF NOT EXISTS idx_gps_points_time ON gps_points(normalised_time);`,
	`CREATE INDEX IF NOT EXISTS idx_gps_points_lat_lng ON gps_points(lat, lng);`,

	`CREATE TABLE IF NOT EXISTS gps_point_sources (
    id INTEGER PRIMARY KEY,
    gps_point_id INTEGER NOT NULL,
    source_file TEXT NOT NULL,
    raw_row INTEGER NOT NULL,
    imported_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (gps_point_id) REFERENCES gps_points(id),
    UNIQUE (gps_point_id, source_file, raw_row)
);`,

	`CREATE TABLE IF NOT EXISTS sites (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    address TEXT,
    lat REAL NOT NULL,
    lng REAL NOT NULL,
    range_m REAL,
    min_destination_minutes REAL,
    type TEXT,
    important INTEGER NOT NULL DEFAULT 1,
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);`,

	`CREATE TABLE IF NOT EXISTS routes (
    id INTEGER PRIMARY KEY,
    from_site_id INTEGER NOT NULL,
    to_site_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    confidence TEXT,
    notes TEXT,
    expected_distance_min_km REAL,
    expected_distance_max_km REAL,
    expected_duration_min_min REAL,
    expected_duration_max_min REAL,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_site_id) REFERENCES sites(id),
    FOREIGN KEY (to_site_id) REFERENCES sites(id),
    UNIQUE (from_site_id, to_site_id)
);`,

	`CREATE TABLE IF NOT EXISTS processing_runs (
    id INTEGER PRIMARY KEY,
    started_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TEXT,
    status TEXT NOT NULL DEFAULT 'RUNNING',
    app_version TEXT,
    algorithm_version TEXT,
    config_hash TEXT,
    config_json TEXT,
    gps_start_time TEXT,
    gps_end_time TEXT,
    notes TEXT
);`,

	`CREATE TABLE IF NOT EXISTS trips (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    trip_index INTEGER NOT NULL,
    trip_status TEXT NOT NULL,
    jitter_category TEXT,
    start_gps_point_id INTEGER,
    end_gps_point_id INTEGER,
    departure_time TEXT NOT NULL,
    destination_time TEXT NOT NULL,
    departure_site_id INTEGER,
    destination_site_id INTEGER,
    departure_lat REAL,
    departure_lng REAL,
    destination_lat REAL,
    destination_lng REAL,
    distance_km REAL,
    duration_hours REAL,
    top_speed_kph REAL,
    average_speed_kph REAL,
    site_duration_hours REAL,
    continuity_status TEXT,
    route_id INTEGER,
    route_match_status TEXT,
    route_diagnostic_status TEXT,
    route_notes TEXT,
    flags TEXT,
    FOREIGN KEY (run_id) REFERENCES processing_runs(id),
    FOREIGN KEY (start_gps_point_id) REFERENCES gps_points(id),
    FOREIGN KEY (end_gps_point_id) REFERENCES gps_points(id),
    FOREIGN KEY (departure_site_id) REFERENCES sites(id),
    FOREIGN KEY (destination_site_id) REFERENCES sites(id),
    FOREIGN KEY (route_id) REFERENCES routes(id),
    UNIQUE (run_id, trip_index)
);`,

	`CREATE TABLE IF NOT EXISTS trip_waypoints (
    id INTEGER PRIMARY KEY,
    trip_id INTEGER NOT NULL,
    gps_point_id INTEGER NOT NULL,
    point_index INTEGER NOT NULL,
    FOREIGN KEY (trip_id) REFERENCES trips(id),
    FOREIGN KEY (gps_point_id) REFERENCES gps_points(id),
    UNIQUE (trip_id, gps_point_id),
    UNIQUE (trip_id, point_index)
);`,

	`CREATE TABLE IF NOT EXISTS gps_point_classifications (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    gps_point_id INTEGER NOT NULL,
    point_status TEXT NOT NULL,
    movement_status TEXT,
    quality_status TEXT,
    reason TEXT,
    flags TEXT,
    FOREIGN KEY (run_id) REFERENCES processing_runs(id),
    FOREIGN KEY (gps_point_id) REFERENCES gps_points(id),
    UNIQUE (run_id, gps_point_id)
);`,

	`CREATE TABLE IF NOT EXISTS route_stats (
    id INTEGER PRIMARY KEY,
    run_id INTEGER NOT NULL,
    from_site_id INTEGER NOT NULL,
    to_site_id INTEGER NOT NULL,
    trip_count INTEGER NOT NULL,
    median_distance_km REAL,
    min_distance_km REAL,
    max_distance_km REAL,
    median_duration_min REAL,
    min_duration_min REAL,
    max_duration_min REAL,
    first_seen TEXT,
    last_seen TEXT,
    suggested_name TEXT,
    FOREIGN KEY (run_id) REFERENCES processing_runs(id),
    FOREIGN KEY (from_site_id) REFERENCES sites(id),
    FOREIGN KEY (to_site_id) REFERENCES sites(id),
    UNIQUE (run_id, from_site_id, to_site_id)
);`,

	`CREATE TABLE IF NOT EXISTS issues (
    id INTEGER PRIMARY KEY,
    run_id INTEGER,
    issue_type TEXT NOT NULL,
    gps_point_id INTEGER,
    trip_id INTEGER,
    route_id INTEGER,
    site_id INTEGER,
    status TEXT,
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (run_id) REFERENCES processing_runs(id),
    FOREIGN KEY (gps_point_id) REFERENCES gps_points(id),
    FOREIGN KEY (trip_id) REFERENCES trips(id),
    FOREIGN KEY (route_id) REFERENCES routes(id),
    FOREIGN KEY (site_id) REFERENCES sites(id)
);`,
}
