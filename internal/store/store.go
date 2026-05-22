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
		point_hash text not null unique,
		raw_dt text not null,
		normalised_time text not null,
		lat real not null,
		lng real not null,
		altitude real,
		angle real,
		speed_kph real,
		params_raw text,
		params_json text,
		first_source_file text,
		first_raw_row integer,
		last_source_file text,
		last_raw_row integer,
		seen_count integer not null default 1,
		imported_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
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
		name text not null unique,
		address text,
		lat real not null,
		lng real not null,
		range_m real,
		min_destination_minutes real,
		type text,
		important integer not null default 1,
		notes text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now'))
	)`,
	`create table if not exists routes (
		id integer primary key,
		name text not null,
		from_site_id integer not null,
		to_site_id integer not null,
		expected_distance_min_km real,
		expected_distance_max_km real,
		expected_duration_min_min real,
		expected_duration_max_min real,
		confidence real,
		notes text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		foreign key (from_site_id) references sites(id) on delete restrict,
		foreign key (to_site_id) references sites(id) on delete restrict,
		unique (from_site_id, to_site_id)
	)`,
	`create table if not exists processing_runs (
		id integer primary key,
		started_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		completed_at text,
		status text not null default 'RUNNING',
		app_version text,
		algorithm_version text,
		config_hash text,
		config_json text,
		gps_start_time text,
		gps_end_time text,
		notes text
	)`,
	`create table if not exists trips (
		id integer primary key,
		run_id integer not null,
		trip_index integer not null,
		from_site_id integer,
		to_site_id integer,
		departure_time text,
		arrival_time text,
		duration_minutes real,
		distance_km real,
		point_count integer,
		movement_points integer,
		stationary_points integer,
		is_check integer not null default 0,
		route_id integer,
		notes text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		foreign key (run_id) references processing_runs(id) on delete cascade,
		foreign key (from_site_id) references sites(id) on delete set null,
		foreign key (to_site_id) references sites(id) on delete set null,
		foreign key (route_id) references routes(id) on delete set null,
		unique (run_id, trip_index)
	)`,
	`create table if not exists trip_waypoints (
		id integer primary key,
		trip_id integer not null,
		gps_point_id integer not null,
		seq integer not null,
		time text,
		lat real,
		lng real,
		speed_kph real,
		foreign key (trip_id) references trips(id) on delete cascade,
		foreign key (gps_point_id) references gps_points(id) on delete cascade,
		unique (trip_id, seq)
	)`,
	`create table if not exists gps_point_classifications (
		id integer primary key,
		run_id integer not null,
		gps_point_id integer not null,
		point_status text not null,
		movement_status text,
		quality_status text,
		reason text,
		flags text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		foreign key (run_id) references processing_runs(id) on delete cascade,
		foreign key (gps_point_id) references gps_points(id) on delete cascade,
		unique (run_id, gps_point_id)
	)`,
	`create table if not exists route_stats (
		id integer primary key,
		run_id integer not null,
		from_site_id integer,
		to_site_id integer,
		trip_count integer not null default 0,
		median_distance_km real,
		min_distance_km real,
		max_distance_km real,
		median_duration_min real,
		min_duration_min real,
		max_duration_min real,
		first_seen text,
		last_seen text,
		suggested_name text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		updated_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		foreign key (run_id) references processing_runs(id) on delete cascade,
		foreign key (from_site_id) references sites(id) on delete set null,
		foreign key (to_site_id) references sites(id) on delete set null
	)`,
	`create table if not exists issues (
		id integer primary key,
		run_id integer not null,
		issue_type text not null,
		gps_point_id integer,
		trip_id integer,
		route_id integer,
		site_id integer,
		status text,
		notes text,
		created_at text not null default (strftime('%Y-%m-%dT%H:%M:%fZ','now')),
		foreign key (run_id) references processing_runs(id) on delete cascade,
		foreign key (gps_point_id) references gps_points(id) on delete set null,
		foreign key (trip_id) references trips(id) on delete set null,
		foreign key (route_id) references routes(id) on delete set null,
		foreign key (site_id) references sites(id) on delete set null
	)`,
	`create index if not exists idx_gps_points_time on gps_points(normalised_time)`,
	`create index if not exists idx_gps_points_lat_lng on gps_points(lat, lng)`,
	`create index if not exists idx_gps_point_sources_point on gps_point_sources(gps_point_id)`,
	`create index if not exists idx_routes_from_to on routes(from_site_id, to_site_id)`,
	`create index if not exists idx_trips_run on trips(run_id)`,
	`create index if not exists idx_issues_issue_type on issues(issue_type)`,
}
