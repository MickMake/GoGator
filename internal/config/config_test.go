package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWithoutNewSectionsUsesSafeDefaults(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "gogator.yaml")
	data := "timezone: Australia/Sydney\ntrip_detection:\n  min_stop_duration_seconds: 60\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Engine.Enabled || !cfg.Engine.CompatibilityMode {
		t.Fatalf("engine defaults changed: %+v", cfg.Engine)
	}
	if cfg.Engine.TripSource != "legacy" {
		t.Fatalf("expected default engine.trip_source legacy, got %q", cfg.Engine.TripSource)
	}
	if !cfg.Engine.EngineMode.RequireMinReadiness || cfg.Engine.EngineMode.MinReadiness != "GoodMatch" {
		t.Fatalf("expected conservative engine_mode defaults, got %+v", cfg.Engine.EngineMode)
	}
	if cfg.Engine.Audit.Enabled || cfg.Valhalla.Enabled || cfg.H3.Enabled || cfg.PostGIS.Enabled {
		t.Fatalf("future integrations should be disabled by default: %+v %+v %+v %+v", cfg.Engine.Audit, cfg.Valhalla, cfg.H3, cfg.PostGIS)
	}
}

func TestLoadNewSectionsOverrideDefaults(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "gogator.yaml")
	data := "engine:\n  enabled: true\n  compatibility_mode: true\n  stay_detection:\n    enabled: true\n  motion:\n    enabled: true\n  quality:\n    enabled: true\n  audit:\n    enabled: true\nvalhalla:\n  enabled: true\nh3:\n  enabled: true\npostgis:\n  enabled: true\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Engine.StayDetection.Enabled || !cfg.Engine.Motion.Enabled || !cfg.Engine.Quality.Enabled || !cfg.Engine.Audit.Enabled {
		t.Fatalf("engine nested sections not loaded: %+v", cfg.Engine)
	}
	if !cfg.Valhalla.Enabled || !cfg.H3.Enabled || !cfg.PostGIS.Enabled {
		t.Fatalf("integration toggles not loaded: %+v %+v %+v", cfg.Valhalla, cfg.H3, cfg.PostGIS)
	}
}

func TestLoadNestedSectionSiblingResetsToParentSection(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "gogator.yaml")
	data := "engine:\n  stay_detection:\n    enabled: true\n  compatibility_mode: false\n  trip_source: shadow\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Engine.CompatibilityMode {
		t.Fatalf("expected compatibility_mode=false, got true")
	}
	if cfg.Engine.TripSource != "shadow" {
		t.Fatalf("expected trip_source=shadow, got %q", cfg.Engine.TripSource)
	}
}

func TestLoadPostGISScaffoldFields(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "gogator.yaml")
	data := `postgis:
  enabled: true
  dsn: postgresql://user:pass@localhost/db
  match_radius_meters: 250
  audit_enabled: true
`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.PostGIS.DSN == "" || cfg.PostGIS.MatchRadiusMeters != 250 || !cfg.PostGIS.AuditEnabled {
		t.Fatalf("expected postgis scaffold fields loaded, got %+v", cfg.PostGIS)
	}
}
