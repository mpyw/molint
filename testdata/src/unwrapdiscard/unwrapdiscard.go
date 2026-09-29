// Package unwrapdiscard covers unwrap-discard: Get with its second result
// discarded while the value is used.
package unwrapdiscard

import (
	"github.com/samber/mo"
)

type T struct{ N int }

func use(...any)        {}
func pair(int, bool)    {}
func pairErr(*T, error) {}

// ===== Option =====

func Discarded(o mo.Option[int]) {
	v, _ := o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	use(v)
}

func DiscardedPointer(o mo.Option[*T]) {
	v, _ := o.Get() // want `^Get on mo\.Option\[\*T\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	use(v)
}

// Not reported: ok is checked.
func Checked(o mo.Option[int]) {
	v, ok := o.Get()
	if !ok {
		return
	}
	use(v)
}

// Not reported: ok is read, even if it is not checked.
func OkReadNotChecked(o mo.Option[int]) {
	v, ok := o.Get()
	use(v, ok)
}

// Not reported: the value is discarded and ok is used.
func ValueDiscarded(o mo.Option[int]) bool {
	_, ok := o.Get()
	return ok
}

// Not reported: both are discarded; nothing is used.
func BothDiscardedStatement(o mo.Option[int]) {
	o.Get()
}

// Not reported: both are discarded with blanks.
func BothDiscardedBlank(o mo.Option[int]) {
	_, _ = o.Get()
}

// Not reported: both results are passed on.
func PassedOn(o mo.Option[int]) {
	pair(o.Get())
}

// Not reported by unwrap-discard: both results are returned.
func Returned(o mo.Option[int]) (int, bool) { // want `^Returned reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	return o.Get()
}

// ----- The value used in different ways -----

func UsedByReturn(o mo.Option[*T]) *T {
	v, _ := o.Get() // want `^Get on mo\.Option\[\*T\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	return v
}

func UsedByField(o mo.Option[*T]) int {
	v, _ := o.Get() // want `^Get on mo\.Option\[\*T\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	return v.N
}

func UsedByComparison(o mo.Option[*T]) bool {
	v, _ := o.Get() // want `^Get on mo\.Option\[\*T\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	return v != nil
}

func UsedByStore(o mo.Option[int], m map[string]int) {
	v, _ := o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	m["k"] = v
}

func UsedByArithmetic(o mo.Option[int]) int {
	v, _ := o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	return v + 1
}

func UsedInIf(o mo.Option[int]) {
	if v, _ := o.Get(); v > 0 { // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
		use()
	}
}

// Not reported: assigning to the blank identifier gives the value's Extract
// no referrer, so nothing is used.
func UsedOnlyByBlank(o mo.Option[int]) {
	v, _ := o.Get()
	_ = v
}

// ok is assigned to the blank identifier only: its Extract has no referrer,
// so ok is discarded.
func OkOnlyByBlank(o mo.Option[int]) {
	v, ok := o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	_ = ok
	use(v)
}

func AssignedToExisting(o mo.Option[int]) int {
	var v int
	v, _ = o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	return v
}

func InLiteral(o mo.Option[int]) func() int {
	return func() int {
		v, _ := o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
		return v
	}
}

// ----- Receivers and types -----

func ThroughPointer(p *mo.Option[int]) {
	v, _ := p.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	use(v)
}

func Generic[E any](o mo.Option[E]) E {
	v, _ := o.Get() // want `^Get on mo\.Option\[E\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	return v
}

type OptInt = mo.Option[int]

func Alias(o OptInt) {
	v, _ := o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	use(v)
}

// Not reported: OrElse chooses a default in plain sight.
func OrElse(o mo.Option[int]) {
	use(o.OrElse(0))
}

// Not reported: a method value is not followed.
func MethodValue(o mo.Option[int]) {
	get := o.Get
	v, _ := get()
	use(v)
}

// ===== Result =====

func ResultDiscarded(r mo.Result[int]) {
	v, _ := r.Get() // want `^Get on mo\.Result\[int\] discards the error; check it \[unwrap-discard\]$`
	use(v)
}

func ResultDiscardedPointer(r mo.Result[*T]) *T {
	v, _ := r.Get() // want `^Get on mo\.Result\[\*T\] discards the error; check it \[unwrap-discard\]$`
	return v
}

// Not reported: the error is checked.
func ResultChecked(r mo.Result[int]) error {
	v, err := r.Get()
	if err != nil {
		return err
	}
	use(v)
	return nil
}

// Not reported: the value is discarded and the error is used.
func ResultValueDiscarded(r mo.Result[int]) error {
	_, err := r.Get()
	return err
}

// Not reported: both are discarded.
func ResultBothDiscarded(r mo.Result[int]) {
	r.Get()
	_, _ = r.Get()
}

// Not reported: both results are passed on.
func ResultPassedOn(r mo.Result[*T]) {
	pairErr(r.Get())
}

// Reported although IsError was checked first: the rule looks only at Get.
func ResultErrorMethod(r mo.Result[int]) {
	if r.IsError() {
		return
	}
	v, _ := r.Get() // want `^Get on mo\.Result\[int\] discards the error; check it \[unwrap-discard\]$`
	use(v)
}

// Not reported: IsPresent, then MustGet.
func PresentThenMustGet(o mo.Option[int]) {
	if o.IsPresent() {
		use(o.MustGet())
	}
}
