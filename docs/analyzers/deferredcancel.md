# deferredcancel

Requires the cancel function returned by `context.WithCancel`, `WithTimeout`, `WithDeadline` (and the `...Cause` variants) to be deferred on the very next statement. Deferring it straight away guarantees the context's resources are released on every return path, including panics and early returns added later.

The analyzer needs type information, so aliased imports and shadowed names are handled. It accepts `defer cancel()`, `defer cancel(nil)` for `WithCancelCause`, and a deferred function literal that calls `cancel()` as one of its top-level statements (`defer func() { cancel(); wg.Done() }()`). For a variable declared first and assigned in branches (`var cancel context.CancelFunc; if d > 0 { ctx, cancel = ... } else { ... }`), a single `defer cancel()` after the branches is enough.

```go
// Bad
ctx, cancel := context.WithTimeout(parent, time.Second)
result, err := fetch(ctx)
defer cancel() // not the next statement

// Good
ctx, cancel := context.WithTimeout(parent, time.Second)
defer cancel()
result, err := fetch(ctx)
```

A cancel func that escapes the function is skipped, because ownership has moved: when it is returned, passed as an argument, stored in a field or composite literal, assigned to another variable, captured by a `go` statement, or referenced inside any function literal that is not itself the deferred call (`time.AfterFunc(d, func() { cancel() })`, `t.Cleanup(func() { cancel() })`, `f := func() { cancel() }`). A blank `_` cancel is left to `go vet`'s `lostcancel`.

## Known limitations

- Loop bodies are skipped: a `defer` inside a `for` or `range` body piles up until the function returns, so the fix there is a per-iteration function, not a `defer`.
- Only assignments are matched, so `var ctx, cancel = context.WithCancel(...)` is not checked.
- A cancel variable that is not declared in the enclosing function (a package-level variable, a parameter, or one captured from an outer function) is skipped, since it is managed elsewhere.

## Difference from `lostcancel`

`lostcancel` reports a cancel function that is not *called on some path*. `deferredcancel` is stricter about style: it wants the `defer` immediately after creation, so a manual `cancel()` at the end of the function is reported here even though `lostcancel` accepts it on straight-line code.
