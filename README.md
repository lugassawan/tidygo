# tidygo

A [golangci-lint](https://golangci-lint.run/) v2 plugin that enforces code ordering and structural conventions in Go source files.

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

Add to your `go.mod`:

```sh
go get github.com/lugassawan/tidygo
```

Import in a blank import file (e.g. `tools/lint.go`):

```go
//go:build lint

package tools

import _ "github.com/lugassawan/tidygo"
```

Configure `.golangci.yml`:

```yaml
version: "2"

plugins:
  module:
    tidygo:
      path: github.com/lugassawan/tidygo

linters:
  enable:
    - funcname
    - maxparams
    - nolateconst
    - nolateexport
    - nolocalstruct
```

## Development

Prerequisites: [mise](https://mise.jdx.dev/)

```sh
mise install && mise trust && make init
```

| Target     | Description                                    |
|------------|------------------------------------------------|
| `lint`     | Run golangci-lint on all packages              |
| `fmt`      | Format all Go files with gofmt and golines     |
| `test`     | Run all tests                                  |
| `coverage` | Generate HTML coverage report                  |
| `init`     | Configure git hooks and trust mise config      |

## License

MIT License — see [LICENSE](LICENSE) for details.
