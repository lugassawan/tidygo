# Analyzer Details

## funcname

Ensures function and method names match the pattern `^(_|[a-zA-Z0-9]+)$`. Underscores are not permitted.

```go
// Bad
func get_user() {}
func TestSomething_Case() {}
func (t T) bad_method() {}

// Good
func getUser() {}
func TestSomethingCase() {}
func (t T) goodMethod() {}
func _() {} // blank identifier is allowed
```

## maxparams

Reports functions and function literals with more than 7 parameters. This includes both named and anonymous parameters.

```go
// Bad — 8 parameters
func process(a, b, c, d, e, f, g, h int) {}

// Bad — function literal
var fn = func(a, b, c, d, e, f, g, h int) {}

// Good — exactly 7
func process(a, b, c, d, e, f, g int) {}

// Recommended refactor: use an options struct
type ProcessOpts struct {
    A, B, C, D, E, F, G, H int
}
func process(opts ProcessOpts) {}
```

## nolateconst

Forbids package-level `const` and `var` declarations that appear after any function or method declaration. Type declarations are not affected.

```go
// Bad
const a = 1

func foo() {}

const b = 2  // error: const declaration after function declaration
var c = 3    // error: var declaration after function declaration

// Good
const a = 1
const b = 2
var c = 3

type T struct{}

func foo() {}
```

## nolateexport

Enforces export-first ordering within a file:

- **Standalone functions:** an exported function must not appear after an unexported standalone function. `init()` is exempt.
- **Methods:** an exported method must not appear after an unexported method of the *same receiver type*. Each receiver type is tracked independently (pointer and generic receivers count as the base type), so a layout of "type A's methods, then type B's methods" is fine. It does not matter which file declares the type. Receivers are keyed by the type name as written, so methods declared through a type alias (`type A = T`) are tracked separately from methods declared on `T`; mixed ordering across the two is not reported.

```go
// Bad
func Exported() {}
func unexported() {}
func AlsoExported() {} // error: should appear before unexported functions

// Good
func Exported() {}
func AlsoExported() {}
func unexported() {}
func init() {}         // exempt
```

Method ordering applies per receiver type, even when the type lives in another file:

```go
// a.go
type T struct{}
type U struct{}

// b.go
func (t T) helper() {}
func (t T) Shown() {}  // error: exported method T.Shown should appear before unexported methods of T

func (u U) Other() {}  // OK — no unexported method of U precedes it, even though T has one
```

When the receiver type is declared in the same file, `funcorder` reports similar problems. The two linters overlap in that case; nolateexport additionally covers types declared elsewhere.

### Options

| Option | Default | Description |
|---|---|---|
| `unexported-funcs-last` | `false` | Also require exported methods to appear before unexported *standalone* functions. |

Off by default, so teams that keep unexported standalone helpers next to their callers are unaffected by this option. (The per-type method rule above is always on and still applies to unexported helper *methods*.) With it enabled:

```go
// Bad
func stableGitEnv() []string { return nil }

func (r *Runner) Run() {} // error: exported method Runner.Run should appear before unexported standalone functions
```

Enable it under the tidygo plugin's `settings`:

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          nolateexport:
            unexported-funcs-last: true
```

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

## nolocalstruct

Forbids named struct type declarations inside function bodies. Structs should be declared at package level. Non-struct local types (e.g. `type myInt int`) are allowed.

```go
// Bad
func process() {
    type Request struct{ URL string } // error: declare at package level
    _ = Request{}
}

// Good
type Request struct{ URL string }

func process() {
    _ = Request{}
}

// OK — non-struct local types are allowed
func convert() {
    type myInt int
    _ = myInt(0)
}
```
