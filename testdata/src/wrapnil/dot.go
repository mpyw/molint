package wrapnil

import (
	. "github.com/samber/mo"
)

// ===== A dot import =====

func DotSome() Option[*T] {
	return Some[*T](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func DotOk() Result[*T] {
	return Ok[*T](nil) // want `^mo\.Ok is given nil; pass a non-nil value \[wrap-nil\]$`
}

// Not reported: a non-nil value through a dot import.
func DotSomeValue() Option[*T] {
	return Some(&T{})
}
