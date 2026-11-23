# confusingop

Reports an assignment where `=` is written directly against a unary `+`, `-`, `!` or `^`. Go parses `x =+ 1` as `x = +1`, so the program compiles and silently assigns instead of adding. The diagnostic is raised only when there is a space before `=` and nothing between `=` and the operator; `x = +1`, an operator on the next line, and consistently compact code such as `x=-1` are accepted as deliberate. Only the first right-hand value is checked, so `a, b =-1, 2` is reported too. The operator may also start a larger expression: `total =+ price * qty` is reported because `+price` is its leftmost operand.

```go
// Bad
x =+ 1               // meant x += 1
x =- 1               // meant x -= 1
x =^ 1               // meant x ^= 1
total =+ price * qty // meant total += price * qty
ok =!done            // missing space? write ok = !done

// Good
x += 1
x = -1   // a space shows the unary operator is intended
```

## Known limitations

`gofmt` rewrites `x =+ 1` to `x = +1`, which hides the typo. The analyzer only sees it in source that has not been formatted yet, so run the linter before the formatter (for example in an editor on save, or in CI before `gofmt`).

Only assignment statements are checked. Declarations such as `var y =-1` and `const c =+1` are not reported.

golangci-lint also reports only one issue per line by default (`issues.uniq-by-line`). If the `gofmt` formatter flags the same line first, the `confusingop` finding is hidden; set `uniq-by-line: false` to see both.
