package busywait_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/busywait"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, busywait.Analyzer, "example")
}
