package app

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"context"

	"gogator/engine"
	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/output"
	"gogator/internal/routes"
	"gogator/internal/sites"
	"gogator/internal/store"
)

type Options struct {
	Input      string
	Inputs     []string
	ConfigPath string
	SitesPath  string
	RoutesPath string
	Timezone   string
}

type processResult struct {
	Source            string
	SitesSource       string
	RoutesSource      string
	ConfigPath        string
	Timezone          string
	Points            []gps.RawPoint
	Valid             []gps.Trip
	Jitter            []gps.Trip
	JitterReview      []gps.Trip
	JitterSameSite    []gps.Trip
	RouteObservations []routes.Observation
	RouteAnomalies    []routes.Anomaly
	SiteCount         int
	RouteCount        int
}

type processOutputPaths struct {
	Expanded          string
	Processed         string
	Jitter            string
	JitterSameSite    string
	Audit             string
	RouteObservations string
	RouteAnomalies    string
}

func RunProcess(opts Options) error {
	if len(opts.Inputs) == 0 && opts.Input != "" {
		opts.Inputs = []string{opts.Input}
	}
	if len(opts.Inputs) == 0 {
		return fmt.Errorf("missing input CSV")
	}
	cfg, err := loadProcessConfig(&opts)
	if err != nil {
		return err
	}
	if math.Abs(cfg.RawTime.CorrectionHours) > 0.000001 {
		fmt.Fprintf(os.Stderr, "warning: raw_time.correction_hours=%.2f shifts raw tracker timestamps before local date grouping; use only for known malformed exports\n", cfg.RawTime.CorrectionHours)
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("timezone %q: %w", cfg.Timezone, err)
	}

	return runProcessCombined(opts, cfg, loc)
}

func RunProcessGPS(opts Options) error {
	cfg, err := loadProcessConfig(&opts)
	if err != nil {
		return err
	}
	if math.Abs(cfg.RawTime.CorrectionHours) > 0.000001 {
		fmt.Fprintf(os.Stderr, "warning: raw_time.correction_hours=%.2f is ignored for already-loaded SQLite GPS points\n", cfg.RawTime.CorrectionHours)
	}

	points, err := store.LoadGPSPointsForProcess()
	if err != nil {
		return fmt.Errorf("load gps points from database: %w", err)
	}
	if len(points) == 0 {
		return fmt.Errorf("no GPS points found in database; run: gogator load gator from <file>")
	}

	siteList, err := store.LoadSitesForProcess(cfg)
	if err != nil {
		return fmt.Errorf("load sites from database: %w", err)
	}
	if len(siteList) == 0 {
		fmt.Fprintf(os.Stderr, "warning: loaded 0 sites from database; all sites will be CHECK\n")
	} else {
		fmt.Fprintf(os.Stderr, "loaded sites: %d from database\n", len(siteList))
	}

	routeRules, err := store.LoadRoutesForProcess()
	if err != nil {
		return fmt.Errorf("load routes from database: %w", err)
	}
	if len(routeRules) == 0 {
		fmt.Fprintf(os.Stderr, "loaded routes: 0 from database; observations will still be generated\n")
	} else {
		fmt.Fprintf(os.Stderr, "loaded routes: %d from database\n", len(routeRules))
	}

	res := runProcessPipeline(points, siteList, routeRules, cfg)
	res.Source = "gogator.sqlite"
	res.SitesSource = "database"
	res.RoutesSource = "database"
	res.ConfigPath = opts.ConfigPath
	res.Timezone = cfg.Timezone

	paths := processPaths("gogator")
	if err := writeProcessOutputs(paths, res); err != nil {
		return err
	}
	printProcessSummary(res)
	printProcessWrites(paths)
	printProcessErrors(res)
	return nil
}

func loadProcessConfig(opts *Options) (config.Config, error) {
	if opts.ConfigPath == "" {
		opts.ConfigPath = DefaultConfigPath()
	}
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return cfg, err
	}
	if envTZ := os.Getenv("GOGATOR_TIMEZONE"); envTZ != "" {
		cfg.Timezone = envTZ
	}
	if opts.Timezone != "" {
		cfg.Timezone = opts.Timezone
	}
	if opts.SitesPath != "" {
		cfg.Sites = opts.SitesPath
	}
	if opts.RoutesPath != "" {
		cfg.Routes = opts.RoutesPath
	}
	if cfg.Sites == "" {
		cfg.Sites = "sites.csv"
	}
	if cfg.Routes == "" {
		cfg.Routes = "routes.csv"
	}
	return cfg, nil
}

