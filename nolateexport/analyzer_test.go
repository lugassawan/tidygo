package nolateexport_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/nolateexport"
)

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, nolateexport.Analyzer, "example")
}

func TestNew(t *testing.T) {
	cases := []struct {
		name string
		cfg  nolateexport.Config
		pkg  string
	}{
		{name: "per-type method ordering", cfg: nolateexport.Config{}, pkg: "example"},
		{name: "type declared in another file", cfg: nolateexport.Config{}, pkg: "splitfile"},
		{
			name: "per-file state does not leak, option on",
			cfg:  nolateexport.Config{UnexportedFuncsLast: true},
			pkg:  "splitfile",
		},
		{name: "unexported-funcs-last on", cfg: nolateexport.Config{UnexportedFuncsLast: true}, pkg: "funcslast"},
		{name: "unexported-funcs-last off by default", cfg: nolateexport.Config{}, pkg: "funcslastoff"},
	}
	testdata := analysistest.TestData()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			analysistest.Run(t, testdata, nolateexport.New(c.cfg), c.pkg)
		})
	}
}
