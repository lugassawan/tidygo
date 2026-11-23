// Package nonestedswitch defines an analyzer that forbids a switch or type
// switch nested inside the case clause of another switch or type switch.
// See docs/analyzers/nonestedswitch.md.
package nonestedswitch

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports switch statements nested inside another switch statement.
var Analyzer = &analysis.Analyzer{
	Name:     "nonestedswitch",
	Doc:      "forbids a switch statement nested inside another switch; extract the inner switch into a function",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{
		(*ast.SwitchStmt)(nil),
		(*ast.TypeSwitchStmt)(nil),
	}

	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if push && nestedInSwitch(stack[:len(stack)-1]) {
			pass.Reportf(
				n.Pos(),
				"switch statement nested inside another switch; extract the inner switch into a function",
			)
		}
		return true
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// nestedInSwitch reports whether the nearest switch-like ancestor in the same
// function is a switch or type switch. ancestors runs outermost to innermost.
func nestedInSwitch(ancestors []ast.Node) bool {
	for _, ancestor := range slices.Backward(ancestors) {
		switch ancestor.(type) {
		case *ast.SwitchStmt, *ast.TypeSwitchStmt:
			return true
		case *ast.SelectStmt, *ast.FuncLit, *ast.FuncDecl:
			return false
		}
	}
	return false
}
