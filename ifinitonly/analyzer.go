// Package ifinitonly defines an analyzer that forbids `if` init variables used
// only in the condition. The call can go straight into the condition instead.
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

// initVar returns the variable declared by a single-variable `:=` init statement.
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
// and nowhere in the body or else branch. Several references in the condition
// are kept, because inlining would evaluate the initializer more than once.
// A reference under `&` is kept too, because `&f()` is not valid Go, and so is
// one whose evaluation inlining would change (see changesEvaluation).
func onlyInCond(ifs *ast.IfStmt, name string) bool {
	return countRefs(ifs.Cond, name) == 1 &&
		countRefs(ifs.Body, name) == 0 &&
		countRefs(ifs.Else, name) == 0 &&
		!addressTaken(ifs.Cond, name) &&
		!changesEvaluation(ifs.Cond, name)
}

// changesEvaluation reports whether a reference to name in cond sits where
// inlining the initializer would run it a different number of times: inside a
// function literal (once per call) or in the right operand of && or ||
// (skipped when the left operand short-circuits).
func changesEvaluation(cond ast.Expr, name string) bool {
	var path []ast.Node
	found := false
	ast.Inspect(cond, func(n ast.Node) bool {
		if n == nil {
			path = path[:len(path)-1]
			return false
		}
		path = append(path, n)
		if isIdentNamed(exprOf(n), name) {
			found = found || runsConditionally(path)
		}
		return !found
	})
	return found
}

// runsConditionally reports whether the last node of path is evaluated
// lazily or repeatedly relative to the start of the path.
func runsConditionally(path []ast.Node) bool {
	for i := len(path) - 1; i > 0; i-- {
		switch p := path[i-1].(type) {
		case *ast.FuncLit:
			return true
		case *ast.BinaryExpr:
			if (p.Op == token.LAND || p.Op == token.LOR) && ast.Node(p.Y) == path[i] {
				return true
			}
		}
	}
	return false
}

// exprOf returns n as an expression, or nil when it is not one.
func exprOf(n ast.Node) ast.Expr {
	e, _ := n.(ast.Expr)
	return e
}

// addressTaken reports whether the condition contains `&name`.
func addressTaken(cond ast.Expr, name string) bool {
	found := false
	ast.Inspect(cond, func(n ast.Node) bool {
		u, ok := n.(*ast.UnaryExpr)
		if ok && u.Op == token.AND {
			found = found || isIdentNamed(ast.Unparen(u.X), name)
		}
		return !found
	})
	return found
}

func isIdentNamed(e ast.Expr, name string) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == name
}
