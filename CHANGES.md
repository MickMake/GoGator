# Changes

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
