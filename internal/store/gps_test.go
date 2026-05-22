package store

import (
	"encoding/csv"
	"os"
	"reflect"
	"testing"
	"time"

	"gogator/internal/config"
)

func TestImportGPSDeduplicatesPointsAndSources(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Init(DefaultPath); err != nil {
		t.Fatal(err)
	}

	csvData := "dt,lat,lng,altitude,angle,speed,params\n" +
		"2026-05-01 00:00:00,-33.0,151.0,10,90,42,io1=1\n" +
		"2026-05-01 00:01:00,-33.1,151.1,11,91,43,io1=0\n"
	if err := os.WriteFile("gps.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}

	loc := time.UTC
	cfg := config.Default()
	first, err := ImportGPS([]string{"gps.csv"}, loc, cfg)
	if err != nil {
		t.Fatalf("first import: %v", err)
	}
	if first.Files != 1 || first.RawRows != 2 || first.GPSPoints != 2 || first.SourceRows != 2 {
		t.Fatalf("unexpected first result: %#v", first)
	}

	counts, _, err := Status(DefaultPath)
	if err != nil {
		t.Fatal(err)
	}
	if counts.GPSPoints != 2 {
		t.Fatalf("want 2 gps_points, got %d", counts.GPSPoints)
	}

	second, err := ImportGPS([]string{"gps.csv"}, loc, cfg)
	if err != nil {
		t.Fatalf("second import: %v", err)
	}
	if second.GPSPoints != 0 || second.SourceRows != 0 {
		t.Fatalf("duplicate source import was not idempotent: %#v", second)
	}

	counts, _, err = Status(DefaultPath)
	if err != nil {
		t.Fatal(err)
	}
	if counts.GPSPoints != 2 {
		t.Fatalf("want 2 gps_points after duplicate import, got %d", counts.GPSPoints)
	}
}

func TestImportGPSRecordsSecondSourceForSamePoint(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Init(DefaultPath); err != nil {
		t.Fatal(err)
	}

	csvData := "dt,lat,lng,altitude,angle,speed,params\n2026-05-01 00:00:00,-33.0,151.0,10,90,42,io1=1\n"
	if err := os.WriteFile("gps-a.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("gps-b.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}

	loc := time.UTC
	cfg := config.Default()
	if _, err := ImportGPS([]string{"gps-a.csv"}, loc, cfg); err != nil {
		t.Fatalf("import a: %v", err)
	}
	second, err := ImportGPS([]string{"gps-b.csv"}, loc, cfg)
	if err != nil {
		t.Fatalf("import b: %v", err)
	}
	if second.GPSPoints != 0 || second.SourceRows != 1 {
		t.Fatalf("unexpected second source result: %#v", second)
	}

	db, err := Open(DefaultPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	var seenCount int
	if err := db.QueryRow(`SELECT seen_count FROM gps_points`).Scan(&seenCount); err != nil {
		t.Fatal(err)
	}
	if seenCount != 2 {
		t.Fatalf("want seen_count 2, got %d", seenCount)
	}
}

func TestExportGPSFlattensSortedParamsWithoutMetadata(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Init(DefaultPath); err != nil {
		t.Fatal(err)
	}

	csvData := "dt,lat,lng,altitude,angle,speed,params\n" +
		"2026-05-01 00:00:00,-33.0,151.0,10,90,42,zeta=9,alpha=1,io1=on\n"
	if err := os.WriteFile("gps.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := ImportGPS([]string{"gps.csv"}, time.UTC, config.Default()); err != nil {
		t.Fatalf("import gps: %v", err)
	}
	if err := ExportGPS("gps.tsv"); err != nil {
		t.Fatalf("export gps: %v", err)
	}

	f, err := os.Open("gps.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("want header plus one row, got %d", len(rows))
	}

	wantHeader := []string{"Raw DT", "Normalised Time", "Lat", "Lng", "Altitude", "Angle", "Speed KPH", "alpha", "io1", "zeta"}
	if !reflect.DeepEqual(rows[0], wantHeader) {
		t.Fatalf("unexpected header\nwant: %#v\n got: %#v", wantHeader, rows[0])
	}
	for _, forbidden := range []string{"Params Raw", "Params JSON", "First Source File", "Seen Count", "Imported At"} {
		for _, h := range rows[0] {
			if h == forbidden {
				t.Fatalf("metadata/audit column leaked into gps export: %s", forbidden)
			}
		}
	}
	wantRow := []string{"2026-05-01 00:00:00", "2026-05-01T00:00:00Z", "-33", "151", "10", "90", "42", "1", "on", "9"}
	if !reflect.DeepEqual(rows[1], wantRow) {
		t.Fatalf("unexpected row\nwant: %#v\n got: %#v", wantRow, rows[1])
	}
}

func TestExportGPSEmptyDBWritesCoreHeaderOnly(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Init(DefaultPath); err != nil {
		t.Fatal(err)
	}
	if err := ExportGPS("empty.tsv"); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open("empty.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.Comma = '\t'
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("want header only, got %#v", rows)
	}
	if !reflect.DeepEqual(rows[0], gpsExportCoreHeader) {
		t.Fatalf("unexpected empty header: %#v", rows[0])
	}
}
