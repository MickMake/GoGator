package main

import (
	"fmt"
	"os"
	"strconv"

	"gogator/internal/app"
)

const appName = "gogator"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "process":
		process(os.Args[2:])
	case "add_route":
		addRoute(os.Args[2:])
	case "commands":
		commands()
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func process(args []string) {
	opts := app.Options{ConfigPath: app.DefaultConfigPath()}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--timezone":
			i++
			requireValue(args, i, a)
			opts.Timezone = args[i]
		case "--config":
			i++
			requireValue(args, i, a)
			opts.ConfigPath = args[i]
		case "--sites":
			i++
			requireValue(args, i, a)
			opts.SitesPath = args[i]
		case "--routes":
			i++
			requireValue(args, i, a)
			opts.RoutesPath = args[i]
		case "--help", "-h":
			processHelp()
			return
		default:
			if len(a) > 2 && a[:2] == "--" {
				fmt.Fprintf(os.Stderr, "unknown option: %s\n", a)
				os.Exit(2)
			}
			opts.Inputs = append(opts.Inputs, a)
			if opts.Input == "" {
				opts.Input = a
			}
		}
	}
	if len(opts.Inputs) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s process <raw-gps.csv> [more-raw-gps.csv ...] [--timezone Australia/Sydney] [--config gogator.yaml] [--sites addresses.csv] [--routes routes.csv]\n", appName)
		os.Exit(2)
	}
	if err := app.RunProcess(opts); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func addRoute(args []string) {
	var observations string
	var idxText string
	routesPath := ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--routes":
			i++
			requireValue(args, i, a)
			routesPath = args[i]
		case "--help", "-h":
			addRouteHelp()
			return
		default:
			if len(a) > 2 && a[:2] == "--" {
				fmt.Fprintf(os.Stderr, "unknown option: %s\n", a)
				os.Exit(2)
			}
			if observations == "" {
				observations = a
				continue
			}
			if idxText == "" {
				idxText = a
				continue
			}
			fmt.Fprintf(os.Stderr, "unexpected argument: %s\n", a)
			os.Exit(2)
		}
	}
	if observations == "" || idxText == "" {
		fmt.Fprintf(os.Stderr, "usage: %s add_route <route_observations.csv> <index> [--routes routes.csv]\n", appName)
		os.Exit(2)
	}
	idx, err := strconv.Atoi(idxText)
	if err != nil || idx <= 0 {
		fmt.Fprintf(os.Stderr, "invalid observation index: %s\n", idxText)
		os.Exit(2)
	}
	if err := app.RunAddRoute(observations, idx, routesPath); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func requireValue(args []string, i int, flag string) {
	if i >= len(args) {
		fmt.Fprintf(os.Stderr, "%s requires a value\n", flag)
		os.Exit(2)
	}
}

func usage() {
	fmt.Printf(`GoGator processes Gator/Teltonika raw GPS CSV exports.

Usage:
  %[1]s process <raw-gps.csv> [more-raw-gps.csv ...] [--timezone Australia/Sydney] [--config gogator.yaml] [--sites addresses.csv] [--routes routes.csv]
  %[1]s add_route <route_observations.csv> <index> [--routes routes.csv]
  %[1]s commands

Defaults:
  config:   ./gogator.yaml, falling back to ./gatorlog.yaml for older folders
  sites:    config sites value, otherwise ./addresses.csv or addresses.csv beside each input
  routes:   config routes value, otherwise ./routes.csv or routes.csv beside each input
  timezone: --timezone, then GOGATOR_TIMEZONE, then GATORLOG_TIMEZONE, then config, otherwise Australia/Sydney

Outputs use each input filename prefix:
  <input>_processed.csv
  <input>_expanded.csv
  <input>_jitter.csv
  <input>_route_observations.csv
  <input>_route_anomalies.csv
  <input>_audit.csv

For intent and examples, run:
  %[1]s commands
`, appName)
}

func commands() {
	fmt.Printf(`GoGator command guide

Command: %[1]s process <raw-gps.csv> [more-raw-gps.csv ...]
Intent:  Convert one or more raw Gator/Teltonika GPS exports into deterministic spreadsheet-ready trip logs.
Use when: You have monthly or ad-hoc raw GPS exports and want processed trips, expanded raw rows, jitter rejects, audit output, and route review files.
Examples:
  %[1]s process 2026-04_raw.csv
  %[1]s process 2026-04_raw.csv 2026-05_raw.csv 2026-06_raw.csv
  %[1]s process 2026-04_raw.csv --timezone Australia/Sydney
  %[1]s process exports/2026-04_raw.csv --config gogator.yaml --sites addresses.csv --routes routes.csv
Notes:
  - Raw GPS remains canonical because processed Gator drive/stop exports have shown suspicious timestamp and trip-detection behaviour.
  - Matched sites keep the observed/noisy GPS coordinate from the tracker output.
  - Expanded/audit detail should preserve useful tracker signals such as io24, io251, pdop, io14, io247, io253, io303, g0, g1, and g2.
  - g0/g1/g2 are raw X/Y/Z acceleration vectors: X left/right, Y forward/back, Z up/down.
  - Unknown or weakly proven destinations are labelled CHECK, because goblins dislike evidence.

Command: %[1]s add_route <route_observations.csv> <index>
Intent:  Promote one observed common route into routes.csv so future runs can recognise it.
Use when: route_observations.csv shows a repeated From Site -> To Site path that you consider normal.
Examples:
  %[1]s add_route 2026-04_route_observations.csv 3
  %[1]s add_route exports/2026-04_route_observations.csv 3 --routes routes.csv
Notes:
  - Routes are advisory only; they should not rewrite destinations silently.
  - The selected observation is appended to routes.csv using the Index column.

Command: %[1]s commands
Intent:  Show this extended command guide with examples and practical command intent.
Use when: You or Codex need a reminder of what the CLI does without spelunking through source files.
`, appName)
}

func processHelp() {
	fmt.Printf(`Usage:
  %[1]s process <raw-gps.csv> [more-raw-gps.csv ...] [--timezone Australia/Sydney] [--config gogator.yaml] [--sites addresses.csv] [--routes routes.csv]

Intent:
  Process one or more raw GPS CSVs as the source of truth and produce spreadsheet-friendly trip logs. Preserve enough raw tracker detail to debug suspicious timestamp, trip-detection, GPS, and accelerometer behaviour.

Examples:
  %[1]s process 2026-04_raw.csv
  %[1]s process 2026-04_raw.csv 2026-05_raw.csv
  %[1]s process 2026-04_raw.csv --sites addresses.csv --routes routes.csv
`, appName)
}

func addRouteHelp() {
	fmt.Printf(`Usage:
  %[1]s add_route <route_observations.csv> <index> [--routes routes.csv]

Intent:
  Append one indexed observed route to routes.csv.

Examples:
  %[1]s add_route 2026-04_route_observations.csv 3
  %[1]s add_route 2026-04_route_observations.csv 3 --routes routes.csv
`, appName)
}
