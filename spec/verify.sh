#!/usr/bin/env bash
# Verify every FSL spec. The failing ones are expected to fail, and exist to
# hold a counterexample. A naive spec models a rule that was rejected. A gap
# spec models the chosen rule on a case it accepts as a limitation. See
# spec/README.md.
#
# Every verdict is read from the JSON `result`, never from the exit code. fslc
# exits non-zero for a parse error and an internal error as well as for a
# violation, so an exit-code test would let a failing spec rot into
# unparseable text and still call it "violated, as intended".
#
# Each failing spec is also pinned to the invariant it must break. A spec that
# fails for some new, unrelated reason has stopped holding its counterexample.
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
  echo "fslc not found; skipping the specs. Install: https://github.com/ymm-oss/fsl" >&2
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

## These must NOT verify. Each is a rule that was rejected, and its
## counterexample is the reason. Pinned to the invariant it must break.
naive_specs=(
  nil_value_nocheck        ReportMeansANilIsReturned
  nil_value_nophi          CheckedOrConstantNilIsReported
  pairing_naive            ReportMeansNilWithBad
  pairing_nonrecursive     ReportMeansNilWithBad
  pairing_bare_return      ReportMeansNilWithBad
  pairing_load_whole       ReportMeansNilWithBad
  local_store_last         ConstantNilIsReported
  local_store_escape       ReportMeansANilIsLoaded
  result_zero_nocheck      ReportMeansAZeroIsUsed
  result_zero_nokill       ReportMeansAZeroIsUsed
)

## These must NOT verify either. Each is the chosen rule on a case it accepts
## as a gap: it reports, and no run gives the value. Pinned the same way.
gap_specs=(
  nil_value_correlated     ReportMeansANilIsReturned
  nil_value_dead_check     ReportMeansANilIsReturned
  return_store_deferred    ReportMeansANilIsReturned
)

status=0

failing=()
for ((i = 0; i < ${#naive_specs[@]}; i += 2)); do failing+=("${naive_specs[$i]}"); done
for ((i = 0; i < ${#gap_specs[@]}; i += 2)); do failing+=("${gap_specs[$i]}"); done

## Every spec in the directory must be named above, or nothing would run it.
for f in *.fsl; do
  name=${f%.fsl}
  if [[ " ${proving[*]} ${failing[*]} " != *" $name "* ]]; then
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
## Checks that a spec fails on its pinned invariant. $3 is naive or gap.
check_failing() {
  local f=$1 want=$2 kind=$3 out
  out=$(fslc verify "$f.fsl" --depth 8 2> /dev/null)
  case "$(field "$out" result)" in
    violated)
      if [[ "$(field "$out" invariant)" == "$want" ]]; then
        echo "  ok       $f.fsl ($kind: violated on $want, as intended)"
      else
        echo "  FAILED   $f.fsl ($kind) broke on $(field "$out" invariant), want $want" >&2
        status=1
      fi
      ;;
    verified | proved)
      echo "  FAILED   $f.fsl ($kind) verified, but it must hold its counterexample" >&2
      status=1
      ;;
    *)
      echo "  FAILED   $f.fsl ($kind) did not run: $(field "$out" result)" >&2
      status=1
      ;;
  esac
}

for ((i = 0; i < ${#naive_specs[@]}; i += 2)); do
  check_failing "${naive_specs[$i]}" "${naive_specs[$((i + 1))]}" naive
done
for ((i = 0; i < ${#gap_specs[@]}; i += 2)); do
  check_failing "${gap_specs[$i]}" "${gap_specs[$((i + 1))]}" gap
done
exit $status
