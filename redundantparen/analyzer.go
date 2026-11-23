// Package redundantparen defines an analyzer that reports unnecessary
// parentheses around a whole return value, assignment or declaration
// right-hand side, or call argument. Parentheses inside binary operands are
// left alone because they often clarify mixed operators.
package redundantparen

import (
	"go/ast"
	"go/token"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports redundant parentheses.
var Analyzer = &analysis.Analyzer{
	Name:     "redundantparen",
	Doc:      "forbids unnecessary parentheses around return values, assignment right-hand sides and call arguments",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{(*ast.ParenExpr)(nil)}

	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		paren, ok := n.(*ast.ParenExpr)
		if !push || !ok || len(stack) < 2 {
			return true
		}
		// Only the outer pair of an unformatted ((x)) sits in operand position,
		// so it alone is reported. gofmt collapses doubled parentheses anyway.
		if isRedundant(paren, stack[len(stack)-2], stack) {
			pass.Reportf(paren.Pos(), "redundant parentheses")
		}
		return true
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// isRedundant reports whether paren is a top-level operand whose parentheses
// can be removed without changing how the code parses.
func isRedundant(paren *ast.ParenExpr, parent ast.Node, stack []ast.Node) bool {
	if !isOperand(parent, paren) {
		return false
	}
	// `if x := (T{}); x.ok {` does not parse without the parentheses.
	required := containsCompositeLit(paren.X) && inHeader(stack, paren.Pos())
	return !required
}

// isOperand reports whether e is a top-level operand of parent: a returned
// value, a right-hand side, or a call argument. The callee of a call is never
// an operand, so conversions such as (*T)(nil) are kept.
func isOperand(parent ast.Node, e ast.Expr) bool {
	switch p := parent.(type) {
	case *ast.ReturnStmt:
		return slices.Contains(p.Results, e)
	case *ast.AssignStmt:
		return slices.Contains(p.Rhs, e)
	case *ast.ValueSpec:
		return slices.Contains(p.Values, e)
	case *ast.CallExpr:
		return slices.Contains(p.Args, e)
	}
	return false
}

// containsCompositeLit reports whether e holds a composite literal outside any
// function literal; braces inside a function body are unambiguous.
func containsCompositeLit(e ast.Expr) bool {
	found := false
	ast.Inspect(e, func(n ast.Node) bool {
		switch n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.CompositeLit:
			found = true
		}
		return !found
	})
	return found
}

// inHeader reports whether pos lies in the header of the nearest enclosing
// if, for, range or switch statement, that is before the opening brace of its
// body. A function literal ends the search: its body is a fresh context.
func inHeader(stack []ast.Node, pos token.Pos) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		var body *ast.BlockStmt
		switch s := stack[i].(type) {
		case *ast.FuncLit:
			return false
		case *ast.IfStmt:
			body = s.Body
		case *ast.ForStmt:
			body = s.Body
		case *ast.RangeStmt:
			body = s.Body
		case *ast.SwitchStmt:
			body = s.Body
		case *ast.TypeSwitchStmt:
			body = s.Body
		default:
			continue
		}
		return pos < body.Lbrace
	}
	return false
}
