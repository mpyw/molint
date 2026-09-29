// Package resultzero covers result-zero: the kinds of zero mo.Result, every
// use, and what is not a use.
package resultzero

import (
	"errors"

	"github.com/samber/mo"
)

type T struct{ N int }

var errX = errors.New("x")

func compute() mo.Result[int] { return mo.Ok(1) }
func consume(mo.Result[int])  {}
func decode(*mo.Result[int])  {}
func cond() bool              { return true }
func cleanup()                {}
func use(...any)              {}

// ===== The zero constant =====

func ReturnLiteral() mo.Result[int] {
	return mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func ReturnVar() mo.Result[int] {
	var r mo.Result[int]
	return r // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func ReturnPointerResult() mo.Result[*T] {
	return mo.Result[*T]{} // want `^a zero mo\.Result\[\*T\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: the variable is overwritten before its use.
func Overwritten() mo.Result[int] {
	var r mo.Result[int]
	r = mo.Ok(1)
	return r
}

// Not reported: overwritten with a call's result.
func OverwrittenByCall() mo.Result[int] {
	r := mo.Result[int]{}
	r = compute()
	return r
}

// Not reported: built with mo.Ok and mo.Err.
func Built(c bool) mo.Result[int] {
	if c {
		return mo.Ok(1)
	}
	return mo.Err[int](errX)
}

// Not reported: a zero mo.Option is None, which is fine.
func ZeroOption() mo.Option[int] {
	return mo.Option[int]{}
}

// Not reported: the named result is set before the bare return.
func NamedSet() (r mo.Result[int]) {
	r = mo.Ok(1)
	return
}

func NamedBare() (r mo.Result[int]) {
	return // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// ===== φ =====

func PhiIf(c bool) mo.Result[int] {
	var r mo.Result[int]
	if c {
		r = mo.Ok(1)
	}
	return r // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: both edges are built.
func PhiIfElse(c bool) mo.Result[int] {
	var r mo.Result[int]
	if c {
		r = mo.Ok(1)
	} else {
		r = mo.Err[int](errX)
	}
	return r
}

func PhiSwitch(k int) mo.Result[int] {
	var r mo.Result[int]
	switch k {
	case 0:
		r = mo.Ok(0)
	case 1:
		r = mo.Err[int](errX)
	}
	return r // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func PhiLoop(xs []int) mo.Result[int] {
	var r mo.Result[int]
	for _, x := range xs {
		if x > 0 {
			r = mo.Ok(x)
			break
		}
	}
	return r // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: the loop starts from a built value.
func PhiLoopBuilt(xs []int) mo.Result[int] {
	r := mo.Err[int](errX)
	for _, x := range xs {
		if x > 0 {
			r = mo.Ok(x)
			break
		}
	}
	return r
}

// ===== Loads of a local that is not lifted =====

// One report per return: the store the return makes into its result
// variable is not a use.
func DeferZero() mo.Result[int] {
	defer cleanup()
	return mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: with a defer, the stored value is built.
func DeferBuilt() mo.Result[int] {
	defer cleanup()
	return mo.Ok(1)
}

func DeferNamedBare(c bool) (r mo.Result[int]) {
	defer cleanup()
	if c {
		r = mo.Ok(1)
	}
	return // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: a function literal that stores into the variable stops the
// following.
func Captured() mo.Result[int] {
	var r mo.Result[int]
	func() { r = mo.Ok(1) }()
	return r
}

// A return with operands is judged on the value it stores, whatever
// captures the result.
func DeferCaptured() (r mo.Result[int]) {
	defer func() {
		if r.IsOk() {
			r = mo.Err[int](errX)
		}
	}()
	return mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: a bare return follows the stores reaching it, and a
// deferred literal that stores into the result stops that.
func DeferStoringBare() (r mo.Result[int]) {
	defer func() { r = mo.Ok(1) }()
	return
}

// The initializing store and the address: these three are judged by the
// same rule. The store that initializes a local is never a use itself; the
// loads of the local are judged instead. A load is followed unless the
// address is used by something other than a load or a store.

// Not reported: &r is passed to a call, so the load of r is not followed.
func AddressEscapes() mo.Result[int] {
	var r mo.Result[int]
	decode(&r)
	return r
}

// Not reported: the same with an initializing store. That store is not a
// use, and the load is not followed because &r is passed to a call.
func AddressEscapesInitialized() mo.Result[int] {
	r := mo.Result[int]{}
	decode(&r)
	return r
}

// &r is taken, so r is not lifted, but its address is used only by a store
// and a load. The load is followed to the initializing store, which brings
// the zero. The initializing store is not reported; the return is.
func AddressLoadedOnly() mo.Result[int] {
	r := mo.Result[int]{}
	p := &r
	return *p // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// ===== new =====

func NewNotStored() mo.Result[int] {
	p := new(mo.Result[int])
	return *p // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: the value is stored before the load.
func NewStored() mo.Result[int] {
	p := new(mo.Result[int])
	*p = mo.Ok(1)
	return *p
}

// Not reported: the pointer is passed to a call before the load, which may
// store through it.
func NewPassed() mo.Result[int] {
	p := new(mo.Result[int])
	decode(p)
	return *p
}

// ===== Uses =====

func Argument() {
	consume(mo.Result[int]{}) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func ArgumentVar() {
	var r mo.Result[int]
	consume(r) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// A conversion to an interface on the way to an argument is still an
// argument.
func ArgumentInterface() {
	use(mo.Result[int]{}) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func ArgumentDeferred() {
	defer consume(mo.Result[int]{}) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// An implicit conversion has no position of its own, so the report is where
// the converted value is used.
func ConvertedToInterface() {
	var x any = mo.Result[int]{}
	use(x) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: the argument is built.
func ArgumentBuilt() {
	consume(mo.Ok(1))
}

func StoreThroughPointer(dst *mo.Result[int]) {
	*dst = mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

type Holder struct{ R mo.Result[int] }

func StoreField(h *Holder) {
	h.R = mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

var global mo.Result[int]

func StoreGlobal() {
	global = mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func StoreSliceElem(rs []mo.Result[int]) {
	var r mo.Result[int]
	rs[0] = r // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: the stored value is built.
func StoreBuilt(dst *mo.Result[int]) {
	*dst = mo.Err[int](errX)
}

func MethodCall() bool {
	var r mo.Result[int]
	return r.IsOk() // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func MethodCallOnLiteral() int {
	return mo.Result[int]{}.OrElse(1) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func MethodCallGet() error {
	var r mo.Result[int]
	_, err := r.Get() // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
	return err
}

func Send(ch chan<- mo.Result[int]) {
	ch <- mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

func MapUpdate(m map[string]mo.Result[int]) {
	m["k"] = mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: a comparison with the zero constant on the way stops a zero,
// as a nil check does.
func ComparedAndReplaced(c bool) mo.Result[int] {
	var r mo.Result[int]
	if c {
		r = mo.Ok(1)
	}
	if r == (mo.Result[int]{}) {
		r = mo.Err[int](errX)
	}
	return r
}

// The branch where the comparison says zero gives a zero.
func ComparedZero(r mo.Result[int]) mo.Result[int] {
	if r == (mo.Result[int]{}) {
		return r // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
	}
	return r
}

// ===== Not uses =====

// Not reported: a comparison is not a use.
func Compare(r mo.Result[int]) bool {
	return r == mo.Result[int]{}
}

// Not reported: a comparison with a zero variable is not a use.
func CompareVar(r mo.Result[int]) bool {
	var zero mo.Result[int]
	return r != zero
}

// Not reported: an assignment to the blank identifier is not a use.
func Blank() {
	_ = mo.Result[int]{}
}

// ===== Not followed =====

// Not reported: a parameter is not a zero Result.
func Param(r mo.Result[int]) mo.Result[int] {
	return r
}

// Not reported: a call's result is not followed.
func Call() mo.Result[int] {
	return compute()
}

// Not reported: a zero Result held in a field is not followed.
func FieldLoad(h Holder) mo.Result[int] {
	return h.R
}

// Not reported: a field left out of a composite literal is not followed.
func InStruct() Holder {
	return Holder{}
}

// A zero Result written into a field of a composite literal is a store.
func InStructField() Holder {
	return Holder{R: mo.Result[int]{}} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// A zero Result written into an element of a composite literal is a store.
func InSliceElement() []mo.Result[int] {
	return []mo.Result[int]{{}} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// Not reported: the elements are built.
func InSliceBuilt() []mo.Result[int] {
	return []mo.Result[int]{mo.Ok(1)}
}

// ===== Generics and aliases =====

func Generic[E any]() mo.Result[E] {
	return mo.Result[E]{} // want `^a zero mo\.Result\[E\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

type IntResult = mo.Result[int]

func Alias() IntResult {
	return IntResult{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// ===== Function literals =====

// result-zero has no exemption for a function literal.
func InLiteral() func() mo.Result[int] {
	return func() mo.Result[int] {
		return mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
	}
}

// ===== range-over-func =====

func RangeFunc(seq func(func(int) bool)) mo.Result[int] {
	for x := range seq {
		if x > 0 {
			return mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
		}
	}
	return mo.Ok(0)
}

// ===== Other types of mo =====

// Not reported: only mo.Result has a zero value that means Ok.
func ZeroIO() mo.IO[int] {
	return mo.IO[int]{}
}

// Not reported: the body sets one result only on some iterations, and the
// function returns after the loop.
func RangeSomeResults(seq func(func(int) bool)) (p *T, r mo.Result[int]) {
	r = mo.Ok(0)
	for x := range seq {
		if x > 0 {
			r = mo.Ok(x)
		}
		p = &T{}
	}
	return
}

// ===== Stores from function literals =====

// A function literal that stores a zero into a captured variable makes a
// use: the loads of the variable are no longer followed.
func StoredByLiteral() mo.Result[int] {
	r := mo.Ok(1)
	func() {
		r = mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
	}()
	return r
}

// The same from the body of a range-over-func loop.
func StoredByRangeBody(seq func(func(int) bool)) mo.Result[int] {
	r := mo.Ok(1)
	for x := range seq {
		if x > 0 {
			r = mo.Result[int]{} // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
		}
	}
	return r
}

// An implicit conversion that flows through a φ is reported where the
// converted value is used.
func ConvertedThroughPhi(c bool) {
	var x any
	if c {
		x = mo.Result[int]{}
	} else {
		x = 1
	}
	use(x) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
}

// An explicit conversion is reported where it is written.
func ConvertedExplicitly() {
	x := any(mo.Result[int]{}) // want `^a zero mo\.Result\[int\] is Ok with a zero value; build it with mo\.Ok or mo\.Err \[result-zero\]$`
	use(x)
}
