// Package funcname defines an analyzer that forbids underscores in function names.
package funcname

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var validName = regexp.MustCompile(`^(_|[a-zA-Z0-9]+)$`)

// Analyzer reports function names that contain underscores.
var Analyzer = &analysis.Analyzer{
	Name:     "funcname",
	Doc:      "forbids underscores in function names",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}
		name := fn.Name.Name

		if !validName.MatchString(name) {
			pass.Reportf(fn.Name.Pos(),
				"Rename function %q to match the regular expression %s",
				name, validName.String(),
			)
		}
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}
