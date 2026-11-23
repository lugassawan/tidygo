// Package dupfunc defines an analyzer that reports functions whose bodies are
// identical to an earlier function in the same package. Only bodies of at
// least DefaultMinStmts top-level statements are compared, test files and
// generated files are skipped, and function literals are ignored.
//
// Bodies are compared structurally: positions, comments and layout are
// ignored, but identifier names, literals and operators must match exactly, so
// a renamed local variable makes a body different. Each later duplicate is
// reported against the first function by file name and offset, so a //nolint
// directive belongs on the later function.
package dupfunc

import (
	"cmp"
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// DefaultMinStmts is the smallest number of top-level statements a body needs to be compared.
const DefaultMinStmts = 3

// Analyzer reports functions with the same body as an earlier function in the package.
var Analyzer = &analysis.Analyzer{
	Name: "dupfunc",
	Doc:  "forbids functions whose bodies are identical to an earlier function in the same package",
	Run:  run,
}

// candidate is a function eligible for comparison together with its source position.
type candidate struct {
	decl *ast.FuncDecl
	pos  token.Position
}

func run(pass *analysis.Pass) (any, error) {
	candidates := collect(pass)

	// Sort by position so the "first" function, and so every report, is
	// deterministic regardless of the order the files were loaded in.
	slices.SortFunc(candidates, func(a, b candidate) int {
		return cmp.Or(
			strings.Compare(a.pos.Filename, b.pos.Filename),
			cmp.Compare(a.pos.Offset, b.pos.Offset),
		)
	})

	first := make(map[string]candidate)
	for _, c := range candidates {
		key := bodyKey(c.decl.Body)
		orig, seen := first[key]
		if !seen {
			first[key] = c
			continue
		}
		pass.Reportf(c.decl.Name.Pos(), "function %s has the same body as %s (%s:%d)",
			funcName(c.decl), funcName(orig.decl), filepath.Base(orig.pos.Filename), orig.pos.Line)
	}

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// collect returns the functions large enough to compare, skipping test and generated files.
func collect(pass *analysis.Pass) []candidate {
	var out []candidate
	for _, f := range pass.Files {
		if strings.HasSuffix(pass.Fset.File(f.Pos()).Name(), "_test.go") || ast.IsGenerated(f) {
			continue
		}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Body == nil || len(fd.Body.List) < DefaultMinStmts {
				continue
			}
			out = append(out, candidate{decl: fd, pos: pass.Fset.Position(fd.Pos())})
		}
	}
	return out
}

// bodyKey serialises the shape of body: node kinds, identifier names, literal
// values and operator tokens. A nil node from ast.Inspect closes its parent,
// so the string encodes the tree structure and not just the node sequence.
func bodyKey(body *ast.BlockStmt) string {
	var b strings.Builder
	ast.Inspect(body, func(n ast.Node) bool {
		if n == nil {
			b.WriteByte(')')
			return true
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
		fmt.Fprintf(b, " %s", x.Tok)
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
	recv := recvTypeName(fd.Recv.List[0].Type)
	if recv == "" {
		return fd.Name.Name
	}
	return recv + "." + fd.Name.Name
}

// recvTypeName returns the base type name of a receiver, looking through
// pointers, parentheses and type parameters.
func recvTypeName(e ast.Expr) string {
	switch t := e.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return recvTypeName(t.X)
	case *ast.ParenExpr:
		return recvTypeName(t.X)
	case *ast.IndexExpr:
		return recvTypeName(t.X)
	case *ast.IndexListExpr:
		return recvTypeName(t.X)
	}
	return ""
}
