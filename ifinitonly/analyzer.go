// Package ifinitonly defines an analyzer that forbids `if` init variables used
// only in the condition. The call can go straight into the condition instead.
// See docs/analyzers/ifinitonly.md.
package ifinitonly

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports `if` init variables that are referenced only once, in the condition.
var Analyzer = &analysis.Analyzer{
	Name:     "ifinitonly",
	Doc:      "forbids if init variables used only in the condition",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	insp.Preorder([]ast.Node{(*ast.IfStmt)(nil)}, func(n ast.Node) {
		ifs, ok := n.(*ast.IfStmt)
		if !ok {
			return
		}
		ident, ok := initVar(ifs)
		if !ok || !onlyInCond(ifs, ident.Name) {
			return
		}
		pass.Reportf(
			ident.Pos(),
			"variable %q in if init is only used in the condition; inline it into the condition",
			ident.Name,
		)
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func initVar(ifs *ast.IfStmt) (*ast.Ident, bool) {
	as, ok := ifs.Init.(*ast.AssignStmt)
	if !ok || as.Tok != token.DEFINE || len(as.Lhs) != 1 {
		return nil, false
	}
	ident, ok := as.Lhs[0].(*ast.Ident)
	if !ok || ident.Name == "_" {
		return nil, false
	}
	return ident, true
}

// onlyInCond reports whether name is referenced exactly once, in the condition,
// and nowhere in the body or else branch. That one reference must also be safe
// to inline: not under `&`, and not conditional or repeated (right of &&/||,
// inside a function literal).
func onlyInCond(ifs *ast.IfStmt, name string) bool {
	cond := scanRefs(ifs.Cond, name)
	inlinable := cond.count == 1 && cond.lazy == 0 && !cond.addr
	unusedElsewhere := scanRefs(ifs.Body, name).count == 0 && scanRefs(ifs.Else, name).count == 0
	return inlinable && unusedElsewhere
}
