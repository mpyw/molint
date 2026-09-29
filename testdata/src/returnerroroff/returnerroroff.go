// Package returnerroroff has the same kinds of code as returnerror, checked
// with the default flags: return-error is off, so nothing is reported.
package returnerroroff

import "io"

type T struct{ N int }

func Parse(s string) (int, error) { return len(s), nil }

func Pointer() (*T, error) { return &T{}, nil }

func Two() (string, int, error) { return "", 0, nil }

func BoolError() (bool, error) { return false, nil }

func ErrorError() (error, error) { return nil, nil }

func ValueBoolError() (int, bool, error) { return 0, false, nil }

func Decode[E any](b []byte) (E, error) {
	var z E
	return z, nil
}

type S struct{}

func (S) Value() (int, error) { return 0, nil }

func (*S) Pointer() (int, error) { return 0, nil }

type Store interface {
	Load(string) (*T, error)
}

type LoadFunc func(string) (*T, error)

type J struct{}

func (J) MarshalJSON() ([]byte, error) { return []byte("{}"), nil }

type Wr struct{}

func (Wr) Write(p []byte) (int, error) { return len(p), nil }

var _ io.Writer = Wr{}

// Not reported by return-nil either: the error is not nil.
func NilErr() (*T, error) { return nil, io.EOF }
