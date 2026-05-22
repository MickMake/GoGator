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

assert_file_not_contains() {
  local path="$1"
  local needle="$2"
  if grep -Fq "$needle" "$path"; then
    printf 'File %s was:\n' "$path" >&2
    cat "$path" >&2
    fail "expected $path not to contain: $needle"
  fi
}

log "Build CLI with vendored dependencies"
(
  cd "$ROOT"
  go build -mod=vendor -o "$BIN" ./cmd/gogator
)

log "Show command help"
commands_output="$($BIN commands)"
assert_contains "$commands_output" "import raw [from] <file...>" "commands help"
assert_contains "$commands_output" "export raw [[as] file]" "commands help"
assert_contains "$commands_output" "export gps [[as] file]" "commands help"
assert_contains "$commands_output" "set gps params" "commands help"
assert_contains "$commands_output" "export routes" "commands help"

log "Initialise database"
(
  cd "$DATA"
  init_output="$($BIN db init)"
  assert_contains "$init_output" "initialised database" "db init"
)

log "Check renamed raw import command"
(
  cd "$DATA"
  if "$BIN" import gps from gps-a.csv >import-gps.out 2>&1; then
    fail "import gps unexpectedly succeeded"
  fi
  assert_file_contains import-gps.out "import gps is not a command; use import raw"
)

log "Check renamed raw import command"
(
  cd "$DATA"
  if "$BIN" import gps from gps-a.csv >import-gps.out 2>&1; then
    fail "import gps unexpectedly succeeded"
  fi
  assert_file_contains import-gps.out "import gps is not a command; use import raw"
)

log "Import raw CSV, export raw CSV, export clean flattened GPS, and verify idempotent row count"

cat > "$DATA/raw-a.csv" <<'CSV'
dt,lat,lng,altitude,angle,speed,params
2026-05-01 00:00:00,-33.000000,151.000000,10,90,42,"zeta=9,alpha=1,io1=1"
2026-05-01 00:01:00,-33.100000,151.100000,11,91,43,"alpha=2,io1=0"
CSV
(
  cd "$DATA"
  import_output="$($BIN import raw from raw-a.csv)"
  assert_contains "$import_output" "imported raw files: 1" "raw import"
  assert_contains "$import_output" "new gps points: 2" "raw import"
  assert_contains "$import_output" "new gps point sources: 2" "raw import"

  repeat_output="$($BIN import raw raw-a.csv)"
  assert_contains "$repeat_output" "new gps points: 0" "raw duplicate import"
  assert_contains "$repeat_output" "new gps point sources: 0" "raw duplicate import"
  $BIN export raw as exported-raw.csv
  assert_file_contains exported-raw.csv "dt,lat,lng,altitude,angle,speed,params"
  assert_file_contains exported-raw.csv '2026-05-01 00:00:00,-33,151,10,90,42,"zeta=9,alpha=1,io1=1"'
  assert_file_contains exported-raw.csv '2026-05-01 00:01:00,-33.1,151.1,11,91,43,"alpha=2,io1=0"'

  raw_reimport_output="$($BIN import raw exported-raw.csv)"
  assert_contains "$raw_reimport_output" "new gps points: 0" "raw export re-import"

  $BIN export gps as exported-gps.csv
  assert_file_contains exported-gps.csv $'Raw DT,Normalised Time,Lat,Lng,Altitude,Angle,Speed KPH,gpslev,gsmlev,pdop,io1'
  assert_file_contains exported-gps.csv $'g0,g1,g2,alpha,zeta'
  assert_file_contains exported-gps.csv "2026-05-01 00:00:00"
  assert_file_contains exported-gps.csv $',1,'
  assert_file_contains exported-gps.csv $',19'
  assert_file_contains exported-gps.csv $',2'
  assert_file_not_contains exported-gps.csv "Params Raw"
  assert_file_not_contains exported-gps.csv "Params JSON"
  assert_file_not_contains exported-gps.csv "First Source File"
  assert_file_not_contains exported-gps.csv "Seen Count"

  status_output="$($BIN db status)"
  assert_contains "$status_output" "gps_points: 2" "db status after raw import"
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

log "Route promotion path is still wired under add route"
(
  cd "$DATA"
  if "$BIN" add route 1 from missing-route-observations.tsv >legacy.out 2>&1; then
    fail "route promotion unexpectedly succeeded"
  fi
  assert_file_contains legacy.out "missing-route-observations.tsv"
)

log "Smoke tests passed"
