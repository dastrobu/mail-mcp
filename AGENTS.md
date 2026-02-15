# GitHub Copilot Instructions for Apple Mail MCP Server

This document provides context and guidelines for GitHub Copilot when working on this MCP server project.

## Project Overview

An MCP (Model Context Protocol) server providing programmatic access to macOS Mail.app using Go and JavaScript for Automation (JXA). The server enables AI assistants and other MCP clients to interact with Apple Mail through a clean, typed interface.

**Key Technologies:**
- Go 1.26+
- MCP Go SDK v1.2.0+ (github.com/modelcontextprotocol/go-sdk)
- JXA (JavaScript for Automation) for Mail.app interaction
- HTTP and STDIO transports for MCP communication (HTTP recommended for automation permissions)

## Architecture

### Core Principles

1. **Single Binary**: All JXA scripts are embedded using `//go:embed` to create a self-contained executable
2. **Dual Transport**: Supports both HTTP (recommended) and STDIO transports
3. **Modular Design**: Each tool is in its own file within `internal/tools/` for maintainability
4. **macOS Only**: Relies on Mail.app and JXA, which are macOS-specific
5. **Nested Mailbox Support**: All tools support hierarchical mailboxes via mailbox path arrays

## Go Development Guidelines

### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Always run `gofmt -w` before committing
- Use meaningful variable names: `mailbox` not `mb`, `messageID` not `mid`
- Prefer `any` over `interface{}` for better readability
- Use `context.Context` for all operations
- Handle errors explicitly - never ignore them

### Typed Flags

Use typed flags for better validation and completion support:

```go
// internal/opts/typed_flags/transport_flag.go
type Transport string

const (
    TransportStdio Transport = "stdio"
    TransportHTTP  Transport = "http"
)

var TransportValues = []Transport{
    TransportStdio,
    TransportHTTP,
}

// Implement flags.Completer and flags.Unmarshaler
func (t *Transport) Complete(match string) (completions []flags.Completion)
func (t *Transport) UnmarshalFlag(value string) error
```

**Benefits:**
- Type safety
- Automatic validation
- Tab completion support
- Clear valid values

### Subcommands

The server supports subcommands for managing launchd services:

```bash
# Show help
./apple-mail-mcp -h                    # Main help
./apple-mail-mcp launchd -h            # Launchd subcommands
./apple-mail-mcp launchd create -h     # Create options

# Run the server (default command)
./apple-mail-mcp
./apple-mail-mcp --transport=http

# Set up launchd service (HTTP mode, no Terminal parent process)
./apple-mail-mcp launchd create
./apple-mail-mcp --port=3000 launchd create
./apple-mail-mcp --debug launchd create  # With debug logging

# Remove launchd service
./apple-mail-mcp launchd remove
```

**Implementation:**
- Commands are parsed in `internal/opts/opts.go`
- `launchd` is a main command with subcommands: `create` and `remove`
- Launchd logic is in `internal/launchd/setup.go`
- Main command handler is in `main.go`
- Help is enabled via `flags.HelpFlag` and explicitly printed via `parser.WriteHelp(os.Stdout)` when `flags.ErrHelp` is detected

**Why launchd?** When launched via launchd, the binary runs without a parent process, so macOS grants automation permissions to the binary itself rather than Terminal or another parent application.

**Service Label:** `com.github.dastrobu.apple-mail-mcp`

### Bash Completion

The server supports bash completion for commands and flags:

```bash
# Generate completion script
./apple-mail-mcp completion bash > /usr/local/etc/bash_completion.d/apple-mail-mcp

# Or source directly
source <(./apple-mail-mcp completion bash)
```

**Implementation:**
- `internal/completion/bash.go` - Generates bash completion script
- `internal/opts/typed_flags/*.go` - Typed flags implement `flags.Completer`
- go-flags automatically provides completion via `GO_FLAGS_COMPLETION=1` env var

**Testing completion:**
```bash
GO_FLAGS_COMPLETION=1 ./apple-mail-mcp --transport=
# Output: --transport=http, --transport=stdio

GO_FLAGS_COMPLETION=1 ./apple-mail-mcp launchd ""
# Output: create, remove
```

