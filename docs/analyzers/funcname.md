# funcname

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

## Options

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
