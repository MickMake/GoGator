package engine

import (
	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/routes"
	"gogator/internal/sites"
)

type Input struct {
	Points       []gps.RawPoint
	Sites        []sites.Site
	Routes       []routes.Route
	Config       config.Config
	EngineConfig EngineConfig
}

type EngineConfig struct {
	Enabled           bool
	CompatibilityMode bool
	StayDetection     StayConfig
	Visits            VisitConfig
	Excursions        ExcursionConfig
	TripBuilder       TripBuilderConfig
	Motion            MotionConfig
	Quality           bool
	Audit             bool
	Valhalla          bool
	H3                bool
	PostGIS           bool
}

type Result struct {
	Points            []gps.RawPoint
	Evidence          EvidenceSet
	Motion            MotionEvidence
	Stays             StayEvidence
	Visits            VisitEvidence
	Excursions        ExcursionEvidence
	CandidateTrips    CandidateTripEvidence
	TripComparison    CandidateTripComparison
	Valid             []gps.Trip
	Jitter            []gps.Trip
	JitterReview      []gps.Trip
	JitterSameSite    []gps.Trip
	RouteObservations []routes.Observation
	RouteAnomalies    []routes.Anomaly
	SiteCount         int
	RouteCount        int
}
