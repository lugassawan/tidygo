// Package tidygo registers custom lint analyzers as a golangci-lint v2 module plugin.
package tidygo

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/lugassawan/tidygo/confusingop"
	"github.com/lugassawan/tidygo/constcond"
	"github.com/lugassawan/tidygo/emptycomment"
	"github.com/lugassawan/tidygo/emptyfunc"
	"github.com/lugassawan/tidygo/funcname"
	"github.com/lugassawan/tidygo/ifacename"
	"github.com/lugassawan/tidygo/ifinitonly"
	"github.com/lugassawan/tidygo/maxboolops"
	"github.com/lugassawan/tidygo/maxcaselines"
	"github.com/lugassawan/tidygo/maxcases"
	"github.com/lugassawan/tidygo/maxparams"
	"github.com/lugassawan/tidygo/nolateconst"
	"github.com/lugassawan/tidygo/nolateexport"
	"github.com/lugassawan/tidygo/nolocalstruct"
	"github.com/lugassawan/tidygo/nonestedswitch"
	"github.com/lugassawan/tidygo/redundantparen"
)

func init() {
	register.Plugin("tidygo", newPlugin)
}

// settings mirrors the golangci-lint `linters.settings.custom.tidygo.settings` block.
// Unknown keys are rejected by register.DecodeSettings, so typos fail loudly.
type settings struct {
	IfaceName    ifaceNameSettings    `json:"ifacename"`
	MaxBoolOps   maxSettings          `json:"maxboolops"`
	MaxCaseLines maxSettings          `json:"maxcaselines"`
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
	maxboolops   maxboolops.Config
	maxcaselines maxcaselines.Config
	maxcases     maxcases.Config
	nolateexport nolateexport.Config
}

type plugin struct {
	ifacename    *analysis.Analyzer
	maxboolops   *analysis.Analyzer
	maxcaselines *analysis.Analyzer
	maxcases     *analysis.Analyzer
	nolateexport *analysis.Analyzer
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		confusingop.Analyzer,
		constcond.Analyzer,
		emptycomment.Analyzer,
		emptyfunc.Analyzer,
		funcname.Analyzer,
		p.ifacename,
		ifinitonly.Analyzer,
		p.maxboolops,
		p.maxcaselines,
		p.maxcases,
		maxparams.Analyzer,
		nolateconst.Analyzer,
		p.nolateexport,
		nolocalstruct.Analyzer,
		nonestedswitch.Analyzer,
		redundantparen.Analyzer,
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
		maxboolops:   maxboolops.New(cfg.maxboolops),
		maxcaselines: maxcaselines.New(cfg.maxcaselines),
		maxcases:     maxcases.New(cfg.maxcases),
		nolateexport: nolateexport.New(cfg.nolateexport),
	}, nil
}

// validateMax rejects negative limits so a typo does not silently fall back to the default.
func validateMax(s settings) error {
	limits := []struct {
		name string
		max  int
	}{
		{"maxboolops", s.MaxBoolOps.Max},
		{"maxcaselines", s.MaxCaseLines.Max},
		{"maxcases", s.MaxCases.Max},
	}
	for _, l := range limits {
		if l.max < 0 {
			return fmt.Errorf("%s.max must not be negative, got %d", l.name, l.max)
		}
	}
	return nil
}

func decodeConfig(conf any) (config, error) {
	s, err := register.DecodeSettings[settings](conf)
	if err != nil {
		return config{}, err
	}
	if err := validateMax(s); err != nil {
		return config{}, err
	}
	return config{
		ifacename:    ifacename.Config{AllowedNames: s.IfaceName.AllowedNames},
		maxboolops:   maxboolops.Config{Max: s.MaxBoolOps.Max},
		maxcaselines: maxcaselines.Config{Max: s.MaxCaseLines.Max},
		maxcases:     maxcases.Config{Max: s.MaxCases.Max},
		nolateexport: nolateexport.Config{UnexportedFuncsLast: s.NoLateExport.UnexportedFuncsLast},
	}, nil
}
