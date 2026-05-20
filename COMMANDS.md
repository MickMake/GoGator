# GoGator Commands

## `gogator process <raw-gps.csv>`

Intent: process raw Gator/Teltonika CSV rows into deterministic trip logs.

Examples:

```bash
gogator process 2026-04_raw.csv
gogator process 2026-04_raw.csv --timezone Australia/Sydney
gogator process exports/2026-04_raw.csv --config gogator.yaml --sites addresses.csv --routes routes.csv
```

Outputs:

```text
<input>_processed.csv
<input>_expanded.csv
<input>_jitter.csv
<input>_route_observations.csv
<input>_route_anomalies.csv
<input>_audit.csv
```

## `gogator add_route <route_observations.csv> <index>`

Intent: append an indexed observed route to `routes.csv`.

Examples:

```bash
gogator add_route 2026-04_route_observations.csv 3
gogator add_route 2026-04_route_observations.csv 3 --routes my_routes.csv
```

## `gogator commands`

Intent: show command examples and command purpose directly from the binary.

Example:

```bash
gogator commands
```
