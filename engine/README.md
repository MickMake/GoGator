# engine

`engine` is GoGator's explicit orchestration seam for in-memory GPS processing.

## Current state (this run)

- `engine.Run(ctx, input)` is the compatibility wrapper around the existing processing pipeline.
- It orchestrates the current sequence: sort/recalculate, classify, build trips, collapse to important sites, apply routes, and split jitter.
- Core algorithms remain in `internal/gps` and `internal/routes` with no intended behaviour changes.
- Compatibility mode is currently the only active behaviour path.

## Config scaffolding (placeholders)

- New configuration sections are now accepted and passed through `engine.Input.EngineConfig`:
  - `engine`
  - `engine.stay_detection`
  - `engine.motion`
  - `engine.quality`
  - `engine.audit`
  - `valhalla`
  - `h3`
  - `postgis`
- Safe defaults keep behaviour unchanged:
  - `engine.enabled: true`
  - `engine.compatibility_mode: true`
  - `engine.audit.enabled: false`
  - `valhalla.enabled: false`
  - `h3.enabled: false`
  - `postgis.enabled: false`
- These options are currently scaffolding only; they are parsed and passed through but do not alter trip logic yet.

## Intended future state

- `engine` becomes the staged GPS intelligence boundary where future processing implementations can be swapped in behind a stable API.
- `internal/app/process.go` remains responsible for CLI-facing concerns (loading files/config and writing CSV outputs), while `engine` owns processing orchestration.

## Explicit non-goals for this run

- No Valhalla integration.
- No H3 integration.
- No PostGIS integration.
- No trip-detection algorithm rewrite.
- No CSV schema/header changes.
- Valhalla, H3, and PostGIS are intentionally inactive and not required for build/test/runtime.


## Passive evidence and quality scoring (v0.26.4)

- `engine.Run` now extracts passive per-point evidence into `engine.Result.Evidence`.
- Evidence captures tracker signal fields when present (`io24`, `io251`, `io14`, `pdop`, `gpslev`, `gsmlev`, `g0`, `g1`, `g2`, `io247`, `io253`, `io303`) and tolerates missing fields safely.
- A deterministic point quality score is available with bands: `Good`, `Usable`, `Poor`, `Invalid`, `Unknown`.
- Quality scoring is passive instrumentation only for now: it does not alter trip boundaries, jitter handling, route application, important-site collapse, or CSV output schemas.
- `engine.quality.enabled` acts as a passive switch for scoring. When disabled, quality remains `Unknown` while evidence extraction still runs.
- Intended future use: motion/stay/trip stages may consume this evidence, but this release keeps compatibility-first behaviour unchanged.


## Passive motion classification with hysteresis (v0.26.5)

- Added passive motion diagnostics (`MotionSample`, `MotionSegment`) with states `Moving`, `Stationary`, `Unknown`, `Gap`, and `Noise`.
- Classification combines existing point evidence/quality, speed, `io24`, `io251`, coordinate validity, duplicate timestamps, and timestamp gaps.
- Conservative hysteresis requires repeated contrary samples before changing between moving and stationary, and tolerates brief uncertainty.
- Motion remains passive in this release: it does not alter trip construction, important-site collapsing, route application, or CSV output headers.
- Intended future use: staged stay detection and future trip-building refinements can consume motion diagnostics without changing current compatibility behaviour today.


## Passive stay/dwell detection (v0.26.6)

- Added passive stay clustering diagnostics in `engine` using motion state, point quality, duration, point count, radius/spread, and time gaps.
- Stay diagnostics include types (`SiteStop`, `UnknownStop`, `Pause`, `Traffic`, `PickupCandidate`, `GapInferredStop`, `NoiseCluster`), confidence, reasons, representative coordinates, duration, radius, and source point indexes.
- Optional site proximity tagging marks stays near known sites as `SiteStop` without changing existing site-collapsing/trip logic.
- Stay detection remains passive and does not alter trip construction, jitter outputs, routes, or CSV headers in this release.


## Passive visit and excursion diagnostics (v0.26.7)

- Added passive visit modelling derived from detected stays. Visits carry timing, representative coordinates, stay linkage, optional known-site match, confidence, reasons, and point traceability.
- Added passive excursion modelling for movement legs between ordered visits, including timing windows, duration, approximate leg distance, confidence, reasons, and point traceability.
- Visit/excursion diagnostics are passive only in this release: they do not alter trip boundaries, important-site collapse, route application, jitter handling, or CSV headers.
- Intended future use: the replacement trip builder can consume visit/excursion diagnostics once explicitly enabled in a future staged run.


## Passive replacement trip-builder prototype (v0.26.8)

- Added passive candidate trip building from visit transitions and excursion evidence.
- Candidate trips include timing, origin/destination labels, point index boundaries, approximate distance, type classification, confidence, reasons, and quality warnings.
- Added boundary confidence facets for origin/destination boundaries, movement, GPS quality, gap/noise impact, site matching, and duration plausibility.
- Added lightweight deterministic comparison diagnostics against legacy valid/jitter trips to support staged audit and migration.
- Candidate trips and comparisons are diagnostics only and are **not** official output; legacy valid/jitter generation remains unchanged.


## Optional engine diagnostics output (v0.26.9)

- `engine` now exposes a read-only `Result.Diagnostics()` snapshot for output writers.
- `internal/output` can write optional `*_engine_*.csv` diagnostics for passive stages: points, motion, stays, visits, excursions, candidate trips, and trip comparison.
- Diagnostics are disabled by default. Enable with `engine.audit.enabled: true` plus either `engine.audit.output_diagnostics: true` (all) or per-file toggles.
- Legacy-compatible processed trip output remains the official output path in this release.
