package mo

import (
	"fmt"
)

// Ok builds a Result when value is valid.
// Play: https://go.dev/play/p/PDwADdzNoyZ
func Ok[T any](value T) Result[T] {
	return Result[T]{
		value: value,
		isErr: false,
	}
}

// Err builds a Result when value is invalid.
// Play: https://go.dev/play/p/PDwADdzNoyZ
func Err[T any](err error) Result[T] {
	return Result[T]{
		err:   err,
		isErr: true,
	}
}

// Errf builds a Result when value is invalid.
// Errf formats according to a format specifier and returns the error as a value that satisfies Result[T].
// Play: https://go.dev/play/p/N43w92SM-Bs
func Errf[T any](format string, a ...any) Result[T] {
	return Err[T](fmt.Errorf(format, a...))
}

// TupleToResult convert a pair of T and error into a Result.
// Play: https://go.dev/play/p/KWjfqQDHQwa
func TupleToResult[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}
	return Ok(value)
}

// Result represents a result of an action having one
// of the following output: success or failure.
// An instance of Result is an instance of either Ok or Err.
// It could be compared to `Either[error, T]`.
type Result[T any] struct {
	isErr bool
	value T
	err   error
}

// IsOk returns true when value is valid.
// Play: https://go.dev/play/p/sfNvBQyZfgU
func (r Result[T]) IsOk() bool {
	return !r.isErr
}

// IsError returns true when value is invalid.
// Play: https://go.dev/play/p/xkV9d464scV
func (r Result[T]) IsError() bool {
	return r.isErr
}

// Error returns error when value is invalid or nil.
// Play: https://go.dev/play/p/CSkHGTyiXJ5
func (r Result[T]) Error() error {
	return r.err
}

// Get returns value and error.
// Play: https://go.dev/play/p/8KyX3z6TuNo
func (r Result[T]) Get() (T, error) {
	if r.isErr {
		return empty[T](), r.err
	}

	return r.value, nil
}

// MustGet returns value when Result is valid or panics.
// Play: https://go.dev/play/p/8LSlndHoTAE
func (r Result[T]) MustGet() T {
	if r.isErr {
		panic(r.err)
	}

	return r.value
}

// OrElse returns value when Result is valid or default value.
// Play: https://go.dev/play/p/MN_ULx0soi6
func (r Result[T]) OrElse(fallback T) T {
	if r.isErr {
		return fallback
	}

	return r.value
}

// OrEmpty returns value when Result is valid or empty value.
// Play: https://go.dev/play/p/rdKtBmOcMLh
func (r Result[T]) OrEmpty() T {
	return r.value
}
