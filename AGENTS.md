# Codex/Agent Notes for GoGator

## Project intent

GoGator is a local Go 1.22 CLI for processing raw Gator/Teltonika GPS CSV exports into deterministic, spreadsheet-friendly trip logs.

Raw GPS CSV is canonical. Do not use Gator's processed drive/stop exports as the source of truth; they have shown suspicious timestamp and trip-detection behaviour.

## Non-negotiable behaviours

- Preserve the current file-based process command: `gogator process <raw-gps.csv> [more-raw-gps.csv ...]`.
- Preserve the current route promotion command: `gogator add route <index> from <route_observations.csv>`.
- Do not reintroduce `gogator add_route`.
- Preserve the staged natural-ish command shape shown by `gogator commands`.
- Preserve stable output columns unless the user explicitly asks for new columns.
- Keep observed/noisy tracker GPS coordinates in processed output when a site matches.
- Do not replace processed output GPS with canonical `addresses.csv` coordinates.
- Prefer durable rules over one-off fixes for individual CHECK rows.
- Keep routes advisory only. They may annotate or flag, but must not silently rewrite destinations.
- Treat routes as directional: `A -> B` and `B -> A` are different routes.
- Build and test before packaging.
- Preserve tracker signal detail: `io24`, `io251`, `pdop`, `io14`, `io247`, `io253`, `io303`, `g0`, `g1`, and `g2` are useful evidence, not disposable noise.
- Treat `g0`, `g1`, and `g2` as raw X/Y/Z accelerometer values: X left/right, Y forward/back, Z up/down.
- Treat `CHANGES.md` as append-only project history. Never rewrite older entries.

## Development environment

- Go version target: 1.22.
- Module: `gogator`.
- Main package: `./cmd/gogator`.

Common checks:

```bash
go test ./...
go build -o gogator ./cmd/gogator
./gogator commands
```

## Key source areas

- `cmd/gogator/main.go`: CLI parsing and command help.
- `internal/app/process.go`: process orchestration, default config resolution, output generation.
- `internal/gps/read.go`: raw CSV parsing, row numbering, time correction.
- `internal/gps/trips.go`: movement, stationary clusters, trip building, jitter suppression.
- `internal/sites/sites.go`: addresses.csv/TSV loading and site matching.
- `internal/routes/routes.go`: route rules, observations, anomalies, route promotion support.
- `internal/output/csv.go`: output schemas. Treat this as high-stability.

## Packaging convention

Deliver project zips as versioned archives with a top-level directory matching the archive name, for example:

```text
GoGator-v0.26.zip
└── GoGator-v0.26/
```

Append new user-visible changes to `CHANGES.md`.

## Context reference files

- `TRACKER_SIGNALS.md`: raw params, movement/idling signals, accelerometer axis meanings, crash/driving-style fields, and debugging value.
- `DESIGN_NOTES.md`: why raw GPS is canonical, local enrichment model, evidence-preservation rules, CHECK philosophy, and output stability.
- `COMMANDS.md`: CLI command intent and examples.
