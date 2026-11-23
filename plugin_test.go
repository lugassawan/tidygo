package tidygo

import (
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/ifacename"
	"github.com/lugassawan/tidygo/nolateexport"
)

const analyzerName = "nolateexport"

const (
	dupFuncKey       = "dupfunc"
	elseIfElseKey    = "elseifelse"
	enabledKey       = "enabled"
	funcNameKey      = "funcname"
	ifaceNameKey     = "ifacename"
	handlerIface     = "Handler"
	maxBoolOpsKey    = "maxboolops"
	maxCasesKey      = "maxcases"
	maxCaseLinesKey  = "maxcaselines"
	maxKey           = "max"
	minStatementsKey = "min-statements"
	misspelledMaxKey = "maximum"
	noAnonNestedKey  = "noanonnested"
	noLocalStructKey = "nolocalstruct"

	methodColocationKey = "methodcolocation"
	allowFileSuffixKey  = "allow-file-suffix"
)

// maxLimits lists the analyzers that share the single-limit settings block.
var maxLimits = []struct {
	key string
	get func(config) int
}{
	{maxBoolOpsKey, func(c config) int { return c.maxboolops.Max }},
	{maxCaseLinesKey, func(c config) int { return c.maxcaselines.Max }},
	{maxCasesKey, func(c config) int { return c.maxcases.Max }},
}

func TestDecodeConfig(t *testing.T) {
	t.Run("nil settings give the zero config", func(t *testing.T) {
		got, err := decodeConfig(nil)
		if err != nil {
			t.Fatalf("decodeConfig(nil) error = %v", err)
		}
		if !reflect.DeepEqual(got, config{}) {
			t.Errorf("decodeConfig(nil) = %+v, want zero config", got)
		}
	})

	t.Run("unexported-funcs-last turns the option on", func(t *testing.T) {
		conf := map[string]any{analyzerName: map[string]any{"unexported-funcs-last": true}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if !got.nolateexport.UnexportedFuncsLast {
			t.Errorf("decodeConfig() = %+v, want UnexportedFuncsLast=true", got)
		}
	})

	t.Run("unknown key is rejected", func(t *testing.T) {
		conf := map[string]any{analyzerName: map[string]any{"unexported-func-last": true}}
		_, err := decodeConfig(conf)
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want unknown-field error")
		}
		if !strings.Contains(err.Error(), "unexported-func-last") {
			t.Errorf("error %q does not name the unknown field", err)
		}
	})
}

func TestDecodeConfigIfaceName(t *testing.T) {
	t.Run("allowed-names reaches the ifacename config", func(t *testing.T) {
		conf := map[string]any{ifaceNameKey: map[string]any{"allowed-names": []any{handlerIface, "Reader"}}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		want := ifacename.Config{AllowedNames: []string{handlerIface, "Reader"}}
		if !slices.Equal(got.ifacename.AllowedNames, want.AllowedNames) {
			t.Errorf("decodeConfig() ifacename = %+v, want %+v", got.ifacename, want)
		}
		if got.nolateexport != (nolateexport.Config{}) {
			t.Errorf("decodeConfig() nolateexport = %+v, want zero config", got.nolateexport)
		}
	})

	t.Run("unknown ifacename key is rejected", func(t *testing.T) {
		conf := map[string]any{ifaceNameKey: map[string]any{"allowed-name": []any{handlerIface}}}
		_, err := decodeConfig(conf)
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want unknown-field error")
		}
		if !strings.Contains(err.Error(), "allowed-name") {
			t.Errorf("error %q does not name the unknown field", err)
		}
	})
}

func TestDecodeConfigMaxReachesConfig(t *testing.T) {
	for _, l := range maxLimits {
		t.Run(l.key, func(t *testing.T) {
			got, err := decodeConfig(map[string]any{l.key: map[string]any{maxKey: 3}})
			if err != nil {
				t.Fatalf("decodeConfig() error = %v", err)
			}
			if v := l.get(got); v != 3 {
				t.Errorf("decodeConfig() %s max = %d, want 3", l.key, v)
			}
			// Setting one limit must not leak into the others.
			for _, other := range maxLimits {
				if v := other.get(got); other.key != l.key && v != 0 {
					t.Errorf("decodeConfig() %s max = %d, want 0 when only %s is set", other.key, v, l.key)
				}
			}
		})
	}
}

func TestDecodeConfigMaxRejectsUnknownKey(t *testing.T) {
	for _, l := range maxLimits {
		t.Run(l.key, func(t *testing.T) {
			_, err := decodeConfig(map[string]any{l.key: map[string]any{"maximum": 3}})
			if err == nil {
				t.Fatal("decodeConfig() error = nil, want unknown-field error")
			}
			if !strings.Contains(err.Error(), "maximum") {
				t.Errorf("error %q does not name the unknown field", err)
			}
		})
	}
}

