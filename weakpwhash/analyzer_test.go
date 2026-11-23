package weakpwhash_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/weakpwhash"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()

	analysistest.Run(t, testdata, weakpwhash.Analyzer, "example")
}
