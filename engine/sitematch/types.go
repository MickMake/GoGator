package sitematch

import "fmt"

type SiteMatcher interface {
	Match(req SiteMatchRequest) (SiteMatchResult, error)
}

type SiteMatchRequest struct {
	Latitude     float64
	Longitude    float64
	RadiusMeters float64
}

type SiteMatchResult struct {
	Candidates []SiteMatchCandidate
	Confidence SiteMatchConfidence
	Warnings   []SiteMatchWarning
}

type SiteMatchCandidate struct {
	SiteName       string
	SiteAddress    string
	SiteLatitude   float64
	SiteLongitude  float64
	DistanceMeters float64
	Metadata       map[string]string
}

type SiteMatchWarning struct {
	Code    string
	Message string
}

type SiteMatchConfidence struct {
	Score float64
}

var ErrPostGISNotImplemented = fmt.Errorf("postgis site matcher is scaffold-only and not implemented")