func TestDecodeConfigMaxRejectsNegative(t *testing.T) {
	for _, l := range maxLimits {
		t.Run(l.key, func(t *testing.T) {
			_, err := decodeConfig(map[string]any{l.key: map[string]any{maxKey: -1}})
			if err == nil {
				t.Fatal("decodeConfig() error = nil, want negative-max error")
			}
			if !strings.Contains(err.Error(), l.key) {
				t.Errorf("error %q does not name the setting", err)
			}
		})
	}
}

func TestDecodeConfigElseIfElse(t *testing.T) {
	t.Run("enabled turns the analyzer on", func(t *testing.T) {
		got, err := decodeConfig(map[string]any{elseIfElseKey: map[string]any{enabledKey: true}})
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if !got.elseifelse {
			t.Errorf("decodeConfig() = %+v, want elseifelse=true", got)
		}
	})

	t.Run("unknown key is rejected", func(t *testing.T) {
		_, err := decodeConfig(map[string]any{elseIfElseKey: map[string]any{"enable": true}})
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want unknown-field error")
		}
		if !strings.Contains(err.Error(), "enable") {
			t.Errorf("error %q does not name the unknown field", err)
		}
	})
}

func TestBuildAnalyzersElseIfElseIsOptIn(t *testing.T) {
	t.Run("absent by default", func(t *testing.T) {
		if got := buildAnalyzer(t, nil, elseIfElseKey); got != nil {
			t.Errorf("BuildAnalyzers() has %s without it being enabled", got.Name)
		}
	})

	t.Run("absent when disabled explicitly", func(t *testing.T) {
		conf := map[string]any{elseIfElseKey: map[string]any{enabledKey: false}}
		if got := buildAnalyzer(t, conf, elseIfElseKey); got != nil {
			t.Errorf("BuildAnalyzers() has %s while it is disabled", got.Name)
		}
	})

	t.Run("present and working when enabled", func(t *testing.T) {
		conf := map[string]any{elseIfElseKey: map[string]any{enabledKey: true}}
		got := buildAnalyzer(t, conf, elseIfElseKey)
		if got == nil {
			t.Fatal("BuildAnalyzers() has no elseifelse analyzer")
		}
		testdata, err := filepath.Abs(filepath.Join(elseIfElseKey, "testdata"))
		if err != nil {
			t.Fatalf("resolving testdata: %v", err)
		}
		analysistest.Run(t, testdata, got, "example")
	})
}

func TestBuildAnalyzersRegistersAnalyzers(t *testing.T) {
	p, err := newPlugin(nil)
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	registered := make(map[string]bool, len(analyzers))
	for _, a := range analyzers {
		registered[a.Name] = true
	}
	for _, name := range []string{
		"atomiccreate",
		"busywait",
		"confusingop",
		"constcond",
		"deferredcancel",
		"dupfunc",
		"emptycomment",
		"emptyfunc",
		"ifinitonly",
		"maxboolops",
		"methodcolocation",
		"nonestedswitch",
		"redundantparen",
		"txrollback",
	} {
		t.Run(name, func(t *testing.T) {
			if !registered[name] {
				t.Errorf("BuildAnalyzers() has no %s analyzer", name)
			}
		})
	}
}

// The type-aware analyzers read pass.TypesInfo, which golangci-lint only fills in
// when the plugin asks for the types-info load mode.
func TestGetLoadModeIsTypesInfo(t *testing.T) {
	p, err := newPlugin(nil)
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	if got := p.GetLoadMode(); got != register.LoadModeTypesInfo {
		t.Errorf("GetLoadMode() = %q, want %q", got, register.LoadModeTypesInfo)
	}
}

func TestNewPluginPropagatesDecodeError(t *testing.T) {
	if _, err := newPlugin(map[string]any{"bogus": true}); err == nil {
		t.Fatal("newPlugin() error = nil, want error for unknown setting")
	}
}

func TestNewPluginWiresConfigIntoAnalyzer(t *testing.T) {
	p, err := newPlugin(map[string]any{analyzerName: map[string]any{"unexported-funcs-last": true}})
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	var got *analysis.Analyzer
	for _, a := range analyzers {
		if a.Name == analyzerName {
			got = a
		}
	}
	if got == nil {
		t.Fatal("BuildAnalyzers() has no nolateexport analyzer")
	}
	// funcslast only passes if unexported-funcs-last reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join("nolateexport", "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "funcslast")
}

