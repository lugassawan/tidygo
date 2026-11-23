// Package elseifelse defines an analyzer that reports if-else-if chains that do
// not end in a final else clause. It is opt-in in the tidygo plugin.
// See docs/analyzers/elseifelse.md.
package elseifelse

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports if-else-if chains without a final else.
var Analyzer = &analysis.Analyzer{
	Name:     "elseifelse",
	Doc:      "requires an if-else-if chain to end in a final else clause",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{(*ast.IfStmt)(nil)}

	// links holds the else-if statements already walked from an earlier chain
	// head, so each chain is examined once. Preorder visits a head before its links.
	links := make(map[*ast.IfStmt]bool)

	insp.Preorder(nodeFilter, func(n ast.Node) {
		head, ok := n.(*ast.IfStmt)
		if !ok || links[head] {
			return
		}
		last := head
		for {
			next, isIf := last.Else.(*ast.IfStmt)
			if !isIf {
				break
			}
			links[next] = true
			last = next
		}
		if last != head && last.Else == nil {
			pass.Reportf(last.Pos(), "if-else-if chain has no final else")
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}
