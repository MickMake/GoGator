package engine

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/routes"
	"gogator/internal/sites"
)

type ValidationMetrics struct {
	LegacyValidTripCount          int
	LegacyJitterTripCount         int
	EngineCandidateTripCount      int
	EngineOfficialValidTripCount  int
	EngineOfficialJitterTripCount int
	ShadowReadiness               ShadowReadiness
	UnmatchedLegacyValidCount     int
	UnmatchedCandidateCount       int
	LargestBoundaryDeltaMinutes   float64
	LowConfidenceCandidateCount   int
	GapNoiseAffectedCount         int
	RouteSignatureGroupCount      int
	ValhallaDiagnosticCount       int
	EngineModeAccepted            bool
	EngineModeReasons             []string
}

func buildValidationMetricsFromResult(res Result) ValidationMetrics {
	s := res.TripComparison.ShadowSummary
	return ValidationMetrics{
		LegacyValidTripCount:          s.LegacyValidTripCount,
		LegacyJitterTripCount:         s.LegacyJitterTripCount,
		EngineCandidateTripCount:      len(res.CandidateTrips.Trips),
		EngineOfficialValidTripCount:  len(res.Valid),
		EngineOfficialJitterTripCount: len(res.Jitter),
		ShadowReadiness:               s.Readiness,
		UnmatchedLegacyValidCount:     s.UnmatchedLegacyValidTripCount,
		UnmatchedCandidateCount:       s.UnmatchedCandidateTripCount,
		LargestBoundaryDeltaMinutes:   s.LargestBoundaryDeltaMinutes,
		LowConfidenceCandidateCount:   s.LowConfidenceCandidateCount,
		GapNoiseAffectedCount:         s.GapAffectedCandidateCount + s.NoiseAffectedCandidateCount,
		RouteSignatureGroupCount:      len(res.RouteGroups),
		ValhallaDiagnosticCount:       len(res.MapMatchDiagnostics),
	}
}

