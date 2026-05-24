# engine

`engine` is GoGator's explicit orchestration seam for in-memory GPS processing.

## Current status (v0.26.18)

Implemented now:
- `engine.Run(ctx, input)` is the compatibility orchestration seam wrapping the existing legacy processing sequence.
- Passive engine diagnostics are implemented for evidence, quality, motion, stays, visits, excursions, candidate trips, trip comparison, shadow summary/mismatches, engine-mode selection, map-match diagnostics, route signatures, and route groups.
- `engine.trip_source` supports `legacy` (default), `shadow`, and experimental `engine`.

Passive/diagnostic-only today:
- Candidate trip building and candidate-vs-legacy comparison.
- Valhalla map-match diagnostics.
- Route signatures and deterministic grouping.
- Shadow readiness summaries and mismatch reports.

Experimental/opt-in only:
- `engine.trip_source: engine` official-output selection path.
- Engine-mode readiness policy gating and optional explicit fallback to legacy.

## Trip source modes

- `legacy` (default): official processed valid/jitter output comes from the legacy trip pipeline.
- `shadow`: official output still comes from legacy; shadow diagnostics are used for comparison only.
- `engine`: official output uses adapted engine candidate trips, guarded by readiness validation (`engine.engine_mode`).

Defaults remain conservative:
- `engine.trip_source: legacy`
- `engine.engine_mode.fallback_to_legacy_on_reject: false` (no silent fallback)

## Valhalla status

- Optional scaffold and client live under `engine/mapmatch`.
- Disabled by default (`valhalla.enabled: false`).
- No HTTP calls occur when disabled.
- When enabled, map-match failures are captured as diagnostics and remain advisory-only by default.

## Route signature / H3 placeholder status

- Route signatures are deterministic and diagnostic-only.
- No external H3 dependency is required.
- No CGO requirement is introduced.
- Grouping uses identical signature keys for passive diagnostics only.

## PostGIS scaffold status

- Optional scaffold under `engine/sitematch`.
- Disabled by default (`postgis.enabled: false`), so no DB connection is attempted by default.
- If explicitly enabled without full implementation, scaffold returns clear validation/not-implemented errors.
- `engine/sitematch/schema.sql` documents intended future tables/indexing and is documentation-only in current releases.

## Recommended safe workflow

1. Run normal default legacy mode (`engine.trip_source: legacy`).
2. Enable diagnostics (`engine.audit.enabled` and specific outputs, or `output_diagnostics`).
3. Inspect shadow summaries/mismatches on known representative data.
4. Test `engine` mode only on known datasets with strict readiness policy.
5. Keep fallback explicit (`fallback_to_legacy_on_reject`) and review selection diagnostics.

## Stability guarantees for this stage

- Default behaviour remains legacy.
- CLI command names/arguments stay unchanged.
- Official processed output schemas/headers stay unchanged.
- Routes remain advisory and do not silently rewrite destinations.
