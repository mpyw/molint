// Package unwrapnil covers unwrap-nil: OrEmpty, and OrElse with a nil value,
// on an Option or a Result of a pointer.
package unwrapnil

import (
	"github.com/samber/mo"
)

type T struct{ N int }

func get() *T    { return &T{} }
func use(...any) {}

// ===== OrEmpty =====

func OrEmpty(o mo.Option[*T]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

// OrEmpty is reported always, even on an Option that is present.
func OrEmptySome() {
	use(mo.Some(&T{}).OrEmpty()) // want `^OrEmpty on mo\.Option\[\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

func OrEmptyResult(r mo.Result[*T]) {
	use(r.OrEmpty()) // want `^OrEmpty on mo\.Result\[\*T\] gives nil when it is an error; use Get and check the error, or OrElse with a non-nil value \[unwrap-nil\]$`
}

func OrEmptyPtrPtr(o mo.Option[**T]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[\*\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

// Not reported: OrEmpty of a value type chooses the zero value in plain
// sight.
func OrEmptyInt(o mo.Option[int]) {
	use(o.OrEmpty())
}

// A nil interface, map, func or channel breaks on use too.
func OrEmptyInterface(o mo.Option[error]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[error\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

func OrEmptyMap(o mo.Option[map[string]int]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[map\[string\]int\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

func OrEmptyFunc(o mo.Option[func()]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[func\(\)\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

func OrEmptyChan(r mo.Result[chan int]) {
	use(r.OrEmpty()) // want `^OrEmpty on mo\.Result\[chan int\] gives nil when it is an error; use Get and check the error, or OrElse with a non-nil value \[unwrap-nil\]$`
}

// Not reported: a nil slice works as an empty one.
func OrEmptySlice(o mo.Option[[]int]) {
	use(o.OrEmpty())
}

// Not reported: MustGet panics rather than give nil.
func MustGet(o mo.Option[*T]) {
	use(o.MustGet())
}

// Not reported: MustGet on a Result.
func MustGetResult(r mo.Result[*T]) {
	use(r.MustGet())
}

// Not reported: ToPointer is not one of the calls, although it gives nil
// when the Option is empty.
func ToPointer(o mo.Option[*T]) {
	use(o.ToPointer())
}

// ----- Named pointer types -----

type P *T

func OrEmptyNamedPointer(o mo.Option[P]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[P\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

// ----- Generics -----

func OrEmptyGeneric[E any](o mo.Option[*E]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[\*E\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

// Not reported: a type parameter is not a pointer, whatever its constraint.
func OrEmptyTypeParam[E ~*int](o mo.Option[E]) {
	use(o.OrEmpty())
}

// ----- Calls that are not followed -----

// Not reported: a method value is not followed.
func MethodValue(o mo.Option[*T]) {
	f := o.OrEmpty
	use(f())
}

// Not reported: a method expression is called through a thunk.
func MethodExpression(o mo.Option[*T]) {
	use(mo.Option[*T].OrElse(o, nil))
	use(mo.Option[*T].OrEmpty(o))
}

// ----- Aliases -----

type PT = *T

// The alias is resolved in the message.
func OrEmptyAliasPointer(o mo.Option[PT]) {
	use(o.OrEmpty()) // want `^OrEmpty on mo\.Option\[\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

// ===== OrElse =====

func OrElseNil(o mo.Option[*T]) {
	use(o.OrElse(nil)) // want `^OrElse\(nil\) on mo\.Option\[\*T\] gives nil when it is empty; pass a non-nil value \[unwrap-nil\]$`
}

func OrElseNilResult(r mo.Result[*T]) {
	use(r.OrElse(nil)) // want `^OrElse\(nil\) on mo\.Result\[\*T\] gives nil when it is an error; pass a non-nil value \[unwrap-nil\]$`
}

func OrElseVar(o mo.Option[*T]) {
	var def *T
	use(o.OrElse(def)) // want `^OrElse\(nil\) on mo\.Option\[\*T\] gives nil when it is empty; pass a non-nil value \[unwrap-nil\]$`
}

func OrElsePhi(o mo.Option[*T], c bool) {
	var def *T
	if c {
		def = get()
	}
	use(o.OrElse(def)) // want `^OrElse\(nil\) on mo\.Option\[\*T\] gives nil when it is empty; pass a non-nil value \[unwrap-nil\]$`
}

func OrElseChecked(o mo.Option[*T], def *T) {
	if def == nil {
		use(o.OrElse(def)) // want `^OrElse\(nil\) on mo\.Option\[\*T\] gives nil when it is empty; pass a non-nil value \[unwrap-nil\]$`
	}
}

// Not reported: the default is not nil.
func OrElseAddr(o mo.Option[*T]) {
	use(o.OrElse(&T{}))
}

// Not reported: a parameter is not a nil value.
func OrElseParam(o mo.Option[*T], def *T) {
	use(o.OrElse(def))
}

// Not reported: a call's result is not followed.
func OrElseCall(o mo.Option[*T]) {
	use(o.OrElse(get()))
}

// Not reported: the nil edge is replaced by a default.
func OrElseDefaulted(o mo.Option[*T], c bool) {
	var def *T
	if c {
		def = get()
	}
	if def == nil {
		def = &T{}
	}
	use(o.OrElse(def))
}

func OrElseNilInterface(o mo.Option[error]) {
	use(o.OrElse(nil)) // want `^OrElse\(nil\) on mo\.Option\[error\] gives nil when it is empty; pass a non-nil value \[unwrap-nil\]$`
}

func OrElseNilMap(o mo.Option[map[string]int]) {
	use(o.OrElse(nil)) // want `^OrElse\(nil\) on mo\.Option\[map\[string\]int\] gives nil when it is empty; pass a non-nil value \[unwrap-nil\]$`
}

// Not reported: a made map is not nil.
func OrElseMadeMap(o mo.Option[map[string]int]) {
	use(o.OrElse(map[string]int{}))
}

// Not reported: a nil slice works as an empty one.
func OrElseNilSlice(o mo.Option[[]int]) {
	use(o.OrElse(nil))
}

func OrElseGeneric[E any](o mo.Option[*E]) {
	use(o.OrElse(nil)) // want `^OrElse\(nil\) on mo\.Option\[\*E\] gives nil when it is empty; pass a non-nil value \[unwrap-nil\]$`
}

// ===== Receivers reached in other ways =====

func ThroughPointer(p *mo.Option[*T]) {
	use(p.OrEmpty()) // want `^OrEmpty on mo\.Option\[\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

type Wrap struct{ mo.Option[*T] }

func Promoted(w Wrap) {
	use(w.OrEmpty()) // want `^OrEmpty on mo\.Option\[\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
}

// ===== Function literals =====

// unwrap-nil has no exemption for a function literal.
func InLiteral(o mo.Option[*T]) func() *T {
	return func() *T {
		return o.OrEmpty() // want `^OrEmpty on mo\.Option\[\*T\] gives nil when it is empty; use Get and check ok, or OrElse with a non-nil value \[unwrap-nil\]$`
	}
}

type S struct{}

func (S) OrEmpty() *T { return &T{} }

// Not reported: a method of another type with the same name.
func OtherOrEmpty(s S) {
	use(s.OrEmpty())
}
