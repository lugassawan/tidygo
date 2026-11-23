# weakpwhash

Reports password-hashing calls with a constant salt or a constant work factor below a conservative floor. A constant salt lets one precomputed table attack every password at once, and a low cost makes brute force cheap. The floors are fixed and deliberately lenient (OWASP recommends 600,000 PBKDF2-HMAC-SHA256 iterations, for example), so they catch clearly weak settings rather than certify strong ones:

| Call | Salt checked | Reported when |
|---|---|---|
| `crypto/pbkdf2.Key`, `golang.org/x/crypto/pbkdf2.Key` | yes | iteration count < 100,000 |
| `golang.org/x/crypto/scrypt.Key` | yes | N < 32,768 (2^15) |
| `golang.org/x/crypto/argon2.Key`, `IDKey` | yes | memory < 19,456 KiB |
| `golang.org/x/crypto/bcrypt.GenerateFromPassword` | no | cost 4 to 9 |

The analyzer needs type information. A salt is constant when it is `nil`, a conversion of a constant (`[]byte("salt")`), or a `[]byte{...}` literal whose elements are all constants. bcrypt replaces a cost below `MinCost` (4) with `DefaultCost` (10), so costs 0 to 3 are safe and are not reported.

```go
// Bad
key := argon2.IDKey(pw, []byte("static"), 1, 4*1024, 4, 32)
hash, err := bcrypt.GenerateFromPassword(pw, 6)

// Good
salt := make([]byte, 16)
if _, err := rand.Read(salt); err != nil {
	return err
}
key := argon2.IDKey(pw, salt, 2, 19*1024, 1, 32)
hash, err := bcrypt.GenerateFromPassword(pw, bcrypt.DefaultCost)
```

## Known limitations

- Variables are not followed: `salt := []byte("static")` followed by `Key(pw, salt, ...)` is not reported, only a constant written at the call.
- A cost that is not a compile-time constant (a variable, a field or a configuration value) is skipped.
- Only the argon2 memory parameter is checked, not its time and thread counts; for scrypt only N is checked, not r and p. A deliberate memory-for-time trade (for example argon2 with 12 MiB and 3 passes) is therefore still reported; suppress it with `//nolint:tidygo // reason`.
- Test files are checked like any other file, so `bcrypt.GenerateFromPassword(pw, bcrypt.MinCost)` (the usual bcrypt test speed-up) and known-answer tests with a fixed salt are reported. Suppress them with `//nolint:tidygo // reason` or a `linters.exclusions` rule for `_test.go` files.
- Other password-hashing libraries (`alexedwards/argon2id`, `pbkdf2` wrappers) are not covered.
