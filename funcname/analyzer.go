// Package funcname defines an analyzer that forbids underscores in function names.
// Config.IncludeVars extends the same rule to variables, constants, parameters,
// results and receivers.
package funcname

import (
	"go/ast"
	"go/token"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var validName = regexp.MustCompile(`^(_|[a-zA-Z0-9]+)$`)

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// IncludeVars also checks variable and constant names, parameters, results and
	// receivers against the same pattern as function names.
	IncludeVars bool
}

// Analyzer reports function names that contain underscores.
var Analyzer = New(Config{})

// New builds a funcname analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "funcname",
		Doc:      "forbids underscores in function names",
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

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}
	if cfg.IncludeVars {
		nodeFilter = append(nodeFilter,
			(*ast.FuncLit)(nil),
			(*ast.GenDecl)(nil),
			(*ast.AssignStmt)(nil),
			(*ast.RangeStmt)(nil),
		)
	}

	c := checker{pass: pass}
	insp.Preorder(nodeFilter, func(n ast.Node) {
		switch n := n.(type) {
		case *ast.FuncDecl:
			c.funcDecl(n, cfg.IncludeVars)
		case *ast.FuncLit:
			c.signature(n.Type)
		case *ast.GenDecl:
			c.genDecl(n)
		case *ast.AssignStmt:
			c.define(n.Tok, n.Lhs...)
		case *ast.RangeStmt:
			c.define(n.Tok, n.Key, n.Value)
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

type checker struct {
	pass *analysis.Pass
}

func (c checker) funcDecl(fn *ast.FuncDecl, includeVars bool) {
	c.ident(fn.Name, "function")
	if includeVars {
		c.fields(fn.Recv, "receiver")
		c.signature(fn.Type)
	}
}

// signature checks the named parameters and results of a function type.
// Type parameters are not checked.
func (c checker) signature(ft *ast.FuncType) {
	c.fields(ft.Params, "parameter")
	c.fields(ft.Results, "result")
}

// genDecl checks the names declared by var and const declarations.
func (c checker) genDecl(gd *ast.GenDecl) {
	var kind string
	switch gd.Tok {
	case token.VAR:
		kind = "variable"
	case token.CONST:
		kind = "constant"
	default:
		return
	}
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			c.ident(name, kind)
		}
	}
}

// define checks the left-hand side of a short variable declaration, which also
// covers type switch bindings and select receive clauses. Plain assignments and
// range clauses using = only reuse existing variables, so they are skipped.
// A := that redeclares an existing name is checked again, because telling it
// apart from a new variable needs type information this analyzer does not use.
func (c checker) define(tok token.Token, exprs ...ast.Expr) {
	if tok != token.DEFINE {
		return
	}
	for _, e := range exprs {
		if id, ok := e.(*ast.Ident); ok {
			c.ident(id, "variable")
		}
	}
}

func (c checker) fields(list *ast.FieldList, kind string) {
	if list == nil {
		return
	}
	for _, field := range list.List {
		for _, name := range field.Names {
			c.ident(name, kind)
		}
	}
}

func (c checker) ident(id *ast.Ident, kind string) {
	if validName.MatchString(id.Name) {
		return
	}
	c.pass.Reportf(id.Pos(),
		"Rename %s %q to match the regular expression %s",
		kind, id.Name, validName.String(),
	)
}
