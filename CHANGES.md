
## v0.26.16

- Added passive engine route-signature diagnostics for candidate trips using a deterministic internal spatial grid placeholder (H3-style cell IDs without external H3 dependency).
- Signature generation now prefers passive Valhalla matched shapes when present and safely falls back to candidate raw traces; insufficient traces are skipped with warnings.
- Added passive deterministic route grouping diagnostics by identical route signatures.
- Added optional diagnostic CSV outputs for route signatures and route groups when enabled, without changing official processed/jitter CSV schemas.
- Added focused tests for deterministic signatures, duplicate-cell collapse, insufficient-point warnings, map-matched source preference, and grouping behavior.

## v0.26.17

- Added optional engine-side PostGIS site-matching scaffolding under `engine/sitematch` with a clean interface boundary, deterministic `NoopSiteMatcher`, and tested in-memory radius matcher helper for future staged integration.
- Added PostGIS scaffold config fields (`postgis.dsn`, `postgis.match_radius_meters`, `postgis.audit_enabled`) while preserving safe defaults (`postgis.enabled: false`) and no database requirement by default.
- Added scaffold `PostGISSiteMatcher` constructor validation and clear unsupported/not-implemented matching errors when explicitly enabled before a real DB-backed implementation exists.
- Added `engine/sitematch/schema.sql` documenting intended future PostGIS tables/indexes for sites and audit persistence (visit/stay and route/candidate audit).
- Preserved default legacy behaviour and official output schemas; no CLI command or processed CSV header changes.

# Changes

## v0.26.15

- Added passive Valhalla map-match diagnostics for engine candidate trips using the existing optional `engine/mapmatch` client path; default behaviour remains legacy and unchanged when Valhalla is disabled.
- Kept map-matching advisory-only: successful/failed match outcomes are captured in engine diagnostics without changing official valid/jitter output, trip-source selection, routes, readiness decisions, or processed CSV headers.
- Extended optional engine diagnostics output with `_engine_mapmatch.csv` when diagnostics are enabled, preserving existing diagnostic CSV schemas.
- Added focused tests for disabled/no-call behaviour, enabled mocked map-matching behaviour, and passive diagnostic emission.

## v0.26.14

- Added optional Valhalla HTTP map-matching client scaffolding under `engine/mapmatch` with a clean map-matching interface boundary and a safe `NoopMapMatcher` default path.
- Extended `valhalla` config scaffolding with `base_url`, `timeout_seconds`, `endpoint`, and `max_points_per_request`; defaults remain disabled and non-impacting.
- Wired engine-side matcher construction only (Noop when disabled, Valhalla client when enabled) without altering trip building, routes, diagnostics decisions, or processed CSV output.
- Added mocked HTTP unit tests for request/response handling, malformed JSON, HTTP errors, empty input safety, and disabled-default behaviour without requiring a live Valhalla server.


## v0.26.13

- Hardened `engine.trip_source: engine` candidate adaptation with conservative confidence handling, traceability flags, and robust unknown-endpoint safety while preserving stable processed CSV headers.
- Added deterministic rejection/selection hardening for noise-affected candidates and expanded engine selection diagnostics with candidate/official/rejected counts.
- Added edge-case tests for candidate adaptation, jitter classification, and fallback selection diagnostics; legacy/shadow defaults remain unchanged.

## v0.26.12

- Added engine-mode readiness validation and conservative safety policy controls under `engine.engine_mode` for experimental `engine.trip_source: engine` output selection.
- Engine mode now rejects unsafe candidate output by default with clear reasons (readiness, empty candidates, unmatched legacy rate, boundary drift, low-confidence/gap-affected candidates).
- Added explicit optional fallback control `engine.engine_mode.fallback_to_legacy_on_reject`; no silent fallback occurs unless enabled.
- Added engine selection diagnostics support (`*_engine_selection.csv`) including requested/selected source, accepted/rejected, fallback usage, readiness, and reasons.
- Default behaviour remains unchanged (`engine.trip_source: legacy`).

## v0.26.11

