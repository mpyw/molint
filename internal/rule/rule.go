// Package rule names the rules of molint.
//
// A rule's name is shared by its flag, its diagnostics, and the directives
// that silence it, so it is spelled once, here.
package rule

// Name is the name of a rule.
type Name string

const (
	// ReturnNil reports a nil pointer result.
	ReturnNil Name = "return-nil"
	// ReturnBool reports a signature that ends in a bool after other results.
	ReturnBool Name = "return-bool"
	// ReturnError reports a signature that ends in an error after other
	// results.
	ReturnError Name = "return-error"
	// WrapNil reports nil given to mo.Some, mo.Ok or mo.Err.
	WrapNil Name = "wrap-nil"
	// ResultZero reports a use of a zero mo.Result.
	ResultZero Name = "result-zero"
	// UnwrapNil reports OrEmpty, or OrElse with nil, on an Option or a Result
	// of a pointer.
	UnwrapNil Name = "unwrap-nil"
	// UnwrapDiscard reports Get with its second result discarded.
	UnwrapDiscard Name = "unwrap-discard"
)

// All lists every rule, in the order the documentation gives them.
var All = []Name{ReturnNil, ReturnBool, ReturnError, WrapNil, ResultZero, UnwrapNil, UnwrapDiscard}

// OnByDefault reports whether a rule is on when no flag says otherwise.
// return-error is for those who want mo.Result everywhere, so it is off.
func OnByDefault(n Name) bool {
	return n != ReturnError
}

// Known reports whether s names a rule.
func Known(s string) bool {
	for _, n := range All {
		if string(n) == s {
			return true
		}
	}
	return false
}
