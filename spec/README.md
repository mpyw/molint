# Formal specs

Machine-checked statements about the path reasoning in [design/rules.md](../design/rules.md), written in [FSL](https://github.com/ymm-oss/fsl) and verified with `fslc`. `./verify.sh` re-runs them.

They check the **design**, not the implementation. Each spec models a rule against what happens at run time:

1. `init` picks the program from free choice bits: which operands each edge brings, where the checks are, what escapes. Every program is a start, and it never changes.
2. The rule is a pure `def` over that program. Its verdict, and the witness below, are fixed before the run starts (in `init`, or by a `judge` action).
3. The run then takes any branches, and each value from outside is nil or not, freely. The run keeps its own record of each value, such as "holds the nil constant" or "passed a nil check". The rule never reads that record.
4. The invariants relate the verdict to the run's record.

"Reported" means **some run** gives the value, with branches taken as independent, as design/rules.md defines a nil value. Each spec proves this with a witness run: the rule's verdict picks the edges, the rounds and the outside values. The invariant says that a run making those choices gives the value, and is never blocked: every check, round and exit it takes is enabled when it gets there. The converse says that every run that gives a constant's nil (or zero), or one that passed a check, is reported.

Confirming that the Go code follows the rules takes the fixtures under `testdata/src`.

## What is proved

All six are `proved` under `--engine induction`.

| Spec | Rule | Claim |
| --- | --- | --- |
| `nil_value.fsl` | Nil values: the constant, φ-nodes in a loop and a join, nil checks on join edges and on loop edges | Reported exactly when some run returns a nil that is a constant's or passed a nil check. A nil can take three rounds of the loop to arrive. A check on the way stops it at its own edge |
| `pairing.fsl` | `return-nil` with a trailing `error` or `bool`, through φ-nodes | Reported exactly when some run returns such a nil pointer with a nil error (or `true`). The pairing continues into the loop's φ-nodes, and a pair may come from two rounds. A check `u != nil` on an edge settles u before its φ-nodes are taken apart; `u == nil` makes it nil |
| `pairing_stores.fsl` | `return-nil` for a bare return in a function with a defer | Reported exactly when some run returns the nil constant with a nil error (or `true`). Both loads are read along the same path back to their stores. Stored values that are φ-nodes of one block are paired along its edges. A store after the loop kills the ones before it |
| `pairing_load.fsl` | `return-nil` pairing a φ-node with a load of a followed variable | Reported exactly when some run returns the pair. The load is read, along each edge, as the stores reaching the end of that edge's predecessor |
| `local_store.fsl` | Loads of a local that is not lifted | Reported exactly when some run loads the nil constant, for a variable no code can store into through its address. The Alloc counts as a store of nil. A literal that only loads the variable keeps it followed: the converse is stated over the run, so a rule that drops it breaks |
| `result_zero.fsl` | `result-zero`, for a variable SSA lifts | Reported exactly when some run passes on a zero Result that is a constant's or passed a comparison. A comparison with the zero constant stops a zero at its own edge. An overwrite before the use kills it. A comparison is not a use: the run records which instructions pass the Result on, and the claim is stated over that record |

Each also witnesses what it gives up: a nil (or zero) from outside that no check tests is not reported. That is nilaway's job.

### Rule mutations each spec catches

The specs were checked by editing the rule by hand and re-running. Each edit below breaks a proved invariant under `--engine induction`, and one of the named claims under BMC depth 8 with the auxiliary invariants and the `reachable` witnesses left out.

| Spec | Edit to the rule | Claim that breaks under BMC |
| --- | --- | --- |
| `nil_value.fsl` | A check on a join edge ignored | `ReportMeansANilIsReturned` |
| | An `IsNil` join edge not counted | `CheckedOrConstantNilIsReported` |
| | φ-nodes not followed, or followed one level only | `CheckedOrConstantNilIsReported` |
| | An `IsNil` loop edge not counted | `CheckedOrConstantNilIsReported` |
| | A check on a loop edge ignored when chaining φ-nodes | `ReportMeansANilIsReturned` |
| | A value from outside counted as nil | `ReportMeansANilIsReturned` |
| | The entry's nil not counted | `CheckedOrConstantNilIsReported` |
| `pairing.fsl` | Values from the loop judged whole (no recursion) | `ReportMeansNilWithBad` |
| | `u != nil` on an edge ignored | `ReportMeansNilWithBad` |
| | `u == nil` on a φ-node dropped | `NilWithBadIsReported` |
| | The first store of the loop kept, not the last | `ReportMeansNilWithBad` |
| | The loop not followed, or followed one round only | `NilWithBadIsReported` |
| | A value from outside counted as nil or Bad | `ReportMeansNilWithBad` |
| `pairing_stores.fsl` | `f()` counted as nil; `true` for every error; the bool's zero counted as `true` | `ReportMeansNilWithBad` |
| | The Alloc's zero forgotten; one round of the loop only | `NilWithBadIsReported` |
| | An arm not needed without the loop | `WitnessNeverBlocked` |
| | The edges of P judged apart; the store after the loop ignored | `ReportMeansNilWithBad` |
| | Each load judged apart | `WitnessNeverBlocked` |
| `pairing_load.fsl` | The load judged whole; an arm's store not killing | `ReportMeansNilWithBad` |
| | The Alloc's zero forgotten; one store of the earlier if only | `NilWithBadIsReported` |
| `local_store.fsl` | `followed()` false; a load-only literal not followed | `ConstantNilIsReported` |
| | An escaped variable followed; a later store not killing; `f()` counted as nil | `ReportMeansANilIsLoaded` |
| | The Alloc's zero forgotten; the loop not allowed zero rounds; arms always killing the entry | `ConstantNilIsReported` |
| `result_zero.fsl` | A comparison not stopping; an overwrite not killing; `f()` counted as zero; a comparison counted as a use | `ReportMeansAZeroIsUsed` |
| | `EqZero` not counted; the earlier φ-node not followed, or one edge of it only | `ZeroUseIsReported` |

One edit survives, and it is equivalent: in `nil_value.fsl`, letting a `NotNil` loop edge bring the constant nil. Such a check can never succeed, and ASSUME-2 leaves it out.

## What fails, and why

These are not regressions. Each fails on the named invariant, and `verify.sh` pins it. There are two kinds.

### Naive: a rejected rule

The counterexample is the reason the rule was rejected. A naive spec that over-reports runs only programs the proved rule does not report, so the report is the naive rule's own. Its checks can all succeed and fail, so the edge the witness is blocked on is a live one.

| Spec | Rule it models | Counterexample fslc finds | Breaks |
| --- | --- | --- | --- |
| `nil_value_nocheck.fsl` | A nil check on the way does not stop a nil | `a, b := f(), nil; for cond() { a, b = b, a }; if a != nil { return a }` is reported | `ReportMeansANilIsReturned` |
| `nil_value_nophi.fsl` | A φ-node is never a nil value | `a := (*T)(nil); for cond() { if x := f(); x != nil { a = x } }; return a` is missed: zero rounds return nil | `CheckedOrConstantNilIsReported` |
| `pairing_naive.fsl` | The pointer and the error are judged apart | `u, err := (*T)(nil), error(nil); for next() { u, err = f(), g() }; switch { case a: return u, errX; case u != nil: return u, err }` is reported | `ReportMeansNilWithBad` |
| `pairing_nonrecursive.fsl` | φ-nodes of an earlier block are judged whole | `u, err := &T{}, nil; for next() { if b { u, err = nil, g() } else { u, err = h(), g() } }; if u == nil { return u, err }` is reported | `ReportMeansNilWithBad` |
| `pairing_bare_return.fsl` | Each load of a bare return is judged on its own reaching stores | With a defer, `u = f(); for next() { if b { u, err = nil, g() } else { u, err = nil, errX } }; return` is reported | `ReportMeansNilWithBad` |
| `pairing_load_whole.fsl` | A load paired with a φ-node is judged whole | With err captured by a load-only defer, `if a { err = g() }; switch { case c0: u, err = nil, errX; case c1: u = f(); default: u, err = f(), nil }; return u, err` is reported | `ReportMeansNilWithBad` |
| `local_store_last.fsl` | A load reads the last store in program order | `var p *T; for cond() { switch { case a: p = nil; case b: p = f(); default: p = &T{} } }; return` is missed: zero rounds load nil | `ConstantNilIsReported` |
| `local_store_escape.fsl` | A variable is followed whatever holds its address | `var p *T; for cond() { p = &T{} }; set(&p); return` is reported, and `set` stores into p | `ReportMeansANilIsLoaded` |
| `result_zero_nocheck.fsl` | A comparison on the way does not stop a zero | `var r mo.Result[T]; if c { r = f() }; if r != (mo.Result[T]{}) { use(r) }` is reported | `ReportMeansAZeroIsUsed` |
| `result_zero_nokill.fsl` | A zero ever stored counts at the use | `...; if r == (mo.Result[T]{}) { ... }; r = f(); return r` is reported | `ReportMeansAZeroIsUsed` |

### Gap: a limitation the chosen rule accepts

The rule as design/rules.md writes it reports these, although no run gives the value. They are listed as gaps or consequences there. Each proved spec leaves them out by an `ASSUME` that names the spec.

| Spec | Counterexample | Why it is accepted | Named by | Breaks |
| --- | --- | --- | --- | --- |
| `nil_value_correlated.fsl` | `p := nil; if c { p = &T{} }; if c { return p }` is reported. The two branches test the same condition | A nil value is nil on some path, with branches taken as independent | ASSUME-1 of every proved spec | `ReportMeansANilIsReturned` |
| `nil_value_dead_check.fsl` | `p := &T{}; if p == nil { return p }` is reported. The branch is dead | Only a nil check of the value itself narrows the paths; whether it can succeed is not read | ASSUME-2 of `nil_value.fsl`, `pairing.fsl`, `result_zero.fsl` | `ReportMeansANilIsReturned` |
| `return_store_deferred.fsl` | `defer func() { if p == nil { p = def } }(); return nil` is reported. F returns `def` | A return with operands is judged on what it stores, and this one says nil | ASSUME-4 of `local_store.fsl` | `ReportMeansANilIsReturned` |

One more gap has no spec. A check on a loaded variable that SSA does not lift is not carried to later loads: a nil check, and a comparison with the zero Result as well. So `if p == nil { p = def }` on a captured p, or on a named result in a function with a defer, is reported through the Alloc's zero. `local_store.fsl` leaves it out by its ASSUME-3, and `result_zero.fsl` models lifted variables only.

### Not specified

| Topic | Why no spec |
| --- | --- |
| Range-over-func returns (`flow/range.go`) | `RangeReturns` picks the stores that carry a return statement's position. That is attribution of syntax to a return, not a path fact: a store has the position or it does not. `rangeBinding` reads a captured variable in the body as what reached the loop's closure. That holds when the body only loads the variable, which is the load-only-literal case `local_store.fsl` proves followed. A body that stores into it makes the variable one a literal stores into, which is not followed |
| Directives | Set logic over the table in design/rules.md: a spec would restate it |
| ChangeType | The rule and the run both see through it |

## Running them

```bash
./verify.sh
```

The script reads each verdict from `fslc`'s JSON with Python, never from its exit code. The proved specs must be `proved` under induction. The naive and gap specs must be `violated` on the named invariant, and the output says which kind each is. A spec in the directory that is named in no list fails the run. Without `fslc` the script skips the specs, except in CI (`CI` set), where it fails. CI installs the fslc version below.

## Mutation

`fslc mutate <spec> --depth 8`, measured on fslc 4.0.0:

| Spec | Mutants | Killed | Kill rate |
| --- | --- | --- | --- |
| `nil_value.fsl` | 200 | 152 | 76% |
| `pairing.fsl` | 200 | 148 | 74% |
| `pairing_stores.fsl` | 200 | 165 | 82% |
| `pairing_load.fsl` | 82 | 70 | 85% |
| `local_store.fsl` | 141 | 123 | 87% |
| `result_zero.fsl` | 155 | 125 | 81% |

Most survivors change which free choice bit picks an operand, or the bounds of the bit index. Both leave the set of programs the same. The rest touch state the run writes before it reads, the `abort` action of a run whose checks all fail, or a ghost that only the auxiliary invariants read.
