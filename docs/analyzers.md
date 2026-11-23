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

## ifinitonly

Forbids an `if` init statement whose `:=` variable is used only in the condition. The extra name adds nothing, because the call can go straight into the condition.

```go
// Bad
if n := len(items); n > 10 {
	return errTooMany
}

// Good
if len(items) > 10 {
	return errTooMany
}

// Good: err is used in the body
if err := f(); err != nil {
	return err
}
```

A variable is reported only when it is referenced **exactly once**, in the condition, and nowhere in the body or `else` branches. With several references (`n > 0 && n < 10`), inlining would evaluate the call more than once, so the variable is kept.

Inlining is a suggestion, and the analyzer cannot prove it is always safe without type information. It skips the cases it can recognise:

- `&n` in the condition, because `&f()` is not valid Go.
- A reference inside a function literal in the condition (`func() bool { return n > 0 }()`), because the call would run on every invocation.
- A reference in the right operand of `&&` or `||` (`enabled && n > 0`), because inlining would make the call conditional.

It cannot see these, so check them before inlining:

- Evaluation order: in `if n := f(); g() > n`, inlining makes `g()` run before `f()`.
- Addressability: if `n` is the receiver of a pointer-receiver method or is sliced (`n.PtrMethod()`, `n[:2]`), inlining a non-addressable call result may not compile.

Only a single-variable `:=` init is checked. Multi-value inits such as `v, ok := m[k]`, plain assignment (`n = f()`) and other init statement kinds are ignored.

The analyzer runs without type information, so shadowing is resolved syntactically. A redeclaration of the same name in the body (`n := 5`, a closure parameter, a nested `if`/`for`/`switch`/`select` binding) is not counted as a use of the init variable. Where the two cannot be told apart it errs on the side of not reporting. For example, a struct-literal key such as `T{n: 1}` counts as a use, because it could be a map key that reads the variable.

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
