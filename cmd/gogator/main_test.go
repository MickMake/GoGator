package main

import (
	"os"
	"strings"
	"testing"
)

func TestLegacyAddRouteUnknown(t *testing.T) {
	err := run([]string{"add_route"})
	if err == nil || !strings.Contains(err.Error(), "unknown command: add_route") {
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

func TestGPSParamsSettingsStubsRecognized(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"set", "gps", "params", "io66,io67"}, "not implemented yet: set gps params"},
		{[]string{"show", "gps", "params"}, "not implemented yet: show gps params"},
		{[]string{"reset", "gps", "params"}, "not implemented yet: reset gps params"},
	} {
		err := run(tc.args)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Fatalf("%v got %v", tc.args, err)
		}
	}
}

func TestImportRawRecognized(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = run([]string{"db", "init"})
	csvData := "dt,lat,lng,altitude,angle,speed,params\n2026-05-01 00:00:00,-33.0,151.0,10,90,42,io1=1\n"
	if err := os.WriteFile("raw.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"import", "raw", "from", "raw.csv"}); err != nil {
		t.Fatalf("import raw from: %v", err)
	}
	if err := run([]string{"import", "raw", "raw.csv"}); err != nil {
		t.Fatalf("import raw direct: %v", err)
	}
}

func TestImportGPSRejected(t *testing.T) {
	err := run([]string{"import", "gps", "from", "raw.csv"})
	if err == nil || !strings.Contains(err.Error(), "import gps is not a command; use import raw") {
		t.Fatalf("got %v", err)
	}
}

func TestExportGPSRecognized(t *testing.T) {
	t.Chdir(t.TempDir())
	_ = run([]string{"db", "init"})
	csvData := "dt,lat,lng,altitude,angle,speed,params\n" + `2026-05-01 00:00:00,-33.0,151.0,10,90,42,"zeta=9,alpha=1"` + "\n"
	if err := os.WriteFile("raw.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"import", "raw", "from", "raw.csv"}); err != nil {
		t.Fatalf("import raw: %v", err)
	}
	if err := run([]string{"export", "gps", "as", "gps.tsv"}); err != nil {
		t.Fatalf("export gps: %v", err)
	}
	data, err := os.ReadFile("gps.tsv")
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"Raw DT\tNormalised Time\tLat\tLng\tAltitude\tAngle\tSpeed KPH\tgpslev\tgsmlev\tpdop\tio1", "\talpha\tzeta", "2026-05-01 00:00:00", "\t1\t9"} {
		if !strings.Contains(text, want) {
			t.Fatalf("export missing %q in:\n%s", want, text)
		}
	}
}

func TestExportRawRecognized(t *testing.T) {
	t.Chdir(t.TempDir())

	_ = run([]string{"db", "init"})

	csvData := "dt,lat,lng,altitude,angle,speed,params\n" +
		`2026-05-01 00:00:00,-33.000000,151.000000,10,90,42,"zeta=9,alpha=1"` + "\n"

	if err := os.WriteFile("raw.csv", []byte(csvData), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := run([]string{"import", "raw", "from", "raw.csv"}); err != nil {
		t.Fatalf("import raw: %v", err)
	}

	if err := run([]string{"export", "raw", "as", "exported-raw.csv"}); err != nil {
		t.Fatalf("export raw: %v", err)
	}

	data, err := os.ReadFile("exported-raw.csv")
	if err != nil {
		t.Fatal(err)
	}

	text := string(data)

	for _, want := range []string{
		"dt,lat,lng,altitude,angle,speed,params",
		`2026-05-01 00:00:00,-33,151,10,90,42,"zeta=9,alpha=1"`,
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("export raw missing %q in:\n%s", want, text)
		}
	}
}

func TestDBBackupAndVacuumRecognized(t *testing.T) {
	t.Chdir(t.TempDir())
	if err := run([]string{"db", "init"}); err != nil {
		t.Fatalf("db init: %v", err)
	}
	if err := run([]string{"db", "backup", "as", "backup.sqlite"}); err != nil {
		t.Fatalf("db backup: %v", err)
	}
	if _, err := os.Stat("backup.sqlite"); err != nil {
		t.Fatalf("backup missing: %v", err)
	}
	if err := run([]string{"db", "vacuum"}); err != nil {
		t.Fatalf("db vacuum: %v", err)
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
