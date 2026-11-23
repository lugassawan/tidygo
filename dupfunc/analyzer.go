// Package dupfunc defines an analyzer that reports functions whose bodies and
// signatures are identical to an earlier function in the same package. Only bodies of at
// least DefaultMinStmts top-level statements are compared (Config.MinStmts
// overrides the minimum), test files and generated files are skipped, and
// function literals and init functions are ignored.
//
// The signatures must match too: functions whose parameter or result types
// differ, such as clampInt and clampFloat, cannot be merged without generics and
// are not reported. Types are compared as the type checker resolves them, so
// byte matches uint8, an alias matches its target, and generic functions match up
// to the names of their type parameters. Receivers and parameter names are not
// compared, so methods on different types with the same signature are reported.
// Methods of different generic types match only when their signatures do not
// mention the receiver's type parameters, which are distinct for each type.
//
// Bodies are compared structurally: positions, comments and layout are
// ignored, but identifier names, literals and operators must match exactly, so
// a renamed local variable makes a body different. Each later duplicate is
// reported against the first earlier function with an identical signature, by
// file name and offset, so a //nolint directive belongs on the later function.
package dupfunc

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/lugassawan/tidygo/internal/receiver"
)

// DefaultMinStmts is the smallest number of top-level statements a body needs to
// be compared when Config.MinStmts is not positive.
const DefaultMinStmts = 3

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// MinStmts is the smallest number of top-level statements a body needs to be
	// compared. Values <= 0 select DefaultMinStmts.
	MinStmts int
}

// Analyzer reports functions with the same body and signature as an earlier
// function in the package, comparing bodies of at least DefaultMinStmts statements.
var Analyzer = New(Config{})

// New builds a dupfunc analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	minStmts := cfg.MinStmts
	if minStmts <= 0 {
		minStmts = DefaultMinStmts
	}
	return &analysis.Analyzer{
		Name: "dupfunc",
		Doc:  "forbids functions whose bodies and signatures are identical to an earlier function in the same package",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, minStmts)
		},
	}
}

// candidate is a function eligible for comparison together with its source position.
type candidate struct {
	decl *ast.FuncDecl
	pos  token.Position
	// fn is the type checker's object for the function, nil when it has none.
	fn *types.Func
	// sig is the canonical signature, resolved on first use; see signature.
	sig         *types.Signature
	sigResolved bool
}

func run(pass *analysis.Pass, minStmts int) (any, error) {
	candidates := collect(pass, minStmts)

	// Sort by position so the "first" function, and so every report, is
	// deterministic regardless of the order the files were loaded in.
	slices.SortFunc(candidates, func(a, b *candidate) int {
		return cmp.Or(
			strings.Compare(a.pos.Filename, b.pos.Filename),
			cmp.Compare(a.pos.Offset, b.pos.Offset),
		)
	})

	// Functions with the same body are grouped, and within a group each
	// signature class keeps only its first function: a later function is
	// reported against the first earlier one whose signature is identical.
	classes := make(map[string][]*candidate)
	for _, c := range candidates {
		key := bodyKey(c.decl.Body)
		orig := findMatch(classes[key], c)
		if orig == nil {
			classes[key] = append(classes[key], c)
			continue
		}
		pass.Reportf(c.decl.Name.Pos(), "function %s has the same body as %s (%s:%d)",
			funcName(c.decl), funcName(orig.decl), filepath.Base(orig.pos.Filename), orig.pos.Line)
	}

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// collect returns the functions large enough to compare, skipping test and generated files.
func collect(pass *analysis.Pass, minStmts int) []*candidate {
	var out []*candidate
	for _, f := range pass.Files {
		if strings.HasSuffix(pass.Fset.File(f.Pos()).Name(), "_test.go") || ast.IsGenerated(f) {
			continue
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil || len(fd.Body.List) < minStmts || isInit(fd) {
				continue
			}
			fn, _ := pass.TypesInfo.Defs[fd.Name].(*types.Func)
			out = append(out, &candidate{decl: fd, pos: pass.Fset.Position(fd.Pos()), fn: fn})
		}
	}
	return out
}

// findMatch returns the first function in group whose signature is identical to
// c's, or nil. A function without a signature matches nothing: a missed report
// costs less than a false one.
func findMatch(group []*candidate, c *candidate) *candidate {
	if len(group) == 0 {
		return nil
	}
	sig := c.signature()
	if sig == nil {
		return nil
	}
	for _, g := range group {
		if gs := g.signature(); gs != nil && types.Identical(gs, sig) {
			return g
		}
	}
	return nil
}

// signature returns the canonical signature of c, or nil when the type checker
// recorded none. It is resolved lazily, so a function that shares its body with
// no other never pays for it.
func (c *candidate) signature() *types.Signature {
	if !c.sigResolved {
		c.sigResolved = true
		if c.fn != nil {
			c.sig = canonicalSignature(c.fn)
		}
	}
	return c.sig
}

// canonicalSignature returns the signature of fn. Each method of a generic type
// declares its own receiver type parameters, so (*box[T]).put(x T) and
// (*box[X]).other(x X) would never be identical; the signature is therefore
// re-read through the receiver type instantiated with its own type parameters,
// which gives every method of the type one shared set. Where that fails, the raw
// signature is used.
func canonicalSignature(fn *types.Func) *types.Signature {
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.RecvTypeParams().Len() == 0 {
		return sig
	}
	if canon := instantiatedMethodSignature(fn, sig); canon != nil {
		return canon
	}
	return sig
}

// instantiatedMethodSignature looks fn up on its receiver's base type
// instantiated with that type's own type parameters, or returns nil.
func instantiatedMethodSignature(fn *types.Func, sig *types.Signature) *types.Signature {
	t := types.Unalias(sig.Recv().Type())
	if p, ok := t.(*types.Pointer); ok {
		t = types.Unalias(p.Elem())
	}
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}
	origin := named.Origin()
	tparams := origin.TypeParams()
	targs := make([]types.Type, tparams.Len())
	for i := range targs {
		targs[i] = tparams.At(i)
	}
	inst, err := types.Instantiate(nil, origin, targs, false)
	if err != nil {
		return nil
	}
	obj, _, _ := types.LookupFieldOrMethod(types.NewPointer(inst), true, fn.Pkg(), fn.Name())
	m, ok := obj.(*types.Func)
	if !ok {
		return nil
	}
	canon, _ := m.Type().(*types.Signature)
	return canon
}

