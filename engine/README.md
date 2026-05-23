# engine

`engine` introduces an isolated seam around GoGator's existing in-memory GPS processing pipeline.

Current state:
- `engine.Run(ctx, input)` delegates to the same sequence used by `internal/app/process.go`.
- No behaviour changes are intended.
- Existing trip detection, jitter handling, important-site collapsing, and route enrichment remain in `internal/gps` and `internal/routes`.

This package exists so future GPS intelligence work can be implemented behind a stable engine API without changing CLI behaviour.
