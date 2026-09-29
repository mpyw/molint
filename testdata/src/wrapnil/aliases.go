package wrapnil

import (
	m "github.com/samber/mo"
)

// ===== An import name and type aliases =====

type OptT = m.Option[*T]

type Opt[E any] = m.Option[E]

func RenamedImport() OptT {
	return m.Some[*T](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func GenericAlias() Opt[*T] {
	var o Opt[*T] = m.Some[*T](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
	return o
}

// Not reported: an alias does not change what is given.
func GenericAliasValue() Opt[*T] {
	return m.Some(&T{})
}
