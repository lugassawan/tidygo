package confusingop_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/confusingop"
)

// fixture is written to a temp dir instead of living under testdata/ because
// gofmt rewrites "x =+ 1" to "x = +1", which would silently erase the cases
// under test whenever the pre-commit hook or `make fmt` formats the tree.
//
// The want patterns use "." and character classes instead of quotes and
// escapes so they survive both this raw string and analysistest's parsing.
const fixture = `package example

func typos(b bool) {
	x := 1
	var a, c int

	x =+ 1 // want ".=[+]. looks like a typo for .[+]=."
	x =- 1 // want ".=[-]. looks like a typo for .[-]=."
	b =!b  // want ".=[!]. looks like a typo for .[!]=."
	x =^ 1 // want ".=[\\^]. looks like a typo for .[\\^]=."

	a, c =-1, 2 // want ".=[-]. looks like a typo for .[-]=."

	x = +1
	x = -x
	x += 1
	x -= 1
	y := -1
	x =
		+1
	a, c = 1, -2

	_, _, _, _ = x, y, a, c
}
`

func TestAnalyzer(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "src", "example")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("creating fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "example.go"), []byte(fixture), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	analysistest.Run(t, dir, confusingop.Analyzer, "example")
}
