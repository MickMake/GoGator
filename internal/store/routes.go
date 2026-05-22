package store

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type RouteRecord struct {
	FromSite                  string
	ToSite                    string
	Name                      string
	Confidence                string
	Notes                     string
	ExpectedDistanceMinKM     sql.NullFloat64
	ExpectedDistanceMaxKM     sql.NullFloat64
	ExpectedDurationMinMin    sql.NullFloat64
	ExpectedDurationMaxMin    sql.NullFloat64
}

var routeHeader = []string{"From", "To", "Name", "Confidence", "Notes", "Expected Distance Min Km", "Expected Distance Max Km", "Expected Duration Min Min", "Expected Duration Max Min"}

func ImportRoutes(path string) (int, error) {
	db, err := Open(DefaultPath)
	if err != nil { return 0, err }
	defer db.Close()
	routes, err := parseRoutesFile(path)
	if err != nil { return 0, err }
	tx, err := db.Begin()
	if err != nil { return 0, err }
	defer tx.Rollback()
	for _, r := range routes {
		if err := upsertRouteTx(tx, r); err != nil { return 0, err }
	}
	if err := tx.Commit(); err != nil { return 0, err }
	return len(routes), nil
}

func ExportRoutes(path string) error {
	db, err := Open(DefaultPath)
	if err != nil { return err }
	defer db.Close()
	rows, err := db.Query(`SELECT fs.name,ts.name,r.name,COALESCE(r.confidence,''),COALESCE(r.notes,''),r.expected_distance_min_km,r.expected_distance_max_km,r.expected_duration_min_min,r.expected_duration_max_min FROM routes r JOIN sites fs ON fs.id=r.from_site_id JOIN sites ts ON ts.id=r.to_site_id ORDER BY fs.name,ts.name`)
	if err != nil { return err }
	defer rows.Close()
	f, err := os.Create(path)
	if err != nil { return err }
	defer f.Close()
	w := csv.NewWriter(f)
	w.Comma = '\t'
	if err := w.Write(routeHeader); err != nil { return err }
	for rows.Next() {
		var r RouteRecord
		if err := rows.Scan(&r.FromSite, &r.ToSite, &r.Name, &r.Confidence, &r.Notes, &r.ExpectedDistanceMinKM, &r.ExpectedDistanceMaxKM, &r.ExpectedDurationMinMin, &r.ExpectedDurationMaxMin); err != nil { return err }
rec := []string{
	r.FromSite,
	r.ToSite,
	r.Name,
	r.Confidence,
	r.Notes,
	trimNullFloat(r.ExpectedDistanceMinKM),
	trimNullFloat(r.ExpectedDistanceMaxKM),
	trimNullFloat(r.ExpectedDurationMinMin),
	trimNullFloat(r.ExpectedDurationMaxMin),
}

		if err := w.Write(rec); err != nil { return err }
	}
	if err := rows.Err(); err != nil { return err }
	w.Flush()
	return w.Error()
}

func UpsertRoute(r RouteRecord) error {
	db, err := Open(DefaultPath)
	if err != nil { return err }
	defer db.Close()
	return upsertRouteTx(db, r)
}

func DeleteRoute(fromSite, toSite string) error {
	db, err := Open(DefaultPath)
	if err != nil { return err }
	defer db.Close()
	fromID, err := siteID(db, fromSite)
	if err != nil { return err }
	toID, err := siteID(db, toSite)
	if err != nil { return err }
	res, err := db.Exec(`DELETE FROM routes WHERE from_site_id=? AND to_site_id=?`, fromID, toID)
	if err != nil { return err }
	n, _ := res.RowsAffected()
	if n == 0 { return fmt.Errorf("route not found: %s -> %s", fromSite, toSite) }
	return nil
}

