// Package coverage reaches the less common branches of the analysis.
package coverage

import (
	"errors"
	"unsafe"
)

type T struct{ P *T }

var errX = errors.New("x") // an error variable is no pointer, so it has no fact

func KnownError(p *T) (*T, error) {
	err := errors.New("x")
	if err != nil {
		return p, err
	}
	return p, nil // want `with a nil error, from parameter p`
}

func ErrorLoop(n int) (*T, error) {
	var err error
	for range n {
		if n > 2 {
			err = errX
		}
	}
	return nil, err // want `with a nil error, from a nil constant`
}

func ErrorLoopOK(n int, e error) (*T, error) { // want ErrorLoopOK:"nonnil r0"
	err := e
	for i := range n {
		if i > 2 {
			err = errX
		}
	}
	return nil, err
}

func CheckedNilError(p *T, err error) (*T, error) {
	if err == nil {
		return p, err // want `with a nil error, from parameter p`
	}
	return nil, err
}

func CommaOK(m map[int]*T) (*T, error) { // want CommaOK:"nonnil r0"
	u, _ := m[0]
	return u, errX
}

func OtherCall() (*T, error) { // want OtherCall:"nonnil r0"
	u, _ := Pair()
	_, err := Pair()
	return u, err
}

func Pair() (*T, error) { // want Pair:"nonnil r0"
	return &T{}, nil
}

func Receive(ch chan *T) *T {
	return <-ch // want `from a channel receive`
}

func ReceiveOK(ch chan *T) *T {
	p, _ := <-ch
	return p // want `from a channel receive`
}

func Deref(pp **T) *T {
	return *pp // want `from a pointer load`
}

func Array(a [2]*T) *T {
	return a[0] // want `from an element`
}

func Value(t T) *T {
	return t.P // want `from field P`
}

func MakeT() T { return T{} }

func FieldOfCall() *T {
	return MakeT().P // want `from field P`
}

func MakeArray() [2]*T { return [2]*T{} }

func ElementOfCall() *T {
	return MakeArray()[0] // want `from an element`
}

func CapturedAddr() func() *T {
	var t T
	return func() *T {
		return &t
	}
}

func Convert(p unsafe.Pointer) *T {
	return (*T)(p) // want `from a conversion`
}

func Slice(s []T) *[1]T {
	return (*[1]T)(s) // want `from a conversion`
}

func Local() *T {
	var p *T
	set := func() { p = &T{} }
	set()
	return p // want `from variable p`
}

func Select(a, b chan *T) *T {
	select {
	case p := <-a:
		return p // want `from a value nilproof does not follow`
	case <-b:
		return &T{}
	}
}

func Cycle(n int) *T { // want Cycle:"nonnil r0"
	p := &T{}
	q := &T{}
	for i := range n {
		if i%2 == 0 {
			p = q
		}
	}
	return p
}

func Twice(b bool) *T {
	var p *T
	if b {
		p = &T{}
	}
	if b {
		return p // want `from a nil constant`
	}
	return p // want `from a nil constant`
}

type Named *T

func Change(p *T) Named { // want Change:"nonnil r0"
	if p == nil {
		return Named(&T{})
	}
	return Named(p)
}

type I interface{ M() *T }

func Thunk() func(I) *T {
	return I.M
}

func UseThunk(i I) *T {
	return Thunk()(i) // want `from a call through an interface or a function value`
}

func Wide() (a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q, r, s, t, u, v, w, x, y, z, a1, b1, c1, d1, e1, f1, g1, h1, i1, j1, k1, l1, m1, n1, o1, p1, q1, r1, s1, t1, u1, v1, w1, x1, y1, z1, a2, b2, c2, d2, e2, f2, g2, h2, i2, j2, k2, l2 int, last *T) {
	last = &T{}
	return
}

func UseWide() *T {
	_, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, _, p := Wide()
	return p // want `from a call to Wide, which may return nil`
}
