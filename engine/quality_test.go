package engine

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
)

func TestBuildEvidenceMissingSignalsSafe(t *testing.T) {
	pts := []gps.RawPoint{{Lat: -33, Lng: 151, Time: time.Now().UTC()}}
	ev := buildEvidence(pts, true)
	if ev.Points[0].Signals.IO24 != nil || ev.Points[0].Signals.PDOP != nil {
		t.Fatalf("expected missing signals to stay nil")
	}
}

func TestQualityGoodPoint(t *testing.T) {
	tm := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{
		{Time: tm, Lat: -33, Lng: 151, SpeedKPH: 0, ParamNums: map[string]float64{"pdop": 1.2, "gpslev": 5, "gsmlev": 20}},
		{Time: tm.Add(time.Minute), Lat: -33.001, Lng: 151.001, SpeedKPH: 25, ParamNums: map[string]float64{"pdop": 1.5, "gpslev": 5, "gsmlev": 20, "io24": 1}},
	}
	ev := buildEvidence(pts, true)
	if ev.Points[1].Quality.Band != QualityGood {
		t.Fatalf("expected good band, got %s", ev.Points[1].Quality.Band)
	}
}

func TestQualityPoorGPSPoint(t *testing.T) {
	tm := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{{Time: tm, Lat: -33, Lng: 151, ParamNums: map[string]float64{"pdop": 7, "gpslev": 1, "gsmlev": 2}}}
	ev := buildEvidence(pts, true)
	if ev.Points[0].Quality.Band != QualityPoor {
		t.Fatalf("expected poor band, got %s", ev.Points[0].Quality.Band)
	}
}

func TestQualityInvalidCoordinates(t *testing.T) {
	pts := []gps.RawPoint{{Time: time.Now().UTC(), Lat: 0, Lng: 0}}
	ev := buildEvidence(pts, true)
	if ev.Points[0].Quality.Band != QualityInvalid {
		t.Fatalf("expected invalid")
	}
}

func TestQualityDuplicateTimestamp(t *testing.T) {
	tm := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{{Time: tm, Lat: -33, Lng: 151}, {Time: tm, Lat: -33.01, Lng: 151.01}}
	ev := buildEvidence(pts, true)
	r := ev.Points[1].Quality.Reasons
	found := false
	for _, rr := range r {
		if rr == ReasonDuplicateTimestamp {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected duplicate timestamp reason")
	}
}

func TestQualityDeterministic(t *testing.T) {
	tm := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{{Time: tm, Lat: -33, Lng: 151, ParamNums: map[string]float64{"pdop": 2, "gpslev": 4}}}
	a := buildEvidence(pts, true)
	b := buildEvidence(pts, true)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("expected deterministic evidence")
	}
}

func TestRunQualityIsPassive(t *testing.T) {
	cfg := config.Default()
	cfg.Engine.Quality.Enabled = true
	start := time.Date(2026, 5, 3, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{
		{SourceFile: "n.csv", RawRow: 2, Time: start, Lat: -33.0, Lng: 151.0, SpeedKPH: 0, ParamNums: map[string]float64{"io24": 0, "io251": 1}},
		{SourceFile: "n.csv", RawRow: 3, Time: start.Add(time.Minute), Lat: -33.001, Lng: 151.001, SpeedKPH: 32, ParamNums: map[string]float64{"io24": 1}},
	}
	base, err := Run(context.Background(), Input{Points: clonePoints(pts), Config: config.Default()})
	if err != nil {
		t.Fatal(err)
	}
	withQuality, err := Run(context.Background(), Input{Points: clonePoints(pts), Config: cfg})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(normalizeTrips(base.Valid), normalizeTrips(withQuality.Valid)) || !reflect.DeepEqual(normalizeTrips(base.Jitter), normalizeTrips(withQuality.Jitter)) {
		t.Fatalf("quality scoring must be passive")
	}
	if len(withQuality.Evidence.Points) != len(pts) {
		t.Fatalf("expected evidence points")
	}
}
