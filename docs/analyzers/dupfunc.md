# dupfunc

Reports a function or method whose body and signature are identical to those of an earlier function in the same package. Copy-pasted bodies drift apart when only one copy is fixed; extract the shared logic and call it from both. Only bodies with at least 3 top-level statements are compared by default (see `min-statements`), so trivial getters and one-line wrappers are left alone.

Bodies are compared structurally. Positions, comments and layout are ignored, but identifier names, literals and operators must match exactly. A renamed local variable, a different constant or a different operator makes two bodies different.

Signatures must match too, so two functions with the same body but different parameter or result types are left alone: they cannot be merged without generics. Types are compared as the type checker resolves them, so `byte` matches `uint8`, an alias matches its target, and generic functions match up to the names of their type parameters. Receivers and parameter names are not compared, so two methods on different types with the same signature are still reported, since embedding or a shared helper can often merge them. Methods of different generic types are reported only when their signatures do not mention the receiver's type parameters: `(*stack[T]).push(x T)` and `(*queue[K, V]).push(x K)` use unrelated type parameters and are left alone, while two `reset(n int)` methods are reported.

```go
// Bad
func totalLen(items []string) int {
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}

// error: function sumAll has the same body as totalLen (lengths.go:3)
func sumAll(items []string) int {
	sum := 0
	for _, s := range items {
		sum += len(s)
	}
	return sum
}

// Good: one function, called from both places
```

```go
// Good: the same body over different types is not reported
func clampInt(x, lo, hi int) int {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}

func clampFloat(x, lo, hi float64) float64 {
	if x < lo {
		return lo
	}
	if x > hi {
		return hi
	}
	return x
}
```

Files are ordered by name and offset, and every later duplicate is reported against the **first** earlier function with an identical signature, so the result does not depend on the order the files were loaded in. A group of three copies with the same signature yields two reports, both naming the first; copies whose signatures differ are grouped separately, each group naming its own first function. The diagnostic is reported at the later function, so a `//nolint:tidygo` directive goes there.

Methods are named with their receiver type, for example `function right.count has the same body as left.count (types.go:5)`.

## Known limitations

`_test.go` files and generated files (those with a `// Code generated ... DO NOT EDIT.` header) are skipped entirely, since table-test scaffolding repeats on purpose. Function literals are never compared, and neither are `init` functions: a package may declare several, and one cannot call another to share a body. Duplicates are only found within one package, and only whole bodies are compared, not repeated fragments inside different functions; the `dupl` linter covers larger token-level clones.

Receiver types are not compared, so methods on different types that share a body and signature are reported even when merging them would take a refactor. The type checker does not list a method named `_` among its type's methods, so one on a generic type is not canonicalised: it is matched only when its signature does not mention the receiver's type parameters. Raising `min-statements` reduces such noise, and `//nolint:tidygo` on the later function silences a deliberate copy.

## Options

| Option | Default | Description |
|---|---|---|
| `min-statements` | `3` | Smallest number of top-level statements a body needs to be compared. `0` or unset uses the default; a negative value is an error. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          dupfunc:
            min-statements: 5
```
