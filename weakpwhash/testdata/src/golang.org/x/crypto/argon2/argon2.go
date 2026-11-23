// Package argon2 is a minimal stub of golang.org/x/crypto/argon2 for analyzer tests.
package argon2

func Key(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte {
	return nil
}

func IDKey(password, salt []byte, time, memory uint32, threads uint8, keyLen uint32) []byte {
	return nil
}
