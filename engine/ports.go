package engine

import (
	"gogator/internal/gps"
	"gogator/internal/routes"
)

type ProcessorPorts interface {
	SortAndRecalculate(points []gps.RawPoint)
	Classify(points []gps.RawPoint, in Input) []gps.RawPoint
	BuildTrips(points []gps.RawPoint, in Input) (valid []gps.Trip, jitter []gps.Trip)
	ApplyImportant(valid []gps.Trip, jitter []gps.Trip, in Input) ([]gps.Trip, []gps.Trip)
	ApplyRoutes(valid []gps.Trip, in Input) ([]gps.Trip, []routes.Observation, []routes.Anomaly)
	SplitJitter(jitter []gps.Trip) (review []gps.Trip, sameSite []gps.Trip)
}
