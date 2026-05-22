# GoGator

Deterministic Go CLI processor for Gator/Teltonika raw GPS CSV exports.

GoGator treats the raw GPS export as canonical. Gator processed drive/stop exports are useful as a comparison point, but they are not trusted as the source of truth because they have shown suspicious timestamp and trip-detection behaviour.

## Usage

```bash
gogator process <raw-gps.csv>
gogator process <raw-gps.csv> <more-raw-gps.csv> ...
```

Optional overrides:

```bash
gogator process <raw-gps.csv> --timezone Australia/Sydney --config gogator.yaml --sites sites.csv --routes routes.csv
```

Extended command help:

```bash
gogator
gogator commands
gogator process --help
gogator add route 3 from 2026-04_route_observations.csv
```

## Defaults

- `config` defaults to `gogator.yaml`.
- `sites` defaults to the config value, otherwise `sites.csv`.
- `routes` defaults to the config value, otherwise `routes.csv`.
- timezone priority is `--timezone`, then `GOGATOR_TIMEZONE`, then config, then `Australia/Sydney`.
- if `sites.csv` or `routes.csv` are not found in the current working directory, the app also checks beside each input raw GPS file.

## Commands

### `gogator process <raw-gps.csv> [more-raw-gps.csv ...]`

Convert one or more raw Gator/Teltonika GPS files into deterministic, spreadsheet-ready outputs.

Examples:

```bash
gogator process 2026-04_raw.csv
gogator process 2026-04_raw.csv 2026-05_raw.csv 2026-06_raw.csv
gogator process exports/2026-04_raw.csv --config gogator.yaml --sites sites.csv --routes routes.csv
```

### `gogator add route <index> from <route_observations.csv>`

Promote one observed route into `routes.csv` so future runs can recognise it as expected.

Example:

```bash
gogator add route 3 from 2026-04_route_observations.csv
```

Routes are advisory only. They can confirm common routes and flag anomalies, but must not silently rewrite destinations.

## Raw GPS input

Supported raw GPS fields:

```text
dt,lat,lng,altitude,angle,speed,params
```

Headed and headerless files are supported. Headerless files are assumed to use the field order above.

## Sites CSV

Supported minimum columns:

```csv
Site,Real Address,GPS,Range
Home Sweet Home,"28 New Line Rd, West Pennant Hills NSW 2125, Australia","-33.74154166687418, 151.04808520098027",100
```

`Range` is metres. If blank, the config default radius is used.

The processor matches the nearest site within that row's radius and writes the actual `Site` value from `sites.csv`, e.g. `Home Sweet Home`, not just `Home`.

Important GPS output rule: matched sites keep the observed/noisy GPS coordinate from the tracker row or cluster. GoGator does **not** replace output GPS values with canonical GPS from `sites.csv`.

Supported dwell aliases:

```text
Min Destination Minutes
Min Stop Minutes
Min Dwell Minutes
Destination Min Dwell Minutes
```

## Outputs

Given `2026-04.CSV`, outputs are:

```text
2026-04_processed.csv
2026-04_expanded.csv
2026-04_jitter.csv
2026-04_route_observations.csv
2026-04_route_anomalies.csv
2026-04_audit.csv
```

The processed CSV includes a `Continuity Status` column. It reports whether adjacent trip chaining is clean, repaired, or suspicious.

The audit file includes loaded sites/routes and run settings.

## Time correction

GoGator treats naive raw timestamps as UTC, applies `raw_time.correction_hours`, then renders the result in the configured timezone.

Default:

```yaml
raw_time:
  source: gator_raw_utc
  correction_hours: 0
```

Non-zero correction values are emergency overrides. GoGator warns when a correction is configured because it can move trips onto the wrong local calendar day.

## Route rules

Optional `routes.csv` adds a deterministic route-confidence layer.

Example columns:

```csv
Route Name,From Site,To Site,Expected Distance Min Km,Expected Distance Max Km,Expected Duration Min Min,Expected Duration Max Min,Confidence Boost,Auto Merge Gap Min,Notes
Home Sweet Home to Asquith Public School,Home Sweet Home,Asquith Public School,8,16,10,40,Good,5,School route
```

The processor writes:

```text
<input>_route_observations.csv  # frequent observed routes, sorted by trip count
<input>_route_anomalies.csv     # unknown endpoints or trips outside approved route bands
```

## GPS sanity filters

GoGator includes safeguards for false stationary GPS clusters:

- `stationary_teleport_guard_enabled`: filters stationary/idling points that jump far away from the last known site while odometer and motion signals say the vehicle did not move.
- `same_site_micro_trip_*`: suppresses tiny local loops where the previous destination and current destination are the same known site, but a short CHECK segment appears in between.
- `continuity_repair_enabled`: repairs adjacent trip labels when one side is `CHECK`, the other side is known, and both GPS coordinates represent the same physical place.

## Reference docs

- `AGENTS.md`: project rules and source-map for agent/Codex work.
- `CODEX.md`: compact implementation context.
- `COMMANDS.md`: command intent and examples.
- `TRACKER_SIGNALS.md`: raw params, movement, quality, accelerometer, and crash/driving-style signal notes.
- `DESIGN_NOTES.md`: design intent, why raw GPS is canonical, and evidence-preservation rules.
- `CHANGES.md`: append-only project history.

## Build and test

```bash
go test -mod=vendor ./...
go build -mod=vendor -o gogator ./cmd/gogator
```
