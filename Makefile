.PHONY: build clean test test-scripts fmt vet install-hooks help

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

# Test JXA scripts directly (requires Mail.app running)
test-scripts:
	@echo "Testing list_mailboxes.js..."
	@osascript -l JavaScript internal/tools/scripts/list_mailboxes.js || echo "Failed"
	@echo "\nTesting get_messages.js..."
	@osascript -l JavaScript internal/tools/scripts/get_messages.js "INBOX" "" 5 || echo "Failed"
	@echo "\nTesting search_messages.js..."
	@osascript -l JavaScript internal/tools/scripts/search_messages.js "test" "subject" 10 || echo "Failed"
	@echo "\nNote: get_message_content.js requires a valid message ID"

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

# Display help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  clean        - Remove build artifacts"
	@echo "  test         - Run Go tests"
	@echo "  test-scripts - Test JXA scripts directly (requires Mail.app)"
	@echo "  fmt          - Format Go code"
	@echo "  vet          - Run go vet"
	@echo "  deps         - Download and verify dependencies"
	@echo "  tidy         - Tidy dependencies"
	@echo "  check        - Run fmt, vet, and test"
	@echo "  install-hooks - Install git pre-commit hook"
	@echo "  help         - Display this help message"
