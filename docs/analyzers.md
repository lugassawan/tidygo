# Analyzer Details

## atomiccreate

Reports a file created after an existence check on the same path. `os.Stat` followed by `os.IsNotExist` and then `os.Create` is a check-then-act race: another process can create the file between the two calls, and `os.Create` (and `os.WriteFile`) truncates whatever is there. `os.OpenFile` with `O_CREATE|O_EXCL` does the check and the create in one atomic step.

The analyzer needs type information. It matches an `os.Stat` or `os.Lstat` call whose error is tested in an `if` condition by `os.IsNotExist(err)` or `errors.Is(err, fs.ErrNotExist)` / `errors.Is(err, os.ErrNotExist)`, in either the init form (`if _, err := os.Stat(p); os.IsNotExist(err) {`) or as a separate statement before the `if`. Inside that `if` body (not the `else`), it reports `os.Create(p)`, `os.WriteFile(p, ...)` and `os.OpenFile(p, flag, ...)` on the same path, where `flag` is a constant with `O_CREATE` set and `O_EXCL` unset.

```go
// Bad
if _, err := os.Stat(path); os.IsNotExist(err) {
	f, err := os.Create(path) // racy
	...
}

// Good
f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
if errors.Is(err, fs.ErrExist) {
	// someone else got there first
}
```

"Same path" means the same variable that is not reassigned between the stat and the create, or equal constant values.

### Known limitations

- A non-constant `os.OpenFile` flag is skipped, because the analyzer cannot tell whether `O_EXCL` is set.
- Negated checks (`!os.IsNotExist(err)`) and creates in the `else` branch of the check are not reported. An `else if os.IsNotExist(err)` is checked like any other `if`.
- `os.OpenFile` with `O_CREATE` but not `O_EXCL` is reported whatever else it sets (`O_APPEND`, `O_TRUNC`).
- Helpers are not followed: a stat in one function and the create in another, or a path built twice (`filepath.Join(dir, name)` in both calls), is not matched.
- Creates inside a nested function literal are not attributed to the enclosing check.

## busywait

Forbids a `for` loop that spins on a plain variable: the body is empty, or is only `time.Sleep(...)` or `runtime.Gosched()`. Reading a variable that another goroutine writes, without synchronisation, is a data race, and the loop burns CPU (or polls on a timer) while it waits. Use a channel, `sync.Cond`, `sync.WaitGroup` or `context` instead.

The analyzer needs type information to recognise `time.Sleep` and `runtime.Gosched`. The loop must have a condition and no init or post statement. The condition must be built only from identifiers, selectors, parentheses, unary `!`/`-`/`+`, comparisons, `&&`/`||` and constants. Arithmetic (`x+1 < y`) and pointer dereferences (`*p`) are outside that set, so such loops are not reported. A condition that is a constant (`for true`) waits on no state and is not reported. Any function or method call, channel receive or index expression in the condition makes the loop exempt (type conversions and builtins do not count as calls, so `time.Sleep(time.Duration(n) * time.Millisecond)` is still a spin), which keeps polling idioms such as `atomic.Load*`, `ctx.Err()`, `time.Now().Before(d)` and `len(q)` out of scope. A bare `for {}` has no condition and is covered by staticcheck SA5002.

```go
// Bad
for !done {
}
for !s.ready {
	time.Sleep(10 * time.Millisecond)
}
for flag == false {
	runtime.Gosched()
}

// Good
<-doneCh
for atomic.LoadInt32(&state) == 0 { // an explicit atomic poll is not reported
	time.Sleep(time.Millisecond)
}
for i := 0; i < n; i++ { // has init and post
}
```

### Known limitations

A loop that does any other work in its body is never reported, even if it also spins on a plain variable. An atomic poll is not reported either, though a channel is usually still the better design.

## confusingop

Reports an assignment where `=` is written directly against a unary `+`, `-`, `!` or `^`. Go parses `x =+ 1` as `x = +1`, so the program compiles and silently assigns instead of adding. The diagnostic is raised only when there is a space before `=` and nothing between `=` and the operator; `x = +1`, an operator on the next line, and consistently compact code such as `x=-1` are accepted as deliberate. Only the first right-hand value is checked, so `a, b =-1, 2` is reported too. The operator may also start a larger expression: `total =+ price * qty` is reported because `+price` is its leftmost operand.

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

