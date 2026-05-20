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
}

type TripDetection struct {
	MovingSpeedKPH                     float64
	MinTripDistanceM                   float64
	MinTripDurationSeconds             int
	MinStopDurationSeconds             int
	PoorPDOPThreshold                  float64
	IdealPDOPThreshold                 float64
	SameSiteJitterRadiusM              float64
	AccelerometerNonzeroSupportsMotion bool
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
		Sites:    "addresses.csv",
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
			InferSilentStopGaps:           true,
			SilentStopMinGapMinutes:       5,
			UnknownSiteLabel:              "CHECK",
			HomeSiteName:                  "Home",
		},
		RawTime: RawTime{
			Source:          "gator_raw_utc_plus_one_day",
			CorrectionHours: -24,
		},
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
