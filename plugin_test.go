package tidygo

import (
	"strings"
	"testing"

	"github.com/lugassawan/tidygo/nolateexport"
)

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
		conf := map[string]any{"nolateexport": map[string]any{"unexported-funcs-last": true}}
		got, err := decodeConfig(conf)
		if err != nil {
			t.Fatalf("decodeConfig() error = %v", err)
		}
		if !got.UnexportedFuncsLast {
			t.Errorf("decodeConfig() = %+v, want UnexportedFuncsLast=true", got)
		}
	})

	t.Run("unknown key is rejected", func(t *testing.T) {
		conf := map[string]any{"nolateexport": map[string]any{"unexported-func-last": true}}
		_, err := decodeConfig(conf)
		if err == nil {
			t.Fatal("decodeConfig() error = nil, want unknown-field error")
		}
		if !strings.Contains(err.Error(), "unexported-func-last") {
			t.Errorf("error %q does not name the unknown field", err)
		}
	})
}

func TestNewPluginPropagatesDecodeError(t *testing.T) {
	if _, err := newPlugin(map[string]any{"bogus": true}); err == nil {
		t.Fatal("newPlugin() error = nil, want error for unknown setting")
	}
}
