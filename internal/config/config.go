package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Timezone string
	Sites    string
	Routes   string
	Trip     TripDetection
	Site     SiteMatching
	RawTime  RawTime
	Engine   Engine
	Valhalla Valhalla
	H3       H3
	PostGIS  PostGIS
}

type Engine struct {
	Enabled           bool
	CompatibilityMode bool
	StayDetection     EngineStayDetection
	Visits            EngineVisits
	Excursions        EngineExcursions
	TripBuilder       EngineTripBuilder
	Shadow            EngineShadow
	Motion            EngineMotion
	Quality           EngineQuality
	Audit             EngineAudit
}

type EngineStayDetection struct {
	Enabled                bool
	MinDurationMinutes     float64
	MaxRadiusMeters        float64
	MinPoints              int
	SiteMatchRadiusMeters  float64
	GapInferredStopEnabled bool
}
type EngineVisits struct {
	Enabled                 bool
	MinVisitDurationMinutes float64
}
type EngineExcursions struct {
	Enabled                          bool
	ShortOutAndBackMaxMinutes        float64
	ShortOutAndBackMaxDistanceMeters float64
}
type EngineTripBuilder struct {
	Enabled                bool
	PassiveOnly            bool
	CompareLegacy          bool
	MinTripDurationMinutes float64
	MaxGapMinutes          float64
	LowConfidenceThreshold float64
}
type EngineShadow struct {
	Enabled                        bool
	SummaryEnabled                 bool
	MatchToleranceMinutes          float64
	GoodMatchThresholdPercent      float64
	ExcellentMatchThresholdPercent float64
	WarnOnMajorMismatch            bool
}
type EngineMotion struct {
	Enabled                     bool
	StationarySpeedThresholdKPH float64
	MovingSpeedThresholdKPH     float64
	GapThresholdMinutes         float64
	MinConsecutiveSamples       int
}
type EngineQuality struct{ Enabled bool }
type EngineAudit struct {
	Enabled                bool
	OutputDiagnostics      bool
	OutputPoints           bool
	OutputMotion           bool
	OutputStays            bool
	OutputVisits           bool
	OutputExcursions       bool
	OutputCandidateTrips   bool
	OutputTripComparison   bool
	OutputShadowSummary    bool
	OutputShadowMismatches bool
}
type Valhalla struct{ Enabled bool }
type H3 struct{ Enabled bool }
type PostGIS struct{ Enabled bool }

type TripDetection struct {
	MovingSpeedKPH                     float64
	MinTripDistanceM                   float64
	MinTripDurationSeconds             int
	MinStopDurationSeconds             int
	PoorPDOPThreshold                  float64
	IdealPDOPThreshold                 float64
	SameSiteJitterRadiusM              float64
	AccelerometerNonzeroSupportsMotion bool
	IgnitionAnalysis                   string
	StationaryTeleportGuardEnabled     bool
	StationaryTeleportMinJumpM         float64
	StationaryTeleportRequiresOdometer bool
	SameSiteMicroTripMaxKM             float64
	SameSiteMicroTripMaxMinutes        float64
	SameSiteGuardRadiusM               float64
}

type SiteMatching struct {
	DefaultRadiusM                float64
	DefaultMinDestinationMinutes  float64
	UnknownCheckMinDestinationMin float64
	StationaryDwellRatioRequired  float64
	DwellWindowMinutes            float64
	DwellRequiredInsideRatio      float64
	DwellRequiredStationaryRatio  float64
	DwellMaxSampleGapMinutes      float64
	ContinuityRepairEnabled       bool
	ContinuityMatchMaxMetres      float64
	InferSilentStopGaps           bool
	SilentStopMinGapMinutes       float64
	UnknownSiteLabel              string
	HomeSiteName                  string
}

type RawTime struct {
	Source          string
	CorrectionHours float64
}

