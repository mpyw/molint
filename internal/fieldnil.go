package internal

import (
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"

	"github.com/mpyw/molint/internal/flow"
	"github.com/mpyw/molint/internal/rule"
	"github.com/mpyw/molint/internal/typeutil"
)

// checkFieldNil reports each nil value stored into a pointer field
// (field-nil-store), and each comparison of a pointer field with nil
// (field-nil-compare). A field left out of a composite literal has no store
// in SSA, so field-nil-store relies on exhaustruct to make it written.
//
//declscope:package // run.go calls it
func (c *checker) checkFieldNil(fn *ssa.Function) {
	nils := flow.NilTracer()
	for _, b := range fn.Blocks {
		at := flow.SiteAt(b)
		for _, in := range b.Instrs {
			switch in := in.(type) {
			case *ssa.Store:
				fa, ok := in.Addr.(*ssa.FieldAddr)
				if !ok {
					continue
				}
				if f, name := c.fieldNilTarget(fa.X.Type(), fa.Field); f != nil && nils.Is(in.Val, at) {
					c.report(in.Pos(), rule.FieldNilStore, fmt.Sprintf("%s is set to nil; make it mo.Option[%s]", name, typeutil.TypeString(f.Type(), c.pass.Pkg)))
				}
			case *ssa.BinOp:
				if in.Op != token.EQL && in.Op != token.NEQ {
					continue
				}
				if f, name := c.fieldNilCompared(in); f != nil {
					c.report(in.Pos(), rule.FieldNilCompare, fmt.Sprintf("%s is compared with nil; make it mo.Option[%s]", name, typeutil.TypeString(f.Type(), c.pass.Pkg)))
				}
			}
		}
	}
}

// fieldNilCompared is the field that a comparison with the constant nil
// reads, and its name, or nil when the comparison reads none.
func (c *checker) fieldNilCompared(cmp *ssa.BinOp) (*types.Var, string) {
	x, y := cmp.X, cmp.Y
	if k, ok := x.(*ssa.Const); ok && k.IsNil() {
		x, y = y, x
	}
	if k, ok := y.(*ssa.Const); !ok || !k.IsNil() {
		return nil, ""
	}
	for {
		ct, ok := x.(*ssa.ChangeType)
		if !ok {
			break
		}
		x = ct.X
	}
	switch x := x.(type) {
	case *ssa.Field:
		return c.fieldNilTarget(x.X.Type(), x.Field)
	case *ssa.UnOp:
		if fa, ok := x.X.(*ssa.FieldAddr); ok && x.Op == token.MUL {
			return c.fieldNilTarget(fa.X.Type(), fa.Field)
		}
	}
	return nil, ""
}

// fieldNilTarget is field i of the struct t is, or points to, as declared,
// and its name as a diagnostic spells it, when the rules about fields look
// at it. That is a pointer field, not embedded, declared in the package
// outside a generated file. It is nil, with "", when they do not.
//
// The field is taken from the generic type an instance comes from, so a
// diagnostic spells it the same way at every instance: Box[E].V, of type *E.
func (c *checker) fieldNilTarget(t types.Type, i int) (*types.Var, string) {
	if p, ok := t.Underlying().(*types.Pointer); ok {
		t = p.Elem()
	}
	// Go selects no field through a type parameter, so t is always a struct.
	st := t.Underlying().(*types.Struct)
	f := st.Field(i).Origin()
	if f.Embedded() || !typeutil.IsPointer(f.Type()) || f.Pkg() != c.pass.Pkg || c.inGenerated(f.Pos()) {
		return nil, ""
	}
	if n, ok := types.Unalias(t).(*types.Named); ok {
		return f, typeutil.TypeName(n.Obj()) + "." + f.Name()
	}
	return f, f.Name()
}
