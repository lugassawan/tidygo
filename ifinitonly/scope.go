package ifinitonly

import (
	"go/ast"
	"go/token"
	"slices"
)

// resolver counts the references to one outer variable. The analyzer runs
// without type information, so shadowing is resolved syntactically: whenever a
// scope declares the same name, the rest of that scope no longer refers to the
// outer variable and the walk stops there.
type resolver struct {
	name  string
	count int
}

// countRefs returns how many identifiers in n refer to the variable called name
// that was declared outside n. It is deliberately conservative: when a name
// might be a reference (a composite literal key, for example) it is counted.
func countRefs(n ast.Node, name string) int {
	r := &resolver{name: name}
	r.visit(n)
	return r.count
}

func (r *resolver) visit(n ast.Node) {
	switch n := n.(type) {
	case nil:
	case *ast.Ident:
		if n.Name == r.name {
			r.count++
		}
	case *ast.SelectorExpr:
		r.visit(n.X) // Sel is a field or method name, not a variable
	case *ast.Field:
		r.visit(n.Type) // Names are declarations
	case *ast.LabeledStmt:
		r.visit(n.Stmt) // labels live in their own namespace
	case *ast.BranchStmt: // so do break/continue/goto targets
	case *ast.BlockStmt:
		r.stmts(n.List)
	case *ast.CaseClause:
		r.exprs(n.List)
		r.stmts(n.Body)
	case *ast.CommClause:
		r.commClause(n)
	case *ast.IfStmt:
		r.ifStmt(n)
	case *ast.ForStmt:
		r.forStmt(n)
	case *ast.SwitchStmt:
		r.switchStmt(n)
	case *ast.TypeSwitchStmt:
		r.typeSwitchStmt(n)
	case *ast.RangeStmt:
		r.rangeStmt(n)
	case *ast.FuncLit:
		r.funcLit(n)
	default:
		r.children(n)
	}
}

// visitAll visits each node in order. Nil interface values are skipped.
func (r *resolver) visitAll(nodes ...ast.Node) {
	for _, n := range nodes {
		r.visit(n)
	}
}

func (r *resolver) exprs(list []ast.Expr) {
	for _, e := range list {
		r.visit(e)
	}
}

// children visits the direct children of n through the dispatcher. ast.Inspect
// calls back for n itself first, so that call descends and every other node is
// handed to visit, which does its own recursion (hence return false). The
// trailing nil call that Inspect makes is a no-op in visit.
func (r *resolver) children(n ast.Node) {
	ast.Inspect(n, func(c ast.Node) bool {
		if c == n {
			return true
		}
		r.visit(c)
		return false
	})
}

// stmts walks a statement list. A statement that declares the name shadows it
// for the remainder of the list.
func (r *resolver) stmts(list []ast.Stmt) {
	for _, s := range list {
		if r.declares(s) {
			return
		}
		r.visit(s)
	}
}

// declares reports whether s introduces a new variable called r.name. When it
// does, it also counts the references in s's own initializers: Go scoping
// starts after the declaration, so in `n := n + 1` the right-hand n is the
// outer one.
func (r *resolver) declares(s ast.Stmt) bool {
	switch s := s.(type) {
	case *ast.AssignStmt:
		if s.Tok != token.DEFINE || !hasExpr(s.Lhs, r.name) {
			return false
		}
		r.exprs(s.Rhs)
		return true
	case *ast.DeclStmt:
		return r.declStmt(s)
	}
	return false
}

// declStmt handles var, const and type declarations. Specs before the declaring
// one in the same group are still in the outer scope.
func (r *resolver) declStmt(ds *ast.DeclStmt) bool {
	gd, ok := ds.Decl.(*ast.GenDecl)
	if !ok {
		return false
	}
	idx := slices.IndexFunc(gd.Specs, func(sp ast.Spec) bool { return specDeclares(sp, r.name) })
	if idx < 0 {
		return false
	}
	for _, sp := range gd.Specs[:idx] {
		r.visit(sp)
	}
	if vs, ok := gd.Specs[idx].(*ast.ValueSpec); ok {
		r.exprs(vs.Values)
	}
	return true
}

func (r *resolver) ifStmt(s *ast.IfStmt) {
	if r.declares(s.Init) {
		return
	}
	r.visitAll(s.Init, s.Cond, s.Body, s.Else)
}

func (r *resolver) forStmt(s *ast.ForStmt) {
	if r.declares(s.Init) {
		return
	}
	r.visitAll(s.Init, s.Cond, s.Post, s.Body)
}

func (r *resolver) switchStmt(s *ast.SwitchStmt) {
	if r.declares(s.Init) {
		return
	}
	r.visitAll(s.Init, s.Tag, s.Body)
}

// typeSwitchStmt handles `switch n := x.(type)`: the clauses see the new n.
func (r *resolver) typeSwitchStmt(s *ast.TypeSwitchStmt) {
	if r.declares(s.Init) {
		return
	}
	r.visit(s.Init) // Init is evaluated before the clause variable exists
	if r.declares(s.Assign) {
		return
	}
	r.visitAll(s.Assign, s.Body)
}

func (r *resolver) rangeStmt(s *ast.RangeStmt) {
	r.visit(s.X)
	if s.Tok == token.DEFINE && hasExpr([]ast.Expr{s.Key, s.Value}, r.name) {
		return
	}
	r.visitAll(s.Key, s.Value, s.Body)
}

// commClause handles `case n := <-ch:`: the body sees the new n.
func (r *resolver) commClause(c *ast.CommClause) {
	if r.declares(c.Comm) {
		return
	}
	r.visit(c.Comm)
	r.stmts(c.Body)
}

func (r *resolver) funcLit(fl *ast.FuncLit) {
	if hasField(fl.Type.Params, r.name) || hasField(fl.Type.Results, r.name) {
		return
	}
	// Type and Body are never nil for a FuncLit, so no typed-nil reaches visit.
	r.visitAll(fl.Type, fl.Body)
}

func specDeclares(sp ast.Spec, name string) bool {
	switch sp := sp.(type) {
	case *ast.ValueSpec:
		return slices.ContainsFunc(sp.Names, func(id *ast.Ident) bool { return id.Name == name })
	case *ast.TypeSpec:
		return sp.Name.Name == name
	}
	return false
}

// hasExpr reports whether any expression is the plain identifier name.
func hasExpr(list []ast.Expr, name string) bool {
	return slices.ContainsFunc(list, func(e ast.Expr) bool { return isIdentNamed(e, name) })
}

// hasField reports whether a parameter or result list declares name. fl may be nil.
func hasField(fl *ast.FieldList, name string) bool {
	if fl == nil {
		return false
	}
	return slices.ContainsFunc(fl.List, func(f *ast.Field) bool {
		return slices.ContainsFunc(f.Names, func(id *ast.Ident) bool { return id.Name == name })
	})
}
