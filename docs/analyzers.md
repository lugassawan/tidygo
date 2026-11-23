# Analyzer Details

## confusingop

Reports an assignment where `=` is written directly against a unary `+`, `-`, `!` or `^`. Go parses `x =+ 1` as `x = +1`, so the program compiles and silently assigns instead of adding. The diagnostic is raised only when nothing sits between `=` and the operator; `x = +1` and an operator on the next line are accepted as deliberate. Only the first right-hand value is checked, so `a, b =-1, 2` is reported too. The operator may also start a larger expression: `total =+ price * qty` is reported because `+price` is its leftmost operand.

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

### Known limitations

`gofmt` rewrites `x =+ 1` to `x = +1`, which hides the typo. The analyzer only sees it in source that has not been formatted yet, so run the linter before the formatter (for example in an editor on save, or in CI before `gofmt`).

golangci-lint also reports only one issue per line by default (`issues.uniq-by-line`). If the `gofmt` formatter flags the same line first, the `confusingop` finding is hidden; set `uniq-by-line: false` to see both.

## constcond

Forbids `if` and `for` statements whose condition is the literal `true` or `false`. Such a condition makes one branch dead, or turns a loop into an unconditional one (`for true`) or one that never runs (`for false`). Parentheses are ignored, and `else if` chains and `if` statements with an init clause are covered. `for {}` has no condition and is not reported.

```go
// Bad
if true { run() }
if false { run() }
if ok { a() } else if true { b() }
if x := f(); true { use(x) }
for false { run() }
for true { run() }

// Good
if cfg.Debug { run() }
for { run() }
```

### Known limitations

The plugin runs in syntax-only mode, so the check matches the identifier name only. A local variable that shadows `true` or `false` is still reported; enable the `predeclared` linter to catch such shadowing. Expressions that merely evaluate to a constant (`!true`, `true && x`, a named constant) are not detected.

## emptycomment

Forbids `/* */` block comments that contain only whitespace and `*` characters. Empty block comments are leftovers from deleted text or unfinished stubs and carry no information. `//` comments are never reported, because an empty `//` line separates the paragraphs of a doc comment.

```go
// Bad
/**/
/* */
/*
 *
 */

// Good
/* explanation */
//
```

## emptyfunc

Requires a function body with no statements to hold a comment that explains why it is empty. This covers function and method declarations as well as function literals. A comment must sit between the braces; one above the declaration or after the closing brace does not count. Declarations without a body (assembly or external functions) are skipped.

There are no exemptions and no settings: `func _() {}` and no-op implementations are reported like any other empty function. A one-line comment is the fix, or `//nolint:tidygo` where an empty body is intended and needs no explanation.

```go
// Bad
func (T) Flush() {}  // reported
handler := func() {} // reported

// Good
func (T) Flush() {
	// intentionally empty: T buffers nothing
}
handler := func() { /* no-op */ }
```

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

## ifacename

Requires a single-method interface to be named after its method plus an agent-noun suffix, the way the standard library does (`Read` → `Reader`, `Close` → `Closer`). The check is purely syntactic.

An interface is checked when its body is exactly one exported method. Interfaces that embed another interface, contain a type-set element (`~int | ~string`), declare more than one method, or have an unexported sole method (the sealed-interface pattern, such as `type Node interface{ isNode() }` or `type Option interface{ apply(*cfg) }`) are skipped. Local (function-scope) and generic interfaces are checked too.

Accepted spellings are compared ignoring the case of the first letter. For a method `w`:

| Rule | Example |
|---|---|
| `w` + `er` | `Read` → `Reader`, `String` → `Stringer`, `Marshal` → `Marshaler` |
| `w` ending in `e`: `w` + `r`, or drop the `e` and add `or` | `Close` → `Closer`, `Validate` → `Validator`, `Execute` → `Executor` |
| `w` + `or` | `Visit` → `Visitor`, `Edit` → `Editor` |
| consonant + `y`: replace `y` with `ier` | `Copy` → `Copier`, `Notify` → `Notifier` |
| consonant-vowel-consonant ending (not `w`, `x`, `y`): double the last letter | `Stop` → `Stopper`, `Run` → `Runner`, `Get` → `Getter` |

The same rules apply to the method's first camel-case word with the rest appended unchanged, which covers the standard library's `ReadFrom` → `ReaderFrom`, `WriteTo` → `WriterTo` and `ReadAt` → `ReaderAt`.

```go
// Bad
type FileRead interface{ Read() }      // reported, suggests Reader
type Handler interface{ ServeHTTP() }  // reported, suggests ServerHTTP (see allowed-names)
type closeable interface{ Close() }    // reported, suggests closer

// Good
type Reader interface{ Read() }
type Closer interface{ Close() error }
type ReaderFrom interface{ ReadFrom() }
type Getter[T any] interface{ Get() T }
```

The diagnostic's suggested name is a heuristic example, not the only accepted spelling. For multi-word methods it can read awkwardly (`RoundTrip` suggests `RounderTrip`), and stress-final verbs get an unusual spelling (`Visit` suggests `Visiter`, `Format` suggests `Formater`, `Commit` suggests `Commiter`). The suggestion is always itself an accepted name, and the standard spelling (`Visitor`, `Formatter`, `Committer`) is accepted too.

### Known limitations

The rule is deliberately strict, so a few common, idiomatic shapes are reported and need `allowed-names`:

- **Multi-word methods** are matched by suffixing the *first* word only (`GetUser` → `GetterUser`), so `UserGetter` is reported. The same applies to the "object + agent noun" names in the standard library: `ByteReader` (`ReadByte`), `RuneReader`, `StringWriter` (`WriteString`), `TextMarshaler`, and `Marshaler` for `MarshalJSON`.
- **Exported marker methods** are still reported, because only an *unexported* sole method is skipped: `type Expr interface{ Expr() }` suggests `Exprer` and `type EventMarker interface{ IsEvent() }` suggests `IserEvent`.
- **Type aliases** (`type X = interface{ Read() }`) are checked like any other declaration.
- **`Error() string`** interfaces, because `Error` → `Errorer`.

### Dependencies

Only interfaces declared in the packages being linted are inspected. Interfaces declared in dependencies or the standard library (`io.ReaderFrom`) are never reported, and neither is *using* them (embedding one, or taking one as a parameter), because that is a reference, not a declaration. The only exposure is your own code declaring a name that mirrors a standard library one. The forms that follow the rules above (`Reader`, `ReaderFrom`, `Stringer`) pass; the ones that do not (see Known limitations) need `allowed-names`.

### Options

| Option | Default | Description |
|---|---|---|
| `allowed-names` | `[]` | Interface names exempt from the check, for names that cannot be derived from the method (e.g. `Handler` for `ServeHTTP`, `ByteReader` for `ReadByte`). Matched exactly. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          ifacename:
            allowed-names:
              - Handler
              - ByteReader
```

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

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

- Evaluation order: in `if n := f(); g() > n`, inlining makes `g()` run before `f()`. This also applies when the initializer only reads mutable state, as in `if n := c.cnt; c.Next() > n`, where inlining reads `c.cnt` after `c.Next()` has run.
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