Only assignment statements are checked. Declarations such as `var y =-1` and `const c =+1` are not reported.

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

This analyzer uses no type information, so it matches the identifier name only. A local variable that shadows `true` or `false` is still reported; enable the `predeclared` linter to catch such shadowing. Expressions that merely evaluate to a constant (`!true`, `true && x`, a named constant) are not detected.

## deferredcancel

Requires the cancel function returned by `context.WithCancel`, `WithTimeout`, `WithDeadline` (and the `...Cause` variants) to be deferred on the very next statement. Deferring it straight away guarantees the context's resources are released on every return path, including panics and early returns added later.

The analyzer needs type information, so aliased imports and shadowed names are handled. It accepts `defer cancel()`, `defer cancel(nil)` for `WithCancelCause`, and a deferred function literal that calls `cancel()` as one of its top-level statements (`defer func() { cancel(); wg.Done() }()`). For a variable declared first and assigned in branches (`var cancel context.CancelFunc; if d > 0 { ctx, cancel = ... } else { ... }`), a single `defer cancel()` after the branches is enough.

```go
// Bad
ctx, cancel := context.WithTimeout(parent, time.Second)
result, err := fetch(ctx)
defer cancel() // not the next statement

// Good
ctx, cancel := context.WithTimeout(parent, time.Second)
defer cancel()
result, err := fetch(ctx)
```

A cancel func that escapes the function is skipped, because ownership has moved: when it is returned, passed as an argument, stored in a field or composite literal, assigned to another variable, captured by a `go` statement, or referenced inside any function literal that is not itself the deferred call (`time.AfterFunc(d, func() { cancel() })`, `t.Cleanup(func() { cancel() })`, `f := func() { cancel() }`). A blank `_` cancel is left to `go vet`'s `lostcancel`.

### Known limitations

- Loop bodies are skipped: a `defer` inside a `for` or `range` body piles up until the function returns, so the fix there is a per-iteration function, not a `defer`.
- Only assignments are matched, so `var ctx, cancel = context.WithCancel(...)` is not checked.
- A cancel variable that is not declared in the enclosing function (a package-level variable, a parameter, or one captured from an outer function) is skipped, since it is managed elsewhere.

### Difference from `lostcancel`

`lostcancel` reports a cancel function that is not *called on some path*. `deferredcancel` is stricter about style: it wants the `defer` immediately after creation, so a manual `cancel()` at the end of the function is reported here even though `lostcancel` accepts it on straight-line code.

## dupfunc

Reports a function or method whose body is identical to that of an earlier function in the same package. Copy-pasted bodies drift apart when only one copy is fixed; extract the shared logic and call it from both. Only bodies with at least 3 top-level statements are compared by default (see `min-statements`), so trivial getters and one-line wrappers are left alone.

Bodies are compared structurally. Positions, comments and layout are ignored, but identifier names, literals and operators must match exactly. A renamed local variable, a different constant or a different operator makes two bodies different. Only the body is compared: parameter lists and receivers are not, so two methods on different types with the same body are reported.

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

Files are ordered by name and offset, and every later duplicate is reported against the **first** function, so the result does not depend on the order the files were loaded in. A group of three copies yields two reports, both naming the first. The diagnostic is reported at the later function, so a `//nolint:tidygo` directive goes there.

Methods are named with their receiver type, for example `function right.count has the same body as left.count (types.go:5)`.

### Known limitations

`_test.go` files and generated files (those with a `// Code generated ... DO NOT EDIT.` header) are skipped entirely, since table-test scaffolding repeats on purpose. Function literals are never compared, and neither are `init` functions: a package may declare several, and one cannot call another to share a body. Duplicates are only found within one package, and only whole bodies are compared, not repeated fragments inside different functions; the `dupl` linter covers larger token-level clones.

The analyzer runs without type information, so two identical bodies are reported even if their parameters or receivers have different types (for example `clampInt` and `clampFloat`). Raising `min-statements` reduces such noise, and `//nolint:tidygo` on the later function silences a deliberate copy.

### Options

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

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

## elseifelse

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

### Options

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

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

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

### Options

