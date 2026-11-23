// Package nolocalstruct defines an analyzer that forbids named struct type
// declarations inside function bodies. Structs should be declared at package level.
package nolocalstruct

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports named struct types declared inside function bodies.
var Analyzer = &analysis.Analyzer{
	Name:     "nolocalstruct",
	Doc:      "forbids named struct type declarations inside function bodies",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	// A DeclStmt only occurs inside a function body, so filtering on it visits
	// every local declaration exactly once, including those in nested literals.
	nodeFilter := []ast.Node{(*ast.DeclStmt)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		ds, ok := n.(*ast.DeclStmt)
		if !ok {
			return
		}
		gd, ok := ds.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			return
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, isStruct := ts.Type.(*ast.StructType); isStruct {
				pass.Reportf(
					ts.Pos(),
					"named struct type %q should be declared at package level, not inside a function",
					ts.Name.Name,
				)
			}
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}
