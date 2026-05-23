package engine

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/routes"
	"gogator/internal/sites"
)

func TestDetectStaysStationaryCluster(t *testing.T) {
	pts := motionFixturePoints()
	e := buildEvidence(pts, false)
	m := classifyMotion(e.Points, MotionConfig{Enabled: true, StationarySpeedThresholdKPH: 2, MovingSpeedThresholdKPH: 8, GapThresholdMinutes: 20, MinConsecutiveSamples: 2})
	s := detectStays(e.Points, m, nil, StayConfig{Enabled: true, MinDurationMinutes: 5, MaxRadiusMeters: 120, MinPoints: 3, SiteMatchRadiusMeters: 80, GapInferredStopEnabled: true})
	if len(s.Stays) == 0 {
		t.Fatalf("expected stay")
	}
}

func TestStayDeterministic(t *testing.T) {
	pts := motionFixturePoints()
	e := buildEvidence(pts, false)
	m := classifyMotion(e.Points, MotionConfig{Enabled: true, StationarySpeedThresholdKPH: 2, MovingSpeedThresholdKPH: 8, GapThresholdMinutes: 20, MinConsecutiveSamples: 2})
	cfg := StayConfig{Enabled: true, MinDurationMinutes: 5, MaxRadiusMeters: 120, MinPoints: 3, SiteMatchRadiusMeters: 100, GapInferredStopEnabled: true}
	a := detectStays(e.Points, m, nil, cfg)
	b := detectStays(e.Points, m, nil, cfg)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("not deterministic")
	}
}

func TestDetectStaysRespectsMinPoints(t *testing.T) {
	pts := []gps.RawPoint{
		mkPoint("s.csv", 2, time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC), -33.0, 151.0, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("s.csv", 3, time.Date(2026, 5, 1, 8, 10, 0, 0, time.UTC), -33.00001, 151.00001, 0, map[string]float64{"io24": 0, "io251": 1}),
	}
	e := buildEvidence(pts, false)
	m := classifyMotion(e.Points, MotionConfig{Enabled: true, StationarySpeedThresholdKPH: 2, MovingSpeedThresholdKPH: 8, GapThresholdMinutes: 20, MinConsecutiveSamples: 1})
	s := detectStays(e.Points, m, nil, StayConfig{Enabled: true, MinDurationMinutes: 5, MaxRadiusMeters: 120, MinPoints: 3, SiteMatchRadiusMeters: 80, GapInferredStopEnabled: false})
	if len(s.Stays) != 0 {
		t.Fatalf("expected no stays for cluster smaller than min_points, got %d", len(s.Stays))
	}
}

func TestRunStayDetectionPassiveParity(t *testing.T) {
	cfg := config.Default()
	cfg.Engine.Motion.Enabled = true
	cfg.Engine.StayDetection.Enabled = true
	siteList := []sites.Site{{Name: "Home", Lat: -33, Lng: 151, RadiusM: 120, Important: true}}
	rules := []routes.Route{}
	pts := motionFixturePoints()
	base, _ := Run(context.Background(), Input{Points: clonePoints(pts), Sites: siteList, Routes: rules, Config: cfg, EngineConfig: EngineConfig{Quality: false, Motion: MotionConfig{Enabled: true, StationarySpeedThresholdKPH: 2, MovingSpeedThresholdKPH: 8, GapThresholdMinutes: 20, MinConsecutiveSamples: 2}, StayDetection: StayConfig{Enabled: false}}})
	with, _ := Run(context.Background(), Input{Points: clonePoints(pts), Sites: siteList, Routes: rules, Config: cfg, EngineConfig: EngineConfig{Quality: false, Motion: MotionConfig{Enabled: true, StationarySpeedThresholdKPH: 2, MovingSpeedThresholdKPH: 8, GapThresholdMinutes: 20, MinConsecutiveSamples: 2}, StayDetection: StayConfig{Enabled: true, MinDurationMinutes: 5, MaxRadiusMeters: 120, MinPoints: 3, SiteMatchRadiusMeters: 100, GapInferredStopEnabled: true}}})
	if !reflect.DeepEqual(normalizeTrips(base.Valid), normalizeTrips(with.Valid)) || !reflect.DeepEqual(normalizeTrips(base.Jitter), normalizeTrips(with.Jitter)) || !reflect.DeepEqual(base.RouteObservations, with.RouteObservations) || !reflect.DeepEqual(base.RouteAnomalies, with.RouteAnomalies) {
		t.Fatalf("passive parity failed")
	}
	if len(with.Stays.Stays) == 0 {
		t.Fatalf("expected passive stays")
	}
}

func motionFixturePoints() []gps.RawPoint {
	return []gps.RawPoint{
		mkPoint("m.csv", 2, time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC), -33.0, 151.0, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("m.csv", 3, time.Date(2026, 5, 1, 8, 4, 0, 0, time.UTC), -33.00002, 151.00001, 0.5, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("m.csv", 4, time.Date(2026, 5, 1, 8, 9, 0, 0, time.UTC), -33.00001, 151.0, 0.8, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("m.csv", 5, time.Date(2026, 5, 1, 8, 12, 0, 0, time.UTC), -33.001, 151.001, 30, map[string]float64{"io24": 1}),
	}
}
