package engine

import (
	"reflect"
	"testing"
	"time"

	"gogator/internal/gps"
)

func TestShadowSummaryExcellentMatch(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	cand := CandidateTripEvidence{Trips: []CandidateTrip{{StartTime: base, EndTime: base.Add(10 * time.Minute)}}}
	legacy := []gps.Trip{{Start: base, End: base.Add(10 * time.Minute)}}
	s := buildShadowSummary(cand, legacy, nil, ShadowConfig{Enabled: true, SummaryEnabled: true})
	if s.Readiness != ShadowReadinessExcellent {
		t.Fatalf("got %s", s.Readiness)
	}
}
func TestShadowSummaryUnmatchedTripsReported(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	s := buildShadowSummary(CandidateTripEvidence{Trips: []CandidateTrip{{StartTime: base, EndTime: base.Add(time.Minute)}}}, []gps.Trip{{Start: base.Add(2 * time.Hour), End: base.Add(3 * time.Hour)}}, nil, ShadowConfig{Enabled: true, SummaryEnabled: true})
	if s.UnmatchedLegacyTripCount == 0 || s.UnmatchedCandidateTripCount == 0 {
		t.Fatalf("expected unmatched")
	}
}
func TestShadowSummaryDeterministic(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	cand := CandidateTripEvidence{Trips: []CandidateTrip{{StartTime: base, EndTime: base.Add(10 * time.Minute)}, {StartTime: base.Add(1 * time.Hour), EndTime: base.Add(70 * time.Minute)}}}
	legacy := []gps.Trip{{Start: base, End: base.Add(10 * time.Minute)}, {Start: base.Add(1 * time.Hour), End: base.Add(70 * time.Minute)}}
	a := buildShadowSummary(cand, legacy, nil, ShadowConfig{Enabled: true, SummaryEnabled: true})
	b := buildShadowSummary(cand, legacy, nil, ShadowConfig{Enabled: true, SummaryEnabled: true})
	if !reflect.DeepEqual(a, b) {
		t.Fatal("non deterministic")
	}
}

func TestShadowSummaryPreservesBaselineComparisonWhenSummaryDisabled(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	cand := CandidateTripEvidence{Trips: []CandidateTrip{
		{StartTime: base, EndTime: base.Add(10 * time.Minute)},
		{StartTime: base.Add(3 * time.Hour), EndTime: base.Add(4 * time.Hour)},
	}}
	legacy := []gps.Trip{
		{Start: base.Add(2 * time.Minute), End: base.Add(12 * time.Minute)},
		{Start: base.Add(6 * time.Hour), End: base.Add(7 * time.Hour)},
	}
	cmp := compareCandidateTrips(cand, legacy, nil, ShadowConfig{Enabled: false, SummaryEnabled: false, MatchToleranceMinutes: 20})
	if cmp.ApproxMatchedTrips == 0 {
		t.Fatalf("expected baseline approximate matches")
	}
	if len(cmp.UnmatchedCandidateTrips) == 0 || len(cmp.UnmatchedLegacyTrips) == 0 {
		t.Fatalf("expected unmatched counts when summary disabled: %+v", cmp)
	}
	if cmp.ShadowSummary.Readiness != ShadowReadinessNotEvaluated {
		t.Fatalf("expected not evaluated readiness when summary disabled, got %s", cmp.ShadowSummary.Readiness)
	}
}
