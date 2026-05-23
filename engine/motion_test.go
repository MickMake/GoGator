package engine

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
)

func TestMotionStationarySequence(t *testing.T) {
	pts := motionPoints([]float64{0, 0, 0}, []float64{0, 0, 0}, time.Minute)
	mv := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	for _, s := range mv.Samples {
		if s.State != MotionStationary {
			t.Fatalf("expected stationary, got %s", s.State)
		}
	}
}

func TestMotionMovingSequence(t *testing.T) {
	pts := motionPoints([]float64{20, 25, 30}, []float64{1, 1, 1}, time.Minute)
	mv := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	for _, s := range mv.Samples {
		if s.State != MotionMoving {
			t.Fatalf("expected moving, got %s", s.State)
		}
	}
}

func TestMotionHysteresisSpikeDoesNotFlipImmediately(t *testing.T) {
	pts := motionPoints([]float64{0, 25, 0}, []float64{0, 1, 0}, time.Minute)
	mv := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	if mv.Samples[1].State != MotionStationary {
		t.Fatalf("single spike should not flip state")
	}
}

func TestMotionHysteresisPauseDoesNotEndMovement(t *testing.T) {
	pts := motionPoints([]float64{30, 30, 0, 30}, []float64{1, 1, 0, 1}, time.Minute)
	mv := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	if mv.Samples[2].State != MotionMoving {
		t.Fatalf("single pause should not immediately stop movement")
	}
}

func TestMotionGapAndNoise(t *testing.T) {
	pts := motionPoints([]float64{10, 0, 10}, []float64{1, 0, 1}, 30*time.Minute)
	pts[1].Lat, pts[1].Lng = 0, 0
	mv := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	if mv.Samples[1].State != MotionNoise {
		t.Fatalf("expected noise")
	}
	if mv.Samples[2].State != MotionGap {
		t.Fatalf("expected gap")
	}
}

func TestMotionNoiseDoesNotPoisonHysteresisRecovery(t *testing.T) {
	pts := motionPoints([]float64{0, 0, 0}, []float64{0, 0, 0}, time.Minute)
	pts[1].Lat, pts[1].Lng = 0, 0
	mv := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	if mv.Samples[2].State != MotionStationary {
		t.Fatalf("expected recovery to stationary after noise, got %s", mv.Samples[2].State)
	}
}

func TestMotionDeterministic(t *testing.T) {
	pts := motionPoints([]float64{0, 10, 20}, []float64{0, 1, 1}, time.Minute)
	a := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	b := classifyMotion(buildEvidence(pts, true).Points, defaultMotionConfig())
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("deterministic expected")
	}
}

func TestRunMotionIsPassive(t *testing.T) {
	cfg := config.Default()
	cfg.Engine.Motion.Enabled = true
	start := time.Date(2026, 5, 3, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{{SourceFile: "n.csv", RawRow: 2, Time: start, Lat: -33.0, Lng: 151.0, SpeedKPH: 0, ParamNums: map[string]float64{"io24": 0, "io251": 1}}, {SourceFile: "n.csv", RawRow: 3, Time: start.Add(time.Minute), Lat: -33.001, Lng: 151.001, SpeedKPH: 32, ParamNums: map[string]float64{"io24": 1}}}
	base, _ := Run(context.Background(), Input{Points: clonePoints(pts), Config: config.Default()})
	withMotion, _ := Run(context.Background(), Input{Points: clonePoints(pts), Config: cfg})
	if !reflect.DeepEqual(normalizeTrips(base.Valid), normalizeTrips(withMotion.Valid)) || !reflect.DeepEqual(normalizeTrips(base.Jitter), normalizeTrips(withMotion.Jitter)) {
		t.Fatalf("motion must be passive")
	}
	if len(withMotion.Motion.Samples) != len(pts) {
		t.Fatalf("expected motion diagnostics")
	}
}

func motionPoints(speeds, io24 []float64, step time.Duration) []gps.RawPoint {
	start := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	pts := make([]gps.RawPoint, len(speeds))
	for i := range speeds {
		pts[i] = gps.RawPoint{Time: start.Add(time.Duration(i) * step), Lat: -33 + float64(i)*0.0001, Lng: 151 + float64(i)*0.0001, SpeedKPH: speeds[i], ParamNums: map[string]float64{"io24": io24[i], "gpslev": 4, "pdop": 2}}
	}
	return pts
}

func defaultMotionConfig() MotionConfig {
	return MotionConfig{Enabled: true, StationarySpeedThresholdKPH: 2, MovingSpeedThresholdKPH: 8, GapThresholdMinutes: 20, MinConsecutiveSamples: 2}
}
