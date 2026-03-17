# tidygo

A golangci-lint v2 plugin that enforces code ordering conventions in Go files.

## Prerequisites

- [mise](https://mise.jdx.dev/)

## Setup

```sh
mise install
make init
```

## Make Targets

| Target     | Description                                        |
|------------|----------------------------------------------------|
| `lint`     | Run golangci-lint on all packages                  |
| `fmt`      | Format all Go files with gofmt and golines         |
| `test`     | Run all tests                                      |
| `coverage` | Run tests and generate an HTML coverage report     |
| `init`     | Configure git to use the local `.githooks` directory |

## License

MIT License — see [LICENSE](LICENSE) for details.
