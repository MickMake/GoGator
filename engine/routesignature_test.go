package engine

import (
	"testing"
	"time"

	"gogator/engine/mapmatch"
	"gogator/internal/gps"
)

func TestRouteSignatureDeterministicAndCollapse(t *testing.T) {
	pts := []gps.RawPoint{{Lat: -33, Lng: 151, Time: time.Now()}, {Lat: -33, Lng: 151, Time: time.Now()}, {Lat: -33.01, Lng: 151.01, Time: time.Now()}}
	trips := []CandidateTrip{{SourcePointStart: 0, SourcePointEnd: 2}}
	a := buildRouteSignatures(pts, trips, nil, true, 7)
	b := buildRouteSignatures(pts, trips, nil, true, 7)
	if len(a) != 1 || a[0].Signature != b[0].Signature {
		t.Fatalf("expected deterministic signature")
	}
	if a[0].CellCount >= a[0].PointCount {
		t.Fatalf("expected duplicate collapse")
	}
}
func TestRouteSignatureInsufficientPointsWarning(t *testing.T) {
	pts := []gps.RawPoint{{Lat: -33, Lng: 151, Time: time.Now()}}
	s := buildRouteSignatures(pts, []CandidateTrip{{SourcePointStart: 0, SourcePointEnd: 0}}, nil, true, 7)
	if len(s) != 1 || len(s[0].Warnings) == 0 {
		t.Fatalf("expected warning")
	}
}
func TestRouteSignaturePrefersMapMatch(t *testing.T) {
	pts := []gps.RawPoint{{Lat: -33, Lng: 151, Time: time.Now()}, {Lat: -33.2, Lng: 151.2, Time: time.Now()}}
	trips := []CandidateTrip{{SourcePointStart: 0, SourcePointEnd: 1}}
	mm := []CandidateTripMapMatchDiagnostic{{CandidateTripID: 1, MatchedShape: []mapmatch.MatchedShapePoint{{Lat: -10, Lng: 10}, {Lat: -10.1, Lng: 10.1}}}}
	s := buildRouteSignatures(pts, trips, mm, true, 7)
	if s[0].Source != RouteSignatureSourceMatched {
		t.Fatalf("expected matched source")
	}
}
