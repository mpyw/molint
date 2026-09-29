# molint

Go linter that enforces [samber/mo](https://github.com/samber/mo): absence is [`mo.Option`](https://pkg.go.dev/github.com/samber/mo#Option), not a nil pointer or a trailing `bool`.

> [!WARNING]
> Work in progress. The rules are strict on purpose. They are meant for applications that choose [samber/mo](https://github.com/samber/mo), not for libraries.

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

molint reads the shape of signatures, and follows values only inside one function. It does not look for nil panics. Use it beside [uber-go/nilaway](https://github.com/uber-go/nilaway):

| Tool | Job | Reads |
| --- | --- | --- |
| [uber-go/nilaway](https://github.com/uber-go/nilaway) | Finds possible nil panics | The flow of values across functions |
| molint | Enforces the use of [`mo.Option`](https://pkg.go.dev/github.com/samber/mo#Option) and [`mo.Result`](https://pkg.go.dev/github.com/samber/mo#Result) | The shape of signatures and code |

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
| [`return-nil`](#return-nil) | 🟢 On | A nil pointer result<br>→ Use [`mo.Option`](https://pkg.go.dev/github.com/samber/mo#Option) instead |
| [`return-bool`](#return-bool) | 🟢 On | A signature that ends in a `bool` after other results<br>→ Use [`mo.Option`](https://pkg.go.dev/github.com/samber/mo#Option) instead |
| [`return-error`](#return-error) | 🔴 Off | A signature that ends in an `error` after other results<br>→ Use [`mo.Result`](https://pkg.go.dev/github.com/samber/mo#Result) instead |
| [`wrap-nil`](#wrap-nil) | 🟢 On | nil given to [`mo.Some`](https://pkg.go.dev/github.com/samber/mo#Some), [`mo.Ok`](https://pkg.go.dev/github.com/samber/mo#Ok), or [`mo.Err`](https://pkg.go.dev/github.com/samber/mo#Err)<br>→ Use [`mo.None`](https://pkg.go.dev/github.com/samber/mo#None), or pass a non-nil value |
| [`result-zero`](#result-zero) | 🟢 On | A zero [`mo.Result`](https://pkg.go.dev/github.com/samber/mo#Result)<br>→ Build it with [`mo.Ok`](https://pkg.go.dev/github.com/samber/mo#Ok) or [`mo.Err`](https://pkg.go.dev/github.com/samber/mo#Err) |
| [`unwrap-nil`](#unwrap-nil) | 🟢 On | [`OrEmpty`](https://pkg.go.dev/github.com/samber/mo#Option.OrEmpty), or [`OrElse(nil)`](https://pkg.go.dev/github.com/samber/mo#Option.OrElse), where the nil breaks on use:<br>pointer, interface, `map`, `func` or `chan`<br>→ Use [`Get`](https://pkg.go.dev/github.com/samber/mo#Option.Get) and check `ok`, or give [`OrElse`](https://pkg.go.dev/github.com/samber/mo#Option.OrElse) a non-nil value |
| [`unwrap-discard`](#unwrap-discard) | 🟢 On | [`Get`](https://pkg.go.dev/github.com/samber/mo#Option.Get) with its `ok` or its error discarded<br>→ Check it, or use [`OrElse`](https://pkg.go.dev/github.com/samber/mo#Option.OrElse) |

The line between them: absence must not be dropped silently. A trailing `bool` and a discarded `ok` drop it silently, whatever the type. [`OrEmpty`](https://pkg.go.dev/github.com/samber/mo#Option.OrEmpty) and [`OrElse`](https://pkg.go.dev/github.com/samber/mo#Option.OrElse) choose a default in plain sight. That is fine, unless the default is nil.

Each rule has a flag of its name:

```bash
molint -return-error ./...         # every rule
molint -return-bool=false ./...    # every rule except return-bool and return-error
```

Nothing is reported in a generated file. Test files are checked like any other file.

### `return-nil`

A return must not give a nil pointer. Only pointers are checked: a slice, a map, a func, a channel or an interface may still be returned as nil.

<table>
<thead>
<tr><th>Code</th><th>Valid?</th><th>Reason</th></tr>
</thead>
<tbody>
<tr>
<td>

```go
func F() *T {
	return nil
}
```

</td>
<td>❌</td>
<td>The result is a nil pointer</td>
</tr>
<tr>
<td>

```go
func F() *T {
	var p *T
	return p
}
```

</td>
<td>❌</td>
<td><code>p</code> is never set, so it is nil</td>
</tr>
<tr>
<td>

```go
func F() (*T, error) {
	return nil, nil
}
```

</td>
<td>❌</td>
<td>Both the pointer and the error are nil</td>
</tr>
<tr>
<td>

```go
func F() (*T, error) {
	return nil, ErrNotFound
}
```

</td>
<td>✅</td>
<td>The error is not nil</td>
</tr>
<tr>
<td>

```go
func F() (*T, bool) {
	return nil, false
}
```

</td>
<td>⚠️</td>
<td>Not reported by this rule. <a href="#return-bool"><code>return-bool</code></a> reports the signature</td>
</tr>
<tr>
<td>

```go
func F() (*T, bool) {
	return nil, true
}
```

</td>
<td>❌</td>
<td><code>ok</code> says the value is there, but it is nil</td>
</tr>
<tr>
<td>

```go
func F() (T, bool) {
	return T{}, false
}
```

</td>
<td>⚠️</td>
<td>Not a pointer, so not reported by this rule. <a href="#return-bool"><code>return-bool</code></a> reports the signature</td>
</tr>
<tr>
<td>

```go
func F() (T, error) {
	return T{}, nil
}
```

</td>
<td>✅</td>
<td>Not a pointer, so this rule does not apply</td>
</tr>
<tr>
<td>

```go
func F() []T {
	return nil
}
```

</td>
<td>✅</td>
<td>A nil slice is Go's empty slice, and nothing claims that a value is present</td>
</tr>
</tbody>
</table>

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

#### Fix

```go
func F() mo.Option[*T] {
	if !found {
		return mo.None[*T]()
	}
	return mo.Some(&T{})
}
```

### `return-bool`

A signature must not end in a `bool` after at least one other result.

<table>
<thead>
<tr><th>Signature</th><th>Valid?</th><th>Reason</th></tr>
</thead>
<tbody>
<tr>
<td>

```go
func Find() (User, bool)
```

</td>
<td>❌</td>
<td>Absence is a trailing <code>bool</code>, whatever type comes before it</td>
</tr>
<tr>
<td>

```go
func Cut() (string, string, bool)
```

</td>
<td>❌</td>
<td>The same, with several values</td>
</tr>
<tr>
<td>

```go
func IsAdmin() bool
```

</td>
<td>✅</td>
<td>Nothing comes before the <code>bool</code></td>
</tr>
</tbody>
</table>

Functions, methods, methods of named interfaces, and named function types are checked. A function literal is exempt. So is a method that implements an interface: the interface's own declaration is reported instead, when it is in the package.

#### Fix

```go
func Find() mo.Option[User]
func Cut() mo.Option[lo.Tuple2[string, string]]
```

For several values, use a struct, or a tuple of [samber/lo](https://github.com/samber/lo) such as [`lo.Tuple2`](https://pkg.go.dev/github.com/samber/lo#Tuple2).

> [!NOTE]
> A method is exempt only when molint sees the interface. It must be declared in the package, or in a package that the package imports directly. `MarshalJSON` in a package that does not import `encoding/json` is not exempt.

### `return-error`

The same as [`return-bool`](#return-bool), with `error` in place of `bool`. It is off unless `-return-error` is set.

<table>
<thead>
<tr><th>Signature</th><th>Valid?</th><th>Reason</th></tr>
</thead>
<tbody>
<tr>
<td>

```go
func Find() (User, error)
```

</td>
<td>❌</td>
<td>Failure is a trailing <code>error</code></td>
</tr>
<tr>
<td>

```go
func Load() (Config, Meta, error)
```

</td>
<td>❌</td>
<td>The same, with several values</td>
</tr>
<tr>
<td>

```go
func Close() error
```

</td>
<td>✅</td>
<td>Nothing comes before the <code>error</code></td>
</tr>
</tbody>
</table>

#### Fix

```go
func Find() mo.Result[User]
func Load() mo.Result[lo.Tuple2[Config, Meta]]
```

### `wrap-nil`

<table>
<thead>
<tr><th>Call</th><th>Valid?</th><th>Reason</th></tr>
</thead>
<tbody>
<tr>
<td>

```go
mo.Some[*T](nil)
```

</td>
<td>❌</td>
<td>The option is present and holds nil</td>
</tr>
<tr>
<td>

```go
mo.Ok[*T](nil)
```

</td>
<td>❌</td>
<td>The result is Ok and holds nil</td>
</tr>
<tr>
<td>

```go
mo.Err[T](nil)
```

</td>
<td>❌</td>
<td>The result is an error, and its error is nil</td>
</tr>
<tr>
<td>

```go
mo.Some[map[string]int](nil)
```

</td>
<td>❌</td>
<td>The option is present, and writing to its map panics. The same holds for a func and a channel</td>
</tr>
<tr>
<td>

```go
mo.Some[[]int](nil)
```

</td>
<td>❌</td>
<td><a href="https://pkg.go.dev/github.com/samber/mo#Some"><code>Some</code></a> claims that a value is present, but it holds nil. It then encodes in JSON as <code>null</code>, as <a href="https://pkg.go.dev/github.com/samber/mo#None"><code>mo.None</code></a> does</td>
</tr>
<tr>
<td>

```go
mo.Some([]int{})
```

</td>
<td>✅</td>
<td>An empty slice is not nil, and encodes as <code>[]</code></td>
</tr>
</tbody>
</table>

#### Fix

```go
mo.None[*T]()          // the value is absent
mo.Some(&T{})          // the value is there, and not nil
mo.Err[T](ErrNotFound) // a failure, with a non-nil error
```

### `result-zero`

A zero [`mo.Result`](https://pkg.go.dev/github.com/samber/mo#Result) is Ok, holding the zero value of its type. It must not be used.

```go
func Load() mo.Result[Config] {
	var r mo.Result[Config]
	return r // reported
}
```

A return, an argument, a store, a send, and a method call are uses. A comparison is not. A zero [`mo.Option`](https://pkg.go.dev/github.com/samber/mo#Option) is None, which is fine, so it is not reported.

#### Fix

```go
func Load() mo.Result[Config] {
	cfg, err := read()
	if err != nil {
		return mo.Err[Config](err)
	}
	return mo.Ok(cfg)
}
```

### `unwrap-nil`

<table>
<thead>
<tr><th>Call</th><th>Valid?</th><th>Reason</th></tr>
</thead>
<tbody>
<tr>
<td>

```go
var o mo.Option[*User]
o.OrEmpty()
```

</td>
<td>❌</td>
<td>It gives nil when the option is empty</td>
</tr>
<tr>
<td>

```go
var o mo.Option[*User]
o.OrElse(nil)
```

</td>
<td>❌</td>
<td>The fallback is nil</td>
</tr>
<tr>
<td>

```go
var o mo.Option[*User]
o.OrElse(&guest)
```

</td>
<td>✅</td>
<td>The fallback is not nil</td>
</tr>
<tr>
<td>

```go
var o mo.Option[*User]
o.MustGet()
```

</td>
<td>✅</td>
<td>It panics rather than give nil</td>
</tr>
<tr>
<td>

```go
var n mo.Option[int]
n.OrEmpty()
```

</td>
<td>✅</td>
<td>The zero value works, and is chosen in plain sight</td>
</tr>
<tr>
<td>

```go
var m mo.Option[map[string]int]
m.OrEmpty()
```

</td>
<td>❌</td>
<td>It gives a nil map when the option is empty, and writing to it panics</td>
</tr>
</tbody>
</table>

#### Fix

```go
if u, ok := o.Get(); ok {
	use(u)
}

// Or, with a fallback that is not nil:
u := o.OrElse(&guest)
```

### `unwrap-discard`

<table>
<thead>
<tr><th>Code</th><th>Valid?</th><th>Reason</th></tr>
</thead>
<tbody>
<tr>
<td>

```go
v, _ := o.Get()
use(v)
```

</td>
<td>❌</td>
<td><code>ok</code> is discarded, whatever the type of <code>v</code></td>
</tr>
<tr>
<td>

```go
var r mo.Result[int]
v, _ := r.Get()
use(v)
```

</td>
<td>❌</td>
<td>The error is discarded</td>
</tr>
<tr>
<td>

```go
if v, ok := o.Get(); ok {
	use(v)
}
```

</td>
<td>✅</td>
<td><code>ok</code> is checked</td>
</tr>
<tr>
<td>

```go
o.Get()
```

</td>
<td>✅</td>
<td>Nothing is used</td>
</tr>
</tbody>
</table>

#### Fix

```go
if v, ok := o.Get(); ok {
	use(v)
}

// Or, for a mo.Result:
v, err := r.Get()
if err != nil {
	return err
}

// Or, with a fallback:
v := o.OrElse(10)
```

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
| A nil from a parameter, a field, or a call | Values are followed only inside one function. [uber-go/nilaway](https://github.com/uber-go/nilaway) follows them further |
| [`mo.TupleToOption`](https://pkg.go.dev/github.com/samber/mo#TupleToOption), [`mo.TupleToResult`](https://pkg.go.dev/github.com/samber/mo#TupleToResult), [`mo.EmptyableToOption`](https://pkg.go.dev/github.com/samber/mo#EmptyableToOption) | They check their arguments at run time |
| A zero [`mo.Result`](https://pkg.go.dev/github.com/samber/mo#Result) left out of a composite literal, as in `Holder{}` | Fields are not followed |
| A constructor or method of mo passed as a function value | Calls through function values are not followed |

These are reported although no run returns nil, since branches are taken as independent:

| Case | Instead |
| --- | --- |
| `if c { p = x }; if c { return p }` | Keep the value and its condition together, as in [`mo.Option`](https://pkg.go.dev/github.com/samber/mo#Option) |
| The pointer and the error set on separate branches, then `if err != nil { return nil, err }; return u, nil` | Return from each branch |
| A retry loop that ends with `return nil, err` | Start the error at a sentinel, so that no round leaves it nil |

The full specification, with every limit, is [design/rules.md](design/rules.md).

## License

[MIT](LICENSE)
