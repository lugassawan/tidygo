package funcname_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/funcname"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, funcname.Analyzer, "example")
}

func TestNewIncludeVars(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, funcname.New(funcname.Config{IncludeVars: true}), "includevars")
}

// TestNewDefaultIgnoresVariables pins the opt-in: the example package declares an
// underscore variable that must not be reported without IncludeVars.
func TestNewDefaultIgnoresVariables(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, funcname.New(funcname.Config{}), "example")
}
