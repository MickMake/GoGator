package engine

import (
	"testing"
	"time"

	"gogator/internal/gps"
)

func TestTripComparisonMatchedAndUnmatched(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	cand := CandidateTripEvidence{Trips: []CandidateTrip{{StartTime: base, EndTime: base.Add(10 * time.Minute), OriginLabel: "A", DestinationLabel: "B"}, {StartTime: base.Add(30 * time.Minute), EndTime: base.Add(40 * time.Minute)}}}
	legacy := []gps.Trip{{Start: base.Add(time.Minute), End: base.Add(11 * time.Minute), DepartureSite: "A", DestinationSite: "B"}}
	cmp := compareCandidateTrips(cand, legacy, nil, ShadowConfig{Enabled: true, SummaryEnabled: true})
	if cmp.ApproxMatchedTrips != 1 || len(cmp.UnmatchedCandidateTrips) != 1 || len(cmp.UnmatchedLegacyTrips) != 0 {
		t.Fatalf("unexpected comparison: %+v", cmp)
	}
}

func TestTripComparisonMatchesAgainstValidAndJitter(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	candidates := CandidateTripEvidence{Trips: []CandidateTrip{
		{StartTime: base, EndTime: base.Add(10 * time.Minute), OriginLabel: "A", DestinationLabel: "B"},
		{StartTime: base.Add(30 * time.Minute), EndTime: base.Add(40 * time.Minute), OriginLabel: "X", DestinationLabel: "Y"},
	}}
	legacyValid := []gps.Trip{{Start: base, End: base.Add(10 * time.Minute), DepartureSite: "A", DestinationSite: "B"}}
	legacyJitter := []gps.Trip{{Start: base.Add(30 * time.Minute), End: base.Add(40 * time.Minute), DepartureSite: "J1", DestinationSite: "J2"}}
	cmp := compareCandidateTrips(candidates, legacyValid, legacyJitter, ShadowConfig{Enabled: true, SummaryEnabled: true})
	if cmp.ApproxMatchedTrips != 2 {
		t.Fatalf("expected both candidates matched, got %+v", cmp)
	}
	if len(cmp.UnmatchedLegacyTrips) != 0 {
		t.Fatalf("expected no unmatched legacy trips, got %+v", cmp)
	}
	if cmp.SiteDifferenceCount != 1 {
		t.Fatalf("expected one site mismatch from jitter match, got %+v", cmp)
	}
}

func TestTripComparisonUnmatchedLegacyIncludesJitterAndHonorsTolerance(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	candidates := CandidateTripEvidence{Trips: []CandidateTrip{
		{StartTime: base, EndTime: base.Add(10 * time.Minute)},
	}}
	legacyValid := []gps.Trip{{Start: base, End: base.Add(10 * time.Minute)}}
	legacyJitter := []gps.Trip{{Start: base.Add(40 * time.Minute), End: base.Add(50 * time.Minute)}}
	cmp := compareCandidateTrips(candidates, legacyValid, legacyJitter, ShadowConfig{Enabled: true, SummaryEnabled: true, MatchToleranceMinutes: 5})
	if len(cmp.UnmatchedLegacyTrips) != 1 {
		t.Fatalf("expected unmatched jitter legacy trip to be counted, got %+v", cmp)
	}
	candidatesLate := CandidateTripEvidence{Trips: []CandidateTrip{{StartTime: base.Add(6 * time.Minute), EndTime: base.Add(16 * time.Minute)}}}
	cmpLate := compareCandidateTrips(candidatesLate, legacyValid, nil, ShadowConfig{Enabled: true, SummaryEnabled: true, MatchToleranceMinutes: 5})
	if cmpLate.ApproxMatchedTrips != 0 {
		t.Fatalf("expected no match beyond custom tolerance, got %+v", cmpLate)
	}
}
