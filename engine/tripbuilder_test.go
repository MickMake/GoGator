package engine

import (
	"reflect"
	"testing"
	"time"
)

func TestCandidateTripTypesAndDeterminism(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	vis := VisitEvidence{Visits: []Visit{{StartTime: base, EndTime: base.Add(10 * time.Minute), MatchedSite: "A", Confidence: VisitConfidenceHigh, PointIndexes: []int{1, 2}}, {StartTime: base.Add(20 * time.Minute), EndTime: base.Add(30 * time.Minute), MatchedSite: "B", Confidence: VisitConfidenceHigh, PointIndexes: []int{3, 4}}}}
	ex := detectExcursions(vis, ExcursionConfig{Enabled: true})
	got := detectCandidateTrips(vis, ex, TripBuilderConfig{Enabled: true, PassiveOnly: true, MinTripDurationMinutes: 1})
	if len(got.Trips) != 1 || got.Trips[0].Type != CandidateTripSiteToSite {
		t.Fatalf("expected one site-to-site trip")
	}
	got2 := detectCandidateTrips(vis, ex, TripBuilderConfig{Enabled: true, PassiveOnly: true, MinTripDurationMinutes: 1})
	if !reflect.DeepEqual(got, got2) {
		t.Fatalf("non-deterministic")
	}
}

func TestCandidateKnownToUnknownAndLowConfidence(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	vis := VisitEvidence{Visits: []Visit{{StartTime: base, EndTime: base.Add(10 * time.Minute), MatchedSite: "A", Confidence: VisitConfidenceHigh, PointIndexes: []int{1}}, {StartTime: base.Add(12 * time.Minute), EndTime: base.Add(20 * time.Minute), Confidence: VisitConfidenceLow, Reasons: []VisitReason{VisitReasonFromStayType}, PointIndexes: []int{2}}}}
	got := detectCandidateTrips(vis, ExcursionEvidence{}, TripBuilderConfig{Enabled: true, MinTripDurationMinutes: 5})
	if got.Trips[0].Type != CandidateTripNoiseAffected && got.Trips[0].Type != CandidateTripLowConfidence {
		t.Fatalf("expected noisy/low confidence type")
	}
	if got.Trips[0].SiteMatchBoundary != BoundaryLow {
		t.Fatalf("expected site match low")
	}
}