func Default() Config {
	return Config{
		Timezone: "Australia/Sydney",
		Sites:    "sites.csv",
		Routes:   "routes.csv",
		Trip: TripDetection{
			MovingSpeedKPH:                     7,
			MinTripDistanceM:                   150,
			MinTripDurationSeconds:             60,
			MinStopDurationSeconds:             300,
			PoorPDOPThreshold:                  5,
			IdealPDOPThreshold:                 2,
			SameSiteJitterRadiusM:              100,
			AccelerometerNonzeroSupportsMotion: false,
			IgnitionAnalysis:                   "auto",
			StationaryTeleportGuardEnabled:     true,
			StationaryTeleportMinJumpM:         250,
			StationaryTeleportRequiresOdometer: true,
			SameSiteMicroTripMaxKM:             1.5,
			SameSiteMicroTripMaxMinutes:        10,
			SameSiteGuardRadiusM:               750,
		},
		Site: SiteMatching{
			DefaultRadiusM:                100,
			DefaultMinDestinationMinutes:  5,
			UnknownCheckMinDestinationMin: 10,
			StationaryDwellRatioRequired:  0.70,
			DwellWindowMinutes:            180,
			DwellRequiredInsideRatio:      0.70,
			DwellRequiredStationaryRatio:  0.70,
			DwellMaxSampleGapMinutes:      90,
			ContinuityRepairEnabled:       true,
			ContinuityMatchMaxMetres:      75,
			InferSilentStopGaps:           true,
			SilentStopMinGapMinutes:       5,
			UnknownSiteLabel:              "CHECK",
			HomeSiteName:                  "Home",
		},
		RawTime: RawTime{
			Source:          "gator_raw_utc",
			CorrectionHours: 0,
		},
		Engine: Engine{
			Enabled:           true,
			CompatibilityMode: true,
			StayDetection: EngineStayDetection{
				Enabled:                false,
				MinDurationMinutes:     5,
				MaxRadiusMeters:        120,
				MinPoints:              3,
				SiteMatchRadiusMeters:  100,
				GapInferredStopEnabled: true,
			},
			Visits:      EngineVisits{Enabled: false, MinVisitDurationMinutes: 5},
			Excursions:  EngineExcursions{Enabled: false, ShortOutAndBackMaxMinutes: 20, ShortOutAndBackMaxDistanceMeters: 5000},
			TripBuilder: EngineTripBuilder{Enabled: false, PassiveOnly: true, CompareLegacy: false, MinTripDurationMinutes: 1, MaxGapMinutes: 20, LowConfidenceThreshold: 0.4},
			Shadow:      EngineShadow{Enabled: false, SummaryEnabled: false, MatchToleranceMinutes: 20, GoodMatchThresholdPercent: 70, ExcellentMatchThresholdPercent: 90, WarnOnMajorMismatch: false},
			Motion: EngineMotion{
				Enabled:                     false,
				StationarySpeedThresholdKPH: 2,
				MovingSpeedThresholdKPH:     8,
				GapThresholdMinutes:         20,
				MinConsecutiveSamples:       2,
			},
			Quality: EngineQuality{Enabled: false},
			Audit:   EngineAudit{Enabled: false, OutputDiagnostics: false, OutputPoints: false, OutputMotion: false, OutputStays: false, OutputVisits: false, OutputExcursions: false, OutputCandidateTrips: false, OutputTripComparison: false, OutputShadowSummary: false, OutputShadowMismatches: false},
		},
		Valhalla: Valhalla{Enabled: false},
		H3:       H3{Enabled: false},
		PostGIS:  PostGIS{Enabled: false},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, err
	}
	defer f.Close()

	sectionByLevel := map[int]string{}
	section := ""
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if i := strings.Index(line, "#"); i >= 0 {
			line = line[:i]
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ":") && indent == 0 {
			section = strings.TrimSuffix(line, ":")
			sectionByLevel = map[int]string{0: section}
			continue
		} else if strings.HasSuffix(line, ":") {
			level := indent / 2
			for k := range sectionByLevel {
				if k > level {
					delete(sectionByLevel, k)
				}
			}
			sectionByLevel[level] = strings.TrimSuffix(line, ":")
			parts := make([]string, 0, level+1)
			for i := 0; i <= level; i++ {
				if s, ok := sectionByLevel[i]; ok && s != "" {
					parts = append(parts, s)
				}
			}
			section = strings.Join(parts, ".")
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		if indent == 0 {
			section = ""
		} else {
			level := indent / 2
			parts := make([]string, 0, level+1)
			for i := 0; i < level; i++ {
				if s, ok := sectionByLevel[i]; ok && s != "" {
					parts = append(parts, s)
				}
			}
			section = strings.Join(parts, ".")
		}
		apply(&cfg, section, key, val)
	}
	return cfg, scanner.Err()
}

