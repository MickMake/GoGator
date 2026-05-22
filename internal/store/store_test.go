package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitIsIdempotent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "gogator.sqlite")
	if err := Init(path); err != nil {
		t.Fatalf("init first: %v", err)
	}
	if err := Init(path); err != nil {
		t.Fatalf("init second: %v", err)
	}
}

func TestRequiredTablesColumnsAndIndexesExist(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "gogator.sqlite")
	if err := Init(path); err != nil {
		t.Fatalf("init: %v", err)
	}
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	for _, table := range []string{"gps_points", "sites", "routes", "processing_runs", "trips", "gps_point_classifications", "route_stats", "issues"} {
		var name string
		if err := db.QueryRow(`select name from sqlite_master where type='table' and name=?`, table).Scan(&name); err != nil {
			t.Fatalf("missing table %s: %v", table, err)
		}
	}
	requireColumn(t, db, "gps_points", "point_hash")
	requireColumn(t, db, "gps_points", "lat")
	requireColumn(t, db, "gps_points", "lng")
	requireColumn(t, db, "gps_points", "imported_at")
	requireColumn(t, db, "sites", "name")
	requireColumn(t, db, "routes", "from_site_id")
	requireColumn(t, db, "routes", "to_site_id")
	requireColumn(t, db, "processing_runs", "completed_at")
	requireColumn(t, db, "trips", "run_id")
	requireColumn(t, db, "trips", "trip_index")
	requireColumn(t, db, "gps_point_classifications", "run_id")
	requireColumn(t, db, "route_stats", "median_distance_km")
	requireColumn(t, db, "issues", "issue_type")
	requireIndex(t, db, "idx_gps_points_time")
	requireIndex(t, db, "idx_gps_points_lat_lng")
}

func TestStatusBaselineCountsOnNewDB(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "gogator.sqlite")
	if err := Init(path); err != nil {
		t.Fatalf("init: %v", err)
	}
	counts, version, err := Status(path)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if version == "" {
		t.Fatal("expected sqlite version")
	}
	if counts.GPSPoints != 0 || counts.Sites != 0 || counts.Routes != 0 || counts.ProcessingRuns != 0 || counts.Trips != 0 || counts.Issues != 0 {
		t.Fatalf("unexpected counts: %+v", counts)
	}
}

func TestStatusMissingDBReturnsHelpfulError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "missing.sqlite")
	_, _, err := Status(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "no such table") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func requireColumn(t *testing.T, db *sql.DB, table, column string) {
	t.Helper()
	rows, err := db.Query("pragma table_info(" + table + ")")
	if err != nil {
		t.Fatalf("table_info %s: %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan: %v", err)
		}
		if name == column {
			return
		}
	}
	t.Fatalf("missing column %s.%s", table, column)
}

func requireIndex(t *testing.T, db *sql.DB, index string) {
	t.Helper()
	var name string
	if err := db.QueryRow(`select name from sqlite_master where type='index' and name=?`, index).Scan(&name); err != nil {
		t.Fatalf("missing index %s: %v", index, err)
	}
}
