# Codex/Agent Notes for GoGator

## Project intent

GoGator is a local Go 1.22 CLI for processing raw Gator/Teltonika GPS CSV exports into deterministic, spreadsheet-friendly trip logs.

Raw GPS CSV is canonical. Do not use Gator's processed drive/stop exports as the source of truth.

## Non-negotiable behaviours

- Preserve the CLI shape: `gogator process`, `gogator add_route`, `gogator commands`.
- Preserve stable output columns unless a user explicitly asks for new columns.
- Keep observed/noisy tracker GPS coordinates in processed output when a site matches.
- Do not replace processed output GPS with canonical `addresses.csv` coordinates.
- Prefer durable rules over one-off fixes for individual CHECK rows.
- Keep routes advisory only. They may annotate or flag, but must not silently rewrite destinations.
- Build and test before packaging.

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
- `internal/routes/routes.go`: route rules, observations, anomalies, add_route support.
- `internal/output/csv.go`: output schemas. Treat this as high-stability.

## Packaging convention

Deliver project zips as versioned archives with a top-level directory matching the archive name, for example:

```text
GoGator-v0.13.zip
└── GoGator-v0.13/
```

Maintain `CHANGES.md` with user-visible changes.
