.PHONY: build clean test fmt vet install-hooks doctoc help

# Binary name
BINARY=apple-mail-mcp

# Build the binary
build:
	go build -o $(BINARY) .

# Clean build artifacts
clean:
	rm -f $(BINARY)
	go clean

# Run Go tests
test:
	go test -v -count=1 ./...



# Format Go code
fmt:
	gofmt -w .

# Run go vet
vet:
	go vet ./...

# Download and verify dependencies
deps:
	go mod download
	go mod verify

# Tidy dependencies
tidy:
	go mod tidy

# Run all checks (format, vet, test)
check: fmt vet test

# Install git hooks
install-hooks:
	@./scripts/install-hooks.sh

# Update README Table of Contents
doctoc:
	npx doctoc --maxlevel 2 README.md --github --title "## Table of Contents"

# Display help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run Go tests"

	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Download and verify dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  check        - Run fmt, vet, and test"
	@echo "  install-hooks - Install git pre-commit hook"
	@echo "  doctoc       - Update README Table of Contents"
	@echo "  help         - Display this help message"
