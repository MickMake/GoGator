package sitematch

import (
	"fmt"
	"strings"
)

type PostGISConfig struct {
	Enabled           bool
	DSN               string
	MatchRadiusMeters float64
	AuditEnabled      bool
}

type PostGISSiteMatcher struct {
	cfg PostGISConfig
}

func NewPostGISSiteMatcher(cfg PostGISConfig) (*PostGISSiteMatcher, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("postgis site matcher is disabled")
	}
	if strings.TrimSpace(cfg.DSN) == "" {
		return nil, fmt.Errorf("postgis enabled but dsn is empty")
	}
	if cfg.MatchRadiusMeters <= 0 {
		return nil, fmt.Errorf("postgis enabled but match_radius_meters must be greater than zero")
	}
	return &PostGISSiteMatcher{cfg: cfg}, nil
}

func (p *PostGISSiteMatcher) Match(req SiteMatchRequest) (SiteMatchResult, error) {
	_ = req
	if p == nil {
		return SiteMatchResult{}, ErrPostGISNotImplemented
	}
	return SiteMatchResult{}, ErrPostGISNotImplemented
}
