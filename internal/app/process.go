package app

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gogator/internal/config"
	"gogator/internal/gps"
	"gogator/internal/output"
	"gogator/internal/routes"
	"gogator/internal/sites"
)

type Options struct {
	Input      string
	Inputs     []string
	ConfigPath string
	SitesPath  string
	RoutesPath string
	Timezone   string
}

func RunProcess(opts Options) error {
	if len(opts.Inputs) == 0 && opts.Input != "" {
		opts.Inputs = []string{opts.Input}
	}
	if len(opts.Inputs) == 0 {
		return fmt.Errorf("missing input CSV")
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = DefaultConfigPath()
	}
	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return err
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
	if math.Abs(cfg.RawTime.CorrectionHours) > 0.000001 {
		fmt.Fprintf(os.Stderr, "warning: raw_time.correction_hours=%.2f shifts raw tracker timestamps before local date grouping; use only for known malformed exports\n", cfg.RawTime.CorrectionHours)
	}

	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("timezone %q: %w", cfg.Timezone, err)
	}

	return runProcessCombined(opts, cfg, loc)
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
	sort.SliceStable(points, func(i, j int) bool {
		if points[i].Time.Equal(points[j].Time) {
			if points[i].SourceFile == points[j].SourceFile {
				return points[i].RawRow < points[j].RawRow
			}
			return points[i].SourceFile < points[j].SourceFile
		}
		return points[i].Time.Before(points[j].Time)
	})
	gps.RecalculatePointDeltas(points)

	points = gps.Classify(points, cfg)
	valid, jitter := gps.BuildTrips(points, cfg, siteList)
	valid, jitter = gps.CollapseToImportantSites(valid, jitter, cfg, siteList)
	valid, routeObservations, routeAnomalies := routes.Apply(valid, routeRules, cfg.Site.UnknownSiteLabel)
	jitterReview, jitterSameSite := splitSameSiteJitter(jitter)

	prefix := output.Prefix(primaryInput)
	if len(opts.Inputs) > 1 {
		prefix = filepath.Join(filepath.Dir(primaryInput), commonOutputName(opts.Inputs))
	}
	expanded := prefix + "_expanded.csv"
	processed := prefix + "_processed.csv"
	jitterPath := prefix + "_jitter.csv"
	jitterSameSitePath := prefix + "_jitter_same_site.csv"
	audit := prefix + "_audit.csv"
	routeObservationsPath := prefix + "_route_observations.csv"
	routeAnomaliesPath := prefix + "_route_anomalies.csv"

	if err := output.WriteExpanded(expanded, points); err != nil {
		return err
	}
	if err := output.WriteTrips(processed, valid); err != nil {
		return err
	}
	if err := output.WriteTrips(jitterPath, jitterReview); err != nil {
		return err
	}
	if err := output.WriteTrips(jitterSameSitePath, jitterSameSite); err != nil {
		return err
	}
	if err := output.WriteRouteObservations(routeObservationsPath, routeObservations); err != nil {
		return err
	}
	if err := output.WriteRouteAnomalies(routeAnomaliesPath, routeAnomalies); err != nil {
		return err
	}
	if err := output.WriteAudit(audit, strings.Join(opts.Inputs, ";"), sitesPath, routesPath, opts.ConfigPath, cfg.Timezone, len(points), len(valid), len(jitter), len(siteList), len(routeRules)); err != nil {
		return err
	}

	fmt.Printf("processed raw points: %d\n", len(points))
	fmt.Printf("valid trips: %d\n", len(valid))
	fmt.Printf("rejected jitter: %d\n", len(jitter))
	fmt.Printf("same-site jitter: %d\n", len(jitterSameSite))
	fmt.Printf("wrote: %s\n", processed)
	fmt.Printf("wrote: %s\n", expanded)
	fmt.Printf("wrote: %s\n", jitterPath)
	fmt.Printf("wrote: %s\n", jitterSameSitePath)
	fmt.Printf("wrote: %s\n", routeObservationsPath)
	fmt.Printf("wrote: %s\n", routeAnomaliesPath)
	fmt.Printf("wrote: %s\n", audit)
	return nil
}

func splitSameSiteJitter(jitter []gps.Trip) ([]gps.Trip, []gps.Trip) {
	var review []gps.Trip
	var sameSite []gps.Trip
	for _, t := range jitter {
		from := strings.TrimSpace(t.DepartureSite)
		to := strings.TrimSpace(t.DestinationSite)
		if from != "" && to != "" && strings.EqualFold(from, to) {
			sameSite = append(sameSite, t)
			continue
		}
		review = append(review, t)
	}
	return review, sameSite
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
