package store

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestInitIsIdempotentAndStatusCounts(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "gogator.sqlite")

	if err := Init(path); err != nil {
		t.Fatalf("init first: %v", err)
	}
	if err := Init(path); err != nil {
		t.Fatalf("init second: %v", err)
	}

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

func TestInitCreatesExpectedRun1Schema(t *testing.T) {
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

	requireTables(t, db,
		"gps_points",
		"gps_point_sources",
		"sites",
		"routes",
		"processing_runs",
		"trips",
		"trip_waypoints",
		"gps_point_classifications",
		"route_stats",
		"issues",
	)

	requireColumns(t, db, "gps_points",
		"id", "point_hash", "first_source_file", "first_raw_row", "last_source_file", "last_raw_row",
		"seen_count", "raw_dt", "normalised_time", "lat", "lng", "altitude", "angle", "speed_kph",
		"params_raw", "params_json", "imported_at",
	)
	requireIndexWithColumns(t, db, "gps_points", "idx_gps_points_time", "normalised_time")
	requireIndexWithColumns(t, db, "gps_points", "idx_gps_points_lat_lng", "lat", "lng")

	requireColumns(t, db, "gps_point_sources", "id", "gps_point_id", "source_file", "raw_row", "imported_at")
	requireUniqueConstraintByColumns(t, db, "gps_point_sources", "gps_point_id", "source_file", "raw_row")

	requireColumns(t, db, "sites",
		"id", "name", "address", "lat", "lng", "range_m", "min_destination_minutes", "type", "important", "notes", "created_at", "updated_at",
	)

	requireColumns(t, db, "routes",
		"id", "from_site_id", "to_site_id", "name", "confidence", "notes",
		"expected_distance_min_km", "expected_distance_max_km", "expected_duration_min_min", "expected_duration_max_min",
		"created_at", "updated_at",
	)
	requireUniqueConstraintByColumns(t, db, "routes", "from_site_id", "to_site_id")

	requireColumns(t, db, "processing_runs",
		"id", "started_at", "completed_at", "status", "app_version", "algorithm_version", "config_hash", "config_json", "gps_start_time", "gps_end_time", "notes",
	)

	requireColumns(t, db, "trips",
		"id", "run_id", "trip_index", "trip_status", "jitter_category", "start_gps_point_id", "end_gps_point_id",
		"departure_time", "destination_time", "departure_site_id", "destination_site_id", "departure_lat", "departure_lng", "destination_lat", "destination_lng",
		"distance_km", "duration_hours", "top_speed_kph", "average_speed_kph", "site_duration_hours", "continuity_status", "route_id",
		"route_match_status", "route_diagnostic_status", "route_notes", "flags",
	)
	requireUniqueConstraintByColumns(t, db, "trips", "run_id", "trip_index")

	requireColumns(t, db, "trip_waypoints", "id", "trip_id", "gps_point_id", "point_index")
	requireUniqueConstraintByColumns(t, db, "trip_waypoints", "trip_id", "gps_point_id")
	requireUniqueConstraintByColumns(t, db, "trip_waypoints", "trip_id", "point_index")

	requireColumns(t, db, "gps_point_classifications",
		"id", "run_id", "gps_point_id", "point_status", "movement_status", "quality_status", "reason", "flags",
	)
	requireUniqueConstraintByColumns(t, db, "gps_point_classifications", "run_id", "gps_point_id")

	requireColumns(t, db, "route_stats",
		"id", "run_id", "from_site_id", "to_site_id", "trip_count", "median_distance_km", "min_distance_km", "max_distance_km",
		"median_duration_min", "min_duration_min", "max_duration_min", "first_seen", "last_seen", "suggested_name",
	)
	requireUniqueConstraintByColumns(t, db, "route_stats", "run_id", "from_site_id", "to_site_id")

	requireColumns(t, db, "issues",
		"id", "run_id", "issue_type", "gps_point_id", "trip_id", "route_id", "site_id", "status", "notes", "created_at",
	)
}

func TestStatusOnMissingDBReturnsHelpfulError(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "missing", "gogator.sqlite")

	_, _, err := Status(path)
	if err == nil {
		t.Fatalf("expected status error for missing db path")
	}
	if !strings.Contains(err.Error(), "unable to open database file") {
		t.Fatalf("expected helpful missing-db error, got: %v", err)
	}
}

func requireTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()
	for _, table := range tables {
		if !tableExists(t, db, table) {
			t.Fatalf("expected table %q to exist", table)
		}
	}
}

func tableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()
	var found string
	err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name = ?`, table).Scan(&found)
	if err == sql.ErrNoRows {
		return false
	}
	if err != nil {
		t.Fatalf("query sqlite_master for %q: %v", table, err)
	}
	return found == table
}

func requireColumns(t *testing.T, db *sql.DB, table string, want ...string) {
	t.Helper()
	got := tableColumns(t, db, table)
	for _, col := range want {
		if !slices.Contains(got, col) {
			t.Fatalf("table %q missing column %q; got columns=%v", table, col, got)
		}
	}
}

func tableColumns(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		t.Fatalf("pragma table_info(%q): %v", table, err)
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan pragma table_info(%q): %v", table, err)
		}
		cols = append(cols, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate pragma table_info(%q): %v", table, err)
	}
	return cols
}

func requireIndexWithColumns(t *testing.T, db *sql.DB, table, idxName string, cols ...string) {
	t.Helper()
	indexes := listIndexes(t, db, table)
	if _, ok := indexes[idxName]; !ok {
		t.Fatalf("table %q missing index %q; indexes=%v", table, idxName, mapsKeys(indexes))
	}
	got := indexColumns(t, db, idxName)
	if !slices.Equal(got, cols) {
		t.Fatalf("index %q columns mismatch; got=%v want=%v", idxName, got, cols)
	}
}

func requireUniqueConstraintByColumns(t *testing.T, db *sql.DB, table string, cols ...string) {
	t.Helper()
	indexes := listIndexes(t, db, table)
	for idx, unique := range indexes {
		if !unique {
			continue
		}
		if slices.Equal(indexColumns(t, db, idx), cols) {
			return
		}
	}
	t.Fatalf("table %q missing unique constraint for columns=%v", table, cols)
}

func listIndexes(t *testing.T, db *sql.DB, table string) map[string]bool {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf(`PRAGMA index_list(%s)`, table))
	if err != nil {
		t.Fatalf("pragma index_list(%q): %v", table, err)
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var seq int
		var name string
		var unique int
		var origin string
		var partial int
		if err := rows.Scan(&seq, &name, &unique, &origin, &partial); err != nil {
			t.Fatalf("scan pragma index_list(%q): %v", table, err)
		}
		out[name] = unique == 1
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate pragma index_list(%q): %v", table, err)
	}
	return out
}

func indexColumns(t *testing.T, db *sql.DB, index string) []string {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf(`PRAGMA index_info(%s)`, index))
	if err != nil {
		t.Fatalf("pragma index_info(%q): %v", index, err)
	}
	defer rows.Close()

	var cols []string
	for rows.Next() {
		var seqno, cid int
		var name string
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			t.Fatalf("scan pragma index_info(%q): %v", index, err)
		}
		cols = append(cols, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate pragma index_info(%q): %v", index, err)
	}
	return cols
}

func mapsKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
