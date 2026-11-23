// Package txrollback defines an analyzer that requires a begun *sql.Tx to be
// rolled back through a deferred Rollback, since a leaked transaction pins its
// connection and locks. Transactions that escape the function (returned, passed
// on, stored) are skipped because ownership has moved.
package txrollback

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/callee"
)

// Analyzer reports transactions that are begun without a deferred Rollback.
var Analyzer = &analysis.Analyzer{
	Name:     "txrollback",
	Doc:      "requires a deferred Rollback for every *sql.Tx that is begun and does not escape the function",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// txType names a transaction type by package path and type name.
type txType struct{ pkg, name string }

// trackedTypes lists the transaction types whose Begin must be paired with a
// deferred Rollback.
var trackedTypes = []txType{
	{"database/sql", "Tx"},
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		if ftype, body := funcParts(n); body != nil {
			checkFunc(pass, ftype, body)
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func funcParts(n ast.Node) (*ast.FuncType, *ast.BlockStmt) {
	switch fn := n.(type) {
	case *ast.FuncDecl:
		return fn.Type, fn.Body
	case *ast.FuncLit:
		return fn.Type, fn.Body
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, ftype *ast.FuncType, body *ast.BlockStmt) {
	info := pass.TypesInfo
	begins := collectBegins(info, body)
	if len(begins) == 0 {
		return
	}
	rolledBack := deferredRollbacks(info, body)
	escaped := escapedObjects(info, body)
	markResults(info, ftype, escaped)

	for _, id := range begins {
		obj := info.ObjectOf(id)
		if rolledBack[obj] || escaped[obj] {
			continue
		}
		pass.Reportf(
			id.Pos(),
			"transaction %[1]s is not rolled back on every path: add \"defer %[1]s.Rollback()\" after Begin",
			id.Name,
		)
	}
}

// collectBegins returns the identifiers in body that receive a tracked
// transaction from a call. Nested function literals are their own scope.
func collectBegins(info *types.Info, body *ast.BlockStmt) []*ast.Ident {
	var begins []*ast.Ident
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.AssignStmt:
			begins = append(begins, beginTargets(info, identsOf(node.Lhs), node.Rhs)...)
		case *ast.ValueSpec:
			begins = append(begins, beginTargets(info, node.Names, node.Values)...)
		}
		return true
	})
	return begins
}

// beginTargets filters lhs down to the tracked transaction variables that are
// assigned from a call. A single call on the right feeds every target, as in
// "tx, err := db.Begin()"; otherwise targets pair with values by position.
func beginTargets(info *types.Info, lhs []*ast.Ident, rhs []ast.Expr) []*ast.Ident {
	var targets []*ast.Ident
	for i, id := range lhs {
		if id == nil || id.Name == "_" || !isTrackedVar(info, id) {
			continue
		}
		if isCall(valueAt(rhs, i)) {
			targets = append(targets, id)
		}
	}
	return targets
}

func identsOf(exprs []ast.Expr) []*ast.Ident {
	ids := make([]*ast.Ident, len(exprs))
	for i, e := range exprs {
		ids[i], _ = ast.Unparen(e).(*ast.Ident)
	}
	return ids
}

func valueAt(rhs []ast.Expr, i int) ast.Expr {
	switch {
	case len(rhs) == 1:
		return rhs[0]
	case i < len(rhs):
		return rhs[i]
	}
	return nil
}

func isCall(e ast.Expr) bool {
	if e == nil {
		return false
	}
	_, ok := ast.Unparen(e).(*ast.CallExpr)
	return ok
}

func isTrackedVar(info *types.Info, id *ast.Ident) bool {
	v, ok := info.ObjectOf(id).(*types.Var)
	return ok && isTrackedType(v.Type())
}

// isTrackedType reports whether t is a pointer to one of trackedTypes.
func isTrackedType(t types.Type) bool {
	ptr, ok := types.Unalias(t).(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := types.Unalias(ptr.Elem()).(*types.Named)
	if !ok {
		return false
	}
	return slices.ContainsFunc(trackedTypes, func(tracked txType) bool {
		return tracked.names(named.Obj())
	})
}

func (t txType) names(obj *types.TypeName) bool {
	return obj.Pkg() != nil && obj.Pkg().Path() == t.pkg && obj.Name() == t.name
}

// deferredRollbacks returns the variables whose Rollback is deferred in body,
// either directly or anywhere inside a deferred function literal. Function
// literals that are not deferred are their own scope and are not searched.
func deferredRollbacks(info *types.Info, body *ast.BlockStmt) map[types.Object]bool {
	rolledBack := make(map[types.Object]bool)
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.DeferStmt:
			markRollbacks(info, node.Call, rolledBack)
			return false
		}
		return true
	})
	return rolledBack
}

