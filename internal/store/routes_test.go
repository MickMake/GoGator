package store

import (
	"database/sql"
	"encoding/csv"
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
	if _, err := ImportRoutes("routes.csv"); err != nil {
		t.Fatalf("import csv: %v", err)
	}

	tsvData := "From\tTo\tName\tConfidence\tNotes\nB\tA\tBA\tauto\tn2\n"
	_ = os.WriteFile("routes.tsv", []byte(tsvData), 0o644)
	if _, err := ImportRoutes("routes.tsv"); err != nil {
		t.Fatalf("import tsv: %v", err)
	}

	counts, _, _ := Status(DefaultPath)
	if counts.Routes != 2 {
		t.Fatalf("want 2 routes got %d", counts.Routes)
	}
}

func TestExportImportRoutesRoundTrip(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "Home", Lat: -33, Lng: 151, Important: true})
	_ = UpsertSite(SiteRecord{Name: "Shop", Lat: -33.1, Lng: 151.1, Important: true})
	if err := UpsertRoute(RouteRecord{FromSite: "Home", ToSite: "Shop", Name: "Run", Confidence: "manual"}); err != nil {
		t.Fatal(err)
	}
	if err := ExportRoutes("routes.tsv"); err != nil {
		t.Fatal(err)
	}
	if _, err := ImportRoutes("routes.tsv"); err != nil {
		t.Fatal(err)
	}
	counts, _, _ := Status(DefaultPath)
	if counts.Routes != 1 {
		t.Fatalf("want 1 got %d", counts.Routes)
	}
}

func TestExportRoutesPreservesBlankOptionalNumbersAndExplicitZero(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "Home", Lat: -33, Lng: 151, Important: true})
	_ = UpsertSite(SiteRecord{Name: "Shop", Lat: -33.1, Lng: 151.1, Important: true})
	_ = UpsertSite(SiteRecord{Name: "Depot", Lat: -33.2, Lng: 151.2, Important: true})

	input := "From\tTo\tName\tConfidence\tNotes\tExpected Distance Min Km\tExpected Distance Max Km\tExpected Duration Min Min\tExpected Duration Max Min\n" +
		"Home\tShop\tBlank Run\tmanual\tblank numeric fields\t\t\t\t\n" +
		"Shop\tDepot\tZero Run\tmanual\texplicit zero\t0\t0\t0\t0\n"
	_ = os.WriteFile("routes.tsv", []byte(input), 0o644)
	if _, err := ImportRoutes("routes.tsv"); err != nil {
		t.Fatalf("import routes: %v", err)
	}
	if err := ExportRoutes("exported.tsv"); err != nil {
		t.Fatalf("export routes: %v", err)
	}

	f, err := os.Open("exported.tsv")
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
	if len(rows) != 3 {
		t.Fatalf("want header plus 2 rows, got %d", len(rows))
	}

	blank := rows[1]
	if blank[0] != "Home" || blank[1] != "Shop" {
		t.Fatalf("unexpected first route row: %#v", blank)
	}
	for i, v := range blank[5:9] {
		if v != "" {
			t.Fatalf("blank route numeric field %d exported as %q", i, v)
		}
	}

	zero := rows[2]
	if zero[0] != "Shop" || zero[1] != "Depot" {
		t.Fatalf("unexpected second route row: %#v", zero)
	}
	for i, v := range zero[5:9] {
		if v != "0" {
			t.Fatalf("zero route numeric field %d exported as %q", i, v)
		}
	}
}

func TestExportRoutesHandlesNullNumericFields(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "A", Lat: -33, Lng: 151, Important: true})
	_ = UpsertSite(SiteRecord{Name: "B", Lat: -33.1, Lng: 151.1, Important: true})
	if err := UpsertRoute(RouteRecord{FromSite: "A", ToSite: "B", Name: "AB"}); err != nil {
		t.Fatal(err)
	}
	if err := UpsertRoute(RouteRecord{
		FromSite:              "B",
		ToSite:                "A",
		Name:                  "BA",
		ExpectedDistanceMinKM: sql.NullFloat64{Float64: 0, Valid: true},
	}); err != nil {
		t.Fatal(err)
	}
	if err := ExportRoutes("routes.tsv"); err != nil {
		t.Fatal(err)
	}
}

func TestImportRoutesUnknownSitesAndBadNumber(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "A", Lat: -33, Lng: 151, Important: true})

	_ = os.WriteFile("badfrom.tsv", []byte("From\tTo\nX\tA\n"), 0o644)
	_, err := ImportRoutes("badfrom.tsv")
	if err == nil || !strings.Contains(err.Error(), `unknown site "X"`) {
		t.Fatalf("got %v", err)
	}

	_ = os.WriteFile("badto.tsv", []byte("From\tTo\nA\tX\n"), 0o644)
	_, err = ImportRoutes("badto.tsv")
	if err == nil || !strings.Contains(err.Error(), `unknown site "X"`) {
		t.Fatalf("got %v", err)
	}

	_ = os.WriteFile("badnum.tsv", []byte("From\tTo\tExpected Distance Min Km\nA\tA\toops\n"), 0o644)
	_, err = ImportRoutes("badnum.tsv")
	if err == nil || !strings.Contains(err.Error(), "row 2") || !strings.Contains(err.Error(), "Expected Distance Min Km") {
		t.Fatalf("got %v", err)
	}
}

func TestDeleteRouteDirectional(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = Init(DefaultPath)
	_ = UpsertSite(SiteRecord{Name: "A", Lat: -33, Lng: 151, Important: true})
	_ = UpsertSite(SiteRecord{Name: "B", Lat: -33.1, Lng: 151.1, Important: true})
	_ = UpsertRoute(RouteRecord{FromSite: "A", ToSite: "B", Name: "AB"})
	_ = UpsertRoute(RouteRecord{FromSite: "B", ToSite: "A", Name: "BA"})
	if err := DeleteRoute("A", "B"); err != nil {
		t.Fatal(err)
	}
	counts, _, _ := Status(DefaultPath)
	if counts.Routes != 1 {
		t.Fatalf("want 1 got %d", counts.Routes)
	}
}
