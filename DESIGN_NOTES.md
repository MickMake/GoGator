<!--
PROTECTED FILE - DO NOT EDIT

This file is intentionally protected.
Do not modify, rewrite, reformat, rename, move, or delete this file.

Codex/AI agents:
- You must not change this file.
- If a task appears to require changing this file, stop and report that it is protected.
- Do not "clean up", "modernise", "simplify", or "deduplicate" this file.

PROTECTED FILE - DO NOT EDIT WITHOUT EXPLICIT OWNER APPROVAL

This file defines the non-negotiable goal and mechanics of GoGator.

Codex/AI agents:
- Read this file before every change.
- Do not rewrite, summarise, soften, reinterpret, or "modernise" this file.
- Do not make changes that conflict with this file.
- If a requested task appears to conflict with this file, stop and report the conflict.
-->

# GoGator Project Intent

## Purpose

GoGator turns raw Gator/Teltonika GPS tracker evidence into deterministic, auditable, spreadsheet-friendly trip records for a trade vehicle.

The core output is:

```text
departure site -> travel/route -> destination site -> time spent at destination
```

GoGator must explain how it reached each answer.

A trip row that cannot explain itself is just a spreadsheet wearing a false moustache.

---

## Non-negotiable direction

These points are fixed project intent.

1. Raw GPS evidence is canonical.
2. Gator processed "drives and stops" exports are not canonical truth.
3. The `engine` package owns trip construction.
4. Official trip output must come from the engine path.
5. Existing CLI commands and output files should remain available unless explicitly changed.
6. Command stability must not preserve old trip-construction internals.
7. Runtime trip construction must use the GoGator database as the operating model.
8. CSV/TSV files are import/export/pre-population paths only.
9. Known sites in the GoGator database are the first authority for site matching.
10. PostGIS is a fallback discovery/context/audit mechanism when no known-site match is found.
11. Routes are learned from repeated observed behaviour and route evidence, not primarily hand-authored.
12. `routes.csv` is a pre-population/import/export path only.
13. Valhalla is the preferred map-matching/routing engine.
14. H3 is used for spatial bucketing/signatures where useful.
15. Every output decision must be auditable.
16. Low-confidence output must be shown as `CHECK` or reviewable evidence, not hidden.
17. Tests must validate engine behaviour directly.
18. Legacy GPS output is not the test oracle.
19. Every Codex step must move toward this goal or explicitly say why it cannot.

---

## Source of truth hierarchy

### Raw GPS observations

Raw GPS tracker observations are the canonical source evidence.

Raw data may be noisy, sparse, delayed, duplicated, stale, contradictory, or flat-out odd. The engine must preserve the evidence and attach interpretation beside it. It must not silently rewrite history to make output look cleaner.

Preserve:

- source identity
- source file
- raw row number
- raw timestamp
- normalised timestamp
- raw latitude/longitude
- speed, angle, and altitude where present
- raw params
- decoded tracker signals
- quality flags
- audit/reason codes

### Database operating model

Once raw GPS evidence is loaded, the engine should operate from GoGator's database-backed representation of that evidence.

The database is the operating source for:

- loaded GPS observations
- known sites
- site metadata
- site match radii
- site dwell thresholds
- site audit history
- learned route signatures
- route frequencies
- route observations
- anomalies
- processing runs
- diagnostic evidence

### Flat files

Flat files are not the long-term runtime model.

They exist for:

- import
- export
- review
- bootstrap/pre-population
- portability

Examples:

- `sites.csv` / TSV may pre-populate or export known sites.
- `routes.csv` may pre-populate or export known/approved route records.
- `gogator.yaml` may define deterministic thresholds and runtime defaults.

The engine must not depend on CSV files as its core processing model.

---

## User-facing contract

Current user-facing commands and output files should not vanish simply because the engine internals are improved.

The project may keep command names, output filenames, and CSV headers stable while replacing the underlying trip-construction method.

Important distinction:

```text
Keep the command/output contract.
Do not preserve old trip-construction internals as the authority.
```

Processed CSV columns are high-stability. Expanded/audit/diagnostic outputs are the better place for new evidence.

---

## Engine ownership

Trip construction belongs in `engine`.

The final architecture is:

```text
input/load layer -> database -> engine -> output writers
```

Output writers may continue producing current files. The engine decides trip, stay, site, route, anomaly, and confidence evidence.

The old GPS trip-construction output is not authoritative and must not determine project direction.

---

## Target engine pipeline

The engine pipeline is:

1. ingest and normalise raw evidence
2. clean/derive point evidence
3. quality scoring
4. motion classification
5. stay detection
6. database known-site matching
7. PostGIS fallback matching when no known site is found
8. visit modelling
9. candidate trip construction
10. Valhalla map matching
11. route signature generation
12. route learning/grouping
13. anomaly detection
14. itinerary row generation
15. audit/diagnostic output

Core design choice:

```text
Detect stops from raw clustered behaviour.
Reconstruct travel from map-matched movement.
```

Do not map-match a car park into a road and then wonder why the vehicle appears to have spent six hours politely blocking a lane.

---

## Core model

The exact database schema may evolve, but the engine concepts must remain stable.

### Raw GPS observations

A raw GPS record must preserve enough data to replay and audit engine decisions:

- source identity
- tracker/vehicle identity where available
- raw timestamp
- normalised UTC timestamp
- rendered/local timestamp where useful
- latitude/longitude
- altitude
- heading/angle
- reported speed
- raw params
- decoded tracker signals
- source file and row
- ingestion timestamp

Raw rows are immutable evidence.

### Derived point evidence

Derived point evidence may include:

- delta time from previous usable point
- delta distance
- implied speed
- derived heading
- quality band
- quality flags
- usable-for-stay flag
- usable-for-route flag
- nearest known-site distance
- route/map-match hints
- algorithm version
- config version

Suspicious data should be flagged, not casually deleted.

### Known sites

Known sites are database records. A known site may include:

- site ID
- name
- address
- type/category
- latitude/longitude
- optional polygon
- match radius
- arrival/departure hysteresis radius
- minimum dwell threshold
- priority/importance
- active date range
- notes
- audit history

Known sites are user-approved or imported local knowledge.

### Stays

A stay is a spatial-temporal cluster of tracker evidence.

A stay should preserve:

- start/end time
- duration
- representative coordinate
- cluster radius
- supporting point count
- raw point range
- gap before/after
- classification
- matched site if any
- site match score/confidence
- reasons/flags
- audit evidence

### Trip legs

A trip leg is movement between stays/sites.

A trip leg should preserve:

- origin stay/site
- destination stay/site
- departure/arrival time
- elapsed time
- raw distance
- map-matched distance where available
- moving/idle/gap time
- matched route geometry where available
- route signature
- route group
- map-match score
- anomaly flags
- evidence/reason codes

### Itinerary rows

The itinerary row is the presentation layer.

It should answer:

```text
Where did the vehicle leave from?
When did it leave?
How did it travel?
Where did it arrive?
When did it arrive?
How long did it spend there?
What confidence/review status applies?
```

---

## Quality scoring

The engine must identify and annotate weak or suspicious evidence.

Quality flags should cover:

- duplicate timestamp
- duplicate coordinate
- same timestamp conflict
- impossible speed
- teleport jump
- teleport return
- stale coordinate
- GPS quality weakness when available
- long gap
- cold-start drift
- map-match failure
- missing/contradictory tracker signals

Quality scoring must separate:

```text
usable for stay detection
usable for route reconstruction
```

A point may be useful for one and not the other.

---

## Motion classification

Motion classification should use tracker signals as evidence, not as absolute truth.

Inputs may include:

- reported speed
- derived speed
- distance moved
- time gap
- movement signal
- idling/stationary-style signal
- odometer changes
- accelerometer values when trusted
- GPS quality
- recent state history

Output states should include:

- moving
- stationary
- unknown
- gap
- noise

Use hysteresis. Do not flip states because one lonely point sneezed.

---

## Stay detection

Stay detection is central.

A stay is not just "a point near a site". It is a supported cluster of evidence over time.

Stay detection should consider:

- spatial radius
- dwell duration
- point count
- stationary/motion evidence
- inside-radius ratio
- stationary ratio
- sample gaps
- tracker silence
- known-site context
- repeated behaviour
- confidence and ambiguity

Stays become the main boundary evidence for trip construction.

---

## Database known-site matching

Site matching is database-first.

### Match order

The required order is:

1. known sites in the GoGator database
2. PostGIS spatial/context fallback only when no known-site match is found
3. `CHECK` when there is still not enough durable evidence

PostGIS must not override a confident known-site database match.

### Known-site candidate eligibility

A database known-site candidate is eligible when:

- the site is active for the relevant timestamp
- the site has valid coordinates or geometry
- the observed point/stay/cluster is within the effective match area
- the effective match area uses the site-specific radius/polygon when present
- configured defaults are used only when site-specific values are absent
- dwell evidence satisfies the site threshold when the match represents a visit/destination

