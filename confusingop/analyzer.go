// Package confusingop defines an analyzer that reports assignments written as
// =+, =-, =! or =^, which are usually typos for +=, -=, != or ^=.
package confusingop

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports assignments whose "=" is immediately followed by a unary operator.
var Analyzer = &analysis.Analyzer{
	Name:     "confusingop",
	Doc:      "reports =+, =-, =! and =^ assignments that look like typos for compound operators",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	insp.Preorder([]ast.Node{(*ast.AssignStmt)(nil)}, func(n ast.Node) {
		stmt, ok := n.(*ast.AssignStmt)
		if !ok || stmt.Tok != token.ASSIGN || len(stmt.Rhs) == 0 {
			return
		}
		// Only the first right-hand side can touch the "=": in `a, b =-1, 2`
		// the operator belongs to Rhs[0].
		unary, ok := stmt.Rhs[0].(*ast.UnaryExpr)
		if !ok || !isConfusing(unary.Op) {
			return
		}
		// Nothing between "=" and the operator; "= +1" and "=\n+1" are deliberate.
		if unary.OpPos != stmt.TokPos+1 {
			return
		}
		pass.Reportf(
			stmt.TokPos,
			`"=%[1]s" looks like a typo for "%[1]s="; write "= %[1]s" if the unary operator is intended`,
			unary.Op,
		)
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func isConfusing(op token.Token) bool {
	switch op {
	case token.ADD, token.SUB, token.NOT, token.XOR:
		return true
	default:
		return false
	}
}
