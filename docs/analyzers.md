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

### Options

| Option | Default | Description |
|---|---|---|
| `include-vars` | `false` | Also apply the same pattern to variables, constants, parameters, results and receivers. |

Off by default, so existing configurations see no new findings. With it enabled, these names are checked:

- `var` and `const` names, at package level and inside functions
- names introduced by `:=`, including type switch bindings (`switch v := x.(type)`) and `select` receive clauses
- `range` keys and values declared with `:=`
- receivers, parameters and named results of functions, methods and function literals

Struct fields, type parameters, and the parameters of interface methods and function *types* are not checked, since they declare no variable. Assignments (`=`) and `range ... = ` clauses only reuse an existing variable, so they are not checked either. Without type information, a `:=` that redeclares an existing variable is checked again, so such a name can be reported twice.

```go
// Bad
var max_retries = 3                // error: Rename variable "max_retries" ...
func get(user_id int) {}           // error: Rename parameter "user_id" ...
for key_name := range m {}         // error: Rename variable "key_name" ...

// Good
var maxRetries = 3
func get(userID int) {}
for keyName := range m {}
```

The message names the kind: `variable`, `constant`, `parameter`, `result` or `receiver`. The message for functions is unchanged.

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          funcname:
            include-vars: true
```

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

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

## maxcaselines

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

### Options

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

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

## maxcases

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

### Options

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

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

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

## noanonnested

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
- Any other anonymous struct has depth 1, plus one for every anonymous struct that encloses it through fields, pointers (`*`), slices and arrays, maps and channels.
- Anything else ends the walk: function signatures, interface methods, generic instantiations and composite literals each start counting again from 1.

Only the struct at depth `max + 1` is reported. That is the one to extract, and it gives one diagnostic per nesting chain: a struct three levels deep is reported once, at level two.

A single anonymous struct stays fine, which keeps table-driven tests (`[]struct{ name string }{...}`) untouched. A table row that itself holds an anonymous struct is depth 2 and is reported.

This analyzer is on by default, so it can add findings to an existing codebase. Raise `max` if that is too strict.

### Options

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

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

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

Forbids named struct type declarations inside function bodies. Structs should be declared at package level. Non-struct local types (e.g. `type myInt int`) are allowed unless `all-types` is enabled.

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

### Options

| Option | Default | Description |
|---|---|---|
| `all-types` | `false` | Also report local interfaces, defined types (`type id int`, `type fn func()`) and generic types. |

Off by default, so existing configurations see no new findings. Aliases (`type X = int`) are not new named types and stay allowed, except an alias of a struct (`type X = struct{}`), which is reported as it always was. With it enabled:

```go
// Bad
func convert() {
    type myInt int          // error: named type "myInt" should be declared at package level
    type Reader interface{} // error: named type "Reader" should be declared at package level
}
```

Struct declarations keep the original message (`named struct type %q ...`) in both modes, so existing `nolint` directives and exclusion rules keep matching.

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          nolocalstruct:
            all-types: true
```

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

## nonestedswitch

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

This analyzer has no options.
