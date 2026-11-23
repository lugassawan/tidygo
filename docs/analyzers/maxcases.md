# maxcases

Reports a `switch`, type switch or `select` statement with more than 30 case clauses. The `default` clause counts as a clause. A statement this large is usually better served by a lookup table or by splitting the dispatch.

```go
// Bad — 31 clauses
switch op { // error: switch has 31 case clauses, max is 30
case op1:
case op2:
	// ...
case op31:
}

// Good — a table replaces the switch
var handlers = map[Op]func(){op1: handle1, op2: handle2 /* ... */}
```

The message names the statement kind: `switch`, `type switch` or `select`.

## Options

| Option | Default | Description |
|---|---|---|
| `max` | `30` | Largest allowed number of clauses, `default` included. `0` or unset uses the default; a negative value is an error. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          maxcases:
            max: 20
```