- Added explicit `engine.trip_source` selection with supported values `legacy`, `shadow`, and `engine`, defaulting to `legacy`.
- Preserved default processing behaviour: official valid/jitter outputs continue to come from the legacy pipeline unless `engine.trip_source: engine` is explicitly set.
- Added candidate-trip adapter for experimental engine mode to convert engine candidate trips into existing output trip rows without CSV header changes.
- Kept route observations/anomalies working in engine mode by applying existing route rules to adapted engine-output trips.
- Added validation error for invalid `engine.trip_source` values with clear allowed values.
- Updated engine docs with trip-source mode guidance and recommended shadow-first workflow for testing.

## v0.26.10.1

- Preserved baseline candidate-vs-legacy trip comparison diagnostics when `engine.trip_builder.compare_legacy` is enabled even if shadow summary toggles are off, so trip comparison metrics continue to report expected counts.
- Populated `speed_kmh` in optional `_engine_motion.csv` diagnostics from passive engine point speed evidence while preserving existing headers and opt-in behaviour.
- Hardened `engine.Result.Diagnostics()` snapshot cloning by deep-copying nested diagnostics slices (reasons, point indexes, unmatched trip lists, and shadow summary metrics/mismatches).
- Added regression tests covering legacy comparison with shadow summary disabled, non-blank motion speed diagnostics, and deep-copy snapshot immutability.

## v0.26.10

- Added passive shadow-engine comparison summaries with deterministic readiness ratings and mismatch diagnostics (unmatched trips, boundary drift, and optional origin/destination mismatches) for candidate-vs-legacy trip auditing.
- Added configurable `engine.shadow` thresholds/tolerances and optional diagnostic CSV outputs for shadow summary and mismatch rows, disabled by default.
- Kept official processed valid/jitter trip outputs unchanged; shadow summary remains advisory-only.

## v0.26.9

- Added optional engine diagnostic CSV outputs for passive engine stages (points, motion, stays, visits, excursions, candidate trips, and trip comparison) behind `engine.audit` toggles.
- Diagnostics are disabled by default; existing processed/expanded/jitter/route/audit outputs remain unchanged unless explicitly enabled.
- Added engine/output tests covering default-disabled behaviour and diagnostic header writing.

## v0.26.8

- Added a passive replacement trip-builder prototype in `engine` that derives candidate trips from visit/excursion diagnostics with deterministic type, confidence, reason, and boundary-quality annotations.
- Added lightweight candidate-vs-legacy comparison diagnostics for auditing approximate match counts, unmatched trips, boundary-time drift, and site mismatches.
- Added `engine.trip_builder` config scaffolding with safe defaults (`passive_only: true`) so existing processing behaviour remains unchanged unless explicitly enabled for diagnostics.
- Added focused tests for candidate trip typing/confidence, deterministic generation, and candidate-vs-legacy comparison while preserving compatibility outputs.

## v0.26.7.1

- Fixed process pipeline error handling so `gogator process` and `gogator process gps` now propagate `engine.Run` failures instead of silently returning empty success-style outputs.
- Fixed config parsing for nested sections so sibling keys after nested blocks (for example under `engine`) resolve to the correct parent section.
- Fixed engine evidence sequencing to build passive evidence after point normalization/sort and delta recalculation so diagnostics align with normalized output points.
- Fixed motion hysteresis recovery so `Gap`/`Noise` classifications no longer poison the active state and block immediate recovery on the next valid sample.
- Fixed stay detection to enforce `engine.stay_detection.min_points` before emitting stays from stationary/noise clusters.
- Added focused regression tests for each of the above fixes across app, config, motion, stays, and engine run evidence alignment.

## v0.26.7

- Added passive engine visit modelling from detected stays, including known-site/unknown/pause/noise/home/supplier classifications, confidence, reasons, and stay traceability.
- Added passive engine excursion modelling between visits with movement-leg diagnostics, duration/distance estimates, confidence, and short out-and-back/supplier/gap-affected candidate tagging.
- Added engine config toggles for visits/excursions and safe defaults that preserve existing trip pipeline behaviour.
- Added tests for deterministic visit/excursion classification and compatibility parity to ensure valid/jitter/route outputs remain unchanged.

