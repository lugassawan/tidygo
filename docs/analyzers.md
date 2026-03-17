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

Forbids exported standalone functions that appear after unexported standalone functions. Methods (functions with receivers) and `init()` are exempt.

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

type T struct{}
func (t T) Method() {} // exempt — has receiver
```

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
