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
		unary, ok := leadingUnary(stmt.Rhs[0])
		if !ok || !isConfusing(unary.Op) {
			return
		}
		// Nothing between "=" and the operator; "= +1" and "=\n+1" are deliberate.
		if unary.OpPos != stmt.TokPos+1 {
			return
		}
		// Compact code such as `x=-1` is consistently unspaced, so it is
		// deliberate; the typo shape has a space before "=" but not after.
		if stmt.Lhs[len(stmt.Lhs)-1].End() >= stmt.TokPos {
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

// leadingUnary returns the unary expression that starts e. Unary operators bind
// tighter than any binary one, so in `=+ y * 2` the right-hand side is the
// binary expression (+y) * 2 and the unary is its leftmost operand.
func leadingUnary(e ast.Expr) (*ast.UnaryExpr, bool) {
	for {
		switch x := e.(type) {
		case *ast.BinaryExpr:
			e = x.X
		case *ast.UnaryExpr:
			return x, true
		default:
			return nil, false
		}
	}
}

func isConfusing(op token.Token) bool {
	switch op {
	case token.ADD, token.SUB, token.NOT, token.XOR:
		return true
	default:
		return false
	}
}
