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
mise install && mise trust && make init
```

| Target     | Description                                       |
|------------|---------------------------------------------------|
| `build`    | Build custom golangci-lint with tidygo plugin     |
| `lint`     | Build and run golangci-lint (dogfooding)          |
| `fmt`      | Format all Go files with gofmt and golines        |
| `test`     | Run all tests                                     |
| `coverage` | Generate HTML coverage report                     |
| `init`     | Configure git hooks and trust mise config         |

## License

MIT License — see [LICENSE](LICENSE) for details.
