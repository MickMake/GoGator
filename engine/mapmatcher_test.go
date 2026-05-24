package engine

import (
	"testing"

	"gogator/engine/mapmatch"
	"gogator/internal/config"
)

func TestNewMapMatcherDisabledUsesNoop(t *testing.T) {
	m := newMapMatcher(EngineConfig{Valhalla: ValhallaConfig{Enabled: false}})
	if _, ok := m.(mapmatch.NoopMapMatcher); !ok {
		t.Fatalf("expected noop matcher, got %T", m)
	}
}

func TestConfigDefaultsKeepValhallaDisabled(t *testing.T) {
	cfg := config.Default()
	if cfg.Valhalla.Enabled {
		t.Fatalf("expected valhalla disabled by default")
	}
}
