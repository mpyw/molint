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
)

func init() {
	register.Plugin(molint.Analyzer.Name, newPlugin)
}

// pluginAnalyzer is the plugin built from one settings block.
type pluginAnalyzer struct {
	analyzer *analysis.Analyzer
}

func newPlugin(settings any) (register.LinterPlugin, error) {
	named, err := register.DecodeSettings[map[string]bool](settings)
	if err != nil {
		return nil, fmt.Errorf("reading settings: %w", err)
	}
	a, err := molint.New(molint.Config{Rules: named})
	if err != nil {
		return nil, fmt.Errorf("settings: %w", err)
	}
	return &pluginAnalyzer{analyzer: a}, nil
}

// BuildAnalyzers gives the analyzer that molint.New built from the settings.
// It has no flags, so molint.Analyzer is left alone.
func (p *pluginAnalyzer) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{p.analyzer}, nil
}

// GetLoadMode asks for type information, which buildssa needs.
func (*pluginAnalyzer) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
