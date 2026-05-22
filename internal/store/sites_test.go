package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExportThenImportSitesRoundTripTSV(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Init(DefaultPath); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := UpsertSite(SiteRecord{Name: "Home", Address: "1 Main St", Lat: -33.0, Lng: 151.0, RangeM: 100, MinDestinationMinutes: 10, Type: "Home", Important: true, Notes: "note"}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := ExportSites("sites.tsv"); err != nil {
		t.Fatalf("export: %v", err)
	}
	if _, err := ImportSites("sites.tsv"); err != nil {
		t.Fatalf("import: %v", err)
	}

	counts, _, err := Status(DefaultPath)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if counts.Sites != 1 {
		t.Fatalf("expected one site after upsert roundtrip, got %d", counts.Sites)
	}
}

func TestImportSitesDetectsCSVAndTSVDelimiters(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Init(DefaultPath); err != nil {
		t.Fatalf("init: %v", err)
	}

	csvData := "Site,Address,GPS,Range,Min Destination Minutes,Type,Important,Notes\nA,Addr A,\"-33.1,151.1\",100,5,Customer,yes,n1\n"
	if err := os.WriteFile("sites.csv", []byte(csvData), 0o644); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	if _, err := ImportSites("sites.csv"); err != nil {
		t.Fatalf("import csv: %v", err)
	}

	tsvData := "Site\tAddress\tGPS\tRange\tMin Destination Minutes\tType\tImportant\tNotes\nB\tAddr B\t-33.2,151.2\t200\t6\tSupplier\tno\tn2\n"
	if err := os.WriteFile("sites.tsv", []byte(tsvData), 0o644); err != nil {
		t.Fatalf("write tsv: %v", err)
	}
	if _, err := ImportSites("sites.tsv"); err != nil {
		t.Fatalf("import tsv: %v", err)
	}

	counts, _, err := Status(DefaultPath)
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if counts.Sites != 2 {
		t.Fatalf("expected 2 sites, got %d", counts.Sites)
	}
}

func TestImportSitesMalformedRowFailsFast(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := Init(DefaultPath); err != nil {
		t.Fatalf("init: %v", err)
	}
	bad := strings.Join([]string{
		"Site\tAddress\tGPS\tRange\tMin Destination Minutes\tType\tImportant\tNotes",
		"Bad\tAddr\tnot-a-gps\t100\t5\tType\tyes\tn",
	}, "\n")
	p := filepath.Join(".", "bad.tsv")
	if err := os.WriteFile(p, []byte(bad), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	_, err := ImportSites(p)
	if err == nil || !strings.Contains(err.Error(), "row 2: invalid gps") {
		t.Fatalf("got %v", err)
	}
}
