# Engine Validation Pack (v0.26.19)

This validation pack is an **automated synthetic/local fixture** harness for comparing `legacy` and `shadow` behaviour, plus guarded `engine` readiness checks, without changing production behaviour.

Optional manual commands are included below for **real-data validation** when you have representative GPS CSVs available locally.

## Safety and defaults

- Default trip source remains `legacy`.
- `valhalla.enabled`, `postgis.enabled`, and diagnostics are all disabled by default.
- No external service is required for validation tests.
- Engine mode remains opt-in/experimental.

## Automated local validation (synthetic fixtures)

Run these as part of normal local verification:

```bash
go test ./engine -run 'ValidationPack|ValidationMetrics' -v
go test ./engine -run 'ValidationPack|ValidationMetrics|EngineMode|Shadow|Readiness' -v
```

These tests use deterministic local fixtures only (no real-user CSV fixtures and no Valhalla/PostGIS/H3 runtime requirements).

## Optional manual real-data validation flow

If you have a representative raw GPS CSV locally, run this manual flow before enabling engine mode.

1. Run default legacy processing:

```bash
gogator process <raw-gps.csv>
```

2. Run shadow comparison with diagnostics enabled (official output remains legacy unless you explicitly switch trip source):

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

3. Inspect shadow/diagnostic outputs:
   - shadow summary readiness + reasons
   - legacy valid trip count
   - legacy jitter trip count
   - candidate trip count
   - unmatched legacy valid count
   - unmatched candidate count
   - largest boundary delta
   - low-confidence and gap/noise affected counts
   - route signature/grouping counts (if enabled)
   - mismatch/selection diagnostics from audit output

4. Only then test guarded engine mode:

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
