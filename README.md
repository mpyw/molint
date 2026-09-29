# nilproof

Go linter that forbids returning a pointer it cannot prove to be non-nil.

Where a value may be absent, return `mo.Option[*T]` from [samber/mo](https://github.com/samber/mo) instead of a nil pointer.

> [!WARNING]
> Work in progress. This rule is strict on purpose. It is meant for new applications, not for libraries.
