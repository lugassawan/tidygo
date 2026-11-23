<div align="center">

# tidygo

[![CI](https://github.com/lugassawan/tidygo/actions/workflows/test.yml/badge.svg)](https://github.com/lugassawan/tidygo/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lugassawan/tidygo.svg)](https://pkg.go.dev/github.com/lugassawan/tidygo)
[![Go Report Card](https://goreportcard.com/badge/github.com/lugassawan/tidygo)](https://goreportcard.com/report/github.com/lugassawan/tidygo)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/lugassawan/tidygo)](https://github.com/lugassawan/tidygo/releases)

A [golangci-lint](https://golangci-lint.run/) v2 plugin that enforces code ordering, structural conventions, a few resource-handling bugs and a few crypto misuses in Go source files.

</div>

## Analyzers

| Analyzer | Description |
|---|---|
| `aeadcipher` | Reports unauthenticated cipher modes for encryption (`cipher.NewCBCEncrypter`, `cipher.NewCTR`); use an AEAD such as GCM |
| `atomiccreate` | Reports `os.Create`, `os.WriteFile` and non-exclusive `os.OpenFile` calls guarded by an `os.Stat` not-exist check (use `O_CREATE\|O_EXCL`) |
| `busywait` | Forbids empty or sleep-only `for` loops that spin on a plain variable |
| `confusingop` | Reports `=+`, `=-`, `=!` and `=^` assignments that look like typos for `+=`, `-=`, `!=` and `^=` |
| `constcond` | Forbids `if` and `for` statements whose condition is the literal `true` or `false` |
| `deferredcancel` | Requires the cancel func from `context.WithCancel`/`WithTimeout`/`WithDeadline` to be deferred immediately |
| `emptycomment` | Forbids `/* */` block comments that contain only whitespace and `*` |
| `emptyfunc` | Requires a comment inside function bodies that contain no statements |
| `funcname` | Forbids underscores in function and method names (optionally variables, constants, parameters, results and receivers) |
| `ifacename` | Requires single-method interfaces to be named after their method plus an agent-noun suffix (`Read` → `Reader`) |
| `ifinitonly` | Forbids `if` init variables used only in the condition |
| `jwtnone` | Reports use of the JWT `none` algorithm (`SigningMethodNone`, `UnsafeAllowNoneSignatureType`) in golang-jwt and dgrijalva/jwt-go |
| `maxcaselines` | Forbids `case` clause bodies that span more than 10 lines (configurable) |
| `maxcases` | Forbids `switch`, type switch and `select` statements with more than 30 case clauses (configurable) |
| `maxparams` | Forbids functions with more than 7 parameters |
| `methodcolocation` | Requires a method to be declared in the same file as its receiver type (optionally allowing `<file>_*.go` siblings of the type's file) |
| `noanonnested` | Forbids more than 1 level of anonymous struct nesting; a named struct type does not count (configurable) |
| `nolateconst` | Forbids `const`/`var` declarations after function declarations |
| `nolateexport` | Forbids exported functions after unexported functions, and exported methods after unexported methods of the same type |
| `nolocalstruct` | Forbids named struct type declarations inside function bodies (optionally interfaces and defined types too) |
| `nonestedswitch` | Forbids a `switch` or type switch nested inside another `switch` or type switch |
| `txrollback` | Requires a `*sql.Tx` to have a `defer tx.Rollback()` in the function that begins it |
| `weakpwhash` | Reports password hashing (pbkdf2, scrypt, argon2, bcrypt) with a constant salt or a constant cost below the recommended minimum |

## Installation

Requirements: Go 1.26+, golangci-lint v2.

Create `.custom-gcl.yml`:

```yaml
version: v2.15.0
plugins:
  - module: 'github.com/lugassawan/tidygo'
    version: v0.0.1
```

Configure `.golangci.yml`:

```yaml
version: "2"

linters:
  enable:
    - tidygo
  settings:
    custom:
      tidygo:
        type: "module"
        settings:
          # Optional: apply the underscore rule to variables, constants, parameters,
          # results and receivers as well as functions.
          funcname:
            include-vars: true
          # Optional: exempt interface names that cannot be derived from their method.
          ifacename:
            allowed-names: [Handler]
          # Optional: raise the longest allowed case body (default 10 lines).
          maxcaselines:
            max: 15
          # Optional: raise the most allowed case clauses per switch (default 30).
          maxcases:
            max: 40
          # Optional: allow methods in <type>_*.go files (user.go → user_sql.go).
          methodcolocation:
            allow-file-suffix: true
          # Optional: raise the deepest allowed nesting of anonymous structs (default 1).
          noanonnested:
            max: 2
          # Optional: also require exported methods before unexported standalone functions.
          nolateexport:
            unexported-funcs-last: true
          # Optional: also forbid local interfaces and defined types, not just structs.
          nolocalstruct:
            all-types: true
```

tidygo loads packages with type information, so golangci-lint skips every tidygo analyzer on a package that does not type-check.

Build and run:

```sh
golangci-lint custom
./custom-gcl run ./...
```

## Development

Prerequisites: [mise](https://mise.jdx.dev/)

```sh
make init
```

If you use [rimba](https://github.com/lugassawan/rimba) worktrees, `.rimba/settings.toml` runs `make init` automatically on `rimba add`. rimba skips committed `post_create` commands until you approve them, so run `rimba trust` once per clone, and again whenever the `post_create` commands change.

| Target     | Description                                       |
|------------|---------------------------------------------------|
| `init`     | Trust mise config, install tools, set up git hooks|
| `build`    | Build custom golangci-lint with tidygo plugin     |
| `lint`     | Build and run golangci-lint (dogfooding)          |
| `fmt`      | Format all Go files with gofmt and golines        |
| `test`     | Run all tests                                     |
| `coverage` | Generate HTML coverage report                     |
| `release`  | Tag and push a semver release (VERSION=vX.Y.Z)   |

## License

MIT License — see [LICENSE](LICENSE) for details.
