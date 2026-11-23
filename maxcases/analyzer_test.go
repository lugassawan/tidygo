package maxcases_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/maxcases"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, maxcases.Analyzer, "example")
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		cfg  maxcases.Config
		pkg  string
	}{
		{
			name: "a positive max replaces the default",
			cfg:  maxcases.Config{Max: 3},
			pkg:  "custom",
		},
	}
	testdata := analysistest.TestData()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			analysistest.Run(t, testdata, maxcases.New(c.cfg), c.pkg)
		})
	}
}

func TestNewNonPositiveMaxUsesDefault(t *testing.T) {
	testdata := analysistest.TestData()
	for _, limit := range []int{0, -1} {
		// The example package is written against the default max of 30.
		analysistest.Run(t, testdata, maxcases.New(maxcases.Config{Max: limit}), "example")
	}
}
