// Package implementingtest covers a method whose interface only a test file
// sees. The package is checked alone and with its tests, and both must agree.
package implementingtest

type W struct{}

// W implements implstub.Other, but only the test file imports implstub. So
// the method is not exempt, alone or with the tests.
func (W) OtherGet() (int, bool) { return 0, false } // want `^W\.OtherGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

type U struct{}

// U implements Getter, which only the test file declares. It is not exempt
// either.
func (U) Get() (int, bool) { return 0, false } // want `^U\.Get reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
