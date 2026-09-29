// Package implementing covers the exemption of a method that implements an
// interface, for return-bool and return-nil.
package implementing

import "implstub"

type T struct{}

// ===== An interface in the package =====

type LocalIface interface {
	LocalGet() (int, bool) // want `^LocalIface\.LocalGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
}

type V1 struct{}

// Not reported: V1 implements LocalIface.
func (V1) LocalGet() (int, bool) { return 0, false }

type P1 struct{}

// Not reported: *P1 implements LocalIface.
func (*P1) LocalGet() (int, bool) { return 0, false }

type Mixed interface {
	MixA() (int, bool) // want `^Mixed\.MixA reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	MixB()
}

type M struct{}

// Not reported: M does not implement Mixed, but *M does.
func (M) MixA() (int, bool) { return 0, false }

func (*M) MixB() {}

type Pair interface {
	PairA() (int, bool) // want `^Pair\.PairA reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	PairB()
}

type N struct{}

// N lacks PairB, so it implements nothing.
func (N) PairA() (int, bool) { return 0, false } // want `^N\.PairA reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

type SigI interface {
	Sig() (int, bool) // want `^SigI\.Sig reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
}

type W struct{}

// The name matches, but the signature does not.
func (W) Sig() (string, bool) { return "", false } // want `^W\.Sig reports absence with a trailing bool; return mo\.Option\[string\] instead \[return-bool\]$`

type Multi struct{}

func (Multi) LocalGet() (int, bool) { return 0, false }

// Multi implements LocalIface, but Extra is not one of its methods.
func (Multi) Extra() (int, bool) { return 0, false } // want `^Multi\.Extra reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

// ===== An interface in a package imported directly =====

type G struct{}

// Not reported: G implements implstub.Getter.
func (G) Get() (int, bool) { return 0, false }

// Checked with implstub, so the import is used by more than the methods.
var _ implstub.Getter = G{}

type Op struct{}

// Not reported by return-nil: Op implements implstub.Opener.
func (Op) Open() *implstub.File { return nil }

// ===== An interface in a package imported only indirectly =====

type D struct{}

// implementing does not import implstub/deep, so deep.Deep does not count.
func (D) DeepGet() (int, bool) { return 0, false } // want `^D\.DeepGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

type ED struct{}

// Not reported: ED implements implstub.EmbedsDeep, which is in a package
// imported directly. deep.Deep alone would not count: D above is reported.
func (ED) DeepGet() (int, bool) { return 0, false }

func (ED) X() {}

// ===== Generic interfaces =====

type LocalGeneric[E any] interface {
	GenGet() (E, bool) // want `^LocalGeneric\[E\]\.GenGet reports absence with a trailing bool; return mo\.Option\[E\] instead \[return-bool\]$`
}

type X struct{}

// A generic interface does not count, although X implements LocalGeneric[int].
func (X) GenGet() (int, bool) { return 0, false } // want `^X\.GenGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

var _ LocalGeneric[int] = X{}

type Y struct{}

// An imported generic interface does not count either.
func (Y) GGet() (int, bool) { return 0, false } // want `^Y\.GGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

var _ implstub.GenericGetter[int] = Y{}

type GenMaker[E any] interface {
	GMake() *E
}

type Gm struct{}

// A generic interface does not exempt return-nil.
func (Gm) GMake() *int {
	return nil // want `^Gm\.GMake returns a nil \*int; return mo\.Option\[\*int\] instead \[return-nil\]$`
}

var _ GenMaker[int] = Gm{}

// ===== Embedded interfaces =====

type Reader interface {
	Read2() (int, bool) // want `^Reader\.Read2 reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
}

// Not reported: nothing is declared here but Close.
type ReadCloser interface {
	Reader
	Close()
}

type RC struct{}

// Not reported: RC implements Reader and ReadCloser.
func (RC) Read2() (int, bool) { return 0, false }

func (RC) Close() {}

type OnlyEmbedded interface {
	Reader
	OnlyExtra()
}

type OE struct{}

// Not reported: OE implements OnlyEmbedded through a method of the embedded
// Reader; it implements Reader too.
func (*OE) Read2() (int, bool) { return 0, false }

func (*OE) OnlyExtra() {}

type LocalEmbeds interface {
	implstub.Embedded
	Extra3()
}

type E1 struct{}

// Not reported: E1 implements LocalEmbeds, whose EmbGet comes from an imported
// interface.
func (E1) EmbGet() (int, bool) { return 0, false }

func (E1) Extra3() {}

// ===== Promoted methods =====

type PromI interface {
	PromGet() (int, bool) // want `^PromI\.PromGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	Close()
}

type Inner struct{}

func (Inner) Close() {}

type Outer struct{ Inner }

// Not reported: Outer implements PromI with a Close promoted from Inner.
func (Outer) PromGet() (int, bool) { return 0, false }

type PromJ interface {
	InnerGet() (int, bool) // want `^PromJ\.InnerGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	Other2()
}

type Inner2 struct{}

// Outer2 implements PromJ through this promoted method, but the receiver
// *Inner2 does not implement PromJ.
func (*Inner2) InnerGet() (int, bool) { return 0, false } // want `^\(\*Inner2\)\.InnerGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

type Outer2 struct{ *Inner2 }

func (Outer2) Other2() {}

var _ PromJ = Outer2{}

// ===== Interfaces that are not named =====

type Z struct{}

// An interface literal does not count.
func (Z) AnonGet() (int, bool) { return 0, false } // want `^Z\.AnonGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`

var _ interface{ AnonGet() (int, bool) } = Z{}

// ===== return-nil =====

type Maker interface {
	Make() *T
}

type Mk struct{}

// Not reported: *Mk implements Maker.
func (*Mk) Make() *T { return nil }

type Nm struct{}

// Make2 is not the method of any interface.
func (Nm) Make2() *T {
	return nil // want `^Nm\.Make2 returns a nil \*T; return mo\.Option\[\*T\] instead \[return-nil\]$`
}

// ===== Generic receivers =====

type BoxI interface {
	BoxGet() (int, bool) // want `^BoxI\.BoxGet reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
}

type Box[E any] struct{}

// Not reported: Box[E] implements BoxI.
func (Box[E]) BoxGet() (int, bool) { return 0, false }

type BoxValI interface {
	BoxVal() (int, bool) // want `^BoxValI\.BoxVal reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
}

// Box[int] would implement BoxValI, but the receiver's type Box[E] does not.
func (Box[E]) BoxVal() (E, bool) { // want `^Box\[E\]\.BoxVal reports absence with a trailing bool; return mo\.Option\[E\] instead \[return-bool\]$`
	var z E
	return z, false
}

// ===== Interfaces inside a function body and aliases =====

func declaresLocal() {
	type localGetter interface {
		LocalOnly() (int, bool) // want `^localGetter\.LocalOnly reports absence with a trailing bool; return mo\.Option\[int\] instead \[return-bool\]$`
	}
	var _ localGetter = L{}
}

type L struct{}

// Not reported: L implements an interface declared inside a function body.
func (L) LocalOnly() (int, bool) { return 0, false }

// An alias of an interface literal counts for the exemption. Its method is
// not reported: it is not a method of a named interface type.
type AliasGetter = interface {
	AliasGet() (int, bool)
}

type AG struct{}

// Not reported: AG implements AliasGetter.
func (AG) AliasGet() (int, bool) { return 0, false }
