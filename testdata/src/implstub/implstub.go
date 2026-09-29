// Package implstub declares interfaces that the implementing fixture imports.
// It is not a checked package.
package implstub

import "implstub/deep"

type File struct{}

type Getter interface {
	Get() (int, bool)
}

type Other interface {
	OtherGet() (int, bool)
}

type GenericGetter[E any] interface {
	GGet() (E, bool)
}

type Opener interface {
	Open() *File
}

type Embedded interface {
	EmbGet() (int, bool)
}

// EmbedsDeep embeds deep.Deep, so that a package importing only implstub
// sees DeepGet through it.
type EmbedsDeep interface {
	deep.Deep
	X()
}

// Holder refers to deep, so that deep is loaded without implementing
// importing it.
type Holder struct{ D deep.Deep }
