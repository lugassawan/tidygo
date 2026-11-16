// Package nolateexport defines an analyzer that enforces export-first ordering.
//
// Standalone functions: exported functions must appear before unexported ones.
//
// Methods: within a file, exported methods of a receiver type must appear
// before unexported methods of that same type, wherever the type is declared.
// Methods of different types are tracked independently.
//
// Optionally (Config.UnexportedFuncsLast), exported methods must also appear
// before unexported standalone functions.
package nolateexport

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// UnexportedFuncsLast additionally requires exported methods to appear
	// before unexported standalone functions in the same file.
	UnexportedFuncsLast bool
}

// Analyzer reports exported functions and methods that appear after unexported
// ones, using the default (zero) Config.
var Analyzer = New(Config{})

// New builds a nolateexport analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "nolateexport",
		Doc: "forbids exported functions after unexported standalone functions, " +
			"and exported methods after unexported methods of the same receiver type",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

// fileState tracks what has been seen so far while walking one file's declarations.
type fileState struct {
	seenUnexportedFunc   bool
	seenUnexportedMethod map[string]bool // keyed by receiver base type name
}

func run(pass *analysis.Pass, cfg Config) (any, error) {
	for _, file := range pass.Files {
		state := &fileState{seenUnexportedMethod: map[string]bool{}}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv != nil {
				state.checkMethod(pass, cfg, fn)
			} else {
				state.checkFunc(pass, fn)
			}
		}
	}
	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func (s *fileState) checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	name := fn.Name.Name
	if name == "init" {
		return
	}
	if !ast.IsExported(name) {
		s.seenUnexportedFunc = true
		return
	}
	if s.seenUnexportedFunc {
		pass.Reportf(fn.Pos(), "exported function %s should appear before unexported functions", name)
	}
}

func (s *fileState) checkMethod(pass *analysis.Pass, cfg Config, fn *ast.FuncDecl) {
	if len(fn.Recv.List) == 0 {
		return
	}
	typ := recvTypeName(fn.Recv.List[0].Type)
	name := fn.Name.Name
	if !ast.IsExported(name) {
		s.seenUnexportedMethod[typ] = true
		return
	}
	switch {
	case s.seenUnexportedMethod[typ]:
		pass.Reportf(fn.Pos(), "exported method %s.%s should appear before unexported methods of %s", typ, name, typ)
	case cfg.UnexportedFuncsLast && s.seenUnexportedFunc:
		pass.Reportf(fn.Pos(), "exported method %s.%s should appear before unexported standalone functions", typ, name)
	}
}

// recvTypeName returns the base type name of a receiver expression, unwrapping
// pointers, parentheses and generic instantiations (T, *T, T[K], *T[K, V]).
func recvTypeName(expr ast.Expr) string {
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
