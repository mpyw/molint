package nilproof_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/mpyw/nilproof"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), nilproof.Analyzer, "basic", "errpair", "lib", "crosspkg", "globals", "directives", "generated", "generics", "coverage")
}
