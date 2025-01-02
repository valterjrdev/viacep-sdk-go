TEST_FLAGS := -race -v
LINT_CMD := golangci-lint run

# Target for running unit tests (excluding integration tests)
.PHONY: test
test: clean
	@echo "Running unit tests..."
	@go test $(TEST_FLAGS) -short ./...

# Target for running integration tests (including integration tests)
.PHONY: test-integration
test-integration: clean
	@echo "Running integration tests..."
	@go test $(TEST_FLAGS) ./...

# Target for generating code coverage
.PHONY: coverage
coverage: clean
	@echo "Generating coverage report..."
	@go test $(TEST_FLAGS) -coverprofile=coverage.out ./...

# Target to run golangci-lint
.PHONY: lint
lint:
	@echo "Running lint..."
	@golangci-lint run -v ./...

# Clean the cache
.PHONY: clean
clean:
	@go clean -testcache
