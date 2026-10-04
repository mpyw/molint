// Package fieldnilother declares a struct in another package. The rules
// about fields do not look at its fields.
package fieldnilother

type T struct{ N int }

type Other struct {
	P *T
}
