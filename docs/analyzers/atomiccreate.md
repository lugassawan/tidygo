# atomiccreate

Reports a file created after an existence check on the same path. `os.Stat` followed by `os.IsNotExist` and then `os.Create` is a check-then-act race: another process can create the file between the two calls, and `os.Create` (and `os.WriteFile`) truncates whatever is there. `os.OpenFile` with `O_CREATE|O_EXCL` does the check and the create in one atomic step.

The analyzer needs type information. It matches an `os.Stat` or `os.Lstat` call whose error is tested in an `if` condition by `os.IsNotExist(err)` or `errors.Is(err, fs.ErrNotExist)` / `errors.Is(err, os.ErrNotExist)`, in either the init form (`if _, err := os.Stat(p); os.IsNotExist(err) {`) or as a separate statement before the `if`. Inside that `if` body (not the `else`), it reports `os.Create(p)`, `os.WriteFile(p, ...)` and `os.OpenFile(p, flag, ...)` on the same path, where `flag` is a constant with `O_CREATE` set and `O_EXCL` unset.

```go
// Bad
if _, err := os.Stat(path); os.IsNotExist(err) {
	f, err := os.Create(path) // racy
	...
}

// Good
f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
if errors.Is(err, fs.ErrExist) {
	// someone else got there first
}
```

"Same path" means the same variable that is not reassigned between the stat and the create, or equal constant values.

## Known limitations

- A non-constant `os.OpenFile` flag is skipped, because the analyzer cannot tell whether `O_EXCL` is set.
- Negated checks (`!os.IsNotExist(err)`) and creates in the `else` branch of the check are not reported. An `else if os.IsNotExist(err)` is checked like any other `if`.
- `os.OpenFile` with `O_CREATE` but not `O_EXCL` is reported whatever else it sets (`O_APPEND`, `O_TRUNC`).
- Helpers are not followed: a stat in one function and the create in another, or a path built twice (`filepath.Join(dir, name)` in both calls), is not matched.
- Creates inside a nested function literal are not attributed to the enclosing check.
