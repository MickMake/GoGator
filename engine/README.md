# engine

`engine` is GoGator's explicit orchestration seam for in-memory GPS processing.

## Current state (this run)

- `engine.Run(ctx, input)` is the compatibility wrapper around the existing processing pipeline.
- It orchestrates the current sequence: sort/recalculate, classify, build trips, collapse to important sites, apply routes, and split jitter.
- Core algorithms remain in `internal/gps` and `internal/routes` with no intended behaviour changes.

## Intended future state

- `engine` becomes the staged GPS intelligence boundary where future processing implementations can be swapped in behind a stable API.
- `internal/app/process.go` remains responsible for CLI-facing concerns (loading files/config and writing CSV outputs), while `engine` owns processing orchestration.

## Explicit non-goals for this run

- No Valhalla integration.
- No H3 integration.
- No PostGIS integration.
- No trip-detection algorithm rewrite.
- No CSV schema/header changes.
