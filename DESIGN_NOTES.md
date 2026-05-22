# GoGator Design Notes

## Design goal

GoGator should produce deterministic, spreadsheet-friendly trip logs from raw Gator/Teltonika GPS CSV exports.

The goal is not to guess perfectly. The goal is to make repeatable decisions, preserve enough evidence for review, and mark uncertain cases as `CHECK` instead of pretending the goblins have filled in a statutory declaration.

## Canonical data rule

Raw GPS CSV is canonical.

Gator's processed "drives and stops" exports are not canonical because they have shown suspicious timestamp and trip-detection behaviour. They may be used as rough reference material, but GoGator should not rely on them for truth.

## Local enrichment model

GoGator enriches raw tracker observations using local reference files:

- `sites.csv` / TSV: known sites, real addresses, GPS centres, match radii, dwell thresholds.
- `routes.csv`: approved or expected common routes.
- `gogator.yaml`: deterministic processing thresholds and defaults.

The app should never silently create permanent rules from one odd trip. Generate observations/anomalies instead, then let the user approve routes or adjust reference files.

## Preserve evidence

Avoid losing detail from the source data. In particular:

- preserve raw row numbers
- preserve raw timestamps and corrected rendered times where relevant
- preserve stable `params` columns in expanded output
- preserve movement/idling/GPS quality/accelerometer/crash-style fields
- preserve observed/noisy GPS coordinates in processed output
- avoid replacing observed evidence with cleaned canonical reference values

## Site matching principle

When a raw point or stationary cluster matches a known site:

- write the matched site name from `sites.csv`
- write the real address from `sites.csv`
- keep the observed tracker GPS coordinate in the trip output

The canonical site GPS already exists in `sites.csv`. The processed trip table should show what the tracker actually reported.

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

## Route principle

Routes are advisory only.

A matching route may increase confidence, annotate a trip, or flag an anomaly. It must not silently rewrite departure or destination sites.

## Output stability

Processed CSV columns are high-stability. Avoid adding columns unless explicitly requested. The expanded/audit outputs are better places to preserve richer diagnostic detail.
