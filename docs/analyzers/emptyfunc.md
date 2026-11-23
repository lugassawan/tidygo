# emptyfunc

Requires a function body with no statements to hold a comment that explains why it is empty. This covers function and method declarations as well as function literals. A comment must sit between the braces; one above the declaration or after the closing brace does not count. Declarations without a body (assembly or external functions) are skipped.

There are no exemptions and no settings: `func _() {}` and no-op implementations are reported like any other empty function. A one-line comment is the fix, or `//nolint:tidygo` where an empty body is intended and needs no explanation.

```go
// Bad
func (T) Flush() {}  // reported
handler := func() {} // reported

// Good
func (T) Flush() {
	// intentionally empty: T buffers nothing
}
handler := func() { /* no-op */ }
```
