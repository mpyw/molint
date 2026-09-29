// Package typeutil holds the type questions every part of the analysis asks.
package typeutil

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// errorType is the predeclared error type.
var errorType = types.Universe.Lookup("error").Type()

// IsPointer reports whether a value of type t is a pointer. A named type whose
// underlying type is a pointer is one too. A type parameter is not, whatever
// its constraint says.
func IsPointer(t types.Type) bool {
	if _, ok := types.Unalias(t).(*types.TypeParam); ok {
		return false
	}
	_, ok := t.Underlying().(*types.Pointer)
	return ok
}

// ErrorResult is the index of the error result of sig: its last result, when
// that is of type error. It is -1 when sig has none.
func ErrorResult(sig *types.Signature) int {
	r := sig.Results()
	if n := r.Len(); n > 0 && types.Identical(r.At(n-1).Type(), errorType) {
		return n - 1
	}
	return -1
}

// PointerResults lists the indices of the pointer results of sig.
func PointerResults(sig *types.Signature) []int {
	var out []int
	for i := range sig.Results().Len() {
		if IsPointer(sig.Results().At(i).Type()) {
			out = append(out, i)
		}
	}
	return out
}

// Qualifier spells a package by its name, and the package pkg not at all.
func Qualifier(pkg *types.Package) types.Qualifier {
	return func(p *types.Package) string {
		if p == pkg {
			return ""
		}
		return p.Name()
	}
}

// FuncName spells fn as a diagnostic in pkg names it: F, pkg.F, T.M,
// (*T).M or (*pkg.T).M. An instance of a generic function is spelled as its
// origin.
func FuncName(fn *ssa.Function, pkg *types.Package) string {
	if o := fn.Origin(); o != nil {
		fn = o
	}
	qual := Qualifier(pkg)
	if recv := fn.Signature.Recv(); recv != nil {
		t := types.TypeString(recv.Type(), qual)
		if _, ok := recv.Type().(*types.Pointer); ok {
			t = "(" + t + ")"
		}
		return t + "." + fn.Name()
	}
	if fn.Pkg != nil && fn.Pkg.Pkg != pkg {
		return fn.Pkg.Pkg.Name() + "." + fn.Name()
	}
	return fn.Name()
}

// Func is the declared function behind fn, or nil when fn is a function
// literal. fn is not an instance of a generic function: the caller takes its
// origin first. A synthetic wrapper stands for the method it wraps, whose
// results it returns as they are.
func Func(fn *ssa.Function) *types.Func {
	obj, ok := fn.Object().(*types.Func)
	if !ok {
		return nil
	}
	return obj.Origin()
}
