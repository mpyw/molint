package returnbool

import "testing"

// A test file is checked like any other file.
func helper() (string, bool) { // want `^helper reports absence with a trailing bool; return mo\.Option\[string\] instead \[return-bool\]$`
	return "", false
}

// Not reported: a test function has no results.
func TestHelper(t *testing.T) {
	if _, ok := helper(); ok {
		t.Fatal("ok")
	}
}
