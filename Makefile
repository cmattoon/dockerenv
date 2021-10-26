.DEFAULT_GOAL := help

.PHONY: help
help: ## display this message
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## build dockerenv binary
build: tidy
	go build -o bin/dockerenv cmd/dockerenv/dockerenv.go

.PHONY: all
all: nix osx

.PHONY: nix
nix:
	GOOS=linux GOARCH=amd64 go build -o bin/dockerenv-linux cmd/dockerenv/dockerenv.go

.PHONY: osx
osx:
	GOOS=darwin GOARCH=amd64 go build -o bin/dockerenv-osx cmd/dockerenv/dockerenv.go

.PHONY: test
test: ## run tests
test:
	go test ./...

.PHONY: tidy
tidy: ## go mod tidy
tidy:
	go mod tidy && go mod vendor