func apply(cfg *Config, section, key, val string) {
	switch section {
	case "":
		switch key {
		case "timezone":
			cfg.Timezone = val
		case "sites":
			cfg.Sites = val
		case "routes":
			cfg.Routes = val
		}
	case "trip_detection":
		switch key {
		case "moving_speed_kph":
			cfg.Trip.MovingSpeedKPH = f(val, cfg.Trip.MovingSpeedKPH)
		case "min_trip_distance_m":
			cfg.Trip.MinTripDistanceM = f(val, cfg.Trip.MinTripDistanceM)
		case "min_trip_duration_seconds":
			cfg.Trip.MinTripDurationSeconds = i(val, cfg.Trip.MinTripDurationSeconds)
		case "min_stop_duration_seconds":
			cfg.Trip.MinStopDurationSeconds = i(val, cfg.Trip.MinStopDurationSeconds)
		case "poor_pdop_threshold":
			cfg.Trip.PoorPDOPThreshold = f(val, cfg.Trip.PoorPDOPThreshold)
		case "ideal_pdop_threshold":
			cfg.Trip.IdealPDOPThreshold = f(val, cfg.Trip.IdealPDOPThreshold)
		case "same_site_jitter_radius_m":
			cfg.Trip.SameSiteJitterRadiusM = f(val, cfg.Trip.SameSiteJitterRadiusM)
		case "accelerometer_nonzero_supports_motion":
			cfg.Trip.AccelerometerNonzeroSupportsMotion = b(val, cfg.Trip.AccelerometerNonzeroSupportsMotion)
		case "ignition_analysis":
			cfg.Trip.IgnitionAnalysis = ignitionMode(val, cfg.Trip.IgnitionAnalysis)
		case "stationary_teleport_guard_enabled":
			cfg.Trip.StationaryTeleportGuardEnabled = b(val, cfg.Trip.StationaryTeleportGuardEnabled)
		case "stationary_teleport_min_jump_m":
			cfg.Trip.StationaryTeleportMinJumpM = f(val, cfg.Trip.StationaryTeleportMinJumpM)
		case "stationary_teleport_requires_odometer":
			cfg.Trip.StationaryTeleportRequiresOdometer = b(val, cfg.Trip.StationaryTeleportRequiresOdometer)
		case "same_site_micro_trip_max_km":
			cfg.Trip.SameSiteMicroTripMaxKM = f(val, cfg.Trip.SameSiteMicroTripMaxKM)
		case "same_site_micro_trip_max_minutes":
			cfg.Trip.SameSiteMicroTripMaxMinutes = f(val, cfg.Trip.SameSiteMicroTripMaxMinutes)
		case "same_site_guard_radius_m":
			cfg.Trip.SameSiteGuardRadiusM = f(val, cfg.Trip.SameSiteGuardRadiusM)
		}
	case "site_matching":
		switch key {
		case "default_radius_m":
			cfg.Site.DefaultRadiusM = f(val, cfg.Site.DefaultRadiusM)
		case "default_min_destination_minutes":
			cfg.Site.DefaultMinDestinationMinutes = f(val, cfg.Site.DefaultMinDestinationMinutes)
		case "unknown_check_min_destination_minutes":
			cfg.Site.UnknownCheckMinDestinationMin = f(val, cfg.Site.UnknownCheckMinDestinationMin)
		case "stationary_dwell_ratio_required":
			cfg.Site.StationaryDwellRatioRequired = f(val, cfg.Site.StationaryDwellRatioRequired)
		case "dwell_window_minutes":
			cfg.Site.DwellWindowMinutes = f(val, cfg.Site.DwellWindowMinutes)
		case "dwell_required_inside_ratio":
			cfg.Site.DwellRequiredInsideRatio = f(val, cfg.Site.DwellRequiredInsideRatio)
		case "dwell_required_stationary_ratio":
			cfg.Site.DwellRequiredStationaryRatio = f(val, cfg.Site.DwellRequiredStationaryRatio)
		case "dwell_max_sample_gap_minutes":
			cfg.Site.DwellMaxSampleGapMinutes = f(val, cfg.Site.DwellMaxSampleGapMinutes)
		case "continuity_repair_enabled":
			cfg.Site.ContinuityRepairEnabled = b(val, cfg.Site.ContinuityRepairEnabled)
		case "continuity_match_max_metres", "continuity_match_max_meters":
			cfg.Site.ContinuityMatchMaxMetres = f(val, cfg.Site.ContinuityMatchMaxMetres)
		case "infer_silent_stop_gaps":
			cfg.Site.InferSilentStopGaps = b(val, cfg.Site.InferSilentStopGaps)
		case "silent_stop_min_gap_minutes":
			cfg.Site.SilentStopMinGapMinutes = f(val, cfg.Site.SilentStopMinGapMinutes)
		case "unknown_site_label":
			cfg.Site.UnknownSiteLabel = val
		case "home_site_name":
			cfg.Site.HomeSiteName = val
		}
	case "raw_time":
		switch key {
		case "source":
			cfg.RawTime.Source = val
		case "correction_hours":
			cfg.RawTime.CorrectionHours = f(val, cfg.RawTime.CorrectionHours)
		}
	case "engine":
		switch key {
		case "enabled":
			cfg.Engine.Enabled = b(val, cfg.Engine.Enabled)
		case "compatibility_mode":
			cfg.Engine.CompatibilityMode = b(val, cfg.Engine.CompatibilityMode)
		}
	case "engine.stay_detection":
		switch key {
		case "enabled":
			cfg.Engine.StayDetection.Enabled = b(val, cfg.Engine.StayDetection.Enabled)
		case "min_duration_minutes":
			cfg.Engine.StayDetection.MinDurationMinutes = f(val, cfg.Engine.StayDetection.MinDurationMinutes)
		case "max_radius_meters", "max_radius_metres":
			cfg.Engine.StayDetection.MaxRadiusMeters = f(val, cfg.Engine.StayDetection.MaxRadiusMeters)
		case "min_points":
			cfg.Engine.StayDetection.MinPoints = i(val, cfg.Engine.StayDetection.MinPoints)
		case "site_match_radius_meters", "site_match_radius_metres":
			cfg.Engine.StayDetection.SiteMatchRadiusMeters = f(val, cfg.Engine.StayDetection.SiteMatchRadiusMeters)
		case "gap_inferred_stop_enabled":
			cfg.Engine.StayDetection.GapInferredStopEnabled = b(val, cfg.Engine.StayDetection.GapInferredStopEnabled)
		}

	case "engine.visits":
		switch key {
		case "enabled":
			cfg.Engine.Visits.Enabled = b(val, cfg.Engine.Visits.Enabled)
		case "min_visit_duration_minutes":
			cfg.Engine.Visits.MinVisitDurationMinutes = f(val, cfg.Engine.Visits.MinVisitDurationMinutes)
		}
	case "engine.excursions":
		switch key {
		case "enabled":
			cfg.Engine.Excursions.Enabled = b(val, cfg.Engine.Excursions.Enabled)
		case "short_out_and_back_max_minutes":
			cfg.Engine.Excursions.ShortOutAndBackMaxMinutes = f(val, cfg.Engine.Excursions.ShortOutAndBackMaxMinutes)
		case "short_out_and_back_max_distance_meters", "short_out_and_back_max_distance_metres":
			cfg.Engine.Excursions.ShortOutAndBackMaxDistanceMeters = f(val, cfg.Engine.Excursions.ShortOutAndBackMaxDistanceMeters)
		}
	case "engine.trip_builder":
		switch key {
		case "enabled":
			cfg.Engine.TripBuilder.Enabled = b(val, cfg.Engine.TripBuilder.Enabled)
		case "passive_only":
			cfg.Engine.TripBuilder.PassiveOnly = b(val, cfg.Engine.TripBuilder.PassiveOnly)
		case "compare_legacy":
			cfg.Engine.TripBuilder.CompareLegacy = b(val, cfg.Engine.TripBuilder.CompareLegacy)
		case "min_trip_duration_minutes":
			cfg.Engine.TripBuilder.MinTripDurationMinutes = f(val, cfg.Engine.TripBuilder.MinTripDurationMinutes)
		case "max_gap_minutes":
			cfg.Engine.TripBuilder.MaxGapMinutes = f(val, cfg.Engine.TripBuilder.MaxGapMinutes)
		case "low_confidence_threshold":
			cfg.Engine.TripBuilder.LowConfidenceThreshold = f(val, cfg.Engine.TripBuilder.LowConfidenceThreshold)
		}

	case "engine.shadow":
		switch key {
		case "enabled":
			cfg.Engine.Shadow.Enabled = b(val, cfg.Engine.Shadow.Enabled)
		case "summary_enabled":
			cfg.Engine.Shadow.SummaryEnabled = b(val, cfg.Engine.Shadow.SummaryEnabled)
		case "match_tolerance_minutes":
			cfg.Engine.Shadow.MatchToleranceMinutes = f(val, cfg.Engine.Shadow.MatchToleranceMinutes)
		case "good_match_threshold_percent":
			cfg.Engine.Shadow.GoodMatchThresholdPercent = f(val, cfg.Engine.Shadow.GoodMatchThresholdPercent)
		case "excellent_match_threshold_percent":
			cfg.Engine.Shadow.ExcellentMatchThresholdPercent = f(val, cfg.Engine.Shadow.ExcellentMatchThresholdPercent)
		case "warn_on_major_mismatch":
			cfg.Engine.Shadow.WarnOnMajorMismatch = b(val, cfg.Engine.Shadow.WarnOnMajorMismatch)
		}

	case "engine.motion":
		switch key {
		case "enabled":
			cfg.Engine.Motion.Enabled = b(val, cfg.Engine.Motion.Enabled)
		case "stationary_speed_threshold_kmh":
			cfg.Engine.Motion.StationarySpeedThresholdKPH = f(val, cfg.Engine.Motion.StationarySpeedThresholdKPH)
		case "moving_speed_threshold_kmh":
			cfg.Engine.Motion.MovingSpeedThresholdKPH = f(val, cfg.Engine.Motion.MovingSpeedThresholdKPH)
		case "gap_threshold_minutes":
			cfg.Engine.Motion.GapThresholdMinutes = f(val, cfg.Engine.Motion.GapThresholdMinutes)
		case "min_consecutive_samples":
			cfg.Engine.Motion.MinConsecutiveSamples = i(val, cfg.Engine.Motion.MinConsecutiveSamples)
		}
	case "engine.quality":
		if key == "enabled" {
			cfg.Engine.Quality.Enabled = b(val, cfg.Engine.Quality.Enabled)
		}
	case "engine.audit":
		switch key {
		case "enabled":
			cfg.Engine.Audit.Enabled = b(val, cfg.Engine.Audit.Enabled)
		case "output_diagnostics":
			cfg.Engine.Audit.OutputDiagnostics = b(val, cfg.Engine.Audit.OutputDiagnostics)
		case "output_points":
			cfg.Engine.Audit.OutputPoints = b(val, cfg.Engine.Audit.OutputPoints)
		case "output_motion":
			cfg.Engine.Audit.OutputMotion = b(val, cfg.Engine.Audit.OutputMotion)
		case "output_stays":
			cfg.Engine.Audit.OutputStays = b(val, cfg.Engine.Audit.OutputStays)
		case "output_visits":
			cfg.Engine.Audit.OutputVisits = b(val, cfg.Engine.Audit.OutputVisits)
		case "output_excursions":
			cfg.Engine.Audit.OutputExcursions = b(val, cfg.Engine.Audit.OutputExcursions)
		case "output_candidate_trips":
			cfg.Engine.Audit.OutputCandidateTrips = b(val, cfg.Engine.Audit.OutputCandidateTrips)
		case "output_trip_comparison":
			cfg.Engine.Audit.OutputTripComparison = b(val, cfg.Engine.Audit.OutputTripComparison)
		case "output_shadow_summary":
			cfg.Engine.Audit.OutputShadowSummary = b(val, cfg.Engine.Audit.OutputShadowSummary)
		case "output_shadow_mismatches":
			cfg.Engine.Audit.OutputShadowMismatches = b(val, cfg.Engine.Audit.OutputShadowMismatches)
		}
	case "valhalla":
		if key == "enabled" {
			cfg.Valhalla.Enabled = b(val, cfg.Valhalla.Enabled)
		}
	case "h3":
		if key == "enabled" {
			cfg.H3.Enabled = b(val, cfg.H3.Enabled)
		}
	case "postgis":
		if key == "enabled" {
			cfg.PostGIS.Enabled = b(val, cfg.PostGIS.Enabled)
		}
	}
}

func f(s string, d float64) float64 {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return d
	}
	return v
}
func i(s string, d int) int {
	v, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return d
	}
	return v
}
func b(s string, d bool) bool {
	v, err := strconv.ParseBool(strings.ToLower(strings.TrimSpace(s)))
	if err != nil {
		return d
	}
	return v
}
func ignitionMode(s, d string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "auto", "wired", "unwired":
		return s
	default:
		return d
	}
}
