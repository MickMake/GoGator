# GoGator

Deterministic Go CLI processor for Gator/Teltonika raw GPS CSV exports.

GoGator treats the raw GPS export as canonical. Gator's processed drive/stop exports are useful as a comparison point, but they are not trusted as the source of truth. This is how we avoid giving spreadsheet goblins a clipboard and authority.

## Usage

```bash
gogator process <raw-gps.csv>
```

Optional overrides:

```bash
gogator process <raw-gps.csv> --timezone Australia/Sydney --config gogator.yaml --sites addresses.csv --routes routes.csv
```

Extended command help:

```bash
gogator commands
gogator process --help
gogator add_route --help
```

## Defaults

- `config` defaults to `gogator.yaml`, falling back to `gatorlog.yaml` for older folders.
- `sites` defaults to the config value, otherwise `addresses.csv`.
- `routes` defaults to the config value, otherwise `routes.csv`.
- timezone priority is `--timezone`, then `GOGATOR_TIMEZONE`, then legacy `GATORLOG_TIMEZONE`, then config, then `Australia/Sydney`.
- if `addresses.csv` or `routes.csv` are not found in the current working directory, the app also checks beside the input raw GPS file.

## Commands

### `gogator process <raw-gps.csv>`

Intent: convert raw Gator/Teltonika GPS rows into deterministic, spreadsheet-ready outputs.

Examples:

```bash
gogator process 2026-04_raw.csv
gogator process exports/2026-04_raw.csv --config gogator.yaml --sites addresses.csv --routes routes.csv
```

### `gogator add_route <route_observations.csv> <index>`

Intent: promote one observed route into `routes.csv` so future runs can recognise it as expected.

Examples:

```bash
gogator add_route 2026-04_route_observations.csv 3
gogator add_route 2026-04_route_observations.csv 3 --routes my_routes.csv
```

Routes are advisory only. They can confirm common routes and flag anomalies, but must not silently rewrite destinations.

## Raw GPS input

Supported raw GPS fields:

```text
dt,lat,lng,altitude,angle,speed,params
```

Headed and headerless files are supported. Headerless files are assumed to use the field order above.

Raw row numbering matches source file line numbers:

- headed file: row 1 is the header and first data row is raw row 2
- headerless file: first data row is raw row 1

## Address/site CSV

Supported minimum columns:

```csv
Site,Real Address,GPS,Range
Home Sweet Home,"28 New Line Rd, West Pennant Hills NSW 2125, Australia","-33.74154166687418, 151.04808520098027",100
```

`Range` is metres. If blank, the config default radius is used.

The processor matches the nearest site within that row's radius and writes the actual `Site` value from `addresses.csv`, e.g. `Home Sweet Home`, not just `Home`.

Important GPS output rule: matched sites keep the observed/noisy GPS coordinate from the tracker row or cluster. GoGator does **not** replace output GPS values with canonical GPS from `addresses.csv`.

Google Sheets may export the Addresses tab as TSV while still naming it `addresses.csv`. GoGator detects the delimiter from the header row so address fields containing commas do not trick the parser into reading zero sites.

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

The audit file includes loaded sites/routes and run settings.

## Time correction

The raw GPS export can contain timestamps that are not the same clock as the vehicle's real local time. GoGator treats naive raw timestamps as UTC, applies `raw_time.correction_hours`, then renders the result in the configured timezone.

Default:

```yaml
raw_time:
  source: gator_raw_utc_plus_one_day
  correction_hours: -24
```

Verified April 2026 examples:

- raw `2026-04-01 20:09:04` -> Sydney `2026-04-01 07:09:04`
- raw `2026-04-02 04:31:53` -> Sydney `2026-04-01 15:31:53`

If a future Gator export changes its clock behaviour, adjust only `raw_time.correction_hours` in `gogator.yaml` and re-run.

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

Processed trip rows include:

```text
Route Name
Route Confidence
Route Match Status
Route Expected Distance Range
Route Expected Duration Range
Route Notes
```

## GPS sanity filters

GoGator includes safeguards for false stationary GPS clusters:

- `stationary_teleport_guard_enabled`: filters stationary/idling points that jump far away from the last known site while odometer and motion signals say the vehicle did not move.
- `same_site_micro_trip_*`: suppresses tiny local loops where the previous destination and current destination are the same known site, but a short CHECK segment appears in between.

Useful defaults in `gogator.yaml`:

```yaml
trip_detection:
  stationary_teleport_guard_enabled: true
  stationary_teleport_min_jump_m: 250
  stationary_teleport_requires_odometer: true
  same_site_micro_trip_max_km: 1.5
  same_site_micro_trip_max_minutes: 10
  same_site_guard_radius_m: 750
```

## Stationary dwell destinations

Known sites can include site-level destination dwell rules in `addresses.csv`/TSV:

```text
Site	Real Address	GPS	Range	Min Destination Minutes	Site Type
Home Sweet Home	28 New Line Rd...	-33.74154166687418, 151.04808520098027	100	3	Home
Bunnings Castle Hill	14 Victoria Ave...	-33.72816963501526, 150.97700866421283	200	8	Supplier
```

`Min Destination Minutes` means minimum stationary dwell time inside the site radius. Time spent moving slowly through the radius does not count as destination dwell.

Relevant defaults:

```yaml
site_matching:
  default_radius_m: 100
  default_min_destination_minutes: 5
  unknown_check_min_destination_minutes: 10
  stationary_dwell_ratio_required: 0.70
  infer_silent_stop_gaps: true
  silent_stop_min_gap_minutes: 5
```

## Build and test

```bash
go test ./...
go build -o gogator ./cmd/gogator
```
