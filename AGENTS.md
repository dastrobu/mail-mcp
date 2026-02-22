# GitHub Copilot Instructions for Apple Mail MCP Server

This document provides context and guidelines for GitHub Copilot when working on this MCP server project. To keep this document concise, detailed specifications have been split into targeted reference files.

## Reference Documentation

Always refer to these documents for specific subsystem details:
- **JXA & Mail Interaction:** [`docs/JXA_QUICK_REFERENCE.md`](docs/JXA_QUICK_REFERENCE.md) and [`docs/MAIL_EDITING.md`](docs/MAIL_EDITING.md)
- **Nested Mailbox Support:** [`internal/tools/scripts/NESTED_MAILBOX_SUPPORT.md`](internal/tools/scripts/NESTED_MAILBOX_SUPPORT.md)
- **Rich Text & Styling:** [`docs/RICH_TEXT_DESIGN.md`](docs/RICH_TEXT_DESIGN.md)
- **Launchd Management:** [`docs/LAUNCHD_SERVICE_MANAGEMENT.md`](docs/LAUNCHD_SERVICE_MANAGEMENT.md)

## Project Overview & Architecture

An MCP server providing programmatic, read-only access to macOS Mail.app.
- **Technologies:** Go 1.26+, MCP Go SDK v1.2.0+, JXA (JavaScript for Automation).
- **Core Principles:** Single binary (`//go:embed` for scripts), dual transport (HTTP/STDIO, HTTP recommended for permissions), modular design (one file per tool in `internal/tools/`).

## Tool Implementation Guidelines

### 1. Go Registration Pattern
Each tool lives in `internal/tools/tool_name.go`.

```go
//go:embed scripts/example.js
var exampleScript string

type ExampleInput struct {
    // CRITICAL: jsonschema must be a PLAIN string description. NO key=value pairs!
    Mailbox string `json:"mailbox" jsonschema:"Name of the mailbox to access"`
}

func RegisterExampleTool(srv *mcp.Server) {
    mcp.AddTool(srv,
        &mcp.Tool{
            Name: "example_tool",
            Description: "...",
            Annotations: &mcp.ToolAnnotations{
                Title:           "Example Tool",
                ReadOnlyHint:    true,
                IdempotentHint:  true,
                DestructiveHint: new(false), // Go 1.26+ new() syntax
                OpenWorldHint:   new(true),  // Interacts with external Mail.app
            },
        },
        handleExampleTool,
    )
}

func handleExampleTool(ctx context.Context, req *mcp.CallToolRequest, input ExampleInput) (*mcp.CallToolResult, any, error) {
    // Only apply defaults for optional parameters here. Validation belongs in JXA.
    data, err := jxa.Execute(ctx, exampleScript, input.Mailbox)
    // ALWAYS return generic types (map[string]any or slices) for the data output, not custom structs.
    return nil, data, err 
}
```

### 2. JXA Script Structure
All JXA scripts (`internal/tools/scripts/*.js`) MUST strictly follow this pattern:

```javascript
function run(argv) {
    const Mail = Application('Mail');
    Mail.includeStandardAdditions = true;
    
    // 1. CRITICAL: Check if running FIRST
    if (!Mail.running()) {
        return JSON.stringify({
            success: false,
            error: 'Mail.app is not running...',
            errorCode: 'MAIL_APP_NOT_RUNNING'
        });
    }
    
    // 2. Logging setup (NEVER use console.log)
    const logs = [];
    function log(msg) { logs.push(msg); }
    
    // 3. Argument parsing & strict validation
    const mailbox = argv[0] || '';
    if (!mailbox) return JSON.stringify({ success: false, error: 'Mailbox required' });
    
    // 4. Execution wrapped in try/catch
    try {
        const result = doSomething(mailbox);
        
        // 5. CRITICAL: Always wrap output in 'data' field alongside success and logs
        return JSON.stringify({
            success: true,
            data: result,
            logs: logs.join("\n")
        });
    } catch (e) {
        // Return errorCode: 'MAIL_APP_NO_PERMISSIONS' for automation permission denials
        return JSON.stringify({ success: false, error: e.toString() });
    }
}
```

### 3. JXA Best Practices (Quick Summary)

- **Dereference Properties:** Always use `()` to get values: `mailbox.name()` NOT `mailbox.name`. Exception: `.length`.
- **Use `whose()` for Filtering:** Always use `whose({ id: messageId })()` instead of JS `for`/`filter`. It is O(1) constant time and ~150x faster.
- **Nested Mailboxes:** Always accept and traverse mailbox paths as arrays `["Account", "Inbox", "Subfolder"]`. See `NESTED_MAILBOX_SUPPORT.md`.
- **Modern JS:** Use `const`, `let`, template literals, and arrow functions.
- **Dates:** Always convert to ISO: `msg.dateReceived().toISOString()`.

## Build & Testing

Always ensure the server builds and runs cleanly without panicking after changes. Schema parsing errors will trigger panics on startup.

```bash
# Verify it builds and starts correctly
go build -o apple-mail-mcp . && timeout 2s ./apple-mail-mcp

# Run tests
go test -v -count=1 ./...
```

## Common Pitfalls Checklist

1.  [ ] Did you use `//go:embed` for the script?
2.  [ ] Does the JXA script verify `Mail.running()` first?
3.  [ ] Did you wrap the final JXA output payload inside a `data` field? (e.g. `data: result`)
4.  [ ] Is your Go return type for tools `map[string]any` or `[]map[string]any` instead of a typed struct?
5.  [ ] Are your `jsonschema` tags plain strings without `key=value` parameters?
6.  [ ] Are you dereferencing AppleScript object specifiers with `()`?
7.  [ ] Did you use `whose()` for finding messages rather than iterating arrays?
8.  [ ] Are you using `log()` instead of `console.log()`?
9.  [ ] Did you ignore any errors silently?
