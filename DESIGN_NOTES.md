<!--
PROTECTED FILE - DO NOT EDIT

This file is intentionally protected.
Do not modify, rewrite, reformat, rename, move, or delete this file.

Codex/AI agents:
- You must not change this file.
- If a task appears to require changing this file, stop and report that it is protected.
- Do not "clean up", "modernise", "simplify", or "deduplicate" this file.
-->

# GoGator Design Notes

## Design goal

GoGator should produce deterministic, spreadsheet-friendly trip logs from raw Gator/Teltonika GPS exports.

The goal is not to guess perfectly. The goal is to make repeatable decisions, preserve enough evidence for review, and mark uncertain cases as `CHECK` instead of pretending the goblins have filled in a statutory declaration.

## Canonical data rule

Raw GPS observations are canonical source evidence.

Gator's processed "drives and stops" exports are not canonical because they have shown suspicious timestamp and trip-detection behaviour. They may be used as rough reference material, but GoGator must not rely on them for truth.

Once raw GPS data is loaded, the engine should operate from GoGator's database-backed representation of that evidence. Flat files are import/export interchange, not the engine's long-term operating model.

## Engine ownership principle

Trip construction belongs in `engine`.

The old `internal/gps` trip construction path is not the authority and must not be treated as the test oracle. Existing CLI commands and output files should remain available, but the trip-construction internals must move to the engine path.

A useful test asks whether the engine output is explainable, auditable, and consistent for the configured input data. It does not ask whether the engine reproduced whatever one old binary happened to emit.

## Database-first enrichment model

The engine enriches raw tracker observations using database-backed reference and learned data.

The database is the engine's operating source for:

- loaded GPS observations
- known sites
- site metadata, match radii, dwell thresholds, and audit history
- learned route signatures
- route frequencies
- route observations and anomalies
- processing runs and diagnostic evidence

Flat files are import/export mechanisms:

- `sites.csv` / TSV may pre-populate or export known sites.
- `routes.csv` may pre-populate or export known/approved routes.
- `gogator.yaml` may define deterministic thresholds and runtime defaults.

The engine must not depend on CSV files as its core model. CSV inputs are a way to seed, review, or exchange data. Runtime trip construction should use the database.

## Known-route learning principle

Routes are not primarily hand-authored rules.

The engine should learn known routes from repeated observed travel patterns. A route becomes "known" because it appears often enough, consistently enough, and with enough evidence to be useful.

`routes.csv` is only a pre-population/import/export path for route knowledge. It is not the long-term authority and must not limit the engine to manually curated routes.

Known-route learning should use evidence such as:

- repeated origin/destination pairs
- route signatures
- map-matched paths
- H3/spatial corridor patterns
- observed frequency
- duration and distance bands
- anomaly history

A matching route may increase confidence, annotate a trip, or flag an anomaly. It must not silently rewrite departure or destination sites.

## Preserve evidence

Avoid losing detail from the source data. In particular:

- preserve raw row numbers
- preserve raw timestamps and corrected rendered times where relevant
- preserve stable `params` columns in expanded output
- preserve movement/idling/GPS quality/accelerometer/crash-style fields
- preserve observed/noisy GPS coordinates in processed output
- avoid replacing observed evidence with cleaned canonical reference values

## Site matching principle

Site matching is database-first.

The engine must first match a raw point, stay, or stationary cluster against known sites stored in the GoGator database. Known sites are the primary site authority because they represent user-approved places, imported site records, and accumulated local knowledge.

If no known site is found with sufficient confidence, the engine may then use PostGIS-backed spatial/context matching as a fallback discovery and audit mechanism.

The matching order is:

1. known sites in the GoGator database
2. PostGIS spatial/context fallback when no known-site match is found
3. `CHECK` when there is still not enough durable evidence

When a match is made:

- write the matched site name from the database-backed known-site record when available
- write the real address from the known-site record when available
- keep the observed tracker GPS coordinate in the trip output
- store enough audit evidence to explain why the site matched
- do not silently create or promote a new known site without review

PostGIS may suggest candidate places, clusters, or spatial context, but it must not override a confident known-site database match. The processed trip table should show what the tracker actually reported, with the match decision explainable through audit evidence.

## CHECK is a feature

`CHECK` means GoGator did not have enough durable evidence to make a confident deterministic decision. It should not be treated as failure.

Prefer `CHECK` over fragile one-off fixes. Good fixes usually improve one of these general concepts:

- site radius
- minimum stationary dwell
- silent stop gap handling
- stationary teleport filtering
- same-site micro-trip suppression
- route observation/anomaly review
- movement signal interpretation
- database-backed site evidence
- learned route confidence

## Route principle

Routes are advisory, learned, and auditable.

Routes should help explain travel patterns. They should not become brittle command-and-control rules that force a trip to fit an expected answer.

A matching route may increase confidence, annotate a trip, or flag an anomaly. It must not silently rewrite departure or destination sites.

## Output stability

Processed CSV columns are high-stability. Avoid adding columns unless explicitly requested. The expanded/audit outputs are better places to preserve richer diagnostic detail.

Output stability means current command names and output files should not vanish just because the engine replaces the trip-construction internals. It does not mean old GPS trip construction remains authoritative.

## Locked intent

These points are fixed project direction:

- Raw GPS evidence is canonical.
- The engine owns trip construction.
- Existing CLI commands and output files should remain available.
- Current command stability must not preserve old trip-construction internals.
- The engine should operate from the database, not depend on CSV files as its core model.
- Site matching is database-first: known sites in the GoGator database are checked before any PostGIS fallback.
- PostGIS is a fallback discovery/context/audit mechanism when no known-site match is found.
- `sites.csv` and `routes.csv` are import/export or pre-population paths only.
- The engine should learn known routes from repeated observed behaviour and frequency.
- Legacy GPS output is not the authority and is not the test oracle.
- Tests should validate engine behaviour directly.