func runProcessCombined(opts Options, cfg config.Config, loc *time.Location) error {
	primaryInput := opts.Inputs[0]
	sitesPath := resolveSiblingPath(cfg.Sites, primaryInput)
	siteList, err := sites.Load(sitesPath, cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load sites %s: %v; all sites will be CHECK\n", sitesPath, err)
	}
	if len(siteList) == 0 {
		fmt.Fprintf(os.Stderr, "warning: loaded 0 sites from %s; all sites will be CHECK\n", sitesPath)
	} else {
		fmt.Fprintf(os.Stderr, "loaded sites: %d from %s\n", len(siteList), sitesPath)
	}

	routesPath := resolveSiblingPath(cfg.Routes, primaryInput)
	routeRules, err := routes.Load(routesPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not load routes %s: %v; route rules will be skipped\n", routesPath, err)
	}
	if len(routeRules) == 0 {
		fmt.Fprintf(os.Stderr, "loaded routes: 0 from %s; observations will still be generated\n", routesPath)
	} else {
		fmt.Fprintf(os.Stderr, "loaded routes: %d from %s\n", len(routeRules), routesPath)
	}

	var points []gps.RawPoint
	for _, input := range opts.Inputs {
		if len(opts.Inputs) > 1 {
			fmt.Printf("loading: %s\n", input)
		}
		pts, err := gps.ReadRawCSV(input, loc, cfg)
		if err != nil {
			return err
		}
		points = append(points, pts...)
	}
	res := runProcessPipeline(points, siteList, routeRules, cfg)
	res.Source = strings.Join(opts.Inputs, ";")
	res.SitesSource = sitesPath
	res.RoutesSource = routesPath
	res.ConfigPath = opts.ConfigPath
	res.Timezone = cfg.Timezone

	prefix := output.Prefix(primaryInput)
	if len(opts.Inputs) > 1 {
		prefix = filepath.Join(filepath.Dir(primaryInput), commonOutputName(opts.Inputs))
	}
	paths := processPaths(prefix)
	if err := writeProcessOutputs(paths, res); err != nil {
		return err
	}

	printProcessSummary(res)
	printProcessWrites(paths)
	return nil
}

func processPaths(prefix string) processOutputPaths {
	return processOutputPaths{
		Expanded:          prefix + "_expanded.csv",
		Processed:         prefix + "_processed.csv",
		Jitter:            prefix + "_jitter.csv",
		JitterSameSite:    prefix + "_jitter_same_site.csv",
		Audit:             prefix + "_audit.csv",
		RouteObservations: prefix + "_route_observations.csv",
		RouteAnomalies:    prefix + "_route_anomalies.csv",
	}
}

func writeProcessOutputs(paths processOutputPaths, res processResult) error {
	if err := output.WriteExpanded(paths.Expanded, res.Points); err != nil {
		return err
	}
	if err := output.WriteTrips(paths.Processed, res.Valid); err != nil {
		return err
	}
	if err := output.WriteTrips(paths.Jitter, res.JitterReview); err != nil {
		return err
	}
	if err := output.WriteTrips(paths.JitterSameSite, res.JitterSameSite); err != nil {
		return err
	}
	if err := output.WriteRouteObservations(paths.RouteObservations, res.RouteObservations); err != nil {
		return err
	}
	if err := output.WriteRouteAnomalies(paths.RouteAnomalies, res.RouteAnomalies); err != nil {
		return err
	}
	return output.WriteAudit(paths.Audit, res.Source, res.SitesSource, res.RoutesSource, res.ConfigPath, res.Timezone, len(res.Points), len(res.Valid), len(res.Jitter), res.SiteCount, res.RouteCount)
}

func printProcessSummary(res processResult) {
	fmt.Printf("processed raw points: %d\n", len(res.Points))
	fmt.Printf("valid trips: %d\n", len(res.Valid))
	fmt.Printf("rejected jitter: %d\n", len(res.Jitter))
	fmt.Printf("same-site jitter: %d\n", len(res.JitterSameSite))
}

func printProcessWrites(paths processOutputPaths) {
	fmt.Printf("wrote: %s\n", paths.Processed)
	fmt.Printf("wrote: %s\n", paths.Expanded)
	fmt.Printf("wrote: %s\n", paths.Jitter)
	fmt.Printf("wrote: %s\n", paths.JitterSameSite)
	fmt.Printf("wrote: %s\n", paths.RouteObservations)
	fmt.Printf("wrote: %s\n", paths.RouteAnomalies)
	fmt.Printf("wrote: %s\n", paths.Audit)
}

func printProcessErrors(res processResult) {
	if len(res.RouteAnomalies) == 0 {
		fmt.Printf("errors encountered: 0\n")
		return
	}
	fmt.Printf("errors encountered: %d\n", len(res.RouteAnomalies))
	for _, a := range res.RouteAnomalies {
		fmt.Printf("error: trip=%d from=%s to=%s status=%s notes=%s raw_rows=%d-%d\n", a.TripIndex, a.FromSite, a.ToSite, a.Status, a.Notes, a.RawStartRow, a.RawEndRow)
	}
}

