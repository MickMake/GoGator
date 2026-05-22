package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gogator/internal/app"
)

const appName = "gogator"
const version = "v0.25"

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
	case "add_route":
		return fmt.Errorf("add_route has been replaced by: gogator add route")
	case "add":
		return add(args[1:])
	case "delete":
		return deleteCmd(args[1:])
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
		return fmt.Errorf("not implemented yet: process gps")
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

func add(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%w: usage: gogator add <site|route> ...", errUsage)
	}
	switch args[0] {
	case "site":
		if err := parsePairs(args[1:], map[string]bool{"name": true, "gps": true, "address": false, "range": false, "dwell": false, "type": false, "important": false, "notes": false}); err != nil {
			return err
		}
		return fmt.Errorf("not implemented yet: add site")
	case "route":
		if len(args) >= 4 {
			if idx, err := strconv.Atoi(args[1]); err == nil && idx > 0 && args[2] == "from" {
				return app.RunAddRoute(args[3], idx, "")
			}
		}
		if err := parsePairs(args[1:], map[string]bool{"from": true, "to": true, "name": false, "confidence": false, "notes": false}); err != nil {
			return err
		}
		return fmt.Errorf("not implemented yet: add route")
	default:
		return fmt.Errorf("%w: unknown add target: %s", errUsage, args[0])
	}
}

func deleteCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator delete <site|route> ...", errUsage)
	}
	switch args[0] {
	case "site", "route":
		return fmt.Errorf("not implemented yet: delete %s", args[0])
	default:
		return fmt.Errorf("%w: unknown delete target: %s", errUsage, args[0])
	}
}

func importCmd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("%w: usage: gogator import <gps|sites|routes> ...", errUsage)
	}
	switch args[0] {
	case "gps", "sites", "routes":
		return fmt.Errorf("not implemented yet: import %s", args[0])
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
	case "gps", "sites", "routes", "trips", "jitter", "stats", "issues", "paths":
		return fmt.Errorf("not implemented yet: export %s", args[0])
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
	case "init", "status", "backup", "vacuum":
		return fmt.Errorf("not implemented yet: db %s", args[0])
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
  Future DB commands:      recognised now, implemented over staged releases

Commands:
  process <gps.csv...>                         Process raw GPS CSV files using current file-based workflow.
  process gps ...                              Planned DB-backed GPS processing.

  db init                                      Planned: initialise gogator.sqlite.
  db status                                    Planned: show database status.
  db backup ...                                Planned: backup database.
  db vacuum                                    Planned: compact database.

  import gps [from] <file...>                  Planned: import GPS tracker CSV rows into the database.
  import sites [from] <file>                   Planned: import site definitions.
  import routes [from] <file>                  Planned: import directional route definitions.

  export gps [date] [[as] file]                Planned: export GPS tracker points.
  export sites [[as] file]                     Planned: export site definitions.
  export routes [[as] file]                    Planned: export directional route definitions.
  export trips [date] [[as] file]              Planned: export processed trips.
  export jitter [date] [[as] file]             Planned: export jitter/suppressed trip rows.
  export stats [date] [[as] file]              Planned: export route/trip statistics.
  export issues [date] [[as] file]             Planned: export review/problem rows.
  export paths [route/date] [[as] file]        Planned: export trip/route waypoint evidence.

  add site <name/value pairs>                  Planned: add or replace one site.
  add route <name/value pairs>                 Planned: add or replace one route.
  add route <index> from <stats.csv|tsv>       Promote one observed route into routes.csv.
  delete site <name> [anyway]                  Planned: delete one site with safety checks.
  delete route from <site> to <site> [anyway]  Planned: delete one directional route rule.

  commands                                    Show extended command help with examples.
  command                                     Alias for commands.
  version                                     Print version.
  help                                        Show this help.

Examples:
  %[1]s process 2026-04_raw.csv
  %[1]s process 2026-04_raw.csv 2026-05_raw.csv
  %[1]s process 2026-04_raw.csv --timezone Australia/Sydney
  %[1]s add route 3 from 2026-04_route_observations.csv

  %[1]s db init
  %[1]s import gps from 2026-04_raw.csv
  %[1]s import sites from sites.tsv
  %[1]s import routes from routes.tsv
  %[1]s export trips during 2026 as trips.tsv
  %[1]s export gps during 2026-04 as gps.tsv
  %[1]s export paths from "Home Sweet Home" to "Asquith Public School" during 2026 as school-run-paths.tsv

Configuration:
  GOGATOR_TIMEZONE             Optional timezone override.

`, appName, version)
}

func commands() {
	fmt.Printf(`GoGator command guide

Supported date formats:
  YYYY, YYYY-MM, or YYYY-MM-DD

Command families:

  process
    Intent: Convert raw Gator/Teltonika GPS CSV exports into spreadsheet-ready trip logs.
    Current:
      %[1]s process <gps.csv...> [--timezone Australia/Sydney] [--config gogator.yaml] [--sites sites.csv] [--routes routes.csv]
    Planned:
      %[1]s process gps during 2026
      %[1]s process gps from 2026 to 2027
    Notes:
      The current file-based process command remains supported while SQLite work is staged.

  db
    Intent: Low-level database maintenance.
    Planned:
      %[1]s db init
      %[1]s db status
      %[1]s db backup
      %[1]s db vacuum

  import
    Intent: Load many records into the future SQLite evidence store.
    Planned:
      %[1]s import gps 2026-04_raw.csv
      %[1]s import gps from 2026-04_raw.csv 2026-05_raw.csv
      %[1]s import sites from sites.tsv
      %[1]s import routes from routes.tsv
    Notes:
      Use gps, not raw. Use sites, not addresses; address is a field on a site.

  export
    Intent: Export many records or views from the future SQLite evidence store.
    Planned:
      %[1]s export gps during 2026-04 as gps.tsv
      %[1]s export sites as sites.tsv
      %[1]s export routes as routes.tsv
      %[1]s export trips during 2026 as trips.tsv
      %[1]s export jitter during 2026 as jitter.tsv
      %[1]s export stats during 2026 as stats.tsv
      %[1]s export issues during 2026 as issues.tsv
      %[1]s export paths from "Home Sweet Home" to "Asquith Public School" during 2026 as paths.tsv

  add
    Intent: Add or replace one record.
    Planned:
      %[1]s add site name "Bunnings Thornleigh" gps "-33.72816964,150.97700866" range 200 type Supplier important yes
      %[1]s add route from "Home Sweet Home" to "Asquith Public School" name "School run"
    Current:
      %[1]s add route 3 from 2026-04_route_observations.csv
    Notes:
      add route from a route observations file preserves the old add_route workflow under the new command shape.

  delete
    Intent: Delete one record, with future safety checks.
    Planned:
      %[1]s delete site "Bunnings Thornleigh"
      %[1]s delete route from "Home Sweet Home" to "Asquith Public School"
    Notes:
      Future delete commands should refuse dangerous deletes unless an explicit escape hatch such as anyway is supplied.

  commands
    Intent: Show this command guide.

  version
    Intent: Print the GoGator version.

`, appName)
}

func processHelp() {
	fmt.Printf(`Usage:
  %[1]s process <raw-gps.csv> [more-raw-gps.csv ...] [--timezone Australia/Sydney] [--config gogator.yaml] [--sites sites.csv] [--routes routes.csv]

Intent:
  Process one or more raw GPS CSV files using the current file-based workflow.

Examples:
  %[1]s process 2026-04_raw.csv
  %[1]s process 2026-04_raw.csv 2026-05_raw.csv
  %[1]s process 2026-04_raw.csv --timezone Australia/Sydney
`, appName)
}
