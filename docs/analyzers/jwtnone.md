# jwtnone

Reports any reference to `SigningMethodNone` or `UnsafeAllowNoneSignatureType` from `github.com/golang-jwt/jwt` (v3), `github.com/golang-jwt/jwt/v4`, `github.com/golang-jwt/jwt/v5`, `github.com/dgrijalva/jwt-go` or the `github.com/form3tech-oss/jwt-go` fork. `SigningMethodNone` signs a token with no signature, and `UnsafeAllowNoneSignatureType` is the key a parser needs to accept one, so either lets an attacker forge claims.

The analyzer needs type information and matches the resolved object, so import aliases and dot imports are covered, while a local identifier with the same name is not.

```go
// Bad
token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

// Good
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, err := token.SignedString(secret)
parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))
```

## Known limitations

- Every reference is reported, including defensive ones such as `if tok.Method == jwt.SigningMethodNone { reject }` or a test asserting that a none token is rejected. Suppress those with `//nolint:tidygo // reason`, or exclude `_test.go` files with a `linters.exclusions` rule.
- Only references to the two identifiers are reported. A parser that accepts any algorithm because it has no `WithValidMethods` check is not detected, and neither is a runtime `alg` header check.
- Runtime or string lookups of the none algorithm are not detected: `jwt.GetSigningMethod("none")`, or an algorithm name read from configuration or a token header, returns the none method without naming either identifier.
- Other JWT libraries (`lestrrat-go/jwx`, other forks of `dgrijalva/jwt-go`) are not covered.
