# Engine Validation Pack (v0.26.19)

This validation pack is for **real-data comparison** between `legacy`, `shadow`, and optional `engine` mode without changing production behaviour.

## Safety and defaults

- Default trip source remains `legacy`.
- `valhalla.enabled`, `postgis.enabled`, and diagnostics are all disabled by default.
- No external service is required for validation tests.
- Engine mode remains opt-in/experimental.

## Recommended run flow

1. Run default legacy processing:

```bash
gogator process <raw-gps.csv>
```

2. Run shadow comparison (official output still legacy):

```yaml
engine:
  trip_source: shadow
  trip_builder:
    enabled: true
    compare_legacy: true
  shadow:
    enabled: true
    summary_enabled: true
  audit:
    enabled: true
    output_diagnostics: true
```

3. Inspect validation metrics in diagnostics/tests:
   - legacy valid trip count
   - legacy jitter trip count
   - candidate trip count
   - shadow readiness
   - unmatched legacy valid count
   - unmatched candidate count
   - largest boundary delta
   - low-confidence and gap/noise affected counts
   - route signature/grouping counts (if enabled)

4. Test engine mode safely only when shadow readiness is acceptable:

```yaml
engine:
  trip_source: engine
  engine_mode:
    require_min_readiness: true
    min_readiness: GoodMatch
    fallback_to_legacy_on_reject: false
```

If rejected, review readiness reasons before any promotion.

## “Good enough to promote” guideline

A practical threshold is:
- readiness `GoodMatch` or better,
- low unmatched valid legacy rate,
- low boundary drift,
- minimal low-confidence/gap/noise candidate warnings.

Treat routes and map-match diagnostics as advisory only.

## Test harness

Use:

```bash
go test ./engine -run ValidationPack
```

The harness validates deterministic comparison metrics and ensures validation works with local fixtures/synthetic points only (no Valhalla/PostGIS/H3 services required).
