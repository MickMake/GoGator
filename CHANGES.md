# Changes

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
