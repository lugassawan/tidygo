package example

import (
	"crypto/aes"
	"crypto/cipher"
	stdcipher "crypto/cipher"
)

func cbcEncrypt(key, iv []byte) cipher.BlockMode {
	block, _ := aes.NewCipher(key)
	return cipher.NewCBCEncrypter(block, iv) // want `cipher.NewCBCEncrypter produces unauthenticated ciphertext`
}

func ctr(key, iv []byte) cipher.Stream {
	block, _ := aes.NewCipher(key)
	return cipher.NewCTR(block, iv) // want `cipher.NewCTR produces unauthenticated ciphertext`
}

func aliased(key, iv []byte) cipher.Stream {
	block, _ := aes.NewCipher(key)
	return stdcipher.NewCTR(block, iv) // want `cipher.NewCTR produces unauthenticated ciphertext`
}

func cbcDecrypt(key, iv []byte) cipher.BlockMode {
	block, _ := aes.NewCipher(key)
	return cipher.NewCBCDecrypter(block, iv)
}

func gcm(key []byte) cipher.AEAD {
	block, _ := aes.NewCipher(key)
	aead, _ := cipher.NewGCM(block)
	return aead
}

type local struct{}

func (local) NewCTR() {}

func method() {
	local{}.NewCTR()
}