func TestNewPluginWiresMaxCasesConfigIntoAnalyzer(t *testing.T) {
	p, err := newPlugin(map[string]any{maxCasesKey: map[string]any{maxKey: 3}})
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	var got *analysis.Analyzer
	for _, a := range analyzers {
		if a.Name == maxCasesKey {
			got = a
		}
	}
	if got == nil {
		t.Fatal("BuildAnalyzers() has no maxcases analyzer")
	}
	// The custom package only passes if max reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join("maxcases", "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "custom")
}

func TestNewPluginWiresMaxBoolOpsConfigIntoAnalyzer(t *testing.T) {
	got := buildAnalyzer(t, map[string]any{maxBoolOpsKey: map[string]any{maxKey: 1}}, maxBoolOpsKey)
	if got == nil {
		t.Fatal("BuildAnalyzers() has no maxboolops analyzer")
	}
	// The custom package only passes if max reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join(maxBoolOpsKey, "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "custom")
}

func TestNewPluginWiresMaxCaseLinesConfigIntoAnalyzer(t *testing.T) {
	p, err := newPlugin(map[string]any{maxCaseLinesKey: map[string]any{maxKey: 3}})
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	var got *analysis.Analyzer
	for _, a := range analyzers {
		if a.Name == maxCaseLinesKey {
			got = a
		}
	}
	if got == nil {
		t.Fatal("BuildAnalyzers() has no maxcaselines analyzer")
	}
	// The custom package only passes if max reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join(maxCaseLinesKey, "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "custom")
}

func TestNewPluginWiresIfaceNameConfigIntoAnalyzer(t *testing.T) {
	p, err := newPlugin(map[string]any{ifaceNameKey: map[string]any{"allowed-names": []any{handlerIface}}})
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	var got *analysis.Analyzer
	for _, a := range analyzers {
		if a.Name == ifaceNameKey {
			got = a
		}
	}
	if got == nil {
		t.Fatal("BuildAnalyzers() has no ifacename analyzer")
	}
	// The allowed package only passes if allowed-names reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join("ifacename", "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "allowed")
}

