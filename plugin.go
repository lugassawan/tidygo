// Package tidygo registers custom lint analyzers as a golangci-lint v2 module plugin.
package tidygo

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/lugassawan/tidygo/funcname"
	"github.com/lugassawan/tidygo/maxparams"
	"github.com/lugassawan/tidygo/nolateconst"
	"github.com/lugassawan/tidygo/nolateexport"
	"github.com/lugassawan/tidygo/nolocalstruct"
)

func init() {
	register.Plugin("tidygo", newPlugin)
}

func newPlugin(_ any) (register.LinterPlugin, error) {
	return &plugin{}, nil
}

type plugin struct{}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		funcname.Analyzer,
		maxparams.Analyzer,
		nolateconst.Analyzer,
		nolateexport.Analyzer,
		nolocalstruct.Analyzer,
	}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}
