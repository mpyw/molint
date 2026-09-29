// Package returnnil covers return-nil in functions without a trailing error
// or bool: the kinds of nil value, and how a check or a φ edge stops one.
package returnnil

import (
	"bytes"
	"iter"
	"unsafe"
)

type T struct{ N int }

func ok(*T) bool   { return true }
func get() *T      { return &T{} }
func next(p *T) *T { return p }
func cond() bool   { return true }
func log()         {}
func use(...any)   {}

// ===== The constant nil =====

func Nil() *T {
	return nil // want `^Nil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: an address is not nil.
func Addr() *T {
	return &T{}
}

func NilPtrPtr() **T {
	return nil // want `^NilPtrPtr returns a nil \*\*T; return mo\.Option\[\*\*T\] instead \[return-nil\]$`
}

// A type from another package is named by its package name.
func NilImported() *bytes.Buffer {
	return nil // want `^NilImported returns a nil \*bytes\.Buffer; return mo\.Option\[\*bytes\.Buffer\] instead \[return-nil\]$`
}

func VarNeverSet() *T {
	var p *T
	return p // want `^VarNeverSet returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the variable is set before the return.
func VarSet() *T {
	var p *T
	p = get()
	return p
}

func TypedNil() *T {
	return (*T)(nil) // want `^TypedNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: a parameter is not a nil value.
func Param(p *T) *T {
	return p
}

// Not reported: a call's result is not followed.
func Call() *T {
	return get()
}

// Not reported: a field is not followed.
func Field(s struct{ P *T }) *T {
	return s.P
}

// Not reported: a map lookup is not followed.
func MapLookup(m map[string]*T) *T {
	return m["k"]
}

var global *T

// Not reported: a global is not followed.
func Global() *T {
	return global
}

// ===== Types that are not pointers =====

// Not reported: unsafe.Pointer's underlying type is not a pointer.
func UnsafePointer() unsafe.Pointer {
	return nil
}

// Not reported: an interface is not a pointer.
func Interface() any {
	return nil
}

// Not reported: an error is not a pointer.
func Error() error {
	return nil
}

// Not reported: a slice, a map, a func and a chan are not pointers.
func Slice() []int     { return nil }
func Map() map[int]int { return nil }
func Func() func()     { return nil }
func Chan() chan int   { return nil }

// ===== Named pointer types and ChangeType =====

type P *T

type Q *T

func NamedPointer() P {
	return nil // want `^NamedPointer returns a nil P; return mo\.Option\[P\] instead \[return-nil\]$`
}

func ChangeTypeOfChecked(p *T) P {
	if p == nil {
		return P(p) // want `^ChangeTypeOfChecked returns a nil P; return mo\.Option\[P\] instead \[return-nil\]$`
	}
	return P(p)
}

func ChangeTypeNamed(q Q) P {
	if q == nil {
		return P(q) // want `^ChangeTypeNamed returns a nil P; return mo\.Option\[P\] instead \[return-nil\]$`
	}
	return &T{}
}

// Not reported: a conversion of a value that is not nil.
func ChangeTypeOfParam(q Q) P {
	return P(q)
}

// ===== A dominating nil check =====

func CheckedNil(p *T) *T {
	if p == nil {
		return p // want `^CheckedNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	}
	return p
}

// Not reported: the branch checks the value to be non-nil.
func CheckedNonNil(p *T) *T {
	if p != nil {
		return p
	}
	return &T{}
}

func CheckedNilReversed(p *T) *T {
	if nil == p {
		return p // want `^CheckedNilReversed returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	}
	return p
}

func CheckedNotEqualElse(p *T) *T {
	if p != nil {
		return p
	} else {
		return p // want `^CheckedNotEqualElse returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	}
}

func CheckedNegated(p *T) *T {
	if !(p != nil) {
		return p // want `^CheckedNegated returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	}
	return p
}

func CheckedViaVariable(p *T) *T {
	isNil := p == nil
	if isNil {
		return p // want `^CheckedViaVariable returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	}
	return p
}

func CheckedAnd(p *T, c bool) *T {
	if p == nil && c {
		return p // want `^CheckedAnd returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	}
	return &T{}
}

// Not reported: with ||, the branch does not dominate on p == nil.
func CheckedOr(p *T, c bool) *T {
	if p == nil || c {
		return p
	}
	return &T{}
}

func CheckedSwitch(p *T) *T {
	switch {
	case p == nil:
		return p // want `^CheckedSwitch returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	default:
		return p
	}
}

// Not reported: the check is on another value.
func CheckedOther(p, q *T) *T {
	if q == nil {
		return p
	}
	return q
}

// ===== φ from if and else =====

func PhiIf(c bool) *T {
	var p *T
	if c {
		p = get()
	}
	return p // want `^PhiIf returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: both edges bring a value that is not nil.
func PhiIfElse(c bool) *T {
	var p *T
	if c {
		p = get()
	} else {
		p = &T{}
	}
	return p
}

func PhiElseNil(c bool) *T {
	p := get()
	if c {
		p = nil
	}
	return p // want `^PhiElseNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

func PhiReturnInBranch(c bool) *T {
	if c {
		return get()
	}
	return nil // want `^PhiReturnInBranch returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: a nil check on the edge replaces the nil with a default.
func PhiDefaulted(c bool, def *T) *T {
	var p *T
	if c {
		p = get()
	}
	if p == nil {
		p = def
	}
	return p
}

// Not reported: the same, with a parameter.
func ParamDefaulted(p, def *T) *T {
	if p == nil {
		p = def
	}
	return p
}

func PhiCheckedButNotReplaced(c bool) *T {
	var p *T
	if c {
		p = get()
	}
	if p == nil {
		log()
	}
	return p // want `^PhiCheckedButNotReplaced returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the return is dominated by the != check.
func PhiCheckedNonNil(c bool) *T {
	var p *T
	if c {
		p = get()
	}
	if p != nil {
		return p
	}
	return &T{}
}

// Not reported: the return is on the false edge of == nil.
func PhiEarlyReturn(c bool) *T {
	var p *T
	if c {
		p = get()
	}
	if p == nil {
		return &T{}
	}
	return p
}

func PhiCheckedNil(c bool) *T {
	var p *T
	if c {
		p = get()
	}
	if p == nil {
		return p // want `^PhiCheckedNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
	}
	return p
}

// Not reported: nested ifs whose edges all bring a value that is not nil.
func PhiNested(a, b bool) *T {
	var p *T
	if a {
		if b {
			p = get()
		} else {
			p = &T{}
		}
	} else {
		p = get()
	}
	return p
}

func PhiNestedMissing(a, b bool) *T {
	var p *T
	if a {
		if b {
			p = get()
		}
	} else {
		p = get()
	}
	return p // want `^PhiNestedMissing returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// ===== switch =====

func SwitchNoDefault(k int) *T {
	var p *T
	switch k {
	case 1:
		p = get()
	case 2:
		p = &T{}
	}
	return p // want `^SwitchNoDefault returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: every case, including default, sets the value.
func SwitchDefault(k int) *T {
	var p *T
	switch k {
	case 1:
		p = get()
	default:
		p = &T{}
	}
	return p
}

func SwitchCaseNil(k int) *T {
	p := get()
	switch k {
	case 1:
		p = nil
	}
	return p // want `^SwitchCaseNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: a type switch binds a value from a type assertion.
func TypeSwitch(x any) *T {
	switch v := x.(type) {
	case *T:
		return v
	}
	return &T{}
}

// ===== Loops =====

func RangeBreak(xs []*T) *T {
	var found *T
	for _, x := range xs {
		if ok(x) {
			found = x
			break
		}
	}
	return found // want `^RangeBreak returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

func RangeFallThrough(xs []*T) *T {
	for _, x := range xs {
		if ok(x) {
			return x
		}
	}
	return nil // want `^RangeFallThrough returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the value is set before the loop and each round.
func ForReassign(n int) *T {
	p := &T{}
	for i := 0; i < n; i++ {
		p = next(p)
	}
	return p
}

func ForSetsNil() *T {
	p := &T{}
	for cond() {
		p = nil
	}
	return p // want `^ForSetsNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the loop exits only on the false edge of p == nil.
func ForUntilNonNil() *T {
	var p *T
	for p == nil {
		p = get()
	}
	return p
}

func ForUntilNil() *T {
	p := get()
	for p != nil {
		p = next(p)
	}
	return p // want `^ForUntilNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// ===== goto and labeled break =====

func Goto(c bool) *T {
	var p *T
	if c {
		goto done
	}
	p = get()
done:
	return p // want `^Goto returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the goto skips only a reassignment of a value that is not nil.
func GotoDefault(c bool) *T {
	p := &T{}
	if c {
		goto done
	}
	p = get()
done:
	return p
}

func LabeledBreak(rows [][]*T) *T {
	var p *T
outer:
	for _, row := range rows {
		for _, x := range row {
			if ok(x) {
				p = x
				break outer
			}
		}
	}
	return p // want `^LabeledBreak returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the value starts non-nil.
func LabeledBreakDefault(rows [][]*T) *T {
	p := &T{}
outer:
	for _, row := range rows {
		for _, x := range row {
			if ok(x) {
				p = x
				break outer
			}
		}
	}
	return p
}

// ===== No return =====

// Not reported: the function panics.
func Panics() *T {
	panic("never")
}

// Not reported: the function loops forever.
func Forever() *T {
	for {
	}
}

// Not reported: the only nil path panics.
func PanicsOnNil(c bool) *T {
	var p *T
	if c {
		p = get()
	}
	if p == nil {
		panic("nil")
	}
	return p
}

// ===== Several results =====

func SecondPointer() (int, *T) {
	return 0, nil // want `^SecondPointer returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// One report per nil pointer result.
func TwoPointers() (*T, *bytes.Buffer) {
	return nil, nil // want `^TwoPointers returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$` `^TwoPointers returns a nil \*bytes\.Buffer; return mo\.Option\[\*bytes\.Buffer\] instead \[return-nil\]$`
}

func TwoPointersOneNil() (*T, *T) {
	return &T{}, nil // want `^TwoPointersOneNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: neither pointer is nil.
func TwoPointersNoNil() (*T, *T) {
	return &T{}, get()
}

// ===== return f() =====

func pair() (int, *T) { return 0, &T{} }

// Not reported: the results of another call are not followed.
func ReturnCall() (int, *T) {
	return pair()
}

// ===== Named results and bare returns =====

func NamedBare() (p *T) {
	return // want `^NamedBare returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the named result is set before the bare return.
func NamedBareSet() (p *T) {
	p = &T{}
	return
}

func NamedBarePhi(c bool) (p *T) {
	if c {
		p = get()
	}
	return // want `^NamedBarePhi returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

func NamedExplicitNil() (p *T) {
	p = &T{}
	return nil // want `^NamedExplicitNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// ===== defer =====

func cleanup() {}

func DeferNil() *T {
	defer cleanup()
	return nil // want `^DeferNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the value stored on the way to the return is not nil.
func DeferNonNil() *T {
	defer cleanup()
	return &T{}
}

func DeferNamedBare(c bool) (p *T) {
	defer cleanup()
	if c {
		p = get()
	}
	return // want `^DeferNamedBare returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: every store on the way to the return is not nil.
func DeferNamedSet(c bool) (p *T) {
	defer cleanup()
	if c {
		p = get()
	} else {
		p = &T{}
	}
	return
}

// A return with operands is judged on the values it stores, whatever
// captures the result.
func DeferCapturedResult() (p *T) {
	defer func() {
		if p == nil {
			p = &T{}
		}
	}()
	return nil // want `^DeferCapturedResult returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: a bare return follows the stores reaching it, and a
// deferred literal that stores into the result stops that.
func DeferStoringLiteralBare() (p *T) {
	defer func() { p = &T{} }()
	return
}

// A literal that captures the result and only loads it does not stop the
// following.
func ReadingLiteralResult() (p *T) {
	f := func() { use(p) }
	f()
	return nil // want `^ReadingLiteralResult returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

func ReadingLiteralResultBare() (p *T) {
	f := func() { use(p) }
	f()
	return // want `^ReadingLiteralResultBare returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the result is stored with a value that is not nil, and a
// nil stored before is overwritten on the way.
func DeferOverwritten() (p *T) {
	defer cleanup()
	p = nil
	p = get()
	return
}

// ===== Captures and escapes =====

// Not reported: a function literal that stores into the variable stops the
// following.
func Captured() *T {
	var p *T
	set := func() { p = &T{} }
	set()
	return p
}

// A literal that only reads the variable does not stop the following. The
// variable is never stored, so its load gives the zero of its allocation.
func CapturedRead() *T {
	var p *T
	func() { use(p) }()
	return p // want `^CapturedRead returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// The same with an explicit nil store.
func CapturedReadStored() *T {
	p := (*T)(nil)
	func() { use(p) }()
	return p // want `^CapturedReadStored returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the literal only reads the variable, and the store that
// reaches the load is not nil.
func CapturedReadSet() *T {
	var p *T
	p = &T{}
	func() { use(p) }()
	return p
}

var sink **T

// Not reported: a variable whose address is stored anywhere is not followed.
func AddressStored() *T {
	var p *T
	sink = &p
	return p
}

func fill(pp **T) { *pp = &T{} }

// Not reported: a variable whose address is passed to a call is not followed.
func AddressEscapes() *T {
	var p *T
	fill(&p)
	return p
}

// ===== Function literals =====

// Not reported: a function literal is exempt.
func Literal() func() *T {
	return func() *T {
		return nil
	}
}

// Not reported: a function literal is exempt, even when called at once.
func LiteralCalled() *T {
	return func() *T {
		if cond() {
			return nil
		}
		return &T{}
	}()
}

// Not reported: a literal inside a variable.
var LiteralVar = func() *T { return nil }

// ===== Generics =====

func Generic[E any]() *E {
	return nil // want `^Generic\[E\] returns a nil \*E; return mo\.Option\[\*E\] instead \[return-nil\]$`
}

// Not reported: the result is a type parameter, which is not a pointer.
func GenericZero[E any]() E {
	var z E
	return z
}

// Not reported: a type parameter is not a pointer, whatever its constraint.
func GenericPointerConstraint[E ~*int]() E {
	return nil
}

// Not reported: the generic pointer result is not nil.
func GenericNew[E any]() *E {
	return new(E)
}

type Box[E any] struct{ v *E }

func (b *Box[E]) Get() *E {
	return nil // want `^\(\*Box\[E\]\)\.Get returns a nil \*E; return mo\.Option\[\*E\] instead \[return-nil\]$`
}

// Not reported: a field of the receiver is not followed.
func (b Box[E]) Field() *E {
	return b.v
}

// ===== Methods =====

type S struct{ p *T }

func (s S) Value() *T {
	return nil // want `^S\.Value returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

func (s *S) Pointer() *T {
	return nil // want `^\(\*S\)\.Pointer returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: the method returns a field.
func (s *S) FieldValue() *T {
	return s.p
}

// Not reported: the receiver itself is not a nil value.
func (s *S) Self() *S {
	return s
}

func (s *S) SelfChecked() *S {
	if s == nil {
		return s // want `^\(\*S\)\.SelfChecked returns a nil \*S; return mo\.Option\[\*S\] instead \[return-nil\]$`
	}
	return s
}

// ===== Aliases =====

type PT = *T

// The alias is resolved in the message.
func Alias() PT {
	return nil // want `^Alias returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// ===== range-over-func =====

// A return inside the body is judged as a return of the enclosing function.
func RangeFunc(seq iter.Seq[*T]) *T {
	for x := range seq {
		if ok(x) {
			return x
		}
		if cond() {
			return nil // want `^RangeFunc returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
		}
	}
	return nil // want `^RangeFunc returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: every return gives a value that is not nil.
func RangeFuncNonNil(seq iter.Seq[*T]) *T {
	for x := range seq {
		if x != nil {
			return x
		}
	}
	return &T{}
}
