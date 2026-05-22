# To Implement

This file tracks agreed GoGator work that is not implemented yet. It is intentionally practical rather than aspirational, because aspirational TODO files tend to breed in the walls.

## Command model

### Raw tracker data

`raw` means source/backup/round-trip tracker data.

Implemented command:

```bash
gogator import raw from tracker.csv
gogator export raw as tracker.csv
```

Planned commands/behaviour:

```bash
```

Rules:

- `import raw` replaces the old planned/current `import gps` meaning.
- `import gps` is not a command; use `import raw`.
- `import raw` ingests original Gator/Teltonika tracker CSV exports.
- Gator tracker exports are plain CSV.
- `import raw` should also accept TSV input.
- `export raw` exports tracker-style CSV columns: `dt,lat,lng,altitude,angle,speed,params`.
- `export raw` must export a round-trippable raw file.
- A raw CSV exported by GoGator should be suitable for importing again with `import raw`.
- For CSV source data, the intended ideal is:

```text
original Gator CSV -> import raw -> export raw -> diff original.csv exported.csv
```

with zero meaningful changes.

Notes:

- Preserve original raw fields and field ordering where possible.
- Avoid adding source metadata, audit columns, or GoGator-only columns to `export raw`.
- If exact byte-for-byte preservation proves impossible because of CSV quoting/newline differences, document the exact limitation and preserve semantic equality. The goal remains zero diff for normal Gator CSV exports.
- Current SQLite stores GPS numeric values as floats, so exported raw numeric formatting is semantic rather than byte-for-byte identical to the original imported text.

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
gogator process gps ...
gogator export trips ...
gogator export jitter ...
gogator export stats ...
gogator export issues ...
gogator export paths ...
gogator db backup
gogator db vacuum
gogator set gps params ...
gogator show gps params
gogator reset gps params
```

## Process GPS

Planned purpose:

```text
SQLite gps_points + sites + routes -> trips + classifications + route_stats + issues + waypoints
```

Current file-based workflow remains:

```bash
gogator process <raw-gps.csv...>
```

Future decision:

- Decide whether `process <raw-gps.csv...>` becomes a convenience wrapper around `import raw`, `process gps`, and default exports.
- No migration/backwards compatibility requirement yet because the tool is not actively in production use.

## Export commands after process gps

These become meaningful after `process gps` exists:

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
