package generics

func New[T any]() *T { // want New:"nonnil r0"
	return new(T)
}

func Zero[T any]() *T {
	var p *T
	return p // want `from a nil constant`
}

func First[T any](xs []*T) *T {
	return xs[0] // want `from an element`
}

func Use() *int { // want Use:"nonnil r0"
	return New[int]()
}

func UseZero() *int {
	return Zero[int]() // want `from a call to Zero, which may return nil`
}

type Box[T any] struct{ v T }

func (b *Box[T]) Ptr() *T { // want Ptr:"nonnil r0"
	return &b.v
}

func UseBox(b *Box[string]) *string { // want UseBox:"nonnil r0"
	return b.Ptr()
}

// P is a type parameter, not a pointer, whatever its constraint says.
func Param[P ~*int](p P) P {
	return p
}
