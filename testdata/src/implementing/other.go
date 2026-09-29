package implementing

// This file does not import implstub itself: the package imports it in
// implementing.go, which is enough.

type O struct{}

// Not reported: O implements implstub.Other.
func (O) OtherGet() (int, bool) { return 0, false }
