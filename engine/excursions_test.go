package engine

import (
	"reflect"
	"testing"
	"time"
)

func TestExcursionsBetweenVisitsAndOutBack(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	visits := VisitEvidence{Visits: []Visit{
		{StartTime: base, EndTime: base.Add(10 * time.Minute), Latitude: -33, Longitude: 151, MatchedSite: "A", Type: VisitKnownSite, Confidence: VisitConfidenceHigh, PointIndexes: []int{1, 2}},
		{StartTime: base.Add(20 * time.Minute), EndTime: base.Add(25 * time.Minute), Latitude: -33.001, Longitude: 151.001, Type: VisitUnknown, Confidence: VisitConfidenceMedium, PointIndexes: []int{3, 4}},
		{StartTime: base.Add(30 * time.Minute), EndTime: base.Add(35 * time.Minute), Latitude: -33, Longitude: 151, MatchedSite: "A", Type: VisitKnownSite, Confidence: VisitConfidenceHigh, PointIndexes: []int{5, 6}},
	}}
	e := detectExcursions(visits, ExcursionConfig{Enabled: true, ShortOutAndBackMaxMinutes: 25, ShortOutAndBackMaxDistance: 500})
	if len(e.Excursions) != 2 {
		t.Fatalf("expected 2 excursions")
	}
	if e.Excursions[1].Type != ExcursionShortOutAndBack {
		t.Fatalf("expected short out and back")
	}
	e2 := detectExcursions(visits, ExcursionConfig{Enabled: true, ShortOutAndBackMaxMinutes: 25, ShortOutAndBackMaxDistance: 500})
	if !reflect.DeepEqual(e, e2) {
		t.Fatalf("not deterministic")
	}
}

func TestExcursionsFallbackShortOutAndBackRespectsDistanceCap(t *testing.T) {
	base := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	cfg := ExcursionConfig{Enabled: true, ShortOutAndBackMaxMinutes: 25, ShortOutAndBackMaxDistance: 500}
	visitsFar := VisitEvidence{Visits: []Visit{
		{StartTime: base, EndTime: base.Add(10 * time.Minute), Latitude: -33, Longitude: 151, MatchedSite: "A", Type: VisitKnownSite, Confidence: VisitConfidenceHigh},
		{StartTime: base.Add(20 * time.Minute), EndTime: base.Add(25 * time.Minute), Latitude: -33.02, Longitude: 151.02, Type: VisitUnknown, Confidence: VisitConfidenceMedium},
		{StartTime: base.Add(30 * time.Minute), EndTime: base.Add(35 * time.Minute), Latitude: -33, Longitude: 151, MatchedSite: "A", Type: VisitKnownSite, Confidence: VisitConfidenceHigh},
	}}
	eFar := detectExcursions(visitsFar, cfg)
	if eFar.Excursions[1].Type == ExcursionShortOutAndBack {
		t.Fatalf("expected far fallback excursion not to be short out and back")
	}

	visitsNear := VisitEvidence{Visits: []Visit{
		{StartTime: base, EndTime: base.Add(10 * time.Minute), Latitude: -33, Longitude: 151, MatchedSite: "A", Type: VisitKnownSite, Confidence: VisitConfidenceHigh},
		{StartTime: base.Add(20 * time.Minute), EndTime: base.Add(25 * time.Minute), Latitude: -33.001, Longitude: 151.001, Type: VisitUnknown, Confidence: VisitConfidenceMedium},
		{StartTime: base.Add(30 * time.Minute), EndTime: base.Add(35 * time.Minute), Latitude: -33, Longitude: 151, MatchedSite: "A", Type: VisitKnownSite, Confidence: VisitConfidenceHigh},
	}}
	eNear := detectExcursions(visitsNear, cfg)
	if eNear.Excursions[1].Type != ExcursionShortOutAndBack {
		t.Fatalf("expected near fallback excursion to be short out and back")
	}
}
