# elseifelse

Requires an `if … else if` chain to end in a final `else`. A chain without one silently ignores any case the conditions did not anticipate; a final `else` states what happens then, even if the answer is to do nothing or to fail loudly. A plain `if` and an `if … else` are never reported.

**This analyzer is off by default**, because many codebases deliberately leave the last branch implicit. Turn it on with `elseifelse.enabled`.

```go
// Bad
if x < 0 {
	neg()
} else if x > 0 { // error: if-else-if chain has no final else
	pos()
}

// Good
if x < 0 {
	neg()
} else if x > 0 {
	pos()
} else {
	zero()
}

// OK: no else-if, so no chain
if x < 0 {
	neg()
}
```

A chain is reported once, at the `if` of its last link, however long it is. An `if` nested in an `else` block (`else { if … }`) is not part of the chain, and a chain nested in the body of another chain is checked on its own.

## Options

| Option | Default | Description |
|---|---|---|
| `enabled` | `false` | Turns the analyzer on. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          elseifelse:
            enabled: true
```
