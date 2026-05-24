package engine

import (
	"errors"
	"testing"

	"gogator/engine/mapmatch"
	"gogator/internal/config"
)

type fakeMatcher struct{}

func (fakeMatcher) Match(req mapmatch.MatchRequest) (mapmatch.MatchResult, error) { return mapmatch.MatchResult{}, nil }
func (fakeMatcher) Validate() error                                                { return errors.New("validate-called") }

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

func TestValidateMapMatcherDisabledDoesNotValidate(t *testing.T) {
	err := validateMapMatcher(EngineConfig{Valhalla: ValhallaConfig{Enabled: false}}, fakeMatcher{})
	if err != nil {
		t.Fatalf("expected nil err when disabled, got %v", err)
	}
}

func TestValidateMapMatcherEnabledCallsValidate(t *testing.T) {
	err := validateMapMatcher(EngineConfig{Valhalla: ValhallaConfig{Enabled: true}}, fakeMatcher{})
	if err == nil || err.Error() != "validate-called" {
		t.Fatalf("expected validate error, got %v", err)
	}
}