### Evidence used for matching

Database known-site matching should use:

- representative stay coordinate
- arrival/departure cluster points
- distance to site centre or polygon
- point count inside match area
- dwell duration
- inside-radius ratio
- stationary ratio
- site-specific dwell threshold
- site-specific radius
- site priority/importance
- recent/historical visit behaviour
- ambiguity margin against other candidate sites

### Tie-breaking

When multiple known sites match, the engine should choose deterministically.

Suggested order:

1. strongest confidence score
2. explicit polygon match over radius-only match
3. longest qualifying dwell evidence
4. highest inside-area ratio
5. highest stationary ratio
6. smallest distance to site centre/geometry
7. higher site priority/importance
8. stable site ID ordering as final tie-breaker

Ambiguous matches should become `CHECK`/review rather than guesses.

### Required audit evidence

A site match should be able to explain itself.

Audit evidence should include:

- match source: `database_known_site`
- matched site ID
- matched site name
- matched site type
- distance metres
- effective radius/metres or polygon result
- dwell duration
- required dwell duration
- inside ratio
- stationary ratio
- confidence
- ambiguity margin
- reason codes
- source point range
- algorithm version
- config version

### New site promotion

The engine must not silently create or promote a new known site.

Unknown repeated stops may produce candidates for review. A human or explicit command should promote them.

---

## PostGIS fallback

PostGIS is a fallback discovery/context/audit mechanism when no known-site database match is found.

It may support:

- spatial proximity search
- polygon/geometry matching
- clustering repeated unknown stops
- candidate unknown-site discovery
- richer audit persistence
- spatial context analysis

PostGIS fallback may suggest:

- possible places
- possible clusters
- possible unknown site candidates
- spatial context

It must not silently override a confident database known-site match.

If PostGIS cannot identify a durable candidate, the engine should emit `CHECK` or an unknown stop with audit evidence.

---

## Known-route learning

Routes are not primarily hand-authored rules.

A route becomes known because repeated observed evidence shows that it is known.

Route learning should use:

- repeated origin/destination pairs
- matched road-segment sequences
- Valhalla matched shapes/attributes
- H3 route cell sequences
- simplified polyline signatures
- duration bands
- distance bands
- frequency
- anomaly history
- time-of-day/context where useful

`routes.csv` is only a pre-population/import/export mechanism. It must not limit the engine to manually curated routes.

A learned route should be auditable and versioned.

---

## Route principle

Routes are advisory, learned, and auditable.

A matching route may:

- increase confidence
- annotate a trip
- identify a usual route
- identify an unusual detour
- flag duration/distance anomaly
- support route grouping

A route must not silently rewrite departure or destination site truth.

Site truth comes from site/stay evidence, not route expectation.

---

## Valhalla map matching

Valhalla is the preferred self-hosted map-matching/routing engine.

Use it to reconstruct movement, not to decide car-park stays.

Valhalla/map matching should support:

- trace/map matching from movement points
- route geometry
- matched distance
- road segment sequence or useful equivalent
- route attributes where available
- confidence/failure status
- route signatures/grouping

Failure policy:

- Valhalla must be optional for normal local tests unless explicitly configured.
- If disabled, the engine should use deterministic fallback/no-op behaviour.
- If enabled and unavailable, startup/config validation may fail early.
- Per-trip map-match failure should become audit evidence, not silent corruption.

---

## H3 / spatial signatures

H3 is useful for:

- route signatures
- stay signatures
- repeated unknown-stop clustering
- heatmaps
- spatial bucketing
- route similarity
- "where does this vehicle tend to stop?" analysis

H3 should be wrapped behind an engine abstraction so the rest of the engine does not depend directly on H3 internals everywhere.

When H3 is disabled or unavailable, deterministic fallback signatures are acceptable, but must be documented.

---

## Trip construction

Trip construction should be stay-to-stay.

A trip starts when the vehicle leaves a confirmed stay/site.

A trip ends when the next confirmed stay/site begins.

Pauses remain inside a trip unless site evidence says otherwise.

Store/supplier pickup stops may either become their own trip legs or be attached to the parent destination visit, but the choice must be explicit and auditable.

Large gaps may create inferred boundaries only when evidence supports it.

Short movements, same-site loops, noise clusters, and low-confidence candidates should be preserved as review/jitter/check evidence rather than promoted blindly.

---

## Anomaly detection

