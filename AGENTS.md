# GoGator Agent Instructions

Before changing anything, read `AUTHORITATIVE_SPECIFICATION.md`.

That file is the controlling project specification. If this file, code, tests, generated output, or legacy behaviour conflicts with `AUTHORITATIVE_SPECIFICATION.md`, the authoritative specification wins.

## Read first

1. `AUTHORITATIVE_SPECIFICATION.md`
2. `ROUTE_MAPPING_PROCEDURE.md`
3. `TRACKER_SIGNALS.md`
4. `engine/README.md`
5. relevant source files

## Protected files

Do not edit, rename, move, reformat, summarise, or delete protected files unless the owner explicitly gives the matching approval phrase.

Protected files:

- `AUTHORITATIVE_SPECIFICATION.md`
- `AGENTS.md`
- `CODEX.md`
- `ROUTE_MAPPING_PROCEDURE.md`
- `TRACKER_SIGNALS.md`
- `engine/README.md`

Append-only:

- `CHANGES.md`

## Non-negotiable agent rules

- Do not make legacy GPS output authoritative.
- Do not use old GPS output as the correctness oracle.
- Do not remove current CLI commands without explicit instruction.
- Do not silently change processed CSV columns.
- Do not make CSV files the runtime model.
- Do not treat `routes.csv` as route authority.
- Do not let routes rewrite site truth.
- Do not hide low-confidence data.
- Do not remove auditability.
- Do not make external services mandatory for ordinary tests unless explicitly requested.

## Required checks

~~~bash
go test ./...
go build -o gogator ./cmd/gogator
./gogator commands
~~~

## Required final report

~~~text
Goal:
Files changed:
Behaviour changes:
Output/schema changes:
Tests run:
Pass/fail:
Known limitations:
Suggested next step:
~~~