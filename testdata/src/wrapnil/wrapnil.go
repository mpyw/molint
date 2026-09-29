// Package wrapnil covers wrap-nil: a nil value given to mo.Some, mo.Ok or
// mo.Err.
package wrapnil

import (
	"errors"

	"github.com/samber/mo"
)

type T struct{ N int }

var errX = errors.New("x")

func get() *T       { return &T{} }
func getErr() error { return nil }
func cond() bool    { return true }
func use(...any)    {}

// ===== mo.Some =====

func SomeNil() mo.Option[*T] {
	return mo.Some[*T](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeTypedNilInferred() mo.Option[*T] {
	return mo.Some((*T)(nil)) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeVarNeverSet() mo.Option[*T] {
	var p *T
	return mo.Some(p) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

// Not reported: an address is not nil.
func SomeAddr() mo.Option[*T] {
	return mo.Some(&T{})
}

// Not reported: a parameter is not a nil value.
func SomeParam(p *T) mo.Option[*T] {
	return mo.Some(p)
}

// Not reported: a call's result is not followed.
func SomeCall() mo.Option[*T] {
	return mo.Some(get())
}

func SomeChecked(p *T) mo.Option[*T] {
	if p == nil {
		return mo.Some(p) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
	}
	return mo.Some(p)
}

// Not reported: the branch checks the value to be non-nil.
func SomeCheckedNonNil(p *T) mo.Option[*T] {
	if p != nil {
		return mo.Some(p)
	}
	return mo.None[*T]()
}

func SomePhi(c bool) mo.Option[*T] {
	var p *T
	if c {
		p = get()
	}
	return mo.Some(p) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

// Not reported: the nil edge is replaced by a default.
func SomePhiDefaulted(c bool) mo.Option[*T] {
	var p *T
	if c {
		p = get()
	}
	if p == nil {
		p = &T{}
	}
	return mo.Some(p)
}

// Not reported: a variable a function literal captures is not followed.
func SomeCaptured() mo.Option[*T] {
	var p *T
	func() { p = get() }()
	return mo.Some(p)
}

// The call is reported where it is, not only in a return.
func SomeStatement() {
	o := mo.Some[*T](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
	use(o)
}

// Not reported: nothing is given to mo.None.
func NoneGiven() mo.Option[*T] {
	return mo.None[*T]()
}

// ----- Interfaces -----

func SomeInterfaceNil() mo.Option[error] {
	return mo.Some[error](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeAnyNil() mo.Option[any] {
	return mo.Some[any](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeInterfaceVar() mo.Option[error] {
	var err error
	return mo.Some(err) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

// Not reported: an interface that holds a nil *T is not a nil value.
func SomeInterfaceTypedNil() mo.Option[any] {
	var p *T
	return mo.Some[any](p)
}

// Not reported: a non-nil interface.
func SomeInterfaceValue() mo.Option[error] {
	return mo.Some(errX)
}

// ----- Maps, funcs, channels and slices -----

// A nil map, func or channel breaks on use: writing to the map, calling the
// func, or using the channel.
func SomeMapNil() mo.Option[map[int]int] {
	return mo.Some[map[int]int](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeFuncNil() mo.Option[func()] {
	return mo.Some[func()](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeChanNil() mo.Option[chan int] {
	return mo.Some[chan int](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeMapVar() mo.Option[map[string]*T] {
	var m map[string]*T
	return mo.Some(m) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

// Not reported: a made map is not nil.
func SomeMapMade() mo.Option[map[int]int] {
	return mo.Some(make(map[int]int))
}

// Not reported: a nil slice works as an empty one.
func SomeSliceNil() mo.Option[[]int] { return mo.Some[[]int](nil) }

// Not reported: a value type.
func SomeInt() mo.Option[int] {
	return mo.Some(0)
}

// ----- Named pointer types -----

type P *T

func SomeNamedPointer() mo.Option[P] {
	return mo.Some[P](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

func SomeChangeType(p *T) mo.Option[P] {
	if p == nil {
		return mo.Some(P(p)) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
	}
	return mo.Some(P(p))
}

// ----- Generics -----

func SomeGeneric[E any]() mo.Option[*E] {
	return mo.Some[*E](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

// Not reported: a type parameter value is not followed as a pointer.
func SomeGenericValue[E any](v E) mo.Option[E] {
	return mo.Some(v)
}

// Not reported: a type parameter counts as none of the types a nil breaks,
// whatever its constraint.
func SomeGenericPointerConstraint[E ~*int]() mo.Option[E] {
	return mo.Some[E](nil)
}

// Not reported: the same for a type parameter constrained by any.
func SomeGenericZero[E any]() mo.Option[E] {
	var z E
	return mo.Some(z)
}

// ----- Calls that are not followed -----

// A constructor held in a local variable lowers to a static call.
func SomeFuncValue() mo.Option[*T] {
	f := mo.Some[*T]
	return f(nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

// Not reported: a constructor passed through a parameter is not followed.
func SomeThroughParam(f func(*T) mo.Option[*T]) mo.Option[*T] {
	return f(nil)
}

// Not reported: the caller of SomeThroughParam is not followed into it.
func CallsThroughParam() mo.Option[*T] {
	return SomeThroughParam(mo.Some[*T])
}

type ctors struct {
	some func(*T) mo.Option[*T]
}

// Not reported: a constructor held in a field is not followed.
func SomeThroughField() mo.Option[*T] {
	c := ctors{some: mo.Some[*T]}
	return c.some(nil)
}

// Not reported: these check their arguments at run time.
func Tuple() mo.Option[*T]     { return mo.TupleToOption[*T](nil, true) }
func Emptyable() mo.Option[*T] { return mo.EmptyableToOption[*T](nil) }

// ===== mo.Ok =====

func OkNil() mo.Result[*T] {
	return mo.Ok[*T](nil) // want `^mo\.Ok is given nil; pass a non-nil value \[wrap-nil\]$`
}

func OkVar() mo.Result[*T] {
	var p *T
	return mo.Ok(p) // want `^mo\.Ok is given nil; pass a non-nil value \[wrap-nil\]$`
}

func OkInterfaceNil() mo.Result[any] {
	return mo.Ok[any](nil) // want `^mo\.Ok is given nil; pass a non-nil value \[wrap-nil\]$`
}

func OkPhi(c bool) mo.Result[*T] {
	var p *T
	if c {
		p = get()
	}
	return mo.Ok(p) // want `^mo\.Ok is given nil; pass a non-nil value \[wrap-nil\]$`
}

// Not reported: an address is not nil.
func OkAddr() mo.Result[*T] {
	return mo.Ok(&T{})
}

// Not reported: a nil slice works as an empty one.
func OkSliceNil() mo.Result[[]int] {
	return mo.Ok[[]int](nil)
}

func OkMapNil() mo.Result[map[int]int] {
	return mo.Ok[map[int]int](nil) // want `^mo\.Ok is given nil; pass a non-nil value \[wrap-nil\]$`
}

// Not reported: a checked value on its non-nil edge.
func OkChecked(p *T) mo.Result[*T] {
	if p != nil {
		return mo.Ok(p)
	}
	return mo.Err[*T](errX)
}

// Not reported: these check their arguments at run time.
func TupleResult() mo.Result[*T] { return mo.TupleToResult[*T](nil, nil) }

// ===== mo.Err =====

func ErrNil() mo.Result[int] {
	return mo.Err[int](nil) // want `^mo\.Err is given a nil error; pass a non-nil error \[wrap-nil\]$`
}

func ErrVar() mo.Result[int] {
	var err error
	return mo.Err[int](err) // want `^mo\.Err is given a nil error; pass a non-nil error \[wrap-nil\]$`
}

// The type of the value does not matter for mo.Err.
func ErrNilPointerResult() mo.Result[*T] {
	return mo.Err[*T](nil) // want `^mo\.Err is given a nil error; pass a non-nil error \[wrap-nil\]$`
}

func ErrCheckedWrongWay() mo.Result[int] {
	err := getErr()
	if err == nil {
		return mo.Err[int](err) // want `^mo\.Err is given a nil error; pass a non-nil error \[wrap-nil\]$`
	}
	return mo.Ok(1)
}

// Not reported: the common check gives the error on its non-nil edge.
func ErrChecked() mo.Result[int] {
	if err := getErr(); err != nil {
		return mo.Err[int](err)
	}
	return mo.Ok(1)
}

// Not reported: a call's result is not followed, even if it may be nil.
func ErrCall() mo.Result[int] {
	return mo.Err[int](getErr())
}

// Not reported: a non-nil error.
func ErrValue() mo.Result[int] {
	return mo.Err[int](errX)
}

type MyErr struct{}

func (*MyErr) Error() string { return "my" }

// Not reported: an error that holds a nil *MyErr is not a nil value.
func ErrTypedNil() mo.Result[int] {
	var e *MyErr
	return mo.Err[int](e)
}

func ErrPhi(c bool) mo.Result[int] {
	var err error
	if c {
		err = errX
	}
	return mo.Err[int](err) // want `^mo\.Err is given a nil error; pass a non-nil error \[wrap-nil\]$`
}

// Not reported: Errf builds its error.
func ErrfBuilds() mo.Result[int] {
	return mo.Errf[int]("x %d", 1)
}

// A constructor held in a local variable lowers to a static call.
func ErrFuncValue() mo.Result[int] {
	f := mo.Err[int]
	return f(nil) // want `^mo\.Err is given a nil error; pass a non-nil error \[wrap-nil\]$`
}

// ===== Function literals =====

// wrap-nil has no exemption for a function literal.
func InLiteral() func() mo.Option[*T] {
	return func() mo.Option[*T] {
		return mo.Some[*T](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
	}
}

// A variable of the function, given to mo.Some inside a range-over-func
// body, holds what reached the loop.
func SomeInRangeBody(seq func(func(int) bool)) mo.Option[*T] {
	var p *T
	for x := range seq {
		if x == 0 {
			return mo.Some(p) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
		}
	}
	return mo.None[*T]()
}
