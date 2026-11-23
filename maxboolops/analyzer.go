// Package maxboolops defines an analyzer that reports boolean expressions with
// too many && and || operators; call arguments and function literals start a new
// expression. Config.Max overrides the default limit of DefaultMax.
// See docs/analyzers/maxboolops.md.
package maxboolops

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// DefaultMax is the operator limit used when Config.Max is not positive.
const DefaultMax = 3

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// Max is the largest allowed number of && and || operators in one boolean
	// expression. Values <= 0 select DefaultMax.
	Max int
}

// Analyzer reports boolean expressions with more than DefaultMax && and || operators.
var Analyzer = New(Config{})

// New builds a maxboolops analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	limit := cfg.Max
	if limit <= 0 {
		limit = DefaultMax
	}
	return &analysis.Analyzer{
		Name:     "maxboolops",
		Doc:      "forbids boolean expressions with too many && and || operators",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, limit)
		},
	}
}

func run(pass *analysis.Pass, limit int) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{(*ast.BinaryExpr)(nil)}

	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		expr, ok := n.(*ast.BinaryExpr)
		if !ok || !isBoolOp(expr) || !isChainRoot(stack) {
			return true
		}
		if count := countOps(expr); count > limit {
			pass.Reportf(expr.Pos(), "boolean expression has %d && / || operators, max is %d", count, limit)
		}
		return true
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func isBoolOp(e *ast.BinaryExpr) bool {
	return e.Op == token.LAND || e.Op == token.LOR
}

// isChainRoot reports whether the node on top of stack is not nested inside
// another && / || expression, looking through parentheses and unary !.
func isChainRoot(stack []ast.Node) bool {
	for i := len(stack) - 2; i >= 0; i-- {
		switch p := stack[i].(type) {
		case *ast.ParenExpr:
			continue
		case *ast.UnaryExpr:
			if p.Op == token.NOT {
				continue
			}
			return true
		case *ast.BinaryExpr:
			return !isBoolOp(p)
		default:
			return true
		}
	}
	return true
}

// countOps counts the && and || operators reachable from e through
// parentheses, unary ! and other && / || nodes. Any other operand is a leaf.
func countOps(e ast.Expr) int {
	switch x := e.(type) {
	case *ast.ParenExpr:
		return countOps(x.X)
	case *ast.UnaryExpr:
		if x.Op == token.NOT {
			return countOps(x.X)
		}
	case *ast.BinaryExpr:
		if isBoolOp(x) {
			return 1 + countOps(x.X) + countOps(x.Y)
		}
	}
	return 0
}
