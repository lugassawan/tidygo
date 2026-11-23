package constcond_test

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/lugassawan/tidygo/constcond"
)

// parenFixture is written to a temp dir instead of living under testdata/
// because gofmt strips the parentheses from "if (true)", which would silently
// turn the case into the plain one whenever the tree is formatted.
const parenFixture = `package example

func parenthesized(ok bool) {
	if (true) { // want "condition is always true"
		_ = ok
	}

	for ((false)) { // want "loop condition is always false"
		_ = ok
	}

	if (ok) {
		_ = ok
	}
}
`

func TestAnalyzer(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, constcond.Analyzer, "example")
}

func TestAnalyzerParenthesizedConditions(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "src", "example")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("creating fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "example.go"), []byte(parenFixture), 0o600); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}
	analysistest.Run(t, dir, constcond.Analyzer, "example")
}
