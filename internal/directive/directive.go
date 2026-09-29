// Package directive reads the //molint: comments in a package.
//
// One directive exists. //molint:ignore // <reason> silences a report on its
// own line or the line below it. The reason is required, so that the next
// reader can tell a decision from a shortcut. Any other directive is
// reported.
package directive

import (
	"go/ast"
	"go/token"
	"strings"
)

// tool is the tool name of every directive, as in //molint:ignore.
// Directives follow the syntax of https://go.dev/doc/comment#directives, and
// go/ast parses them. A comment with a space after the slashes is prose.
const tool = "molint"

// Set is what the directives in a package say.
type Set struct {
	fset     *token.FileSet
	ignores  map[string]map[int]*ignore
	problems []Problem
}

// Problem is a directive that is malformed or unknown.
type Problem struct {
	Pos     token.Pos
	Message string
}

// ignore is one ignore comment, and whether it silenced anything.
type ignore struct {
	pos  token.Pos
	used bool
}

// Scan reads the directives in files. A generated file is skipped: nothing is
// reported in it, so neither its ignores nor its problems mean anything.
func Scan(fset *token.FileSet, files []*ast.File) *Set {
	s := &Set{fset: fset, ignores: make(map[string]map[int]*ignore)}
	for _, f := range files {
		if !ast.IsGenerated(f) {
			s.scanFile(f)
		}
	}
	return s
}

func (s *Set) scanFile(f *ast.File) {
	// Keyed by the file and line on disk. A //line directive renames the
	// positions below it, and two of them can give two lines one number.
	lines := make(map[int]*ignore)
	s.ignores[s.fset.PositionFor(f.Pos(), false).Filename] = lines
	for _, cg := range f.Comments {
		for _, cm := range cg.List {
			name, args, reason, ok := parse(cm)
			switch {
			case !ok:
			case name != "ignore":
				s.problems = append(s.problems, Problem{Pos: cm.Pos(),
					Message: "unknown directive molint:" + name})
			case args != "":
				s.problems = append(s.problems, Problem{Pos: cm.Pos(),
					Message: "molint:ignore takes no argument; write the reason after //"})
			case !reason:
				s.problems = append(s.problems, Problem{Pos: cm.Pos(),
					Message: "molint:ignore needs a reason after //"})
			default:
				lines[s.fset.PositionFor(cm.Pos(), false).Line] = &ignore{pos: cm.Pos()}
			}
		}
	}
}

// parse returns the directive's name and arguments, and whether a reason
// follows them, when cm is one of this tool's. The reason is a trailing
// comment, so //molint:ignore // why and //molint:ignore//why are both an
// ignore with a reason. Text after the name that is not behind // is an
// argument, which the ignore does not take, so that every tool of this family
// reads a directive the same way.
func parse(cm *ast.Comment) (name, args string, reason, ok bool) {
	text := cm.Text
	if body, line := strings.CutPrefix(text, "//"); line {
		if before, after, found := strings.Cut(body, "//"); found {
			text = "//" + before
			reason = strings.TrimSpace(after) != ""
		}
	}
	d, ok := ast.ParseDirective(cm.Slash, text)
	if !ok || d.Tool != tool {
		return "", "", false, false
	}
	return d.Name, d.Args, reason, true
}

// Ignored reports whether an ignore comment on the line of pos, or the line
// above it, silences a report there, and records that it did.
func (s *Set) Ignored(pos token.Pos) bool {
	p := s.fset.PositionFor(pos, false)
	lines := s.ignores[p.Filename]
	for _, l := range []int{p.Line, p.Line - 1} {
		if ig, ok := lines[l]; ok {
			ig.used = true
			return true
		}
	}
	return false
}

// Unused lists the ignore comments that silenced nothing.
func (s *Set) Unused() []token.Pos {
	var out []token.Pos
	for _, lines := range s.ignores {
		for _, ig := range lines {
			if !ig.used {
				out = append(out, ig.pos)
			}
		}
	}
	return out
}

// Problems lists the directives that are malformed or unknown.
func (s *Set) Problems() []Problem {
	return s.problems
}
