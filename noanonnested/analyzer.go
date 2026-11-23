// Package noanonnested defines an analyzer that reports anonymous struct types
// nested too deeply inside other anonymous structs. Config.Max overrides the
// default limit of DefaultMax. See docs/analyzers.md for how depth is counted.
package noanonnested

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// DefaultMax is the nesting limit used when Config.Max is not positive.
const DefaultMax = 1

// Config controls the optional behaviour of the analyzer.
type Config struct {
	// Max is the deepest allowed nesting of anonymous structs. Values <= 0 select DefaultMax.
	Max int
}

// Analyzer reports anonymous structs nested deeper than DefaultMax levels.
var Analyzer = New(Config{})

// New builds a noanonnested analyzer with the given configuration.
func New(cfg Config) *analysis.Analyzer {
	limit := cfg.Max
	if limit <= 0 {
		limit = DefaultMax
	}
	return &analysis.Analyzer{
		Name:     "noanonnested",
		Doc:      "forbids anonymous structs nested too deeply inside other anonymous structs",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, limit)
		},
	}
}

func run(pass *analysis.Pass, limit int) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	nodeFilter := []ast.Node{(*ast.StructType)(nil)}

	insp.WithStack(nodeFilter, func(n ast.Node, push bool, stack []ast.Node) bool {
		if !push {
			return true
		}
		// Reporting only at limit+1 marks the struct to extract and keeps one
		// diagnostic per nesting chain instead of one per level.
		if d := depth(stack); d == limit+1 {
			pass.Reportf(n.Pos(), "anonymous struct nested %d levels deep, max is %d; extract a named type", d, limit)
		}
		return true
	})

	return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
}

// depth counts the anonymous structs from the struct at the top of stack up
// through its enclosing type expressions. A struct declared by a type spec is a
// named (or aliased) type and contributes nothing; any other parent ends the walk.
func depth(stack []ast.Node) int {
	count := 0
	for i, node := range slices.Backward(stack) {
		switch node.(type) {
		case *ast.StructType:
			if i > 0 {
				if _, declared := stack[i-1].(*ast.TypeSpec); declared {
					return count
				}
			}
			count++
		case *ast.Field, *ast.FieldList, *ast.StarExpr, *ast.ArrayType,
			*ast.MapType, *ast.ChanType, *ast.ParenExpr:
			// Transparent wrappers between one struct and the next.
		default:
			return count
		}
	}
	return count
}
