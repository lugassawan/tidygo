// Package busywait reports loops that spin on a plain variable or field, with an
// empty body or a lone time.Sleep or runtime.Gosched. An unsynchronised read is a
// data race that burns CPU; conditions with calls or receives may be polls and are skipped.
// See docs/analyzers/busywait.md.
package busywait

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/callee"
)

// Analyzer reports for loops that busy-wait on plain state.
var Analyzer = &analysis.Analyzer{
	Name:     "busywait",
	Doc:      "reports loops that busy-wait on a plain variable instead of using a channel, sync.Cond or context",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
	}
	insp.Preorder([]ast.Node{(*ast.ForStmt)(nil)}, func(n ast.Node) {
		loop, ok := n.(*ast.ForStmt)
		if !ok || !isBusyWait(pass.TypesInfo, loop) {
			return
		}
		pass.Reportf(
			loop.For,
			"loop busy-waits on %s; use a channel, sync.Cond or context instead",
			types.ExprString(loop.Cond),
		)
	})
	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func isBusyWait(info *types.Info, loop *ast.ForStmt) bool {
	if loop.Cond == nil || loop.Init != nil || loop.Post != nil {
		return false
	}
	// `for true` waits on no state; constcond covers it.
	if info.Types[loop.Cond].Value != nil {
		return false
	}
	return isPlainState(loop.Cond) && isSpinBody(info, loop.Body)
}

// isPlainState reports whether expr reads only variables, constants and
// literals, with no calls, receives or index expressions.
func isPlainState(expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident, *ast.BasicLit:
		return true
	case *ast.ParenExpr:
		return isPlainState(e.X)
	case *ast.SelectorExpr:
		return isPlainState(e.X)
	case *ast.UnaryExpr:
		return isPlainUnaryOp(e.Op) && isPlainState(e.X)
	case *ast.BinaryExpr:
		return isPlainBinaryOp(e.Op) && isPlainState(e.X) && isPlainState(e.Y)
	}
	return false
}

func isPlainUnaryOp(op token.Token) bool {
	return op == token.NOT || op == token.SUB || op == token.ADD
}

func isPlainBinaryOp(op token.Token) bool {
	switch op {
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ, token.LAND, token.LOR:
		return true
	}
	return false
}

func isSpinBody(info *types.Info, body *ast.BlockStmt) bool {
	switch len(body.List) {
	case 0:
		return true
	case 1:
		return isYield(info, body.List[0])
	}
	return false
}

// isYield reports whether stmt is a bare time.Sleep(x) with no call inside x,
// or a runtime.Gosched().
func isYield(info *types.Info, stmt ast.Stmt) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	switch {
	case callee.IsFunc(info, call, "runtime", "Gosched"):
		return true
	case callee.IsFunc(info, call, "time", "Sleep"):
		return len(call.Args) == 1 && !containsCall(info, call.Args[0])
	}
	return false
}

// containsCall reports whether expr calls a function. Type conversions
// (time.Duration(n)) and builtins (len(q)) run no user code, so they do not count.
func containsCall(info *types.Info, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			tv := info.Types[call.Fun]
			found = found || (!tv.IsType() && !tv.IsBuiltin())
		}
		return !found
	})
	return found
}
