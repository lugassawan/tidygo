// Package tidygo registers custom lint analyzers as a golangci-lint v2 module plugin.
package tidygo

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/lugassawan/tidygo/funcname"
	"github.com/lugassawan/tidygo/ifacename"
	"github.com/lugassawan/tidygo/ifinitonly"
	"github.com/lugassawan/tidygo/maxcases"
	"github.com/lugassawan/tidygo/maxparams"
	"github.com/lugassawan/tidygo/nolateconst"
	"github.com/lugassawan/tidygo/nolateexport"
	"github.com/lugassawan/tidygo/nolocalstruct"
	"github.com/lugassawan/tidygo/nonestedswitch"
)

func init() {
	register.Plugin("tidygo", newPlugin)
}

// settings mirrors the golangci-lint `linters.settings.custom.tidygo.settings` block.
// Unknown keys are rejected by register.DecodeSettings, so typos fail loudly.
type settings struct {
	IfaceName    ifaceNameSettings    `json:"ifacename"`
	MaxCases     maxSettings          `json:"maxcases"`
	NoLateExport noLateExportSettings `json:"nolateexport"`
}

type ifaceNameSettings struct {
	AllowedNames []string `json:"allowed-names"`
}

// maxSettings is the settings block shared by the analyzers that take a single limit.
// Zero or unset selects the analyzer's default; negative values are rejected.
type maxSettings struct {
	Max int `json:"max"`
}

type noLateExportSettings struct {
	UnexportedFuncsLast bool `json:"unexported-funcs-last"`
}

// config holds the decoded per-analyzer configuration.
type config struct {
	ifacename    ifacename.Config
	maxcases     maxcases.Config
	nolateexport nolateexport.Config
}

type plugin struct {
	ifacename    *analysis.Analyzer
	maxcases     *analysis.Analyzer
	nolateexport *analysis.Analyzer
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		funcname.Analyzer,
		p.ifacename,
		ifinitonly.Analyzer,
		p.maxcases,
		maxparams.Analyzer,
		nolateconst.Analyzer,
		p.nolateexport,
		nolocalstruct.Analyzer,
		nonestedswitch.Analyzer,
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
		maxcases:     maxcases.New(cfg.maxcases),
		nolateexport: nolateexport.New(cfg.nolateexport),
	}, nil
}

func decodeConfig(conf any) (config, error) {
	s, err := register.DecodeSettings[settings](conf)
	if err != nil {
		return config{}, err
	}
	if s.MaxCases.Max < 0 {
		return config{}, fmt.Errorf("maxcases.max must not be negative, got %d", s.MaxCases.Max)
	}
	return config{
		ifacename:    ifacename.Config{AllowedNames: s.IfaceName.AllowedNames},
		maxcases:     maxcases.Config{Max: s.MaxCases.Max},
		nolateexport: nolateexport.Config{UnexportedFuncsLast: s.NoLateExport.UnexportedFuncsLast},
	}, nil
}
