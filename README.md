<div align="center">

# tidygo

[![CI](https://github.com/lugassawan/tidygo/actions/workflows/test.yml/badge.svg)](https://github.com/lugassawan/tidygo/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lugassawan/tidygo.svg)](https://pkg.go.dev/github.com/lugassawan/tidygo)
[![Go Report Card](https://goreportcard.com/badge/github.com/lugassawan/tidygo)](https://goreportcard.com/report/github.com/lugassawan/tidygo)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GitHub Release](https://img.shields.io/github/v/release/lugassawan/tidygo)](https://github.com/lugassawan/tidygo/releases)

A [golangci-lint](https://golangci-lint.run/) v2 plugin that enforces code ordering and structural conventions in Go source files.

</div>

## Analyzers

| Analyzer | Description |
|---|---|
| `funcname` | Forbids underscores in function and method names (optionally variables, constants, parameters, results and receivers) |
| `ifacename` | Requires single-method interfaces to be named after their method plus an agent-noun suffix (`Read` → `Reader`) |
| `ifinitonly` | Forbids `if` init variables used only in the condition |
| `maxcaselines` | Forbids `case` clause bodies that span more than 10 lines (configurable) |
| `maxcases` | Forbids `switch`, type switch and `select` statements with more than 30 case clauses (configurable) |
| `maxparams` | Forbids functions with more than 7 parameters |
| `noanonnested` | Forbids anonymous structs nested more than 1 level deep inside other anonymous structs (configurable) |
| `nolateconst` | Forbids `const`/`var` declarations after function declarations |
| `nolateexport` | Forbids exported functions after unexported functions, and exported methods after unexported methods of the same type |
| `nolocalstruct` | Forbids named struct type declarations inside function bodies (optionally interfaces and defined types too) |
| `nonestedswitch` | Forbids a `switch` or type switch nested inside another `switch` or type switch |

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

| Target     | Description                                       |
|------------|---------------------------------------------------|
| `init`     | Install tools, trust mise config, set up git hooks|
| `build`    | Build custom golangci-lint with tidygo plugin     |
| `lint`     | Build and run golangci-lint (dogfooding)          |
| `fmt`      | Format all Go files with gofmt and golines        |
| `test`     | Run all tests                                     |
| `coverage` | Generate HTML coverage report                     |
| `release`  | Tag and push a semver release (VERSION=vX.Y.Z)   |

## License

MIT License — see [LICENSE](LICENSE) for details.
