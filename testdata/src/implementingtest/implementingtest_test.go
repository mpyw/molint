package implementingtest

import "implstub"

var _ implstub.Other = W{}

// Getter is declared in a test file, so a method outside the test files is
// not exempt by it.
type Getter interface {
	Get() (int, bool) // want `^Getter\.Get reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
}

type V struct{}

// Not reported: a method in a test file sees what the test files see.
func (V) OtherGet() (int, bool) { return 0, false }
