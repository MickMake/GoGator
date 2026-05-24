package engine

import (
	"context"
	"reflect"
	"slices"
	"sort"
	"testing"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/routes"
	"gogator/internal/sites"
)

func TestRunMatchesLegacyPipeline(t *testing.T) {
	cfg := config.Default()
	siteList := []sites.Site{
		{Name: "Home", Address: "1 Home St", Lat: -33.00000, Lng: 151.00000, RadiusM: 120, MinDestinationMinutes: 1, Important: true},
		{Name: "Work", Address: "99 Work Rd", Lat: -33.01000, Lng: 151.01000, RadiusM: 120, MinDestinationMinutes: 1, Important: true},
	}
	rules := []routes.Route{{Name: "Home to Work", FromSite: "Home", ToSite: "Work", DistanceMinKM: 0.2, DistanceMaxKM: 5, DurationMinMin: 1, DurationMaxMin: 30, ConfidenceBoost: "Good"}}

	start := time.Date(2026, 5, 1, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{
		mkPoint("b.csv", 2, start.Add(3*time.Minute), -33.0002, 151.0002, 28, map[string]float64{"io24": 1}),
		mkPoint("a.csv", 2, start, -33.0000, 151.0000, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("b.csv", 3, start.Add(4*time.Minute), -33.0010, 151.0010, 35, map[string]float64{"io24": 1}),
		mkPoint("b.csv", 4, start.Add(5*time.Minute), -33.0100, 151.0100, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("a.csv", 3, start.Add(1*time.Minute), -33.0000, 151.0000, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("a.csv", 4, start.Add(2*time.Minute), -33.0000, 151.0000, 0, map[string]float64{"io24": 0, "io251": 1}),
	}

	expected := runLegacy(clonePoints(pts), siteList, rules, cfg)
	actual, err := Run(context.Background(), Input{Points: clonePoints(pts), Sites: siteList, Routes: rules, Config: cfg})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	if !reflect.DeepEqual(normalizeTrips(actual.Valid), normalizeTrips(expected.Valid)) {
		t.Fatalf("valid trips mismatch")
	}
	if !reflect.DeepEqual(normalizeTrips(actual.Jitter), normalizeTrips(expected.Jitter)) {
		t.Fatalf("jitter mismatch")
	}
	if !reflect.DeepEqual(actual.RouteObservations, expected.RouteObservations) {
		t.Fatalf("route observations mismatch")
	}
	if !reflect.DeepEqual(actual.RouteAnomalies, expected.RouteAnomalies) {
		t.Fatalf("route anomalies mismatch")
	}
	if !reflect.DeepEqual(normalizeTrips(actual.JitterReview), normalizeTrips(expected.JitterReview)) {
		t.Fatalf("jitter review mismatch")
	}
	if !reflect.DeepEqual(normalizeTrips(actual.JitterSameSite), normalizeTrips(expected.JitterSameSite)) {
		t.Fatalf("same-site jitter mismatch")
	}
}

func TestRunRoutesAppliedThroughEngineBoundary(t *testing.T) {
	cfg := config.Default()
	siteList := []sites.Site{
		{Name: "Alpha", Address: "A", Lat: -33.0, Lng: 151.0, RadiusM: 120, MinDestinationMinutes: 1, Important: true},
		{Name: "Beta", Address: "B", Lat: -33.01, Lng: 151.01, RadiusM: 120, MinDestinationMinutes: 1, Important: true},
	}
	rules := []routes.Route{{Name: "Alpha-Beta", FromSite: "Alpha", ToSite: "Beta", DistanceMinKM: 0.2, DistanceMaxKM: 5, DurationMinMin: 1, DurationMaxMin: 30, ConfidenceBoost: "Good"}}
	start := time.Date(2026, 5, 2, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{
		mkPoint("r.csv", 2, start, -33.0, 151.0, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("r.csv", 3, start.Add(1*time.Minute), -33.0, 151.0, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("r.csv", 4, start.Add(2*time.Minute), -33.0008, 151.0008, 30, map[string]float64{"io24": 1}),
		mkPoint("r.csv", 5, start.Add(3*time.Minute), -33.01, 151.01, 0, map[string]float64{"io24": 0, "io251": 1}),
	}

	res, err := Run(context.Background(), Input{Points: clonePoints(pts), Sites: siteList, Routes: rules, Config: cfg})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(res.Valid) == 0 {
		t.Fatalf("expected at least one valid trip")
	}
	if len(res.RouteObservations) == 0 {
		t.Fatalf("expected route observations from engine boundary")
	}
}

func TestRunDefaultEngineConfigIsBehaviorNeutral(t *testing.T) {
	cfg := config.Default()
	siteList := []sites.Site{{Name: "Home", Address: "A", Lat: -33.0, Lng: 151.0, RadiusM: 120, MinDestinationMinutes: 1, Important: true}}
	start := time.Date(2026, 5, 3, 8, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{
		mkPoint("n.csv", 2, start, -33.0, 151.0, 0, map[string]float64{"io24": 0, "io251": 1}),
		mkPoint("n.csv", 3, start.Add(1*time.Minute), -33.001, 151.001, 32, map[string]float64{"io24": 1}),
	}

	base, err := Run(context.Background(), Input{Points: clonePoints(pts), Sites: siteList, Config: cfg})
	if err != nil {
		t.Fatalf("Run base error: %v", err)
	}
	withCfg, err := Run(context.Background(), Input{
		Points: clonePoints(pts), Sites: siteList, Config: cfg,
		EngineConfig: EngineConfig{
			Enabled:           cfg.Engine.Enabled,
			CompatibilityMode: cfg.Engine.CompatibilityMode,
			StayDetection:     StayConfig{Enabled: cfg.Engine.StayDetection.Enabled, MinDurationMinutes: cfg.Engine.StayDetection.MinDurationMinutes, MaxRadiusMeters: cfg.Engine.StayDetection.MaxRadiusMeters, MinPoints: cfg.Engine.StayDetection.MinPoints, SiteMatchRadiusMeters: cfg.Engine.StayDetection.SiteMatchRadiusMeters, GapInferredStopEnabled: cfg.Engine.StayDetection.GapInferredStopEnabled},
			Visits:            VisitConfig{Enabled: cfg.Engine.Visits.Enabled, MinVisitDurationMinutes: cfg.Engine.Visits.MinVisitDurationMinutes},
			Excursions:        ExcursionConfig{Enabled: cfg.Engine.Excursions.Enabled, ShortOutAndBackMaxMinutes: cfg.Engine.Excursions.ShortOutAndBackMaxMinutes, ShortOutAndBackMaxDistance: cfg.Engine.Excursions.ShortOutAndBackMaxDistanceMeters},
			TripBuilder:       TripBuilderConfig{Enabled: cfg.Engine.TripBuilder.Enabled, PassiveOnly: cfg.Engine.TripBuilder.PassiveOnly, CompareLegacy: cfg.Engine.TripBuilder.CompareLegacy, MinTripDurationMinutes: cfg.Engine.TripBuilder.MinTripDurationMinutes, MaxGapMinutes: cfg.Engine.TripBuilder.MaxGapMinutes, LowConfidenceThreshold: cfg.Engine.TripBuilder.LowConfidenceThreshold},
			Motion:            MotionConfig{Enabled: cfg.Engine.Motion.Enabled, StationarySpeedThresholdKPH: cfg.Engine.Motion.StationarySpeedThresholdKPH, MovingSpeedThresholdKPH: cfg.Engine.Motion.MovingSpeedThresholdKPH, GapThresholdMinutes: cfg.Engine.Motion.GapThresholdMinutes, MinConsecutiveSamples: cfg.Engine.Motion.MinConsecutiveSamples},
			Quality:           cfg.Engine.Quality.Enabled,
			Audit:             cfg.Engine.Audit.Enabled,
			Valhalla:          ValhallaConfig{Enabled: cfg.Valhalla.Enabled},
			H3:                cfg.H3.Enabled,
			PostGIS:           cfg.PostGIS.Enabled,
		},
	})
	if err != nil {
		t.Fatalf("Run with config error: %v", err)
	}
	if !reflect.DeepEqual(normalizeTrips(base.Valid), normalizeTrips(withCfg.Valid)) || !reflect.DeepEqual(normalizeTrips(base.Jitter), normalizeTrips(withCfg.Jitter)) {
		t.Fatalf("default engine config should not change behavior")
	}
}

func TestRunBuildsEvidenceFromNormalizedPoints(t *testing.T) {
	cfg := config.Default()
	start := time.Date(2026, 5, 3, 9, 0, 0, 0, time.UTC)
	pts := []gps.RawPoint{
		mkPoint("n.csv", 5, start.Add(2*time.Minute), -33.002, 151.002, 12, map[string]float64{"io24": 1, "pdop": 2}),
		mkPoint("n.csv", 3, start, -33.0, 151.0, 0, map[string]float64{"io24": 0, "io251": 1, "pdop": 2}),
		mkPoint("n.csv", 4, start.Add(time.Minute), -33.001, 151.001, 10, map[string]float64{"io24": 1, "pdop": 2}),
	}
	res, err := Run(context.Background(), Input{Points: clonePoints(pts), Config: cfg, EngineConfig: EngineConfig{Quality: true}})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(res.Points) != len(res.Evidence.Points) {
		t.Fatalf("points/evidence length mismatch: %d vs %d", len(res.Points), len(res.Evidence.Points))
	}
	for i := range res.Points {
		if !res.Points[i].Time.Equal(res.Evidence.Points[i].Time) {
			t.Fatalf("evidence should align with normalized points at index %d", i)
		}
	}
}

func clonePoints(in []gps.RawPoint) []gps.RawPoint {
	out := make([]gps.RawPoint, len(in))
	for i, p := range in {
		cp := p
		if p.Params != nil {
			cp.Params = make(map[string]string, len(p.Params))
			for k, v := range p.Params {
				cp.Params[k] = v
			}
		}
		if p.ParamNums != nil {
			cp.ParamNums = make(map[string]float64, len(p.ParamNums))
			for k, v := range p.ParamNums {
				cp.ParamNums[k] = v
			}
		}
		out[i] = cp
	}
	return out
}

func normalizeTrips(in []gps.Trip) []gps.Trip {
	out := make([]gps.Trip, len(in))
	for i, t := range in {
		cp := t
		cp.Flags = append([]string(nil), t.Flags...)
		slices.Sort(cp.Flags)
		cp.Points = make([]gps.RawPoint, len(t.Points))
		for j, p := range t.Points {
			pp := p
			pp.Flags = append([]string(nil), p.Flags...)
			slices.Sort(pp.Flags)
			cp.Points[j] = pp
		}
		out[i] = cp
	}
	return out
}

func runLegacy(points []gps.RawPoint, siteList []sites.Site, routeRules []routes.Route, cfg config.Config) Result {
	sort.SliceStable(points, func(i, j int) bool {
		if points[i].Time.Equal(points[j].Time) {
			if points[i].SourceFile == points[j].SourceFile {
				return points[i].RawRow < points[j].RawRow
			}
			return points[i].SourceFile < points[j].SourceFile
		}
		return points[i].Time.Before(points[j].Time)
	})
	gps.RecalculatePointDeltas(points)
	points = gps.Classify(points, cfg)
	valid, jitter := gps.BuildTrips(points, cfg, siteList)
	valid, jitter = gps.CollapseToImportantSites(valid, jitter, cfg, siteList)
	valid, observations, anomalies := routes.Apply(valid, routeRules, cfg.Site.UnknownSiteLabel)
	review, sameSite := LegacyAdapter{}.SplitJitter(jitter)
	return Result{Points: points, Valid: valid, Jitter: jitter, JitterReview: review, JitterSameSite: sameSite, RouteObservations: observations, RouteAnomalies: anomalies, SiteCount: len(siteList), RouteCount: len(routeRules)}
}

func mkPoint(file string, row int, tm time.Time, lat, lng, speed float64, nums map[string]float64) gps.RawPoint {
	params := map[string]string{}
	for k, v := range nums {
		params[k] = "1"
		if v == 0 {
			params[k] = "0"
		}
	}
	return gps.RawPoint{SourceFile: file, RawRow: row, Time: tm, RawDT: tm.Format(time.RFC3339), Lat: lat, Lng: lng, SpeedKPH: speed, Params: params, ParamNums: nums}
}
