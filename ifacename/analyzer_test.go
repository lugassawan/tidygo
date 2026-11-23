package ifacename_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/ifacename"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, ifacename.Analyzer, "example")
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		cfg  ifacename.Config
		pkg  string
	}{
		{
			name: "allowed-names exempts only the listed names",
			cfg:  ifacename.Config{AllowedNames: []string{"Handler"}},
			pkg:  "allowed",
		},
	}
	testdata := analysistest.TestData()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			analysistest.Run(t, testdata, ifacename.New(c.cfg), c.pkg)
		})
	}
}
