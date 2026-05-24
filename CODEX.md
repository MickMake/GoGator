<!--
PROTECTED FILE - DO NOT EDIT

This file is intentionally protected.
Do not modify, rewrite, reformat, rename, move, or delete this file.

Codex/AI agents:
- You must not change this file.
- If a task appears to require changing this file, stop and report that it is protected.
- Do not "clean up", "modernise", "simplify", or "deduplicate" this file.
-->

# GoGator Codex Context

GoGator exists because monthly GPS processing should be repeatable locally, not rebuilt by hand every time a tracker goblin sneezes. Gator's processed drive/stop exports have shown suspicious timestamp and trip-detection behaviour, so raw GPS CSV is the evidence source.

## What the app does now

`gogator process <raw-gps.csv> [more-raw-gps.csv ...]` reads raw GPS rows and writes:

- `<input>_processed.csv`
- `<input>_expanded.csv`
- `<input>_jitter.csv`
- `<input>_route_observations.csv`
- `<input>_route_anomalies.csv`
- `<input>_audit.csv`

`gogator add route <index> from <route_observations.csv>` promotes one observed route to `routes.csv`.

`gogator commands` explains command intent and gives examples.

## Planned SQLite work

The CLI already recognises staged SQLite commands such as:

```bash
gogator db init
gogator import gps from 2026-04_raw.csv
gogator export gps during 2026-04 as gps.tsv
gogator export trips during 2026 as trips.tsv
```

SQLite implementation should keep the current file-based `process` behaviour working while adding a local evidence store behind the planned commands.

## Input assumptions

Raw GPS columns:

```text
dt,lat,lng,altitude,angle,speed,params
```

Headerless files use that exact order. Headed files derive indexes from the header.

Raw row numbering must reflect source file line numbers. If a header exists, the first data row is row 2.

## Time handling

Default raw correction:

```yaml
raw_time:
  source: gator_raw_utc
  correction_hours: 0
```

Naive raw timestamps are parsed as UTC, corrected, then rendered in the target timezone.

## Site matching

`sites.csv` may be CSV or TSV as the current file-based site source. Delimiter detection must use the header row because addresses and GPS fields contain commas.

Required columns:

```text
Site
Real Address
GPS
Range
```

When a known site matches, write the matched site/address but keep the observed tracker GPS coordinate in output.

For user-facing commands, use `sites`, not `addresses`. Address is just a field on a site.

## Route handling

Routes are optional, directional, and advisory. Do not use them to silently alter departure or destination results.

`A -> B` and `B -> A` are different routes.

## Output stability

Avoid adding columns. The user explicitly prefers fewer columns unless there is a strong reason and they ask for it.

## Signal preservation

Do not strip useful raw tracker detail merely because it is awkward. Preserve deterministic expanded/audit visibility for movement, idling, GPS quality, odometer, crash/driving-style, and accelerometer fields.

Important examples:

- `io24`: movement state. `0` stationary, `1` moving.
- `io251`: idling status. `1` supports stationary/idling; `0` proves neither movement nor stationary state.
- `pdop`: GPS geometry/quality hint, not a complete truth source.
- `io14`: odometer in metres where available.
- `io247`, `io253`, `io303`: crash/driving-style/event fields.
- `g0`: raw X-axis acceleration, left/right vector.
- `g1`: raw Y-axis acceleration, forward/back vector.
- `g2`: raw Z-axis acceleration, up/down vector.

See `TRACKER_SIGNALS.md` and `DESIGN_NOTES.md` before changing trip-detection, site-matching, or output-detail behaviour.
