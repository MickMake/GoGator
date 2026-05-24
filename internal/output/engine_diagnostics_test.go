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

func TestWriteEngineDiagnosticsShadowSummaryHeader(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "x_engine_shadow_summary.csv")
	err := WriteEngineDiagnostics(engine.Diagnostics{ShadowSummary: engine.ShadowSummary{Metrics: []engine.ShadowMetric{{Name: "x", Value: "1"}}}}, EngineDiagnosticPaths{ShadowSummary: p}, EngineDiagnosticOptions{Enabled: true, OutputShadowSummary: true})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "metric,value,severity,notes") {
		t.Fatalf("missing header")
	}
}
