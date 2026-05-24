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
	if _, err := runProcessPipeline(nil, nil, nil, cfg); err != nil {
		t.Fatal(err)
	}
	cfg.Engine.TripSource = "shadow"
	if _, err := runProcessPipeline(nil, nil, nil, cfg); err != nil {
		t.Fatal(err)
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
