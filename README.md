# Apple Mail MCP Server

A Model Context Protocol (MCP) server providing programmatic access to macOS Mail.app using Go and JavaScript for Automation (JXA).

## Overview

This MCP server enables AI assistants and other MCP clients to interact with Apple Mail on macOS. It provides read-only access to mailboxes, messages, and search functionality through a clean, typed interface.

## Features

- **List Accounts**: Enumerate all configured email accounts with their properties
- **List Mailboxes**: Enumerate all available mailboxes and accounts
- **Get Message Content**: Fetch detailed content of individual messages
- **Get Selected Messages**: Retrieve currently selected message(s) in Mail.app
- **Reply to Message**: Create a reply to a message and save it as a draft

## Requirements

- macOS (Mail.app is macOS-only)
- Go 1.25 or later
- Mail.app must be running and configured with at least one email account
- Automation permissions for the terminal/application running the server (see [Troubleshooting](#troubleshooting))

## Installation

### From Source

```bash
git clone https://github.com/dastrobu/apple-mail-mcp.git
cd apple-mail-mcp
go build -o apple-mail-mcp
```

## Usage

The server supports two transport modes: STDIO (default) and HTTP.

### STDIO Transport (Default)

For use with MCP clients like Claude Desktop:

```bash
./apple-mail-mcp
```

Or explicitly:

```bash
./apple-mail-mcp --transport=stdio
```

### HTTP Transport

For web-based clients or development:

```bash
# Start HTTP server on default port 8787
./apple-mail-mcp --transport=http

# Start on custom port
./apple-mail-mcp --transport=http --port=3000

# Start on custom host and port
./apple-mail-mcp --transport=http --host=0.0.0.0 --port=3000
```

The HTTP server provides a streamable HTTP transport compatible with MCP clients.

### Command-Line Options

```
--transport=[stdio|http]  Transport type (default: stdio)
--port=PORT              HTTP port (default: 8787, only used with --transport=http)
--host=HOST              HTTP host (default: localhost, only used with --transport=http)
--debug                  Enable debug logging of tool calls and results to stderr
--help                   Show help message
```

Options can also be set via environment variables:
```
TRANSPORT=http
PORT=8787
HOST=localhost
DEBUG=true
```

Command-line flags take precedence over environment variables. You can also use a `.env` file for local development.

#### Debug Mode

When `--debug` is enabled, the server logs all MCP protocol interactions to stderr, including:
- **Initialize requests**: Client capabilities and initialization parameters
- **Tools/list requests**: When the client requests the list of available tools
- **Tool calls**: Input parameters for each tool invocation
- **Tool results**: Output data or errors from tool execution

This is useful for troubleshooting and understanding what data the MCP client is requesting:

```bash
./apple-mail-mcp --debug
```

Example debug output:
```
[DEBUG] MCP Request: initialize
Params: {
  "capabilities": {
    "roots": {
      "listChanged": true
    }
  },
  "clientInfo": {
    "name": "claude-desktop",
    "version": "1.0.0"
  },
  "protocolVersion": "2024-11-05"
}
[DEBUG] MCP Response: initialize
Result: {
  "capabilities": {
    "tools": {}
  },
  "protocolVersion": "2024-11-05",
  "serverInfo": {
    "name": "apple-mail",
    "version": "0.1.0"
  }
}
[DEBUG] MCP Request: tools/list
[DEBUG] MCP Response: tools/list
Result: {
  "tools": [
    {
      "name": "list_accounts",
      "description": "Lists all configured email accounts..."
    },
    ...
  ]
}
[DEBUG] MCP Request: tools/call
Params: {
  "name": "list_accounts",
  "arguments": {
    "enabled": true
  }
}
[DEBUG] MCP Response: tools/call
Result: {
  "content": [
    {
      "type": "text",
      "text": "{\"accounts\":[{\"name\":\"Work\",\"enabled\":true}],\"count\":1}"
    }
  ]
}
```

Since the server uses STDIO for MCP communication (on stdout), debug logs are written to stderr and won't interfere with the protocol.

On first run, the server performs a connectivity check to verify Mail.app is accessible. If this fails, see the [Troubleshooting](#troubleshooting) section below.

### Configuration

#### For STDIO Transport (Claude Desktop)

Add to your MCP client configuration (e.g., Claude Desktop):

```json
{
  "mcpServers": {
    "apple-mail": {
      "command": "/path/to/apple-mail-mcp"
    }
  }
}
```

#### For HTTP Transport

Configure your MCP client to connect to:
```
http://localhost:8787
```

Or the custom host/port you specified with `--port` flag or `PORT` environment variable.

## Troubleshooting

### Automation Permission Errors

If you see an error like:
```
Mail.app startup check failed: osascript execution failed: signal: killed
```

This means macOS is blocking the server from controlling Mail.app. You need to grant automation permissions:

#### Option 1: Using System Settings (Recommended)

1. Open **System Settings** (or **System Preferences** on older macOS)
2. Go to **Privacy & Security** → **Automation**
3. Find the application running the server in the list:
   - For Terminal: Look for **Terminal** or **iTerm**
   - For Claude Desktop: Look for **Claude**
   - For other apps: Look for the parent application name
4. Enable the checkbox next to **Mail** for that application
5. Restart the MCP server

#### Option 2: Triggering the Permission Dialog

If the app doesn't appear in Automation settings:

1. Run the server manually from Terminal to trigger the permission prompt:
   ```bash
   ./apple-mail-mcp
   ```
2. macOS should show a dialog asking: "Terminal would like to control Mail.app"
3. Click **OK** to grant permission
4. The server should now work

#### Option 3: Using tccutil (Advanced)

For automation or CI/CD scenarios, you can use `tccutil` to reset permissions:

```bash
# Reset automation permissions (forces a new prompt)
tccutil reset AppleEvents

# Then run the server to trigger the permission dialog
./apple-mail-mcp
```

**Note**: After granting permissions, you may need to restart the application running the MCP server (e.g., Claude Desktop, Terminal) for changes to take effect.

### Mail.app Not Running

If you see:
```
Mail.app is not running. Please start Mail.app and try again.
```

Simply open Mail.app and try again. Mail.app must be running for the server to work.

### Server Exits Immediately

The server performs a startup check to verify Mail.app is accessible. If the check fails, the server will exit with an error message. Common causes:

1. Mail.app is not running → Start Mail.app
2. Missing automation permissions → See [Automation Permission Errors](#automation-permission-errors) above
3. Mail.app is starting up → Wait a few seconds and try again

## Examples

### Testing HTTP Transport

You can test the HTTP transport using curl:

```bash
# Start the server (uses default port 8787)
./apple-mail-mcp --transport=http

# Or use environment variable
TRANSPORT=http PORT=8787 ./apple-mail-mcp

# In another terminal, test the endpoint
# Note: Proper MCP clients will handle session management
curl -i http://localhost:8787/

# Expected response:
# HTTP/1.1 400 Bad Request
# Bad Request: GET requires an Mcp-Session-Id header
```

For actual MCP communication over HTTP, use an MCP client library that supports the streamable HTTP transport protocol.

### Using with Claude Desktop (STDIO)

The default STDIO transport is designed for use with MCP clients like Claude Desktop. Simply add the configuration as shown in the [Configuration](#configuration) section above.

## Available Tools

### list_accounts

Lists all configured email accounts in Apple Mail.

**Parameters:**
- `enabled` (boolean, optional): Filter to only show enabled accounts (default: false)

**Output:**
- Array of account objects with:
  - `name`: Account name
  - `enabled`: Whether the account is enabled
  - `emailAddresses`: Array of email addresses associated with the account
  - `mailboxCount`: Number of mailboxes in the account
- `count`: Total number of accounts

**Example Output:**
```json
{
  "accounts": [
    {
      "name": "Exchange",
      "enabled": true,
      "emailAddresses": ["user@example.com"],
      "mailboxCount": 22
    }
  ],
  "count": 1
}
```

### list_mailboxes

Lists all available mailboxes across all Mail accounts.

**Output:**
- Array of mailbox objects with name and account information

### get_message_content

Fetches the full content of a specific message including body, headers, recipients, and attachments.

**Parameters:**
- `account` (string, required): Name of the email account
- `mailbox` (string, required): Name of the mailbox (e.g., "INBOX", "Sent")
- `message_id` (integer, required): The unique ID of the message

**Output:**
- Full message object including:
  - Basic fields: id, subject, sender, replyTo
  - Dates: dateReceived, dateSent
  - Content: content (body text), allHeaders
  - Status: readStatus, flaggedStatus
  - Recipients: toRecipients, ccRecipients, bccRecipients (with name and address)
  - Attachments: array of attachment objects with name, fileSize, and downloaded status
  - Note: mimeType is not included for attachments due to Mail.app API limitations

**Error Handling:**
- The tool gracefully handles missing or unavailable fields
- If a field cannot be accessed, it returns a safe default value (empty string, empty array, etc.)
- Clear error messages are provided for common issues:
  - Invalid account or mailbox names
  - Message not found or has been deleted
  - Missing required parameters

### get_selected_messages

Gets the currently selected message(s) in the frontmost Mail.app viewer window.

**Parameters:**
- None (operates on current selection)

**Output:**
- Object containing:
  - `count`: Number of selected messages
  - `messages`: Array of message objects, each with:
    - `id`: Unique message identifier
    - `subject`: Subject line
    - `sender`: Sender email address
    - `dateReceived`: When the message was received (ISO 8601)
    - `dateSent`: When the message was sent (ISO 8601)
    - `readStatus`: Whether the message has been read
    - `flaggedStatus`: Whether the message is flagged
    - `junkMailStatus`: Whether the message is marked as junk
    - `mailbox`: Name of the mailbox containing the message
    - `account`: Name of the account containing the message

**Behavior:**
- Returns empty array if no messages are selected
- Returns error if no Mail viewer windows are open
- Can return multiple messages if multiple are selected
- Selection state is transient and can change between calls

**Example Output:**
```json
{
  "count": 1,
  "messages": [
    {
      "id": 123456,
      "subject": "Meeting Tomorrow",
      "sender": "colleague@example.com",
      "dateReceived": "2024-02-11T10:30:00Z",
      "dateSent": "2024-02-11T10:25:00Z",
      "readStatus": true,
      "flaggedStatus": false,
      "junkMailStatus": false,
      "mailbox": "INBOX",
      "account": "Work"
    }
  ]
}
```

### reply_to_message

Creates a reply to a specific message and saves it as a draft in the Drafts mailbox. Mail.app automatically includes the quoted original message. The reply is NOT sent automatically - it remains in drafts for review and manual sending.

**Parameters:**
- `account` (string, required): Name of the email account
- `mailbox` (string, required): Name of the mailbox containing the message to reply to
- `message_id` (integer, required): The unique ID of the message to reply to
- `reply_content` (string, required): The content/body of the reply message (will be prepended to the automatically quoted original message)
- `opening_window` (boolean, optional): Whether to show the window for the reply message. Default is false.
- `reply_to_all` (boolean, optional): Whether to reply to all recipients. Default is false (reply to sender only).

**Output:**
- Object containing:
  - `draft_id`: ID of the created draft message
  - `subject`: Subject line of the reply (prefixed with "Re: ")
  - `to_recipients`: Array of recipient email addresses
  - `drafts_mailbox`: Name of the Drafts mailbox where the reply was saved
  - `message`: Confirmation message

**Behavior:**
- Creates a reply with "Re: " prefix on the subject
- Sets the recipient to the original message sender (or all recipients if `reply_to_all` is true)
- Mail.app automatically formats and includes the quoted original message
- Your `reply_content` is prepended to the automatic quote
- Saves the reply in the account's Drafts mailbox (not sent)
- Maintains email thread context with proper headers

**Error Handling:**
- Clear error messages for invalid account or mailbox names
- Message not found or has been deleted
- Missing Drafts mailbox
- Missing required parameters

## Architecture

The server is built with:
- **Go**: Main server implementation using the MCP Go SDK
- **JXA (JavaScript for Automation)**: Scripts embedded in the binary for Mail.app interaction
  - See [JXA Documentation](https://developer.apple.com/library/archive/releasenotes/InterapplicationCommunication/RN-JavaScriptForAutomation/Articles/Introduction.html#//apple_ref/doc/uid/TP40014508) for more details
  - See [Mac Automation Scripting Guide](https://developer.apple.com/library/archive/documentation/LanguagesUtilities/Conceptual/MacAutomationScriptingGuide/index.html#//apple_ref/doc/uid/TP40016239-CH56-SW1) for comprehensive automation documentation
- **STDIO Transport**: Simple, stateless communication protocol

All JXA scripts are embedded at compile time using `//go:embed`, making the server a single, self-contained binary.

## Development

### Git Hooks

This project includes a pre-commit hook that automatically runs `go fmt` on all staged Go files before committing. This ensures consistent code formatting across the project.

#### Installing Git Hooks

Run the installation script:

```bash
make install-hooks
```

Or manually:

```bash
./scripts/install-hooks.sh
```

The pre-commit hook will:
- Run `go fmt` on all staged `.go` files
- Automatically stage the formatted files
- Only run if there are staged Go files

#### Manual Formatting

You can also format code manually:

```bash
make fmt
```

Or directly:

```bash
gofmt -w .
```

### Build

```bash
make build
```

### Test JXA Scripts

```bash
make test-scripts
```

### Clean

```bash
make clean
```

## Project Structure

```
apple-mail-mcp/
├── cmd/
│   └── apple-mail-mcp/      # Main application entry point
│       └── main.go
├── internal/
│   ├── jxa/                  # JXA script execution
│   │   └── executor.go
│   └── tools/                # MCP tool implementations
│       ├── scripts/          # Embedded JXA scripts
│       │   ├── list_accounts.js
│       │   ├── list_mailboxes.js
│       │   ├── get_message_content.js
│       │   ├── get_selected_messages.js
│       │   └── reply_to_message.js
│       ├── list_accounts.go
│       ├── list_mailboxes.go
│       ├── get_message_content.go
│       ├── get_selected_messages.go
│       ├── reply_to_message.go
│       └── tools.go          # Tool registration and helpers
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

The project follows standard Go project layout:
- `cmd/` - Main application packages
- `internal/` - Private application code
  - `jxa/` - JXA script execution utilities
  - `tools/` - Individual MCP tool implementations (one file per tool)
  - `tools/scripts/` - JavaScript for Automation scripts embedded into the binary

Each tool is implemented in its own file within `internal/tools/`, making the codebase modular and easy to maintain.

## Error Handling

The server provides detailed error messages to help diagnose issues:

- **Script Errors**: Clear messages indicating what went wrong in JXA scripts
- **Missing Data**: Descriptive errors when expected data is not found
- **Invalid Parameters**: Validation errors with hints about correct usage
- **Argument Context**: Error messages include the arguments passed to help debugging

All tools handle errors gracefully and return informative error messages rather than generic failures.

## Limitations

- **macOS only**: Relies on Mail.app and JXA
- **Mostly read-only**: Only the `reply_to_message` tool creates drafts; no emails are sent automatically
- **Mail.app required**: Mail.app must be running for the server to work
- **Attachment MIME types**: Due to Mail.app API limitations, MIME types are not available for attachments

## Security & Privacy

- All operations are read-only except `reply_to_message` which creates drafts
- Draft replies are not sent automatically - they require manual review and sending
- No data is transmitted outside of the MCP connection
- The server runs locally on your machine
- Mail.app's security and permissions apply

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

MIT License - see LICENSE file for details

## Acknowledgments

Built with the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) from Anthropic.
