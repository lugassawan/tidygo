// Package receiver resolves method receiver types, so analyzers that group or
// name methods by their type share one definition of "the receiver's type".
package receiver

import "go/ast"

// TypeName returns the base type name of a receiver expression, unwrapping
// pointers, parentheses and generic instantiations (T, *T, (T), T[K], *T[K, V]).
// It returns "" for a receiver that does not resolve to a local type name, such
// as the qualified pkg.T or a composite type.
func TypeName(expr ast.Expr) string {
	for {
		switch e := expr.(type) {
		case *ast.Ident:
			return e.Name
		case *ast.StarExpr:
			expr = e.X
		case *ast.ParenExpr:
			expr = e.X
		case *ast.IndexExpr:
			expr = e.X
		case *ast.IndexListExpr:
			expr = e.X
		default:
			return ""
		}
	}
}
