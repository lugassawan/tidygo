# maxcaselines

Reports a `case` or `default` clause (in a `switch`, type switch or `select`) whose body spans more than 10 lines. Long case bodies hide the shape of the switch; extract the body into a function named for what the case does.

The span runs from the **first statement's start line** through the **last statement's end line**, inclusive. Blank and comment lines *between* statements count, and a single statement that wraps over several lines counts in full. Comments before the first statement or after the last one do not. A clause with an empty body is never reported. The diagnostic is reported at the clause.

```go
// Bad — the body spans 11 lines
switch kind {
case kindA: // error: case clause body spans 11 lines, max is 10
	a1()
	// ...
	a11()
}

// Good
switch kind {
case kindA:
	handleA()
case kindB:
	handleB()
}
```

## Options

| Option | Default | Description |
|---|---|---|
| `max` | `10` | Largest allowed body span, in lines. `0` or unset uses the default; a negative value is an error. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          maxcaselines:
            max: 15
```
