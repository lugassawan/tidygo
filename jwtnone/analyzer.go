// Package jwtnone defines an analyzer that reports uses of the JWT "none"
// algorithm (SigningMethodNone, UnsafeAllowNoneSignatureType), which lets an
// attacker forge claims. Matching is by resolved object, not by name.
// See docs/analyzers/jwtnone.md.
package jwtnone

import (
	"go/ast"
	"slices"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Analyzer reports references to jwt.SigningMethodNone and
// jwt.UnsafeAllowNoneSignatureType.
var Analyzer = &analysis.Analyzer{
	Name:     "jwtnone",
	Doc:      "reports use of the JWT \"none\" signing method, which produces and accepts unsigned tokens",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// jwtPackages are the import paths of the JWT libraries that define the none
// algorithm.
var jwtPackages = []string{
	"github.com/golang-jwt/jwt",
	"github.com/golang-jwt/jwt/v4",
	"github.com/golang-jwt/jwt/v5",
	"github.com/dgrijalva/jwt-go",
	"github.com/form3tech-oss/jwt-go",
}

// Diagnostics, keyed by the reported identifier.
const (
	messageSigningMethodNone = "jwt.SigningMethodNone produces unsigned tokens: " +
		"sign with an HMAC, RSA, ECDSA or EdDSA method"
	messageAllowNone = "jwt.UnsafeAllowNoneSignatureType lets unsigned tokens verify: " +
		"remove it and restrict the accepted signing methods"
)

var messages = map[string]string{
	"SigningMethodNone":            messageSigningMethodNone,
	"UnsafeAllowNoneSignatureType": messageAllowNone,
}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	insp.Preorder([]ast.Node{(*ast.Ident)(nil)}, func(n ast.Node) {
		id, ok := n.(*ast.Ident)
		if !ok {
			return
		}
		msg, found := messages[id.Name]
		if !found || pass.TypesInfo == nil {
			return
		}
		obj := pass.TypesInfo.Uses[id]
		if obj != nil && obj.Pkg() != nil && slices.Contains(jwtPackages, obj.Pkg().Path()) {
			pass.Reportf(id.Pos(), "%s", msg)
		}
	})

	return nil, nil //nolint:nilnil // analyzer produces no result
}
