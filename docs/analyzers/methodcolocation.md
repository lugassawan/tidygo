# methodcolocation

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

## Known limitations

- Only types declared at file scope are tracked. A method on a type alias (`type A = T`) is never reported, because receivers are matched by name and aliases are not resolved to the type they stand for.
- Build constraints are recognised only in the `//go:build` form, not the legacy `// +build` form.
- With `allow-file-suffix`, the stem comes from the file that declares the type, not from the type's name: `type User` declared in `account.go` accepts methods in `account_sql.go`, not `user_sql.go`.

## Options

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
