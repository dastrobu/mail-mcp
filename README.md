# Apple Mail MCP Server

[![CI](https://github.com/dastrobu/apple-mail-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/dastrobu/apple-mail-mcp/actions/workflows/ci.yaml)

A Model Context Protocol (MCP) server providing programmatic access to macOS Mail.app using Go and JavaScript for Automation (JXA).

## Overview

This MCP server enables AI assistants and other MCP clients to interact with Apple Mail on macOS. It provides read-only access to mailboxes, messages, and search functionality through a clean, typed interface.

## Security & Privacy

- **Human-in-the-loop design**: No emails are sent automatically - all drafts require manual sending. This prevents agents from sending emails without human oversight.
- No data transmitted outside of the MCP connection
- Runs locally on your machine
- Mail.app's security and permissions apply

## Features

- **List Accounts**: Enumerate all configured email accounts with their properties
- **List Mailboxes**: Enumerate all available mailboxes and accounts
- **Get Message Content**: Fetch detailed content of individual messages
- **Get Selected Messages**: Retrieve currently selected message(s) in Mail.app
- **Reply to Message**: Create a reply to a message and save it as a draft
- **Create Outgoing Message**: Create new email drafts with optional Markdown formatting
- **Rich Text Support**: Format emails with Markdown (headings, bold, italic, lists, code blocks, and more)

## Requirements

- macOS (Mail.app is macOS-only)
- Go 1.26 or later
- Mail.app must be running and configured with at least one email account
- Automation permissions for the terminal/application running the server

## Installation

```bash
go build -o apple-mail-mcp
```

## Usage

The server supports two transport modes: STDIO (default) and HTTP.

### STDIO Transport (Default)

```bash
./apple-mail-mcp
```

### HTTP Transport

```bash
# Default port 8787
./apple-mail-mcp --transport=http

# Custom port
./apple-mail-mcp --transport=http --port=3000

# Custom host and port
./apple-mail-mcp --transport=http --host=0.0.0.0 --port=3000
```

### Command-Line Options

```
--transport=[stdio|http]  Transport type (default: stdio)
--port=PORT              HTTP port (default: 8787, only used with --transport=http)
--host=HOST              HTTP host (default: localhost, only used with --transport=http)
--debug                  Enable debug logging of tool calls and results to stderr
--rich-text-styles=PATH  Path to custom rich text styles YAML file (uses embedded default if not specified)
--help                   Show help message
```

Options can also be set via environment variables:
```
TRANSPORT=http
PORT=8787
HOST=localhost
DEBUG=true
RICH_TEXT_STYLES=/path/to/custom_styles.yaml
```

### Debug Mode

When `--debug` is enabled, the server logs all MCP protocol interactions and JXA script diagnostics to stderr, including tool calls, results, and JXA script logs. See [DEBUG_LOGGING.md](DEBUG_LOGGING.md) for details.

```bash
./apple-mail-mcp --debug
```

### Configuration

#### For STDIO Transport

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

```
http://localhost:8787
```

## Troubleshooting

### Automation Permission Errors

If you see:
```
Mail.app startup check failed: osascript execution failed: signal: killed
```

Grant automation permissions:

1. Open **System Settings** → **Privacy & Security** → **Automation**
2. Find the application running the server (Terminal, Claude, etc.)
3. Enable the checkbox next to **Mail**
4. Restart the MCP server

**Alternative:** Run the server from Terminal to trigger the permission prompt, then click **OK**.

**Note:** You may need to restart the application running the MCP server for changes to take effect.

### Mail.app Not Running

Simply open Mail.app. It must be running for the server to work.

## Available Tools

### list_accounts

Lists all configured email accounts in Apple Mail.

**Parameters:**
- `enabled` (boolean, optional): Filter to only show enabled accounts (default: false)

**Output:**
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

### get_selected_messages

Gets the currently selected message(s) in the frontmost Mail.app viewer window.

**Parameters:**
- None (operates on current selection)

**Output:**
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

Creates a reply to a specific message and saves it as a draft in the Drafts mailbox. Mail.app automatically includes the quoted original message. The reply is NOT sent automatically.

**Parameters:**
- `account` (string, required): Name of the email account
- `mailbox` (string, required): Name of the mailbox containing the message to reply to
- `message_id` (integer, required): The unique ID of the message to reply to
- `reply_content` (string, required): The content/body of the reply message
- `opening_window` (boolean, optional): Whether to show the window for the reply message. Default is false.
- `reply_to_all` (boolean, optional): Whether to reply to all recipients. Default is false.

**Output:**
- Object containing:
  - `draft_id`: ID of the created draft message
  - `subject`: Subject line of the reply (prefixed with "Re: ")
  - `to_recipients`: Array of recipient email addresses
  - `drafts_mailbox`: Name of the Drafts mailbox where the reply was saved
  - `message`: Confirmation message

**Important Notes:**
- The returned `draft_id` is obtained after a 4-second sync delay
- If creating multiple drafts rapidly, wait 4+ seconds between operations
- See [docs/DRAFT_MANAGEMENT.md](docs/DRAFT_MANAGEMENT.md) for details

### create_outgoing_message

Creates a new outgoing email message with optional Markdown formatting. The message is saved but NOT sent automatically - you must send it manually in Mail.app.

**Parameters:**
- `subject` (string, required): Subject line of the email
- `content` (string, required): Email body content (supports Markdown formatting when `content_format` is "markdown")
- `content_format` (string, optional): Content format: "plain" or "markdown". Default is "markdown"
- `to_recipients` (array of strings, required): List of To recipient email addresses
- `cc_recipients` (array of strings, optional): List of CC recipient email addresses
- `bcc_recipients` (array of strings, optional): List of BCC recipient email addresses
- `sender` (string, optional): Sender email address (uses default account if omitted)
- `opening_window` (boolean, optional): Whether to show the compose window. Default is false

**Rich Text Formatting:**

When `content_format` is set to "markdown", the content is parsed as Markdown and rendered with rich text styling:

**Supported Markdown Elements:**
- **Headings**: `# H1` through `###### H6`
- **Bold**: `**bold text**`
- **Italic**: `*italic text*`
- **Bold+Italic**: `***bold and italic text***`
- **Strikethrough**: `~~strikethrough text~~`
- **Inline Code**: `` `code` ``
- **Code Blocks**: ` ```code block``` `
- **Blockquotes**: `> quote`
- **Lists**: Unordered (`-`, `*`) and ordered (`1.`, `2.`)
- **Nested Lists**: Up to 4 levels deep
- **Links**: `[text](url)` (rendered as "text (url)")
- **Horizontal Rules**: `---`
- **Hard Line Breaks**: Two spaces at end of line creates line break within paragraph

**Example:**
```json
{
  "subject": "Project Update",
  "content": "# Weekly Report\n\nThis week we:\n\n- Completed **Phase 1**\n- Started *Phase 2*\n\n## Code Changes\n\n```\nfunction example() {\n  return true;\n}\n```",
  "content_format": "markdown",
  "to_recipients": ["team@example.com"],
  "opening_window": false
}
```

**Custom Styling:**

You can customize rich text styling by providing a YAML configuration file:

```bash
./apple-mail-mcp --rich-text-styles=/path/to/custom_styles.yaml
```

**Margin Support:**

Block elements (headings, code blocks, blockquotes, lists) support `margin_top` and `margin_bottom` properties (measured in font points) to add spacing:

```yaml
styles:
  h1:
    font: "Helvetica-Bold"
    size: 24
    margin_top: 12    # 12 point empty line before heading
    margin_bottom: 6   # 6 point empty line after heading
  
  code_block:
    margin_top: 6
    margin_bottom: 6
  
  blockquote:
    margin_top: 6
    margin_bottom: 6
  
  list:
    margin_top: 6     # Applied to entire list, not individual items
    margin_bottom: 6
```

**Note:** Margins are applied to the block as a whole (e.g., entire list), not to individual items within the block.

See [docs/RICH_TEXT_DESIGN.md](docs/RICH_TEXT_DESIGN.md) for the complete styling specification and examples.

**Output:**
- Object containing:
  - `outgoing_id`: ID of the created OutgoingMessage
  - `subject`: Subject line
  - `sender`: Sender email address
  - `to_recipients`: Array of To recipient addresses
  - `cc_recipients`: Array of CC recipient addresses
  - `bcc_recipients`: Array of BCC recipient addresses
  - `message`: Confirmation message
  - `warning`: (optional) Warning if some recipients couldn't be added

**Important Notes:**
- The OutgoingMessage only exists in memory while Mail.app is running
- For persistent drafts that survive Mail.app restart, use `reply_to_message` instead
- The message is NOT sent automatically - manual sending required
- Default format is Markdown (rich text enabled by default)
- Plain text content works as Markdown with no special characters (renders as single paragraph)
- Use `content_format: "plain"` to explicitly bypass Markdown parsing
- Rich text rendering errors fail immediately with clear error messages (no silent fallback to plain text)

## Architecture

- **Go**: Main server implementation using the MCP Go SDK
- **JXA (JavaScript for Automation)**: Scripts embedded in the binary for Mail.app interaction
- **STDIO Transport**: Simple, stateless communication protocol

All JXA scripts are embedded at compile time using `//go:embed`, making the server a single, self-contained binary.

## Development

### Build

```bash
make build
```

### Git Hooks

Install pre-commit hooks that run `go fmt`:

```bash
make install-hooks
```

### Format

```bash
make fmt
```

### Clean

```bash
make clean
```

## Error Handling

The server provides detailed error messages including:
- Script errors with clear descriptions
- Missing data with descriptive errors
- Invalid parameters with usage hints
- Argument context for debugging

## Limitations

- **macOS only**: Relies on Mail.app and JXA
- **Mail.app required**: Mail.app must be running
- **Attachment MIME types**: Not available due to Mail.app API limitations

### Rich Text Limitations

Due to JXA and Mail.app RichText API constraints:

- **Strikethrough**: Rendered as styled text (gray color) but not actual strikethrough formatting. Mail.app's RichText API doesn't support true strikethrough via JXA.
- **Links**: Rendered as "text (url)" format, not clickable links. Creating clickable links programmatically via JXA is not straightforward with Mail.app's API.
- **Background colors**: Not supported by Mail.app's RichText API (only foreground colors)
- **Tables**: Not implemented (would require complex grid layout)
- **Images**: Use Mail.app attachments instead
- **Dark mode**: All colors use character-level styling for consistency. Mail.app automatically adapts character-level colors in dark mode.

For strikethrough and links, the text is styled distinctively (different color/font) to indicate the formatting intent, but the actual strikethrough line or clickable link behavior is not available through JXA automation.



## License

MIT License - see LICENSE file for details

## Acknowledgments

Built with the [MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) from Anthropic.