| Option | Default | Description |
|---|---|---|
| `include-vars` | `false` | Also apply the same pattern to variables, constants, parameters, results and receivers. |

Off by default, so existing configurations see no new findings. With it enabled, these names are checked:

- `var` and `const` names, at package level and inside functions
- names introduced by `:=`, including type switch bindings (`switch v := x.(type)`) and `select` receive clauses
- `range` keys and values declared with `:=`
- receivers, parameters and named results of functions, methods and function literals

Struct fields, type parameters, and the parameters of interface methods and function *types* are not checked, since they declare no variable. Assignments (`=`) and `range ... = ` clauses only reuse an existing variable, so they are not checked either. Without type information, a `:=` that redeclares an existing variable is checked again, so such a name can be reported twice, and a `//nolint` on the original declaration does not cover the later `:=`.

The pattern is ASCII-only (`[a-zA-Z0-9]`), as it already is for function names, so a valid Go identifier such as `σ` or `参数` is reported. Leave `include-vars` off if your code uses non-ASCII identifiers.

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

## maxboolops

Reports a boolean expression with more than 3 `&&` and `||` operators. Long chains are hard to read and easy to get wrong; give the sub-conditions names or move them into a predicate function.

Each `&&`/`||` operator counts once. Parentheses and unary `!` do not split an expression, so `(a && b) || !(c && d)` has 3 operators. Any other operand (a comparison, a call, a function literal) is a leaf: a chain inside a call argument or a closure body is its own expression and is counted separately. That keeps a struct literal with several small `a && b` fields from adding up into one finding. The diagnostic is reported once per expression, at its start.

```go
// Bad — 4 operators
ok := a && b && c && d && e // error: boolean expression has 4 && / || operators, max is 3

// Good
inRange := lo <= x && x <= hi
allowed := user.Active && (user.Admin || user.Owner)
ok := inRange && allowed
```

### Options

| Option | Default | Description |
|---|---|---|
| `max` | `3` | Largest allowed number of `&&` and `||` operators in one expression. `0` or unset uses the default; a negative value is an error. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          maxboolops:
            max: 4
```

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

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

## methodcolocation

Requires a method to be declared in the same file as its receiver type, so that reading `user.go` shows the type and everything it can do. The check runs per package and reports the method name, pointing at the file that declares the type.

The following are exempt:

- methods in `_test.go` files;
- generated files, on either the method or the type side;
- files with a `//go:build` constraint, and files with a `_<GOOS>`, `_<GOARCH>` or `_<GOOS>_<GOARCH>` name suffix (`user_linux.go`), because platform-specific methods have to live in their own files;
- receiver types that are not declared in the package.

```go
// user.go
type User struct{ ID int }

// helpers.go
func (u User) Name() string { return "" } // reported: should be in user.go
```

### Known limitations

- Only types declared at file scope are tracked. A method on a type alias (`type A = T`) is never reported, because receivers are matched by name and aliases are not resolved to the type they stand for.
- Build constraints are recognised only in the `//go:build` form, not the legacy `// +build` form.
- With `allow-file-suffix`, the stem comes from the file that declares the type, not from the type's name: `type User` declared in `account.go` accepts methods in `account_sql.go`, not `user_sql.go`.

### Options

| Option | Default | Description |
|---|---|---|
| `allow-file-suffix` | `false` | Allow a method in `<stem>_*.go` when the type is declared in `<stem>.go` (`user.go` with `user_sql.go`, `user_json.go`). Only that exact `<stem>_` prefix qualifies. |

```yaml
linters:
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          methodcolocation:
            allow-file-suffix: true
```

Unknown keys are rejected, so a typo in the option name fails loudly instead of being silently ignored.

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
- Any other anonymous struct has depth 1, plus one for every anonymous struct that encloses it through fields, pointers (`*`), slices and arrays, maps and channels. A named slice of anonymous structs (`type Rows []struct{ A struct{ ... } }`) therefore counts its element struct as the first level, so the inner struct is depth 2 and reported.
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

## redundantparen

Reports unnecessary parentheses around a whole value in a position where they can never change the meaning: an element of a `return`, the right-hand side of an assignment or declaration (`=`, `:=`, `op=`, `var x = …`), or a call argument. Doubled parentheses such as `((x))` are reported too, once, at the outer pair. Other parentheses inside a binary operand, such as `(a && b) || c` or `(a + b) * c`, are never reported, because they often clarify mixed operators or change grouping.

