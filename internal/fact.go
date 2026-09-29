package internal

import (
	"fmt"
	"strings"
)

// Fact carries what a function is proven to return to the packages that
// import it.
type Fact struct {
	// NonNil lists the pointer results proven non-nil. For a function whose
	// last result is an error, a result is proven non-nil whenever that error
	// is nil.
	NonNil []int
}

// AFact marks Fact as an analysis fact.
func (*Fact) AFact() {}

// String renders the fact the way the tests spell it: "nonnil r0, r2".
func (f *Fact) String() string {
	parts := make([]string, len(f.NonNil))
	for i, r := range f.NonNil {
		parts[i] = fmt.Sprintf("r%d", r)
	}
	return "nonnil " + strings.Join(parts, ", ")
}

// GlobalFact marks a package-level pointer variable proven non-nil: every
// store to it in its own package is proven, and nothing else takes its
// address.
type GlobalFact struct{}

// AFact marks GlobalFact as an analysis fact.
func (*GlobalFact) AFact() {}

// String renders the fact the way the tests spell it.
func (*GlobalFact) String() string { return "nonnil" }
