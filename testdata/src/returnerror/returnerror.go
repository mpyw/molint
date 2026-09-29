// Package returnerror covers return-error. It is checked with -return-error
// set.
package returnerror

import (
	"errors"
	"io"
)

type T struct{ N int }

var errX = errors.New("x")

// ===== Functions =====

func Parse(s string) (int, error) { // want `^Parse reports failure with a trailing error; return mo\.Result\[int\] instead \[return-error\]$`
	return len(s), nil
}

func Pointer() (*T, error) { // want `^Pointer reports failure with a trailing error; return mo\.Result\[\*T\] instead \[return-error\]$`
	return &T{}, nil
}

func Two() (string, int, error) { // want `^Two reports failure with a trailing error; return mo\.Result of a struct, or of lo\.Tuple2 instead \[return-error\]$`
	return "", 0, nil
}

// With more than nine values, lo has no tuple to suggest.
func Ten() (int, int, int, int, int, int, int, int, int, int, error) { // want `^Ten reports failure with a trailing error; return mo\.Result of a struct instead \[return-error\]$`
	return 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, nil
}

func BoolError() (bool, error) { // want `^BoolError reports failure with a trailing error; return mo\.Result\[bool\] instead \[return-error\]$`
	return false, nil
}

// An error before the trailing error is a value like any other.
func ErrorError() (error, error) { // want `^ErrorError reports failure with a trailing error; return mo\.Result\[error\] instead \[return-error\]$`
	return nil, nil
}

// return-bool does not report it: the bool is not last.
func ValueBoolError() (int, bool, error) { // want `^ValueBoolError reports failure with a trailing error; return mo\.Result of a struct, or of lo\.Tuple2 instead \[return-error\]$`
	return 0, false, nil
}

// return-error does not report it: the error is not last.
func ValueErrorBool() (int, error, bool) { // want `^ValueErrorBool reports absence with a trailing bool; return mo\.Option of a struct, or of lo\.Tuple2 instead \[return-bool\]$`
	return 0, nil, false
}

// Both rules report the same function, on different lines.
func NilNil() (*T, error) { // want `^NilNil reports failure with a trailing error; return mo\.Result\[\*T\] instead \[return-error\]$`
	return nil, nil // want `^NilNil returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: nothing comes before the error.
func Only() error { return nil }

// Not reported: the error is not last.
func ErrorFirst() (error, int) { return nil, 0 }

// Not reported: only an error identical to the predeclared error counts.
type NamedErr interface{ Error() string }

func Named() (int, NamedErr) { return 0, nil }

// An alias of error is identical to error.
type AliasErr = error

func Aliased() (int, AliasErr) { // want `^Aliased reports failure with a trailing error; return mo\.Result\[int\] instead \[return-error\]$`
	return 0, nil
}

// Not reported: a concrete error type is not error.
type MyErr struct{}

func (MyErr) Error() string { return "" }

func Concrete() (int, *MyErr) { return 0, &MyErr{} }

// ===== Generics =====

func Decode[E any](b []byte) (E, error) { // want `^Decode\[E\] reports failure with a trailing error; return mo\.Result\[E\] instead \[return-error\]$`
	var z E
	return z, nil
}

// ===== Methods, interfaces and function types =====

type S struct{}

func (S) Value() (int, error) { // want `^S\.Value reports failure with a trailing error; return mo\.Result\[int\] instead \[return-error\]$`
	return 0, nil
}

func (*S) Pointer() (int, error) { // want `^\(\*S\)\.Pointer reports failure with a trailing error; return mo\.Result\[int\] instead \[return-error\]$`
	return 0, nil
}

type Store interface {
	Load(string) (*T, error) // want `^Store\.Load reports failure with a trailing error; return mo\.Result\[\*T\] instead \[return-error\]$`
	Close() error
}

type LoadFunc func(string) (*T, error) // want `^LoadFunc reports failure with a trailing error; return mo\.Result\[\*T\] instead \[return-error\]$`

// Not reported: a function literal is exempt.
func Literal() {
	f := func() (int, error) { return 0, errX }
	_ = f
}

// ===== Implementing an interface =====

type Mem struct{}

// Not reported: *Mem implements Store.
func (*Mem) Load(string) (*T, error) { return &T{}, nil }

func (*Mem) Close() error { return nil }

type Wr struct{}

// Not reported: Wr implements io.Writer, and the package imports io.
func (Wr) Write(p []byte) (int, error) { return len(p), nil }

// Not reported: Wr implements io.Reader too.
func (Wr) Read(p []byte) (int, error) { return 0, io.EOF }

type J struct{}

// The package does not import encoding/json, so json.Marshaler does not
// count.
func (J) MarshalJSON() ([]byte, error) { // want `^J\.MarshalJSON reports failure with a trailing error; return mo\.Result\[\[\]byte\] instead \[return-error\]$`
	return []byte("{}"), nil
}
