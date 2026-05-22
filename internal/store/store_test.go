package store

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gogator.sqlite")
	if err := Init(path); err != nil {
		t.Fatalf("init first: %v", err)
	}
	if err := Init(path); err != nil {
		t.Fatalf("init second: %v", err)
	}
}

func TestRequiredRun1SchemaExists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gogator.sqlite")
	if err := Init(path); err != nil {
		t.Fatalf("init: %v", err)
	}
	db, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	for _, table := range []string{"gps_points", "gps_point_sources", "sites", "routes", "processing_runs", "trips", "trip_waypoints", "gps_point_classifications", "route_stats", "issues"} {
		requireTable(t, db, table)
	}
	requireColumn(t, db, "gps_points", "normalised_time")
	requireColumn(t, db, "gps_points", "speed_kph")
	for _, c := range []string{"first_source_file", "first_raw_row", "last_source_file", "last_raw_row", "seen_count"} {
		requireColumn(t, db, "gps_points", c)
	}

	requireColumnWithNotNull(t, db, "sites", "lat")
	requireColumnWithNotNull(t, db, "sites", "lng")
	requireColumnType(t, db, "sites", "range_m", "REAL")
	requireColumnType(t, db, "sites", "min_destination_minutes", "REAL")

	requireColumnWithNotNull(t, db, "routes", "name")
	requireColumn(t, db, "routes", "confidence")

	for _, c := range []string{"app_version", "algorithm_version", "config_hash", "config_json", "gps_start_time", "gps_end_time"} {
		requireColumn(t, db, "processing_runs", c)
	}
	requireDefaultContains(t, db, "processing_runs", "status", "RUNNING")

	for _, c := range []string{"run_id", "trip_index", "departure_time", "arrival_time", "distance_km", "duration_minutes"} {
		requireColumn(t, db, "trips", c)
	}

	requireColumnWithNotNull(t, db, "gps_point_classifications", "run_id")
	requireColumnWithNotNull(t, db, "gps_point_classifications", "gps_point_id")
	requireColumnWithNotNull(t, db, "gps_point_classifications", "point_status")
	for _, c := range []string{"movement_status", "quality_status", "reason", "flags"} {
		requireColumn(t, db, "gps_point_classifications", c)
	}
	requireUniqueConstraint(t, db, "gps_point_classifications", []string{"run_id", "gps_point_id"})

	for _, c := range []string{"run_id", "from_site_id", "to_site_id", "trip_count", "median_distance_km", "min_distance_km", "max_distance_km", "median_duration_min", "min_duration_min", "max_duration_min", "first_seen", "last_seen", "suggested_name"} {
		requireColumn(t, db, "route_stats", c)
	}

	for _, c := range []string{"run_id", "issue_type", "gps_point_id", "trip_id", "route_id", "site_id", "status", "notes"} {
		requireColumn(t, db, "issues", c)
	}

	requireIndexOnColumns(t, db, "idx_gps_points_time", []string{"normalised_time"})
	requireIndex(t, db, "idx_gps_points_lat_lng")
}

func TestStatusBaselineCountsOnNewDB(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gogator.sqlite")
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

func TestStatusMissingDBReturnsError(t *testing.T) {
	_, _, err := Status(filepath.Join(t.TempDir(), "missing.sqlite"))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "no such table") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func requireTable(t *testing.T, db *sql.DB, table string) {
	var n string
	if err := db.QueryRow(`select name from sqlite_master where type='table' and name=?`, table).Scan(&n); err != nil {
		t.Fatalf("missing table %s: %v", table, err)
	}
}
func requireIndex(t *testing.T, db *sql.DB, index string) {
	var n string
	if err := db.QueryRow(`select name from sqlite_master where type='index' and name=?`, index).Scan(&n); err != nil {
		t.Fatalf("missing index %s: %v", index, err)
	}
}

func requireColumn(t *testing.T, db *sql.DB, table, column string) {
	_, _, _, ok := getColumnInfo(t, db, table, column)
	if !ok {
		t.Fatalf("missing column %s.%s", table, column)
	}
}
func requireColumnWithNotNull(t *testing.T, db *sql.DB, table, column string) {
	_, _, notNull, ok := getColumnInfo(t, db, table, column)
	if !ok || !notNull {
		t.Fatalf("expected NOT NULL column %s.%s", table, column)
	}
}
func requireColumnType(t *testing.T, db *sql.DB, table, column, typ string) {
	colType, _, _, ok := getColumnInfo(t, db, table, column)
	if !ok || strings.ToUpper(colType) != typ {
		t.Fatalf("expected %s.%s type %s, got %s", table, column, typ, colType)
	}
}
func requireDefaultContains(t *testing.T, db *sql.DB, table, column, want string) {
	_, dflt, _, ok := getColumnInfo(t, db, table, column)
	if !ok || !strings.Contains(strings.ToUpper(dflt), strings.ToUpper(want)) {
		t.Fatalf("expected default for %s.%s to contain %s, got %q", table, column, want, dflt)
	}
}

func getColumnInfo(t *testing.T, db *sql.DB, table, column string) (colType, dflt string, notNull bool, ok bool) {
	rows, err := db.Query("pragma table_info(" + table + ")")
	if err != nil {
		t.Fatalf("table_info %s: %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, nn, pk int
		var name, ct string
		var dv any
		if err := rows.Scan(&cid, &name, &ct, &nn, &dv, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		if name == column {
			if dv != nil {
				dflt = dv.(string)
			}
			return ct, dflt, nn == 1, true
		}
	}
	return "", "", false, false
}

func requireUniqueConstraint(t *testing.T, db *sql.DB, table string, want []string) {
	idxRows, err := db.Query("pragma index_list(" + table + ")")
	if err != nil {
		t.Fatalf("index_list %s: %v", table, err)
	}
	defer idxRows.Close()
	for idxRows.Next() {
		var seq int
		var name string
		var unique int
		var origin, partial any
		if err := idxRows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			t.Fatalf("scan index_list: %v", err)
		}
		if unique != 1 {
			continue
		}
		if indexColumnsMatch(t, db, name, want) {
			return
		}
	}
	t.Fatalf("missing unique constraint on %s columns %v", table, want)
}

func requireIndexOnColumns(t *testing.T, db *sql.DB, index string, want []string) {
	if !indexColumnsMatch(t, db, index, want) {
		t.Fatalf("index %s does not match columns %v", index, want)
	}
}
func indexColumnsMatch(t *testing.T, db *sql.DB, index string, want []string) bool {
	rows, err := db.Query("pragma index_info(" + index + ")")
	if err != nil {
		t.Fatalf("index_info %s: %v", index, err)
	}
	defer rows.Close()
	got := []string{}
	for rows.Next() {
		var seqno, cid int
		var name string
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			t.Fatalf("scan index_info: %v", err)
		}
		got = append(got, name)
	}
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
