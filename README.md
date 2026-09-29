# molint

Go linter that enforces [samber/mo](https://github.com/samber/mo): absence is `mo.Option`, not a nil pointer or a trailing `bool`.

> [!WARNING]
> Work in progress. The rules are strict on purpose. They are meant for applications that choose samber/mo, not for libraries.

## Overview

```go
func Find(name string) *User {
	u, ok := users[name]
	if !ok {
		return nil
	}
	return u
}

func Lookup(name string) (*User, bool) {
	u, ok := users[name]
	return u, ok
}

func Guest(o mo.Option[*User]) *User {
	return o.OrEmpty()
}
```

```console
$ molint ./...
user.go:12:3: Find returns a nil *User; return mo.Option[*User] instead [return-nil]
user.go:17:6: Lookup reports absence with a trailing bool; return mo.Option[*User] instead [return-bool]
user.go:23:18: OrEmpty on mo.Option[*User] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value [unwrap-nil]
```

Every report ends with the name of its rule. That name is what `//molint:ignore` takes.

molint reads the shape of signatures, and follows values only inside one function. It does not look for nil panics. Use it beside [nilaway](https://github.com/uber-go/nilaway):

| Tool | Job | Reads |
| --- | --- | --- |
| nilaway | Finds possible nil panics | The flow of values across functions |
| molint | Enforces the use of `mo.Option` and `mo.Result` | The shape of signatures and code |

## Install

| Method | Command | Needs |
| --- | --- | --- |
| `go tool` | `go get -tool github.com/mpyw/molint/cmd/molint@latest` | Go 1.27+ |
| `go install` | `go install github.com/mpyw/molint/cmd/molint@latest` | Go 1.27+ |

```bash
molint ./...                                 # or: go tool molint ./...
go vet -vettool=$(which molint) ./...        # through go vet, with its cache
```

## Rules

| Rule | Default | Reports |
| --- | --- | --- |
| [`return-nil`](#return-nil) | On | A nil pointer result |
| [`return-bool`](#return-bool) | On | A signature that ends in a `bool` after other results |
| [`return-error`](#return-error) | **Off** | A signature that ends in an `error` after other results |
| [`wrap-nil`](#wrap-nil) | On | nil given to `mo.Some`, `mo.Ok`, or `mo.Err` |
| [`result-zero`](#result-zero) | On | A zero `mo.Result` |
| [`unwrap-nil`](#unwrap-nil) | On | `OrEmpty`, or `OrElse(nil)`, on an Option or a Result of a pointer |
| [`unwrap-discard`](#unwrap-discard) | On | `Get` with its `ok` or its error discarded |

The line between them: absence must not be dropped silently. A trailing `bool` and a discarded `ok` drop it silently, whatever the type. `OrEmpty` and `OrElse` choose a default in plain sight. That is fine, unless the default is nil.

Each rule has a flag of its name:

```bash
molint -return-error ./...         # every rule
molint -return-bool=false ./...    # every rule except return-bool and return-error
```

Nothing is reported in a generated file. Test files are checked like any other file.

### `return-nil`

A return must not give a nil pointer.

| Return | Valid |
| --- | --- |
| `return nil` in `func F() *T` | ❌ |
| `var p *T; return p` | ❌ |
| `return nil, nil` in `func F() (*T, error)` | ❌ |
| `return nil, err` in `func F() (*T, error)` | ✅ The error is not nil |
| `return nil, false` in `func F() (*T, bool)` | ✅ For this rule. `return-bool` reports the signature |
| `return nil, true` in `func F() (*T, bool)` | ❌ |

A nil is followed through branches and loops in the function. A nil check on the way stops it:

```go
var p *T
if c {
	p = get()
}
if p == nil {
	p = def
}
return p // not reported
```

A function literal is exempt. So is a method that implements an interface, since the interface fixes its signature.

Fix: return `mo.Option[*T]`.

### `return-bool`

A signature must not end in a `bool` after at least one other result.

| Signature | Valid |
| --- | --- |
| `func Find() (User, bool)` | ❌ Whatever type comes before the `bool` |
| `func Cut() (string, string, bool)` | ❌ |
| `func IsAdmin() bool` | ✅ Nothing comes before the `bool` |

Functions, methods, methods of named interfaces, and named function types are checked. A function literal is exempt. So is a method that implements an interface: the interface's own declaration is reported instead, when it is in the package.

Fix: return `mo.Option[User]`. For several values, return `mo.Option` of a struct, or of a tuple of [samber/lo](https://github.com/samber/lo), such as `lo.Tuple2`.

> [!NOTE]
> A method is exempt only when molint sees the interface. It must be declared in the package, or in a package that the package imports directly. `MarshalJSON` in a package that does not import `encoding/json` is not exempt.

### `return-error`

The same as `return-bool`, with `error` in place of `bool`. It is off unless `-return-error` is set.

| Signature | Valid |
| --- | --- |
| `func Find() (*User, error)` | ❌ |
| `func Load() (Config, Meta, error)` | ❌ |
| `func Close() error` | ✅ |

Fix: return `mo.Result[*User]`.

### `wrap-nil`

| Call | Valid |
| --- | --- |
| `mo.Some[*T](nil)` | ❌ The option is present and holds nil |
| `mo.Ok[*T](nil)` | ❌ |
| `mo.Err[T](nil)` | ❌ The result is an error, and its error is nil |
| `mo.Some[[]int](nil)` | ✅ Only pointers and interfaces count for `Some` and `Ok` |

Fix: `mo.None[*T]()`, a non-nil value, or a non-nil error.

### `result-zero`

A zero `mo.Result` is Ok, holding the zero value of its type. It must not be used.

```go
func Load() mo.Result[Config] {
	var r mo.Result[Config]
	return r // reported
}
```

A return, an argument, a store, a send, and a method call are uses. A comparison is not. A zero `mo.Option` is None, which is fine, so it is not reported.

Fix: build it with `mo.Ok` or `mo.Err`.

### `unwrap-nil`

| Call on `mo.Option[*T]` or `mo.Result[*T]` | Valid |
| --- | --- |
| `o.OrEmpty()` | ❌ It gives nil when the option is empty |
| `o.OrElse(nil)` | ❌ |
| `o.OrElse(&guest)` | ✅ |
| `o.MustGet()` | ✅ It panics rather than give nil |
| `o.OrEmpty()` on `mo.Option[int]` | ✅ The zero value is chosen in plain sight |

Fix: `Get` with a check of `ok`, or `OrElse` with a non-nil value.

### `unwrap-discard`

| Call | Valid |
| --- | --- |
| `v, _ := o.Get()`, then `v` is used | ❌ Whatever the type of `v` |
| `v, _ := r.Get()` on a `mo.Result`, then `v` is used | ❌ The error is discarded |
| `if v, ok := o.Get(); ok { ... }` | ✅ |
| `o.Get()` as a statement | ✅ Nothing is used |

Fix: check `ok` or the error, or use `OrElse`.

## Ignoring a report

Write `//molint:ignore` with the rules to silence and a reason after `//`.

```go
//molint:ignore return-bool // callers rely on the comma-ok form
func Lookup(name string) (*User, bool) {
```

| Placement | Silences |
| --- | --- |
| On a line of its own | The line below |
| After code | Its own line |

| Directive | Result |
| --- | --- |
| `//molint:ignore return-nil, wrap-nil // reason` | Silences both rules |
| `//molint:ignore // reason` | Silences every rule |
| No reason | Reported, and silences nothing |
| An unknown rule name | Reported |
| An ignore that silences nothing | Reported as unused, unless every rule it names is turned off |

## Limits

These are not checked:

| Case | Why |
| --- | --- |
| A nil from a parameter, a field, or a call | Values are followed only inside one function. nilaway follows them further |
| `mo.TupleToOption`, `mo.TupleToResult`, `mo.EmptyableToOption` | They check their arguments at run time |
| A zero `mo.Result` left out of a composite literal, as in `Holder{}` | Fields are not followed |
| A constructor or method of mo passed as a function value | Calls through function values are not followed |

These are reported although no run returns nil, since branches are taken as independent:

| Case | Instead |
| --- | --- |
| `if c { p = x }; if c { return p }` | Keep the value and its condition together, as in `mo.Option` |
| The pointer and the error set on separate branches, then `if err != nil { return nil, err }; return u, nil` | Return from each branch |
| A retry loop that ends with `return nil, err` | Start the error at a sentinel, so that no round leaves it nil |

The full specification, with every limit, is [design/rules.md](design/rules.md).

## License

[MIT](LICENSE)
