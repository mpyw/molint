// Package returnnilpair covers return-nil in functions whose last result is
// an error or a bool, and the pairing of φ edges.
package returnnilpair

import (
	"errors"
	"fmt"
)

type T struct{ N int }

var errX = errors.New("x")

func get() *T             { return &T{} }
func getErr() (*T, error) { return &T{}, nil }
func mayFail() error      { return nil }
func cond() bool          { return true }
func log()                {}
func cleanup()            {}

// ===== Last result is error =====

func NilNil() (*T, error) {
	return nil, nil // want `^NilNil returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: the error is not nil.
func NilErr() (*T, error) {
	return nil, errX
}

// Not reported: the error is a call's result.
func NilNewErr() (*T, error) {
	return nil, fmt.Errorf("x: %d", 1)
}

// Not reported: the error is a call's result, even if it may be nil.
func NilCallErr() (*T, error) {
	return nil, mayFail()
}

// Not reported: the error is a parameter.
func NilParamErr(err error) (*T, error) {
	return nil, err
}

// Not reported: the pointer is not nil.
func ValueNil() (*T, error) {
	return &T{}, nil
}

func NilVarErr() (*T, error) {
	var err error
	return nil, err // want `^NilVarErr returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: the common check returns the error on its non-nil edge.
func CheckedErr() (*T, error) {
	err := mayFail()
	if err != nil {
		return nil, err
	}
	return &T{}, nil
}

func CheckedErrWrongWay() (*T, error) {
	err := mayFail()
	if err == nil {
		return nil, err // want `^CheckedErrWrongWay returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
	}
	return &T{}, nil
}

type MyErr struct{}

func (*MyErr) Error() string { return "my" }

// Not reported: a nil *MyErr in an error is a non-nil interface; it is not a
// nil value.
func NilTypedErr() (*T, error) {
	var e *MyErr
	return nil, e
}

// Not reported: return f() passes another call's results.
func ReturnCall() (*T, error) {
	return getErr()
}

// An alias of error is identical to error.
type AliasErr = error

func NilAliasErr() (*T, AliasErr) {
	return nil, nil // want `^NilAliasErr returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// A named interface is not identical to error: the function has no trailing
// error, so the nil pointer is always reported.
type NamedErr interface{ Error() string }

func NilNamedErr() (*T, NamedErr) {
	return nil, errX // want `^NilNamedErr returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// An error that is not last does not count.
func ErrorNotLast() (*T, error, int) {
	return nil, errX, 0 // want `^ErrorNotLast returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

func TwoPointersNilErr() (*T, *T, error) {
	return &T{}, nil, nil // want `^TwoPointersNilErr returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: with a non-nil error, neither nil pointer is reported.
func TwoPointersErr() (*T, *T, error) {
	return nil, nil, errX
}

// ===== φ pairing with error =====

// Not reported: each edge brings a non-nil pointer or a non-nil error.
func PairIfElse(b bool) (*T, error) {
	var u *T
	var err error
	if b {
		u = get()
	} else {
		err = errX
	}
	return u, err
}

func PairMissingErr(b bool) (*T, error) {
	var u *T
	var err error
	if b {
		u = get()
	} else {
		log()
	}
	return u, err // want `^PairMissingErr returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

func PairCrossed(b bool) (*T, error) {
	var u *T
	var err error
	if b {
		u = get()
		err = errX
	}
	return u, err // want `^PairCrossed returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: a switch whose every edge pairs correctly.
func PairSwitch(k int) (*T, error) {
	var u *T
	var err error
	switch k {
	case 0:
		u = get()
	case 1:
		u = &T{}
	default:
		err = errX
	}
	return u, err
}

func PairSwitchMissing(k int) (*T, error) {
	var u *T
	var err error
	switch k {
	case 0:
		u = get()
	case 1:
		err = errX
	}
	return u, err // want `^PairSwitchMissing returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: nested ifs with empty joins merge into one block of φs,
// and each edge pairs correctly.
func PairNested(a, b bool) (*T, error) {
	var u *T
	var err error
	if a {
		if b {
			u = get()
		} else {
			err = errX
		}
	} else {
		err = errX
	}
	return u, err
}

// Not reported: the inner join is not empty, so its φs stay in their own
// block; pairing recurses into them along the same edge.
func PairNestedNonEmptyJoin(a, b bool) (*T, error) {
	var u *T
	var err error
	if a {
		if b {
			u = get()
		} else {
			err = errX
		}
		log()
	} else {
		err = errX
	}
	return u, err
}

func PairNestedNonEmptyJoinMissing(a, b bool) (*T, error) {
	var u *T
	var err error
	if a {
		if b {
			u = get()
		}
		log()
	} else {
		err = errX
	}
	return u, err // want `^PairNestedNonEmptyJoinMissing returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: a loop that ends with either a value or an error.
func PairLoop(xs []*T) (*T, error) {
	var u *T
	var err error
	for _, x := range xs {
		if x != nil {
			u = x
			break
		}
	}
	if u == nil {
		err = errX
	}
	return u, err
}

// Not reported: named results set on each path before a bare return.
func PairNamed(b bool) (u *T, err error) {
	if b {
		u = get()
	} else {
		err = errX
	}
	return
}

func PairNamedBare() (u *T, err error) {
	return // want `^PairNamedBare returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: two bare returns, each with one of the two set.
func PairNamedTwoReturns(b bool) (u *T, err error) {
	if b {
		u = get()
		return
	}
	err = errX
	return
}

// Not reported: with a defer, the stores on the way to the return pair up
// along each path.
func PairDefer(b bool) (u *T, err error) {
	defer cleanup()
	if b {
		u = get()
	} else {
		err = errX
	}
	return
}

func PairDeferNil() (*T, error) {
	defer cleanup()
	return nil, nil // want `^PairDeferNil returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: with a defer, a nil pointer with an error.
func PairDeferErr() (*T, error) {
	defer cleanup()
	return nil, errX
}

// A return with operands is judged on the values it stores, whatever
// captures the results.
func PairDeferCaptured() (u *T, err error) {
	defer func() {
		if err == nil && u == nil {
			err = errX
		}
	}()
	return nil, nil // want `^PairDeferCaptured returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// The recover pattern: the deferred literal stores into err, but the return
// has operands.
func PairRecover() (u *T, err error) {
	defer func() {
		if recover() != nil {
			err = errX
		}
	}()
	return nil, nil // want `^PairRecover returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// A deferred literal that only loads err does not stop a bare return from
// following the stores: on the path where c is false, both stay nil.
func PairDeferLogBare(c bool) (u *T, err error) {
	defer func() {
		if err != nil {
			log()
		}
	}()
	if c {
		u = get()
	}
	return // want `^PairDeferLogBare returns a nil \*T with a nil error; return an error, or mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported: the same literal, with each path pairing correctly.
func PairDeferLogBarePaired(c bool) (u *T, err error) {
	defer func() {
		if err != nil {
			log()
		}
	}()
	if c {
		u = get()
	} else {
		err = errX
	}
	return
}

// Not reported: a deferred literal that stores into err stops a bare return
// from following it.
func PairDeferStoringBare() (u *T, err error) {
	defer func() {
		if u == nil {
			err = errX
		}
	}()
	return
}

// ===== Last result is bool =====

func NilTrue() (*T, bool) { // want `^NilTrue reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	return nil, true // want `^NilTrue returns a nil \*T with ok true; return mo\.Option\[\*T\] \[return-nil\]$`
}

// Not reported by return-nil: return-bool reports the signature.
func NilFalse() (*T, bool) { // want `^NilFalse reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	return nil, false
}

// Not reported by return-nil: the bool is not the constant true.
func NilParamBool(ok bool) (*T, bool) { // want `^NilParamBool reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	return nil, ok
}

// Not reported by return-nil: the bool is a call's result.
func NilCallBool() (*T, bool) { // want `^NilCallBool reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	return nil, cond()
}

// Not reported by return-nil: each edge pairs nil with false.
func PairBool(b bool) (*T, bool) { // want `^PairBool reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	var u *T
	ok := false
	if b {
		u = get()
		ok = true
	}
	return u, ok
}

func PairBoolAlwaysTrue(b bool) (*T, bool) { // want `^PairBoolAlwaysTrue reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	var u *T
	ok := true
	if b {
		u = get()
	}
	return u, ok // want `^PairBoolAlwaysTrue returns a nil \*T with ok true; return mo\.Option\[\*T\] \[return-nil\]$`
}

func PairBoolCrossed(b bool) (*T, bool) { // want `^PairBoolCrossed reports absence with a trailing bool; return mo\.Option\[\*T\] instead \[return-bool\]$`
	var u *T
	ok := true
	if b {
		u = get()
		ok = false
	}
	return u, ok // want `^PairBoolCrossed returns a nil \*T with ok true; return mo\.Option\[\*T\] \[return-nil\]$`
}

// A named bool is not bool: the function has no trailing bool, so the nil
// pointer is always reported, and return-bool does not report it.
type MyBool bool

func NilMyBool() (*T, MyBool) {
	return nil, false // want `^NilMyBool returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}
