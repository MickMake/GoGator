#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

BIN="$WORK/gogator"
DATA="$WORK/data"
mkdir -p "$DATA"

log() {
  printf '\n==> %s\n' "$*"
}

fail() {
  printf 'SMOKE TEST FAILED: %s\n' "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if ! printf '%s' "$haystack" | grep -Fq "$needle"; then
    printf 'Output was:\n%s\n' "$haystack" >&2
    fail "$label: expected to find: $needle"
  fi
}

assert_file_contains() {
  local path="$1"
  local needle="$2"
  if ! grep -Fq "$needle" "$path"; then
    printf 'File %s was:\n' "$path" >&2
    cat "$path" >&2
    fail "expected $path to contain: $needle"
  fi
}

log "Build CLI with vendored dependencies"
(
  cd "$ROOT"
  go build -mod=vendor -o "$BIN" ./cmd/gogator
)

log "Show command help"
commands_output="$($BIN commands)"
assert_contains "$commands_output" "import gps [from] <file...>" "commands help"
assert_contains "$commands_output" "export routes" "commands help"

log "Initialise database"
(
  cd "$DATA"
  init_output="$($BIN db init)"
  assert_contains "$init_output" "initialised database" "db init"
)

log "Import GPS CSV and verify idempotent row count"
cat > "$DATA/gps-a.csv" <<'CSV'
dt,lat,lng,altitude,angle,speed,params
2026-05-01 00:00:00,-33.000000,151.000000,10,90,42,io1=1
2026-05-01 00:01:00,-33.100000,151.100000,11,91,43,io1=0
CSV
(
  cd "$DATA"
  import_output="$($BIN import gps from gps-a.csv)"
  assert_contains "$import_output" "imported gps files: 1" "gps import"
  assert_contains "$import_output" "new gps points: 2" "gps import"
  assert_contains "$import_output" "new gps point sources: 2" "gps import"

  repeat_output="$($BIN import gps gps-a.csv)"
  assert_contains "$repeat_output" "new gps points: 0" "gps duplicate import"
  assert_contains "$repeat_output" "new gps point sources: 0" "gps duplicate import"

  status_output="$($BIN db status)"
  assert_contains "$status_output" "gps_points: 2" "db status after gps import"
)

log "Import/export sites with blank TSV fields"
cat > "$DATA/sites.tsv" <<'TSV'
Site	Address	GPS	Range	Min Destination Minutes	Type	Important	Notes
Home		-33.00000000,151.00000000	100	5	Home	yes	
Shop		-33.10000000,151.10000000	150	5	Supplier	yes	
Depot		-33.20000000,151.20000000	150	5	Depot	yes	
TSV
(
  cd "$DATA"
  sites_output="$($BIN import sites from sites.tsv)"
  assert_contains "$sites_output" "imported 3 site(s)" "sites import"
  $BIN export sites as exported-sites.tsv
  assert_file_contains exported-sites.tsv "Home"
  assert_file_contains exported-sites.tsv "Shop"
)

log "Add/delete directional route"
(
  cd "$DATA"
  add_route_output="$($BIN add route from Home to Shop name SupplyRun confidence manual notes smoke)"
  assert_contains "$add_route_output" "upserted route: Home -> Shop" "add route"
  status_output="$($BIN db status)"
  assert_contains "$status_output" "routes: 1" "db status after add route"
  delete_route_output="$($BIN delete route from Home to Shop)"
  assert_contains "$delete_route_output" "deleted route: Home -> Shop" "delete route"
  status_output="$($BIN db status)"
  assert_contains "$status_output" "routes: 0" "db status after delete route"
)

log "Import/export routes and preserve blank optional numbers versus explicit zero"
cat > "$DATA/routes.tsv" <<'TSV'
From	To	Name	Confidence	Notes	Expected Distance Min Km	Expected Distance Max Km	Expected Duration Min Min	Expected Duration Max Min
Home	Shop	BlankRun	manual	blank numbers				
Shop	Depot	ZeroRun	manual	zero numbers	0	0	0	0
TSV
(
  cd "$DATA"
  routes_output="$($BIN import routes from routes.tsv)"
  assert_contains "$routes_output" "imported 2 route(s)" "routes import"
  $BIN export routes as exported-routes.tsv
  awk -F '\t' 'NR == 2 { if ($6 != "" || $7 != "" || $8 != "" || $9 != "") exit 1 }' exported-routes.tsv || fail "blank optional route numbers were not preserved"
  awk -F '\t' 'NR == 3 { if ($6 != "0" || $7 != "0" || $8 != "0" || $9 != "0") exit 1 }' exported-routes.tsv || fail "explicit zero route numbers were not preserved"
)

log "Legacy route promotion path is still wired"
(
  cd "$DATA"
  if "$BIN" add route 1 from missing-route-observations.tsv >legacy.out 2>&1; then
    fail "legacy route promotion unexpectedly succeeded"
  fi
  assert_file_contains legacy.out "missing-route-observations.tsv"
)

log "Smoke tests passed"
