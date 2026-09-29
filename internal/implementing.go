package internal

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"
)

// implementingMethods finds the methods of the package that implement an
// interface. The interface must be named or aliased and not generic. It must
// be declared in the package, at package level or in a function body, or in
// a package the package imports directly. An interface elsewhere is not
// seen, so a method that implements only that one is not exempt.
//
// A package is checked twice when it has tests: alone, and with its test
// files. A method outside the test files is judged by what those files see
// alone. So an interface that only a test file declares or imports does not
// exempt it in one check but not the other.
//
//declscope:package // checker.go computes it once per pass
func (c *checker) implementingMethods() map[*types.Func]bool {
	all := make(map[string][]*types.Interface)
	plain := make(map[string][]*types.Interface)
	add := func(tn *types.TypeName, inPlain bool) {
		if n, ok := tn.Type().(*types.Named); ok && n.TypeParams().Len() > 0 {
			return
		}
		iface, ok := tn.Type().Underlying().(*types.Interface)
		if !ok {
			return
		}
		for i := range iface.NumMethods() {
			name := iface.Method(i).Name()
			all[name] = append(all[name], iface)
			if inPlain {
				plain[name] = append(plain[name], iface)
			}
		}
	}
	for _, obj := range c.pass.TypesInfo.Defs {
		if tn, ok := obj.(*types.TypeName); ok {
			add(tn, !c.implementingInTest(tn))
		}
	}
	plainImports := make(map[string]bool)
	for _, f := range c.pass.Files {
		if c.implementingInTest(f) {
			continue
		}
		for _, spec := range f.Imports {
			plainImports[strings.Trim(spec.Path.Value, `"`)] = true
		}
	}
	for _, imp := range c.pass.Pkg.Imports() {
		scope := imp.Scope()
		for _, name := range scope.Names() {
			if tn, ok := scope.Lookup(name).(*types.TypeName); ok {
				add(tn, plainImports[imp.Path()])
			}
		}
	}

	out := make(map[*types.Func]bool)
	for _, f := range c.pass.Files {
		byMethod := plain
		if c.implementingInTest(f) {
			byMethod = all
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Recv == nil {
				continue
			}
			m := c.pass.TypesInfo.Defs[fd.Name].(*types.Func)
			recv := m.Type().(*types.Signature).Recv().Type()
			if p, ok := recv.(*types.Pointer); ok {
				recv = p.Elem()
			}
			for _, iface := range byMethod[m.Name()] {
				if types.Implements(recv, iface) || types.Implements(types.NewPointer(recv), iface) {
					out[m] = true
					break
				}
			}
		}
	}
	return out
}

// implementingInTest reports whether a declaration or a file is in a test
// file.
func (c *checker) implementingInTest(n interface{ Pos() token.Pos }) bool {
	return strings.HasSuffix(c.pass.Fset.PositionFor(n.Pos(), false).Filename, "_test.go")
}
