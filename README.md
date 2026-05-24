# GoGator

GoGator turns raw Gator/Teltonika GPS tracker evidence into deterministic, auditable, spreadsheet-friendly trip records for a trade vehicle.

Raw GPS observations are canonical evidence. GoGator must preserve enough evidence to explain each trip decision.

## Project authority

`PROJECT_INTENT.md` is the gold-standard project specification.

If any other document, implementation detail, or legacy behaviour conflicts with `PROJECT_INTENT.md`, the project intent wins.

## Runtime direction

The intended architecture is:

~~~text
input/load layer -> database -> engine -> output writers
~~~

The `engine` package owns trip construction.

Current command names and output files should remain available unless explicitly changed, but command stability must not preserve old trip-construction internals.

## Basic usage

~~~bash
gogator process <raw-gps.csv>
gogator process <raw-gps.csv> <more-raw-gps.csv> ...
gogator commands
~~~

## Current compatibility note

The file-based processing workflow remains available as part of the user-facing contract.

CSV/TSV files such as `sites.csv` and `routes.csv` are compatibility, import/export, review, and pre-population paths. They are not the long-term runtime source of truth.

## Reference documents

Read these in order:

1. `PROJECT_INTENT.md` — authoritative project intent
2. `COMMANDS.md` — command examples and command surface
3. `ROUTE_MAPPING_PROCEDURE.md` — logical trip-decision flow
4. `TRACKER_SIGNALS.md` — raw tracker params and signal meanings
5. `DESIGN_NOTES.md` — deeper rationale and audit principles
6. `engine/README.md` — engine package direction

## Build and test

~~~bash
go test ./...
go build -o gogator ./cmd/gogator
./gogator commands
~~~