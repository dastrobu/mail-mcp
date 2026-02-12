# GitHub Copilot Instructions for Apple Mail MCP Server

This document provides context and guidelines for GitHub Copilot when working on this MCP server project.

## Project Overview

An MCP (Model Context Protocol) server providing programmatic access to macOS Mail.app using Go and JavaScript for Automation (JXA). The server enables AI assistants and other MCP clients to interact with Apple Mail through a clean, typed interface.

**Key Technologies:**
- Go 1.26+
- MCP Go SDK v1.2.0+ (github.com/modelcontextprotocol/go-sdk)
- JXA (JavaScript for Automation) for Mail.app interaction
- STDIO transport for MCP communication

## Architecture

### Project Structure

```
apple-mail-mcp/
├── cmd/
│   └── mail-mcp-server/              # Main application entry point
│       └── main.go
├── internal/
│   ├── jxa/                          # JXA script execution
│   │   └── executor.go
│   └── tools/                        # MCP tool implementations
│       ├── scripts/                  # Embedded JXA scripts
│       │   ├── list_accounts.js
│       │   ├── list_mailboxes.js
│       │   ├── get_messages.js
│       │   ├── get_message_content.js
│       │   └── search_messages.js
│       ├── list_accounts.go          # Individual tool implementations
│       ├── list_mailboxes.go
│       ├── get_messages.go
│       ├── get_message_content.go
│       ├── search_messages.go
│       └── tools.go                  # Tool registration and helpers
├── .github/
│   ├── workflows/
│   │   └── ci.yaml                   # CI/CD pipeline
│   └── copilot-instructions.md       # This file
├── go.mod                            # Go module dependencies
├── go.sum
├── Makefile                          # Build commands
└── README.md                         # User documentation

```

### Core Principles

1. **Single Binary**: All JXA scripts are embedded using `//go:embed` to create a self-contained executable
2. **STDIO Only**: Server uses STDIO transport exclusively (no HTTP/SSE)
3. **Modular Design**: Each tool is in its own file within `internal/tools/` for maintainability
4. **macOS Only**: Relies on Mail.app and JXA, which are macOS-specific

## Go Development Guidelines

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Always run `gofmt -w` before committing
- Use meaningful variable names: `mailbox` not `mb`, `messageID` not `mid`
- Prefer `any` over `interface{}` for better readability
- Use `context.Context` for all operations
- Handle errors explicitly - never ignore them

### MCP Server Implementation

The server uses the MCP Go SDK v1.2.0+. Key patterns:

### Tool Registration Pattern

Each tool is implemented in its own file within `internal/tools/`. Here's the pattern:

```go
// internal/tools/example_tool.go
package tools

import (
    "context"
    _ "embed"
    "fmt"
    
    "github.com/dastrobu/apple-mail-mcp/internal/jxa"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed scripts/example.js
var exampleScript string

// ExampleInput defines input parameters for example_tool
type ExampleInput struct {
    Param string `json:"param" jsonschema:"title=Parameter,description=Description of parameter"`
}

// RegisterExampleTool registers the example_tool with the MCP server
func RegisterExampleTool(srv *mcp.Server) {
    mcp.AddTool(srv,
        &mcp.Tool{
            Name:        "example_tool",
            Description: "Clear description of what the tool does",
            Annotations: &mcp.ToolAnnotations{
                Title:           "Example Tool",
                ReadOnlyHint:    true,  // All tools are read-only
                IdempotentHint:  true,  // Most read operations are idempotent
                DestructiveHint: new(false),  // Go 1.26+ new() syntax
                OpenWorldHint:   new(true),   // Interacts with Mail.app
            },
        },
        handleExampleTool,
    )
}

func handleExampleTool(ctx context.Context, request *mcp.CallToolRequest, input ExampleInput) (*mcp.CallToolResult, any, error) {
    // Execute JXA script using the jxa package
    data, err := jxa.Execute(ctx, exampleScript, input.Param)
    if err != nil {
        return nil, nil, fmt.Errorf("failed to execute example_tool: %w", err)
    }
    
    // Return nil result to let SDK auto-populate from output
    return nil, data, nil
}
```

Then register in `internal/tools/tools.go`:
```go
func RegisterAll(srv *mcp.Server) {
    RegisterExampleTool(srv)
    // ... other tools
}
```

