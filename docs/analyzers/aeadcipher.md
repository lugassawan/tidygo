# aeadcipher

Reports `cipher.NewCBCEncrypter` and `cipher.NewCTR`. CBC and CTR hide the plaintext but do not authenticate the ciphertext, so an attacker who can modify it can flip plaintext bits (CTR) or mount padding-oracle attacks (CBC) without detection. An AEAD such as AES-GCM or ChaCha20-Poly1305 encrypts and authenticates in one step.

The analyzer needs type information, so import aliases and dot imports are matched. `cipher.NewCBCDecrypter` is not reported, because decrypting legacy data needs it. CFB, OFB and RSA PKCS#1 v1.5 are left to staticcheck `SA1019`, which reports them as deprecated.

```go
// Bad
stream := cipher.NewCTR(block, iv)

// Good
aead, err := cipher.NewGCM(block)
```

## Known limitations

- A CTR or CBC construction that is paired with a separate MAC (encrypt-then-MAC) is still reported; suppress it with `//nolint:tidygo // reason`.
- Only the `crypto/cipher` constructors are matched. Third-party wrappers around them are not followed.
