// Package maxcases defines an analyzer that reports switch, type switch and
// select statements with too many case clauses. The default clause counts as
// a clause. Config.Max overrides the default limit of DefaultMax.
package maxcases

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// DefaultMax is the clause limit used when Config.Max is not positive.
const DefaultMax = 30

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// Max is the largest allowed number of clauses. Values <= 0 select DefaultMax.
	Max int
}

// Analyzer reports switch-like statements with more than DefaultMax clauses.
var Analyzer = New(Config{})

// New builds a maxcases analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	limit := cfg.Max
	if limit <= 0 {
		limit = DefaultMax
	}
	return &analysis.Analyzer{
		Name:     "maxcases",
		Doc:      "forbids switch, type switch and select statements with too many case clauses",
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
		(*ast.SwitchStmt)(nil),
		(*ast.TypeSwitchStmt)(nil),
		(*ast.SelectStmt)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		kind, count := describe(n)
		if count > limit {
			pass.Reportf(n.Pos(), "%s has %d case clauses, max is %d", kind, count, limit)
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// describe returns the statement kind and its clause count.
func describe(n ast.Node) (string, int) {
	switch s := n.(type) {
	case *ast.SwitchStmt:
		return "switch", len(s.Body.List)
	case *ast.TypeSwitchStmt:
		return "type switch", len(s.Body.List)
	case *ast.SelectStmt:
		return "select", len(s.Body.List)
	}
	return "", 0
}