#### Tool Input/Output Types

**Input Types:**
- Define structs with `jsonschema` tags for validation
- Use descriptive field names and documentation
- Make all parameters required - avoid optional parameters and defaults
- **jsonschema format**: Use plain description strings only (e.g., `jsonschema:"Description here"`)
- **Do NOT use**: `key=value` pairs like `required,title=X,description=Y` - the SDK will reject these

```go
type ExampleInput struct {
    Account string `json:"account" jsonschema:"Name of the email account"`
    Mailbox string `json:"mailbox" jsonschema:"Name of the mailbox to access"`
    Limit   int    `json:"limit" jsonschema:"Maximum number of items (1-1000)"`
    Enabled bool   `json:"enabled" jsonschema:"Filter flag (true or false)"`
}
```

**Important:** The jsonschema tag accepts only a plain description string, NOT `key=value` pairs like `required,title=X,description=Y`. The SDK will reject tags with that format.

**Output Types:**
- Always use generic JSON types: `map[string]any` or slices
- DO NOT define custom output structs
- Provides flexibility as APIs evolve

```go
// Good
var output map[string]any
output = map[string]any{
    "messages": []map[string]any{
        {"subject": "Test", "from": "user@example.com"},
    },
    "count": 1,
}
return nil, output, nil

// Bad - don't do this
type ExampleOutput struct {
    Messages []Message
    Count    int
}
```

#### Tool Annotations

All tools MUST set `Annotations` to provide metadata:

- `Title`: Human-readable tool name
- `ReadOnlyHint`: Always `true` (this is a read-only server)
- `IdempotentHint`: `true` for tools that produce same result on repeated calls
- `OpenWorldHint`: Set to `new(true)` (interacts with Mail.app external system)
- `DestructiveHint`: Set to `new(false)` (read-only operations)

**Go 1.26+ new() Syntax:**
Go 1.26 introduces enhanced `new()` that accepts expressions to create pointers:
```go
new(true)    // Creates *bool pointing to true
new(false)   // Creates *bool pointing to false
new("text")  // Creates *string pointing to "text"
new(42)      // Creates *int pointing to 42
```

### JXA Script Embedding

All JXA scripts must be embedded using `//go:embed` in their respective tool files:

```go
//go:embed scripts/my_script.js
var myScript string
```

Scripts are located in `internal/tools/scripts/` and embedded into the tool files.

**Benefits:**
- Single binary distribution
- No runtime file dependencies
- Scripts bundled at compile time
- Each tool owns its script

### JXA Script Guidelines

#### Structure

All JXA scripts must follow this pattern:

```javascript
function run(argv) {
    const Mail = Application('Mail');
    Mail.includeStandardAdditions = true;
    
    // Parse arguments - extract without defaults
    const param1 = argv[0] || '';
    const param2 = argv[1] || '';
    const param3 = argv[2] ? parseInt(argv[2]) : 0;
    
    // Validate required arguments explicitly
    if (!param1) {
        return JSON.stringify({
            success: false,
            error: 'Parameter 1 is required'
        });
    }
    
    if (!param2) {
        return JSON.stringify({
            success: false,
            error: 'Parameter 2 is required'
        });
    }
    
    if (!param3 || param3 < 1) {
        return JSON.stringify({
            success: false,
            error: 'Parameter 3 is required and must be at least 1'
        });
    }
    
    try {
        // Perform operations
        const result = doSomething(param1, param2, param3);
        
        // Return success with data wrapped in 'data' field
        return JSON.stringify({
            success: true,
            data: result
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
- Parse all arguments at the top using `argv[n] || ''` or `parseInt(argv[n])`
- Validate each required argument explicitly with clear error messages
- Return error JSON immediately if validation fails
- Only proceed to try-catch block after all validations pass
- Never use default values in argument parsing - make parameters required

#### Mail.app Interaction

- Initialize: `const Mail = Application('Mail'); Mail.includeStandardAdditions = true;`
- Access properties via methods: `mailbox.name()`, `message.subject()`
- Convert dates to ISO: `msg.dateReceived().toISOString()`
- Return JSON strings: `JSON.stringify({ success: true, data: result })`

#### Output Format

**CRITICAL:** All JXA scripts MUST wrap their output in a `data` field:

```javascript
// Correct format
return JSON.stringify({
    success: true,
    data: {
        accounts: [...],
        count: 5
    }
});

