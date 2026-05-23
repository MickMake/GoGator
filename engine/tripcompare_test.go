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
	cmp := compareCandidateTrips(cand, legacy, nil)
	if cmp.ApproxMatchedTrips != 1 || len(cmp.UnmatchedCandidateTrips) != 1 || len(cmp.UnmatchedLegacyTrips) != 0 {
		t.Fatalf("unexpected comparison: %+v", cmp)
	}
}
