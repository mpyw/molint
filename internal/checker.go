package internal

import (
	"cmp"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/mpyw/molint/internal/directive"
	"github.com/mpyw/molint/internal/rule"
)

// checker is one pass's state.
//
//declscope:package
type checker struct {
	pass *analysis.Pass
	// directives is what the package's //molint: comments say.
	directives *directive.Set
	// exempt holds the methods that implement an interface. Their
	// signatures are fixed by the interface.
	exempt map[*types.Func]bool
	//declscope:private
	cfg Config
	// generated holds the files marked as generated, by name. Nothing is
	// reported in them.
	//
	//declscope:private
	generated map[string]bool
	// diagnostics holds the reports until the pass ends, when they are
	// reported in the order of their positions.
	//
	//declscope:private
	diagnostics []analysis.Diagnostic
}

//declscope:package
func newChecker(pass *analysis.Pass, cfg Config) *checker {
	c := &checker{
		pass:       pass,
		cfg:        cfg,
		directives: directive.Scan(pass.Fset, pass.Files, pass.ReadFile),
		generated:  make(map[string]bool),
	}
	for _, f := range pass.Files {
		if directive.Generated(f) {
			c.generated[pass.Fset.PositionFor(f.Pos(), false).Filename] = true
		}
	}
	c.exempt = c.implementingMethods()
	return c
}

// report reports msg at pos for rule r, unless r is off, pos is in a
// generated file, or an ignore silences it there.
//
//declscope:package
func (c *checker) report(pos token.Pos, r rule.Name, msg string) {
	if !c.cfg.On(r) || c.inGenerated(pos) {
		return
	}
	if c.directives.Ignored(pos, r) {
		return
	}
	c.diagnostics = append(c.diagnostics, analysis.Diagnostic{Pos: pos, Message: msg + " [" + string(r) + "]"})
}

// inGenerated reports whether pos is in a file marked as generated.
//
//declscope:package // run.go skips the functions there
func (c *checker) inGenerated(pos token.Pos) bool {
	return c.generated[c.pass.Fset.PositionFor(pos, false).Filename]
}

// reportf reports a message that belongs to no rule, such as a problem with
// a directive. No ignore silences it.
//
//declscope:package
func (c *checker) reportf(pos token.Pos, msg string) {
	c.diagnostics = append(c.diagnostics, analysis.Diagnostic{Pos: pos, Message: msg})
}

// flush reports what the pass found, in the order of the positions.
//
//declscope:package
func (c *checker) flush() {
	slices.SortStableFunc(c.diagnostics, func(a, b analysis.Diagnostic) int { return cmp.Compare(a.Pos, b.Pos) })
	for _, d := range c.diagnostics {
		c.pass.Report(d)
	}
}
