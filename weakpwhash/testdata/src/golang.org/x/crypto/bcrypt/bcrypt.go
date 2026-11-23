// Package bcrypt is a minimal stub of golang.org/x/crypto/bcrypt for analyzer tests.
package bcrypt

const (
	MinCost     int = 4
	MaxCost     int = 31
	DefaultCost int = 10
)

func GenerateFromPassword(password []byte, cost int) ([]byte, error) {
	return nil, nil
}