// WRONG - do not return data at top level
return JSON.stringify({
    success: true,
    accounts: [...],
    count: 5
});
```

The `jxa.Execute` function expects and extracts the `data` field. Scripts that don't wrap output will fail with "missing 'data' field" error.

#### Argument Handling and Validation

**Separation of Concerns:**

- **Go Layer**: Handles optional parameter defaults only
- **JXA Layer**: Validates all arguments and enforces constraints

**Go Side - Default Values Only:**
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

All required arguments must be validated before processing:

1. Parse arguments at the top of the function with safe fallbacks
2. Validate each required argument explicitly
3. Return error JSON with descriptive message for missing/invalid arguments
4. Only enter try-catch block after validation passes

```javascript
// Correct pattern - parse with safe fallbacks
const accountName = argv[0] || '';
const mailboxName = argv[1] || '';
const limit = argv[2] ? parseInt(argv[2]) : 0;

// Validate each argument explicitly
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

// Now safe to proceed with try-catch
try {
    // ... implementation
}
```

**WRONG - Don't use defaults in argument parsing:**
```javascript
// Bad - don't do this
const accountName = argv[0] || 'default';  // No! Validate instead
const limit = parseInt(argv[1]) || 50;     // No! Validate instead
```

**Why This Pattern:**
- Go layer keeps defaults close to the API definition
- JXA layer enforces data integrity before execution
- Clear error messages come from the validation layer
- No redundant validation between layers

#### Error Handling

- Validate all arguments before try-catch
- Always wrap operations in try-catch
- Return structured JSON: `{success: bool, data/error: ...}`
- Include descriptive error messages

### Error Handling Pattern

The `internal/jxa` package handles JXA execution and error checking:

```go
// Execute JXA using the jxa package
data, err := jxa.Execute(ctx, script, args...)
if err != nil {
    return nil, nil, fmt.Errorf("failed to execute tool: %w", err)
}

// data is already validated and extracted by jxa.Execute
return nil, data, nil
```

The `jxa.Execute` function:
- Runs the osascript command
- Parses JSON output
- Checks for script-level errors
- Extracts and returns the `data` field (REQUIRED in all scripts)
- Returns descriptive errors for any failures

**Important:** Scripts must wrap all output in a `data` field. The executor will fail if the `data` field is missing.

## Testing

### Running Tests

```bash
# Run all tests (disable cache)
go test -v -count=1 ./...

# Test specific package
go test -v -count=1 ./internal/scripts
```

### Test Structure

- Prefer table-driven tests for multiple scenarios
- Test error cases and edge conditions
- Mock external dependencies where possible

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    map[string]any
        wantErr bool
    }{
        {"valid input", "test", map[string]any{"result": "ok"}, false},
        {"empty input", "", nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ExampleFunc(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExampleFunc() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Assert results
        })
    }
}
```

## Build & Development

### Commands

```bash
# Build
make build
# or
go build -o mail-mcp-server

# Clean
make clean

# Format code
gofmt -w .

# Vet code
go vet ./...

# Run tests
go test -v -count=1 ./...

# Test server starts (important after any changes!)
./mail-mcp-server
# Should start without panicking. Press Ctrl+C to stop.
```

### Testing After Changes

**CRITICAL**: Always test that the server starts after making changes:

```bash
# Rebuild and test
go build -o mail-mcp-server . && ./mail-mcp-server &
SERVER_PID=$!
sleep 2
if ps -p $SERVER_PID > /dev/null; then
    echo "Server started successfully"
    kill $SERVER_PID
else
    echo "Server failed to start"
fi
```

Or simply:
```bash
go build -o mail-mcp-server . && timeout 2s ./mail-mcp-server
# If it times out (exit code 124), the server started successfully
# If it panics immediately, you'll see the error
```

### CI/CD

- `.github/workflows/ci.yaml` defines the build pipeline
- CI runs on macOS runners (Mail.app dependency)
- Steps: format check, vet, build, test, lint
- Always ensure CI passes before merging

## Common Patterns

### JXA Script Execution

JXA execution is handled by the `internal/jxa` package:

```go
// internal/jxa/executor.go
func Execute(ctx context.Context, script string, args ...string) (map[string]any, error) {
    // Build osascript command
    cmdArgs := []string{"-l", "JavaScript", "-e", script}
    cmdArgs = append(cmdArgs, args...)
    
    cmd := exec.CommandContext(ctx, "osascript", cmdArgs...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("script execution failed: %w (output: %s)", err, string(output))
    }
    
    // Parse JSON output
    var result map[string]any
    if err := json.Unmarshal(output, &result); err != nil {
        return nil, fmt.Errorf("failed to parse script output: %w (output: %s)", err, string(output))
    }
    
    // Check for script-level errors
    if success, ok := result["success"].(bool); !ok || !success {
        errMsg := "unknown error"
        if errVal, ok := result["error"].(string); ok {
            errMsg = errVal
        }
        return nil, fmt.Errorf("script error: %s", errMsg)
    }
    
    // Extract and return data
    data, ok := result["data"].(map[string]any)
    if !ok {
        return nil, fmt.Errorf("invalid script output format: missing or invalid 'data' field")
    }
    
    return data, nil
}
```

### Context Handling

Always pass and respect context for cancellation:

```go
func myHandler(ctx context.Context, request *mcp.CallToolRequest, input MyInput) (*mcp.CallToolResult, any, error) {
    // Check context before expensive operations
    select {
    case <-ctx.Done():
        return nil, nil, ctx.Err()
    default:
    }
    
    // Pass context to child operations
    result, err := executeJXAScript(ctx, script, args...)
    // ...
}
```

## Important Constraints

1. **macOS Only**: Mail.app and JXA are macOS-specific
2. **Mail.app Required**: Mail.app must be running for operations to succeed
3. **Read-Only**: Never implement operations that modify mail data
4. **STDIO Transport**: Only support stdio, no HTTP/SSE
5. **No External State**: Keep server stateless for simplicity

## Security Considerations

- All operations are read-only by design
- No sending emails or modifying messages
- Server runs locally on user's machine
- Mail.app's security and permissions apply
- Never log sensitive email content

## Documentation

- Always update `README.md` when adding new tools or features
- Document all exported functions and types
- Keep inline comments minimal - prefer self-documenting code
- Update this file when adding significant patterns or guidelines

## Common Pitfalls

1. **Don't forget `//go:embed` directive** for JXA scripts
2. **Mail.app must be running** - scripts fail if not running
3. **Always use `map[string]any` for outputs** - not custom structs
4. **Always wrap JXA output in `data` field** - executor requires this
5. **Validate in JXA, defaults in Go** - keep validation in JXA scripts, only handle defaults in Go
6. **Safe argument parsing** - use `argv[2] ? parseInt(argv[2]) : 0` not `parseInt(argv[2])`
7. **jsonschema tags use plain strings** - NOT `key=value` format (e.g., use `jsonschema:"Description"` not `jsonschema:"required,description=X"`)
8. **Always test server starts** - run `./mail-mcp-server` after changes to catch schema errors
9. **Check script success field** before processing data
10. **Type assertions need checking**: `val, ok := x.(string)`
11. **JSON numbers are float64** - convert to int as needed
12. **Context cancellation** - always pass and respect ctx

## Adding New Tools

Checklist for adding a new MCP tool:

1. Create JXA script in `internal/tools/scripts/` directory
2. Create new tool file: `internal/tools/my_new_tool.go`
3. Add `//go:embed scripts/my_script.js` directive
4. Define input struct with jsonschema tags
5. Implement `RegisterMyNewTool(srv *mcp.Server)` function
6. Implement handler function following patterns above
7. Add tool registration to `RegisterAll()` in `internal/tools/tools.go`
8. Write tests if needed
9. Update `README.md` with tool documentation
10. Test with an MCP client (e.g., Claude Desktop)

Each tool should be self-contained in its own file within `internal/tools/`.

## Version Control

- Write clear, descriptive commit messages
- Keep commits focused and atomic
- Run `gofmt -w` before committing
- Ensure tests pass before pushing
- Avoid interactive Git operations in automation

## Questions?

When implementing features:
1. Check existing patterns in the codebase
2. Refer to MCP Go SDK documentation
3. Follow Go best practices
4. Keep it simple and maintainable
5. Test with Mail.app running
