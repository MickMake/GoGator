package engine

import (
	"fmt"

	"gogator/engine/sitematch"
)

func newSiteMatcher(cfg EngineConfig) (sitematch.SiteMatcher, error) {
	if !cfg.PostGIS.Enabled {
		return sitematch.NoopSiteMatcher{}, nil
	}
	matcher, err := sitematch.NewPostGISSiteMatcher(sitematch.PostGISConfig{Enabled: cfg.PostGIS.Enabled, DSN: cfg.PostGIS.DSN, MatchRadiusMeters: cfg.PostGIS.MatchRadiusMeters, AuditEnabled: cfg.PostGIS.AuditEnabled})
	if err != nil {
		return nil, fmt.Errorf("postgis site matcher: %w", err)
	}
	return matcher, nil
}
