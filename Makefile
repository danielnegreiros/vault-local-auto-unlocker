COVERAGE_FILE := coverage.out
GO_BUILD_OUTPUT := bin/app

.PHONY: all help build tidy lint vet test coverage format clean profile

##@ General

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

clean: ## Remove generated files
	@echo "Cleaning up..."
	rm -rf $(GO_BUILD_OUTPUT) $(COVERAGE_FILE)

##@ Build

build: tidy ## Build app
	@echo "Building the application..."
	go build -o $(GO_BUILD_OUTPUT) main.go

##@ Formatting and Linting

format: ## Format code
	@echo "Formatting the code..."
	gofmt -s -w .

tidy: ## Clean up go.mod and go.sum
	@echo "Tidying up module dependencies..."
	go mod tidy

lint: ## Run static code analysis
	@echo "Running linter..."
	golangci-lint run ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

##@ Testing

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	go test ./... -coverprofile temp/c.out
	go tool cover -html temp/c.out -o temp/c.html

profile: ## Run benchmarks with profiling
	@echo "Running benchmarks..."
	for pkg in $(shell go list ./...); do \
			go test -bench=. -benchmem -cpuprofile $$pkg.cpu.prof -memprofile $$pkg.mem.prof $$pkg; \
	done

integration:
	podman stop vault-new  > /dev/null 2>&1 || true
	podman rm vault-new  > /dev/null 2>&1 || true
	rm -rf tests/vault/data
	mkdir -p tests/vault/data
	cd tests/vault/ && podman-compose up -d && cd -
	go run main.go
	# podman stop vault-new  > /dev/null 2>&1 || true
	# podman rm vault-new  > /dev/null 2>&1 || true

itgr_idem:
	cd tests/vault/ && podman-compose restart && cd -
	go run main.go

run:
	go run main.go

##@ Utilities

all: format tidy lint vet test coverage build ## Run all stages: format, tidy, lint, vet, test, coverage, build
