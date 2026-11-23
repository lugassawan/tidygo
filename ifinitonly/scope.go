package ifinitonly

import (
	"go/ast"
	"go/token"
	"slices"
)

type refs struct {
	count int  // every reference
	lazy  int  // references that run conditionally or repeatedly: right of &&/||, inside a func literal
	addr  bool // some reference is the operand of &
}

// resolver counts the references to one outer variable. The analyzer runs
// without type information, so shadowing is resolved syntactically: whenever a
// scope declares the same name, the rest of that scope no longer refers to the
// outer variable and the walk stops there.
type resolver struct {
	name  string
	refs  refs
	depth int // > 0 while inside a lazy region
}

// scanRefs summarises the references in n to the variable called name that was
// declared outside n. It is deliberately conservative: when a name might be a
// reference (a composite literal key, for example) it is counted.
func scanRefs(n ast.Node, name string) refs {
	r := &resolver{name: name}
	r.visit(n)
	return r.refs
}

func (r *resolver) visit(n ast.Node) {
	switch n := n.(type) {
	case nil:
	case *ast.Ident:
		r.ident(n)
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
	case *ast.BinaryExpr:
		r.binaryExpr(n)
	case *ast.UnaryExpr:
		r.unaryExpr(n)
	default:
		r.children(n)
	}
}

func (r *resolver) ident(id *ast.Ident) {
	if id.Name != r.name {
		return
	}
	r.refs.count++
	if r.depth > 0 {
		r.refs.lazy++
	}
}

// visitAll skips nil interface values.
func (r *resolver) visitAll(nodes ...ast.Node) {
	for _, n := range nodes {
		r.visit(n)
	}
}

func (r *resolver) visitLazy(n ast.Node) {
	r.depth++
	r.visit(n)
	r.depth--
}

func (r *resolver) exprs(list []ast.Expr) {
	for _, e := range list {
		r.visit(e)
	}
}

// children routes the direct children of n through visit so the scope rules apply to them.
func (r *resolver) children(n ast.Node) {
	ast.Inspect(n, func(c ast.Node) bool {
		if c == n {
			return true
		}
		r.visit(c)
		return false
	})
}

// stmts stops at a statement that declares the name: it shadows the outer
// variable for the remainder of the list.
func (r *resolver) stmts(list []ast.Stmt) {
	for _, s := range list {
		if r.countInitsIfDeclares(s) {
			return
		}
		r.visit(s)
	}
}

// countInitsIfDeclares reports whether s introduces a new variable called
// r.name. When it does, it also counts the references in s's own initializers
// (Go scoping starts after the declaration, so in `n := n + 1` the right-hand n
// is the outer one); when it does not, it counts nothing.
func (r *resolver) countInitsIfDeclares(s ast.Stmt) bool {
	switch s := s.(type) {
	case *ast.AssignStmt:
		if s.Tok != token.DEFINE || !hasExpr(s.Lhs, r.name) {
			return false
		}
		r.exprs(s.Rhs)
		return true
	case *ast.DeclStmt:
		return r.countDeclInitsIfDeclares(s)
	}
	return false
}

// countDeclInitsIfDeclares handles var, const and type declarations. Specs
// before the declaring one in the same group are still in the outer scope.
func (r *resolver) countDeclInitsIfDeclares(ds *ast.DeclStmt) bool {
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
	if r.countInitsIfDeclares(s.Init) {
		return
	}
	r.visitAll(s.Init, s.Cond, s.Body, s.Else)
}

func (r *resolver) forStmt(s *ast.ForStmt) {
	if r.countInitsIfDeclares(s.Init) {
		return
	}
	r.visitAll(s.Init, s.Cond, s.Post, s.Body)
}

func (r *resolver) switchStmt(s *ast.SwitchStmt) {
	if r.countInitsIfDeclares(s.Init) {
		return
	}
	r.visitAll(s.Init, s.Tag, s.Body)
}

// typeSwitchStmt handles `switch n := x.(type)`: the clauses see the new n.
func (r *resolver) typeSwitchStmt(s *ast.TypeSwitchStmt) {
	if r.countInitsIfDeclares(s.Init) {
		return
	}
	r.visit(s.Init) // Init is evaluated before the clause variable exists
	if r.countInitsIfDeclares(s.Assign) {
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
	if r.countInitsIfDeclares(c.Comm) {
		return
	}
	r.visit(c.Comm)
	r.stmts(c.Body)
}

// funcLit treats the body as lazy: it runs once per call, not once per evaluation.
func (r *resolver) funcLit(fl *ast.FuncLit) {
	if hasField(fl.Type.Params, r.name) || hasField(fl.Type.Results, r.name) {
		return
	}
	// Type and Body are never nil for a FuncLit, so no typed-nil reaches visit.
	r.visit(fl.Type)
	r.visitLazy(fl.Body)
}

// binaryExpr treats the right operand of && and || as lazy: it may be skipped.
func (r *resolver) binaryExpr(b *ast.BinaryExpr) {
	r.visit(b.X)
	if b.Op == token.LAND || b.Op == token.LOR {
		r.visitLazy(b.Y)
		return
	}
	r.visit(b.Y)
}

// unaryExpr records `&name`: the call that would replace name is not addressable.
func (r *resolver) unaryExpr(u *ast.UnaryExpr) {
	if u.Op == token.AND && isIdentNamed(ast.Unparen(u.X), r.name) {
		r.refs.addr = true
	}
	r.visit(u.X)
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

func isIdentNamed(e ast.Expr, name string) bool {
	id, ok := e.(*ast.Ident)
	return ok && id.Name == name
}