## v0.26.6

- Added passive engine stay/dwell detection with deterministic stationary clustering based on motion state, point quality, duration, radius/spread, point count, and gap evidence.
- Added structured stay diagnostics (`Stay`, types, confidence, reasons, representative coordinates, durations, and source point traceability) in `engine.Result` without forcing caller adoption.
- Added optional passive known-site proximity tagging for stays and optional gap-inferred stop diagnostics.
- Added stay detection config defaults/parsing under `engine.stay_detection` and kept defaults behavior-neutral.
- Added stay-focused engine tests and run-level parity checks to assert existing valid/jitter/route outputs remain unchanged.

## v0.26.5

- Added passive engine motion classification with deterministic hysteresis across `Moving`, `Stationary`, `Unknown`, `Gap`, and `Noise` states.
- Motion classification uses existing passive evidence/quality plus speed, tracker movement/idling signals (`io24`, `io251`), coordinate validity, duplicate timestamps, and time-gap detection.
- Added passive motion diagnostics to `engine.Result` as samples and segments; callers are not required to consume them and processing behaviour is unchanged.
- Added engine motion tests for stable stationary/moving runs, hysteresis against one-point spikes/pauses, noise/gap handling, determinism, and run-level passive parity.

## v0.26.4

- Added passive engine evidence extraction for per-point tracker signals in the new engine evidence model (`PointEvidence`, `SignalEvidence`, `EvidenceSet`) without changing current trip decisions.
- Added deterministic passive point quality scoring (`QualityScore`, `QualityReason`) with `Good`/`Usable`/`Poor`/`Invalid`/`Unknown` bands and conservative checks for coordinate validity, GPS quality hints, timestamp continuity, and implausible jumps.
- Wired passive evidence into `engine.Result` so callers can inspect diagnostics without requiring adoption; existing process/CSV/route behaviour remains unchanged.
- Added engine tests covering missing signal fields, quality classifications, duplicate timestamp handling, determinism, and passive run parity.
- Updated engine documentation for passive instrumentation and clarified future staged use by motion/stay/trip logic.

## v0.26.3

- Added engine configuration scaffolding with safe defaults for `engine`, `engine.stay_detection`, `engine.motion`, `engine.quality`, `engine.audit`, `valhalla`, `h3`, and `postgis`.
- Wired engine-related config into `engine.Input` so `engine.Run` receives future-facing toggles without changing current processing behaviour.
- Added config and engine coverage to assert old configs still load, new sections parse with safe defaults, and engine compatibility behaviour remains unchanged.
- Updated engine docs and sample config to mark Valhalla/H3/PostGIS as intentionally inactive placeholders in this release.

## v0.28

- Moved the processing orchestration seam fully into `engine.Run`, keeping file/config loading and output writing in `internal/app/process.go`.
- Expanded `engine` compatibility tests to assert parity for route application and jitter split outputs through the engine boundary.
- Updated `engine/README.md` to document the compatibility-wrapper status, future seam intent, and explicit non-goals for this run.

## v0.27

- Added a new root-level `engine/` package with `engine.Run(ctx, input)` as an isolated processing seam.
- Wired `internal/app/process.go` to execute the existing pipeline through the new engine seam with no intended behaviour change.
- Added engine parity tests with a synthetic dataset to assert matching trips, jitter splitting, route observations, and route anomalies against the legacy sequence.

## v0.26

