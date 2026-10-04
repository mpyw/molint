// Package fieldniloff has the same kinds of code as fieldnilstore and
// fieldnilcompare, checked with the default flags: both rules are off, so
// nothing is reported.
package fieldniloff

type T struct{ N int }

type S struct {
	A int
	P *T
}

func Keyed() S {
	return S{A: 1, P: nil}
}

func Assign(s *S) {
	s.P = nil
}

func Equal(s *S) bool {
	return s.P == nil
}
