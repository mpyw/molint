package internal

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/mpyw/molint/internal/rule"
	"github.com/mpyw/molint/internal/typeutil"
)

// checkShapes reports the signatures that end in a bool or an error after
// other results: functions, methods, methods of named interfaces, and named
// function types. A function literal is not one of them, and neither is a
// method that implements an interface.
//
//declscope:package // run.go calls it
func (c *checker) checkShapes() {
	info := c.pass.TypesInfo
	for _, f := range c.pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.FuncDecl:
				fn, ok := info.Defs[n.Name].(*types.Func)
				if ok && !c.exempt[fn] {
					c.checkShape(n.Name.Pos(), typeutil.FuncName(fn, c.pass.Pkg), fn.Type().(*types.Signature))
				}
			case *ast.TypeSpec:
				c.checkTypeSpecShape(n)
			}
			return true
		})
	}
}

// checkTypeSpecShape checks a named function type, or the methods of a named
// interface type. An alias names no new type, so it is not checked.
func (c *checker) checkTypeSpecShape(ts *ast.TypeSpec) {
	if ts.Assign.IsValid() {
		return
	}
	tn := c.pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
	switch t := ts.Type.(type) {
	case *ast.FuncType:
		if sig, ok := tn.Type().Underlying().(*types.Signature); ok {
			c.checkShape(ts.Name.Pos(), typeutil.TypeName(tn), sig)
		}
	case *ast.InterfaceType:
		for _, field := range t.Methods.List {
			for _, name := range field.Names {
				if m, ok := c.pass.TypesInfo.Defs[name].(*types.Func); ok {
					c.checkShape(name.Pos(), typeutil.TypeName(tn)+"."+name.Name, m.Type().(*types.Signature))
				}
			}
		}
	}
}

// checkShape reports sig, named name at pos, when it ends in a bool or an
// error after at least one other result.
func (c *checker) checkShape(pos token.Pos, name string, sig *types.Signature) {
	n := sig.Results().Len()
	if n < 2 {
		return
	}
	switch typeutil.TrailingOf(sig) {
	case typeutil.TrailingBool:
		c.report(pos, rule.ReturnBool, fmt.Sprintf("%s reports absence with a trailing bool; return %s instead", name, c.shapeFix("Option", sig)))
	case typeutil.TrailingError:
		c.report(pos, rule.ReturnError, fmt.Sprintf("%s reports failure with a trailing error; return %s instead", name, c.shapeFix("Result", sig)))
	}
}

// shapeFix names what to return instead: mo.X[T] for one value, and a
// struct or a tuple of samber/lo, which has Tuple2 to Tuple9, for several.
func (c *checker) shapeFix(mo string, sig *types.Signature) string {
	values := sig.Results().Len() - 1
	switch {
	case values == 1:
		return fmt.Sprintf("mo.%s[%s]", mo, typeutil.TypeString(sig.Results().At(0).Type(), c.pass.Pkg))
	case values <= 9:
		return fmt.Sprintf("mo.%s of a struct, or of lo.Tuple%d", mo, values)
	}
	return fmt.Sprintf("mo.%s of a struct", mo)
}