The engine should flag anomalies instead of forcing output to look clean.

Anomalies may include:

- impossible jump
- poor GPS evidence
- unexplained stop
- unusual route
- unusual duration
- unusual distance
- endpoint mismatch
- known-site ambiguity
- long data gap
- map-match failure
- route outlier
- repeated unknown stop

Anomalies should be reviewable and useful for improving configuration or known-site/route data.

---

## Audit requirements

Every important decision must be explainable.

Audit output should preserve:

- algorithm version
- config version
- source data version
- site data version
- route data version
- map data version where available
- raw point IDs/ranges
- quality flags
- motion reasons
- stay reasons
- site match reasons
- route match reasons
- confidence components
- review status
- CHECK reasons

Do not only store a final score. Store the components.

---

## Confidence and review

Confidence should be composed from evidence.

Useful confidence components include:

- point quality
- motion confidence
- stay confidence
- site match confidence
- route/map-match confidence
- ambiguity margin
- gap penalty
- anomaly penalty

Low-confidence rows must remain visible.

Use statuses such as:

- auto
- check
- needs_review
- manually_corrected
- rejected/noise

`CHECK` is a feature. It means the engine is honest.

---

## Testing principle

Tests must validate engine behaviour directly.

Do not test engine correctness by asking whether it reproduces old GPS output.

Good tests ask:

```text
Given this input data and this configuration, is the engine output explainable, auditable, deterministic, and consistent with the intended mechanics?
```

Tests should include:

- synthetic known-good cases
- noisy GPS
- duplicate timestamps
- impossible jumps
- same-site micro trips
- long gaps
- known-site matches
- ambiguous known-site matches
- unknown stops
- repeated unknown-stop clustering
- route learning/grouping
- map-match success/failure
- PostGIS disabled path
- PostGIS fallback path when configured
- CLI/output contract stability

Regression fixtures should compare semantic output, not fragile formatting.

---

## Codex alignment rules

Every Codex run must follow these rules.

### Read first

Before changing code, read:

- `PROJECT_INTENT.md`
- `DESIGN_NOTES.md`
- `engine/README.md`
- `AGENTS.md`
- `CODEX.md`
- `COMMANDS.md`
- relevant engine files
- relevant store/database files
- relevant output files

### Do not drift

Codex must not:

- make legacy GPS output authoritative
- use old GPS output as the oracle
- remove current CLI commands without explicit instruction
- silently change processed CSV columns
- make CSV files the runtime model
- make PostGIS override known-site database matches
- treat `routes.csv` as the route authority
- treat routes as site-truth rewrites
- add one-off rules where an evidence/scoring model is needed
- hide low-confidence data
- remove auditability
- make external services mandatory for ordinary tests unless explicitly requested

### Required final report

Each Codex run must report:

```text
Goal:
Files changed:
Behaviour changes:
Output/schema changes:
Tests run:
Pass/fail:
Known limitations:
Suggested next step:
```

---

## Protected-file policy

This file should be protected.

Suggested protected files:

- `PROJECT_INTENT.md`
- `DESIGN_NOTES.md`
- `engine/README.md`
- `AGENTS.md`
- `CODEX.md`

Suggested append-only file:

- `CHANGES.md`

Protected means:

- do not edit
- do not reformat
- do not summarise
- do not rename
- do not delete
- do not "clean up"

Changes require explicit owner approval.

---

## Acceptance definition

The engine is doing its job when:

1. raw GPS can be loaded into the database
2. known sites are loaded/stored in the database
3. engine trip construction runs from database-backed evidence
4. current commands still produce expected output files
5. official trip output comes from the engine path
6. site matching checks database known sites first
7. PostGIS fallback runs only when no known-site match exists
8. stays are generated from spatial/dwell/context evidence
9. trips are generated from stay-to-stay movement
10. Valhalla route intelligence is optional but supported
11. H3/spatial signatures support route/stay grouping
12. routes can be learned from repeated observed behaviour
13. route matching is advisory and does not rewrite site truth
14. anomalies and CHECK rows are visible
15. audit output explains decisions
16. tests validate engine semantics directly
17. old GPS output is not the oracle

---

## If in doubt

When uncertain, choose the option that best preserves:

1. raw evidence
2. database-first engine operation
3. site/stay/route auditability
4. deterministic replay
5. current command availability
6. engine ownership of trip construction

Do not choose the option that merely keeps old GPS output comfortable.

Legacy is not the captain. At best, it is a map found in the glovebox with coffee stains and suspicious handwriting.
