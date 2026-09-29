// Package generated has one generated file and files that are not.
package generated

type T struct{ N int }

// The same code outside the generated file is reported.
func UserNil() *T {
	return nil // want `^UserNil returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

func UserBool() (int, bool) { // want `^UserBool reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	return 0, false
}

type U struct{}

// Not reported: U implements GenIface, declared in the generated file.
func (U) Find() (*T, bool) { return &T{}, true }
