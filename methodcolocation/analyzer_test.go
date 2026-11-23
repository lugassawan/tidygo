package methodcolocation_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/methodcolocation"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, methodcolocation.Analyzer, "example")
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		cfg  methodcolocation.Config
		pkg  string
	}{
		{
			name: "allow-file-suffix permits only <stem>_<suffix>.go siblings",
			cfg:  methodcolocation.Config{AllowFileSuffix: true},
			pkg:  "suffix",
		},
	}
	testdata := analysistest.TestData()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			analysistest.Run(t, testdata, methodcolocation.New(c.cfg), c.pkg)
		})
	}
}
