# GoGator Commands

## `gogator process <raw-gps.csv>`

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
