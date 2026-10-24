package internal

import (
	"errors"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestRunWithoutSSA(t *testing.T) {
	if _, err := Run(&analysis.Pass{ResultOf: map[*analysis.Analyzer]any{}}, Config{}); !errors.Is(err, ErrRunWithoutSSA) {
		t.Errorf("Run() error = %v, want %v", err, ErrRunWithoutSSA)
	}
}
