package nolateexport

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"golang.org/x/tools/go/analysis"
)

// Receivers that cannot be resolved to a local type name (e.g. pkg.T) must not
// share a key, or unrelated types would be cross-reported.
func TestRunSkipsUnresolvableReceivers(t *testing.T) {
	file := parseSource(t, `package p
func (a pkg.T) helper() {}
func (a pkg.T) Exported() {}
func (b other.U) helper() {}
func (b other.U) Exported() {}
`)
	// A method with no receiver entries is malformed and rejected by the type
	// checker, so analysistest fixtures cannot reach this guard.
	file.Decls = append(file.Decls, &ast.FuncDecl{Name: ast.NewIdent("Orphan"), Recv: &ast.FieldList{}})

	for _, cfg := range []Config{{}, {UnexportedFuncsLast: true}} {
		var got []analysis.Diagnostic
		pass := &analysis.Pass{
			Files:  []*ast.File{file},
			Report: func(d analysis.Diagnostic) { got = append(got, d) },
		}
		if _, err := run(pass, cfg); err != nil {
			t.Fatalf("run() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("run(%+v) reported %d diagnostics, want none: %v", cfg, len(got), got)
		}
	}
}

func parseSource(t *testing.T, src string) *ast.File {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "p.go", src, 0)
	if err != nil {
		t.Fatalf("parsing source: %v", err)
	}
	return file
}
