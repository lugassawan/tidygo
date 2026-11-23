// Package constcond defines an analyzer that forbids if and for statements whose
// condition is the literal true or false.
package constcond

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports if/for statements whose condition is the literal true or false.
var Analyzer = &analysis.Analyzer{
	Name:     "constcond",
	Doc:      "forbids if and for statements whose condition is the literal true or false",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{
		(*ast.IfStmt)(nil),
		(*ast.ForStmt)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch stmt := n.(type) {
		case *ast.IfStmt:
			if lit, ok := boolLiteral(stmt.Cond); ok {
				pass.Reportf(
					stmt.Cond.Pos(),
					"condition is always %s; remove the branch or use a real condition",
					lit,
				)
			}
		case *ast.ForStmt:
			if lit, ok := boolLiteral(stmt.Cond); ok {
				pass.Reportf(
					stmt.Cond.Pos(),
					"loop condition is always %s; remove the loop or use a real condition",
					lit,
				)
			}
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// boolLiteral returns "true" or "false" when cond, ignoring parentheses, is
// that identifier. The analyzer uses no type information, so it matches on
// the name only: a shadowed true or false is still reported.
func boolLiteral(cond ast.Expr) (string, bool) {
	if cond == nil {
		return "", false
	}
	id, ok := ast.Unparen(cond).(*ast.Ident)
	if !ok || (id.Name != "true" && id.Name != "false") {
		return "", false
	}
	return id.Name, true
}
