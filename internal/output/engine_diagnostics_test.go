package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gogator/engine"
)

func TestWriteEngineDiagnosticsDisabledByDefault(t *testing.T) {
	d := t.TempDir()
	paths := EngineDiagnosticPaths{Stays: filepath.Join(d, "x_engine_stays.csv")}
	if err := WriteEngineDiagnostics(engine.Diagnostics{}, paths, EngineDiagnosticOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(paths.Stays); !os.IsNotExist(err) {
		t.Fatalf("expected no file")
	}
}

func TestWriteEngineDiagnosticsStaysHeader(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "x_engine_stays.csv")
	err := WriteEngineDiagnostics(engine.Diagnostics{}, EngineDiagnosticPaths{Stays: p}, EngineDiagnosticOptions{Enabled: true, OutputStays: true})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "stay_id,start_time,end_time") {
		t.Fatalf("missing header: %s", string(b))
	}
}

func TestWriteEngineDiagnosticsCandidateTripsHeader(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "x_engine_candidate_trips.csv")
	err := WriteEngineDiagnostics(engine.Diagnostics{}, EngineDiagnosticPaths{CandidateTrips: p}, EngineDiagnosticOptions{Enabled: true, OutputCandidateTrips: true})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "candidate_trip_id,start_time,end_time") {
		t.Fatalf("missing header")
	}
}
