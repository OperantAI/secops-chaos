BINARY_NAME = experiments-runtime-tool
REPO_NAME = github.com/operantai/$(BINARY_NAME)
GIT_COMMIT = $(shell git rev-list -1 HEAD)
BUILD_DATE = $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION = $(shell git describe --tags --always --dirty)
LD_FLAGS = "-X $(REPO_NAME)/cmd.GitCommit=$(GIT_COMMIT) -X $(REPO_NAME)/cmd.Version=$(GIT_COMMIT) -X $(REPO_NAME)/cmd.BuildDate=$(BUILD_DATE)"

all: fmt vet test build

build: ## Build binary
	@go build -o "bin/$(BINARY_NAME)" -ldflags $(LD_FLAGS) main.go

fmt: ## Run go fmt
	@go fmt ./...

vet: ## Run go vet
	@go vet ./...

lint: ## Run linter
	@golangci-lint run

test: ## Run tests
	@go test -cover -v ./...

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
