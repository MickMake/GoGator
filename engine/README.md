# engine

`engine` is GoGator's explicit processing seam for the v1 migration.

## Purpose

The package gives GoGator a testable, auditable orchestration boundary while preserving current CLI and output behaviour. For v1 migration work, stability wins over novelty.

## v1 migration guardrail

- Current `gogator` CLI behaviour and processed output schemas must remain stable unless explicitly changed.
- `internal/app/process.go` calls `engine.Run(...)` as the seam.
- Legacy `internal/gps` processing logic remains available and authoritative until engine output is proven by tests and explicitly promoted.

## Current seam

- Entry point: `engine.Run(ctx, input)`.
- Input/result contract: `engine.Input` and `engine.Result` (plus diagnostics snapshots) in `engine/types.go`.
- Relationship to legacy: the orchestration path still wraps and preserves legacy trip outputs by default (`engine.trip_source: legacy`).

## Target staged pipeline

The staged target pipeline remains:

1. evidence
2. quality
3. motion
4. stays
5. visits
6. candidate trips
7. map matching
8. route signatures
9. route grouping
10. diagnostics
11. final adaptation to existing outputs

## Implemented now (v1.1 baseline)

- Engine seam orchestration through `engine.Run(...)`.
- Passive diagnostics for evidence, quality, motion, stays, visits, excursions, candidate trips, trip comparison, shadow summary/readiness, map-match diagnostics, route signatures, and route groups.
- Conservative trip-source modes with default legacy output (`legacy`, `shadow`, experimental `engine`).
- Compatibility-focused tests across engine stages and diagnostics behaviours.

## Scaffold-only / deferred

- No replacement trip algorithm is promoted as default behaviour.
- Map matching integration remains optional and advisory (Valhalla scaffold/client path).
- PostGIS-backed site matching remains scaffolded/optional and disabled by default.
- Route signatures/grouping remain passive diagnostics, not destination-rewrite logic.
- Any future promotion of engine-selected output requires explicit readiness evidence and tests.

## Baseline checks

```bash
go test ./...
go build -o gogator ./cmd/gogator
./gogator commands
```
