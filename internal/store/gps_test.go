package store

import (
	"os"
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