func TestDecodeConfigNoLocalStruct(t *testing.T) {
	t.Run("all-types reaches the nolocalstruct config", func(t *testing.T) {
		conf := map[string]any{noLocalStructKey: map[string]any{"all-types": true}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if !got.nolocalstruct.AllTypes {
			t.Errorf("decodeConfig() nolocalstruct = %+v, want AllTypes=true", got.nolocalstruct)
		}
	})

	t.Run("unknown nolocalstruct key is rejected", func(t *testing.T) {
		requireUnknownKey(t, map[string]any{noLocalStructKey: map[string]any{"all-type": true}}, "all-type")
	})
}

func TestNewPluginWiresNoLocalStructConfigIntoAnalyzer(t *testing.T) {
	got := buildNamed(t, map[string]any{noLocalStructKey: map[string]any{"all-types": true}}, noLocalStructKey)
	// The alltypes package only passes if all-types reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join(noLocalStructKey, "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "alltypes")
}

func TestDecodeConfigNoAnonNested(t *testing.T) {
	t.Run("max reaches the noanonnested config", func(t *testing.T) {
		conf := map[string]any{noAnonNestedKey: map[string]any{maxKey: 2}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if got.noanonnested.Max != 2 {
			t.Errorf("decodeConfig() noanonnested = %+v, want Max=2", got.noanonnested)
		}
	})

	t.Run("unknown noanonnested key is rejected", func(t *testing.T) {
		requireUnknownKey(t, map[string]any{noAnonNestedKey: map[string]any{misspelledMaxKey: 2}}, misspelledMaxKey)
	})

	t.Run("negative max is rejected", func(t *testing.T) {
		_, err := decodeConfig(map[string]any{noAnonNestedKey: map[string]any{maxKey: -1}})
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want negative-max error")
		}
		if !strings.Contains(err.Error(), noAnonNestedKey) {
			t.Errorf("error %q does not name the setting", err)
		}
	})
}

func TestNewPluginWiresNoAnonNestedConfigIntoAnalyzer(t *testing.T) {
	got := buildNamed(t, map[string]any{noAnonNestedKey: map[string]any{maxKey: 2}}, noAnonNestedKey)
	// The custom package only passes if max reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join(noAnonNestedKey, "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "custom")
}

func TestDecodeConfigFuncName(t *testing.T) {
	t.Run("include-vars reaches the funcname config", func(t *testing.T) {
		conf := map[string]any{funcNameKey: map[string]any{"include-vars": true}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if !got.funcname.IncludeVars {
			t.Errorf("decodeConfig() funcname = %+v, want IncludeVars=true", got.funcname)
		}
	})

	t.Run("unknown funcname key is rejected", func(t *testing.T) {
		requireUnknownKey(t, map[string]any{funcNameKey: map[string]any{"include-var": true}}, "include-var")
	})
}

func TestNewPluginWiresFuncNameConfigIntoAnalyzer(t *testing.T) {
	got := buildNamed(t, map[string]any{funcNameKey: map[string]any{"include-vars": true}}, funcNameKey)
	// The includevars package only passes if include-vars reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join(funcNameKey, "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "includevars")
}

func TestDecodeConfigDupFunc(t *testing.T) {
	t.Run("min-statements reaches the dupfunc config", func(t *testing.T) {
		got, err := decodeConfig(map[string]any{dupFuncKey: map[string]any{minStatementsKey: 2}})
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if got.dupfunc.MinStmts != 2 {
			t.Errorf("decodeConfig() dupfunc = %+v, want MinStmts=2", got.dupfunc)
		}
	})

	t.Run("unknown dupfunc key is rejected", func(t *testing.T) {
		requireUnknownKey(t, map[string]any{dupFuncKey: map[string]any{"min-statement": 2}}, "min-statement")
	})

	t.Run("negative min-statements is rejected", func(t *testing.T) {
		_, err := decodeConfig(map[string]any{dupFuncKey: map[string]any{minStatementsKey: -1}})
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want negative-value error")
		}
		if !strings.Contains(err.Error(), dupFuncKey) {
			t.Errorf("error %q does not name the setting", err)
		}
	})
}

func TestNewPluginWiresDupFuncConfigIntoAnalyzer(t *testing.T) {
	got := buildNamed(t, map[string]any{dupFuncKey: map[string]any{minStatementsKey: 2}}, dupFuncKey)
	// The custom package only passes if min-statements reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join(dupFuncKey, "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "custom")
}

func TestDecodeConfigMethodColocation(t *testing.T) {
	t.Run("allow-file-suffix defaults to off", func(t *testing.T) {
		got, err := decodeConfig(nil)
		if err != nil {
			t.Fatalf("decodeConfig(nil) error = %v", err)
		}
		if got.methodcolocation.AllowFileSuffix {
			t.Errorf("decodeConfig(nil) = %+v, want AllowFileSuffix=false", got.methodcolocation)
		}
	})

	t.Run("allow-file-suffix turns the option on", func(t *testing.T) {
		conf := map[string]any{methodColocationKey: map[string]any{allowFileSuffixKey: true}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if !got.methodcolocation.AllowFileSuffix {
			t.Errorf("decodeConfig() = %+v, want AllowFileSuffix=true", got.methodcolocation)
		}
	})

	t.Run("unknown methodcolocation key is rejected", func(t *testing.T) {
		conf := map[string]any{methodColocationKey: map[string]any{"allow-file-suffixes": true}}
		requireUnknownKey(t, conf, "allow-file-suffixes")
	})
}

func TestNewPluginWiresMethodColocationConfigIntoAnalyzer(t *testing.T) {
	conf := map[string]any{methodColocationKey: map[string]any{allowFileSuffixKey: true}}
	got := buildNamed(t, conf, methodColocationKey)
	// The suffix package only passes if allow-file-suffix reached the analyzer.
	testdata, err := filepath.Abs(filepath.Join(methodColocationKey, "testdata"))
	if err != nil {
		t.Fatalf("resolving testdata: %v", err)
	}
	analysistest.Run(t, testdata, got, "suffix")
}

// buildAnalyzer builds the plugin from conf and returns the analyzer called name, or nil if it is not registered.
func buildAnalyzer(t *testing.T, conf any, name string) *analysis.Analyzer {
	t.Helper()
	p, err := newPlugin(conf)
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	for _, a := range analyzers {
		if a.Name == name {
			return a
		}
	}
	return nil
}

// buildNamed is buildAnalyzer that fails the test when the analyzer is not registered.
func buildNamed(t *testing.T, conf map[string]any, name string) *analysis.Analyzer {
	t.Helper()
	got := buildAnalyzer(t, conf, name)
	if got == nil {
		t.Fatalf("BuildAnalyzers() has no %s analyzer", name)
	}
	return got
}

// requireUnknownKey asserts that decoding conf fails and names the offending key.
func requireUnknownKey(t *testing.T, conf map[string]any, key string) {
	t.Helper()
	_, err := decodeConfig(conf)
	if err == nil {
		t.Fatal("decodeConfig() error = nil, want unknown-field error")
	}
	if !strings.Contains(err.Error(), key) {
		t.Errorf("error %q does not name the unknown field", err)
	}
}