- Implemented the first `gogator process gps` SQLite bridge: it reads GPS points, sites, and routes from `gogator.sqlite`, reuses the existing in-memory process pipeline, and writes the standard `gogator_*` process CSV output family.
- Added process-gps store readers for SQLite GPS points, sites, and routes.
- Added CLI and smoke coverage for `process gps`.
- Added recognised-but-stubbed Google vendor commands: `load google` and `dump google`, both returning `not implemented yet` for now.
- Renamed vendor tracker interchange commands: `import raw` is now `load gator`, and `export raw` is now `dump gator`.
- Removed the old `import raw` and `export raw` command forms; GoGator is still alpha before v1.0, so no backwards compatibility alias is kept.
- Implemented `gogator db backup [[as] file]`, creating a new SQLite backup file with SQLite `VACUUM INTO`.
- Implemented `gogator db vacuum` to compact the default SQLite database in place.
- Added focused store tests, CLI tests, and smoke coverage for database backup and vacuum.
- Documented that Gator dumps currently preserve accurate representation rather than exact source text: trailing numeric zeroes may be cropped and parameter order may differ.
- Implemented `gogator export raw`, writing tracker-style CSV with columns `dt,lat,lng,altitude,angle,speed,params`.
- Wired `export raw [[as] file]` through the CLI and updated help wording from planned to implemented.
- Added raw export tests covering CSV shape, preserved raw params, deterministic row order, and re-import safety.
- Added smoke coverage for raw CSV export and re-import.
- Documented that raw export currently preserves semantic numeric values, not exact original decimal formatting, because SQLite stores parsed floats.
- Added SQLite Run 2 site commands: `import sites`, `export sites`, `add site`, and `delete site` (SQLite-backed).
- Added SQLite foundation with a new internal store package and default database path `gogator.sqlite`.
- Implemented `gogator db init` to create the SQLite database schema and indexes idempotently.
- Implemented `gogator db status` to report database path, SQLite version, and core table counts (`gps_points`, `sites`, `routes`, `processing_runs`, `trips`, `issues`).
- Added focused tests for DB init idempotency and DB status baseline counts.

## v0.23

- Excluded route anomalies from `*_route_observations.csv` summaries so outlier trips no longer pollute learned route min/max/median values.
- Kept route anomalies in `*_route_anomalies.csv` for review while preventing them from being promoted through `gogator add_route` by accident.
- Split same-site jitter rows into a separate `*_jitter_same_site.csv` output.
- `*_jitter.csv` now contains the remaining jitter rows that are more likely to need manual review.
- Added `*_jitter_same_site.csv` to `.gitignore` with the other generated outputs.

## v0.22

- Moved `Continuity Status` in processed trip output to immediately after `Site Duration`.
- Replaced the processed output `Source` column with `Filename` as column 1.
- `Filename` is now populated from the raw GPS file that contributed the trip.
- Expanded output also includes a `Filename` column for each raw point.
- When multiple raw GPS files are supplied to `gogator process`, GoGator now reads all files, normalises timestamps, sorts all points by time, recalculates point deltas, and processes them as one combined timeline.
- Multi-file output is written once using a combined filename range such as `<first>_to_<last>_processed.csv`, rather than one output set per input file.

## v0.21

- Added `trip_detection.ignition_analysis` with supported values `auto`, `wired`, and `unwired`.
- Default ignition analysis mode is `auto`.
- In `auto`, GoGator trusts ignition only when raw data contains both ignition-on and ignition-off values in `io1`; otherwise it behaves as `unwired`.
- In `wired`, `io1=0` is treated as strong stationary/parked evidence and suppresses movement classification for that point.
- In `unwired`, ignition is ignored and GoGator continues using speed, `io24`, dwell, and GPS evidence.

## v0.20

- Added support for `Type` and `Important` columns in `addresses.csv`.
- `Type` is arbitrary user text for classifying a site, such as `Customer`, `Vendor`, `Home`, or any other label.
- `Important` accepts mixed-case yes/no style values; `yes`, `y`, `true`, and `1` are treated as important.
- When the `Important` column is omitted, sites default to important for backwards compatibility.
- Added an important-site ledger pass that treats unimportant sites and `CHECK` rows as pass-through noise, collapsing them into the next important site where possible.
- Same-important-site loops created by unimportant fragments are moved to jitter instead of the main processed output.
- Set `go.mod` to Go 1.25 to match the target development version.

## v0.19

