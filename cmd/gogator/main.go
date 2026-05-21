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

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, errUsage) {
			usage()
			os.Exit(2)
		}
		os.Exit(1)
	}
}

var errUsage = errors.New("usage error")

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return fmt.Errorf("%w: missing command", errUsage)
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
	case "commands":
		commands()
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
			if err := requireValue(args, i, a); err != nil { return err }
			opts.Timezone = args[i]
		case "--config":
			i++
			if err := requireValue(args, i, a); err != nil { return err }
			opts.ConfigPath = args[i]
		case "--sites":
			i++
			if err := requireValue(args, i, a); err != nil { return err }
			opts.SitesPath = args[i]
		case "--routes":
			i++
			if err := requireValue(args, i, a); err != nil { return err }
			opts.RoutesPath = args[i]
		case "--help", "-h":
			processHelp(); return nil
		default:
			if strings.HasPrefix(a, "--") { return fmt.Errorf("%w: unknown option: %s", errUsage, a) }
			opts.Inputs = append(opts.Inputs, a)
			if opts.Input == "" { opts.Input = a }
		}
	}
	if len(opts.Inputs) == 0 { return fmt.Errorf("%w: usage: %s process <raw-gps.csv> ...", errUsage, appName) }
	if err := app.RunProcess(opts); err != nil { return fmt.Errorf("error: %w", err) }
	return nil
}

func add(args []string) error {
	if len(args) == 0 { return fmt.Errorf("%w: usage: gogator add <site|route> ...", errUsage) }
	switch args[0] {
	case "site":
		if err := parsePairs(args[1:], map[string]bool{"name":true,"gps":true,"address":false,"range":false,"dwell":false,"type":false,"important":false,"notes":false}); err != nil { return err }
		return fmt.Errorf("not implemented yet: add site")
	case "route":
		if len(args) >= 4 {
			if idx, err := strconv.Atoi(args[1]); err == nil && idx > 0 && args[2] == "from" {
				return app.RunAddRoute(args[3], idx, "")
			}
		}
		if err := parsePairs(args[1:], map[string]bool{"from":true,"to":true,"name":false,"confidence":false,"notes":false}); err != nil { return err }
		return fmt.Errorf("not implemented yet: add route")
	default:
		return fmt.Errorf("%w: unknown add target: %s", errUsage, args[0])
	}
}

func deleteCmd(args []string) error { if len(args)<1 {return fmt.Errorf("%w: usage: gogator delete <site|route> ...",errUsage)}; return fmt.Errorf("not implemented yet: delete %s", args[0]) }
func importCmd(args []string) error { if len(args)<1 {return fmt.Errorf("%w: usage: gogator import <gps|sites|routes> ...",errUsage)}; return fmt.Errorf("not implemented yet: import %s", args[0]) }
func exportCmd(args []string) error { if len(args)<1 {return fmt.Errorf("%w: usage: gogator export <gps|sites|routes|trips|jitter|stats|issues|paths> ...",errUsage)}; return fmt.Errorf("not implemented yet: export %s", args[0]) }
func dbCmd(args []string) error { if len(args)<1 {return fmt.Errorf("%w: usage: gogator db <init|status|backup|vacuum>",errUsage)}; return fmt.Errorf("not implemented yet: db %s", args[0]) }

func parsePairs(args []string, allowed map[string]bool) error {
	seen := map[string]bool{}
	for i:=0;i<len(args);i+=2 {
		k:=args[i]
		if _,ok:=allowed[k];!ok { return fmt.Errorf("%w: unknown field: %s", errUsage, k) }
		if i+1>=len(args) { return fmt.Errorf("%w: missing value for field: %s", errUsage, k) }
		seen[k]=true
	}
	for k,req:= range allowed { if req && !seen[k] { return fmt.Errorf("%w: missing required field: %s", errUsage, k) } }
	return nil
}

func requireValue(args []string, i int, flag string) error {
	if i >= len(args) { return fmt.Errorf("%w: %s requires a value", errUsage, flag) }
	return nil
}

func usage() { fmt.Printf("Usage: %s <command>\nRun '%s commands' for command tree.\n", appName, appName) }
func commands() { fmt.Printf("%s\n  db <init|status|backup|vacuum> (planned)\n  import <gps|sites|routes> ... (planned)\n  export <gps|sites|routes|trips|jitter|stats|issues|paths> ... (planned)\n  add <site|route> ...\n  delete <site|route> ... (planned)\n  process <raw-gps.csv...>\n  process gps ... (planned)\n  commands\n  help\n", appName) }
func processHelp() { fmt.Printf("Usage: %s process <raw-gps.csv> [more-raw-gps.csv ...] [--timezone Australia/Sydney] [--config gogator.yaml] [--sites addresses.csv] [--routes routes.csv]\n", appName) }
