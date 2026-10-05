# Rules

This file specifies what molint reports. It is the source of truth for the fixtures under `testdata/src/` and the formal specs under `spec/`. When a rule changes, change this file first, then its spec and its fixtures, then the code. The decisions behind it are in [#1](https://github.com/mpyw/molint/issues/1).

## Scope

molint enforces the use of [samber/mo](https://github.com/samber/mo). It reads the shape of signatures, and it follows values only inside one function. It does not look for nil panics: [nilaway](https://github.com/uber-go/nilaway) does that.

| Group | Rules | What it enforces |
| --- | --- | --- |
| Shape of results | `return-nil`, `return-bool`, `return-error` | Absence and failure are `mo.Option` and `mo.Result`, not a nil pointer, a trailing `bool`, or a trailing `error` |
| Shape of fields | `field-nil-store`, `field-nil-compare` | Absence in a field is `mo.Option`, not a nil pointer |
| Use of mo | `wrap-nil`, `result-zero`, `unwrap-nil`, `unwrap-discard` | An `Option` or a `Result` never holds or gives a nil it should not |

The principle behind the lines: absence must not be dropped silently. A trailing `bool` and a discarded `ok` drop it silently, whatever the type. `OrEmpty` and `OrElse` choose a default in plain sight, which is fine unless the default is nil.

### Terms

| Term | Meaning |
| --- | --- |
| Pointer | A type whose underlying type is a pointer. A type parameter is not one, whatever its constraint |
| mo | The package `github.com/samber/mo` |
| `Option[T]`, `Result[T]` | An instance of `mo.Option` or `mo.Result` |
| Nil value | See [Nil values](#nil-values) |
| Generated file | A file with a `// Code generated ... DO NOT EDIT.` comment, as `ast.IsGenerated` reads it |

Nothing is reported in a generated file. Test files are checked like any other file.

## Nil values

`return-nil`, `field-nil-store`, `wrap-nil` and `unwrap-nil` report a value that is nil on some path to that point. The branches on the way are taken as independent: only a nil check of the value itself narrows the paths. Only these count:

| Value | Example |
| --- | --- |
| The constant nil | `return nil`. A variable declared without a value and never set before the use, such as `var p *T; return p`, is the constant in SSA |
| A φ with an edge that brings a nil value | `var p *T; if c { p = x }; return p` |
| A `ChangeType` of a nil value | A conversion between pointer types |
| A value that a dominating branch checks to be nil | `if p == nil { return p }` |
| A load of a local variable that is not lifted, where a store reaching it is a nil value | A result of a function with a `defer`, or a variable a function literal captures |

A φ edge, and every step above, is judged at its own edge: a nil check on the way stops it. So `if p == nil { p = def }; return p` is not a nil value.

A load of a local that is not lifted follows the stores reaching it within the function. The `Alloc` itself counts as a store of the zero value at its position, since SSA emits no store for `var p *T` or a named result. A function literal that captures the variable and only loads it does not stop this. `defer func() { if err != nil { log(err) } }()` is one. When anything else uses the variable's address, its value is not followed, and it is then not a nil value. That covers a function literal that stores into it, `&x` passed to a call, and `&x` stored anywhere.

Anything else is not a nil value: a parameter, a field, a call's result, a map lookup. nilaway is the tool for those.

## `return-nil`

A return must not give a nil value in a pointer result.

| Function | A nil pointer result is reported when |
| --- | --- |
| No trailing `error` or `bool` result | Always |
| Last result is `error` | The error is also a nil value. So `return nil, nil` is reported and `return nil, err` is not |
| Last result is `bool` | The bool is the constant `true`. So `return nil, false` is not reported: `return-bool` reports the signature |

"Last result is `error`" means a type identical to the predeclared `error`. "Last result is `bool`" means a type identical to the predeclared `bool`.

A return is judged along each path that reaches it through φs. Where the error or bool is a φ, each incoming edge is judged on its own. Every other φ of the same block is read along the same edge. When the values on that edge are φs of the same earlier block, the same is done there, recursively. So this is not reported, and neither is the same shape nested inside another `if`:

```go
var u *T
var err error
if b {
	u = x
} else {
	err = e
}
return u, err
```

A bare return may load both results from variables SSA does not lift, as in a function with a `defer`. It is judged the same way. Each path back to the stores is judged on its own, with both results read on that path.

Some returns store their results into local variables, then load them. These are a function with a `defer`, a function whose named results a function literal captures, and a `return` inside the body of a range-over-func loop. A `return` statement with operands is judged on the values it stores, whatever captures the variables. So `return nil` is reported even when a deferred function replaces the nil later. The return says nil, and that is what the rule reads. A bare `return` is judged on the stores reaching it, followed as [Nil values](#nil-values) says. A `return` inside a range-over-func body is judged as a return of the enclosing function, not of the synthetic function literal.

Exempt:

- a function literal;
- a method that implements an interface (see [Implementing an interface](#implementing-an-interface)).

Message: `F returns a nil *T; return mo.Option[*T] instead [return-nil]`. With a nil error: `F returns a nil *T with a nil error; return an error, or mo.Option[*T] [return-nil]`. With `true`: `F returns a nil *T with ok true; return mo.Option[*T] [return-nil]`.

## `return-bool`

A signature must not end in a `bool` result after at least one other result.

| Signature | Reported |
| --- | --- |
| `func F() (T, bool)` | Yes, whatever `T` is |
| `func F() (A, B, bool)` | Yes |
| `func F() bool` | No. Nothing comes before the `bool` |
| `func F() (bool, error)` | No. The `bool` is not last |
| `func F() (T, MyBool)` with `type MyBool bool` | No. Only the predeclared `bool` counts |

Checked declarations:

| Declaration | Reported at |
| --- | --- |
| A function or a method | Its name |
| A method of a named interface type | The method's name |
| A named function type, `type F func() (T, bool)` | The type's name |
| A named interface type declared inside a function body | The method's name |
| A named function type declared inside a function body | The type's name |

An alias of an interface or function literal, such as `type A = interface{ M() (int, bool) }`, is not reported, since it names no new type. It still counts for the exemption.

Exempt:

- a function literal, since the function it is passed to usually fixes its signature;
- a method that implements an interface (see [Implementing an interface](#implementing-an-interface)). The interface's own declaration is reported instead, when it is in the package.

Message: `F reports absence with a trailing bool; return mo.Option[T] instead [return-bool]`. With two to nine values before the `bool`, the suggestion is `mo.Option of a struct, or of lo.TupleN`, with N the count. With more, it is `mo.Option of a struct`.

### Names and types in messages

| Declaration | Spelled |
| --- | --- |
| Function | `F`, or `F[E]` for a generic one |
| Method | `S.M` or `(*S).M`, with type parameters as `(*Box[E]).Get` |
| Interface method | `Finder.Find`, or `GenericFinder[E].Find` |
| Function type | `F`, or `F[E]` for a generic one. Several parameters are joined by `, `, as in `Get[K, V]` |

A type from another package is qualified by its package name, as in `*bytes.Buffer`. Aliases are resolved at every level, so `type PT = *T` prints as `*T`, and `mo.Option[PT]` as `mo.Option[*T]`. A named pointer type prints as its name.

## `return-error`

The same as `return-bool`, with `error` in place of `bool`. It is off unless `-return-error` is set.

| Signature | Reported |
| --- | --- |
| `func F() (T, error)` | Yes |
| `func F() (A, B, error)` | Yes |
| `func F() error` | No |

Message: `F reports failure with a trailing error; return mo.Result[T] instead [return-error]`. With several values, `mo.Result of a struct, or of lo.TupleN`.

## `field-nil-store`

A nil value must not be stored into a pointer field.

> [!IMPORTANT]
> Run [exhaustruct](https://github.com/GaijinEntertainment/go-exhaustruct) beside it. Without it, a field left out of a composite literal hides a nil from this rule. What the rule reports without it is still right. It only misses more.

A field counts when all of these hold:

- its type is a pointer;
- it is not embedded, since `mo.Option` would lose the promoted fields and methods;
- it is declared in the package, in a file that is not generated. A field of an anonymous struct counts too.

A store counts wherever it is written:

| Write | Example |
| --- | --- |
| A keyed element of a composite literal | `S{F: nil}`, `&S{F: nil}` |
| A positional element of a composite literal | `S{1, nil}` |
| An assignment to the field | `s.F = nil`, `p.F = nil`, `s.Inner.F = nil` |

The value is judged as [Nil values](#nil-values) says. So `var p *T; if c { p = x }; s.F = p` is reported. `if p == nil { p = def }; s.F = p` is not.

A field left out of a composite literal is not followed. SSA emits no store for it, since the new struct is already zero. The exhaustruct linter makes every field written. Then `S{A: 1}` becomes `S{A: 1, F: nil}`, and `field-nil-store` reports that.

> [!NOTE]
> A struct filled after it is built is reported too. `s := S{A: 1, F: nil}; s.F = x` reports the nil in the literal. Compute `x` first, then build `S` with it.

A field whose type points back to its own struct, as in `type Node struct { Next *Node }`, is not exempt. `mo.Option[Node]` would be an invalid recursive type, so the message suggests `mo.Option[*Node]`, as `return-nil` does.

Message: `S.F is set to nil; make it mo.Option[*T] [field-nil-store]`. The struct is spelled as in [Names and types in messages](#names-and-types-in-messages), as in `Box[E].F`. A field of an anonymous struct is spelled by its name alone, as `F`.

## `field-nil-compare`

A pointer field must not be compared with nil. A nil check says the field may be absent, so the field should be an `mo.Option`. It is off unless `-field-nil-compare` is set.

A field counts as for [`field-nil-store`](#field-nil-store). This rule does not need exhaustruct. It also finds a field that only a decoder such as `json.Unmarshal` leaves nil.

| Comparison | Reported |
| --- | --- |
| `s.F == nil`, `p.F != nil`, `nil == s.F` | Yes |
| `switch s.F { case nil: }` | Yes |
| `v := s.F; if v == nil {}` | Yes. A local variable SSA lifts is the field read itself |
| `v := s.F; if c { v = x }; if v == nil {}` | No. The value is a φ, not the field read |
| A field of a struct declared in another package | No |

A comparison counts when one operand is the constant nil and the other is a read of the field. A read is a `Field`, or a load through a `FieldAddr`, seen through any `ChangeType`.

> [!NOTE]
> A field set on first use is reported too, as in `if c.client == nil { c.client = dial() }`. It is absent until then. Use `sync.OnceValue`, or make it an `mo.Option`.

Message: `S.F is compared with nil; make it mo.Option[*T] [field-nil-compare]`. The field is spelled as for `field-nil-store`.

## `wrap-nil`

A nil value must not be passed to these constructors of mo:

| Call | Reported when |
| --- | --- |
| `mo.Some[T](v)` | `T` can be nil, and `v` is a nil value |
| `mo.Ok[T](v)` | `T` can be nil, and `v` is a nil value |
| `mo.Err[T](err)` | `err` is a nil value |

`T` can be nil when it is a pointer, an interface, a map, a func, a channel or a slice. A slice counts too, although a nil slice works as an empty one. A present option that holds nil says two things at once, and samber/mo encodes `mo.Some[[]int](nil)` in JSON as `null`, as it does `mo.None`. A type parameter counts as none of these, whatever its constraint.

A call counts when SSA calls the constructor statically. A constructor held in a local variable, as in `f := mo.Some[*T]; f(nil)`, lowers to a static call and is reported. One passed through a parameter or a field is not followed.

Message: `mo.Some is given nil; pass a non-nil value, or use mo.None [wrap-nil]`. `mo.Ok is given nil; pass a non-nil value [wrap-nil]`. `mo.Err is given a nil error; pass a non-nil error [wrap-nil]`.

## `result-zero`

A zero `Result[T]` must not be used. It is Ok, holding the zero value of `T`.

A zero `Result[T]` is one of:

| Value | Example |
| --- | --- |
| The zero constant of `Result[T]` | `mo.Result[T]{}`, or `var r mo.Result[T]` used before any assignment |
| A φ with an edge that brings one | `var r mo.Result[T]; if c { r = mo.Ok(v) }; return r` |
| A load of a local that is not lifted, where a store reaching it is one | As for nil values |
| A load through `new(mo.Result[T])` before any store | `p := new(mo.Result[T]); return *p` |

A use is reported: a return, an argument, a conversion to an interface, a store, a method call on it, a send, a map update. A conversion is reported where it is. An implicit conversion has no position of its own, so it is reported where the converted value is used. These are not uses:

- a comparison;
- the store that initializes a local variable, as in `r := mo.Result[int]{}` for a variable whose address is taken. The loads of that variable are judged instead;
- the store a `return` makes into its result variable. The return is judged instead, once.

A comparison with the zero constant on the way stops a zero, as a nil check does. So `if r == (mo.Result[int]{}) { r = mo.Err[int](e) }; return r` is not reported.

A zero `Result` written into a field or an element, as in `Holder{R: mo.Result[int]{}}`, is a store and is reported. A field left out of a composite literal, as in `Holder{}`, is not followed. exhaustruct makes it written, as for [`field-nil-store`](#field-nil-store).

Message: `a zero mo.Result[T] is Ok with a zero value; build it with mo.Ok or mo.Err [result-zero]`.

## `unwrap-nil`

These calls must not be made on an `Option[T]` or a `Result[T]` whose nil breaks on use. That is a `T` that is a pointer, an interface, a map, a func or a channel. A slice is left out, unlike for [`wrap-nil`](#wrap-nil): a nil slice that comes out works as an empty one.

| Call | Reported when |
| --- | --- |
| `o.OrEmpty()` | Always |
| `o.OrElse(v)` | `v` is a nil value |

`o.MustGet()` is not reported: it panics rather than give nil. Any other `T` is not reported, a slice included: `OrEmpty` then chooses a zero value that works, in plain sight.

Message: `OrEmpty on mo.Option[*T] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value [unwrap-nil]`. For `OrElse`: `OrElse(nil) on mo.Option[*T] gives nil when it is empty; pass a non-nil value [unwrap-nil]`. For a `Result`, the messages are these:

- `OrEmpty on mo.Result[*T] gives nil when it is an error; use Get and check the error, or OrElse with a non-nil value [unwrap-nil]`
- `OrElse(nil) on mo.Result[*T] gives nil when it is an error; pass a non-nil value [unwrap-nil]`

## `unwrap-discard`

The second result of `Get` must not be discarded, whatever `T` is.

| Call | Reported when |
| --- | --- |
| `v, _ := o.Get()` on an `Option[T]` | The `ok` is discarded and the value is used |
| `v, _ := r.Get()` on a `Result[T]` | The error is discarded and the value is used |

A result is discarded when no `Extract` of it exists, or its `Extract` has no referrers. So `v, ok := o.Get(); _ = ok; use(v)` discards `ok`. A value is used when its `Extract` has a referrer. So `v, _ := o.Get(); _ = v` uses nothing, and is not reported. Discarding both, as in `o.Get()` as a statement or `_, _ = o.Get()`, is not reported either.

Message: `Get on mo.Option[T] discards ok; check it, or use OrElse [unwrap-discard]`. For `Result`: `Get on mo.Result[T] discards the error; check it [unwrap-discard]`.

## Implementing an interface

A method implements an interface when all of these hold:

- the interface is a named or aliased interface type that is not generic;
- it is declared in the package, at package level or inside a function, or in a package the package imports directly;
- the method's name is one of the interface's methods;
- for a method outside the test files, the interface is declared, or its package imported, outside the test files too;
- the receiver's type, or a pointer to it, implements the interface.

A package with tests is checked twice: alone, and with its test files. The rule for test files makes both checks agree on a method. Otherwise an interface that only a test imports would exempt the method in one check and not the other, and no ignore could satisfy both.

## Positions

| Report | Position |
| --- | --- |
| `return-nil`, `result-zero` at a return | The `return` keyword, also for a return inside a range-over-func body |
| `field-nil-store` | The colon of a keyed element. The value of a positional element. The field's name in an assignment |
| `field-nil-compare` | The operator of the comparison. The `nil` of a `case nil` |
| A rule about a call | The opening parenthesis of the call. A trailing ignore goes on that line, which is the first line of a call that spans several |
| `return-bool`, `return-error` | The name of the declaration |
| `result-zero` at a store, a send, or a map update | The statement |

## Directives

`//molint:ignore <rules> // <reason>` silences the named rules. It is a line comment: `/*molint:ignore*/` is not a directive. A directive on a line of its own silences the line below. A directive that trails code silences its own line only. The rules are comma-separated. With no rules, it silences every rule. Reports about directives themselves cannot be silenced. An ignore aimed at one silences nothing, so it is reported as unused.

| Directive | Result |
| --- | --- |
| `//molint:ignore return-bool // reason` | Silences `return-bool` there |
| `//molint:ignore return-nil, wrap-nil // reason` | Silences both |
| `//molint:ignore // reason` | Silences every rule |
| No reason after `//` | Reported: `molint:ignore needs a reason after //` |
| An unknown rule name | Reported: `unknown rule x in molint:ignore`. The known rules beside it still apply. An ignore whose rules are all unknown is not also reported as unused |
| An empty rule name, as in `return-nil, // reason` | Reported: `empty rule name in molint:ignore`. The known rules beside it still apply, as for an unknown name |
| An ignore that silences nothing | Reported: `unused molint:ignore directive`. An ignore whose every rule is turned off by a flag is not reported |
| Any other `//molint:` directive | Reported: `unknown directive molint:x` |

A directive in a generated file is not read. An ignore without a reason is reported and silences nothing. When two ignores silence one report, both count as used. An ignore that names several rules is used when any of them silences something.

## Flags

Each rule has a boolean flag of its name. Every rule is on by default, except `return-error` and `field-nil-compare`. The golangci-lint plugin takes the same names, as keys of its `settings`. A rule left out keeps its default, and an unknown name is an error.

```bash
molint ./...                                  # every rule except return-error and field-nil-compare
molint -return-error -field-nil-compare ./... # every rule
molint -return-bool=false ./...               # every rule on by default, except return-bool
```

## Gaps

| Gap | Why |
| --- | --- |
| A nil that comes from outside the function, or from a field or a call | molint follows values only inside one function |
| Branches whose conditions depend on each other, as in `if c { p = x }; if c { return p }` | Branches are taken as independent, so this is reported although no run returns nil. A nil check that can never succeed is reported too, as in `p := &T{}; if p == nil { return p }` |
| A nil check, or a comparison with the zero `Result`, on a variable that SSA does not lift. `if p == nil { p = def }` on a captured `p`, or on a named result in a function with a `defer`, is one | A later load follows the stores that reach it. The check on an earlier load of the same variable is not carried over. A nil constant stored on an earlier path is then reported |
| `mo.TupleToOption`, `mo.TupleToResult`, `mo.EmptyableToOption` | They check their arguments at run time |
| A field left out of a composite literal, for `field-nil-store` and `result-zero` | SSA emits no store for it. exhaustruct makes it written |
| A struct born zero outside a composite literal: `var s S`, `new(S)`, `make([]S, n)`, a map lookup that misses, or a decoder such as `json.Unmarshal` | Following each field until the struct is used needs definite assignment. molint gave that up with the non-nil contract. `field-nil-compare` still finds such a field where it is checked for nil |
| A nil stored into a field of a struct declared in another package | Whether that file is generated is not known without facts, so generated code such as protobuf would be reported |
| A zero struct stored as a whole, as in `*p = S{}` or `s.Inner = Inner{}` | Only stores into a pointer field itself are read. exhaustruct reports the literal |
| The initializer of a package-level variable, as in `var s = S{P: nil}` | It runs in the package's initializer, which SSA builds as a synthetic function. No rule that reads SSA analyzes it |
| A constructor or method of mo reached indirectly. That is through a parameter, a field, a method value (`f := o.OrEmpty; f()`), or a method expression (`mo.Option[*T].OrElse(o, nil)`) | They are called through a function value, a bound wrapper, or a thunk, which is not followed |
| An error checked after the pointer and the error are set on separate branches. `switch { case a: u = x; default: err = e }` then `if err != nil { return nil, err }; return u, nil` is one | The check is on another value, so it does not narrow the pointer. This is reported. Return from each branch instead |
| A retry loop that returns its last error. `for range 3 { if u, err = f(); err == nil { return u, nil } }` then `return nil, err` is one | The loop is taken as able to run no rounds, which leaves the error nil. Start the error at a sentinel instead |
| An ignore above a `//line` comment | The ignore targets the `//line` comment's own line, so it is reported as unused |
| A method that implements an interface of a package its package does not import directly | The interface is not seen, so the method is not exempt. `MarshalJSON` in a package that does not import `encoding/json` is reported by `return-error`. Write an ignore |
