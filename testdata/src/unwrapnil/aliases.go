package unwrapnil

import (
	m "github.com/samber/mo"
)

// ===== An import name and type aliases =====

type OptT = m.Option[*T]

type Res[E any] = m.Result[E]

func Alias(o OptT) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

func GenericAlias(r Res[*T]) {
	use(r.OrElse(nil)) // want `^OrElse\(nil\) on mo\.Result\[\*T\] gives nil when it is an error; pass a non-nil value \[unwrap-nil\]$`
}

// Not reported: a generic alias of a value type.
func GenericAliasInt(r Res[int]) {
	use(r.OrEmpty())
}