func TestValidationPackLegacyShadowAndEngineModes(t *testing.T) {
	siteList := []sites.Site{
		{Name: "Home", Address: "1 Home St", Lat: -33.00000, Lng: 151.00000, RadiusM: 120, MinDestinationMinutes: 1, Important: true},
		{Name: "Work", Address: "99 Work Rd", Lat: -33.01000, Lng: 151.01000, RadiusM: 120, MinDestinationMinutes: 1, Important: true},
	}
	rules := []routes.Route{{Name: "Home to Work", FromSite: "Home", ToSite: "Work", DistanceMinKM: 0.2, DistanceMaxKM: 5, DurationMinMin: 1, DurationMaxMin: 30, ConfidenceBoost: "Good"}}
	start := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{
		mkPoint("v.csv", 2, start, -33.0000, 151.0000, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("v.csv", 3, start.Add(1*time.Minute), -33.0000, 151.0000, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("v.csv", 4, start.Add(2*time.Minute), -33.0006, 151.0006, 28, map[string]float64{"io24": 1}),
		mkPoint("v.csv", 5, start.Add(3*time.Minute), -33.0015, 151.0015, 35, map[string]float64{"io24": 1}),
		mkPoint("v.csv", 6, start.Add(4*time.Minute), -33.0100, 151.0100, 0, map[string]float64{"io24": 0, "io251": 1}),
	}

	legacyCfg := config.Default()
	legacyRes, err := Run(context.Background(), Input{Points: clonePoints(pts), Sites: siteList, Routes: rules, Config: legacyCfg})
	if err != nil {
		t.Fatalf("legacy Run error: %v", err)
	}
	legacyMetrics := buildValidationMetricsFromResult(legacyRes)
	if legacyMetrics.EngineOfficialValidTripCount == 0 {
		t.Fatalf("expected non-zero official valid count in legacy run: %+v", legacyMetrics)
	}
	if legacyMetrics.ShadowReadiness != "" && legacyMetrics.ShadowReadiness != ShadowReadinessNotEvaluated {
		t.Fatalf("expected legacy run to skip shadow readiness, got %s", legacyMetrics.ShadowReadiness)
	}

	shadowCfg := config.Default()
	shadowCfg.Engine.TripBuilder.Enabled = true
	shadowCfg.Engine.StayDetection.Enabled = true
	shadowCfg.Engine.Visits.Enabled = true
	shadowCfg.Engine.Excursions.Enabled = true
	shadowCfg.Engine.TripBuilder.CompareLegacy = true
	shadowCfg.Engine.Shadow.Enabled = true
	shadowCfg.Engine.Shadow.SummaryEnabled = true
	shadowCfg.Engine.RouteSignatures.Enabled = true
	shadowCfg.Engine.RouteGrouping.Enabled = true

	shadowRes, err := Run(context.Background(), Input{Points: clonePoints(pts), Sites: siteList, Routes: rules, Config: shadowCfg})
	if err != nil {
		t.Fatalf("shadow Run error: %v", err)
	}
	shadowMetrics := buildValidationMetricsFromResult(shadowRes)
	if shadowMetrics.ShadowReadiness == ShadowReadinessNotEvaluated {
		t.Fatalf("expected shadow readiness computed, got %s", shadowMetrics.ShadowReadiness)
	}
	if shadowMetrics.ValhallaDiagnosticCount != 0 {
		t.Fatalf("expected no valhalla diagnostics when disabled, got %d", shadowMetrics.ValhallaDiagnosticCount)
	}
	if shadowMetrics.LegacyValidTripCount != legacyMetrics.EngineOfficialValidTripCount {
		t.Fatalf("expected shadow legacy baseline count to match legacy official valid count, legacy=%d shadow=%d", legacyMetrics.EngineOfficialValidTripCount, shadowMetrics.LegacyValidTripCount)
	}

	policy := EngineModePolicy{RequireMinReadiness: true, MinReadiness: ShadowReadinessGoodMatch, AllowLowConfidence: false, AllowGapAffected: false, AllowEmptyCandidates: false, FallbackToLegacyOnReject: false, MaxUnmatchedLegacyPercent: 20, MaxBoundaryDeltaMinutes: 20, RejectNoiseAffected: true}
	sel, err := ValidateEngineMode(shadowRes.CandidateTrips, shadowRes.TripComparison.ShadowSummary, policy)
	if err != nil {
		t.Fatalf("ValidateEngineMode error: %v", err)
	}
	shadowMetrics.EngineModeAccepted = sel.Accepted
	shadowMetrics.EngineModeReasons = append([]string(nil), sel.Reasons...)
	if len(shadowRes.CandidateTrips.Trips) == 0 && sel.Accepted {
		t.Fatalf("expected empty-candidate fixture to reject engine mode")
	}
	if len(shadowRes.CandidateTrips.Trips) == 0 && len(sel.Reasons) == 0 {
		t.Fatalf("expected clear rejection reasons for empty candidate fixture")
	}
	if shadowMetrics.ShadowReadiness == "" || shadowMetrics.ShadowReadiness == ShadowReadinessNotEvaluated {
		t.Fatalf("expected shadow run to produce a non-empty readiness state, got %q", shadowMetrics.ShadowReadiness)
	}
	if legacyMetrics.ShadowReadiness == shadowMetrics.ShadowReadiness {
		t.Fatalf("expected distinct readiness states between legacy and shadow runs: legacy=%q shadow=%q", legacyMetrics.ShadowReadiness, shadowMetrics.ShadowReadiness)
	}
}

func TestValidationMetricsDeterministic(t *testing.T) {
	r := Result{CandidateTrips: CandidateTripEvidence{Trips: []CandidateTrip{{}}}, TripComparison: CandidateTripComparison{ShadowSummary: ShadowSummary{LegacyValidTripCount: 1, LegacyJitterTripCount: 2, CandidateTripCount: 1, Readiness: ShadowReadinessGoodMatch}}}
	a := buildValidationMetricsFromResult(r)
	b := buildValidationMetricsFromResult(r)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("non-deterministic validation metrics: a=%+v b=%+v", a, b)
	}
}
