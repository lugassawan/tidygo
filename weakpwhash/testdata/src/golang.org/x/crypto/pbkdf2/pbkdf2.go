// Package pbkdf2 is a minimal stub of golang.org/x/crypto/pbkdf2 for analyzer tests.
package pbkdf2

import "hash"

func Key(password, salt []byte, iter, keyLen int, h func() hash.Hash) []byte {
	return nil
}