func markRollbacks(info *types.Info, deferred *ast.CallExpr, into map[types.Object]bool) {
	root := ast.Node(deferred)
	if lit, ok := ast.Unparen(deferred.Fun).(*ast.FuncLit); ok {
		root = lit.Body
	}
	ast.Inspect(root, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if obj := rollbackReceiver(info, call); obj != nil {
				into[obj] = true
			}
		}
		return true
	})
}

// rollbackReceiver returns the variable on which call invokes Rollback, or nil
// when call is not a tracked Rollback or its receiver is not a plain variable.
func rollbackReceiver(info *types.Info, call *ast.CallExpr) types.Object {
	if !isRollback(info, call) {
		return nil
	}
	sel, ok := ast.Unparen(call.Fun).(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	recv, ok := ast.Unparen(sel.X).(*ast.Ident)
	if !ok {
		return nil
	}
	return info.ObjectOf(recv)
}

func isRollback(info *types.Info, call *ast.CallExpr) bool {
	return slices.ContainsFunc(trackedTypes, func(tracked txType) bool {
		return callee.IsMethod(info, call, tracked.pkg, tracked.name, "Rollback")
	})
}

// escapedObjects returns the variables in body whose value leaves the
// function's control: returned, passed to a call, stored, sent, or used in a go
// statement. Nested function literals are searched too, so that a capture is
// never mistaken for a leak-free use.
func escapedObjects(info *types.Info, body *ast.BlockStmt) map[types.Object]bool {
	escaped := make(map[types.Object]bool)
	ast.Inspect(body, func(n ast.Node) bool {
		if goStmt, ok := n.(*ast.GoStmt); ok {
			markAllUses(info, goStmt, escaped)
		}
		for _, e := range escapingExprs(n) {
			markIdent(info, e, escaped)
		}
		return true
	})
	return escaped
}

// escapingExprs returns the operands of n that, when they are plain variables,
// hand their value to someone else.
func escapingExprs(n ast.Node) []ast.Expr {
	switch node := n.(type) {
	case *ast.ReturnStmt:
		return node.Results
	case *ast.CallExpr:
		return node.Args
	case *ast.AssignStmt:
		return node.Rhs
	case *ast.ValueSpec:
		return node.Values
	case *ast.SendStmt:
		return []ast.Expr{node.Value}
	case *ast.UnaryExpr:
		if node.Op == token.AND {
			return []ast.Expr{node.X}
		}
	case *ast.CompositeLit:
		return compositeValues(node)
	}
	return nil
}

func compositeValues(lit *ast.CompositeLit) []ast.Expr {
	values := make([]ast.Expr, 0, len(lit.Elts))
	for _, elt := range lit.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			elt = kv.Value
		}
		values = append(values, elt)
	}
	return values
}

// markResults marks the named result variables of a function as escaped: they
// are returned to the caller, even by a bare "return".
func markResults(info *types.Info, ftype *ast.FuncType, into map[types.Object]bool) {
	if ftype == nil || ftype.Results == nil {
		return
	}
	for _, field := range ftype.Results.List {
		for _, name := range field.Names {
			markIdent(info, name, into)
		}
	}
}

func markAllUses(info *types.Info, root ast.Node, into map[types.Object]bool) {
	ast.Inspect(root, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok {
			markIdent(info, id, into)
		}
		return true
	})
}

func markIdent(info *types.Info, e ast.Expr, into map[types.Object]bool) {
	id, ok := ast.Unparen(e).(*ast.Ident)
	if !ok {
		return
	}
	if obj := info.ObjectOf(id); obj != nil {
		into[obj] = true
	}
}
