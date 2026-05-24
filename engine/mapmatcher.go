package engine

import (
	"time"

	"gogator/engine/mapmatch"
)

func newMapMatcher(cfg EngineConfig) mapmatch.MapMatcher {
	if !cfg.Valhalla.Enabled {
		return mapmatch.NoopMapMatcher{}
	}
	return mapmatch.NewValhallaMapMatcher(mapmatch.ValhallaConfig{BaseURL: cfg.Valhalla.BaseURL, Timeout: time.Duration(cfg.Valhalla.TimeoutSeconds) * time.Second, Endpoint: cfg.Valhalla.Endpoint, MaxPointsPerRequest: cfg.Valhalla.MaxPointsPerRequest})
}
