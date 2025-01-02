# Define the common test flags
TEST_FLAGS := -race -v

# Target for running unit tests (excluding integration tests)
test: clean
	@echo "Running unit tests..."
	@go test $(TEST_FLAGS) -short ./...

# Target for running integration tests (including integration tests)
test-integration: clean
	@echo "Running integration tests..."
	@go test $(TEST_FLAGS) ./...

# Target for generating code coverage
coverage: clean
	@echo "Generating coverage report..."
	@go test $(TEST_FLAGS) -coverprofile=coverage.out ./...

# Clean the cache
clean:
	@go clean -testcache
