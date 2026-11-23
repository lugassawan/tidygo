package deferredcancel_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/deferredcancel"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, deferredcancel.Analyzer, "example")
}
