package confusingop_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/confusingop"
)

// fixture is written at test time because gofmt rewrites "x =+ 1" to "x = +1",
// which would erase the cases under test.
const fixture = `package example

func typos(b, d bool) {
	x := 1
	var a, c int

	// Expectations below use "." and [x] instead of quotes and escapes so
	// they survive both this raw string and analysistest's parsing.
	x =+ 1 // want ".=[+]. looks like a typo for .[+]=."
	x =- 1 // want ".=[-]. looks like a typo for .[-]=."
	b =!b  // want ".=[!]. looks like a typo for .[!]=."
	x =^ 1 // want ".=[\\^]. looks like a typo for .[\\^]=."

	a, c =-1, 2 // want ".=[-]. looks like a typo for .[-]=."

	// The unary operator can be the leftmost operand of a binary expression.
	x =+ a * 2 // want ".=[+]. looks like a typo for .[+]=."
	x =-a + 1  // want ".=[-]. looks like a typo for .[-]=."
	b =!b && d // want ".=[!]. looks like a typo for .[!]=."

	x = +1
	x = -x
	x += 1
	x -= 1
	y := -1
	x =
		+1
	a, c = 1, -2
	x = +a * 2
	x = a + -1
	x = a * -1

	// Compact code with no space before "=" is deliberate, not a typo.
	x=-1
	b=!b
	a=+1

	_, _, _, _, _ = x, y, a, c, d
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
