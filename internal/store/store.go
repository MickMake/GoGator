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
	`create table if not exists gps_points (
		id integer primary key,
		evidence_hash text not null unique,
		raw_dt text not null,
		normalised_time text not null,
		latitude real not null,
		longitude real not null,
		altitude real,
		angle real,
		speed_kph real,
		params_raw text,
		params_json text,
		seen_count integer not null default 1,
		first_source_file text,
		first_raw_row integer,
		last_source_file text,
		last_raw_row integer,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	)`,
	`create table if not exists gps_point_sources (
		gps_point_id integer not null,
		source_file text not null,
		raw_row integer not null,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		primary key (gps_point_id, source_file, raw_row),
		foreign key (gps_point_id) references gps_points(id) on delete cascade
	)`,
	`create table if not exists sites (
		id integer primary key,
		site_name text not null unique,
		address text,
		latitude real,
		longitude real,
		range_m integer,
		min_destination_minutes integer,
		type text,
		important integer not null default 1,
		notes text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	)`,
	`create table if not exists routes (
		id integer primary key,
		route_name text,
		from_site text not null,
		to_site text not null,
		expected_distance_min_km real,
		expected_distance_max_km real,
		expected_duration_min_min integer,
		expected_duration_max_min integer,
		confidence_boost real,
		auto_merge_gap_min integer,
		notes text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		unique (from_site, to_site)
	)`,
	`create table if not exists processing_runs (
		id integer primary key,
		started_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		finished_at text,
		status text not null default 'running',
		notes text
	)`,
	`create table if not exists trips (
		id integer primary key,
		processing_run_id integer,
		started_at text,
		ended_at text,
		from_site text,
		to_site text,
		status text,
		notes text,
		foreign key (processing_run_id) references processing_runs(id) on delete set null
	)`,
	`create table if not exists trip_waypoints (
		id integer primary key,
		trip_id integer not null,
		gps_point_id integer,
		seq integer not null,
		foreign key (trip_id) references trips(id) on delete cascade,
		foreign key (gps_point_id) references gps_points(id) on delete set null,
		unique (trip_id, seq)
	)`,
	`create table if not exists gps_point_classifications (
		id integer primary key,
		gps_point_id integer not null,
		classification text not null,
		reason text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		foreign key (gps_point_id) references gps_points(id) on delete cascade
	)`,
	`create table if not exists route_stats (
		id integer primary key,
		route_id integer,
		period_start text,
		period_end text,
		trip_count integer not null default 0,
		notes text,
		foreign key (route_id) references routes(id) on delete set null
	)`,
	`create table if not exists issues (
		id integer primary key,
		kind text not null,
		related_trip_id integer,
		related_gps_point_id integer,
		details text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		foreign key (related_trip_id) references trips(id) on delete set null,
		foreign key (related_gps_point_id) references gps_points(id) on delete set null
	)`,
	`create index if not exists idx_gps_points_normalised_time on gps_points(normalised_time)`,
	`create index if not exists idx_gps_point_sources_point on gps_point_sources(gps_point_id)`,
	`create index if not exists idx_routes_from_to on routes(from_site, to_site)`,
	`create index if not exists idx_trips_run on trips(processing_run_id)`,
	`create index if not exists idx_issues_kind on issues(kind)`,
}
