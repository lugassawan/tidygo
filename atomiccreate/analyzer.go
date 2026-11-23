// Package atomiccreate defines an analyzer for the check-then-create race: after
// an os.Stat not-exist check, os.Create can clobber a file a concurrent writer
// made in between. The atomic form is os.OpenFile with O_CREATE|O_EXCL.
//
// It matches os.Stat/os.Lstat whose error is tested by os.IsNotExist or
// errors.Is(err, fs.ErrNotExist), alone or joined by &&, then os.Create,
// os.WriteFile or a non-exclusive constant-flag os.OpenFile of the same path in
// the then-body. Paths match as equal string constants or as the same
// identifier/selector chain (cfg.Path) not assigned in between. Helpers,
// aliased paths and negated checks are not followed.
package atomiccreate

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/lugassawan/tidygo/internal/callee"
)

// Analyzer reports os.Create, os.WriteFile and non-exclusive os.OpenFile calls
// that follow an os.Stat not-exist check on the same path.
var Analyzer = &analysis.Analyzer{
	Name:     "atomiccreate",
	Doc:      "requires os.OpenFile with O_CREATE|O_EXCL instead of creating a file after an os.Stat existence check",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

const message = "os.%s after os.Stat existence check is racy; use os.OpenFile with O_CREATE|O_EXCL"

// creators lists the os functions that create a file without exclusivity.
var creators = []string{"Create", "WriteFile", "OpenFile"}

// checker holds the per-function state needed to resolve errors and paths.
type checker struct {
	pass     *analysis.Pass
	info     *types.Info
	assigns  []*ast.AssignStmt // every assignment in the function, in source order
	reported map[token.Pos]bool
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
		if body := funcBody(n); body != nil {
			newChecker(pass, body).checkBody(body)
		}
	})
	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func funcBody(n ast.Node) *ast.BlockStmt {
	switch fn := n.(type) {
	case *ast.FuncDecl:
		return fn.Body
	case *ast.FuncLit:
		return fn.Body
	}
	return nil
}

func newChecker(pass *analysis.Pass, body *ast.BlockStmt) *checker {
	c := &checker{pass: pass, info: pass.TypesInfo, reported: map[token.Pos]bool{}}
	ast.Inspect(body, func(n ast.Node) bool {
		if as, ok := n.(*ast.AssignStmt); ok {
			c.assigns = append(c.assigns, as)
		}
		return true
	})
	return c
}

// inspectShallow visits root's subtree without entering nested function
// literals, which are analyzed as functions of their own.
func inspectShallow(root ast.Node, visit func(ast.Node)) {
	ast.Inspect(root, func(n ast.Node) bool {
		if _, nested := n.(*ast.FuncLit); nested {
			return false
		}
		if n != nil {
			visit(n)
		}
		return true
	})
}

func (c *checker) checkBody(body *ast.BlockStmt) {
	inspectShallow(body, func(n ast.Node) {
		if ifs, ok := n.(*ast.IfStmt); ok {
			c.checkIf(ifs)
		}
	})
}

// checkIf scans only the then-body: an else branch means the file exists.
func (c *checker) checkIf(ifs *ast.IfStmt) {
	stats := c.notExistStats(ifs.Cond)
	if len(stats) == 0 {
		return
	}
	inspectShallow(ifs.Body, func(n ast.Node) {
		if call, ok := n.(*ast.CallExpr); ok {
			c.checkCreate(call, stats)
		}
	})
}

// notExistStats returns the os.Stat/os.Lstat calls whose error is tested for
// non-existence by one of the && operands of cond.
func (c *checker) notExistStats(cond ast.Expr) []*ast.CallExpr {
	var stats []*ast.CallExpr
	for _, term := range conjuncts(cond) {
		id := c.notExistErr(term)
		if id == nil {
			continue
		}
		if stat := c.statFor(id, term.Pos()); stat != nil {
			stats = append(stats, stat)
		}
	}
	return stats
}

// conjuncts flattens a chain of && into its operands.
func conjuncts(e ast.Expr) []ast.Expr {
	switch x := e.(type) {
	case *ast.ParenExpr:
		return conjuncts(x.X)
	case *ast.BinaryExpr:
		if x.Op == token.LAND {
			return append(conjuncts(x.X), conjuncts(x.Y)...)
		}
	}
	return []ast.Expr{e}
}

// notExistErr returns the error identifier tested by os.IsNotExist(err) or
// errors.Is(err, fs.ErrNotExist), or nil when term is neither.
func (c *checker) notExistErr(term ast.Expr) *ast.Ident {
	call, ok := term.(*ast.CallExpr)
	if !ok {
		return nil
	}
	switch {
	case callee.IsFunc(c.info, call, "os", "IsNotExist") && len(call.Args) == 1:
		return identOf(call.Args[0])
	case callee.IsFunc(c.info, call, "errors", "Is") && len(call.Args) == 2 && c.isErrNotExist(call.Args[1]):
		return identOf(call.Args[0])
	}
	return nil
}

func identOf(e ast.Expr) *ast.Ident {
	id, ok := e.(*ast.Ident)
	if !ok {
		return nil
	}
	return id
}

