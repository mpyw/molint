// Package plugin registers molint as a golangci-lint module plugin.
//
// Import it from .custom-gcl.yml to build a golangci-lint binary that holds
// molint. The settings in .golangci.yml map each rule's name to whether it is
// on, as the flags of the molint command do. A rule left out keeps its
// default.
package plugin

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/mpyw/molint"
	"github.com/mpyw/molint/internal"
	"github.com/mpyw/molint/internal/rule"
)

func init() {
	register.Plugin(molint.Analyzer.Name, newPlugin)
}

// pluginRules is the plugin built from one settings block.
type pluginRules struct {
	// on holds the rules the settings name.
	on map[rule.Name]bool
}

func newPlugin(settings any) (register.LinterPlugin, error) {
	named, err := register.DecodeSettings[map[string]bool](settings)
	if err != nil {
		return nil, fmt.Errorf("reading settings: %w", err)
	}
	on := make(map[rule.Name]bool, len(named))
	for name, v := range named {
		if !rule.Known(name) {
			return nil, fmt.Errorf("unknown rule %q in settings", name)
		}
		on[rule.Name(name)] = v
	}
	return &pluginRules{on: on}, nil
}

// BuildAnalyzers gives molint's analyzer, with the rules the settings turn on.
// It is a copy, so the flags of molint.Analyzer are left alone.
func (p *pluginRules) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	cfg := internal.Config{On: p.ruleOn}
	return []*analysis.Analyzer{{
		Name:     molint.Analyzer.Name,
		Doc:      molint.Analyzer.Doc,
		URL:      molint.Analyzer.URL,
		Requires: molint.Analyzer.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return internal.Run(pass, cfg)
		},
	}}, nil
}

// GetLoadMode asks for type information, which buildssa needs.
func (*pluginRules) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func (p *pluginRules) ruleOn(r rule.Name) bool {
	if v, ok := p.on[r]; ok {
		return v
	}
	return rule.OnByDefault(r)
}
