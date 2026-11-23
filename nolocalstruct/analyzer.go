// Package nolocalstruct defines an analyzer that forbids named struct type
// declarations inside function bodies. Config.AllTypes extends the rule to
// every named type declared in a function.
// See docs/analyzers/nolocalstruct.md.
package nolocalstruct

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const (
	// structMessage is kept byte-identical across releases so existing nolint
	// directives and exclusion rules keep matching.
	structMessage = "named struct type %q should be declared at package level, not inside a function"
	typeMessage   = "named type %q should be declared at package level, not inside a function"
)

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// AllTypes also reports local interfaces, defined types and generic types,
	// not just structs. Aliases are still skipped unless they alias a struct.
	AllTypes bool
}

// Analyzer reports named struct types declared inside function bodies.
var Analyzer = New(Config{})

// New builds a nolocalstruct analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "nolocalstruct",
		Doc:      "forbids named struct type declarations inside function bodies",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

func run(pass *analysis.Pass, cfg Config) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	// A DeclStmt only occurs inside a function body, so filtering on it visits
	// every local declaration exactly once, including those in nested literals.
	nodeFilter := []ast.Node{(*ast.DeclStmt)(nil)}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		ds, ok := n.(*ast.DeclStmt)
		if !ok {
			return
		}
		gd, ok := ds.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			return
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if msg, report := messageFor(ts, cfg); report {
				pass.Reportf(ts.Pos(), msg, ts.Name.Name)
			}
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

func messageFor(ts *ast.TypeSpec, cfg Config) (string, bool) {
	if _, isStruct := ts.Type.(*ast.StructType); isStruct {
		return structMessage, true
	}
	if !cfg.AllTypes || ts.Assign.IsValid() {
		return "", false
	}
	return typeMessage, true
}
