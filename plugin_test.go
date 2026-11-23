package tidygo

import (
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/ifacename"
	"github.com/lugassawan/tidygo/nolateexport"
)

const analyzerName = "nolateexport"

const (
	funcNameKey      = "funcname"
	ifaceNameKey     = "ifacename"
	handlerIface     = "Handler"
	maxCasesKey      = "maxcases"
	maxCaseLinesKey  = "maxcaselines"
	maxKey           = "max"
	misspelledMaxKey = "maximum"
	noAnonNestedKey  = "noanonnested"
	noLocalStructKey = "nolocalstruct"
)

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

func TestDecodeConfigMaxCases(t *testing.T) {
	t.Run("max reaches the maxcases config", func(t *testing.T) {
		conf := map[string]any{maxCasesKey: map[string]any{maxKey: 3}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if got.maxcases.Max != 3 {
			t.Errorf("decodeConfig() maxcases = %+v, want Max=3", got.maxcases)
		}
	})

	t.Run("unknown maxcases key is rejected", func(t *testing.T) {
		conf := map[string]any{maxCasesKey: map[string]any{misspelledMaxKey: 3}}
		_, err := decodeConfig(conf)
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want unknown-field error")
		}
		if !strings.Contains(err.Error(), misspelledMaxKey) {
			t.Errorf("error %q does not name the unknown field", err)
		}
	})

	t.Run("negative max is rejected", func(t *testing.T) {
		conf := map[string]any{maxCasesKey: map[string]any{maxKey: -1}}
		_, err := decodeConfig(conf)
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want negative-max error")
		}
		if !strings.Contains(err.Error(), maxCasesKey) {
			t.Errorf("error %q does not name the setting", err)
		}
	})
}

func TestDecodeConfigMaxCaseLines(t *testing.T) {
	t.Run("max reaches the maxcaselines config", func(t *testing.T) {
		conf := map[string]any{maxCaseLinesKey: map[string]any{maxKey: 3}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if got.maxcaselines.Max != 3 {
			t.Errorf("decodeConfig() maxcaselines = %+v, want Max=3", got.maxcaselines)
		}
		if got.maxcases.Max != 0 {
			t.Errorf("decodeConfig() maxcases = %+v, want zero config", got.maxcases)
		}
	})

	t.Run("unknown maxcaselines key is rejected", func(t *testing.T) {
		conf := map[string]any{maxCaseLinesKey: map[string]any{misspelledMaxKey: 3}}
		_, err := decodeConfig(conf)
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want unknown-field error")
		}
		if !strings.Contains(err.Error(), misspelledMaxKey) {
			t.Errorf("error %q does not name the unknown field", err)
		}
	})

	t.Run("negative max is rejected", func(t *testing.T) {
		conf := map[string]any{maxCaseLinesKey: map[string]any{maxKey: -1}}
		_, err := decodeConfig(conf)
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want negative-max error")
		}
		if !strings.Contains(err.Error(), maxCaseLinesKey) {
			t.Errorf("error %q does not name the setting", err)
		}
	})
}

func TestBuildAnalyzersIncludesIfinitonly(t *testing.T) {
	p, err := newPlugin(nil)
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	for _, a := range analyzers {
		if a.Name == "ifinitonly" {
			return
		}
	}
	t.Fatal("BuildAnalyzers() has no ifinitonly analyzer")
}

func TestBuildAnalyzersIncludesNoNestedSwitch(t *testing.T) {
	p, err := newPlugin(nil)
	if err != nil {
		t.Fatalf("newPlugin() error = %v", err)
	}
	analyzers, err := p.BuildAnalyzers()
	if err != nil {
		t.Fatalf("BuildAnalyzers() error = %v", err)
	}
	for _, a := range analyzers {
		if a.Name == "nonestedswitch" {
			return
		}
	}
	t.Fatal("BuildAnalyzers() has no nonestedswitch analyzer")
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

// buildNamed builds the plugin from conf and returns the analyzer called name.
func buildNamed(t *testing.T, conf map[string]any, name string) *analysis.Analyzer {
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
	t.Fatalf("BuildAnalyzers() has no %s analyzer", name)
	return nil
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
