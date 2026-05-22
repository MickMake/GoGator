# Changes

## v0.26

- Added SQLite foundation with a new internal store package and default database path `gogator.sqlite`.
- Implemented `gogator db init` to create the SQLite database schema and indexes idempotently.
- Implemented `gogator db status` to report database path, SQLite version, and core table counts (`gps_points`, `sites`, `routes`, `processing_runs`, `trips`, `issues`).
- Added focused tests for DB init idempotency and DB status baseline counts.
- Corrective: aligned the v0.26 Run 1 SQLite schema in `internal/store` with the Codex prompt naming and structure (including required tables, key columns, foreign keys, and required indexes).

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
