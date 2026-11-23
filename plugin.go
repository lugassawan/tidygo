// Package tidygo registers custom lint analyzers as a golangci-lint v2 module plugin.
package tidygo

import (
	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/lugassawan/tidygo/confusingop"
	"github.com/lugassawan/tidygo/constcond"
	"github.com/lugassawan/tidygo/emptycomment"
	"github.com/lugassawan/tidygo/funcname"
	"github.com/lugassawan/tidygo/ifacename"
	"github.com/lugassawan/tidygo/ifinitonly"
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
	IfaceName    ifaceNameSettings    `json:"ifacename"`
	NoLateExport noLateExportSettings `json:"nolateexport"`
}

type ifaceNameSettings struct {
	AllowedNames []string `json:"allowed-names"`
}

type noLateExportSettings struct {
	UnexportedFuncsLast bool `json:"unexported-funcs-last"`
}

// config holds the decoded per-analyzer configuration.
type config struct {
	ifacename    ifacename.Config
	nolateexport nolateexport.Config
}

type plugin struct {
	ifacename    *analysis.Analyzer
	nolateexport *analysis.Analyzer
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		confusingop.Analyzer,
		constcond.Analyzer,
		emptycomment.Analyzer,
		funcname.Analyzer,
		p.ifacename,
		ifinitonly.Analyzer,
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
	return &plugin{
		ifacename:    ifacename.New(cfg.ifacename),
		nolateexport: nolateexport.New(cfg.nolateexport),
	}, nil
}

func decodeConfig(conf any) (config, error) {
	s, err := register.DecodeSettings[settings](conf)
	if err != nil {
		return config{}, err
	}
	return config{
		ifacename:    ifacename.Config{AllowedNames: s.IfaceName.AllowedNames},
		nolateexport: nolateexport.Config{UnexportedFuncsLast: s.NoLateExport.UnexportedFuncsLast},
	}, nil
}
