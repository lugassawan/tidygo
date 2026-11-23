package maxboolops_test

import (
	"strconv"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/maxboolops"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, maxboolops.Analyzer, "example")
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		cfg  maxboolops.Config
		pkg  string
	}{
		{
			name: "a positive max replaces the default",
			cfg:  maxboolops.Config{Max: 1},
			pkg:  "custom",
		},
	}
	testdata := analysistest.TestData()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			analysistest.Run(t, testdata, maxboolops.New(c.cfg), c.pkg)
		})
	}
}

func TestNewNonPositiveMaxUsesDefault(t *testing.T) {
	testdata := analysistest.TestData()
	for _, limit := range []int{0, -1} {
		// The example package is written against the default max of 3.
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			analysistest.Run(t, testdata, maxboolops.New(maxboolops.Config{Max: limit}), "example")
		})
	}
}
