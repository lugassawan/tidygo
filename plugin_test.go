package tidygo

import (
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/nolateexport"
)

const analyzerName = "nolateexport"

func TestDecodeConfig(t *testing.T) {
	t.Run("nil settings give the zero config", func(t *testing.T) {
		got, err := decodeConfig(nil)
		if err != nil {
			t.Fatalf("decodeConfig(nil) error = %v", err)
		}
		if got != (nolateexport.Config{}) {
			t.Errorf("decodeConfig(nil) = %+v, want zero config", got)
		}
	})

	t.Run("unexported-funcs-last turns the option on", func(t *testing.T) {
		conf := map[string]any{analyzerName: map[string]any{"unexported-funcs-last": true}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if !got.UnexportedFuncsLast {
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
