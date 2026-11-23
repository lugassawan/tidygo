// Package jwt is a minimal stub of github.com/golang-jwt/jwt/v5 for analyzer tests.
package jwt

type SigningMethod interface {
	Alg() string
}

type signingMethodNone struct{}

func (signingMethodNone) Alg() string { return "none" }

type signingMethodHMAC struct{}

func (signingMethodHMAC) Alg() string { return "HS256" }

type unsafeNoneMagicConstant string

var (
	SigningMethodNone  SigningMethod = signingMethodNone{}
	SigningMethodHS256 SigningMethod = signingMethodHMAC{}
)

const UnsafeAllowNoneSignatureType unsafeNoneMagicConstant = "none signing method allowed"
