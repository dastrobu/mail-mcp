# Contributing to Apple Mail MCP Server

Thank you for your interest in contributing to the Apple Mail MCP Server! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- macOS (Mail.app is macOS-only)
- Go 1.25 or later
- Mail.app configured with at least one email account
- Git

### Setting Up Your Development Environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/apple-mail-mcp.git
   cd apple-mail-mcp
   ```

3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/dastrobu/apple-mail-mcp.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Build the project:
   ```bash
   make build
   ```

## Development Workflow

### Before You Start

1. Check existing issues and pull requests to avoid duplicates
2. For major changes, open an issue first to discuss your proposal
3. Create a new branch for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Making Changes

1. **Write clean, idiomatic Go code**
   - Follow the project's code style (see `.github/copilot-instructions.md`)
   - Use `gofmt` to format your code: `make fmt`
   - Run `go vet` to catch common issues: `make vet`

2. **Test your changes**
   - Write tests for new functionality
   - Run existing tests: `make test`
   - Test JXA scripts if modified: `make test-scripts`
   - Ensure all tests pass before submitting

3. **Update documentation**
   - Update `README.md` if you add new features or tools
   - Add comments to exported functions and types
   - Update `.github/copilot-instructions.md` if adding new patterns

4. **Commit your changes**
   - Write clear, descriptive commit messages
   - Use conventional commit format (optional but recommended):
     ```
     feat: add new search_by_date tool
     fix: handle empty mailbox names correctly
     docs: update installation instructions
     test: add tests for message parsing
     ```

### Testing

#### Go Tests

```bash
# Run all tests
make test

# Run specific test
go test -v -count=1 ./internal/scripts -run TestListMailboxes
```

#### JXA Script Tests

```bash
# Test all scripts (requires Mail.app running)
make test-scripts

# Test individual script
osascript -l JavaScript scripts/list_mailboxes.js
```

#### Manual Testing with MCP Client

Test your changes with an MCP client like Claude Desktop:

1. Build the binary: `make build`
2. Configure your MCP client to use the binary
3. Test the tools through the client interface

### Code Quality Checks

Run all checks before submitting:

```bash
make check
```

This runs:
- `gofmt` - Code formatting
- `go vet` - Static analysis
- `go test` - All tests

### Submitting Your Changes

1. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. Open a pull request on GitHub:
   - Use a clear title and description
   - Reference any related issues
   - Describe what changed and why
   - Include screenshots/examples if applicable

3. Wait for review:
   - Address any feedback from maintainers
   - Keep your branch up to date with main
   - Be patient and respectful

## Code Style Guidelines

### Go Code

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable names
- Prefer `any` over `interface{}`
- Always handle errors explicitly
- Use `context.Context` for operations
- Document exported functions and types
- Keep functions focused and testable

### JXA Scripts

- Always wrap in `function run(argv) { ... }`
- Parse arguments at the top with safe fallbacks: `argv[0] || ''` or `argv[2] ? parseInt(argv[2]) : 0`
- Validate all required arguments explicitly before try-catch
- Return error JSON immediately for validation failures
- Use try-catch for error handling
- Return JSON with `{success: bool, data/error: ...}`
- Always wrap output data in a `data` field
- Initialize Mail.app properly
- Convert dates to ISO strings
- Use descriptive variable names

### Tool Implementation

- Use `jsonschema` tags for input documentation
- Output types must use `map[string]any` or slices
- Set `Annotations` on all tools
- Include clear descriptions

### Argument Handling and Validation

**Separation of Concerns:**
- **Go Layer**: Handles optional parameter defaults only
- **JXA Layer**: Validates all arguments and enforces constraints

**Go Side - Default Values:**
```go
func handleTool(ctx context.Context, request *mcp.CallToolRequest, input ToolInput) (*mcp.CallToolResult, any, error) {
    // Only apply defaults for optional parameters
    limit := input.Limit
    if limit == 0 {
        limit = 5 // default value
    }
    
    // Pass to JXA - validation happens there
    data, err := jxa.Execute(ctx, script, account, mailbox, fmt.Sprintf("%d", limit))
    // ...
}
```

**JXA Side - Full Validation:**
```javascript
function run(argv) {
    const Mail = Application('Mail');
    Mail.includeStandardAdditions = true;
    
    // 1. Parse arguments with safe fallbacks
    const accountName = argv[0] || '';
    const mailboxName = argv[1] || '';
    const limit = argv[2] ? parseInt(argv[2]) : 0;
    
    // 2. Validate each argument explicitly
    if (!accountName) {
        return JSON.stringify({
            success: false,
            error: 'Account name is required'
        });
    }
    
    if (!mailboxName) {
        return JSON.stringify({
            success: false,
            error: 'Mailbox name is required'
        });
    }
    
    if (!limit || limit < 1) {
        return JSON.stringify({
            success: false,
            error: 'Limit is required and must be at least 1'
        });
    }
    
    if (limit > 100) {
        return JSON.stringify({
            success: false,
            error: 'Limit cannot exceed 100'
        });
    }
    
    // 3. Only proceed to try-catch after validation passes
    try {
        // ... implementation
        return JSON.stringify({
            success: true,
            data: { ... }  // Always wrap in data field
        });
    } catch (e) {
        return JSON.stringify({
            success: false,
            error: e.toString()
        });
    }
}
```

**Key Points:**
- Parse with safe fallbacks to avoid NaN/undefined
- Validate each required argument with descriptive errors
- No default values in argument parsing - make parameters required
- Keep validation in JXA layer, not Go layer
- Return errors immediately, don't defer to try-catch

## Adding New Tools

1. Create JXA script in `scripts/` directory
2. Embed script using `//go:embed` in `main.go`
3. Define input struct with jsonschema tags
4. Implement handler function
5. Register tool with `mcp.AddTool()`
6. Write tests
7. Update documentation

Example:

```go
//go:embed scripts/my_tool.js
var myToolScript string

type MyToolInput struct {
    Param string `json:"param" jsonschema:"description=Parameter description"`
}

mcp.AddTool(srv,
    &mcp.Tool{
        Name:        "my_tool",
        Description: "What this tool does",
        Annotations: &mcp.ToolAnnotations{
            Title:           "My Tool",
            ReadOnlyHint:    true,
            IdempotentHint:  true,
            OpenWorldHint:   new(true),  // Go 1.26+ new() syntax
        },
    },
    func(ctx context.Context, request *mcp.CallToolRequest, input MyToolInput) (*mcp.CallToolResult, any, error) {
        // Implementation
    },
)
```

## Reporting Issues

### Bug Reports

Include:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Go version, macOS version
- Mail.app configuration (if relevant)
- Logs or error messages

### Feature Requests

Include:
- Clear description of the feature
- Use case and motivation
- Example of how it would work
- Any relevant prior art

## Code Review Process

1. Maintainers will review your PR
2. Address feedback in new commits
3. Keep discussions focused and professional
4. Once approved, a maintainer will merge

## Questions?

- Open an issue for questions about contributing
- Check existing documentation first
- Be clear and specific in your questions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

Thank you for contributing! 🎉
