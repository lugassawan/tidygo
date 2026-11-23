// Package aeadcipher defines an analyzer that reports unauthenticated block
// cipher modes used for encryption. CBC encryption and CTR give confidentiality
// but no integrity, so ciphertext can be tampered with undetected; an AEAD such
// as GCM or ChaCha20-Poly1305 authenticates it. CBC decryption is not reported
// because reading legacy data needs it.
package aeadcipher

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/lugassawan/tidygo/internal/callee"
)

// Analyzer reports cipher.NewCBCEncrypter and cipher.NewCTR calls.
var Analyzer = &analysis.Analyzer{
	Name:     "aeadcipher",
	Doc:      "reports unauthenticated cipher modes (CBC encryption, CTR) where an AEAD should be used",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

const cipherPkg = "crypto/cipher"

// unauthenticatedModes are the crypto/cipher constructors that encrypt without
// authenticating. CFB and OFB are left to staticcheck SA1019, which reports them
// as deprecated.
var unauthenticatedModes = []string{"NewCBCEncrypter", "NewCTR"}

func run(pass *analysis.Pass) (any, error) {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil, nil //nolint:nilnil // analysis framework guarantees type
	}

	insp.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(n ast.Node) {
		call, ok := n.(*ast.CallExpr)
		if !ok || !callee.IsFunc(pass.TypesInfo, call, cipherPkg, unauthenticatedModes...) {
			return
		}
		pass.Reportf(call.Pos(),
			"cipher.%s produces unauthenticated ciphertext: use an AEAD such as cipher.NewGCM or chacha20poly1305",
			callee.Func(pass.TypesInfo, call).Name())
	})

	return nil, nil //nolint:nilnil // analyzer produces no result
}
