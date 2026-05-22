package main

import (
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
	err := run([]string{"add", "route", "from", "A", "to", "B"})
	if err == nil || !strings.Contains(err.Error(), "not implemented yet: add route") {
		t.Fatalf("got %v", err)
	}
}

func TestAddSiteRecognized(t *testing.T) {
	err := run([]string{"add", "site", "name", "X", "gps", "-33.0,151.0"})
	if err == nil {
		t.Fatalf("expected db-related error")
	}
}

func TestImportRecognized(t *testing.T) {
	if err := run([]string{"import", "gps", "from", "file.csv"}); err == nil || !strings.Contains(err.Error(), "not implemented yet: import gps") {
		t.Fatalf("got %v", err)
	}
	if err := run([]string{"import", "gps", "file.csv"}); err == nil || !strings.Contains(err.Error(), "not implemented yet: import gps") {
		t.Fatalf("got %v", err)
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
	err := run([]string{"db", "status"})
	if err == nil || !(strings.Contains(err.Error(), "database not found") || strings.Contains(err.Error(), "no such table")) {
		t.Fatalf("got %v", err)
	}
}
