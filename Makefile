.PHONY: test lint fmt coverage init build release

build:
	golangci-lint custom

lint: build
	./custom-gcl run ./...

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
	mise install
	mise trust
	git config core.hooksPath .githooks

release:
ifndef VERSION
	$(error VERSION is required. Usage: make release VERSION=v1.0.0)
endif
	@echo "$(VERSION)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$$' || (echo "ERROR: VERSION must match vX.Y.Z" && exit 1)
	git tag -a "$(VERSION)" -m "Release $(VERSION)"
	git push origin "$(VERSION)"
	@echo "Tagged and pushed $(VERSION)"
