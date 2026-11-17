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

// settings mirrors the golangci-lint `linters.settings.custom.tidygo.settings` block.
// Unknown keys are rejected by register.DecodeSettings, so typos fail loudly.
type settings struct {
	NoLateExport noLateExportSettings `json:"nolateexport"`
}

type noLateExportSettings struct {
	UnexportedFuncsLast bool `json:"unexported-funcs-last"`
}

type plugin struct {
	nolateexport *analysis.Analyzer
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		funcname.Analyzer,
		maxparams.Analyzer,
		nolateconst.Analyzer,
		p.nolateexport,
		nolocalstruct.Analyzer,
	}, nil
}

func (p *plugin) GetLoadMode() string {
	return register.LoadModeSyntax
}

func newPlugin(conf any) (register.LinterPlugin, error) {
	cfg, err := decodeConfig(conf)
	if err != nil {
		return nil, err
	}
	return &plugin{nolateexport: nolateexport.New(cfg)}, nil
}

func decodeConfig(conf any) (nolateexport.Config, error) {
	s, err := register.DecodeSettings[settings](conf)
	if err != nil {
		return nolateexport.Config{}, err
	}
	return nolateexport.Config{UnexportedFuncsLast: s.NoLateExport.UnexportedFuncsLast}, nil
}
