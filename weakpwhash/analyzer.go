// Package weakpwhash defines an analyzer that reports password-hashing calls
// with a constant salt or a constant work factor below the current guidance.
// A constant salt lets one precomputed table attack every password; a low cost
// makes brute force cheap. Only compile-time constants are judged, so values
// read from variables or configuration are never reported.
package weakpwhash

import (
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/callee"
)

// Analyzer reports constant salts and low constant costs in password hashing.
var Analyzer = &analysis.Analyzer{
	Name:     "weakpwhash",
	Doc:      "reports password hashing with a constant salt or a constant cost below the recommended minimum",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// noArg marks a rule that has no salt or cost argument to inspect.
const noArg = -1

// rule describes one password-hashing function: where its salt and cost
// arguments sit and which constant costs are too weak. A cost is weak when
// from <= cost < min; from is 0 except for bcrypt, which silently replaces a
// cost below MinCost with DefaultCost.
type rule struct {
	pkg      string
	names    []string
	saltArg  int
	costArg  int
	costName string
	from     int64
	min      int64
}

// keyFunc is the name of the key-derivation entry point shared by the pbkdf2,
// scrypt and argon2 packages.
const keyFunc = "Key"

var rules = []rule{
	{
		pkg: "crypto/pbkdf2", names: []string{keyFunc},
		saltArg: 2, costArg: 3, costName: "iteration count", min: 100_000,
	},
	{
		pkg: "golang.org/x/crypto/pbkdf2", names: []string{keyFunc},
		saltArg: 1, costArg: 2, costName: "iteration count", min: 100_000,
	},
	{
		pkg: "golang.org/x/crypto/scrypt", names: []string{keyFunc},
		saltArg: 1, costArg: 2, costName: "N", min: 1 << 15,
	},
	{
		pkg: "golang.org/x/crypto/argon2", names: []string{keyFunc, "IDKey"},
		saltArg: 1, costArg: 3, costName: "memory", min: 19_456,
	},
	{
		pkg: "golang.org/x/crypto/bcrypt", names: []string{"GenerateFromPassword"},
		saltArg: noArg, costArg: 1, costName: "cost", from: 4, min: 10,
	},
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}
		for _, r := range rules {
			if callee.IsFunc(pass.TypesInfo, call, r.pkg, r.names...) {
				checkCall(pass, call, r)
				return
			}
		}
	})

	return nil, nil //nolint:nilnil // analyzer produces no result
}

func checkCall(pass *analysis.Pass, call *ast.CallExpr, r rule) {
	fn := callee.Func(pass.TypesInfo, call)
	name := fn.Pkg().Name() + "." + fn.Name()

	if arg := argAt(call, r.saltArg); arg != nil && isConstantBytes(pass.TypesInfo, arg) {
		pass.Reportf(arg.Pos(), "%s uses a constant salt: generate a random per-password salt with crypto/rand", name)
	}

	arg := argAt(call, r.costArg)
	if arg == nil {
		return
	}
	if cost, ok := constInt(pass.TypesInfo, arg); ok && cost >= r.from && cost < r.min {
		pass.Reportf(arg.Pos(), "%s %s %d is below the minimum %d", name, r.costName, cost, r.min)
	}
}

// argAt returns the call argument at index i, or nil when the rule has no such
// argument or the call is too short to have it.
func argAt(call *ast.CallExpr, i int) ast.Expr {
	if i == noArg || i >= len(call.Args) {
		return nil
	}
	return call.Args[i]
}

// constInt returns the compile-time integer value of e.
func constInt(info *types.Info, e ast.Expr) (int64, bool) {
	tv, ok := info.Types[e]
	if !ok || tv.Value == nil {
		return 0, false
	}
	v := constant.ToInt(tv.Value)
	if v.Kind() != constant.Int {
		return 0, false
	}
	return constant.Int64Val(v)
}

// isConstantBytes reports whether e is a byte slice whose content is fixed at
// compile time: nil, a conversion of a constant, or a literal of constants.
// Variables are not followed.
func isConstantBytes(info *types.Info, e ast.Expr) bool {
	e = ast.Unparen(e)
	if tv, ok := info.Types[e]; ok && tv.IsNil() {
		return true
	}

	switch x := e.(type) {
	case *ast.CallExpr:
		if tv, ok := info.Types[x.Fun]; !ok || !tv.IsType() || len(x.Args) != 1 {
			return false
		}
		tv, ok := info.Types[x.Args[0]]
		return ok && tv.Value != nil
	case *ast.CompositeLit:
		for _, elt := range x.Elts {
			if tv, ok := info.Types[elt]; !ok || tv.Value == nil {
				return false
			}
		}
		return true
	}
	return false
}
