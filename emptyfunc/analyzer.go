// Package emptyfunc defines an analyzer that requires a comment inside function
// bodies that contain no statements.
package emptyfunc

import (
	"go/ast"
	"sort"

	"golang.org/x/tools/go/analysis"
)

// Analyzer reports functions whose body has no statements and no comment.
var Analyzer = &analysis.Analyzer{
	Name: "emptyfunc",
	Doc:  "requires a comment inside function bodies that contain no statements",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				if isEmpty(file, fn.Body) {
					pass.Reportf(
						fn.Name.Pos(),
						"function %q has an empty body; add a comment explaining why",
						fn.Name.Name,
					)
				}
			case *ast.FuncLit:
				if isEmpty(file, fn.Body) {
					pass.Reportf(fn.Pos(), "function has an empty body; add a comment explaining why")
				}
			}
			return true
		})
	}
	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// isEmpty reports whether body has no statements and no comment between its
// braces. A nil body (an external or assembly function) is never empty.
func isEmpty(file *ast.File, body *ast.BlockStmt) bool {
	if body == nil || len(body.List) != 0 {
		return false
	}
	// file.Comments is sorted by position, so the first group starting after
	// the opening brace is the only one that can lie inside the body.
	i := sort.Search(len(file.Comments), func(i int) bool {
		return file.Comments[i].Pos() > body.Lbrace
	})
	return i == len(file.Comments) || file.Comments[i].Pos() > body.Rbrace
}
