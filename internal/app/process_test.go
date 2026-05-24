package app

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gogator/engine"
	"gogator/internal/config"
	"gogator/internal/gps"
)

func TestRunProcessPipelinePropagatesEngineError(t *testing.T) {
	orig := runEngine
	t.Cleanup(func() { runEngine = orig })

	runEngine = func(_ context.Context, _ engine.Input) (engine.Result, error) {
		return engine.Result{}, errors.New("boom")
	}

	_, err := runProcessPipeline(nil, nil, nil, config.Default())
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "run engine pipeline") {
		t.Fatalf("expected wrapped engine error, got %v", err)
	}
}

func TestResolveTripSourceDefaultsToLegacy(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Engine.TripSource = ""
	got, err := resolveTripSource(cfg)
	if err != nil || got != "legacy" {
		t.Fatalf("expected legacy default, got %q err=%v", got, err)
	}
}

func TestResolveTripSourceInvalid(t *testing.T) {
	t.Parallel()
	cfg := config.Default()
	cfg.Engine.TripSource = "bad"
	_, err := resolveTripSource(cfg)
	if err == nil || !strings.Contains(err.Error(), "legacy, shadow, engine") {
		t.Fatalf("expected clear invalid trip source error, got %v", err)
	}
}

func TestEngineTripSourceUsesCandidateTrips(t *testing.T) {
	orig := runEngine
	t.Cleanup(func() { runEngine = orig })
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	runEngine = func(_ context.Context, _ engine.Input) (engine.Result, error) {
		return engine.Result{
			Points:         []gps.RawPoint{{RawRow: 10, SourceFile: "a.csv", Lat: -33, Lng: 151}, {RawRow: 20, SourceFile: "a.csv", Lat: -34, Lng: 150}},
			Valid:          []gps.Trip{{DepartureSite: "legacy"}},
			CandidateTrips: engine.CandidateTripEvidence{Trips: []engine.CandidateTrip{{StartTime: base, EndTime: base.Add(10 * time.Minute), OriginLabel: "A", DestinationLabel: "B", SourcePointStart: 0, SourcePointEnd: 1}}},
			TripComparison: engine.CandidateTripComparison{ShadowSummary: engine.ShadowSummary{Readiness: engine.ShadowReadinessGoodMatch}},
		}, nil
	}
	cfg := config.Default()
	cfg.Engine.TripSource = "engine"
	res, err := runProcessPipeline(nil, nil, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Valid) != 1 || res.Valid[0].DepartureSite != "A" || res.Valid[0].RawStartRow != 10 {
		t.Fatalf("expected adapted candidate trip, got %+v", res.Valid)
	}
}

