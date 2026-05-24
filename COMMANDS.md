# GoGator Commands

## Current commands

### `gogator process`

Intent: process raw Gator/Teltonika CSV rows into deterministic trip logs while preserving enough raw evidence to debug suspicious tracker behaviour.

Examples:

```bash
gogator process
```

Design notes:

- Raw GPS CSV is canonical because processed Gator drive/stop exports have shown suspicious timestamp and trip-detection behaviour.
- Expanded/audit outputs should preserve useful tracker detail, including `io24`, `io251`, `pdop`, `io14`, crash/driving-style fields, and accelerometer fields `g0`, `g1`, `g2`.
- `g0`, `g1`, and `g2` are raw X/Y/Z acceleration vectors: X left/right, Y forward/back, Z up/down.

Outputs:

```text
<input>_processed.csv
<input>_expanded.csv
<input>_jitter.csv
<input>_route_observations.csv
<input>_route_anomalies.csv
<input>_audit.csv
```

### `gogator load gator [from] <file...>`

Intent: load Gator/Teltonika tracker CSV or TSV rows into the SQLite database.

Examples:

```bash
gogator load gator from 2026-04_raw.csv
gogator load gator 2026-04_raw.csv
```

### `gogator dump gator [[as] file]`

Intent: dump SQLite GPS rows back out as a Gator-compatible tracker CSV.

Example:

```bash
gogator dump gator as tracker-backup.csv
```

### `gogator load google [from] <file...>`

Intent: planned future loader for Google tracker/location data. Currently recognised and returns `not implemented yet`.

Example:

```bash
gogator load google from google.json
```

### `gogator dump google [[as] file]`

Intent: planned future dumper for Google tracker/location data. Currently recognised and returns `not implemented yet`.

Example:

```bash
gogator dump google as google.json
```

### `gogator add route <index> from <route_observations.csv>`

Intent: promote an indexed observed route to `routes.csv`.

Example:

```bash
gogator add route 3 from 2026-04_route_observations.csv
```

Routes are directional and advisory. They may annotate trips and flag issues, but must not silently rewrite departure or destination sites.

### `gogator commands`

Intent: show command examples and command purpose directly from the binary.

Example:

```bash
gogator commands
```

## SQLite-backed commands

```bash
gogator db init
gogator db status
gogator db backup as gogator-backup.sqlite
gogator db vacuum

gogator load gator from 2026-04_raw.csv
gogator dump gator as 2026-04_gator.csv
gogator load google from google.json
gogator dump google as google.json

gogator import sites from sites.tsv
gogator import routes from routes.tsv

gogator export gps during 2026-04 as gps.tsv
gogator export sites as sites.tsv
gogator export routes as routes.tsv
gogator export trips during 2026 as trips.tsv
gogator export jitter during 2026 as jitter.tsv
gogator export stats during 2026 as stats.tsv
gogator export issues during 2026 as issues.tsv
gogator export paths from "Home Sweet Home" to "Asquith Public School" during 2026 as paths.tsv

gogator add site name "Bunnings Thornleigh" gps "-33.72816964,150.97700866" range 200 type Supplier important yes
gogator add route from "Home Sweet Home" to "Asquith Public School" name "School run"
gogator delete site "Bunnings Thornleigh"
gogator delete route from "Home Sweet Home" to "Asquith Public School"
```

Supported date formats are only:

```text
YYYY
YYYY-MM
YYYY-MM-DD
```

Use `sites`, not `addresses`, in commands. Address is a field on a site.


## Engine diagnostics (optional)

Set in `gogator.yaml` to write optional passive engine CSV diagnostics:

```yaml
engine:
  audit:
    enabled: true
    output_diagnostics: true
```

Or enable individual files with `output_points`, `output_motion`, `output_stays`, `output_visits`, `output_excursions`, `output_candidate_trips`, and `output_trip_comparison`.