func runProcessPipeline(points []gps.RawPoint, siteList []sites.Site, routeRules []routes.Route, cfg config.Config) processResult {
	result, err := engine.Run(context.Background(), engine.Input{
		Points:       points,
		Sites:        siteList,
		Routes:       routeRules,
		Config:       cfg,
		EngineConfig: buildEngineConfig(cfg),
	})
	if err != nil {
		return processResult{}
	}
	return processResult{
		Points:            result.Points,
		Valid:             result.Valid,
		Jitter:            result.Jitter,
		JitterReview:      result.JitterReview,
		JitterSameSite:    result.JitterSameSite,
		RouteObservations: result.RouteObservations,
		RouteAnomalies:    result.RouteAnomalies,
		SiteCount:         result.SiteCount,
		RouteCount:        result.RouteCount,
	}
}

func buildEngineConfig(cfg config.Config) engine.EngineConfig {
	return engine.EngineConfig{
		Enabled:           cfg.Engine.Enabled,
		CompatibilityMode: cfg.Engine.CompatibilityMode,
		StayDetection: engine.StayConfig{
			Enabled:                cfg.Engine.StayDetection.Enabled,
			MinDurationMinutes:     cfg.Engine.StayDetection.MinDurationMinutes,
			MaxRadiusMeters:        cfg.Engine.StayDetection.MaxRadiusMeters,
			MinPoints:              cfg.Engine.StayDetection.MinPoints,
			SiteMatchRadiusMeters:  cfg.Engine.StayDetection.SiteMatchRadiusMeters,
			GapInferredStopEnabled: cfg.Engine.StayDetection.GapInferredStopEnabled,
		},
		Visits:     engine.VisitConfig{Enabled: cfg.Engine.Visits.Enabled, MinVisitDurationMinutes: cfg.Engine.Visits.MinVisitDurationMinutes},
		Excursions: engine.ExcursionConfig{Enabled: cfg.Engine.Excursions.Enabled, ShortOutAndBackMaxMinutes: cfg.Engine.Excursions.ShortOutAndBackMaxMinutes, ShortOutAndBackMaxDistance: cfg.Engine.Excursions.ShortOutAndBackMaxDistanceMeters},
		Motion: engine.MotionConfig{
			Enabled:                     cfg.Engine.Motion.Enabled,
			StationarySpeedThresholdKPH: cfg.Engine.Motion.StationarySpeedThresholdKPH,
			MovingSpeedThresholdKPH:     cfg.Engine.Motion.MovingSpeedThresholdKPH,
			GapThresholdMinutes:         cfg.Engine.Motion.GapThresholdMinutes,
			MinConsecutiveSamples:       cfg.Engine.Motion.MinConsecutiveSamples,
		},
		Quality:  cfg.Engine.Quality.Enabled,
		Audit:    cfg.Engine.Audit.Enabled,
		Valhalla: cfg.Valhalla.Enabled,
		H3:       cfg.H3.Enabled,
		PostGIS:  cfg.PostGIS.Enabled,
	}
}

func commonOutputName(inputs []string) string {
	if len(inputs) == 0 {
		return "combined"
	}
	var bases []string
	for _, input := range inputs {
		base := filepath.Base(input)
		ext := filepath.Ext(base)
		bases = append(bases, strings.TrimSuffix(base, ext))
	}
	sort.Strings(bases)
	if len(bases) == 1 {
		return bases[0]
	}
	return bases[0] + "_to_" + bases[len(bases)-1]
}

func RunAddRoute(observationsPath string, observationIndex int, routesPath string) error {
	if observationsPath == "" {
		return fmt.Errorf("missing route observations CSV")
	}
	if observationIndex <= 0 {
		return fmt.Errorf("route observation index must be greater than zero")
	}
	if routesPath == "" {
		routesPath = resolveSiblingPath("routes.csv", observationsPath)
	}
	observations, err := routes.LoadObservations(observationsPath)
	if err != nil {
		return err
	}
	observation, ok := routes.FindObservation(observations, observationIndex)
	if !ok {
		return fmt.Errorf("index %d not found in %s", observationIndex, observationsPath)
	}
	route := routes.RouteFromObservation(observation)
	if err := routes.AppendRoute(routesPath, route); err != nil {
		return err
	}
	fmt.Printf("added route index %d to %s\n", observationIndex, routesPath)
	fmt.Printf("route: %s (%s -> %s)\n", route.Name, route.FromSite, route.ToSite)
	fmt.Printf("distance: %.2f-%.2f km; duration: %.2f-%.2f min\n", route.DistanceMinKM, route.DistanceMaxKM, route.DurationMinMin, route.DurationMaxMin)
	return nil
}

func DefaultConfigPath() string {
	return "gogator.yaml"
}

func resolveSiblingPath(path, inputPath string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	if _, err := os.Stat(path); err == nil {
		return path
	}
	dir := filepath.Dir(inputPath)
	if dir != "." && dir != "" {
		candidate := filepath.Join(dir, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return path
}
