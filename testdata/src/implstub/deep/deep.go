// Package deep declares an interface that implementing does not import
// directly: implstub imports it.
package deep

type Deep interface {
	DeepGet() (int, bool)
}
