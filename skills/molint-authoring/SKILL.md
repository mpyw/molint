---
name: molint-authoring
description: "Write or change Go code in a repository that runs molint, which shows as molint in CI or a mise.toml, or as //molint:ignore comments. Read this before acting on a molint report. Read it too before returning or storing a pointer that may be nil, or adding a field or a result that may be absent. Covers how to fix each rule with samber/mo, the fixes that only hide a problem, and when an ignore is right."
license: MIT
---

# Writing code under molint

molint enforces [samber/mo](https://github.com/samber/mo). Absence is `mo.Option`, not a nil pointer or a trailing `bool`. Failure may be `mo.Result`, not a trailing `error`. An `Option` or a `Result` never holds or gives a nil it should not.

The full specification is `design/rules.md` in the molint repository. This skill says how to write code that satisfies it.

## The model

**A nil pointer must never mean "absent".** Where a value may be missing, the type says so: `mo.Option[T]`. A plain `*T` then always points somewhere.

| Rule | Default | Reports | The fix in one line |
| --- | --- | --- | --- |
| `return-nil` | On | A return that gives a nil pointer | Return `mo.Option` |
| `return-bool` | On | A signature that ends in `bool` after other results | Return `mo.Option` |
| `return-error` | Off | A signature that ends in `error` after other results | Return `mo.Result` |
| `field-nil-store` | Off | nil stored into a pointer field | Make the field `mo.Option` |
| `field-nil-compare` | Off | A pointer field compared with nil | Make the field `mo.Option` |
| `wrap-nil` | On | nil given to `mo.Some`, `mo.Ok` or `mo.Err` | Use `mo.None`, or a non-nil value |
| `result-zero` | On | A zero `mo.Result` used | Build it with `mo.Ok` or `mo.Err` |
| `unwrap-nil` | On | `OrEmpty`, or `OrElse(nil)`, where nil breaks on use | Use `Get` and check `ok`, or `OrElse` with a real default |
| `unwrap-discard` | On | `Get` with its `ok` or its error discarded | Check it |

**Check which rules are on before you trust a clean run.** Each rule has a flag of its name. Read how CI runs molint. `-return-error`, `-field-nil-store` and `-field-nil-compare` turn on the rules that are off by default. `-return-bool=false` turns one off.

molint follows values only inside one function. It does not look for nil panics. [nilaway](https://github.com/uber-go/nilaway) does that, and the two are meant to run together.

## Fixing each report

Each report ends with its rule name in brackets. Fix the cause the rule names. Do not change the code until the report goes away.

### `return-nil` and `return-bool`

Both say that the function reports absence in a way its result type does not show.

```go
// Reported by return-nil.
func Find(id ID) *User {
	if u, ok := users[id]; ok {
		return u
	}
	return nil
}

// Reported by return-bool.
func Lookup(id ID) (*User, bool)
```

```go
func Find(id ID) mo.Option[*User] {
	if u, ok := users[id]; ok {
		return mo.Some(u)
	}
	return mo.None[*User]()
}
```

| Written | Reported | Fix |
| --- | --- | --- |
| `return nil` | Yes | `return mo.None[*T]()` |
| `return nil, nil` with a trailing `error` | Yes | Return an error, or return `mo.Option[*T]` with a nil error |
| `return nil, err` | No | Nothing. The error explains the nil |
| `return nil, false` | Not by `return-nil` | `return-bool` reports the signature. Return `mo.Option` |
| `return nil, true` | Yes | `ok` says the value is there, but it is nil |

For several values before the `bool`, return `mo.Option` of a struct, or of `lo.Tuple2` to `lo.Tuple9` from samber/lo.

> [!WARNING]
> These fixes only hide the absence. Do not use them:
>
> | Fix | Why it is wrong |
> | --- | --- |
> | `return &User{}` in place of `return nil` | An empty user is not "no user". Callers cannot tell them apart |
> | A package-level `var NoUser = &User{}` sentinel | An empty user with a name is still not "no user". Callers must remember to compare with it |
> | Returning `(User, error)` with a `ErrNotFound` only to silence `return-nil` | It turns absence into failure. Do that only when absence is a failure for every caller |

A method that implements an interface is exempt, since the interface fixes its signature. The interface is seen when it is declared in the package, or in a package the package imports directly. When it cannot see it, write an ignore.

### `return-error`

It is off by default. When it is on, return `mo.Result[T]`:

```go
func Load(path string) mo.Result[Config] {
	b, err := os.ReadFile(path)
	if err != nil {
		return mo.Err[Config](err)
	}
	return parse(b)
}
```

`mo.Try` wraps a function that returns `(T, error)`. A method that implements an interface, such as `MarshalJSON`, keeps its signature. When molint cannot see the interface, write an ignore.

### `field-nil-store` and `field-nil-compare`

Both say a pointer field is used as optional. Make it an `mo.Option`:

```go
type User struct {
	Name    string
	Manager mo.Option[*User]
}

u := User{Name: n, Manager: mo.None[*User]()}

if m, ok := u.Manager.Get(); ok {
	notify(m)
}
```

| Situation | Do this |
| --- | --- |
| The field is filled after the struct is built | Compute the value first, then build the struct with it. `field-nil-store` reports the `nil` in the literal |
| The field is set on first use, as in `if c.client == nil { c.client = dial() }` | Use `sync.OnceValue`. Or make the field an `mo.Option` |
| The field points to its own struct, as `Next *Node` | Use `mo.Option[*Node]`. `mo.Option[Node]` is an invalid recursive type |
| The struct is decoded from JSON | `mo.Option` implements `json.Marshaler` and `json.Unmarshaler`. A missing key and `null` both decode as `None` |
| The struct is scanned from a database | `mo.Option` implements `sql.Scanner` and `driver.Valuer` |
| The field comes from a type you do not own, such as protobuf's `*string` | Convert at the boundary with `mo.PointerToOption(p)`, which gives `mo.Option[string]`. Convert back with `o.ToPointer()` |

> [!WARNING]
> With `json:",omitzero"`, an `mo.Option` is left out when it is `None`. It is also left out when it holds a zero value, such as `mo.Some(0)` or `mo.Some("")`, since its `IsZero` reports the value's zero. Do not use `omitzero` where a present zero must be sent. `omitempty` never leaves out a struct, so it has no effect on an `mo.Option`.

`field-nil-store` needs [exhaustruct](https://github.com/GaijinEntertainment/go-exhaustruct). A field left out of a composite literal has no store in SSA, so molint does not see it. exhaustruct makes every field written. Then a left-out pointer becomes `F: nil`, and molint reports it.

Only fields declared in the package count. An embedded field and a field in a generated file do not.

### `wrap-nil`

`mo.Some(nil)` is present and holds nil. That says two things at once.

| Written | Fix |
| --- | --- |
| `mo.Some[*T](nil)` | `mo.None[*T]()` |
| `mo.Some(p)` where `p` may be nil | `mo.PointerToOption(p)` for an `mo.Option[T]`. `mo.EmptyableToOption(p)` keeps `mo.Option[*T]` and gives `None` for nil |
| `mo.Ok[*T](nil)` | Return `mo.Err`, or pass a non-nil value |
| `mo.Err[T](nil)` | Pass a non-nil error. If there is no error, it is `mo.Ok` |

A nil slice is reported too: `mo.Some[[]int](nil)` encodes in JSON as `null`, as `None` does. Use `mo.Some([]int{})`.

### `result-zero`

A zero `mo.Result` is Ok, holding the zero value of its type. That is never what was meant.

```go
var r mo.Result[Config] // reported where r is used
return r
```

Build every `Result` with `mo.Ok`, `mo.Err`, `mo.Errf` or `mo.Try`. A zero `mo.Option` is `None`, which is fine.

A struct field of type `mo.Result` that a composite literal leaves out is a zero `Result` too. molint does not see it. exhaustruct does.

### `unwrap-nil`

`OrEmpty` on an `Option` of a pointer gives nil when it is empty. That is the nil molint exists to stop.

| Written | Fix |
| --- | --- |
| `o.OrEmpty()` on `mo.Option[*T]` | `if v, ok := o.Get(); ok { ... }` |
| `o.OrElse(nil)` | `o.OrElse(def)`, with a `def` that is a real default |
| `o.OrEmpty()` on `mo.Option[int]` or `mo.Option[[]T]` | Nothing. A zero int and a nil slice work, so they are not reported |

`MustGet` panics rather than give nil, so it is not reported. Use it only where absence is a bug, such as right after `IsPresent`.

> [!WARNING]
> `o.OrElse(&T{})` with an empty value made up on the spot satisfies the rule and hides the absence. Give `OrElse` a default the caller would really use, such as a guest user. Otherwise branch on `Get`.

### `unwrap-discard`

`v, _ := o.Get()` drops the one thing `mo.Option` is for.

| Written | Fix |
| --- | --- |
| `v, _ := o.Get(); use(v)` | `if v, ok := o.Get(); ok { use(v) }` |
| `v, _ := r.Get(); use(v)` on a `Result` | `v, err := r.Get(); if err != nil { return ... }` |
| `v, ok := o.Get(); _ = ok; use(v)` | Reported. Assigning `ok` to `_` still discards it. Check `ok` |
| `o.OrElse(def)` | Not reported. The default is chosen in plain sight |

## Reports on code that is correct

molint takes branches as independent. A few shapes are reported although no run gives nil. Change the shape, not the rule:

| Shape | Restructure |
| --- | --- |
| `if c { p = x }; if c { return p }` | Keep the value and its condition together, in an `mo.Option` |
| The pointer and the error set on separate branches, then `if err != nil { return nil, err }; return u, nil` | Return from each branch |
| A retry loop that ends with `return nil, err` | Start the error at a sentinel, so that no round leaves it nil |

## Ignoring a report

```go
//molint:ignore return-bool // callers rely on the comma-ok form of sync.Map
func (c *Cache) Load(key string) (*Entry, bool) {
```

| Rule of thumb | Why |
| --- | --- |
| Name the rules. Do not write a bare `//molint:ignore` | A bare one silences every rule on the line |
| Always write the reason after `//` | Without it, the ignore is reported and silences nothing |
| Remove an ignore once the code no longer needs it | An ignore that silences nothing is reported as unused |

An ignore is right when the signature is fixed by something molint cannot see. An interface in a package that is not imported directly is one. An API that must match another library is another. An ignore is wrong when it only keeps a nil that an `mo.Option` would replace.

A directive on a line of its own silences the line below. One after code silences its own line. For a report on a call, the line is the one with the call's opening parenthesis.

## What molint does not check

| Case | Who checks it |
| --- | --- |
| A nil from a parameter, a field or a call that reaches a dereference | nilaway |
| A field left out of a composite literal | exhaustruct |
| A struct made zero by `var s S`, `new(S)` or a decoder | Nobody. `field-nil-compare` finds such a field where it is checked for nil |
| The initializer of a package-level variable | Nobody. molint does not read it |

A clean molint run does not mean no nil reaches a dereference. Run nilaway beside it.
