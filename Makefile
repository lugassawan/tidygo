.PHONY: test lint fmt coverage init

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	golines -w --max-len=120 .

test:
	go test ./...

coverage:
	mkdir -p coverage
	go test -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html

init:
	mise trust
	git config core.hooksPath .githooks
