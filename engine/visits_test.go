package engine

import (
	"testing"
	"time"

	"gogator/internal/sites"
)

func TestVisitKnownAndUnknownClassification(t *testing.T) {
	stays := StayEvidence{Stays: []Stay{{StartTime: time.Unix(0, 0), EndTime: time.Unix(600, 0), Duration: 10 * time.Minute, Latitude: -33, Longitude: 151, MatchedSite: "Depot", Type: StayTypeSiteStop, Confidence: StayConfidenceHigh, PointIndexes: []int{1, 2, 3}}, {StartTime: time.Unix(1200, 0), EndTime: time.Unix(1320, 0), Duration: 2 * time.Minute, Latitude: -33.01, Longitude: 151.01, Type: StayTypePause, Confidence: StayConfidenceLow, PointIndexes: []int{4}}}}
	v := detectVisits(stays, []sites.Site{{Name: "Depot", SiteType: "Supplier"}}, VisitConfig{Enabled: true, MinVisitDurationMinutes: 5})
	if len(v.Visits) != 2 {
		t.Fatalf("expected 2 visits")
	}
	if v.Visits[0].Type != VisitSupplier {
		t.Fatalf("expected supplier visit")
	}
	if v.Visits[1].Type != VisitPause {
		t.Fatalf("expected pause visit")
	}
}
