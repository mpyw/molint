# Rules

This file specifies what molint reports. It is the source of truth for the fixtures under `testdata/src/` and the formal specs under `spec/`. When a rule changes, change this file first, then its spec and its fixtures, then the code. The decisions behind it are in [#1](https://github.com/mpyw/molint/issues/1).

## Scope

molint enforces the use of [samber/mo](https://github.com/samber/mo). It reads the shape of signatures, and it follows values only inside one function. It does not look for nil panics: [nilaway](https://github.com/uber-go/nilaway) does that.

| Group | Rules | What it enforces |
| --- | --- | --- |
| Shape of results | `return-nil`, `return-bool`, `return-error` | Absence and failure are `mo.Option` and `mo.Result`, not a nil pointer, a trailing `bool`, or a trailing `error` |
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

`return-nil`, `wrap-nil` and `unwrap-nil` report a value that is nil for sure at that point. Only these count:

| Value | Example |
| --- | --- |
| The constant nil | `return nil`. A variable declared without a value and never set before the use, such as `var p *T; return p`, is the constant in SSA |
| A φ with an edge that brings a nil value | `var p *T; if c { p = x }; return p` |
| A `ChangeType` of a nil value | A conversion between pointer types |
| A value that a dominating branch checks to be nil | `if p == nil { return p }` |
| A load of a local variable that is not lifted, where a store reaching it is a nil value | A result of a function with a `defer`, or a variable a function literal captures |

A φ edge, and every step above, is judged at its own edge: a nil check on the way stops it. So `if p == nil { p = def }; return p` is not a nil value.

A load of a local that is not lifted follows the stores reaching it within the function. The `Alloc` itself counts as a store of the zero value at its position, since SSA emits no store for `var p *T` or a named result. A function literal that captures the variable and only loads it, such as `defer func() { if err != nil { log(err) } }()`, does not stop this. When anything else uses the variable's address, its value is not followed, and it is then not a nil value. That covers a function literal that stores into it, `&x` passed to a call, and `&x` stored anywhere.

Anything else is not a nil value: a parameter, a field, a call's result, a map lookup. nilaway is the tool for those.

## `return-nil`

A return must not give a nil value in a pointer result.

| Function | A nil pointer result is reported when |
| --- | --- |
| No trailing `error` or `bool` result | Always |
| Last result is `error` | The error is also a nil value. So `return nil, nil` is reported and `return nil, err` is not |
| Last result is `bool` | The bool is the constant `true`. So `return nil, false` is not reported: `return-bool` reports the signature |

"Last result is `error`" means a type identical to the predeclared `error`. "Last result is `bool`" means a type identical to the predeclared `bool`.

A return is judged along each path that reaches it through φs. Where the error or bool is a φ, each incoming edge is judged on its own, and every other φ of the same block is read along the same edge. When the values on that edge are φs of the same earlier block, the same is done there, recursively. So `var u *T; var err error; if b { u = x } else { err = e }; return u, err` is not reported, and neither is the same shape nested inside another `if`.

Some returns store their results into local variables, then load them: a function with a `defer`, a function whose named results a function literal captures, and a `return` inside the body of a range-over-func loop. A `return` statement with operands is judged on the values it stores, whatever captures the variables. A bare `return` is judged on the stores reaching it, followed as [Nil values](#nil-values) says. A `return` inside a range-over-func body is judged as a return of the enclosing function, not of the synthetic function literal.

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
| A named interface or function type declared inside a function body | The same as at package level |

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

## `wrap-nil`

A nil value must not be passed to these constructors of mo:

| Call | Reported when |
| --- | --- |
| `mo.Some[T](v)` | `T` is a pointer or an interface, and `v` is a nil value |
| `mo.Ok[T](v)` | The same |
| `mo.Err[T](err)` | `err` is a nil value |

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

A use is reported: a return, an argument, a conversion to an interface (reported at the conversion), a store, a method call on it, a send, a map update. These are not uses:

- a comparison;
- the store that initializes a local variable, as in `r := mo.Result[int]{}` for a variable whose address is taken. The loads of that variable are judged instead;
- the store a `return` makes into its result variable. The return is judged instead, once.

A zero `Result` written into a field or an element, as in `Holder{R: mo.Result[int]{}}`, is a store and is reported. A field left out of a composite literal, as in `Holder{}`, is not followed.

Message: `a zero mo.Result[T] is Ok with a zero value; build it with mo.Ok or mo.Err [result-zero]`.

## `unwrap-nil`

These calls must not be made on an `Option[T]` or a `Result[T]` whose `T` is a pointer:

| Call | Reported when |
| --- | --- |
| `o.OrEmpty()` | Always |
| `o.OrElse(v)` | `v` is a nil value |

`o.MustGet()` is not reported: it panics rather than give nil. A `T` that is not a pointer is not reported: `OrEmpty` then chooses the zero value in plain sight.

Message: `OrEmpty on mo.Option[*T] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value [unwrap-nil]`. For `OrElse`: `OrElse(nil) on mo.Option[*T] gives nil when it is empty; pass a non-nil value [unwrap-nil]`. For a `Result`: `OrEmpty on mo.Result[*T] gives nil when it is an error; use Get and check the error, or OrElse with a non-nil value [unwrap-nil]`, and `OrElse(nil) on mo.Result[*T] gives nil when it is an error; pass a non-nil value [unwrap-nil]`.

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

- the interface is a named or aliased, non-generic interface type, declared in the package (at package level or inside a function) or in a package the package imports directly;
- the method's name is one of the interface's methods;
- the receiver's type, or a pointer to it, implements the interface.

## Directives

`//molint:ignore <rules> // <reason>` silences the named rules. A directive on a line of its own silences the line below. A directive that trails code silences its own line only. The rules are comma-separated. With no rules, it silences every rule. Reports about directives themselves cannot be silenced. An ignore aimed at one silences nothing, so it is reported as unused.

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

A directive in a generated file is not read.

## Flags

Each rule has a boolean flag of its name. Every rule is on by default, except `return-error`.

```bash
molint ./...                     # every rule except return-error
molint -return-error ./...       # every rule
molint -return-bool=false ./...  # every rule except return-bool and return-error
```

## Gaps

| Gap | Why |
| --- | --- |
| A nil that comes from outside the function, or from a field or a call | molint follows values only inside one function |
| `mo.TupleToOption`, `mo.TupleToResult`, `mo.EmptyableToOption` | They check their arguments at run time |
| A zero `Result` inside a struct | Fields are not followed |
| A constructor or method of mo passed through a parameter or a field, or a method expression such as `mo.Option[*T].OrElse(o, nil)` | They are called through a function value or a thunk, which is not followed |
| A method that implements an interface of a package its package does not import directly | The interface is not seen, so the method is not exempt. `MarshalJSON` in a package that does not import `encoding/json` is reported by `return-error`. Write an ignore |
