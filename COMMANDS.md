# GoGator Commands

## Current commands

### `gogator process <raw-gps.csv> [more-raw-gps.csv ...]`

Intent: process raw Gator/Teltonika CSV rows into deterministic trip logs while preserving enough raw evidence to debug suspicious tracker behaviour.

Examples:

```bash
gogator process 2026-04_raw.csv
gogator process 2026-04_raw.csv --timezone Australia/Sydney
gogator process exports/2026-04_raw.csv --config gogator.yaml --sites addresses.csv --routes routes.csv
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

## Planned SQLite-backed commands

These commands are recognised by the CLI but will be implemented in staged SQLite work:

```bash
gogator db init
gogator db status

gogator import gps from 2026-04_raw.csv
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
