# constcond

Forbids `if` and `for` statements whose condition is the literal `true` or `false`. Such a condition makes one branch dead, or turns a loop into an unconditional one (`for true`) or one that never runs (`for false`). Parentheses are ignored, and `else if` chains and `if` statements with an init clause are covered. `for {}` has no condition and is not reported.

```go
// Bad
if true { run() }
if false { run() }
if ok { a() } else if true { b() }
if x := f(); true { use(x) }
for false { run() }
for true { run() }

// Good
if cfg.Debug { run() }
for { run() }
```

## Known limitations

This analyzer uses no type information, so it matches the identifier name only. A local variable that shadows `true` or `false` is still reported; enable the `predeclared` linter to catch such shadowing. Expressions that merely evaluate to a constant (`!true`, `true && x`, a named constant) are not detected.
