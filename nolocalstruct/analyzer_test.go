package nolocalstruct_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/nolocalstruct"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, nolocalstruct.Analyzer, "example")
}

func TestNewAllTypes(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, nolocalstruct.New(nolocalstruct.Config{AllTypes: true}), "alltypes")
}

// TestNewDefaultIgnoresNonStructTypes pins the opt-in: without AllTypes the
// example package keeps its allowed `type myInt int` and reports only structs.
func TestNewDefaultIgnoresNonStructTypes(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, nolocalstruct.New(nolocalstruct.Config{}), "example")
}
