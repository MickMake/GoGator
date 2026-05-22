package store

import (
	"database/sql"
	"path/filepath"
	"testing"
)

func TestInitSchemaAndIndexes(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "gogator.sqlite")

	if err := Init(path); err != nil {
		t.Fatalf("init first: %v", err)
	}
	if err := Init(path); err != nil {
		t.Fatalf("init second: %v", err)
	}

	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	requireTables(t, db,
		"gps_points", "gps_point_sources", "sites", "routes", "processing_runs",
		"trips", "trip_waypoints", "gps_point_classifications", "route_stats", "issues",
	)

	requireColumns(t, db, "gps_points", "point_hash")
	requireColumns(t, db, "sites", "name")
	requireColumns(t, db, "routes", "from_site_id", "to_site_id")
	requireColumns(t, db, "processing_runs", "completed_at")
	requireColumns(t, db, "trips", "trip_index")
	requireColumns(t, db, "gps_point_classifications", "run_id")
	requireColumns(t, db, "route_stats", "median_distance_km")
	requireColumns(t, db, "issues", "issue_type")

	requireIndexes(t, db, "idx_gps_points_time", "idx_gps_points_lat_lng")

	counts, version, err := Status(path)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if version == "" {
		t.Fatalf("expected sqlite version")
	}
	if counts.GPSPoints != 0 || counts.Sites != 0 || counts.Routes != 0 || counts.ProcessingRuns != 0 || counts.Trips != 0 || counts.Issues != 0 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}

func requireTables(t *testing.T, db *sql.DB, names ...string) {
	t.Helper()
	for _, name := range names {
		var n string
		err := db.QueryRow(`select name from sqlite_master where type='table' and name=?`, name).Scan(&n)
		if err != nil {
			t.Fatalf("missing table %s: %v", name, err)
		}
	}
}

func requireColumns(t *testing.T, db *sql.DB, table string, want ...string) {
	t.Helper()
	rows, err := db.Query("pragma table_info(" + table + ")")
	if err != nil {
		t.Fatalf("table_info %s: %v", table, err)
	}
	defer rows.Close()

	has := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan table_info %s: %v", table, err)
		}
		has[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows table_info %s: %v", table, err)
	}
	for _, col := range want {
		if !has[col] {
			t.Fatalf("table %s missing column %s", table, col)
		}
	}
}

func requireIndexes(t *testing.T, db *sql.DB, names ...string) {
	t.Helper()
	for _, name := range names {
		var n string
		err := db.QueryRow(`select name from sqlite_master where type='index' and name=?`, name).Scan(&n)
		if err != nil {
			t.Fatalf("missing index %s: %v", name, err)
		}
	}
}
