package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gogator/internal/app"
	"gogator/internal/config"
	"gogator/internal/store"
)

const appName = "gogator"
const version = "v0.26.18"

var errUsage = errors.New("usage error")

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, errUsage) {
			fmt.Fprintf(os.Stderr, "\nRun '%s commands' for extended command help.\n", appName)
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return nil
	}
	switch args[0] {
	case "process":
		return process(args[1:])
	case "load":
		return loadCmd(args[1:])
	case "dump":
		return dumpCmd(args[1:])
	case "add":
		return add(args[1:])
	case "delete":
		return deleteCmd(args[1:])
	case "set":
		return setCmd(args[1:])
	case "show":
		return showCmd(args[1:])
	case "reset":
		return resetCmd(args[1:])
	case "import":
		return importCmd(args[1:])
	case "export":
		return exportCmd(args[1:])
	case "db":
		return dbCmd(args[1:])
	case "commands", "command":
		commands()
		return nil
	case "version":
		fmt.Printf("%s %s\n", appName, version)
		return nil
	case "help", "--help", "-h":
		usage()
		return nil
	default:
		return fmt.Errorf("%w: unknown command: %s", errUsage, args[0])
	}
}

func process(args []string) error {
	if len(args) > 0 && args[0] == "gps" {
		return processGPS(args[1:])
	}
	opts := app.Options{ConfigPath: app.DefaultConfigPath()}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--timezone":
			i++
			if err := requireValue(args, i, a); err != nil {
				return err
			}
			opts.Timezone = args[i]
		case "--config":
			i++
			if err := requireValue(args, i, a); err != nil {
				return err
			}
			opts.ConfigPath = args[i]
		case "--sites":
			i++
			if err := requireValue(args, i, a); err != nil {
				return err
			}
			opts.SitesPath = args[i]
		case "--routes":
			i++
			if err := requireValue(args, i, a); err != nil {
				return err
			}
			opts.RoutesPath = args[i]
		case "--help", "-h":
			processHelp()
			return nil
		default:
			if strings.HasPrefix(a, "--") {
				return fmt.Errorf("%w: unknown option: %s", errUsage, a)
			}
			opts.Inputs = append(opts.Inputs, a)
			if opts.Input == "" {
				opts.Input = a
			}
		}
	}
	if len(opts.Inputs) == 0 {
		return fmt.Errorf("%w: usage: %s process <raw-gps.csv> ...", errUsage, appName)
	}
	if err := app.RunProcess(opts); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

func processGPS(args []string) error {
	opts := app.Options{ConfigPath: app.DefaultConfigPath()}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--timezone":
			i++
			if err := requireValue(args, i, a); err != nil {
				return err
			}
			opts.Timezone = args[i]
		case "--config":
			i++
			if err := requireValue(args, i, a); err != nil {
				return err
			}
			opts.ConfigPath = args[i]
		case "--help", "-h":
			processGPSHelp()
			return nil
		default:
			return fmt.Errorf("%w: usage: %s process gps [--timezone Australia/Sydney] [--config gogator.yaml]", errUsage, appName)
		}
	}
	if err := app.RunProcessGPS(opts); err != nil {
		return fmt.Errorf("error: %w", err)
	}
	return nil
}

func loadCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator load <vendor> [from] <file...>", errUsage)
	}
	switch args[0] {
	case "gator":
		paths, err := parseLoadFileArgs("gator", args[1:])
		if err != nil {
			return err
		}
		return loadGator(paths)
	case "google":
		if _, err := parseLoadFileArgs("google", args[1:]); err != nil {
			return err
		}
		return fmt.Errorf("not implemented yet: load google")
	default:
		return fmt.Errorf("%w: unknown load vendor: %s", errUsage, args[0])
	}
}

func dumpCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator dump <vendor> [as] <file>", errUsage)
	}
	switch args[0] {
	case "gator":
		path := "gator.csv"
		if len(args) > 1 {
			p, err := parseOptionalDumpFile("gator", args[1:])
			if err != nil {
				return err
			}
			path = p
		}
		if err := store.ExportRaw(path); err != nil {
			return fmt.Errorf("dump gator: %w", err)
		}
		fmt.Printf("dumped gator data to %s\n", path)
		return nil
	case "google":
		if len(args) > 1 {
			if _, err := parseOptionalDumpFile("google", args[1:]); err != nil {
				return err
			}
		}
		return fmt.Errorf("not implemented yet: dump google")
	default:
		return fmt.Errorf("%w: unknown dump vendor: %s", errUsage, args[0])
	}
}

