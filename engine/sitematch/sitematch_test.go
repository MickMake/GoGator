package sitematch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoopSiteMatcherReturnsEmptyWithoutError(t *testing.T) {
	matcher := NoopSiteMatcher{}
	res, err := matcher.Match(SiteMatchRequest{Latitude: -33, Longitude: 151, RadiusMeters: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(res.Candidates))
	}
}

func TestMemorySiteMatcherMatchesWithinRadius(t *testing.T) {
	matcher := NewMemorySiteMatcher([]MemorySite{{Name: "Home", Address: "A", Lat: -33, Lng: 151}})
	res, err := matcher.Match(SiteMatchRequest{Latitude: -33.0001, Longitude: 151.0001, RadiusMeters: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Candidates) != 1 {
		t.Fatalf("expected one candidate, got %d", len(res.Candidates))
	}
	if res.Candidates[0].SiteName != "Home" {
		t.Fatalf("expected Home, got %q", res.Candidates[0].SiteName)
	}
}

func TestMemorySiteMatcherOutsideRadiusNoMatch(t *testing.T) {
	matcher := NewMemorySiteMatcher([]MemorySite{{Name: "Home", Address: "A", Lat: -33, Lng: 151}})
	res, err := matcher.Match(SiteMatchRequest{Latitude: -34, Longitude: 152, RadiusMeters: 50})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Candidates) != 0 {
		t.Fatalf("expected no candidates, got %d", len(res.Candidates))
	}
}

func TestPostGISScaffoldValidationAndUnsupported(t *testing.T) {
	if _, err := NewPostGISSiteMatcher(PostGISConfig{Enabled: true, DSN: "", MatchRadiusMeters: 100}); err == nil || !strings.Contains(err.Error(), "dsn is empty") {
		t.Fatalf("expected dsn validation error, got %v", err)
	}
	m, err := NewPostGISSiteMatcher(PostGISConfig{Enabled: true, DSN: "postgres://example", MatchRadiusMeters: 100})
	if err != nil {
		t.Fatalf("unexpected constructor error: %v", err)
	}
	if _, err := m.Match(SiteMatchRequest{Latitude: -33, Longitude: 151, RadiusMeters: 100}); err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("expected not implemented error, got %v", err)
	}
}

func TestSchemaSQLExists(t *testing.T) {
	path := filepath.Join("schema.sql")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected schema.sql to exist: %v", err)
	}
	if !strings.Contains(string(b), "postgis_sites") {
		t.Fatalf("expected postgis_sites scaffold in schema.sql")
	}
}
