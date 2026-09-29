package directives

type T struct{}

func Ignored() *T {
	//nilproof:ignore // absence is signalled to a caller outside this module
	return nil
}

func SameLine() *T {
	return nil //nilproof:ignore // same
}

// An ignore without a reason is tested in internal/directive, since the
// expectation comment on its line would be read as the reason.

func Argument() *T {
	//nilproof:ignore because // want `nilproof:ignore takes no argument; write the reason after //`
	return nil // want `from a nil constant`
}

//nilproof:unknown // want `unknown directive nilproof:unknown`

func Unused() *T { // want Unused:"nonnil r0"
	//nilproof:ignore // nothing here // want `unused nilproof:ignore directive`
	return &T{}
}

// Prose is not a directive.
// nilproof:ignore
func Prose() *T {
	return nil // want `from a nil constant`
}
