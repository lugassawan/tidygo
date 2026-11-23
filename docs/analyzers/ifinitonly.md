# ifinitonly

Forbids an `if` init statement whose `:=` variable is used only in the condition. The extra name adds nothing, because the call can go straight into the condition.

```go
// Bad
if n := len(items); n > 10 {
	return errTooMany
}

// Good
if len(items) > 10 {
	return errTooMany
}

// Good: err is used in the body
if err := f(); err != nil {
	return err
}
```

A variable is reported only when it is referenced **exactly once**, in the condition, and nowhere in the body or `else` branches. With several references (`n > 0 && n < 10`), inlining would evaluate the call more than once, so the variable is kept.

Inlining is a suggestion, and the analyzer cannot prove it is always safe without type information. It skips the cases it can recognise:

- `&n` in the condition, because `&f()` is not valid Go.
- A reference inside a function literal in the condition (`func() bool { return n > 0 }()`), because the call would run on every invocation.
- A reference in the right operand of `&&` or `||` (`enabled && n > 0`), because inlining would make the call conditional.

It cannot see these, so check them before inlining:

- Evaluation order: in `if n := f(); g() > n`, inlining makes `g()` run before `f()`. This also applies when the initializer only reads mutable state, as in `if n := c.cnt; c.Next() > n`, where inlining reads `c.cnt` after `c.Next()` has run.
- Addressability: if `n` is the receiver of a pointer-receiver method or is sliced (`n.PtrMethod()`, `n[:2]`), inlining a non-addressable call result may not compile.

Only a single-variable `:=` init is checked. Multi-value inits such as `v, ok := m[k]`, plain assignment (`n = f()`) and other init statement kinds are ignored.

The analyzer runs without type information, so shadowing is resolved syntactically. A redeclaration of the same name in the body (`n := 5`, a closure parameter, a nested `if`/`for`/`switch`/`select` binding) is not counted as a use of the init variable. Where the two cannot be told apart it errs on the side of not reporting. For example, a struct-literal key such as `T{n: 1}` counts as a use, because it could be a map key that reads the variable.
