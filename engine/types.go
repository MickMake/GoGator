package engine

import (
	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/routes"
	"gogator/internal/sites"
)

type Input struct {
	Points []gps.RawPoint
	Sites  []sites.Site
	Routes []routes.Route
	Config config.Config
}

type Result struct {
	Points            []gps.RawPoint
	Valid             []gps.Trip
	Jitter            []gps.Trip
	JitterReview      []gps.Trip
	JitterSameSite    []gps.Trip
	RouteObservations []routes.Observation
	RouteAnomalies    []routes.Anomaly
	SiteCount         int
	RouteCount        int
}
