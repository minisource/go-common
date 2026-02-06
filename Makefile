.PHONY: test test-coverage lint fmt vet build clean help

# Go Common Library Makefile

# Default target
all: lint test

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests for specific package
test-pkg:
	@echo "Running tests for package $(PKG)..."
	@go test -v ./$(PKG)/...

# Run linter
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	@gofmt -s -w .
	@goimports -w .

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Build (verify compilation)
build:
	@echo "Verifying compilation..."
	@go build ./...

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -f coverage.out coverage.html
	@go clean -cache

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download

# Verify dependencies
verify:
	@echo "Verifying dependencies..."
	@go mod verify

# Run benchmarks
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Generate mocks (if using mockgen)
mocks:
	@echo "Generating mocks..."
	@go generate ./...

# Install development tools
tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golang/mock/mockgen@latest

# Check for vulnerabilities
vuln:
	@echo "Checking for vulnerabilities..."
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Run lint and test (default)"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-pkg      - Run tests for specific package (PKG=package)"
	@echo "  lint          - Run golangci-lint"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  build         - Verify compilation"
	@echo "  clean         - Clean up generated files"
	@echo "  tidy          - Tidy go.mod"
	@echo "  deps          - Download dependencies"
	@echo "  verify        - Verify dependencies"
	@echo "  bench         - Run benchmarks"
	@echo "  mocks         - Generate mocks"
	@echo "  tools         - Install development tools"
	@echo "  vuln          - Check for vulnerabilities"
	@echo ""
	@echo "Examples:"
	@echo "  make test"
	@echo "  make test-pkg PKG=logging"
	@echo "  make test-coverage"
