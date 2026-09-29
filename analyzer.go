// Package nilproof reports a returned pointer that is not proven non-nil.
//
// A nil pointer returned where a value was expected fails far from where it
// was made. nilproof asks every return of a pointer for a proof that it is
// not nil. Where a value may be absent, return mo.Option from
// github.com/samber/mo instead, so that the absence is in the type.
package nilproof

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"

	"github.com/mpyw/nilproof/internal"
)

// Analyzer reports a returned pointer that is not proven non-nil.
var Analyzer = &analysis.Analyzer{
	Name:      "nilproof",
	Doc:       "reports a returned pointer that is not proven non-nil",
	URL:       "https://github.com/mpyw/nilproof",
	Requires:  []*analysis.Analyzer{buildssa.Analyzer},
	FactTypes: []analysis.Fact{new(internal.Fact), new(internal.GlobalFact)},
	Run:       internal.Run,
}

// ErrNoSSA is returned when the pass carries no buildssa result, which means
// the analyzer was registered without its requirement.
var ErrNoSSA = internal.ErrRunWithoutSSA
