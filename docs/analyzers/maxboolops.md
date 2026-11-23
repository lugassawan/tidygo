# maxboolops

Reports a boolean expression with more than 3 `&&` and `||` operators. Long chains are hard to read and easy to get wrong; give the sub-conditions names or move them into a predicate function.

Each `&&`/`||` operator counts once. Parentheses and unary `!` do not split an expression, so `(a && b) || !(c && d)` has 3 operators. Any other operand (a comparison, a call, a function literal) is a leaf: a chain inside a call argument or a closure body is its own expression and is counted separately. That keeps a struct literal with several small `a && b` fields from adding up into one finding. The diagnostic is reported once per expression, at its start.

```go
// Bad — 4 operators
ok := a && b && c && d && e // error: boolean expression has 4 && / || operators, max is 3

// Good
inRange := lo <= x && x <= hi
allowed := user.Active && (user.Admin || user.Owner)
ok := inRange && allowed
```

## Options

| Option | Default | Description |
|---|---|---|
| `max` | `3` | Largest allowed number of `&&` and `||` operators in one expression. `0` or unset uses the default; a negative value is an error. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          maxboolops:
            max: 4
```
