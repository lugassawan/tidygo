// Package maxcaselines defines an analyzer that reports case clauses whose
// body spans too many lines. The span runs from the first statement's start
// line through the last statement's end line, inclusive, so blank and comment
// lines between statements count while leading and trailing comments do not.
// Config.Max overrides the default limit of DefaultMax.
package maxcaselines

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// DefaultMax is the line limit used when Config.Max is not positive.
const DefaultMax = 10

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// Max is the largest allowed body span in lines. Values <= 0 select DefaultMax.
	Max int
}

// Analyzer reports case clause bodies longer than DefaultMax lines.
var Analyzer = New(Config{})

// New builds a maxcaselines analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	limit := cfg.Max
	if limit <= 0 {
		limit = DefaultMax
	}
	return &analysis.Analyzer{
		Name:     "maxcaselines",
		Doc:      "forbids switch and select case clauses whose body spans too many lines",
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

	nodeFilter := []ast.Node{
		(*ast.CaseClause)(nil),
		(*ast.CommClause)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		body := clauseBody(n)
		if len(body) == 0 {
			return
		}
		// Unadjusted positions: //line directives must not skew the span.
		first := pass.Fset.PositionFor(body[0].Pos(), false).Line
		last := pass.Fset.PositionFor(body[len(body)-1].End(), false).Line
		if lines := last - first + 1; lines > limit {
			pass.Reportf(n.Pos(), "case clause body spans %d lines, max is %d", lines, limit)
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// clauseBody returns the statements of a switch or select clause.
func clauseBody(n ast.Node) []ast.Stmt {
	switch c := n.(type) {
	case *ast.CaseClause:
		return c.Body
	case *ast.CommClause:
		return c.Body
	}
	return nil
}
