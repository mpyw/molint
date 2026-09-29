package internal

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/mpyw/molint/internal/directive"
)

// checker is one pass's state. Each stage keeps its own state in a struct
// declared in the file that owns it and embedded here, so that the call sites
// read c.funcs rather than c.summary.funcs, and reaching into another stage's
// state is a boundary crossing.
//
//declscope:package
type checker struct {
	pass *analysis.Pass
	// directives is what the package's //molint: comments say.
	directives *directive.Set
	// generated holds the files marked as generated, by name. Nothing is
	// reported in them, but their functions are summarized like any other.
	generated map[string]bool

	summaryBook
	globalBook
}

//declscope:package
func newChecker(pass *analysis.Pass) *checker {
	c := &checker{
		pass:       pass,
		directives: directive.Scan(pass.Fset, pass.Files),
		generated:  make(map[string]bool),
	}
	for _, f := range pass.Files {
		if ast.IsGenerated(f) {
			c.generated[pass.Fset.Position(f.Pos()).Filename] = true
		}
	}
	return c
}
