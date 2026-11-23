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

Each name links to the analyzer's page, grouped in the [index](docs/analyzers/README.md).

| Analyzer | Description |
|---|---|
| [`aeadcipher`](docs/analyzers/aeadcipher.md) | Reports unauthenticated cipher modes for encryption (CBC, CTR) |
| [`atomiccreate`](docs/analyzers/atomiccreate.md) | Reports file creation guarded by an `os.Stat` not-exist check instead of `O_CREATE\|O_EXCL` |
| [`busywait`](docs/analyzers/busywait.md) | Forbids empty or sleep-only `for` loops that spin on a plain variable |
| [`confusingop`](docs/analyzers/confusingop.md) | Reports `=+`, `=-`, `=!` and `=^` assignments that look like typos |
| [`constcond`](docs/analyzers/constcond.md) | Forbids `if` and `for` conditions that are the literal `true` or `false` |
| [`deferredcancel`](docs/analyzers/deferredcancel.md) | Requires the `context` cancel func to be deferred immediately |
| [`dupfunc`](docs/analyzers/dupfunc.md) | Reports functions with the same body and signature as an earlier one (configurable) |
| [`elseifelse`](docs/analyzers/elseifelse.md) | Requires an `if … else if` chain to end in a final `else` (opt-in) |
| [`emptycomment`](docs/analyzers/emptycomment.md) | Forbids `/* */` comments that contain only whitespace and `*` |
| [`emptyfunc`](docs/analyzers/emptyfunc.md) | Requires a comment inside function bodies with no statements |
| [`funcname`](docs/analyzers/funcname.md) | Forbids underscores in function and method names (configurable) |
| [`ifacename`](docs/analyzers/ifacename.md) | Requires single-method interfaces to be named after their method (`Read` → `Reader`) |
| [`ifinitonly`](docs/analyzers/ifinitonly.md) | Forbids `if` init variables used only in the condition |
| [`jwtnone`](docs/analyzers/jwtnone.md) | Reports use of the JWT `none` algorithm |
| [`maxboolops`](docs/analyzers/maxboolops.md) | Forbids boolean expressions with more than 3 `&&` and `\|\|` operators (configurable) |
| [`maxcaselines`](docs/analyzers/maxcaselines.md) | Forbids `case` bodies longer than 10 lines (configurable) |
| [`maxcases`](docs/analyzers/maxcases.md) | Forbids switches with more than 30 case clauses (configurable) |
| [`maxparams`](docs/analyzers/maxparams.md) | Forbids functions with more than 7 parameters |
| [`methodcolocation`](docs/analyzers/methodcolocation.md) | Requires a method to be declared in the same file as its receiver type (configurable) |
| [`noanonnested`](docs/analyzers/noanonnested.md) | Forbids more than 1 level of anonymous struct nesting (configurable) |
| [`nolateconst`](docs/analyzers/nolateconst.md) | Forbids `const`/`var` declarations after function declarations |
| [`nolateexport`](docs/analyzers/nolateexport.md) | Forbids exported functions and methods after unexported ones (configurable) |
| [`nolocalstruct`](docs/analyzers/nolocalstruct.md) | Forbids named struct types declared inside function bodies (configurable) |
| [`nonestedswitch`](docs/analyzers/nonestedswitch.md) | Forbids a `switch` nested inside another `switch` |
| [`redundantparen`](docs/analyzers/redundantparen.md) | Forbids unnecessary parentheses around return values, assignments and call arguments |
| [`txrollback`](docs/analyzers/txrollback.md) | Requires a `*sql.Tx` to have a `defer tx.Rollback()` |
| [`weakpwhash`](docs/analyzers/weakpwhash.md) | Reports password hashing with a constant salt or a too-low constant cost |

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
          # Optional: raise the fewest statements a function body needs to be compared (default 3).
          dupfunc:
            min-statements: 5
```

Every analyzer's options are on its page. Unknown keys are rejected, so a typo fails loudly.

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