// isInit reports whether fd is a package-level init function. A package may
// declare several, and they cannot be merged by calling one from another.
func isInit(fd *ast.FuncDecl) bool {
	return fd.Recv == nil && fd.Name.Name == "init"
}

// bodyKey serialises the shape of body: node kinds, identifier names, literal
// values and operator tokens. A nil node from ast.Inspect closes its parent,
// so the string encodes the tree structure and not just the node sequence.
func bodyKey(body *ast.BlockStmt) string {
	var b strings.Builder
	ast.Inspect(body, func(n ast.Node) bool {
		switch n.(type) {
		case nil:
			b.WriteByte(')')
			return true
		case *ast.CommentGroup, *ast.Comment:
			// Walk reaches the Doc and Comment fields of local declarations and fields.
			return false
		}
		fmt.Fprintf(&b, "(%T", n)
		writeDetail(&b, n)
		return true
	})
	return b.String()
}

// writeDetail appends what the node kind alone does not capture. Where a
// missing child leaves no trace in the walk (slice bounds, ellipsis), it records
// which parts are present.
func writeDetail(b *strings.Builder, n ast.Node) {
	switch x := n.(type) {
	case *ast.Ident:
		fmt.Fprintf(b, " %s", x.Name)
	case *ast.BasicLit:
		fmt.Fprintf(b, " %s %q", x.Kind, x.Value)
	case *ast.BinaryExpr:
		fmt.Fprintf(b, " %s", x.Op)
	case *ast.UnaryExpr:
		fmt.Fprintf(b, " %s", x.Op)
	case *ast.AssignStmt:
		// The left-hand count tells `a, b, c = f()` from `a, b = c, f()`.
		fmt.Fprintf(b, " %s %d", x.Tok, len(x.Lhs))
	case *ast.ForStmt:
		// `for i++; ; {}` and `for ; ; i++ {}` walk the same children.
		fmt.Fprintf(b, " %t %t %t", x.Init != nil, x.Cond != nil, x.Post != nil)
	case *ast.ValueSpec:
		fmt.Fprintf(b, " %d %t", len(x.Names), x.Type != nil)
	case *ast.CommClause:
		// Comm is a statement like the ones in Body, so `case <-a:` and `default: <-a` look alike.
		fmt.Fprintf(b, " %t", x.Comm != nil)
	case *ast.CompositeLit:
		// An elided type (the inner literal of [][]int{{1}}) leaves no child to tell it from a typed one.
		fmt.Fprintf(b, " %t", x.Type != nil)
	case *ast.IncDecStmt:
		fmt.Fprintf(b, " %s", x.Tok)
	case *ast.BranchStmt:
		fmt.Fprintf(b, " %s", x.Tok)
	case *ast.GenDecl:
		fmt.Fprintf(b, " %s", x.Tok)
	case *ast.RangeStmt:
		fmt.Fprintf(b, " %s", x.Tok)
	case *ast.ChanType:
		fmt.Fprintf(b, " %d", x.Dir)
	case *ast.CallExpr:
		fmt.Fprintf(b, " %t", x.Ellipsis.IsValid())
	case *ast.TypeSpec:
		fmt.Fprintf(b, " %t", x.Assign.IsValid())
	case *ast.SliceExpr:
		fmt.Fprintf(b, " %t %t %t", x.Low != nil, x.High != nil, x.Max != nil)
	}
}

// funcName renders a function as Name, or Recv.Name for a method.
func funcName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}
	recv := receiver.TypeName(fd.Recv.List[0].Type)
	// A qualified pkg.T receiver has no local type name; use the bare name, not ".name".
	if recv == "" {
		return fd.Name.Name
	}
	return recv + "." + fd.Name.Name
}
