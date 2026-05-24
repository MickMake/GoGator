package engine

import (
	"fmt"
	"time"

	"gogator/engine/mapmatch"
)

type mapMatcherValidator interface {
	Validate() error
}

func newMapMatcher(cfg EngineConfig) mapmatch.MapMatcher {
	if !cfg.Valhalla.Enabled {
		return mapmatch.NoopMapMatcher{}
	}
	return mapmatch.NewValhallaMapMatcher(mapmatch.ValhallaConfig{BaseURL: cfg.Valhalla.BaseURL, Timeout: time.Duration(cfg.Valhalla.TimeoutSeconds) * time.Second, Endpoint: cfg.Valhalla.Endpoint, MaxPointsPerRequest: cfg.Valhalla.MaxPointsPerRequest})
}

func validateMapMatcher(cfg EngineConfig, matcher mapmatch.MapMatcher) error {
	if !cfg.Valhalla.Enabled {
		return nil
	}
	validator, ok := matcher.(mapMatcherValidator)
	if !ok {
		return fmt.Errorf("valhalla is enabled but matcher does not support validation")
	}
	if err := validator.Validate(); err != nil {
		return err
	}
	return nil
}
