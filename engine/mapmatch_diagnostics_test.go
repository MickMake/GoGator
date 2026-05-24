package engine

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"gogator/engine/mapmatch"
	"gogator/internal/config"
	"gogator/internal/gps"
)

func TestBuildMapMatchDiagnosticsSkipsInsufficientPoints(t *testing.T) {
	d := buildMapMatchDiagnostics(mapmatch.NoopMapMatcher{}, []gps.RawPoint{{Lat: 1, Lng: 1, Time: time.Now()}}, []CandidateTrip{{SourcePointStart: 0, SourcePointEnd: 0}}, ValhallaConfig{})
	if len(d) != 1 || d[0].Error != "insufficient_points" {
		t.Fatalf("expected insufficient_points, got %+v", d)
	}
}

func TestMapMatchDiagnosticsValhallaDisabledMakesNoHTTPCalls(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cfg := config.Default()
	cfg.Valhalla.Enabled = false
	cfg.Valhalla.BaseURL = ts.URL
	matcher := newMapMatcher(EngineConfig{Valhalla: ValhallaConfig{Enabled: cfg.Valhalla.Enabled, BaseURL: cfg.Valhalla.BaseURL}})
	res := buildMapMatchDiagnostics(matcher, sampleTripPoints(), []CandidateTrip{{SourcePointStart: 0, SourcePointEnd: 3}}, ValhallaConfig{})
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("expected zero HTTP calls, got %d", calls)
	}
	if len(res) == 0 {
		t.Fatalf("expected passive mapmatch diagnostics rows")
	}
}

func TestMapMatchDiagnosticsValhallaEnabledCreatesDiagnostics(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"trip":{"summary":{"length":1.5,"time":120},"legs":[{"maneuvers":[{"street_names":["Main St"]}]}]}}`))
	}))
	defer ts.Close()

	cfg := config.Default()
	cfg.Valhalla.Enabled = true
	cfg.Valhalla.BaseURL = ts.URL
	matcher := newMapMatcher(EngineConfig{Valhalla: ValhallaConfig{Enabled: cfg.Valhalla.Enabled, BaseURL: cfg.Valhalla.BaseURL}})
	res := buildMapMatchDiagnostics(matcher, sampleTripPoints(), []CandidateTrip{{SourcePointStart: 0, SourcePointEnd: 3}}, ValhallaConfig{})
	if atomic.LoadInt32(&calls) == 0 {
		t.Fatalf("expected HTTP calls")
	}
	found := false
	for _, row := range res {
		if row.Matched {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected at least one matched diagnostic row: %+v", res)
	}
}

func sampleTripPoints() []gps.RawPoint {
	start := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	return []gps.RawPoint{
		{Time: start, Lat: -33.000, Lng: 151.000, SpeedKPH: 0, Params: map[string]string{"io24": "0", "io251": "1"}, ParamNums: map[string]float64{"io24": 0, "io251": 1}},
		{Time: start.Add(1 * time.Minute), Lat: -33.000, Lng: 151.000, SpeedKPH: 0, Params: map[string]string{"io24": "0", "io251": "1"}, ParamNums: map[string]float64{"io24": 0, "io251": 1}},
		{Time: start.Add(2 * time.Minute), Lat: -33.001, Lng: 151.001, SpeedKPH: 25, Params: map[string]string{"io24": "1"}, ParamNums: map[string]float64{"io24": 1}},
		{Time: start.Add(3 * time.Minute), Lat: -33.002, Lng: 151.002, SpeedKPH: 30, Params: map[string]string{"io24": "1"}, ParamNums: map[string]float64{"io24": 1}},
		{Time: start.Add(4 * time.Minute), Lat: -33.003, Lng: 151.003, SpeedKPH: 0, Params: map[string]string{"io24": "0", "io251": "1"}, ParamNums: map[string]float64{"io24": 0, "io251": 1}},
		{Time: start.Add(5 * time.Minute), Lat: -33.003, Lng: 151.003, SpeedKPH: 0, Params: map[string]string{"io24": "0", "io251": "1"}, ParamNums: map[string]float64{"io24": 0, "io251": 1}},
	}
}
