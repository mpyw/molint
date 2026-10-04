// Package fieldnilstore covers field-nil-store: a nil value stored into a
// pointer field.
package fieldnilstore

import (
	"fieldnilother"

	"github.com/samber/mo"
)

type T struct{ N int }

type S struct {
	A int
	P *T
}

func get() *T    { return &T{} }
func cond() bool { return true }

// ===== Composite literals =====

func Keyed() S {
	return S{A: 1, P: nil} // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

func KeyedAddress() *S {
	return &S{A: 1, P: nil} // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

func Positional() S {
	return S{1, nil} // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

func TypedNil() S {
	return S{A: 1, P: (*T)(nil)} // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

// Not reported: a field left out of a composite literal has no store.
// exhaustruct reports it instead.
func LeftOut() S {
	return S{A: 1}
}

// Not reported: an address is not nil.
func KeyedValue() S {
	return S{A: 1, P: &T{}}
}

// ===== Assignments =====

func Assign(s *S) {
	s.P = nil // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

func AssignLocal() S {
	var s S
	s.P = nil // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
	return s
}

type Outer struct {
	Inner S
}

func AssignNested(o *Outer) {
	o.Inner.P = nil // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

type Embeds struct {
	S
}

func AssignPromoted(e *Embeds) {
	e.P = nil // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

func AssignElement(xs []S) {
	xs[0].P = nil // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

var global S

func AssignGlobal() {
	global.P = nil // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

// Not reported: a call's result is not a nil value.
func AssignCall(s *S) {
	s.P = get()
}

// Not reported: a parameter is not a nil value.
func AssignParam(s *S, p *T) {
	s.P = p
}

// ===== Nil values =====

func PhiNil(s *S, x *T) {
	var p *T
	if cond() {
		p = x
	}
	s.P = p // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

func CheckedNil(s *S, p *T) {
	if p == nil {
		s.P = p // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
	}
}

// Not reported: the nil check replaces the nil on the way.
func CheckedDefault(s *S, p *T) {
	if p == nil {
		p = &T{}
	}
	s.P = p
}

// Not reported: the store is on the path where p is not nil.
func CheckedNonNil(s *S) {
	var p *T
	if cond() {
		p = get()
	}
	if p != nil {
		s.P = p
	}
}

// ===== Which fields count =====

type Node struct {
	Next *Node
}

// A field that points back to its own struct is not exempt.
func Recursive(n *Node) {
	n.Next = nil // want `^Node\.Next is set to nil; make it mo\.Option\[\*Node\] \[field-nil-store\]$`
}

type Box[E any] struct {
	V *E
}

func GenericInstance() Box[int] {
	return Box[int]{V: nil} // want `^Box\[E\]\.V is set to nil; make it mo\.Option\[\*E\] \[field-nil-store\]$`
}

func GenericBody[E any](b *Box[E]) {
	b.V = nil // want `^Box\[E\]\.V is set to nil; make it mo\.Option\[\*E\] \[field-nil-store\]$`
}

func Anonymous() any {
	return struct{ P *T }{P: nil} // want `^P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

type Ptr *T

type Named struct {
	P Ptr
}

func NamedPointer(n *Named) {
	n.P = nil // want `^Named\.P is set to nil; make it mo\.Option\[Ptr\] \[field-nil-store\]$`
}

func InFuncLit() func(*S) {
	return func(s *S) {
		s.P = nil // want `^S\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
	}
}

func LocalType() any {
	type L struct{ P *T }
	return L{P: nil} // want `^L\.P is set to nil; make it mo\.Option\[\*T\] \[field-nil-store\]$`
}

type EmbedsPointer struct {
	*S
}

// Not reported: an embedded field is not looked at.
func Embedded() EmbedsPointer {
	return EmbedsPointer{S: nil}
}

type NotPointer struct {
	M map[string]int
	F func()
	I any
	X []int
}

// Not reported: only pointer fields are looked at.
func OtherNilable() NotPointer {
	return NotPointer{M: nil, F: nil, I: nil, X: nil}
}

type Generic[P any] struct {
	X P
}

// Not reported: a type parameter is not a pointer, whatever its argument.
func TypeParamField() Generic[*T] {
	return Generic[*T]{X: nil}
}

// Not reported: the struct is declared in another package.
func OtherPackage() fieldnilother.Other {
	return fieldnilother.Other{P: nil}
}

// Not reported: the struct is declared in a generated file.
func GeneratedStruct() Gen {
	return Gen{P: nil}
}

type Holder struct {
	O mo.Option[*T]
}

// Not reported: an Option is not a pointer.
func OptionField() Holder {
	return Holder{O: mo.None[*T]()}
}

// ===== Package initialization =====

// Not reported: the initializer of a package-level variable runs in the
// package's synthetic init function, which no rule analyzes.
var initialized = S{P: nil}

// ===== Directives =====

func Ignored(s *S) {
	s.P = nil //molint:ignore field-nil-store // cleared on purpose
}
