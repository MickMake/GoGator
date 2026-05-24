package engine

import (
	"testing"
	"time"

	"gogator/internal/gps"
)

func TestTripComparisonWorksWhenShadowSummaryDisabled(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	cand := CandidateTripEvidence{Trips: []CandidateTrip{{StartTime: base, EndTime: base.Add(10 * time.Minute)}}}
	legacy := []gps.Trip{{Start: base, End: base.Add(10 * time.Minute)}}
	cmp := compareCandidateTrips(cand, legacy, nil, ShadowConfig{})
	if cmp.CandidateTripCount != 1 || cmp.LegacyValidTripCount != 1 || cmp.ApproxMatchedTrips != 1 {
		t.Fatalf("expected legacy comparison counts with shadow disabled: %+v", cmp)
	}
	if cmp.ShadowSummary.ApproxMatchedTripCount != 0 || cmp.ShadowSummary.Readiness != ShadowReadinessNotEvaluated {
		t.Fatalf("expected disabled shadow summary to remain unevaluated: %+v", cmp.ShadowSummary)
	}
}
