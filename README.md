# molint

Go linter that forbids returning a pointer it cannot prove to be non-nil.

> [!WARNING]
> Work in progress. This rule is strict on purpose. It is meant for new applications, not for libraries.

## Overview

A nil pointer returned where a value was expected fails far from where it was made. molint does not look for nil. It asks every return of a pointer for a proof that the pointer is not nil. A return without one is reported.

Where a value may be absent, put the absence in the type. Return `mo.Option[*T]` from [samber/mo](https://github.com/samber/mo) instead of a nil pointer.

```go
func Lookup(name string) *User {
	return users[name]
}

func Find(name string) (*User, error) {
	u, ok := users[name]
	if !ok {
		return nil, nil
	}
	return u, nil
}
```

```console
$ molint ./...
user.go:16:2: Lookup may return a nil *User, from a map lookup; return mo.Option[*User] where absence is expected
user.go:22:3: Find may return a nil *User with a nil error, from a nil constant; return an error, or mo.Option[*User], where absence is expected
user.go:24:2: Find may return a nil *User with a nil error, from a map lookup; return an error, or mo.Option[*User], where absence is expected
```

Each report names the value where the proof stopped. That is where the nil may come from.

## Install

| Method | Command | Needs |
| --- | --- | --- |
| `go tool` | `go get -tool github.com/mpyw/molint/cmd/molint@latest` | Go 1.27+ |
| `go install` | `go install github.com/mpyw/molint/cmd/molint@latest` | Go 1.27+ |

```bash
molint ./...          # or: go tool molint ./...
```

> [!TIP]
> On a large module, run it through `go vet`. Every dependency is analyzed too, so that a constructor in another package is proven where it is called. Run on its own, the tool holds all of that in one process. On one application, it peaked at 4.6GB alone and at 0.7GB through `go vet`.
>
> ```bash
> go vet -vettool=$(which molint) ./...
> ```

## What counts as proof

A pointer is proven non-nil when it is one of these:

| Value | Example |
| --- | --- |
| An address | `&User{}`, `new(User)`, `&u.Name`, `&items[i]` |
| A value checked on the way | `if u == nil { return ... }` above the return |
| A call to a proven function | `return NewUser(name)` |
| A proven package-level variable | `var client = &http.Client{}`, never set to anything unproven |
| A variable proven on every path | `u := a; if cond { u = b }` with both `a` and `b` proven |

Anything else is not proven.

| Value | Why |
| --- | --- |
| A parameter or a receiver | The caller may pass nil |
| A field | Its zero value is nil |
| A map lookup | A missing key gives nil |
| A type assertion | A nil pointer asserts fine |
| A call through an interface or a function value | Its body cannot be read |

> [!NOTE]
> Functions are proven across packages, the standard library included. A proven function publishes an analysis fact. Inside a package, functions are proven together, so mutual recursion is proven when its base cases are.

## Functions that return an error

A function whose last result is an `error` may return a nil pointer beside a non-nil error. It must not return a nil pointer with a nil error.

| Return | Reported |
| --- | --- |
| `return nil, err` | No |
| `return nil, ErrNotFound` | No |
| `return nil, nil` | **Yes** |
| `return u, nil` with `u` unproven | **Yes** |
| `return u, err` with both from one call | Only when the callee is not proven |

An error other than the constant nil is taken as non-nil. The rule is there to catch `return nil, nil`. It does not try to prove errors.

A pointer returned with an error is usable only when that error is nil. So a call's pointer result is proven only after its error is checked.

```go
u, err := Find(name)
if err != nil {
	return nil, err
}
return u, nil // proven: Find is proven, and err is checked
```

> [!IMPORTANT]
> A call through an interface has no body to read. When it returns an error, it is trusted to follow the Go convention. Its pointer is taken as non-nil once its error is checked. Without an error result, there is no convention to trust.

## What is not reported

| Case | Why |
| --- | --- |
| A generated file | It is marked `// Code generated ... DO NOT EDIT.` Its functions are still proven for their callers |
| A function that never returns | Whatever it would return is never seen |
| A type parameter, even `P ~*T` | It is not a pointer type |

## Ignoring a report

Write `//molint:ignore` on the reported line or the line above it. A reason after `//` is required.

```go
func proxy(*http.Request) (*url.URL, error) {
	//molint:ignore // net/http reads nil, nil as "no proxy"
	return nil, nil
}
```

| Directive | Result |
| --- | --- |
| `//molint:ignore // reason` | Silences the report |
| `//molint:ignore` | Reported: the reason is missing |
| `//molint:ignore reason` | Reported: text outside `//` is an argument, and it takes none |
| An ignore that silences nothing | Reported as unused |

> [!NOTE]
> An ignore silences the report only. The function is still not proven, so a caller that relies on it is reported in turn.

## Fixing a report

molint offers no automatic fix. Most fixes change a signature, which breaks callers in other packages.

| The nil means | Return instead |
| --- | --- |
| The value may be absent, and that is normal | `mo.Option[*T]` |
| The value is absent, and that is a failure | A non-nil error |
| The value is always there | A proven value: a fresh one, or one checked on the way |

## License

[MIT](LICENSE)
