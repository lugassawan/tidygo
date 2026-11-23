# redundantparen

Reports unnecessary parentheses around a whole value in a position where they can never change the meaning: an element of a `return`, the right-hand side of an assignment or declaration (`=`, `:=`, `op=`, `var x = …`), or a call argument. Doubled parentheses such as `((x))` are reported too, once, at the outer pair. Other parentheses inside a binary operand, such as `(a && b) || c` or `(a + b) * c`, are never reported, because they often clarify mixed operators or change grouping.

```go
// Bad
return (a + b)       // error: redundant parentheses
x := (a + b)         // error: redundant parentheses
x += (b)             // error: redundant parentheses
var y = (a)          // error: redundant parentheses
use((a))             // error: redundant parentheses

// Good
return a + b
x := a + b
total := (a + b) * c // grouping matters
ok := (p && q) || r  // clarifies mixed operators
```

The callee of a call is never reported, so conversions such as `(*T)(nil)` and `(func())(nil)` are kept. Parentheses that the parser needs are kept too: a composite literal in the header of an `if`, `for` or `switch` does not parse without them, so `if x := (T{}); x.ok {` is accepted.

## Known limitations

`gofmt` already collapses doubled parentheses, so the doubled-parentheses check only fires on source that has not been formatted yet; run the linter before the formatter (the same caveat as [`confusingop`](confusingop.md)). Parentheses around a type (`var t (int)`), a struct field value or an operand of a unary or binary operator are out of scope.

A composite literal anywhere inside a statement header keeps its parentheses, even in a call argument such as `if f((T{})) {` where they are in fact unnecessary. The analyzer errs on the side of not reporting.
