package engine

import (
	"strings"
	"testing"

	"gogator/engine/sitematch"
)

func TestNewSiteMatcherDisabledUsesNoop(t *testing.T) {
	m, err := newSiteMatcher(EngineConfig{PostGIS: PostGISConfig{Enabled: false}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := m.(sitematch.NoopSiteMatcher); !ok {
		t.Fatalf("expected noop matcher, got %T", m)
	}
}

func TestNewSiteMatcherEnabledReturnsClearErrorWhenScaffoldOnly(t *testing.T) {
	m, err := newSiteMatcher(EngineConfig{PostGIS: PostGISConfig{Enabled: true, DSN: "postgres://example", MatchRadiusMeters: 100}})
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	_, err = m.Match(sitematch.SiteMatchRequest{Latitude: -33, Longitude: 151, RadiusMeters: 100})
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("expected scaffold not implemented error, got %v", err)
	}
}
