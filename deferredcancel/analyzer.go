// Package deferredcancel defines an analyzer that requires the cancel func of context.WithCancel, WithTimeout,
// WithDeadline and their Cause variants to be deferred on the next statement. Unlike govet's lostcancel (cancel
// must be called on some path), it demands `defer cancel()` right after creation; blank or escaping cancel funcs
// are skipped.
package deferredcancel

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/callee"
	"github.com/lugassawan/tidygo/internal/funcs"
)

// Analyzer reports context cancel functions that are not deferred immediately.
var Analyzer = &analysis.Analyzer{
	Name:     "deferredcancel",
	Doc:      "requires the cancel func of context.WithCancel, WithTimeout and WithDeadline to be deferred immediately",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

const contextPkg = "context"

var cancelFuncs = []string{
	"WithCancel", "WithTimeout", "WithDeadline",
	"WithCancelCause", "WithTimeoutCause", "WithDeadlineCause",
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	funcs.Each(insp, func(_ *ast.FuncType, body *ast.BlockStmt) {
		checkFunc(pass, body)
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// checkFunc visits every statement list of body except those of nested function
// literals, which are checked on their own against their own body, and those of
// loops: a defer in a loop body piles up until the function returns, so the fix
// there is a per-iteration function, not a defer.
func checkFunc(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit, *ast.ForStmt, *ast.RangeStmt:
			return false
		case *ast.BlockStmt:
			checkList(pass, body, n.List)
		case *ast.CaseClause:
			checkList(pass, body, n.Body)
		case *ast.CommClause:
			checkList(pass, body, n.Body)
		}
		return true
	})
}

func checkList(pass *analysis.Pass, body *ast.BlockStmt, list []ast.Stmt) {
	for i, stmt := range list {
		var next ast.Stmt
		if i+1 < len(list) {
			next = list[i+1]
		}
		checkStmt(pass, body, stmt, next)
	}
}

// checkStmt reports stmt when it creates a cancel function that neither next
// defers nor the rest of body lets escape. next is nil for the last statement.
func checkStmt(pass *analysis.Pass, body *ast.BlockStmt, stmt, next ast.Stmt) {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return
	}
	fn, cancel, ok := cancelAssign(pass.TypesInfo, assign)
	if !ok {
		return
	}
	obj := pass.TypesInfo.ObjectOf(cancel)
	if obj == nil || isManaged(pass.TypesInfo, body, obj, next) {
		return
	}
	if assign.Tok == token.ASSIGN && defersAfter(pass.TypesInfo, body, obj, assign.End()) {
		return // `var cancel ...; if c { ctx, cancel = ... } else { ... }; defer cancel()`
	}
	pass.Reportf(assign.Pos(),
		"cancel func from context.%s should be deferred immediately: add \"defer %s(%s)\"",
		fn, cancel.Name, deferArg(fn),
	)
}

// deferArg is the argument list of the suggested defer: WithCancelCause returns a
// CancelCauseFunc, which takes the cause.
func deferArg(fn string) string {
	if fn == "WithCancelCause" {
		return "nil"
	}
	return ""
}

// defersAfter reports whether body defers obj's cancel func at a point after pos.
func defersAfter(info *types.Info, body *ast.BlockStmt, obj types.Object, pos token.Pos) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			return false
		case *ast.DeferStmt:
			found = found || (n.Pos() > pos && defersCancel(info, n, obj))
		}
		return !found
	})
	return found
}

// cancelAssign matches `ctx, cancel := context.WithX(...)` (or `=`) and returns
// the context function's name and the cancel identifier. A blank cancel is not
// matched: govet's lostcancel already reports it.
func cancelAssign(info *types.Info, assign *ast.AssignStmt) (string, *ast.Ident, bool) {
	if len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
		return "", nil, false
	}
	call, ok := assign.Rhs[0].(*ast.CallExpr)
	if !ok {
		return "", nil, false
	}
	cancel, ok := assign.Lhs[1].(*ast.Ident)
	if !ok || cancel.Name == "_" {
		return "", nil, false
	}
	fn, ok := cancelFuncName(info, call)
	return fn, cancel, ok
}

