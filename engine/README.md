<!--
PROTECTED FILE - DO NOT EDIT

This file is intentionally protected.
Do not modify, rewrite, reformat, rename, move, or delete this file.

Codex/AI agents:
- You must not change this file.
- If a task appears to require changing this file, stop and report that it is protected.
- Do not "clean up", "modernise", "simplify", or "deduplicate" this file.
-->

# engine

`engine` is GoGator's replacement GPS trip intelligence package.

The fixed goal is to replace the old `internal/gps` trip construction method with the engine method. The engine must become the long-term owner of trip construction.

The engine turns raw tracker points into itinerary records:

```text
departure site -> travel/route -> destination site -> time spent at site
```

## Non-negotiable direction

- `engine` is the destination, not a permanent compatibility wrapper.
- The old GPS trip construction path is temporary migration support.
- The old GPS method is to be replaced completely by the engine method.
- New work must move toward staged engine-owned trip construction.
- Compatibility mode protects existing output during migration only.
- Legacy output must not be authoritative. The engine output is the output now, even while it is still being improved.
- Valhalla, spatial radius plus dwell time plus context, PostGIS site matching/audit, H3 route signatures, and route grouping remain part of the intended engine direction.
- Documentation, tests, and implementation must not describe the migration seam as the final purpose of this package.

## Testing principle

Testing must exercise the engine path directly. Legacy output is not the oracle.

The useful test question is not "does engine reproduce whatever one legacy binary happened to emit?". The useful test question is:

```text
Given this input data and this configured engine behaviour, is this engine output explainable, auditable, and consistent?
```

Existing commands and output files may be used to inspect results, but they must not force the engine back into old trip-construction behaviour.

## Target staged pipeline

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

The final engine should decide trips from accumulated evidence rather than from an expanding set of special-case rules.

## Current v1 migration guardrail

The current implementation is deliberately conservative while the replacement engine is being proven:

- Current `gogator` CLI behaviour and processed output schemas must remain stable unless explicitly changed.
- `internal/app/process.go` calls `engine.Run(...)` as the seam.
- Current commands and output files must remain available, but their trip construction must be driven by the engine path.
- The default trip source must be engine (`engine.trip_source: engine`) so every run tests the engine path directly.

These guardrails do not preserve legacy authority. They exist only to keep the application runnable while engine output becomes the single source of truth.

## Current seam

- Entry point: `engine.Run(ctx, input)`.
- Input/result contract: `engine.Input` and `engine.Result` plus diagnostics snapshots in `engine/types.go`.
- Relationship to existing commands: current CLI commands and output files remain, but official trip output must come from the engine path.

## Implemented now (v1.1 baseline)

- Engine seam orchestration through `engine.Run(...)`.
- Passive diagnostics for evidence, quality, motion, stays, visits, excursions, candidate trips, trip comparison, shadow summary/readiness, map-match diagnostics, route signatures, and route groups.
- Engine-first trip-source direction. `engine` is the intended default while current CLI commands and output files remain available.
- Compatibility-focused tests across engine stages and diagnostics behaviours.

## Scaffold-only / deferred

These items are incomplete migration stages, not changes of direction:

- Replacement trip algorithm is not yet promoted as default behaviour.
- Map matching integration remains optional and advisory through the Valhalla scaffold/client path.
- PostGIS-backed site matching remains scaffolded/optional and disabled by default.
- Route signatures/grouping remain passive diagnostics, not destination-rewrite logic.
- Any future promotion of engine-selected output requires explicit readiness evidence and tests.

## Promotion path

The expected direction is:

1. Make `engine.trip_source: engine` the default path.
2. Route official processed trip output through the engine method.
3. Keep existing user-facing commands and output files available while replacing the trip construction internals.
4. Improve staged engine evidence until candidate trips are explainable and auditable.
5. Test the engine against its own expected behaviour, not against legacy as the source of truth.
6. Remove the legacy GPS trip construction path once the engine path is stable enough to stand alone.

## Baseline checks

```bash
go test ./...
go build -o gogator ./cmd/gogator
./gogator commands
```
