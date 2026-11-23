package maxcaselines_test

import (
	"strconv"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/maxcaselines"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, maxcaselines.Analyzer, "example")
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		cfg  maxcaselines.Config
		pkg  string
	}{
		{
			name: "a positive max replaces the default",
			cfg:  maxcaselines.Config{Max: 3},
			pkg:  "custom",
		},
	}
	testdata := analysistest.TestData()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			analysistest.Run(t, testdata, maxcaselines.New(c.cfg), c.pkg)
		})
	}
}

func TestNewNonPositiveMaxUsesDefault(t *testing.T) {
	testdata := analysistest.TestData()
	for _, limit := range []int{0, -1} {
		// The example package is written against the default max of 10.
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			analysistest.Run(t, testdata, maxcaselines.New(maxcaselines.Config{Max: limit}), "example")
		})
	}
}
