// Package callee resolves call expressions to the functions and methods they
// invoke, using type information so that import aliases, dot imports and
// shadowed identifiers cannot fool the match.
package callee

import (
	"go/ast"
	"go/types"
	"slices"

	"golang.org/x/tools/go/types/typeutil"
)

// IsFunc reports whether call invokes the package-level function pkgPath.Name
// for any of the given names. Methods and calls through function values never
// match.
func IsFunc(info *types.Info, call *ast.CallExpr, pkgPath string, names ...string) bool {
	fn := calledFunc(info, call)
	if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != pkgPath {
		return false
	}
	if sig, ok := fn.Type().(*types.Signature); !ok || sig.Recv() != nil {
		return false
	}
	return slices.Contains(names, fn.Name())
}

// IsMethod reports whether call invokes the method on pkgPath.typeName or
// *pkgPath.typeName.
func IsMethod(info *types.Info, call *ast.CallExpr, pkgPath, typeName, method string) bool {
	fn := calledFunc(info, call)
	if fn == nil || fn.Name() != method {
		return false
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}
	recv := sig.Recv().Type()
	if ptr, ok := recv.(*types.Pointer); ok {
		recv = ptr.Elem()
	}
	named, ok := recv.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == pkgPath && obj.Name() == typeName
}

// calledFunc returns the statically known function or method call invokes, or
// nil for conversions, builtins and calls through function values.
func calledFunc(info *types.Info, call *ast.CallExpr) *types.Func {
	if info == nil {
		return nil
	}
	fn, _ := typeutil.Callee(info, call).(*types.Func)
	return fn
}
