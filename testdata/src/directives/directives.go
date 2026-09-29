// Package directives covers //molint:ignore and the other //molint:
// directives.
package directives

import "github.com/samber/mo"

type T struct{ N int }

func use(...any) {}

// ===== Silencing =====

// Not reported: the directive on the line above silences return-bool.
//
//molint:ignore return-bool // callers rely on the comma-ok form
func Above() (int, bool) { return 0, false }

// Not reported: the directive on the same line silences return-bool.
func SameLine() (int, bool) { //molint:ignore return-bool // callers rely on the comma-ok form
	return 0, false
}

// Not reported: the directive silences return-nil on the line below.
func ReturnNil() *T {
	//molint:ignore return-nil // absence is signalled out of band
	return nil
}

// Not reported: the directive at the end of the line silences return-nil.
func ReturnNilSameLine() *T {
	return nil //molint:ignore return-nil // absence is signalled out of band
}

// Not reported: both rules on the line are silenced.
func TwoRules() (*T, mo.Option[*T]) {
	//molint:ignore return-nil, wrap-nil // both are intended here
	return nil, mo.Some[*T](nil)
}

// Not reported: the rules may be written without a space.
func TwoRulesNoSpace() (*T, mo.Option[*T]) {
	//molint:ignore return-nil,wrap-nil // both are intended here
	return nil, mo.Some[*T](nil)
}

// Not reported: an ignore with no rules silences every rule.
func AllRules() (*T, mo.Option[*T]) {
	//molint:ignore // everything here is intended
	return nil, mo.Some[*T](nil)
}

// Not reported: one directive silences two reports of the same rule.
func TwoReports() (*T, *T) {
	//molint:ignore return-nil // both results are optional here
	return nil, nil
}

// Only the named rule is silenced; wrap-nil is still reported.
func OneOfTwo() (*T, mo.Option[*T]) {
	//molint:ignore return-nil // only the pointer is intended
	return nil, mo.Some[*T](nil) // want `^mo\.Some is given nil; pass a non-nil value, or use mo\.None \[wrap-nil\]$`
}

// Not reported: a named interface method is silenced.
type Finder interface {
	//molint:ignore return-bool // an established interface
	Find(string) (*T, bool)
}

// Not reported: a named function type is silenced.
//
//molint:ignore return-bool // an established callback shape
type FindFunc func(string) (*T, bool)

// ===== Directives that silence nothing =====

// The rule named is not the one reported.
func WrongRule() *T {
	//molint:ignore return-bool // wrong rule // want `^unused molint:ignore directive$`
	return nil // want `^WrongRule returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// The directive covers its own line and the line below, not further.
func TwoLinesBelow() *T {
	//molint:ignore return-nil // too far // want `^unused molint:ignore directive$`
	use()
	return nil // want `^TwoLinesBelow returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Nothing is reported where the directive is.
func NothingToSilence() *T {
	//molint:ignore return-nil // not needed // want `^unused molint:ignore directive$`
	return &T{}
}

// An ignore with no rules that silences nothing.
func NothingToSilenceAll() *T {
	//molint:ignore // not needed // want `^unused molint:ignore directive$`
	return &T{}
}

// A trailing directive silences its own line only, not the line below.
func TrailingDoesNotReachBelow() *T {
	use()      //molint:ignore return-nil // trailing // want `^unused molint:ignore directive$`
	return nil // want `^TrailingDoesNotReachBelow returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// Not reported: an ignore whose every rule is off by flag is not unused.
// return-error is off by default.
//
//molint:ignore return-error // an established signature
func Parse() (int, error) { return 0, nil }

// ===== Malformed directives =====

// A missing or empty reason cannot be pinned here: the expectation comment
// on the directive's line would be read as its reason. A unit test covers
// "molint:ignore needs a reason after //".

// The known rule still silences; the unknown one is reported.
func UnknownRule() *T {
	//molint:ignore return-nil, bogus // intended // want `^unknown rule bogus in molint:ignore$`
	return nil
}

// Not reported as unused: every rule is unknown, so only that is reported.
func AllUnknown() *T {
	//molint:ignore bogus // intended // want `^unknown rule bogus in molint:ignore$`
	return &T{}
}

// The known rule still silences; the empty name is reported.
func EmptyRuleName() *T {
	//molint:ignore return-nil, // intended // want `^empty rule name in molint:ignore$`
	return nil
}

//molint:foo // want `^unknown directive molint:foo$`

// A report about a directive cannot be silenced, so this ignore silences
// nothing.
//
//molint:ignore // tries to silence the line below // want `^unused molint:ignore directive$`
//molint:bar // want `^unknown directive molint:bar$`

//molint:ignorex // want `^unknown directive molint:ignorex$`

// Prose that mentions molint:ignore with a space is not a directive, so it
// neither silences nor is reported as unused.
func Prose() *T {
	// molint:ignore return-nil
	return nil // want `^Prose returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}
