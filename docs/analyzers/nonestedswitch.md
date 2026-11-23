# nonestedswitch

Forbids a `switch` or type switch nested inside a case clause of another `switch` or type switch. Extract the inner switch into a function so each level stays readable.

```go
// Bad
switch a {
case 1:
	switch b { // error: switch statement nested inside another switch
	case 2:
	}
}

// Good
switch a {
case 1:
	handleB(b)
}

func handleB(b int) {
	switch b {
	case 2:
	}
}
```

The nearest enclosing switch-like statement in the same function decides:

- `if` and `for` blocks between the two switches do not hide the nesting; the inner switch is still reported.
- A function literal (or declaration) starts a new scope, so a switch inside a closure is not reported.
- A `select` between the two switches is itself the nearest switch-like statement, so a switch inside it is not reported.
- Sibling switches, one after the other, are fine. Three switches nested in a row produce two reports, one for each inner switch.
