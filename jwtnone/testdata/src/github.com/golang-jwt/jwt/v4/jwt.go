// Package jwt is a minimal stub of github.com/golang-jwt/jwt/v4 for analyzer tests.
package jwt

type SigningMethod interface {
	Alg() string
}

type signingMethodNone struct{}

func (signingMethodNone) Alg() string { return "none" }

type unsafeNoneMagicConstant string

var SigningMethodNone SigningMethod = signingMethodNone{}

const UnsafeAllowNoneSignatureType unsafeNoneMagicConstant = "none signing method allowed"