func TestLegacyAndShadowBypassEngineSelectionValidation(t *testing.T) {
	orig := runEngine
	t.Cleanup(func() { runEngine = orig })
	runEngine = func(_ context.Context, _ engine.Input) (engine.Result, error) {
		return engine.Result{Valid: []gps.Trip{{DepartureSite: "legacy"}}}, nil
	}
	cfg := config.Default()
	cfg.Engine.TripSource = "legacy"
	legacyRes, err := runProcessPipeline(nil, nil, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if legacyRes.EngineDiagnostics.EngineSelection.RequestedTripSource != "legacy" || legacyRes.EngineDiagnostics.EngineSelection.SelectedTripSource != "legacy" {
		t.Fatalf("expected legacy selection/requested, got %+v", legacyRes.EngineDiagnostics.EngineSelection)
	}
	cfg.Engine.TripSource = "shadow"
	shadowRes, err := runProcessPipeline(nil, nil, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if shadowRes.EngineDiagnostics.EngineSelection.RequestedTripSource != "shadow" || shadowRes.EngineDiagnostics.EngineSelection.SelectedTripSource != "legacy" {
		t.Fatalf("expected shadow requested with legacy selected, got %+v", shadowRes.EngineDiagnostics.EngineSelection)
	}
}

func TestEngineRejectWithoutFallback(t *testing.T) {
	orig := runEngine
	t.Cleanup(func() { runEngine = orig })
	runEngine = func(_ context.Context, _ engine.Input) (engine.Result, error) {
		return engine.Result{TripComparison: engine.CandidateTripComparison{ShadowSummary: engine.ShadowSummary{Readiness: engine.ShadowReadinessPoorMatch}}}, nil
	}
	cfg := config.Default()
	cfg.Engine.TripSource = "engine"
	_, err := runProcessPipeline(nil, nil, nil, cfg)
	if err == nil || !strings.Contains(err.Error(), "engine trip output rejected") {
		t.Fatalf("expected rejection error, got %v", err)
	}
}

func TestAdaptCandidateTripsUnknownEndpointsNoPanic(t *testing.T) {
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	valid, jitter := adaptCandidateTrips([]engine.CandidateTrip{{StartTime: base, EndTime: base.Add(5 * time.Minute), SourcePointStart: -1, SourcePointEnd: 999, Confidence: engine.CandidateConfidenceHigh}}, nil)
	if len(valid) != 1 || len(jitter) != 0 {
		t.Fatalf("expected valid conversion, got valid=%d jitter=%d", len(valid), len(jitter))
	}
}

func TestAdaptCandidateTripsLowGapNoiseBecomeJitter(t *testing.T) {
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	cases := []engine.CandidateTrip{{StartTime: base, EndTime: base.Add(5 * time.Minute), Confidence: engine.CandidateConfidenceLow}, {StartTime: base, EndTime: base.Add(5 * time.Minute), Confidence: engine.CandidateConfidenceMedium, Type: engine.CandidateTripGapAffected}, {StartTime: base, EndTime: base.Add(5 * time.Minute), Confidence: engine.CandidateConfidenceHigh, Type: engine.CandidateTripNoiseAffected}}
	valid, jitter := adaptCandidateTrips(cases, nil)
	if len(valid) != 0 || len(jitter) != 3 {
		t.Fatalf("expected all jitter, got valid=%d jitter=%d", len(valid), len(jitter))
	}
}

func TestEngineFallbackSelectionDiagnostics(t *testing.T) {
	orig := runEngine
	t.Cleanup(func() { runEngine = orig })
	runEngine = func(_ context.Context, _ engine.Input) (engine.Result, error) {
		return engine.Result{CandidateTrips: engine.CandidateTripEvidence{Trips: []engine.CandidateTrip{}}, TripComparison: engine.CandidateTripComparison{ShadowSummary: engine.ShadowSummary{Readiness: engine.ShadowReadinessPoorMatch}}, Valid: []gps.Trip{{DepartureSite: "legacy"}}}, nil
	}
	cfg := config.Default()
	cfg.Engine.TripSource = "engine"
	cfg.Engine.EngineMode.FallbackToLegacyOnReject = true
	res, err := runProcessPipeline(nil, nil, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !res.EngineDiagnostics.EngineSelection.FallbackUsed || res.EngineDiagnostics.EngineSelection.SelectedTripSource != "legacy" {
		t.Fatalf("expected legacy fallback selection, got %+v", res.EngineDiagnostics.EngineSelection)
	}
}

func TestEngineTripSourceMapMatchErrorDoesNotLeakIntoWarnings(t *testing.T) {
	orig := runEngine
	t.Cleanup(func() { runEngine = orig })
	base := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	runEngine = func(_ context.Context, _ engine.Input) (engine.Result, error) {
		return engine.Result{
			Points: []gps.RawPoint{
				{RawRow: 10, SourceFile: "a.csv", Lat: -33, Lng: 151},
				{RawRow: 20, SourceFile: "a.csv", Lat: -34, Lng: 150},
			},
			CandidateTrips: engine.CandidateTripEvidence{Trips: []engine.CandidateTrip{
				{
					StartTime:        base,
					EndTime:          base.Add(10 * time.Minute),
					OriginLabel:      "A",
					DestinationLabel: "B",
					SourcePointStart: 0,
					SourcePointEnd:   1,
				},
			}},
			MapMatchDiagnostics: []engine.CandidateTripMapMatchDiagnostic{
				{CandidateTripID: 1, Error: "valhalla_unavailable"},
			},
			TripComparison: engine.CandidateTripComparison{ShadowSummary: engine.ShadowSummary{Readiness: engine.ShadowReadinessGoodMatch}},
		}, nil
	}
	cfg := config.Default()
	cfg.Engine.TripSource = "engine"
	res, err := runProcessPipeline(nil, nil, nil, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Valid) != 1 {
		t.Fatalf("expected one adapted trip, got %d", len(res.Valid))
	}
	for _, f := range res.Valid[0].Flags {
		if f == "engine_warning:mapmatch_error" {
			t.Fatalf("did not expect map-match diagnostic leakage into adapted trip flags: %+v", res.Valid[0].Flags)
		}
	}
	if len(res.EngineDiagnostics.MapMatch) != 1 || res.EngineDiagnostics.MapMatch[0].Error == "" {
		t.Fatalf("expected map-match error to remain in diagnostics, got %+v", res.EngineDiagnostics.MapMatch)
	}
}
