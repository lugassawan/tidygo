package callee_test

import (
	"go/ast"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/callee"
)

// probe reports the calls that callee.IsFunc and callee.IsMethod resolve, so
// the testdata can pin both the positive and the negative cases.
var probe = &analysis.Analyzer{
	Name:     "probe",
	Doc:      "exercises callee.IsFunc and callee.IsMethod",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run: func(pass *analysis.Pass) (any, error) {
		insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		if !ok {
			return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
		}
		insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return
			}
			switch {
			case callee.IsFunc(pass.TypesInfo, call, "context", "WithCancel"):
				pass.Reportf(call.Pos(), "func: context.WithCancel")
			case callee.IsFunc(pass.TypesInfo, call, "context", "WithTimeout", "WithDeadline"):
				pass.Reportf(call.Pos(), "func: context.WithTimeout")
			case callee.IsMethod(pass.TypesInfo, call, "bytes", "Buffer", "Reset"):
				pass.Reportf(call.Pos(), "method: bytes.Buffer.Reset")
			}
		})
		return nil, nil //nolint:nilnil // analysis.Analyzer.Run conventionally returns (nil, nil)
	},
}

func TestHelpers(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), probe, "example")
}
