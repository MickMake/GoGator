# To Implement

This file tracks agreed GoGator work that is not implemented yet. It is intentionally practical rather than aspirational, because aspirational TODO files tend to breed in the walls.

## Command model

### Vendor tracker data

Vendor tracker commands handle external tracker interchange formats.

Implemented command:

```bash
gogator load gator from tracker.csv
gogator dump gator as tracker.csv
```

Recognised but not implemented yet:

```bash
gogator load google from google.json
gogator dump google as google.json
```

Rules:

- `load gator` replaces `import raw`.
- `dump gator` replaces `export raw`.
- `gator` is currently implemented.
- `google` is recognised as a planned vendor and returns `not implemented yet`.
- `load gator` ingests original Gator/Teltonika tracker CSV exports.
- Gator tracker exports are plain CSV.
- `load gator` should also accept TSV input.
- `dump gator` exports tracker-style CSV columns: `dt,lat,lng,altitude,angle,speed,params`.
- `dump gator` must export a round-trippable Gator tracker file.
- A Gator CSV dumped by GoGator should be suitable for loading again with `load gator`.
- For CSV source data, the intended ideal is:

```text
original Gator CSV -> load gator -> dump gator -> diff original.csv dumped.csv
```

with zero meaningful changes.

Notes:

- Preserve original raw fields and field ordering where possible.
- Avoid adding source metadata, audit columns, or GoGator-only columns to `dump gator`.
- If exact byte-for-byte preservation proves impossible because of CSV quoting/newline differences, document the exact limitation and preserve semantic equality. The goal remains zero diff for normal Gator CSV exports.
- Current SQLite stores GPS numeric values as floats, so dumped Gator numeric formatting is semantic rather than byte-for-byte identical to the original loaded text.
- The accepted current Gator dump representation is accurate even when float formatting drops trailing zeroes and parameter order differs from the source file.

### Clean GPS data

`gps` means clean normalised GPS data for spreadsheet/reporting use.

Implemented command:

```bash
gogator export gps as gps.tsv
```

Current rules:

- Exports clean raw GPS point data from SQLite.
- Includes core GPS fields.
- Includes the full documented GPS parameter list in `gps.ParamOrder`.
- Appends unknown observed params alphabetically after documented params.
- Does not include audit/source metadata.
- Does not include `params_raw` or `params_json`.
- Is not currently intended to be a round-trip import format.

## Database administration

Implemented commands:

```bash
gogator db backup as gogator-backup.sqlite
gogator db vacuum
```

Current rules:

- `db backup` creates a new SQLite backup file using SQLite `VACUUM INTO`.
- `db backup` refuses to overwrite an existing destination file.
- `db vacuum` compacts the existing default database in place.

## Process GPS

Implemented command:

```bash
gogator process gps
```

Current behaviour:

- Reads GPS points already loaded into `gogator.sqlite`.
- Reads sites and routes from SQLite.
- Reuses the existing in-memory process pipeline rather than rewriting trip logic.
- Writes the standard process output family using the `gogator_` prefix:
  - `gogator_processed.csv`
  - `gogator_expanded.csv`
  - `gogator_jitter.csv`
  - `gogator_jitter_same_site.csv`
  - `gogator_route_observations.csv`
  - `gogator_route_anomalies.csv`
  - `gogator_audit.csv`
- Prints the same process summary and route anomaly/error count to stdout.
- Does not yet persist processed trips, waypoints, route stats, point classifications, or issues back into SQLite.

Current file-based workflow remains:

```bash
gogator process <raw-gps.csv...>
```

Future decision:

- Decide whether `process <raw-gps.csv...>` becomes a convenience wrapper around `load gator`, `process gps`, and default exports.
- Decide when processed output should be persisted into SQLite tables instead of, or as well as, CSV outputs.

## Settings foundation

Implemented schema:

```sql
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

Planned commands currently stubbed:

```bash
gogator set gps params io66,io67,io200
gogator show gps params
gogator reset gps params
```

Current behaviour:

- Commands are recognised.
- Commands return `not implemented yet`.
- Settings table is not yet used.

Planned behaviour:

- `set gps params ...` should persist a custom GPS export parameter column list.
- `show gps params` should show current GPS export parameter configuration.
- `reset gps params` should return GPS export params to default behaviour.
- Default behaviour should remain documented param order plus unknown observed params alphabetically.

## Remaining command surface

Not implemented yet:

```bash
gogator load google ...
gogator dump google ...
gogator export trips ...
gogator export jitter ...
gogator export stats ...
gogator export issues ...
gogator export paths ...
gogator set gps params ...
gogator show gps params
gogator reset gps params
```

## Export commands after process gps

These become meaningful after processed output is persisted in SQLite:

```bash
gogator export trips as trips.tsv
gogator export jitter as jitter.tsv
gogator export stats as stats.tsv
gogator export issues as issues.tsv
gogator export paths as paths.tsv
```

Each should have smoke-test coverage once implemented.

## Testing expectations

Every implemented command should have:

- focused Go tests where practical;
- CLI-level tests for command recognition and basic behaviour;
- smoke-test coverage in `scripts/smoke.sh` for the main functional path.

Validation commands:

```bash
go test -mod=vendor ./...
go build -mod=vendor -o gogator ./cmd/gogator
bash scripts/smoke.sh
```

Do not use:

```bash
go get
go mod tidy
go mod download
go install
```

unless explicitly requested.
