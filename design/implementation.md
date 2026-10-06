# Implementation notes

This file records design decisions and implementation details. The short, always-loaded instructions are in [AGENTS.md](../AGENTS.md). What molint reports is specified in [rules.md](rules.md). The history of the decisions is in [#1](https://github.com/mpyw/molint/issues/1).

## Project overview

**molint** is a Go linter that enforces [samber/mo](https://github.com/samber/mo). It is built on [`go/analysis`](https://pkg.go.dev/golang.org/x/tools/go/analysis) and [`buildssa`](https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/buildssa).

It reads the shape of signatures, and follows values only inside one function. It has no analysis facts, so it does not analyze dependencies. Finding nil panics across functions is left to [nilaway](https://github.com/uber-go/nilaway).

## Architecture

```text
analyzer.go              Analyzer, one flag per rule
skills.go                Skills: the embedded skills/ directory
plugin/                  the golangci-lint module plugin: the settings in place of the flags
skills/                  the agent skills, installed by `molint skill install`
cmd/molint/              singlechecker entry point, the skill subcommand, and -V=full
internal/                the rules: one flat package, one namespace per file
  run.go                 Run: every rule over the package, then the reports in order
  checker.go             the per-pass state, and reporting through the directives
  config.go              Config: which rules are on
  implementing.go        the methods that implement an interface, which some rules exempt
  shape.go               return-bool and return-error
  returnnil.go           return-nil
  calls.go               the loop over static calls, for the rules about mo calls
  wrapnil.go             wrap-nil
  unwrapnil.go           unwrap-nil
  unwrapdiscard.go       unwrap-discard
  resultzero.go          result-zero
  fieldnil.go            field-nil-store and field-nil-compare
internal/rule/           the rule names, shared by flags, messages and directives
internal/directive/      //molint: comments
internal/flow/           following a value back through one function
internal/nilcheck/       what the nil checks above a block say about a value
internal/typeutil/       type questions, samber/mo, and how messages spell types
```

The rules share one pass's state and nothing else, so they sit in one flat package with a file each. The file boundaries are checked by declscope.

Everything that stands alone has its own package:

| Package | Why it stands alone |
| --- | --- |
| `rule` | Names only. Flags, messages and directives all spell them |
| `directive` | Reads comments and source text only |
| `flow` | Reads SSA only. It knows nothing of rules or of samber/mo |
| `nilcheck` | Reads branches and dominators only |
| `typeutil` | Type questions with no state |

`plugin` builds its own copy of the analyzer from the settings, with the rules on that the settings say. Setting the flags of `molint.Analyzer` instead would be state shared by every run in the process. It lives in the main module, since `plugin-module-register` adds no dependency beyond `golang.org/x/tools`. A nested module would need a second tag per release, and would make `declscope shrink` stand down on `internal/`.

`plugin` builds its own copy of the analyzer from the settings, with the rules on that the settings say. Setting the flags of `molint.Analyzer` instead would be state shared by every run in the process. It lives in the main module, since `plugin-module-register` adds no dependency beyond `golang.org/x/tools`.

### declscope

The repository is checked by [declscope](https://github.com/mpyw/declscope) with `qualify: ondemand` and `exported: true` (`.declscope.yaml`). `declscope shrink` runs before the analyzer.

| Rule | Why |
| --- | --- |
| No `//declscope:core` | A core file hides its names from the naming rule |
| A rule's entry point is `//declscope:shared`, with the file that calls it | `run.go` and `calls.go` call into each rule's file |
| `flow` is split by concept: `site.go`, `tracer.go`, `store.go`, `return.go`, `range.go`, `pair.go` | Each name carries its file's concept, as `SiteAt`, `NilTracer`, `storesReaching`, `RangeReturns`. One file would have made every name carry `flow` |
| `nilcheck.Unknown`, `typeutil.TrailingNone` and `typeutil.Option` stay exported with no ignore | No other package names them, but each enum is incomplete without them. Since declscope 0.18.0, `shrink` judges a `const` block of one type as one set, so the other values keep them exported |

## Working with nilaway

molint and nilaway do not exclude each other. molint stops the ways of writing that make a nil. nilaway finds the nils that still reach a dereference. So a rule that needs to follow values across functions belongs to nilaway, not here.

A sample of 15 cases, with samber/mo v1.17.0 and nilaway `v0.0.0-20260918162853-acb8859b9031`, gave this:

| Case | molint | nilaway |
| --- | --- | --- |
| Code as molint asks: `Get` with `ok` checked, `MustGet`, `OrElse(guest)`, `ForEach`, `IsPresent` then `MustGet`, and the same for `mo.Result` | Nothing | Nothing |
| A map lookup dereferenced, as in `users[name].Name` | Nothing | Reported |
| nil passed as an argument, then dereferenced | Nothing | Reported |
| A field never set, dereferenced directly or through a function that returns it | Nothing | Nothing |
| `return nil`, then the result dereferenced | At the `return` | At the dereference |
| `u, _ := opt.Get()`, then `u.Name` | At `Get` | At the dereference |
| `opt.OrEmpty().Name` | Reported | Nothing |

What this means for molint:

- nilaway does not report code that follows molint's rules. It reads the `ok` of `Get` as a guard.
- Where both report one cause, molint reports the cause and nilaway the dereference. Fixing it as molint says clears both.
- A field that is never set is missed by both. molint gave up proving fields non-nil with the non-nil contract. `field-nil-store` with exhaustruct reports a nil written into a field, and `field-nil-compare` reports a field checked for nil. Only `field-nil-store` is on by default.
- nilaway does not know that `OrEmpty` gives nil for an empty Option. That is what `unwrap-nil` is for.

The sample is small. Run both over a real application before relying on these rows.

## Following a value

`flow.Tracer` tells whether a value may be a constant a rule looks for. `NilTracer` looks for nil, `TrueTracer` for the constant true, and `CheckingTracer` for any constant, such as a zero `mo.Result`.

| Value | Followed to |
| --- | --- |
| `Const` | Itself |
| `ChangeType` | Its operand |
| `Phi` | Each edge, judged at that edge |
| A load of an `Alloc` whose address is only loaded and stored | The stores that reach the load, and the `Alloc` itself as a store of the zero value |
| Anything else | Nothing. It is not a constant |

A nil check that dominates a site settles a value there. `nilcheck` reads a comparison with any zero constant, so a comparison of a struct with its zero value counts too. That is how `result-zero` stops at `if r == (mo.Result[int]{})`.

A function literal that captures a variable and only loads it keeps the variable followed. One that stores into it, or any other use of its address, stops the following.

### Returns

`flow.Returns` reads each `Return` of a function. A return that stores its results and loads them again is read as returning the stored values. A function with a `defer` does this. Only a return statement's own instructions may sit between the store and the load: other stores, `RunDefers`, and loads. So an assignment followed by a call is not mistaken for a return.

A range-over-func body is a synthetic function literal (`Synthetic == "range-over-func yield"`). A `return` in it stores the results into captured variables of the enclosing function. `flow.RangeReturns` maps each captured variable back to the result it holds, through the `MakeClosure` that binds it, and through nested bodies. The enclosing function's own load of those variables is not followed, since the body stores into them.

### Pairing

`flow.Pair` holds a nil pointer to the error or the bool beside it. Where either value is a φ, each incoming edge is judged on its own, with the other value read along the same edge. That repeats while φs remain, into earlier blocks. A value defined before the φ's block takes the checks on the edge too. Before a φ is taken apart, a nil check at the current sites settles the value for the whole path.

A bare return that loads both results from followed variables is paired along each path back to the stores, both values read on that path.

## Fields

`field-nil-store` and `field-nil-compare` do not prove a field non-nil. They look for evidence that a field is used as optional: a nil written into it, or a check for nil on it. A field with no such evidence is trusted. That is a gap, not a hole in a contract, since nothing else relies on the field being non-nil.

| Decision | Why |
| --- | --- |
| A field left out of a composite literal is left to exhaustruct | go/ssa emits no store for it. `compLit` clears memory only when the destination is not fresh, as in `*p = S{...}` |
| Both rules live in one file | They share which fields count and how a field is spelled |
| Only fields declared in the package count | molint has no facts, so it cannot tell whether another package's file is generated. Protobuf and SDK structs are full of `*string` |
| The field is taken from the generic type's origin | A diagnostic then spells `Box[E].V` with `*E` at every instance |
| `field-nil-store` is on by default | A nil written into a field says the field is absent there, so each report is right. Without exhaustruct it misses left-out fields, but it reports nothing wrong |
| `field-nil-compare` is off by default | A check says only that someone thought the field could be nil. It reports a field set on first use, and a check made just in case on a field that is never nil. There the fix is to delete the check, not to make the field an `mo.Option` |

## Rejected designs

| Design | Why not |
| --- | --- |
| Prove every returned pointer non-nil, with facts across packages (the first nilproof) | A method that returns its receiver, as Bubble Tea models do, was most of the reports on one application |
| Make `*T` a non-nil contract everywhere: check every write, trust every read | Every gap became a nil that reaches a trusted read. Closing them needed error proofs, facts derived from dependencies, a field invariant for `mo.Option`, a list of known functions, and definite assignment with defers and package initialization. Two adversarial reviews kept finding more |
| Forbid only known nils, and trust everything else | It is neither sound nor simple, and its value over nilaway was unclear |
| Special-case samber/mo by name beyond its package path | The rules name `Some`, `Ok`, `Err`, `Get`, `OrEmpty` and `OrElse` only. Everything else is read from types |
| Look for interfaces in every package in the import graph | Whether a method is exempt would depend on what some dependency happens to import. Only the package and its direct imports count |
| A list of well-known interfaces, such as `json.Marshaler` | It is a list to maintain, for a rule that is off by default. `return-error` users write an ignore instead |
| Exempt a mo constructor held in a local variable | SSA lowers `f := mo.Some[*T]; f(nil)` to a direct call, so exempting it would take extra work to hide a real nil |
| Report an implicit conversion to an interface where it is written | SSA gives it no position. It is reported where the converted value is used |
| Report diagnostics as they are found | The signature rules run before the SSA rules, so reports came out of line order. They are sorted by position at the end of the pass |

## The command

| Decision | Why |
| --- | --- |
| `-V=full` is registered before the driver | x/tools registers a `-V` that prints `devel` for every binary, but only when no `-V` is registered yet. `go vet -vettool` reads the line to identify the tool |
| A release stamps its version with `-X main.version` | goreleaser builds every platform into `dist/` in one checkout. Each build after the first sees untracked files, so the module version the go command records reads `+dirty` |
| `versionFlag` takes its output, its exit and the executable as fields | The tests reach every branch in one process. The coverage floor leaves no room for code that only a subprocess runs |
| `init` intercepts `molint skill` and registers `-V` | The test binary runs `init` too, so both are covered. `main` only hands over to the driver |
| No exported error for a pass without SSA | A driver runs `Requires` first, so only a hand-built `analysis.Pass` meets it. `internal.errRunWithoutSSA` stays as a guard |

## Testing

```bash
go test ./...            # analysistest over testdata/src, and unit tests
mise x -- ./test_all.sh  # tests, golangci-lint, declscope shrink, declscope, specs, coverage
```

- `testdata/src/*` are analysistest packages, one or more per rule. `github.com/samber/mo` there is a stub, with bodies copied from v1.17.0.
- `analyzer_test.go` runs the default flags, `-return-error`, `-field-nil-compare`, `-return-bool=false`, and `-field-nil-store=false`. It runs a package again after restoring a flag, to prove that flags are read on each run. No test may call `t.Parallel`, since the flags are global.
- A directive without a reason cannot be pinned in a fixture: the expectation comment on its line would be read as the reason. `internal/directive` tests it.
- **Coverage is held at 99.5% of statements** (`coverage.sh`). The only statements left are in `main`, which hands over to the driver. Code no input reaches is deleted, not excluded.

### Formal specs

`spec/*.fsl` model the rules that reason about paths: nil values, pairing, loads of variables that are not lifted, and zero Results. [spec/README.md](../spec/README.md) lists what each proves. `spec/verify.sh` is the gate. **When one of those rules changes, change its spec first.**
