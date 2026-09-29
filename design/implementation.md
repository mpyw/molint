# Implementation notes

This file records design decisions and implementation details. The short, always-loaded instructions are in [AGENTS.md](../AGENTS.md). Open questions are in [#1](https://github.com/mpyw/molint/issues/1).

## Project overview

**molint** is a Go linter that reports a returned pointer it cannot prove non-nil. It is built on [`go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) and [`buildssa`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/buildssa).

**Unproven is reported.** errlogreturn leans toward silence. molint leans the other way, because it is adopted on purpose by a new application that wants the guarantee. Where the proof cannot follow a value, the report says so, and the author returns `mo.Option`, checks the value, or writes an ignore with a reason.

## Architecture

```text
analyzer.go            Analyzer, ErrNoSSA
cmd/molint/          singlechecker entry point
internal/              the engine: one flat package, one namespace per file
  run.go               Run: summarize to a fixpoint, export facts, report
  checker.go           the per-pass state, with each stage's book embedded
  summary.go           function summaries, the fixpoint, and Fact import/export
  global.go            package-level pointer variables and their stores
  judge.go             the rule: which returns must prove which results
  proof.go             the proof of one value at one site
  report.go            diagnostic text
  fact.go              Fact and GlobalFact
internal/nilcheck/     what the nil checks above a block say about a value
internal/typeutil/     pointer and error results, and names as diagnostics spell them
internal/directive/    //molint: comments
```

The engine is flat because its parts are mutually recursive. A summary is the judgement of a function. The judgement proves values. A proof of a call reads the callee's summary, and a proof of a variable reads the variable's summary, whose stores are proven in turn. Splitting that cycle across packages would only add exports.

Everything that stands alone has its own package:

| Package | Why it stands alone |
| --- | --- |
| `nilcheck` | It reads branches and dominators only. It knows nothing of summaries |
| `typeutil` | Type questions with no state |
| `directive` | It reads comments only |

**Keep new code out of the flat package unless it joins that cycle.**

### declscope

The repository is checked by [declscope](https://github.com/mpyw/declscope) with `qualify: ondemand` and `exported: true` (`.declscope.yaml`). `declscope shrink` runs before the analyzer. The rules follow errlogreturn:

| Rule | Why |
| --- | --- |
| No `//declscope:core` | A core file hides its names from the naming rule |
| Each stage's state is a `...Book` struct embedded in `checker` | Call sites read `c.funcs`, and reaching another stage's state is a boundary crossing |
| A book is `//declscope:package`, and each field of it `//declscope:private` | `checker` spells the book, and no other file reads its fields |
| A method shared across files states `//declscope:package` with the reading file | `proofSite.state` and `proofFailure.origin` are read from `judge.go` and `report.go` |
| `nilcheck.Unknown` stays exported with an `overexported` ignore | No other package names it, but the enum is incomplete without its zero value |

`proof` is a type the other files never name. `judge.go` gets one from `newProof` and calls its package-scoped methods.

## The analysis

### Summaries and the fixpoint

A function's summary is the bit set of its pointer results proven on every return. A result past the 64th shifts out of the set, so it is never proven.

Inside a package, summaries are a greatest fixpoint. Every function starts with every pointer result proven. Each round judges every function against the summaries of the moment and takes back what failed. A round that takes nothing back ends it. Rounds only take proofs back, so they end.

| Consequence | Example |
| --- | --- |
| Mutual recursion is proven when its base cases are | `Even` and `Odd` in `errpair` |
| A function that never returns is proven | `Spin` in `errpair`. It is vacuously true, since it returns nothing |

Summaries cross packages as a `Fact` on the declared function. Every dependency is analyzed, the standard library included, so `strings.NewReader` is proven from its body. An instance of a generic function is read through its origin. A bound method or a thunk is read through the method it wraps.

A package-level pointer variable is proven when every store to it in its package is proven, and nothing else takes its address. The stores of variable initializers are in the package's `init` function, which is not among `buildssa`'s source functions, so it is read too. A store from another package, to an exported variable, is not seen.

### The rule

`judge` decides which returns must prove which results:

| Function | A return must prove its pointers when |
| --- | --- |
| No error result | Always |
| Last result is `error` | The error may be nil: the constant, a value checked to be nil, or a φ that may bring either |
| Last result is `error` | The error is taken as non-nil, but the pointer and the error come from one call. Then the callee must be proven |

A return whose error is a φ of its own block is judged along each incoming edge. Every other φ of that block is read along the same edge. So `var u *T; var err error; if b { u = &T{} } else { err = e }; return u, err` is proven.

### The proof

`proof.nonNil` proves one value at one site. A site is the start of a block, or the edge a φ reads the value from.

| Value | Proven when |
| --- | --- |
| `Alloc`, `FieldAddr`, `IndexAddr`, `Global`, `FreeVar` | Always. An address is never nil. A field or element of a nil pointer panics first |
| A value checked by a dominating branch | The check says non-nil (`nilcheck`) |
| `ChangeType` | Its operand is proven |
| `Phi` | Every edge is proven at its own edge |
| A call's result | The callee is proven. When it returns an error, that error must also be checked nil on the way |
| A load of a package-level variable | The variable is proven |

`nilcheck.At` walks the dominator tree. A dominating block with a single predecessor puts that edge on every path. SSA values are never redefined between a check and a use, so the check still holds.

A φ that reaches itself round a loop is taken as proven on the back edge. A φ's result is remembered for the judgement. A failure is always final. A success is remembered only when no cycle was assumed on the way.

## Rejected designs

| Design | Why not |
| --- | --- |
| Reporting `return nil` by syntax | `var p *T; return p`, a bare return of a named result, and a map miss all pass it |
| Proving errors non-nil too | `status.Error`, `errors.WithStack` and similar return nil for a nil input. Every `return nil, wrap(err)` would be reported. An error other than the constant nil is taken as non-nil |
| Trusting the `(*T, error)` convention for every callee | A dependency that really returns `nil, nil` would pass. Bodies are read wherever they exist. Only calls without a body are trusted |
| Granting a fact to a function whose report is ignored | The ignore records that the author accepts nil there. Callers must not rely on it |
| Suggested fixes | Most fixes change a signature and break callers in other packages. See #1 |
| Running molint on itself | molint is a tool, not the kind of application the rule is for. Its engine returns a nil `*proofFailure` to mean "proven" |
| Unexporting `nilcheck.Unknown`, as `declscope shrink` proposed | An exported enum with an unexported zero value cannot be spelled by its users |

## Known limitations

Each is a false report, not a missed one.

| Case | Why |
| --- | --- |
| `if x.f != nil { return x.f }` | Two loads of a field are two SSA values. The check proves the first only |
| A method that returns its receiver | A receiver is a parameter. Bubble Tea models return `m` from `Update` |
| A field set in every constructor | Fields are never proven. The zero value of a struct holds nil |
| `if !ok` after `p, ok := lookup()` | Only comparisons with nil are read. An `ok` flag says nothing about `p` |

## Testing

```bash
go test ./...          # analysistest over testdata/src, and unit tests
mise x -- ./test_all.sh  # tests, golangci-lint, declscope shrink, declscope, coverage
```

- `testdata/src/*` are analysistest packages. `lib` is imported by `crosspkg` to prove facts across packages.
- **Summaries are pinned as facts.** `Fact.String` renders `nonnil r0, r2`, and `GlobalFact.String` renders `nonnil`. Every declared function with a proven pointer result needs a `// want Name:"nonnil r0"` on its declaration line.
- A report is on the `return` line.
- An ignore without a reason cannot be tested in a fixture. The expectation comment on its line would be read as the reason. `internal/directive` tests it instead.
- `coverage/coverage.go` reaches the less common branches: every origin a diagnostic names, error φs round a loop, a result past the 64th.
- **Coverage is held at 99.5% of statements** (`coverage.sh`). The statements left are `main` and the two `AFact` methods. Code no input reaches otherwise is deleted, not excluded.