func loadGator(paths []string) error {
	cfg, err := config.Load(app.DefaultConfigPath())
	if err != nil {
		return err
	}
	if envTZ := os.Getenv("GOGATOR_TIMEZONE"); envTZ != "" {
		cfg.Timezone = envTZ
	}
	loc, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		return fmt.Errorf("timezone %q: %w", cfg.Timezone, err)
	}
	result, err := store.ImportGPS(paths, loc, cfg)
	if err != nil {
		return fmt.Errorf("load gator: %w", err)
	}
	fmt.Printf("loaded gator files: %d\n", result.Files)
	fmt.Printf("gator gps rows read: %d\n", result.RawRows)
	fmt.Printf("new gps points: %d\n", result.GPSPoints)
	fmt.Printf("new gps point sources: %d\n", result.SourceRows)
	return nil
}

func add(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: usage: gogator add <site|route> ...", errUsage)
	}
	switch args[0] {
	case "site":
		if err := parsePairs(args[1:], map[string]bool{"name": true, "gps": true, "address": false, "range": false, "dwell": false, "type": false, "important": false, "notes": false}); err != nil {
			return err
		}
		pairs, _ := pairsMap(args[1:])
		return addSiteDB(pairs)
	case "route":
		if len(args) >= 4 {
			if idx, err := strconv.Atoi(args[1]); err == nil && idx > 0 && args[2] == "from" {
				return app.RunAddRoute(args[3], idx, "")
			}
		}
		if err := parsePairs(args[1:], map[string]bool{"from": true, "to": true, "name": false, "confidence": false, "notes": false}); err != nil {
			return err
		}
		pairs, _ := pairsMap(args[1:])
		if err := store.UpsertRoute(store.RouteRecord{FromSite: pairs["from"], ToSite: pairs["to"], Name: pairs["name"], Confidence: pairs["confidence"], Notes: pairs["notes"]}); err != nil {
			return err
		}
		fmt.Printf("upserted route: %s -> %s\n", pairs["from"], pairs["to"])
		return nil
	default:
		return fmt.Errorf("%w: unknown add target: %s", errUsage, args[0])
	}
}

func deleteCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator delete <site|route> ...", errUsage)
	}
	switch args[0] {
	case "site":
		if len(args) < 2 {
			return fmt.Errorf("%w: usage: gogator delete site <name> [anyway]", errUsage)
		}
		anyway := len(args) > 2 && args[2] == "anyway"
		if err := store.DeleteSite(args[1], anyway); err != nil {
			return err
		}
		fmt.Printf("deleted site: %s\n", args[1])
		return nil
	case "route":
		if err := parsePairs(args[1:], map[string]bool{"from": true, "to": true}); err != nil {
			return err
		}
		pairs, _ := pairsMap(args[1:])
		if err := store.DeleteRoute(pairs["from"], pairs["to"]); err != nil {
			return err
		}
		fmt.Printf("deleted route: %s -> %s\n", pairs["from"], pairs["to"])
		return nil
	default:
		return fmt.Errorf("%w: unknown delete target: %s", errUsage, args[0])
	}
}

func setCmd(args []string) error {
	if isGPSParamsCommand(args) {
		return fmt.Errorf("not implemented yet: set gps params")
	}
	return fmt.Errorf("%w: usage: gogator set gps params <param[,param...]>", errUsage)
}

func showCmd(args []string) error {
	if isGPSParamsCommand(args) {
		return fmt.Errorf("not implemented yet: show gps params")
	}
	return fmt.Errorf("%w: usage: gogator show gps params", errUsage)
}

func resetCmd(args []string) error {
	if isGPSParamsCommand(args) {
		return fmt.Errorf("not implemented yet: reset gps params")
	}
	return fmt.Errorf("%w: usage: gogator reset gps params", errUsage)
}

func isGPSParamsCommand(args []string) bool {
	return len(args) >= 2 && args[0] == "gps" && args[1] == "params"
}

func importCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator import <sites|routes> ...", errUsage)
	}
	switch args[0] {
	case "routes":
		path, err := parseFileArg("routes", args[1:])
		if err != nil {
			return err
		}
		n, err := store.ImportRoutes(path)
		if err != nil {
			return fmt.Errorf("import routes: %w", err)
		}
		fmt.Printf("imported %d route(s) from %s\n", n, path)
		return nil
	case "sites":
		path, err := parseFileArg("sites", args[1:])
		if err != nil {
			return err
		}
		n, err := store.ImportSites(path)
		if err != nil {
			return fmt.Errorf("import sites: %w", err)
		}
		fmt.Printf("imported %d site(s) from %s\n", n, path)
		return nil
	case "addresses":
		return fmt.Errorf("%w: addresses is not a command; use sites instead", errUsage)
	default:
		return fmt.Errorf("%w: unknown import target: %s", errUsage, args[0])
	}
}

func exportCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator export <gps|sites|routes|trips|jitter|stats|issues|paths> ...", errUsage)
	}
	switch args[0] {
	case "gps":
		path := "gps.tsv"
		if len(args) > 1 {
			p, err := parseOptionalAsFile("gps", args[1:])
			if err != nil {
				return err
			}
			path = p
		}
		if err := store.ExportGPS(path); err != nil {
			return fmt.Errorf("export gps: %w", err)
		}
		fmt.Printf("exported gps to %s\n", path)
		return nil
	case "trips", "jitter", "stats", "issues", "paths":
		return fmt.Errorf("not implemented yet: export %s", args[0])
	case "routes":
		path := "routes.tsv"
		if len(args) > 1 {
			p, err := parseOptionalAsFile("routes", args[1:])
			if err != nil {
				return err
			}
			path = p
		}
		if err := store.ExportRoutes(path); err != nil {
			return fmt.Errorf("export routes: %w", err)
		}
		fmt.Printf("exported routes to %s\n", path)
		return nil
	case "sites":
		path := "sites.tsv"
		if len(args) > 1 {
			p, err := parseOptionalAsFile("sites", args[1:])
			if err != nil {
				return err
			}
			path = p
		}
		if err := store.ExportSites(path); err != nil {
			return fmt.Errorf("export sites: %w", err)
		}
		fmt.Printf("exported sites to %s\n", path)
		return nil
	case "addresses":
		return fmt.Errorf("%w: addresses is not a command; use sites instead", errUsage)
	default:
		return fmt.Errorf("%w: unknown export target: %s", errUsage, args[0])
	}
}

func dbCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator db <init|status|backup|vacuum>", errUsage)
	}
	switch args[0] {
	case "init":
		if err := store.Init(store.DefaultPath); err != nil {
			return fmt.Errorf("db init: %w", err)
		}
		fmt.Printf("initialised database: %s\n", store.DefaultPath)
		return nil
	case "status":
		exists, err := store.Exists(store.DefaultPath)
		if err != nil {
			return fmt.Errorf("db status: %w", err)
		}
		if !exists {
			return fmt.Errorf("database not found: %s (run: %s db init)", store.DefaultPath, appName)
		}
		counts, version, err := store.Status(store.DefaultPath)
		if err != nil {
			return fmt.Errorf("db status: %w", err)
		}
		fmt.Printf("database: %s\n", store.DefaultPath)
		fmt.Printf("sqlite: %s\n", version)
		fmt.Printf("gps_points: %d\n", counts.GPSPoints)
		fmt.Printf("sites: %d\n", counts.Sites)
		fmt.Printf("routes: %d\n", counts.Routes)
		fmt.Printf("processing_runs: %d\n", counts.ProcessingRuns)
		fmt.Printf("trips: %d\n", counts.Trips)
		fmt.Printf("issues: %d\n", counts.Issues)
		return nil
	case "backup":
		path := "gogator-backup.sqlite"
		if len(args) > 1 {
			p, err := parseOptionalDBFile("backup", args[1:])
			if err != nil {
				return err
			}
			path = p
		}
		if err := store.Backup(store.DefaultPath, path); err != nil {
			return fmt.Errorf("db backup: %w", err)
		}
		fmt.Printf("backed up database to %s\n", path)
		return nil
	case "vacuum":
		if len(args) > 1 {
			return fmt.Errorf("%w: usage: gogator db vacuum", errUsage)
		}
		if err := store.Vacuum(store.DefaultPath); err != nil {
			return fmt.Errorf("db vacuum: %w", err)
		}
		fmt.Printf("vacuumed database: %s\n", store.DefaultPath)
		return nil
	default:
		return fmt.Errorf("%w: unknown db command: %s", errUsage, args[0])
	}
}

