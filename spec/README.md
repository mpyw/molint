# Formal specs

Machine-checked statements about the path reasoning in [design/rules.md](../design/rules.md), written in [FSL](https://github.com/ymm-oss/fsl) and verified with `fslc`. `./verify.sh` re-runs them.

They check the **design**, not the implementation. Each spec models a rule against what happens at run time:

1. `init` picks the program from free choice bits: which operands each edge brings, where the checks are, what escapes. Every program is a start, and it never changes.
2. The rule is a pure `def` over that program. A `judge` action records its verdict before the run starts.
3. The run then takes any branches, and each value from outside is nil or not, freely.
4. The invariants relate the verdict to what the run can do.

"Reported" means **some run** gives the value, with branches taken as independent, as design/rules.md defines a nil value. Each spec proves this with a witness run: the rule's verdict picks the edges and the outside values, and the invariant says that a run making those choices is never blocked and gives the value. The converse says that every run that gives a constant's nil (or zero) is reported.

Confirming that the Go code follows the rules takes the fixtures under `testdata/src`.

## What is proved

All five are `proved` under `--engine induction`.

| Spec | Rule | Claim |
| --- | --- | --- |
| `nil_value.fsl` | Nil values: the constant, φ-nodes, nil checks on the way | Reported exactly when some run returns a nil that is a constant's or passed a nil check. A loop can take two rounds to bring the nil. A check on the way stops it at its own edge |
| `pairing.fsl` | `return-nil` with a trailing `error` or `bool`, through φ-nodes | Reported exactly when some run returns the nil constant with a nil error (or `true`). The pairing continues into an earlier block when both values on an edge are its φ-nodes |
| `pairing_stores.fsl` | `return-nil` for a bare return in a function with a defer | Reported exactly when some run returns the nil constant with a nil error (or `true`). Both loads are read along the same path back to their stores, as `pairStores` does. In a loop, the two stores may come from different rounds |
| `local_store.fsl` | Loads of a local that is not lifted | Reported exactly when some run loads the nil constant, for a variable that is followed. The Alloc counts as a store of nil. A literal that only reads the variable keeps it followed |
| `result_zero.fsl` | `result-zero` | Reported exactly when some run uses a zero Result that is a constant's or passed a comparison. A comparison with the zero constant on the way stops a zero. A comparison is not a use. An overwrite before the use kills the zero |

Each also witnesses what it gives up: a nil (or zero) from outside that no check tests is not reported. That is nilaway's job.

## What fails, and why

These are not regressions. Each fails on the named invariant, and `verify.sh` pins it. There are two kinds.

### Naive: a rejected rule

The counterexample is the reason the rule was rejected.

| Spec | Rule it models | Counterexample | Breaks |
| --- | --- | --- | --- |
| `nil_value_nocheck.fsl` | A nil check on the way does not stop a nil | `if p == nil { p = def }; return p` is reported | `ReportMeansANilIsReturned` |
| `nil_value_nophi.fsl` | A φ-node is never a nil value | `var p *T; if c { p = x }; return p` is missed | `ConstantNilIsReported` |
| `pairing_naive.fsl` | The pointer and the error are judged apart | `if b { u = x } else { err = e }; return u, err` is reported | `ReportMeansNilWithBad` |
| `pairing_nonrecursive.fsl` | φ-nodes of an earlier block are judged whole | The same shape nested inside another `if`, with `log()` after it, is reported | `ReportMeansNilWithBad` |
| `pairing_bare_return.fsl` | Each load of a bare return is judged on its own reaching stores | `defer mu.Unlock(); if c { u = x } else { err = e }; return` is reported | `ReportMeansNilWithBad` |
| `local_store_last.fsl` | A load reads the last store in program order | `if a { p = nil } else { p = &T{} }; return` is missed | `ConstantNilIsReported` |
| `local_store_escape.fsl` | A variable is followed whatever holds its address | `p = nil; defer func() { p = def }(); return` is reported | `ReportMeansANilIsLoaded` |
| `result_zero_nocheck.fsl` | A comparison on the way does not stop a zero | `if r == (mo.Result[T]{}) { r = mo.Err(e) }; return r` is reported | `ReportMeansAZeroIsUsed` |
| `result_zero_nokill.fsl` | A zero ever stored counts at the use | `var r mo.Result[T]; r = mo.Ok(v); return r` is reported | `ReportMeansAZeroIsUsed` |

### Gap: a limitation the chosen rule accepts

The rule as design/rules.md writes it reports these, although no run gives the value. They are listed as gaps or consequences there. The proved specs leave them out by an `ASSUME` that names the gap.

| Spec | Counterexample | Why it is accepted | Breaks |
| --- | --- | --- | --- |
| `nil_value_correlated.fsl` | `var p *T; if c { p = &T{} }; if c { return p }` is reported. The two branches test the same condition | A nil value is nil on some path, with branches taken as independent | `ReportMeansANilIsReturned` |
| `nil_value_dead_check.fsl` | `p := &T{}; if p == nil { return p }` is reported. The branch is dead | Only a nil check of the value itself narrows the paths; whether it can succeed is not read | `ReportMeansANilIsReturned` |
| `return_store_deferred.fsl` | `defer func() { if p == nil { p = def } }(); return nil` is reported. F returns `def` | A return with operands is judged on what it stores, and this one says nil | `ReportMeansANilIsReturned` |

One more gap has no spec. A nil check on a loaded variable before a bare return does not stop a nil: stores are followed, checks on loaded values are not. So `defer f(); if c { p = x }; if p == nil { p = def }; return` is reported. `local_store.fsl` leaves it out by its ASSUME-3.

## Running them

```bash
./verify.sh
```

The script reads each verdict from `fslc`'s JSON, never from its exit code. The proved specs must be `proved` under induction. The naive and gap specs must be `violated` on the named invariant, and the output says which kind each is. A spec in the directory that is named in no list fails the run.

## Mutation

`fslc mutate <spec> --depth 8`, measured on fslc 4.0.0:

| Spec | Mutants | Killed | Kill rate |
| --- | --- | --- | --- |
| `nil_value.fsl` | 173 | 140 | 81% |
| `pairing.fsl` | 95 | 75 | 79% |
| `pairing_stores.fsl` | PS_TOTAL | PS_KILLED | PS_RATE |
| `local_store.fsl` | 141 | 123 | 87% |
| `result_zero.fsl` | 111 | 75 | 68% |

Most survivors change which free choice bit picks an operand, or the bounds of the bit index. Both leave the set of programs the same. The rest touch state the run writes before it reads, the `abort` action of a run whose checks all fail, or a `follow` update the verdict makes redundant.