func cancelFuncName(info *types.Info, call *ast.CallExpr) (string, bool) {
	for _, name := range cancelFuncs {
		if callee.IsFunc(info, call, contextPkg, name) {
			return name, true
		}
	}
	return "", false
}

// isManaged reports whether the cancel object needs no report: it is not a local
// of body (so it is managed elsewhere), next defers it, or it escapes in body.
func isManaged(info *types.Info, body *ast.BlockStmt, obj types.Object, next ast.Stmt) bool {
	if obj.Pos() < body.Pos() || obj.Pos() >= body.End() {
		return true
	}
	return defersCancel(info, next, obj) || escapes(info, body, obj)
}

// defersCancel reports whether stmt is `defer cancel(...)` or a deferred function
// literal that calls cancel as one of its top-level statements.
func defersCancel(info *types.Info, stmt ast.Stmt, obj types.Object) bool {
	deferred, ok := stmt.(*ast.DeferStmt)
	if !ok {
		return false
	}
	lit, isLit := deferred.Call.Fun.(*ast.FuncLit)
	if !isLit {
		return callsObject(info, deferred.Call, obj)
	}
	for _, s := range lit.Body.List {
		if es, ok := s.(*ast.ExprStmt); ok {
			call, _ := es.X.(*ast.CallExpr)
			if callsObject(info, call, obj) {
				return true
			}
		}
	}
	return false
}

func callsObject(info *types.Info, call *ast.CallExpr, obj types.Object) bool {
	if call == nil {
		return false
	}
	id, ok := call.Fun.(*ast.Ident)
	return ok && info.ObjectOf(id) == obj
}

// escapes reports whether body uses obj in any way beyond calling it or
// reassigning it: returning, passing, storing or aliasing it, or reaching it from
// a go statement or a function literal that is not itself deferred.
func escapes(info *types.Info, body *ast.BlockStmt, obj types.Object) bool {
	scan := &escapeScan{
		info:     info,
		obj:      obj,
		allowed:  make(map[*ast.Ident]bool),
		deferred: make(map[*ast.FuncLit]bool),
	}
	ast.Inspect(body, scan.visit)
	return scan.escaped
}

// escapeScan walks a function body once in preorder. Parents are visited before
// their children, so an identifier is already marked allowed when it is reached.
type escapeScan struct {
	info     *types.Info
	obj      types.Object
	allowed  map[*ast.Ident]bool   // call and declaration or assignment targets
	deferred map[*ast.FuncLit]bool // literals called directly by a defer statement
	escaped  bool
}

func (s *escapeScan) visit(n ast.Node) bool {
	switch n := n.(type) {
	case *ast.AssignStmt:
		s.allow(n.Lhs...)
	case *ast.ValueSpec:
		for _, name := range n.Names {
			s.allowed[name] = true
		}
	case *ast.CallExpr:
		s.allow(n.Fun)
	case *ast.DeferStmt:
		if lit, ok := n.Call.Fun.(*ast.FuncLit); ok {
			s.deferred[lit] = true
		}
	case *ast.GoStmt:
		s.escaped = s.escaped || s.mentions(n)
	case *ast.FuncLit:
		s.escaped = s.escaped || (!s.deferred[n] && s.mentions(n))
	case *ast.Ident:
		s.escaped = s.escaped || (!s.allowed[n] && s.isObj(n))
	}
	return !s.escaped
}

func (s *escapeScan) allow(exprs ...ast.Expr) {
	for _, expr := range exprs {
		if id, ok := expr.(*ast.Ident); ok {
			s.allowed[id] = true
		}
	}
}

func (s *escapeScan) isObj(id *ast.Ident) bool {
	return s.info.ObjectOf(id) == s.obj
}

func (s *escapeScan) mentions(n ast.Node) bool {
	found := false
	ast.Inspect(n, func(node ast.Node) bool {
		id, ok := node.(*ast.Ident)
		found = found || (ok && s.isObj(id))
		return !found
	})
	return found
}
