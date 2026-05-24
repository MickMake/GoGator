package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestWriteEngineDiagnosticsMotionWritesSpeedKPH(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "x_engine_motion.csv")
	diag := engine.Diagnostics{Motion: []engine.MotionSample{{Index: 1, Time: time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC), State: engine.MotionMoving, SpeedKPH: 42.25, Reason: engine.MotionReasonSpeedMoving, Quality: engine.QualityGood}}}
	if err := WriteEngineDiagnostics(diag, EngineDiagnosticPaths{Motion: p}, EngineDiagnosticOptions{Enabled: true, OutputMotion: true}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), ",42.25,") {
		t.Fatalf("expected speed_kmh value in motion diagnostics: %s", string(b))
	}
}
