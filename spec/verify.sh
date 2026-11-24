#!/usr/bin/env bash
# Verify every FSL spec. The proved specs must be proved by induction. The
# naive and gap specs are verified by BMC, and each holds a witness of the
# counterexample. See spec/README.md.
#
# Every verdict is read from the JSON `result`, never from the exit code. fslc
# exits non-zero for a parse error and an internal error too, so a test of the
# exit code alone cannot tell a broken spec from a failing one.
#
# fslc is the FSL verifier: https://github.com/ymm-oss/fsl
set -o pipefail

cd "$(dirname "${BASH_SOURCE[0]}")" || exit 1

if ! command -v fslc > /dev/null 2>&1; then
  # Locally a missing fslc skips the specs. In CI it fails the run: a green
  # job that verified nothing would pass for a verified one.
  if [[ -n "${CI:-}" ]]; then
    echo "fslc not found, and CI is set: the specs cannot be skipped" >&2
    exit 1
  fi
  echo "fslc not found; skipping the specs. Install it with: mise install (mise.toml pins its version)" >&2
  exit 0
fi
if ! command -v python3 > /dev/null 2>&1; then
  echo "python3 not found; it reads fslc's JSON" >&2
  exit 1
fi

## Reads one top-level field of fslc's JSON. A missing field, or output that
## is not JSON, reads as "none".
field() {
  python3 -c '
import json, sys
try:
    v = json.loads(sys.stdin.read()).get(sys.argv[1])
except Exception:
    v = None
print(v if isinstance(v, str) else "none")
' "$2" <<< "$1"
}

## These must verify, and must be inductive rather than true only to a depth.
proving=(nil_value pairing pairing_stores pairing_load local_store result_zero)

## These verify by BMC to a depth. Each holds a reachable witness that is
## the negation of a proved spec's claim. A naive spec models a rule that
## was rejected, and the witness is the reason. A gap spec models the chosen
## rule on a case it accepts as a limitation. They leave out the proved
## spec's auxiliary invariants, so induction cannot prove them.
bounded=(
  nil_value_nocheck
  nil_value_nophi
  pairing_naive
  pairing_nonrecursive
  pairing_bare_return
  pairing_load_whole
  local_store_last
  local_store_escape
  result_zero_nocheck
  result_zero_nokill
  nil_value_correlated
  nil_value_dead_check
  return_store_deferred
)

status=0

## Every spec in the directory must be named above, or nothing would run it.
for f in *.fsl; do
  name=${f%.fsl}
  if [[ " ${proving[*]} ${bounded[*]} " != *" $name "* ]]; then
    echo "  FAILED   $f is named in neither list, so nothing verifies it" >&2
    status=1
  fi
done

for f in "${proving[@]}"; do
  # The depth bounds the search for each reachable's witness. The longest,
  # nil_value's three rounds of the loop, takes six steps.
  out=$(fslc verify "$f.fsl" --engine induction --depth 8 2> /dev/null)
  if [[ "$(field "$out" result)" == "proved" ]]; then
    echo "  ok       $f.fsl (proved)"
  else
    echo "  FAILED   $f.fsl is $(field "$out" result), want proved" >&2
    status=1
  fi
done

for f in "${bounded[@]}"; do
  # Every run of these specs ends within six steps, so depth 8 covers them.
  out=$(fslc verify "$f.fsl" --depth 8 2> /dev/null)
  if [[ "$(field "$out" result)" == "verified" ]]; then
    echo "  ok       $f.fsl (verified to depth 8)"
  else
    echo "  FAILED   $f.fsl is $(field "$out" result), want verified" >&2
    status=1
  fi
done
exit $status