func parsePairs(args []string, allowed map[string]bool) error {
	seen := map[string]bool{}
	for i := 0; i < len(args); i += 2 {
		k := args[i]
		if _, ok := allowed[k]; !ok {
			return fmt.Errorf("%w: unknown field: %s", errUsage, k)
		}
		if i+1 >= len(args) {
			return fmt.Errorf("%w: missing value for field: %s", errUsage, k)
		}
		seen[k] = true
	}
	for k, required := range allowed {
		if required && !seen[k] {
			return fmt.Errorf("%w: missing required field: %s", errUsage, k)
		}
	}
	return nil
}

func requireValue(args []string, i int, flag string) error {
	if i >= len(args) {
		return fmt.Errorf("%w: %s requires a value", errUsage, flag)
	}
	return nil
}

func usage() {
	fmt.Printf(`GoGator processes Gator/Teltonika raw GPS exports into deterministic trip, site, route, and GPS evidence data.

Version: %[2]s

Global notes:
  Default database:        gogator.sqlite
  Default config:          gogator.yaml
  Default timezone:        Australia/Sydney
  Supported date formats:  YYYY, YYYY-MM, or YYYY-MM-DD
  Current process command: file-based CSV processing remains supported

Commands:
  process <gps.csv...>                         Process raw GPS CSV files using current file-based workflow.
  process gps                                  Process loaded SQLite GPS rows and write standard process CSV outputs.

  load gator [from] <file...>                  Load Gator tracker CSV/TSV rows into the database.
  load google [from] <file...>                 Planned: load Google tracker/location data.
  dump gator [[as] file]                       Dump Gator tracker-compatible CSV rows.
  dump google [[as] file]                      Planned: dump Google tracker/location data.

  db init                                      Initialise gogator.sqlite schema.
  db status                                    Show database status and row counts.
  db backup [[as] file]                       Back up database to a new SQLite file.
  db vacuum                                    Compact database.

  import sites [from] <file>                   Import site definitions.
  import routes [from] <file>                  Import directional route definitions.

  export gps [[as] file]                       Export clean GPS tracker rows.
  export sites [[as] file]                     Export site definitions.
  export routes [[as] file]                    Export directional route definitions.
  export trips [date] [[as] file]              Planned: export processed trips.
  export jitter [date] [[as] file]             Planned: export jitter/suppressed trip rows.
  export stats [date] [[as] file]              Planned: export route/trip statistics.
  export issues [date] [[as] file]             Planned: export review/problem rows.
  export paths [route/date] [[as] file]        Planned: export trip/route waypoint evidence.

  set gps params <param[,param...]>            Planned: set GPS export param columns.
  show gps params                              Planned: show GPS export param columns.
  reset gps params                             Planned: reset GPS export params to default behaviour.

  add site <name/value pairs>                  Add or replace one site.
  add route <name/value pairs>                 Add or replace one route.
  add route <index> from <stats.csv|tsv>       Promote one observed route into routes.csv.
  delete site <name> [anyway]                  Delete one site with safety checks.
  delete route from <site> to <site>           Delete one directional route rule.

  commands                                    Show extended command help with examples.
  command                                     Alias for commands.
  version                                     Print version.
  help                                        Show this help.

Examples:
  %[1]s process raw.csv
  %[1]s process raw.csv --timezone Australia/Sydney
  %[1]s process gps
  %[1]s db init
  %[1]s db backup as gogator-backup.sqlite
  %[1]s db vacuum
  %[1]s load gator from raw.csv
  %[1]s dump gator as gator.csv
  %[1]s load google from google.json
  %[1]s export gps as gps.tsv
  %[1]s set gps params io66,io67,io200
  %[1]s import sites from sites.tsv
  %[1]s import routes from routes.tsv
  %[1]s export sites as sites.tsv
  %[1]s export routes as routes.tsv

Configuration:
  GOGATOR_TIMEZONE             Optional timezone override.

`, appName, version)
}