```go
// Bad
return (a + b)       // error: redundant parentheses
x := (a + b)         // error: redundant parentheses
x += (b)             // error: redundant parentheses
var y = (a)          // error: redundant parentheses
use((a))             // error: redundant parentheses

// Good
return a + b
x := a + b
total := (a + b) * c // grouping matters
ok := (p && q) || r  // clarifies mixed operators
```

The callee of a call is never reported, so conversions such as `(*T)(nil)` and `(func())(nil)` are kept. Parentheses that the parser needs are kept too: a composite literal in the header of an `if`, `for` or `switch` does not parse without them, so `if x := (T{}); x.ok {` is accepted.

This analyzer has no options.

### Known limitations

`gofmt` already collapses doubled parentheses, so the doubled-parentheses check only fires on source that has not been formatted yet; run the linter before the formatter (the same caveat as `confusingop`). Parentheses around a type (`var t (int)`), a struct field value or an operand of a unary or binary operator are out of scope.

A composite literal anywhere inside a statement header keeps its parentheses, even in a call argument such as `if f((T{})) {` where they are in fact unnecessary. The analyzer errs on the side of not reporting.

## txrollback

Requires a `*sql.Tx` to have a `defer tx.Rollback()` in the function that begins it. A transaction that returns early without committing or rolling back keeps its connection and any locks until the pool times out. `Rollback` after a successful `Commit` is a harmless no-op that returns `sql.ErrTxDone`, so deferring it right after `Begin` is always safe.

The analyzer needs type information. It tracks any assignment or `var` whose left-hand variable is a `*database/sql.Tx` and whose right-hand side calls a function or method named `Begin…` or `Start…` (case-insensitive), which covers `DB.Begin`, `DB.BeginTx`, `Conn.BeginTx` and wrappers such as `beginTx(db)` or `s.startTransaction(ctx)`. The deferral may be anywhere in the same function, either `defer tx.Rollback()` or a deferred function literal that calls `tx.Rollback()` (including `_ = tx.Rollback()` or inside an `if`). A manual `tx.Rollback()` in an error branch does not count, because it misses the other return paths and panics. A `Begin` inside a `for` or `range` body is checked differently, because a deferred `Rollback` there would pile up until the function returns. It is accepted when the loop body calls `tx.Rollback()` on that variable anywhere (deferred or not), and otherwise reported with the advice to extract the iteration into a function that has its own `defer tx.Rollback()`.

```go
// Bad
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return err
}
if _, err := tx.ExecContext(ctx, q); err != nil {
	return err // leaks the transaction
}
return tx.Commit()

// Good
tx, err := db.BeginTx(ctx, nil)
if err != nil {
	return err
}
defer tx.Rollback()
```

A transaction that escapes the function is skipped, because the caller owns it: when it is returned, passed as an argument (including `defer cleanup(tx)`), stored in a field or composite literal, assigned to another variable, or used in a `go` statement. It is also treated as managed when its `Rollback` is handed to code that runs later: called inside any function literal (`t.Cleanup(func() { _ = tx.Rollback() })`), or passed on as a method value (`return tx.Commit, tx.Rollback, nil`).

### Known limitations

- Only `*database/sql.Tx` is tracked; other transaction types (`pgx.Tx`, `sqlx.Tx`) are not.
- The check does not prove a `Rollback` runs on *every* path, only that one is deferred in the function.
- Passing the transaction to any function counts as handing over ownership, so the common `tx, _ := db.Begin(); repo.insert(tx, row); tx.Commit()` is not reported. Methods called on `tx` itself (`tx.Exec`, `tx.Commit`) are not hand-overs.
- A call starts a transaction only when its name begins with `Begin` or `Start`. A getter such as `tx := txFromContext(ctx)` and a type conversion are not reported, but a wrapper with another name (`openTx`, `newUnitOfWork`) is not tracked either, and neither is a call through a function value.
- A variable assigned from `Begin` twice in one function is reported at each assignment.
- In a loop, any `Rollback` call on the variable in the loop body is accepted, even one that covers only some of the exits, so an iteration that rolls back on one error path and returns early on another is not reported.
