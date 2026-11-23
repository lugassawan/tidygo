// Package funcs visits the function bodies of a package, so analyzers that
// reason about one function at a time share a single traversal.
package funcs

import (
	"go/ast"

	"golang.org/x/tools/go/ast/inspector"
)

// Each calls fn with the signature and body of every function declaration and
// function literal that has a body. A literal nested in a function is visited on
// its own, after the function that contains it.
func Each(insp *inspector.Inspector, fn func(ftype *ast.FuncType, body *ast.BlockStmt)) {
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch f := n.(type) {
		case *ast.FuncDecl:
			if f.Body != nil {
				fn(f.Type, f.Body)
			}
		case *ast.FuncLit:
			fn(f.Type, f.Body)
		}
	})
}
