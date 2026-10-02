// Package typeutil holds the type questions every rule asks, and the way a
// diagnostic spells types and declarations.
package typeutil

import (
	"go/types"
	"strings"

	"golang.org/x/tools/go/ssa"
)

// moPath is the import path of samber/mo.
const moPath = "github.com/samber/mo"

// errorType is the predeclared error type.
var errorType = types.Universe.Lookup("error").Type()

// IsPointer reports whether a value of type t is a pointer: its underlying
// type is one. A type parameter is not, whatever its constraint says.
func IsPointer(t types.Type) bool {
	if _, ok := types.Unalias(t).(*types.TypeParam); ok {
		return false
	}
	_, ok := t.Underlying().(*types.Pointer)
	return ok
}

// CanBeNil reports whether a value of type t can be nil: a type whose nil
// breaks on use, or a slice. A type parameter can be none of these: its
// underlying type is an interface, never a slice.
func CanBeNil(t types.Type) bool {
	if _, ok := t.Underlying().(*types.Slice); ok {
		return true
	}
	return NilBreaks(t)
}

// NilBreaks reports whether a nil of type t breaks on use: a pointer, an
// interface, a map, a func or a channel. A nil slice does not, since it
// works as an empty one. A type parameter counts as none of these, although
// go/types calls its underlying type an interface.
func NilBreaks(t types.Type) bool {
	if _, ok := types.Unalias(t).(*types.TypeParam); ok {
		return false
	}
	switch t.Underlying().(type) {
	case *types.Pointer, *types.Interface, *types.Map, *types.Signature, *types.Chan:
		return true
	}
	return false
}

// Trailing is what the last result of a signature is.
type Trailing int

const (
	// TrailingNone means the last result is neither a bool nor an error. It
	// is the zero value.
	TrailingNone Trailing = iota
	// TrailingBool means the last result is the predeclared bool.
	TrailingBool
	// TrailingError means the last result is the predeclared error.
	TrailingError
)

// TrailingOf is what the last result of sig is. An alias of bool or error
// counts, since it is the same type. A named bool or error type does not.
func TrailingOf(sig *types.Signature) Trailing {
	r := sig.Results()
	if r.Len() == 0 {
		return TrailingNone
	}
	t := r.At(r.Len() - 1).Type()
	switch {
	case types.Identical(t, types.Typ[types.Bool]):
		return TrailingBool
	case types.Identical(t, errorType):
		return TrailingError
	}
	return TrailingNone
}

// Mo is which type of samber/mo a type is.
type Mo int

const (
	// NotMo means the type is not one of samber/mo.
	NotMo Mo = iota
	// Option is mo.Option.
	Option
	// Result is mo.Result.
	Result
)

// MoOf reports which type of samber/mo t is, and its type argument.
func MoOf(t types.Type) (Mo, types.Type) {
	n, ok := types.Unalias(t).(*types.Named)
	if !ok || n.Obj().Pkg() == nil || n.Obj().Pkg().Path() != moPath || n.TypeArgs().Len() != 1 {
		return NotMo, nil
	}
	switch n.Obj().Name() {
	case "Option":
		return Option, n.TypeArgs().At(0)
	case "Result":
		return Result, n.TypeArgs().At(0)
	}
	return NotMo, nil
}

// MoFunc is the name of the samber/mo function fn is, or "". A wrapper,
// such as a bound method or a thunk, is not one.
func MoFunc(fn *ssa.Function) string {
	obj := origin(fn)
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != moPath {
		return ""
	}
	if obj.Type().(*types.Signature).Recv() != nil {
		return ""
	}
	return obj.Name()
}

// MoMethod reports which samber/mo type fn is a method of, its name, and the
// type argument of its receiver. A wrapper, such as a bound method or a
// thunk, is not one.
func MoMethod(fn *ssa.Function) (Mo, string, types.Type) {
	obj := origin(fn)
	if obj == nil || fn.Signature.Recv() == nil {
		return NotMo, "", nil
	}
	// Every method molint looks at has a value receiver. A method with a
	// pointer receiver, such as Scan, is left out by MoOf.
	m, arg := MoOf(fn.Signature.Recv().Type())
	return m, obj.Name(), arg
}

// origin is the declared function behind fn: itself, or the generic function
// it instantiates. A synthetic wrapper has none.
func origin(fn *ssa.Function) *types.Func {
	if fn.Synthetic != "" && !strings.HasPrefix(fn.Synthetic, "instance ") && !strings.HasPrefix(fn.Synthetic, "instantiation ") {
		return nil
	}
	if o := fn.Origin(); o != nil {
		fn = o
	}
	obj, _ := fn.Object().(*types.Func)
	return obj
}

// RecvString spells the type a method of samber/mo is called on, as
// mo.Option[*T]. Only methods with a value receiver are asked about.
func RecvString(fn *ssa.Function, pkg *types.Package) string {
	return TypeString(fn.Signature.Recv().Type(), pkg)
}

// TypeString spells t as a diagnostic does. A type from another package is
// qualified by its package name. Aliases are resolved at every level, except
// any, which is how people write the empty interface.
func TypeString(t types.Type, pkg *types.Package) string {
	var b strings.Builder
	writeType(&b, t, pkg)
	return b.String()
}

func writeType(b *strings.Builder, t types.Type, pkg *types.Package) {
	qual := func(p *types.Package) string {
		if p == pkg {
			return ""
		}
		return p.Name()
	}
	switch t := t.(type) {
	case *types.Alias:
		if t.Obj().Pkg() == nil && t.Obj().Name() == "any" {
			b.WriteString("any")
			return
		}
		writeType(b, types.Unalias(t), pkg)
	case *types.Pointer:
		b.WriteString("*")
		writeType(b, t.Elem(), pkg)
	case *types.Slice:
		b.WriteString("[]")
		writeType(b, t.Elem(), pkg)
	case *types.Named:
		if p := t.Obj().Pkg(); p != nil {
			if q := qual(p); q != "" {
				b.WriteString(q + ".")
			}
		}
		b.WriteString(t.Obj().Name())
		if args := t.TypeArgs(); args.Len() > 0 {
			b.WriteString("[")
			for i := range args.Len() {
				if i > 0 {
					b.WriteString(", ")
				}
				writeType(b, args.At(i), pkg)
			}
			b.WriteString("]")
		}
	default:
		b.WriteString(types.TypeString(t, qual))
	}
}

// FuncName spells a function or a method as a diagnostic does: F, F[E],
// S.M, (*S).M, or (*Box[E]).M.
func FuncName(fn *types.Func, pkg *types.Package) string {
	sig := fn.Type().(*types.Signature)
	recv := sig.Recv()
	if recv == nil {
		return fn.Name() + typeParams(sig.TypeParams())
	}
	t := recv.Type()
	ptr := false
	if p, ok := t.(*types.Pointer); ok {
		t, ptr = p.Elem(), true
	}
	s := TypeString(t, pkg)
	if ptr {
		s = "(*" + s + ")"
	}
	return s + "." + fn.Name()
}

// TypeName spells a named type as a diagnostic does, with its type
// parameters: FindFunc, or GenericFunc[E]. An alias is never asked about.
func TypeName(tn *types.TypeName) string {
	return tn.Name() + typeParams(tn.Type().(*types.Named).TypeParams())
}

func typeParams(tps *types.TypeParamList) string {
	if tps.Len() == 0 {
		return ""
	}
	names := make([]string, tps.Len())
	for i := range tps.Len() {
		names[i] = tps.At(i).Obj().Name()
	}
	return "[" + strings.Join(names, ", ") + "]"
}
