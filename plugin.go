// Package tidygo registers custom lint analyzers as a golangci-lint v2 module plugin.
package tidygo

import (
	"fmt"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"

	"github.com/lugassawan/tidygo/atomiccreate"
	"github.com/lugassawan/tidygo/busywait"
	"github.com/lugassawan/tidygo/confusingop"
	"github.com/lugassawan/tidygo/constcond"
	"github.com/lugassawan/tidygo/deferredcancel"
	"github.com/lugassawan/tidygo/emptycomment"
	"github.com/lugassawan/tidygo/emptyfunc"
	"github.com/lugassawan/tidygo/funcname"
	"github.com/lugassawan/tidygo/ifacename"
	"github.com/lugassawan/tidygo/ifinitonly"
	"github.com/lugassawan/tidygo/maxcaselines"
	"github.com/lugassawan/tidygo/maxcases"
	"github.com/lugassawan/tidygo/maxparams"
	"github.com/lugassawan/tidygo/methodcolocation"
	"github.com/lugassawan/tidygo/noanonnested"
	"github.com/lugassawan/tidygo/nolateconst"
	"github.com/lugassawan/tidygo/nolateexport"
	"github.com/lugassawan/tidygo/nolocalstruct"
	"github.com/lugassawan/tidygo/nonestedswitch"
	"github.com/lugassawan/tidygo/txrollback"
)

func init() {
	register.Plugin("tidygo", newPlugin)
}

// settings mirrors the golangci-lint `linters.settings.custom.tidygo.settings` block.
// Unknown keys are rejected by register.DecodeSettings, so typos fail loudly.
type settings struct {
	FuncName         funcNameSettings         `json:"funcname"`
	IfaceName        ifaceNameSettings        `json:"ifacename"`
	MaxCaseLines     maxSettings              `json:"maxcaselines"`
	MaxCases         maxSettings              `json:"maxcases"`
	MethodColocation methodColocationSettings `json:"methodcolocation"`
	NoAnonNested     maxSettings              `json:"noanonnested"`
	NoLateExport     noLateExportSettings     `json:"nolateexport"`
	NoLocalStruct    noLocalStructSettings    `json:"nolocalstruct"`
}

type funcNameSettings struct {
	IncludeVars bool `json:"include-vars"`
}

type ifaceNameSettings struct {
	AllowedNames []string `json:"allowed-names"`
}

// maxSettings is the settings block shared by the analyzers that take a single limit.
// Zero or unset selects the analyzer's default; negative values are rejected.
type maxSettings struct {
	Max int `json:"max"`
}

type methodColocationSettings struct {
	AllowFileSuffix bool `json:"allow-file-suffix"`
}

type noLateExportSettings struct {
	UnexportedFuncsLast bool `json:"unexported-funcs-last"`
}

type noLocalStructSettings struct {
	AllTypes bool `json:"all-types"`
}

// config holds the decoded per-analyzer configuration.
type config struct {
	funcname         funcname.Config
	ifacename        ifacename.Config
	maxcaselines     maxcaselines.Config
	maxcases         maxcases.Config
	methodcolocation methodcolocation.Config
	noanonnested     noanonnested.Config
	nolateexport     nolateexport.Config
	nolocalstruct    nolocalstruct.Config
}

type plugin struct {
	funcname         *analysis.Analyzer
	ifacename        *analysis.Analyzer
	maxcaselines     *analysis.Analyzer
	maxcases         *analysis.Analyzer
	methodcolocation *analysis.Analyzer
	noanonnested     *analysis.Analyzer
	nolateexport     *analysis.Analyzer
	nolocalstruct    *analysis.Analyzer
}

func (p *plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		atomiccreate.Analyzer,
		busywait.Analyzer,
		confusingop.Analyzer,
		constcond.Analyzer,
		deferredcancel.Analyzer,
		emptycomment.Analyzer,
		emptyfunc.Analyzer,
		p.funcname,
		p.ifacename,
		ifinitonly.Analyzer,
		p.maxcaselines,
		p.maxcases,
		maxparams.Analyzer,
		p.methodcolocation,
		nolateconst.Analyzer,
		p.nolateexport,
		p.noanonnested,
		p.nolocalstruct,
		nonestedswitch.Analyzer,
		txrollback.Analyzer,
	}, nil
}

// GetLoadMode asks for type information: deferredcancel, txrollback, busywait and
// atomiccreate resolve callees through pass.TypesInfo. golangci-lint skips every
// analyzer on a package that fails to type-check, so the syntax-only analyzers go
// quiet on non-compiling code too.
func (p *plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func newPlugin(conf any) (register.LinterPlugin, error) {
	cfg, err := decodeConfig(conf)
	if err != nil {
		return nil, err
	}
	return &plugin{
		funcname:         funcname.New(cfg.funcname),
		ifacename:        ifacename.New(cfg.ifacename),
		maxcaselines:     maxcaselines.New(cfg.maxcaselines),
		maxcases:         maxcases.New(cfg.maxcases),
		methodcolocation: methodcolocation.New(cfg.methodcolocation),
		noanonnested:     noanonnested.New(cfg.noanonnested),
		nolateexport:     nolateexport.New(cfg.nolateexport),
		nolocalstruct:    nolocalstruct.New(cfg.nolocalstruct),
	}, nil
}

// validateMax rejects negative limits so a typo does not silently fall back to the default.
func validateMax(s settings) error {
	limits := []struct {
		name string
		max  int
	}{
		{"maxcaselines", s.MaxCaseLines.Max},
		{"maxcases", s.MaxCases.Max},
		{"noanonnested", s.NoAnonNested.Max},
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
		funcname:         funcname.Config{IncludeVars: s.FuncName.IncludeVars},
		ifacename:        ifacename.Config{AllowedNames: s.IfaceName.AllowedNames},
		maxcaselines:     maxcaselines.Config{Max: s.MaxCaseLines.Max},
		maxcases:         maxcases.Config{Max: s.MaxCases.Max},
		methodcolocation: methodcolocation.Config{AllowFileSuffix: s.MethodColocation.AllowFileSuffix},
		noanonnested:     noanonnested.Config{Max: s.NoAnonNested.Max},
		nolateexport:     nolateexport.Config{UnexportedFuncsLast: s.NoLateExport.UnexportedFuncsLast},
		nolocalstruct:    nolocalstruct.Config{AllTypes: s.NoLocalStruct.AllTypes},
	}, nil
}
