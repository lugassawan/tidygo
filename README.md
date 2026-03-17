<div align="center">

# tidygo

[![CI](https://github.com/lugassawan/tidygo/actions/workflows/test.yml/badge.svg)](https://github.com/lugassawan/tidygo/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lugassawan/tidygo.svg)](https://pkg.go.dev/github.com/lugassawan/tidygo)
[![Go Report Card](https://goreportcard.com/badge/github.com/lugassawan/tidygo)](https://goreportcard.com/report/github.com/lugassawan/tidygo)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A [golangci-lint](https://golangci-lint.run/) v2 plugin that enforces code ordering and structural conventions in Go source files.

</div>

## Analyzers

| Analyzer | Description |
|---|---|
| `funcname` | Forbids underscores in function and method names |
| `maxparams` | Forbids functions with more than 7 parameters |
| `nolateconst` | Forbids `const`/`var` declarations after function declarations |
| `nolateexport` | Forbids exported functions after unexported functions |
| `nolocalstruct` | Forbids named struct type declarations inside function bodies |

## Installation

Requirements: Go 1.26+, golangci-lint v2.

Create `.custom-gcl.yml`:

```yaml
version: v2.10.1
plugins:
  - module: 'github.com/lugassawan/tidygo'
    version: v1.0.0
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
