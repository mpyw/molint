// Package mo is a stub of github.com/samber/mo for the analysistest fixtures.
//
// The declarations are copied from github.com/samber/mo v1.17.0 (option.go,
// result.go, utils.go), with their bodies kept as they are, so that the stub
// behaves like the real library. Only what the fixtures use is kept.
//
// Copyright (c) 2022 Samuel Berthe. MIT License:
// https://github.com/samber/mo/blob/v1.17.0/LICENSE
package mo

import (
	"fmt"
	"reflect"
)

var errOptionNoSuchElement = fmt.Errorf("no such element")

// Some builds an Option when value is present.
// Play: https://go.dev/play/p/iqz2n9n0tDM
func Some[T any](value T) Option[T] {
	return Option[T]{
		isPresent: true,
		value:     value,
	}
}

// None builds an Option when value is absent.
// Play: https://go.dev/play/p/yYQPsYCSYlD
func None[T any]() Option[T] {
	return Option[T]{
		isPresent: false,
	}
}

// TupleToOption builds a Some Option when second argument is true, or None.
// Play: https://go.dev/play/p/gkrg2pZwOty
func TupleToOption[T any](value T, ok bool) Option[T] {
	if ok {
		return Some(value)
	}
	return None[T]()
}

// EmptyableToOption builds a Some Option when value is not empty, or None.
// Play: https://go.dev/play/p/GSpQQ-q-UES
func EmptyableToOption[T any](value T) Option[T] {
	// 🤮
	isZero := reflect.ValueOf(&value).Elem().IsZero()
	if isZero {
		return None[T]()
	}

	return Some(value)
}

// Option is a container for an optional value of type T. If value exists, Option is
// of type Some. If the value is absent, Option is of type None.
type Option[T any] struct {
	isPresent bool
	value     T
}

// IsPresent returns false when value is absent.
// Play: https://go.dev/play/p/nDqIaiihyCA
func (o Option[T]) IsPresent() bool {
	return o.isPresent
}

// Get returns value and presence.
// Play: https://go.dev/play/p/0-JBa1usZRT
func (o Option[T]) Get() (T, bool) {
	if !o.isPresent {
		return empty[T](), false
	}

	return o.value, true
}

// MustGet returns value if present or panics instead.
// Play: https://go.dev/play/p/RVBckjdi5WR
func (o Option[T]) MustGet() T {
	if !o.isPresent {
		panic(errOptionNoSuchElement)
	}

	return o.value
}

// OrElse returns value if present or default value.
// Play: https://go.dev/play/p/TrGByFWCzXS
func (o Option[T]) OrElse(fallback T) T {
	if !o.isPresent {
		return fallback
	}

	return o.value
}

// OrEmpty returns value if present or empty value.
// Play: https://go.dev/play/p/SpSUJcE-tQm
func (o Option[T]) OrEmpty() T {
	return o.value
}

// ToPointer returns value if present or a nil pointer.
// Play: https://go.dev/play/p/KJc2Pv3KNPW
func (o Option[T]) ToPointer() *T {
	if !o.isPresent {
		return nil
	}

	return &o.value
}
