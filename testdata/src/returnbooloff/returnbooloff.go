// Package returnbooloff is checked with -return-bool=false: return-bool
// reports nothing, and the other rules are unchanged.
package returnbooloff

import "github.com/samber/mo"

type T struct{ N int }

// Not reported: return-bool is off, and an ignore whose every rule is off by
// flag is not reported as unused.
//
//molint:ignore return-bool // the comma-ok form is intended
func Lookup(m map[string]int, k string) (int, bool) {
	v, ok := m[k]
	return v, ok
}

// Not reported: return-bool is off.
type Finder interface {
	Find(string) (*T, bool)
}

// Not reported: return-bool is off.
type FindFunc func(string) (*T, bool)

// return-nil is still on, and its rule on ok true does not change.
func NilTrue() (*T, bool) {
	return nil, true // want `^NilTrue returns a nil \*T with ok true; return mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: return-nil leaves nil with false to return-bool, even when
// return-bool is off.
func NilFalse() (*T, bool) {
	return nil, false
}

func Nil() *T {
	return nil // want `^Nil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// The rules of mo are still on.
func Discard(o mo.Option[int]) int {
	v, _ := o.Get() // want `^Get on mo\.Option\[int\] discards ok; check it, or use OrElse \[unwrap-discard\]$`
	return v
}

// Not reported: return-error is still off.
func Parse(s string) (int, error) { return len(s), nil }
