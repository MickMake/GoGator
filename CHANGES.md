# Changes

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
