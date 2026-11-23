// Package txrollback defines an analyzer that requires a begun *sql.Tx to be
// rolled back through a deferred Rollback, since a leaked transaction pins its
// connection and locks.
// See docs/analyzers/txrollback.md.
package txrollback

import (
	"go/ast"
	"go/token"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/callee"
	"github.com/lugassawan/tidygo/internal/funcs"
)

// Analyzer reports transactions that are begun without a deferred Rollback.
var Analyzer = &analysis.Analyzer{
	Name:     "txrollback",
	Doc:      "requires a deferred Rollback for every *sql.Tx that is begun and does not escape the function",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

type txType struct{ pkg, name string }

// trackedTypes lists the transaction types whose Begin must be paired with a
// deferred Rollback.
var trackedTypes = []txType{
	{"database/sql", "Tx"},
}

// beginPrefixes are the lower-cased name prefixes of a function that starts a
// transaction: Begin, BeginTx, beginTx, StartTx, startTransaction.
var beginPrefixes = []string{"begin", "start"}

// Diagnostics. A Begin inside a loop gets its own advice: a deferred Rollback in a
// loop body would pile up until the function returns, so the fix is a function
// that handles one iteration.
const (
	messagePlain = "transaction %[1]s has no deferred Rollback: add \"defer %[1]s.Rollback()\" after Begin"
	messageLoop  = "transaction %[1]s begun in a loop has no Rollback: " +
		"extract the iteration into a function with \"defer %[1]s.Rollback()\""
)

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	funcs.Each(insp, func(ftype *ast.FuncType, body *ast.BlockStmt) {
		checkFunc(pass, ftype, body)
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// begin is a variable that receives a tracked transaction from a call.
type begin struct {
	id   *ast.Ident
	loop ast.Node // innermost enclosing for or range statement, nil outside loops
}

func checkFunc(pass *analysis.Pass, ftype *ast.FuncType, body *ast.BlockStmt) {
	info := pass.TypesInfo
	begins := collectBegins(info, body)
	if len(begins) == 0 {
		return
	}
	managed := deferredRollbacks(info, body)
	handedOffRollbacks(info, body, managed)
	escaped := escapedObjects(info, body)
	markResults(info, ftype, escaped)

	for _, b := range begins {
		obj := info.ObjectOf(b.id)
		if managed[obj] || escaped[obj] || declaredOutside(obj, body) || rollsBackIn(info, b.loop, obj) {
			continue // a transaction declared by an enclosing function is managed there
		}
		format := messagePlain
		if b.loop != nil {
			format = messageLoop
		}
		pass.Reportf(b.id.Pos(), format, b.id.Name)
	}
}

// declaredOutside reports whether obj is declared outside body, as when a closure
// assigns a transaction variable of its enclosing function.
func declaredOutside(obj types.Object, body *ast.BlockStmt) bool {
	return obj == nil || obj.Pos() < body.Pos() || obj.Pos() >= body.End()
}

// rollsBackIn reports whether loop calls Rollback on obj anywhere in its body,
// deferred or not. A loop iteration that rolls back by hand is not a leak in
// every shape, so it is accepted; one that never does is reported.
func rollsBackIn(info *types.Info, loop ast.Node, obj types.Object) bool {
	if loop == nil {
		return false
	}
	found := make(map[types.Object]bool)
	markRollbacksIn(info, loop, found)
	return found[obj]
}

// collectBegins returns the variables in body that receive a tracked transaction
// from a call, each with its innermost enclosing loop. Nested function literals
// are their own scope.
func collectBegins(info *types.Info, body *ast.BlockStmt) []begin {
	var begins []begin
	var walk func(root, loop ast.Node)
	walk = func(root, loop ast.Node) {
		ast.Inspect(root, func(n ast.Node) bool {
			if n == root {
				return true
			}
			switch node := n.(type) {
			case *ast.FuncLit:
				return false
			case *ast.ForStmt, *ast.RangeStmt:
				walk(node, node)
				return false
			case *ast.AssignStmt:
				begins = appendBegins(begins, loop, beginTargets(info, identsOf(node.Lhs), node.Rhs))
			case *ast.ValueSpec:
				begins = appendBegins(begins, loop, beginTargets(info, node.Names, node.Values))
			}
			return true
		})
	}
	walk(body, nil)
	return begins
}

func appendBegins(begins []begin, loop ast.Node, ids []*ast.Ident) []begin {
	for _, id := range ids {
		begins = append(begins, begin{id: id, loop: loop})
	}
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
		if isBeginCall(info, valueAt(rhs, i)) {
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

// isBeginCall reports whether e calls a function that starts a transaction. A
// getter such as txFromContext, a type conversion and a call through a function
// value do not, since none of them hands the caller a transaction to roll back.
func isBeginCall(info *types.Info, e ast.Expr) bool {
	if e == nil {
		return false
	}
	call, ok := ast.Unparen(e).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn := callee.Func(info, call)
	if fn == nil {
		return false
	}
	name := strings.ToLower(fn.Name())
	return slices.ContainsFunc(beginPrefixes, func(prefix string) bool {
		return strings.HasPrefix(name, prefix)
	})
}

func isTrackedVar(info *types.Info, id *ast.Ident) bool {
	v, ok := info.ObjectOf(id).(*types.Var)
	return ok && isTrackedType(v.Type())
}

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
	markRollbacksIn(info, root, into)
}

// handedOffRollbacks marks variables whose Rollback is handed to code that runs
// later: called inside any function literal (t.Cleanup(func() { tx.Rollback() })),
// or passed on as a method value (return tx.Commit, tx.Rollback, nil).
func handedOffRollbacks(info *types.Info, body *ast.BlockStmt, into map[types.Object]bool) {
	called := make(map[ast.Expr]bool)
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.CallExpr:
			called[ast.Unparen(node.Fun)] = true
		case *ast.FuncLit:
			markRollbacksIn(info, node.Body, into)
		case *ast.SelectorExpr:
			if obj := rollbackValueReceiver(info, node); obj != nil && !called[node] {
				into[obj] = true
			}
		}
		return true
	})
}

func rollbackValueReceiver(info *types.Info, sel *ast.SelectorExpr) types.Object {
	fn, ok := info.ObjectOf(sel.Sel).(*types.Func)
	if !ok || fn.Name() != "Rollback" {
		return nil
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil || !isTrackedType(sig.Recv().Type()) {
		return nil
	}
	recv, ok := ast.Unparen(sel.X).(*ast.Ident)
	if !ok {
		return nil
	}
	return info.ObjectOf(recv)
}

func markRollbacksIn(info *types.Info, root ast.Node, into map[types.Object]bool) {
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