**Note:** Zsh completion is not currently supported due to compatibility issues with the go-flags completion system.

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
    
    // Check if Mail.app is running (REQUIRED for all scripts)
    if (!Mail.running()) {
        return JSON.stringify({
            success: false,
            error: 'Mail.app is not running. Please start Mail.app and try again.',
            errorCode: 'MAIL_APP_NOT_AVAILABLE'
        });
    }
    
    // Collect logs instead of using console.log
    const logs = [];
    
    // Helper function to log messages
    function log(message) {
        logs.push(message);
    }
    
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
        
        // Return success with data wrapped in 'data' field and logs at top level
        return JSON.stringify({
            success: true,
            data: result,
            logs: logs.join("\n")
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

#### Mail.app Availability Check

**CRITICAL**: All JXA scripts MUST check if Mail.app is running before any operations:

```javascript
// Check if Mail.app is running (REQUIRED for all scripts)
if (!Mail.running()) {
    return JSON.stringify({
        success: false,
        error: 'Mail.app is not running. Please start Mail.app and try again.',
        errorCode: 'MAIL_APP_NOT_RUNNING'
    });
}

// All Mail.app operations should be wrapped in try-catch
try {
    // Access Mail.app (e.g., Mail.accounts(), Mail.mailboxes())
    const accounts = Mail.accounts();
    // ... do work ...
    
    return JSON.stringify({
        success: true,
        data: { /* results */ },
        logs: logs.join("\n")
    });
} catch (e) {
    // If Mail.app is running but we can't access it, it's a permissions issue
    // (macOS returns generic "Error: An error occurred." for permission denials)
    return JSON.stringify({
        success: false,
        error: 'Permission denied to access Mail.app. Please grant automation permissions in System Settings > Privacy & Security > Automation.',
        errorCode: 'MAIL_APP_NO_PERMISSIONS'
    });
}
```

**Why this is required:**
- Server can start without Mail.app running (important for launchd)
- Provides consistent error handling across all tools
- Distinguishes between "not running" and "no permissions" cases
- Returns user-friendly error messages to the LLM

**Error Code Handling:**
The `jxa.Execute()` function in Go detects these error codes:
- `MAIL_APP_NOT_RUNNING` - Returns: "Mail.app is not running. Please start Mail.app and try again"
- `MAIL_APP_NO_PERMISSIONS` - Returns: "Mail.app automation permission denied. Please grant permission in System Settings > Privacy & Security > Automation"

#### Error Handling and Logging

- **Check Mail.app running status FIRST** (before any other operations)
- Validate all arguments before try-catch
- Always wrap operations in try-catch
- Return structured JSON: `{success: bool, data/error: ..., errorCode?: string}`
- Include descriptive error messages
- **NEVER use console.log()** - use the log() helper function instead
- **NEVER ignore errors silently** - always log errors using the log() helper:
  ```javascript
  // ❌ WRONG - Silent error
  try {
    const value = obj.property();
  } catch (e) {
    // No recipients
  }
  
  // ❌ WRONG - Using console.log
  try {
    const value = obj.property();
  } catch (e) {
    console.log("Error reading property: " + e.toString());
  }
  
  // ✅ CORRECT - Use log() helper
  try {
    const value = obj.property();
  } catch (e) {
    log("Error reading property: " + e.toString());
  }
  ```

**Logging Pattern:**
All JXA scripts MUST use the following logging pattern:
```javascript
function run(argv) {
    const Mail = Application('Mail');
    Mail.includeStandardAdditions = true;
    
    // Collect logs instead of using console.log
    const logs = [];
    
    // Helper function to log messages
    function log(message) {
        logs.push(message);
    }
    
    // ... rest of script
    
    try {
        // ... operations
        return JSON.stringify({
            success: true,
            data: result,
            logs: logs.join("\n")  // Always include logs at top level
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
- Create `logs` array at top of run() function
- Define `log(message)` helper that appends message to logs array
- Use `log()` instead of `console.log()` throughout the script
- Always include `logs: logs.join("\n")` as a top-level property (alongside `success` and `data`)
- Logs are returned to Go layer as a single string with newline separators

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
- Detects `errorCode: 'MAIL_APP_NOT_AVAILABLE'` and returns user-friendly error
- Extracts and returns the `data` field (REQUIRED in all scripts)
- Returns descriptive errors for any failures

**Error Codes:**
Scripts can return error codes in the JSON response:
```javascript
return JSON.stringify({
    success: false,
    error: 'Descriptive error message',
    errorCode: 'ERROR_CODE_HERE'
});
```

Currently supported error codes:
- `MAIL_APP_NOT_RUNNING` - Mail.app is not running (detected by `Mail.running()` check)
- `MAIL_APP_NO_PERMISSIONS` - Mail.app is running but permission denied to access it

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
go build -o apple-mail-mcp

# Clean
make clean

# Format code
gofmt -w .

# Vet code
go vet ./...

# Run tests
go test -v -count=1 ./...

# Test server starts (important after any changes!)
./apple-mail-mcp
# Should start without panicking. Press Ctrl+C to stop.
```

### Testing After Changes

**CRITICAL**: Always test that the server starts after making changes:

```bash
# Rebuild and test
go build -o apple-mail-mcp . && ./apple-mail-mcp &
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
go build -o apple-mail-mcp . && timeout 2s ./apple-mail-mcp
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
4. **Dual Transport Support**: HTTP transport is recommended (permissions go to binary) vs STDIO (permissions go to parent process)
5. **No External State**: Keep server stateless for simplicity

## Security Considerations

- All operations are read-only by design
- No sending emails or modifying messages
- Server runs locally on user's machine
- Mail.app's security and permissions apply
- **HTTP mode recommended**: Automation permissions are granted to the `apple-mail-mcp` binary itself
- **STDIO mode caveat**: Automation permissions are granted to the parent process (Terminal, Claude Desktop, etc.)
- Never log sensitive email content

## Launchd Service Management

The `internal/launchd` package provides programmatic launchd management, split into focused modules:

**Constants (common.go):**
- `Label` - Service identifier: `com.github.dastrobu.apple-mail-mcp`
- `DefaultPort` - Default HTTP port: `8787`
- `DefaultHost` - Default HTTP host: `localhost`
- `DefaultLogPath` - Default log path: `~/Library/Logs/com.github.dastrobu.apple-mail-mcp/apple-mail-mcp.log`
- `DefaultErrPath` - Default error log path: `~/Library/Logs/com.github.dastrobu.apple-mail-mcp/apple-mail-mcp.err`

**Key Functions:**
- `Create(cfg *Config)` - Creates plist, loads service (`create.go`)
  - Supports debug flag via `cfg.Debug` field
  - Supports RunAtLoad flag via `cfg.RunAtLoad` field (default: `true`)
  - If unsuccessful, prints hint about enabling debug logging
- `Remove()` - Unloads service, removes plist (`remove.go`)
- `IsLoaded()` - Checks if service is running (`common.go`)
- `DefaultConfig()` - Returns config with executable path (`create.go`)
  - Uses `which apple-mail-mcp` first to find symlinked path (e.g., `/opt/homebrew/bin/apple-mail-mcp`)
  - Falls back to `os.Executable()` if `which` fails
  - This ensures Homebrew upgrades work correctly (symlink updates, not version-specific Cellar path)
- `PlistPath()` - Returns path to plist file (`common.go`)

**Usage Pattern:**
**Implementation Pattern:**
```go
cfg, err := launchd.DefaultConfig()
if err != nil {
    return err
}

// Override defaults from command-line flags
cfg.Port = options.Port
cfg.Host = options.Host
cfg.Debug = options.Debug
if options.Launchd.Create.DisableRunAtLoad {
    cfg.RunAtLoad = false
}

// Create the service
return launchd.Create(cfg)
```

**Command Structure:**
- Main command: `launchd`
- Subcommands: `create`, `remove`
- Examples: 
  - `./apple-mail-mcp launchd create` - Create service with automatic startup on login
  - `./apple-mail-mcp launchd create --disable-run-at-load` - Create service without automatic startup
  - `./apple-mail-mcp launchd remove` - Remove service

**Files:**
- `internal/launchd/common.go` - Shared constants and utility functions
- `internal/launchd/create.go` - Service creation logic (`Create()`), embeds plist template
- `internal/launchd/templates/launchd.plist.tmpl` - Launchd plist template (embedded)
- `internal/launchd/remove.go` - Service removal logic (`Remove()`)
- `main.go` - Command handling (`createLaunchd()`, `removeLaunchd()`)
- `internal/opts/opts.go` - Command and subcommand parsing
- `internal/opts/typed_flags/` - Typed flag implementations
- `internal/completion/bash.go` - Bash completion generation

**Template Embedding:**
```go
// create.go
//go:embed templates/launchd.plist.tmpl
var plistTemplate string
```

The plist template is embedded at compile time from `internal/launchd/templates/launchd.plist.tmpl`, ensuring the binary is self-contained.

**Debug Flag:**
The template conditionally includes the `--debug` flag based on `cfg.Debug`:
- When enabled: `<string>--debug</string>` is added to ProgramArguments
- When disabled: A commented hint is included showing how to enable it

**RunAtLoad Configuration:**
The template conditionally includes the `RunAtLoad` key based on `cfg.RunAtLoad`:
- When `true` (default): Service starts automatically on login
- When `false` (via `--disable-run-at-load`): Service must be started manually via `launchctl start` or `launchctl kickstart`
- This gives users control over automatic startup behavior

**Log Directory:**
The service creates `~/Library/Logs/com.github.dastrobu.apple-mail-mcp/` directory automatically and writes logs there (standard macOS location for application logs).

**Error Messages:**
All error messages in launchd package start with emojis (❌ for errors, ⚠️ for warnings) for better visibility.

**Homebrew Integration:**
The `.goreleaser.yaml` includes a `post_install` script that automatically recreates the launchd service after `brew upgrade`, preserving all user settings:
```ruby
post_install: |
  # Check if launchd service exists and recreate it after upgrade
  # This preserves all settings (port, host, debug, RunAtLoad) from the existing plist
  plist_path = "#{ENV["HOME"]}/Library/LaunchAgents/com.github.dastrobu.apple-mail-mcp.plist"
  if File.exist?(plist_path)
    # Read the existing plist to extract settings
    plist_content = File.read(plist_path)
    
    # Extract port, host, debug, and RunAtLoad settings from existing plist
    port = plist_content.match(/--port=(\d+)/)
    port_flag = port ? "--port=#{port[1]}" : ""
    
    host = plist_content.match(/--host=([^\s<]+)/)
    host_flag = host ? "--host=#{host[1]}" : ""
    
    has_debug = plist_content.include?("<string>--debug</string>")
    debug_flag = has_debug ? "--debug" : ""
    
    has_run_at_load = plist_content.include?("<key>RunAtLoad</key>")
    run_at_load_flag = has_run_at_load ? "" : "--disable-run-at-load"
    
    # Recreate the service preserving all settings
    cmd = ["#{bin}/apple-mail-mcp"]
    cmd << port_flag unless port_flag.empty?
    cmd << host_flag unless host_flag.empty?
    cmd << debug_flag unless debug_flag.empty?
    cmd << "launchd"
    cmd << "create"
    cmd << run_at_load_flag unless run_at_load_flag.empty?
    
    system(*cmd.reject(&:empty?))
    
    ohai "Recreated launchd service with updated binary (preserved existing settings)"
  end
```

This ensures that:
- After `brew upgrade apple-mail-mcp`, the service is automatically recreated with the new version
- All user settings are preserved: port, host, debug flag, and RunAtLoad setting
- Users don't need to manually run `apple-mail-mcp launchd create` again
- Plist template updates are applied (if any)
- The new binary path is used (via the updated symlink)

**Uninstall Process:**
Users should remove the launchd service BEFORE uninstalling via Homebrew:
```bash
# Step 1: Remove the launchd service
apple-mail-mcp launchd remove

# Step 2: Uninstall the package
brew uninstall apple-mail-mcp

# Step 3 (optional): Remove logs
rm -rf ~/Library/Logs/com.github.dastrobu.apple-mail-mcp/
```

**Why this order matters:** The `launchd remove` command needs the binary to properly unload and remove the service. If the binary is uninstalled first, users must manually clean up:
```bash
# Manual cleanup if binary was already removed
launchctl unload ~/Library/LaunchAgents/com.github.dastrobu.apple-mail-mcp.plist
rm ~/Library/LaunchAgents/com.github.dastrobu.apple-mail-mcp.plist
```

**Note:** GoReleaser doesn't support `uninstall` hooks in the `brews` configuration, so uninstall instructions are provided in the caveats and README.

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
8. **Always test server starts** - run `./apple-mail-mcp` after changes to catch schema errors
9. **Check script success field** before processing data
10. **Type assertions need checking**: `val, ok := x.(string)`
11. **JSON numbers are float64** - convert to int as needed
12. **Context cancellation** - always pass and respect ctx
13. **Use mailboxPath arrays** - not plain mailbox names (supports nested mailboxes)
14. **Dereference Object Specifiers** - always use `()` to get values: `mailbox.name()` not `mailbox.name`
15. **Use whose() for filtering** - ~150x faster than loops, constant-time performance
16. **NEVER use console.log()** - use the `log()` helper function and include `logs` as top-level property

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

## Nested Mailbox Support

All tools must support hierarchical mailboxes (e.g., `Inbox > GitHub`).

### Mailbox Path Representation

Use JSON arrays for mailbox paths:
- Top-level: `["Inbox"]`
- Nested: `["Inbox", "GitHub"]`
- Deeply nested: `["Archive", "2024", "Q1"]`

### Building Mailbox Paths (JXA)

```javascript
function getMailboxPath(mailbox, accountName) {
  const path = [];
  let current = mailbox;

  while (current) {
    const name = current.name();
    if (name === accountName) break;
    path.unshift(name);
    try {
      current = current.container();
    } catch (e) {
      break;
    }
  }
  return path;
}
```

### Navigating to Nested Mailboxes (JXA)

```javascript
// Parse mailboxPath from JSON
const mailboxPath = JSON.parse(mailboxPathStr);

// Navigate using chained lookups
let targetMailbox = account.mailboxes[mailboxPath[0]];
for (let i = 1; i < mailboxPath.length; i++) {
  targetMailbox = targetMailbox.mailboxes[mailboxPath[i]];
}
```

### Go Integration

**Input Type:**
```go
type ToolInput struct {
    Account     string   `json:"account" jsonschema:"Account name"`
    MailboxPath []string `json:"mailboxPath" jsonschema:"Mailbox path array"`
    ID          int      `json:"id" jsonschema:"Message ID"`
}
```

**Execution:**
```go
mailboxPathJSON, err := json.Marshal(input.MailboxPath)
if err != nil {
    return nil, nil, fmt.Errorf("failed to marshal mailbox path: %w", err)
}

data, err := jxa.Execute(ctx, script,
    input.Account,
    string(mailboxPathJSON),
    fmt.Sprintf("%d", input.ID))
```

## JXA Best Practices

Based on [JXA documentation](https://github.com/JXA-Cookbook/JXA-Cookbook) and extensive testing:

### Object Specifier Dereferencing

**Always** call properties as functions to get values:

```javascript
// ❌ Wrong - returns Object Specifier
const name = mailbox.name;

// ✅ Correct - returns JavaScript string
const name = mailbox.name();
```

**Exception:** `.length` works on both Object Specifiers and arrays.

### Name Lookup Syntax

```javascript
// ✅ Direct name lookup (preferred)
const inbox = account.mailboxes["Inbox"];
const github = inbox.mailboxes["GitHub"];

// ❌ Don't loop for name-based lookup
for (let i = 0; i < mailboxes.length; i++) {
  if (mailboxes[i].name() === "Inbox") { /* inefficient */ }
}
```

### Use whose() for Filtering

**Performance:** ~150x faster than loops, constant-time O(1)

```javascript
// ✅ Fast (constant time ~0.3ms)
const matches = mailbox.messages.whose({ id: messageId })();

// ❌ Slow (linear time)
const messages = mailbox.messages();
for (let i = 0; i < messages.length; i++) {
  if (messages[i].id() === messageId) { /* 150x slower */ }
}
```

### whose() Performance Numbers

- `whose()`: 0.00064 ms/record (constant time)
- JavaScript filter: 0.1 ms/record (150x slower)
- Traditional loop: 2 ms/record (3000x slower)

For 500 records: `whose()` = 0.32ms vs loops = 1000ms

### Convert Elements to Arrays

```javascript
// ❌ This fails - not an array
const acc = app.accounts;
acc.forEach(a => console.log(a.name()));

// ✅ Dereference first
const acc = app.accounts();
acc.forEach(a => console.log(a.name()));
```

### Modern JavaScript

```javascript
// ✅ Use const/let, template literals, arrow functions
const Mail = Application("Mail");
const error = `Message ${id} not found in ${mailbox}`;
messages.forEach(msg => console.log(msg.subject()));
```

### Enumeration Property Filtering

```javascript
// ❌ Fails with "Types cannot be converted"
app.accounts.whose({ authentication: "password" })();

// ✅ Use _match with ObjectSpecifier()
app.accounts.whose({
  _match: [ObjectSpecifier().authentication, "password"]
})();
```

## Rich Text Support

### Principles

1. **Business logic in Go, not JXA**: All Markdown parsing, style application, and margin calculations happen in Go
2. **JXA only renders**: JXA receives fully-processed styled blocks and just creates paragraphs
3. **Rune-based positioning**: All inline style positions use rune (Unicode code point) offsets, not byte offsets
4. **Margin blocks**: Margins are represented as separate "margin" type blocks created in Go
5. **Character-level color styling**: All colors are applied at character level (via InlineStyles) for consistent dark mode behavior
6. **Newline splitting**: Code blocks, blockquotes, and list items are split on newlines to avoid Mail.app paragraph splitting breaking inline styles

### UTF-8 Character Handling

**CRITICAL**: Always use rune length, never byte length for character positions:

```go
// ❌ WRONG - byte length
currentPos += len(text)
prefixLen := len(prefix)

// ✅ CORRECT - rune length
currentPos += len([]rune(text))
prefixLen := len([]rune(prefix))
```

**Why**: JXA character arrays use Unicode code point indexing. Multibyte UTF-8 characters (•, emoji, etc.) are multiple bytes but single runes.

### Margin Handling

Margins are handled entirely in Go:

```go
func convertHeading(node *ast.Heading, source []byte, config *RenderingConfig, isFirst bool) ([]StyledBlock, error) {
    var blocks []StyledBlock
    
    // Add margin_top block (skip for first block)
    if !isFirst && style.MarginTop > 0 {
        blocks = append(blocks, StyledBlock{
            Type: "margin",
            Text: "\n",
            Font: style.Font,
            Size: style.MarginTop,
        })
    }
    
    // Add the heading block
    blocks = append(blocks, StyledBlock{...})
    
    // Add margin_bottom block
    if style.MarginBottom > 0 {
        blocks = append(blocks, StyledBlock{
            Type: "margin",
            Text: "\n",
            Font: style.Font,
            Size: style.MarginBottom,
        })
    }
    
    return blocks, nil
}
```

**JXA renders ALL blocks as paragraphs** (simplified):
```javascript
// All blocks rendered the same way - Go adds newlines to block.text
const props = {};
if (block.font) {
    props.font = block.font;
}
if (block.size) {
    props.size = block.size;
}
if (block.color) {
    props.color = block.color;
}

Mail.make({
    new: "paragraph",
    withData: block.text,  // Already includes newline from Go
    withProperties: props,
    at: msg.content,
});
```

**Key Points:**
- No type checking needed - render everything as paragraph
- Font, size, and color are optional (only set if present)
- Newlines handled in Go (appended to `block.text`)
- Margin blocks just have different size, rendered same way

### Character-Level Color Styling and Paragraph Splitting

**CRITICAL**: Mail.app automatically splits text on embedded newlines when creating paragraphs, which breaks absolute character positioning for inline styles.

**Solution**: Always split blocks on newlines before sending to JXA.

**Color Application**:
- All colors MUST be applied at character level via `InlineStyles`, never at paragraph level via `Color` property
- Mail.app auto-adapts character-level colors in dark mode but NOT paragraph-level colors
- This ensures consistent color behavior across light and dark modes

**Implementation Pattern**:
```go
// For code blocks, blockquotes, and any block with potential newlines
text := buf.String()
lines := strings.Split(text, "\n")

for i, line := range lines {
    lineText := line + "\n"
    lineRuneCount := len([]rune(line))
    
    var lineInlineStyles []InlineStyle
    
    // Apply color as character-level style covering entire line
    if style.Color != nil {
        colorStyle := InlineStyle{
            Start: 0,
            End:   lineRuneCount,
            Color: style.Color,
        }
        lineInlineStyles = append(lineInlineStyles, colorStyle)
    }
    
    // Add inline styles from content (bold, italic, etc.)
    // Adjust positions for this line...
    
    blocks = append(blocks, StyledBlock{
        Type:         BlockTypeCodeBlock,
        Text:         lineText,
        Font:         safeString(style.Font),
        Size:         safeInt(style.Size),
        InlineStyles: lineInlineStyles,
    })
}
```

**Key Points**:
- Split on `\n` to create one styled block per line
- Each line gets `lineText = line + "\n"` (add back the newline)
- Character-level color covers `0` to `lineRuneCount` (excludes trailing newline)
- When preserving existing inline styles (like in blockquotes), adjust positions for each line
- Font and size can be set at paragraph level
- Color MUST be set at character level (via InlineStyles)

**Why This Works**:
- One styled block per line = no embedded newlines
- Mail.app doesn't split the paragraph
- Inline style positions remain accurate
- Character-level colors get dark mode adaptation automatically

**Applied To**:
- Code blocks (split by line)
- Blockquotes (split by line, preserve inline styles from content)
- List items (already split by line)
- Regular paragraphs (split on hard line breaks `\n`)

**DON'T**:
- Set `Color` property at paragraph level for any block with inline styles
- Include newlines in blocks without splitting
- Use paragraph-level color for consistency - always use character-level

### Block Type Constants

Block types are defined for documentation/debugging but JXA doesn't check them:

```go
const (
    BlockTypeParagraph      = "paragraph"
    BlockTypeHeading        = "heading"
    BlockTypeCodeBlock      = "code_block"
    BlockTypeBlockquote     = "blockquote"
    BlockTypeListItem       = "list_item"
    BlockTypeHorizontalRule = "horizontal_rule"
)
```

**Note**: Margin blocks have empty Type field - they're distinguished by having only size and newline text.

### Styled Block Structure

```go
type StyledBlock struct {
    Type         string        `json:"type"`           // "paragraph", "heading", "code_block", etc.
    Text         string        `json:"text"`           // Text content (includes trailing \n)
    Font         string        `json:"font,omitempty"` // Optional - paragraph level
    Size         int           `json:"size,omitempty"` // Optional - paragraph level
    Color        *AppleRGB     `json:"color,omitempty"` // DEPRECATED - use InlineStyles instead
    InlineStyles []InlineStyle `json:"inline_styles,omitempty"` // Character-level styling
    Level        int           `json:"level,omitempty"` // For headings and list nesting
}
```

**Key Points:**
- `Text` always includes trailing `\n` (added in Go, not JXA)
- Font and Size are applied at paragraph level
- **Color is DEPRECATED** - always use InlineStyles for color to ensure consistent dark mode behavior
- InlineStyles apply formatting at character level (including color)
- Margin blocks have empty Type, just size and `"\n"` text
- No `MarginTop`/`MarginBottom` fields - margins are separate blocks
- For blocks with newlines (code blocks, blockquotes), split into one block per line
**Example usage**:
```go
return StyledBlock{
    Type: BlockTypeHeading,  // ✅ Use constant
    Text: text,
    // ...
}

// ❌ Don't use magic strings
return StyledBlock{
    Type: "heading",
    // ...
}
```

**Note**: MarginTop/MarginBottom are set on blocks but margins are rendered as separate "margin" type blocks by the Go converter.

### Content Format Enum

Use constants for content format validation:

```go
const (
    ContentFormatPlain    = "plain"
    ContentFormatMarkdown = "markdown"
    ContentFormatDefault  = ContentFormatMarkdown // Default is markdown
)

// Use switch for format handling
switch contentFormat {
case ContentFormatMarkdown:
    // Parse and render
case ContentFormatPlain:
    // Plain text
default:
    return nil, nil, fmt.Errorf("invalid content_format")
}
```

**Never panic on user input** - return errors instead.

## Questions?

When implementing features:
1. Check existing patterns in the codebase
2. Refer to MCP Go SDK documentation
3. Follow Go best practices
4. Keep it simple and maintainable
5. Test with Mail.app running
6. See `docs/MAIL_EDITING.md` for detailed JXA patterns
7. See `internal/tools/scripts/NESTED_MAILBOX_SUPPORT.md` for mailbox handling
8. See `docs/RICH_TEXT_DESIGN.md` for rich text architecture and styling
