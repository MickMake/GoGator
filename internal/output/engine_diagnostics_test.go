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

func TestWriteEngineDiagnosticsSelectionIncludesCounts(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "x_engine_selection.csv")
	diag := engine.Diagnostics{EngineSelection: engine.EngineModeSelection{RequestedTripSource: "engine", SelectedTripSource: "legacy", Accepted: false, FallbackUsed: true, Readiness: engine.ShadowReadinessPoorMatch, CandidateCount: 2, OfficialValidCount: 1, OfficialJitterCount: 0, RejectedCount: 1, Reasons: []string{"empty candidate trip output"}}}
	if err := WriteEngineDiagnostics(diag, EngineDiagnosticPaths{Selection: p}, EngineDiagnosticOptions{Enabled: true, OutputSelection: true}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "candidate_count,official_valid_count,official_jitter_count,rejected_candidate_count") {
		t.Fatalf("missing extended header: %s", string(b))
	}
}

func TestWriteEngineDiagnosticsMapMatchHeader(t *testing.T) {
	d := t.TempDir()
	p := filepath.Join(d, "x_engine_mapmatch.csv")
	diag := engine.Diagnostics{MapMatch: []engine.CandidateTripMapMatchDiagnostic{{CandidateTripID: 1, Matched: true}}}
	if err := WriteEngineDiagnostics(diag, EngineDiagnosticPaths{MapMatch: p, Points: filepath.Join(d, "x_engine_points.csv"), Motion: filepath.Join(d, "x_engine_motion.csv"), Stays: filepath.Join(d, "x_engine_stays.csv"), Visits: filepath.Join(d, "x_engine_visits.csv"), Excursions: filepath.Join(d, "x_engine_excursions.csv"), CandidateTrips: filepath.Join(d, "x_engine_candidate_trips.csv"), TripComparison: filepath.Join(d, "x_engine_trip_comparison.csv"), ShadowSummary: filepath.Join(d, "x_engine_shadow_summary.csv"), ShadowMismatches: filepath.Join(d, "x_engine_shadow_mismatches.csv"), Selection: filepath.Join(d, "x_engine_selection.csv"), RouteSignatures: filepath.Join(d, "x_engine_route_signatures.csv"), RouteGroups: filepath.Join(d, "x_engine_route_groups.csv")}, EngineDiagnosticOptions{Enabled: true, OutputDiagnostics: true}); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if !strings.Contains(string(b), "candidate_trip_id,matched,matched_distance_meters") {
		t.Fatalf("missing mapmatch header: %s", string(b))
	}
}