func upsertRouteTx(exec interface{ Exec(string, ...any) (sql.Result, error); QueryRow(string, ...any) *sql.Row }, r RouteRecord) error {
	fromID, err := siteID(exec, r.FromSite)
	if err != nil { return err }
	toID, err := siteID(exec, r.ToSite)
	if err != nil { return err }
	name := strings.TrimSpace(r.Name)
	if name == "" { name = fmt.Sprintf("%s -> %s", r.FromSite, r.ToSite) }
	_, err = exec.Exec(`INSERT INTO routes(from_site_id,to_site_id,name,confidence,notes,expected_distance_min_km,expected_distance_max_km,expected_duration_min_min,expected_duration_max_min,updated_at)
VALUES(?,?,?,?,?,?,?,?,?,CURRENT_TIMESTAMP)
ON CONFLICT(from_site_id,to_site_id) DO UPDATE SET name=excluded.name,confidence=excluded.confidence,notes=excluded.notes,expected_distance_min_km=excluded.expected_distance_min_km,expected_distance_max_km=excluded.expected_distance_max_km,expected_duration_min_min=excluded.expected_duration_min_min,expected_duration_max_min=excluded.expected_duration_max_min,updated_at=CURRENT_TIMESTAMP`,
		fromID, toID, name, r.Confidence, r.Notes, r.ExpectedDistanceMinKM, r.ExpectedDistanceMaxKM, r.ExpectedDurationMinMin, r.ExpectedDurationMaxMin)
	return err
}

func siteID(q interface{ QueryRow(string, ...any) *sql.Row }, name string) (int64, error) {
	var id int64
	if err := q.QueryRow(`SELECT id FROM sites WHERE name=?`, name).Scan(&id); err != nil {
		if err == sql.ErrNoRows { return 0, fmt.Errorf("unknown site %q", name) }
		return 0, err
	}
	return id, nil
}

func parseRoutesFile(path string) ([]RouteRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil { return nil, err }
	delim := ','
	for _, raw := range bytes.Split(data, []byte("\n")) {
		line := strings.TrimSpace(strings.TrimPrefix(string(raw), "\ufeff"))
		if line == "" { continue }
		if strings.Contains(line, "\t") { delim = '\t' }
		break
	}
	r := csv.NewReader(bytes.NewReader(data))
	r.Comma = delim
	r.FieldsPerRecord = -1
	r.TrimLeadingSpace = false
	rows, err := r.ReadAll()
	if err != nil { return nil, fmt.Errorf("%s: %w", path, err) }
	if len(rows) < 1 { return nil, fmt.Errorf("%s: missing header row", path) }
	idx := map[string]int{}
	for i, h := range rows[0] { idx[strings.ToLower(strings.TrimSpace(strings.TrimPrefix(h, "\ufeff")))] = i }
	req := []string{"from", "to"}
	for _, c := range req { if _, ok := idx[c]; !ok { return nil, fmt.Errorf("%s: missing required column: %s", path, strings.Title(c)) } }
	get := func(row []string, key string) string { if i,ok:=idx[key]; ok && i < len(row) { return strings.TrimSpace(row[i])}; return "" }
	var out []RouteRecord
	for ln, row := range rows[1:] {
		rowNum := ln + 2
		if len(strings.TrimSpace(strings.Join(row, ""))) == 0 { continue }
		from := get(row, "from"); to := get(row, "to")
		if from == "" { return nil, fmt.Errorf("%s: row %d: column From: missing value", path, rowNum) }
		if to == "" { return nil, fmt.Errorf("%s: row %d: column To: missing value", path, rowNum) }

parseNum := func(key, label string) (sql.NullFloat64, error) {
	v := get(row, key)
	if v == "" {
		return sql.NullFloat64{}, nil
	}

	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return sql.NullFloat64{}, fmt.Errorf("%s: row %d: column %s: invalid number", path, rowNum, label)
	}

	return sql.NullFloat64{Float64: n, Valid: true}, nil
}

dmin, err := parseNum("expected distance min km", "Expected Distance Min Km")
if err != nil {
	return nil, err
}

dmax, err := parseNum("expected distance max km", "Expected Distance Max Km")
if err != nil {
	return nil, err
}

tmin, err := parseNum("expected duration min min", "Expected Duration Min Min")
if err != nil {
	return nil, err
}

tmax, err := parseNum("expected duration max min", "Expected Duration Max Min")
if err != nil {
	return nil, err
}

out = append(out, RouteRecord{
	FromSite:               from,
	ToSite:                 to,
	Name:                   get(row, "name"),
	Confidence:             get(row, "confidence"),
	Notes:                  get(row, "notes"),
	ExpectedDistanceMinKM:  dmin,
	ExpectedDistanceMaxKM:  dmax,
	ExpectedDurationMinMin: tmin,
	ExpectedDurationMaxMin: tmax,
})
	}
	return out, nil
}

func trimNullFloat(f sql.NullFloat64) string {
	if !f.Valid {
		return ""
	}
	return trimFloat(f.Float64)
}