- Changed the default raw timestamp correction to `raw_time.correction_hours: 0` and `raw_time.source: gator_raw_utc`.
- Added a warning when `raw_time.correction_hours` is non-zero because timestamp shifts can move trips onto the wrong local calendar day.
- Updated `gogator process` to accept more than one raw GPS CSV in a single command, writing the normal output set for each input file.
- Added sliding-window dwell evidence for site matching, replacing brittle per-point dwell interval handling that could ignore exact hourly tracker pings.
- Added site-matching config for dwell windows, inside-site ratio, stationary ratio, and maximum sample gap.
- Added adjacent-trip continuity repair so `CHECK` can be repaired when the previous destination and next departure are the same physical place.
- Added `Continuity Status` to processed trip output to distinguish normal rows, repaired continuity, unresolved checks, GPS gaps, and known-site conflicts.
- Updated timestamp parsing comments to describe the default UTC tracker timestamp model and the legacy/emergency nature of non-zero correction values.

## v0.14

- Added `TRACKER_SIGNALS.md` with raw params, movement/idling interpretation, GPS quality notes, odometer notes, crash/driving-style fields, and accelerometer axis meanings.
- Added `DESIGN_NOTES.md` explaining why raw GPS is canonical, why Gator processed drive/stop exports are not trusted as truth, and how GoGator should preserve source evidence.
- Updated `README.md`, `AGENTS.md`, `CODEX.md`, and `COMMANDS.md` with richer context so future Codex/GitHub work does not lose detail.
- Documented accelerometer fields: `g0` X-axis left/right, `g1` Y-axis forward/back, and `g2` Z-axis up/down.
- Reinforced that useful raw tracker signals should be preserved in expanded/audit outputs rather than discarded.

## v0.13

- Renamed the Go project from `gatorlog` to `GoGator`.
- Renamed the module to `gogator` and moved the CLI entrypoint to `cmd/gogator`.
- Added the `gogator commands` extended help command with examples and command intent.
- Updated command help to use the `gogator` binary name.
- Changed the default config filename to `gogator.yaml`, with fallback support for legacy `gatorlog.yaml`.
- Added `GOGATOR_TIMEZONE` support while retaining legacy `GATORLOG_TIMEZONE` fallback.
- Updated processed output `Source` value from `gatorlog` to `GoGator`.
- Added Codex/agent context docs: `AGENTS.md`, `CODEX.md`, and `COMMANDS.md`.
- Set `go.mod` to Go 1.22 to match the target development version.
- Fixed raw row numbering for headed CSV files so the first data row is row 2, matching the source file line number.
- Removed the old compiled `gatorlog` binary from the source package.

## v0.12 and earlier inherited behaviour

- Preserved silent stop gap handling.
- Preserved noisy/observed GPS output rather than replacing it with canonical site GPS.
- Preserved destination dwell validation, stationary teleport guard, same-site micro-trip suppression, route observations, and route anomalies.

- Implemented SQLite-backed site commands: import sites, export sites, add site, delete site.

## v0.26.18

- Performed a focused engine stabilisation/cleanup pass without expanding production behaviour or changing default legacy processing/output paths.
- Added focused readiness regression coverage for explicit/non-silent fallback behaviour in engine mode (`fallback_to_legacy_on_reject`), including both reject-without-fallback and reject-with-explicit-fallback cases.
- Verified and preserved conservative defaults across optional systems (`engine.trip_source: legacy`, diagnostics off by default, Valhalla off by default, PostGIS off by default).
- Refreshed `engine/README.md` to clearly separate implemented behaviour, diagnostic-only paths, and experimental/opt-in paths, and documented a recommended safe legacy→diagnostics→shadow→engine workflow.
- Kept official processed/jitter CSV schemas, CLI behaviour, and default output selection unchanged.

## v0.26.19

- Added a focused non-production validation pack for comparing legacy/shadow/engine behaviour via deterministic engine tests and validation metrics, without changing CLI defaults or processed output schemas.
- Added `engine/VALIDATION.md` with a safe real-data validation workflow (legacy first, shadow diagnostics, optional guarded engine mode), key metrics to inspect, and promotion guidance.
- Kept defaults conservative and unchanged (`engine.trip_source: legacy`, Valhalla/PostGIS/diagnostics disabled by default), with no external service requirement for validation.
