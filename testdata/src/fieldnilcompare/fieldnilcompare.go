// Package fieldnilcompare covers field-nil-compare: a pointer field compared
// with nil.
package fieldnilcompare

import (
	"fieldnilother"
)

type T struct{ N int }

type S struct {
	A int
	P *T
}

func get() *T    { return &T{} }
func cond() bool { return true }

// ===== Comparisons =====

func Equal(s *S) bool {
	return s.P == nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

func NotEqual(s S) bool {
	return s.P != nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

func NilFirst(s *S) bool {
	return nil == s.P // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

func TypedNil(s *S) bool {
	return s.P == (*T)(nil) // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

func newS() S { return S{} }

func ValueField() bool {
	return newS().P == nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

func Switch(s *S) int {
	switch s.P {
	case nil: // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
		return 0
	}
	return 1
}

func InCondition(s *S) int {
	if s.P != nil && s.P.N > 0 { // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
		return s.P.N
	}
	return 0
}

// A local variable that SSA lifts is the field read itself.
func Lifted(s *S) bool {
	p := s.P
	return p == nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

type Lazy struct {
	client *T
}

// A field set on first use is reported too.
func (l *Lazy) Client() *T {
	if l.client == nil { // want `^Lazy\.client is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
		l.client = get()
	}
	return l.client
}

type Ptr *T

type Named struct {
	P Ptr
}

func NamedPointer(n Named) bool {
	return n.P == nil // want `^Named\.P is compared with nil; make it mo\.Option\[Ptr\] \[field-nil-compare\]$`
}

// A conversion between pointer types is seen through.
func Converted(s *S) bool {
	return Ptr(s.P) == nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

type Box[E any] struct {
	V *E
}

func GenericBody[E any](b Box[E]) bool {
	return b.V == nil // want `^Box\[E\]\.V is compared with nil; make it mo\.Option\[\*E\] \[field-nil-compare\]$`
}

func Anonymous(v struct{ P *T }) bool {
	return v.P == nil // want `^P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

// Not reported: the value is a φ, not the field read.
func Phi(s *S, x *T) bool {
	p := s.P
	if cond() {
		p = x
	}
	return p == nil
}

// Not reported: the field is compared with something other than nil.
func NotNil(s *S, x *T) bool {
	return s.P == x
}

// Not reported: a comparison of two fields reads no nil.
func TwoFields(a, b *S) bool {
	return a.P == b.P
}

// Not reported: a conversion to an interface is not the field read.
func Interface(s *S) bool {
	return any(s.P) == nil
}

// ===== Which fields count =====

type Double struct {
	PP **T
}

// A pointer to a pointer is a pointer.
func DoublePointer(d Double) bool {
	return d.PP == nil // want `^Double\.PP is compared with nil; make it mo\.Option\[\*\*T\] \[field-nil-compare\]$`
}

// Not reported: what the field points to is compared, not the field.
func ThroughDoublePointer(d Double) bool {
	return *d.PP == nil
}

func DoublePointerToStruct(pp **S) bool {
	return (*pp).P == nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

type EmbedsPointer struct {
	*S
}

// A field promoted through an embedded pointer is the field of S.
func PromotedThroughPointer(e EmbedsPointer) bool {
	return e.P == nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

type PS *S

func NamedPointerToStruct(p PS) bool {
	return p.P == nil // want `^S\.P is compared with nil; make it mo\.Option\[\*T\] \[field-nil-compare\]$`
}

// Not reported: an embedded field is not looked at.
func Embedded(e EmbedsPointer) bool {
	return e.S == nil
}

type NotPointer struct {
	M map[string]int
	F func()
	I any
	X []int
}

// Not reported: only pointer fields are looked at.
func OtherNilable(n NotPointer) bool {
	return n.M == nil || n.F == nil || n.I == nil || n.X == nil
}

// Not reported: the struct is declared in another package.
func OtherPackage(o fieldnilother.Other) bool {
	return o.P == nil
}

// Not reported: the struct is declared in a generated file.
func GeneratedStruct(g Gen) bool {
	return g.P == nil
}

// ===== Directives =====

func Ignored(s *S) bool {
	return s.P == nil //molint:ignore field-nil-compare // checked on purpose
}
