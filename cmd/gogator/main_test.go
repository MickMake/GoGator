package main

import (
	"os"
	"strings"
	"testing"
)

func TestAddRouteRejected(t *testing.T) {
	err := run([]string{"add_route"})
	if err == nil || !strings.Contains(err.Error(), "add_route has been replaced") {
		t.Fatalf("got %v", err)
	}
}

func TestAddRouteManualRecognized(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = run([]string{"db", "init"})
	_ = run([]string{"add", "site", "name", "A", "gps", "-33.0,151.0"})
	_ = run([]string{"add", "site", "name", "B", "gps", "-33.1,151.1"})
	err := run([]string{"add", "route", "from", "A", "to", "B"})
	if err != nil {
		t.Fatalf("got %v", err)
	}
}

func TestAddRoutePromotionPathStillUsed(t *testing.T) {
	err := run([]string{"add", "route", "1", "from", "missing.tsv"})
	if err == nil || !strings.Contains(err.Error(), "open missing.tsv") {
		t.Fatalf("got %v", err)
	}
}

func TestAddSiteRecognized(t *testing.T) {
	t.Chdir(t.TempDir())

	err := run([]string{"add", "site", "name", "X", "gps", "-33.0,151.0"})
	if err == nil || !strings.Contains(err.Error(), "no such table: sites") {
		t.Fatalf("got %v", err)
	}
}

func TestImportGPSRecognized(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = run([]string{"db", "init"})
	csvData := "dt,lat,lng,altitude,angle,speed,params\n2026-05-01 00:00:00,-33.0,151.0,10,90,42,io1=1\n"
	if err := os.WriteFile("gps.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"import", "gps", "from", "gps.csv"}); err != nil {
		t.Fatalf("import gps from: %v", err)
	}
	if err := run([]string{"import", "gps", "gps.csv"}); err != nil {
		t.Fatalf("import gps direct: %v", err)
	}
}

func TestExportRecognized(t *testing.T) {
	err := run([]string{"export", "trips", "during", "2026", "as", "trips.tsv"})
	if err == nil || !strings.Contains(err.Error(), "not implemented yet: export trips") {
		t.Fatalf("got %v", err)
	}
}

func TestExportPathsRecognized(t *testing.T) {
	err := run([]string{"export", "paths", "from", "A", "to", "B", "during", "2026"})
	if err == nil || !strings.Contains(err.Error(), "not implemented yet: export paths") {
		t.Fatalf("got %v", err)
	}
}

func TestProcessGPSRecognized(t *testing.T) {
	err := run([]string{"process", "gps", "during", "2026"})
	if err == nil || !strings.Contains(err.Error(), "not implemented yet: process gps") {
		t.Fatalf("got %v", err)
	}
}

func TestDBStatusMissingDB(t *testing.T) {
	t.Chdir(t.TempDir())

	err := run([]string{"db", "status"})
	if err == nil || !strings.Contains(err.Error(), "database not found") {
		t.Fatalf("got %v", err)
	}
}
