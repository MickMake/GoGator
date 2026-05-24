package mapmatch

import "time"

type MapMatcher interface {
	Match(req MatchRequest) (MatchResult, error)
}

type MatchRequest struct {
	Points              []MatchPoint
	Endpoint            string
	MaxPointsPerRequest int
}

type MatchPoint struct {
	Lat  float64
	Lng  float64
	Time time.Time
}

type MatchResult struct {
	MatchedShape []MatchedShapePoint
	MatchedEdges []MatchedEdge
	DistanceM    float64
	DurationS    float64
	Confidence   MatchConfidence
	Warnings     []MatchWarning
}

type MatchedShapePoint struct {
	Lat float64
	Lng float64
}

type MatchedEdge struct {
	ID   string
	Name string
}

type MatchWarning struct {
	Code    string
	Message string
}

type MatchConfidence struct {
	Score float64
}