func commands() {
	usage()
}

func processHelp() {
	fmt.Printf(`Usage:
  %[1]s process <raw-gps.csv> [more-raw-gps.csv ...] [--timezone Australia/Sydney] [--config gogator.yaml] [--sites sites.csv] [--routes routes.csv]
  %[1]s process gps [--timezone Australia/Sydney] [--config gogator.yaml]

Intent:
  Process one or more raw GPS CSV files using the current file-based workflow, or process already-loaded SQLite GPS rows with process gps.

Examples:
  %[1]s process raw.csv
  %[1]s process raw.csv --timezone Australia/Sydney
  %[1]s process gps
`, appName)
}

func processGPSHelp() {
	fmt.Printf(`Usage:
  %[1]s process gps [--timezone Australia/Sydney] [--config gogator.yaml]

Intent:
  Process GPS rows already loaded into gogator.sqlite and write the standard process CSV outputs with the gogator_ prefix.

Examples:
  %[1]s process gps
`, appName)
}

func pairsMap(args []string) (map[string]string, error) {
	if len(args)%2 != 0 {
		return nil, fmt.Errorf("%w: missing value for field: %s", errUsage, args[len(args)-1])
	}
	m := map[string]string{}
	for i := 0; i < len(args); i += 2 {
		m[args[i]] = args[i+1]
	}
	return m, nil
}

func addSiteDB(m map[string]string) error {
	latlng := strings.Split(m["gps"], ",")
	if len(latlng) < 2 {
		return fmt.Errorf("%w: invalid gps", errUsage)
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(latlng[0]), 64)
	if err != nil {
		return err
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(latlng[1]), 64)
	if err != nil {
		return err
	}
	rng, _ := strconv.ParseFloat(strings.TrimSpace(m["range"]), 64)
	dwell, _ := strconv.ParseFloat(strings.TrimSpace(m["dwell"]), 64)
	important := parseImportant(strings.TrimSpace(m["important"]))
	err = store.UpsertSite(store.SiteRecord{Name: m["name"], Address: m["address"], Lat: lat, Lng: lng, RangeM: rng, MinDestinationMinutes: dwell, Type: m["type"], Important: important, Notes: m["notes"]})
	if err != nil {
		return err
	}
	fmt.Printf("upserted site: %s\n", m["name"])
	return nil
}

func parseImportant(v string) bool {
	if v == "" {
		return true
	}
	v = strings.ToLower(v)
	return v == "1" || v == "true" || v == "yes" || v == "y"
}

func parseFileArg(target string, args []string) (string, error) {
	files, err := parseFileArgs(target, args)
	if err != nil {
		return "", err
	}
	if len(files) != 1 {
		return "", fmt.Errorf("%w: usage: gogator import %s [from] <file>", errUsage, target)
	}
	return files[0], nil
}

func parseFileArgs(target string, args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: usage: gogator import %s [from] <file...>", errUsage, target)
	}
	if args[0] == "from" {
		args = args[1:]
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: usage: gogator import %s [from] <file...>", errUsage, target)
	}
	return args, nil
}

func parseLoadFileArgs(vendor string, args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: usage: gogator load %s [from] <file...>", errUsage, vendor)
	}
	if args[0] == "from" {
		args = args[1:]
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: usage: gogator load %s [from] <file...>", errUsage, vendor)
	}
	return args, nil
}

func parseOptionalAsFile(target string, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if len(args) == 2 && args[0] == "as" {
		return args[1], nil
	}
	return "", fmt.Errorf("%w: usage: gogator export %s [as] <file>", errUsage, target)
}

func parseOptionalDumpFile(vendor string, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if len(args) == 2 && args[0] == "as" {
		return args[1], nil
	}
	return "", fmt.Errorf("%w: usage: gogator dump %s [as] <file>", errUsage, vendor)
}

func parseOptionalDBFile(target string, args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	if len(args) == 2 && args[0] == "as" {
		return args[1], nil
	}
	return "", fmt.Errorf("%w: usage: gogator db %s [as] <file>", errUsage, target)
}
