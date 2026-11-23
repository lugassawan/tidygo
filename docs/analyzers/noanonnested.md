# noanonnested

Reports anonymous struct types nested too deeply inside other anonymous structs. Deeply nested anonymous structs are hard to read, impossible to reuse and awkward to construct; extract a named type instead. Unlike `revive`'s `nested-structs`, this also covers `var` declarations, function parameters and results, and nesting through pointers, slices, maps and channels.

```go
// Bad
var cfg struct {
	DB struct { // error: anonymous struct nested 2 levels deep, max is 1; extract a named type
		Host string
	}
}

// Good
type DB struct{ Host string }

var cfg struct {
	DB DB
}
```

Depth is counted like this:

- A struct that is the right-hand side of a type declaration (named or alias) has depth 0, so `type T struct{ X struct{ ... } }` is depth 1 and allowed.
- Any other anonymous struct has depth 1, plus one for every anonymous struct that encloses it through fields, pointers (`*`), slices and arrays, maps and channels. A named slice of anonymous structs (`type Rows []struct{ A struct{ ... } }`) therefore counts its element struct as the first level, so the inner struct is depth 2 and reported.
- Anything else ends the walk: function signatures, interface methods, generic instantiations and composite literals each start counting again from 1.

Only the struct at depth `max + 1` is reported. That is the one to extract, and it gives one diagnostic per nesting chain: a struct three levels deep is reported once, at level two.

A single anonymous struct stays fine, which keeps table-driven tests (`[]struct{ name string }{...}`) untouched. A table row that itself holds an anonymous struct is depth 2 and is reported.

This analyzer is on by default, so it can add findings to an existing codebase. Raise `max` if that is too strict.

## Options

| Option | Default | Description |
|---|---|---|
| `max` | `1` | Deepest allowed nesting of anonymous structs. `0` or unset uses the default; a negative value is an error. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          noanonnested:
            max: 2
```
