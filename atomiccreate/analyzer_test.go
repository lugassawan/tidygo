package atomiccreate_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/atomiccreate"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, atomiccreate.Analyzer, "example")
}
