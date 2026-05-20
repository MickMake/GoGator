# GoGator Codex Context

GoGator exists because monthly GPS processing should be repeatable locally, not rebuilt by hand every time a tracker goblin sneezes.

## What the app does

`gogator process <raw-gps.csv>` reads raw GPS rows and writes:

- `<input>_processed.csv`
- `<input>_expanded.csv`
- `<input>_jitter.csv`
- `<input>_route_observations.csv`
- `<input>_route_anomalies.csv`
- `<input>_audit.csv`

`gogator add_route <route_observations.csv> <index>` appends one observed route to `routes.csv`.

`gogator commands` explains command intent and gives examples.

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
  correction_hours: -24
```

Naive raw timestamps are parsed as UTC, corrected, then rendered in the target timezone.

## Site matching

`addresses.csv` may be CSV or TSV. Delimiter detection must use the header row because addresses and GPS fields contain commas.

Required columns:

```text
Site
Real Address
GPS
Range
```

When a known site matches, write the matched site/address but keep the observed tracker GPS coordinate in output.

## Route handling

Routes are optional and advisory. Do not use them to silently alter destination results.

## Output stability

Avoid adding columns. The user explicitly prefers fewer columns unless there is a strong reason and they ask for it.
