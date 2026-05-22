package store

import (
	"os"
	"strings"
	"testing"
)

func TestImportRoutesCSVAndTSVAndDirectionalIdentity(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "A", Lat: -33, Lng: 151, Important: true})
	_ = UpsertSite(SiteRecord{Name: "B", Lat: -33.1, Lng: 151.1, Important: true})

	csvData := "From,To,Name,Confidence,Notes\nA,B,AB,manual,n1\n"
	_ = os.WriteFile("routes.csv", []byte(csvData), 0o644)
	if _, err := ImportRoutes("routes.csv"); err != nil { t.Fatalf("import csv: %v", err) }

	tsvData := "From\tTo\tName\tConfidence\tNotes\nB\tA\tBA\tauto\tn2\n"
	_ = os.WriteFile("routes.tsv", []byte(tsvData), 0o644)
	if _, err := ImportRoutes("routes.tsv"); err != nil { t.Fatalf("import tsv: %v", err) }

	counts, _, _ := Status(DefaultPath)
	if counts.Routes != 2 { t.Fatalf("want 2 routes got %d", counts.Routes) }
}

func TestExportImportRoutesRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "Home", Lat: -33, Lng: 151, Important: true})
	_ = UpsertSite(SiteRecord{Name: "Shop", Lat: -33.1, Lng: 151.1, Important: true})
	if err := UpsertRoute(RouteRecord{FromSite: "Home", ToSite: "Shop", Name: "Run", Confidence: "manual"}); err != nil { t.Fatal(err) }
	if err := ExportRoutes("routes.tsv"); err != nil { t.Fatal(err) }
	if _, err := ImportRoutes("routes.tsv"); err != nil { t.Fatal(err) }
	counts, _, _ := Status(DefaultPath)
	if counts.Routes != 1 { t.Fatalf("want 1 got %d", counts.Routes) }
}

func TestImportRoutesUnknownSitesAndBadNumber(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "A", Lat: -33, Lng: 151, Important: true})

	_ = os.WriteFile("badfrom.tsv", []byte("From\tTo\nX\tA\n"), 0o644)
	_, err := ImportRoutes("badfrom.tsv")
	if err == nil || !strings.Contains(err.Error(), `unknown site "X"`) { t.Fatalf("got %v", err) }

	_ = os.WriteFile("badto.tsv", []byte("From\tTo\nA\tX\n"), 0o644)
	_, err = ImportRoutes("badto.tsv")
	if err == nil || !strings.Contains(err.Error(), `unknown site "X"`) { t.Fatalf("got %v", err) }

	_ = os.WriteFile("badnum.tsv", []byte("From\tTo\tExpected Distance Min Km\nA\tA\toops\n"), 0o644)
	_, err = ImportRoutes("badnum.tsv")
	if err == nil || !strings.Contains(err.Error(), "row 2") || !strings.Contains(err.Error(), "Expected Distance Min Km") { t.Fatalf("got %v", err) }
}

func TestDeleteRouteDirectional(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "A", Lat: -33, Lng: 151, Important: true})
	_ = UpsertSite(SiteRecord{Name: "B", Lat: -33.1, Lng: 151.1, Important: true})
	_ = UpsertRoute(RouteRecord{FromSite: "A", ToSite: "B", Name: "AB"})
	_ = UpsertRoute(RouteRecord{FromSite: "B", ToSite: "A", Name: "BA"})
	if err := DeleteRoute("A", "B"); err != nil { t.Fatal(err) }
	counts, _, _ := Status(DefaultPath)
	if counts.Routes != 1 { t.Fatalf("want 1 got %d", counts.Routes) }
}
