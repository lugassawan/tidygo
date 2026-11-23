// Package scrypt is a minimal stub of golang.org/x/crypto/scrypt for analyzer tests.
package scrypt

func Key(password, salt []byte, N, r, p, keyLen int) ([]byte, error) {
	return nil, nil
}
