#!/usr/bin/env bash
# Runs every query in queries/ against the captured JSONL files.
#
# Usage:
#   ./run.sh             # run all queries
#   ./run.sh 04          # run only query 04 (matches by prefix)
#   ./run.sh 04 06       # run queries 04 and 06
#
# Expects:
#   - duckdb on PATH (or set DUCKDB env var to the binary path)
#   - run from anywhere; we resolve our own dir
#   - capture files at $REPO_ROOT/captures/*.jsonl
set -euo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$HERE/.." && pwd)"

# Locate the duckdb binary. Preference order:
#   1. $DUCKDB env var (explicit override)
#   2. duckdb on PATH
#   3. ./duckdb in the reporter dir (run install.sh to populate)
#   4. /tmp/duckdb (where we drop it during testing)
DUCKDB="${DUCKDB:-}"
for candidate in "$(command -v duckdb || true)" "$HERE/duckdb" /tmp/duckdb; do
    if [[ -z "$DUCKDB" && -x "$candidate" ]]; then
        DUCKDB="$candidate"
    fi
done

if [[ -z "${DUCKDB:-}" ]]; then
    cat >&2 <<EOF
duckdb not found. Install with one of:

  # download into the reporter dir (no sudo, no PATH change needed):
  $HERE/install.sh

  # system-wide via apt (Ubuntu 24.04+):
  sudo apt install duckdb

  # or download manually from https://duckdb.org/docs/installation/

After install, rerun: $0 $*
EOF
    exit 1
fi

cd "$REPO_ROOT"

# Build the filter list. If no args, run everything.
filters=("$@")
match() {
    local name="$1"
    if [[ ${#filters[@]} -eq 0 ]]; then return 0; fi
    for f in "${filters[@]}"; do
        if [[ "$name" == *"$f"* ]]; then return 0; fi
    done
    return 1
}

for q in "$HERE/queries/"*.sql; do
    name="$(basename "$q" .sql)"
    if ! match "$name"; then continue; fi
    echo
    echo "════════════════════════════════════════════════════════════════"
    echo "  $name"
    echo "════════════════════════════════════════════════════════════════"
    head -n 1 "$q" | sed 's/^-- //'
    echo
    # views.sql is sourced first so every query sees `obs` / `obs_json`,
    # then views_linkage.sql adds `fingerprints` / `fp_*`, then
    # uuid_names.sql adds the `service_uuid_names` lookup. All no-ops if
    # you don't reference them.
    "$DUCKDB" -c ".read $HERE/views.sql" \
              -c ".read $HERE/views_linkage.sql" \
              -c ".read $HERE/uuid_names.sql" \
              -c ".read $q"
done
