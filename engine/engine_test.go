package engine

import (
	"context"
	"reflect"
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

	expected := runLegacy(append([]gps.RawPoint(nil), pts...), siteList, rules, cfg)
	actual, err := Run(context.Background(), Input{Points: pts, Sites: siteList, Routes: rules, Config: cfg})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}

	if !reflect.DeepEqual(actual.Valid, expected.Valid) {
		t.Fatalf("valid trips mismatch")
	}
	if !reflect.DeepEqual(actual.Jitter, expected.Jitter) {
		t.Fatalf("jitter mismatch")
	}
	if !reflect.DeepEqual(actual.RouteObservations, expected.RouteObservations) {
		t.Fatalf("route observations mismatch")
	}
	if !reflect.DeepEqual(actual.RouteAnomalies, expected.RouteAnomalies) {
		t.Fatalf("route anomalies mismatch")
	}
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