// isErrNotExist reports whether e names fs.ErrNotExist or os.ErrNotExist, which
// are the same value.
func (c *checker) isErrNotExist(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	v, ok := c.info.Uses[sel.Sel].(*types.Var)
	if !ok || v.Pkg() == nil || v.Name() != "ErrNotExist" {
		return false
	}
	return v.Pkg().Path() == "io/fs" || v.Pkg().Path() == "os"
}

// statFor returns the os.Stat/os.Lstat call that last assigned the error id
// refers to before pos, or nil when that assignment is anything else.
func (c *checker) statFor(id *ast.Ident, pos token.Pos) *ast.CallExpr {
	errObj := c.info.ObjectOf(id)
	if errObj == nil {
		return nil
	}
	last := c.lastAssign(errObj, pos)
	if last == nil || len(last.Lhs) != 2 || len(last.Rhs) != 1 {
		return nil
	}
	if lhs := identOf(last.Lhs[1]); lhs == nil || c.info.ObjectOf(lhs) != errObj {
		return nil
	}
	call, ok := last.Rhs[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 1 || !callee.IsFunc(c.info, call, "os", "Stat", "Lstat") {
		return nil
	}
	return call
}

func (c *checker) lastAssign(obj types.Object, before token.Pos) *ast.AssignStmt {
	var last *ast.AssignStmt
	for _, as := range c.assigns {
		if as.Pos() >= before {
			break
		}
		if c.assignsTo(as, obj) {
			last = as
		}
	}
	return last
}

// assignsTo reports whether as writes to obj or to a field reached through it.
func (c *checker) assignsTo(as *ast.AssignStmt, obj types.Object) bool {
	for _, lhs := range as.Lhs {
		if c.pathRoot(lhs) == obj {
			return true
		}
	}
	return false
}

// pathRoot returns the object of the root identifier of an identifier or
// selector chain (p, cfg.Path), or nil for any other expression.
func (c *checker) pathRoot(e ast.Expr) types.Object {
	for {
		switch x := e.(type) {
		case *ast.Ident:
			return c.info.ObjectOf(x)
		case *ast.SelectorExpr:
			e = x.X
		default:
			return nil
		}
	}
}

func (c *checker) checkCreate(call *ast.CallExpr, stats []*ast.CallExpr) {
	name := c.creatorName(call)
	if name == "" || len(call.Args) == 0 || (name == "OpenFile" && !c.createsNonExclusively(call)) {
		return
	}
	for _, stat := range stats {
		if c.samePath(stat.Args[0], call.Args[0], stat.Pos(), call.Pos()) {
			c.report(call, name)
			return
		}
	}
}

func (c *checker) creatorName(call *ast.CallExpr) string {
	for _, name := range creators {
		if callee.IsFunc(c.info, call, "os", name) {
			return name
		}
	}
	return ""
}

// createsNonExclusively reports whether the os.OpenFile flag is a constant with
// O_CREATE set and O_EXCL clear. A non-constant flag is never reported.
func (c *checker) createsNonExclusively(call *ast.CallExpr) bool {
	if len(call.Args) < 2 {
		return false
	}
	tv, ok := c.info.Types[call.Args[1]]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.Int {
		return false
	}
	flag, exact := constant.Int64Val(tv.Value)
	create, okCreate := c.osConst(call, "O_CREATE")
	excl, okExcl := c.osConst(call, "O_EXCL")
	return exact && okCreate && okExcl && flag&create != 0 && flag&excl == 0
}

// osConst reads an integer constant of the os package the call was type-checked
// against. The values differ per GOOS, so they are taken from the analyzed code's
// view of os rather than from the platform the linter itself runs on.
func (c *checker) osConst(call *ast.CallExpr, name string) (int64, bool) {
	fn, _ := typeutil.Callee(c.info, call).(*types.Func)
	if fn == nil || fn.Pkg() == nil {
		return 0, false
	}
	k, ok := fn.Pkg().Scope().Lookup(name).(*types.Const)
	if !ok {
		return 0, false
	}
	return constant.Int64Val(constant.ToInt(k.Val()))
}

// samePath reports whether a and b denote the same path between from and to.
func (c *checker) samePath(a, b ast.Expr, from, to token.Pos) bool {
	if sa, ok := c.constString(a); ok {
		sb, ok := c.constString(b)
		return ok && sa == sb
	}
	root := c.pathRoot(a)
	if root == nil || root != c.pathRoot(b) || types.ExprString(a) != types.ExprString(b) {
		return false
	}
	return !c.reassigned(root, from, to)
}

func (c *checker) constString(e ast.Expr) (string, bool) {
	tv, ok := c.info.Types[e]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}

// reassigned reports whether obj, or a field of it, is assigned strictly between
// from and to.
func (c *checker) reassigned(obj types.Object, from, to token.Pos) bool {
	for _, as := range c.assigns {
		if as.Pos() > from && as.Pos() < to && c.assignsTo(as, obj) {
			return true
		}
	}
	return false
}

func (c *checker) report(call *ast.CallExpr, name string) {
	if c.reported[call.Pos()] {
		return
	}
	c.reported[call.Pos()] = true
	c.pass.Reportf(call.Pos(), message, name)
}
