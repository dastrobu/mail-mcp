# Apple Mail MCP Server

[![CI](https://github.com/dastrobu/apple-mail-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/dastrobu/apple-mail-mcp/actions/workflows/ci.yaml)

A Model Context Protocol (MCP) server providing programmatic access to macOS Mail.app using JavaScript for Automation (JXA).

## Overview

This MCP server enables AI assistants and other MCP clients to interact with Apple Mail on macOS. It provides read-only access to mailboxes, messages, and search functionality through a clean, typed interface.

## Security & Privacy

- **Human-in-the-loop design**: No emails are sent automatically - all drafts require manual sending. This prevents agents from sending emails without human oversight.
- No data transmitted outside of the MCP connection
- Runs locally on your machine
- Grant automation permissions to the MCP server alone, not to the terminal or any other application like Claude Code.
- No credentials to a mail account ot SMTP server required, all interactions happen transparently with the Mail.app.

## Features

- **List Accounts**: Enumerate all configured email accounts with their properties
- **List Mailboxes**: Enumerate all available mailboxes and accounts
- **Get Message Content**: Fetch detailed content of individual messages
- **Get Selected Messages**: Retrieve currently selected message(s) in Mail.app
- **Reply to Message**: Create a reply to a message and save it as a draft
- **Create Outgoing Message**: Create new email drafts with optional Markdown rendering to rich text.
- **Rich Text Support**: Format emails with Markdown (headings, bold, italic, lists, code blocks, and more)

## Requirements

- macOS (Mail.app is macOS-only)
- Mail.app configured with at least one email account (does not need to be running at server startup)
- Automation permissions (see [Automation Permissions](#automation-permissions) below)

## Installation

### Option 1: Homebrew (Recommended)

```bash
# Add the tap
brew tap dastrobu/tap

# Install
brew install apple-mail-mcp

# Set up launchd service (IMPORTANT for proper permissions)
apple-mail-mcp launchd create
```

**Important**: After installation, you must run `apple-mail-mcp launchd create` to set up the launchd service. This ensures automation permissions are granted to the binary itself (not Terminal or Claude Desktop).

### Option 2: Download Binary

Download the latest release from [GitHub Releases](https://github.com/dastrobu/apple-mail-mcp/releases):

- **Intel Mac**: `apple-mail-mcp_*_darwin_amd64.tar.gz`
- **Apple Silicon**: `apple-mail-mcp_*_darwin_arm64.tar.gz`

```bash
# Extract
tar -xzf apple-mail-mcp_*.tar.gz

# Set up launchd service (uses full path to binary)
apple-mail-mcp launchd create
```

### Option 3: Install via Go

```bash
# Install directly from GitHub (requires Go 1.26+)
go install github.com/dastrobu/apple-mail-mcp@latest

# Set up launchd service
apple-mail-mcp launchd create
```

**Note**: Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is in your PATH, or use the full path:
```bash
~/go/bin/apple-mail-mcp launchd create
```

### Option 4: Build from Source

```bash
git clone https://github.com/dastrobu/apple-mail-mcp.git
cd apple-mail-mcp

# Build locally
go build -o apple-mail-mcp .

# Set up launchd service
./apple-mail-mcp launchd create
```

## Usage

The server supports two transport modes: **HTTP (recommended)** and STDIO.

### HTTP Transport (Recommended)

HTTP mode runs the server as a standalone daemon, allowing automation permissions to be granted directly to the `apple-mail-mcp` binary rather than the parent application.

**Important:** To get permissions granted to the binary (not Terminal), you must launch it without Terminal as the parent process.

#### Option 1: Using launchd (Recommended for Production)

Create a launch agent to run the server in the background.

**Quick setup using the built-in subcommand:**

```bash
# Run the setup subcommand
apple-mail-mcp launchd create

# With custom port
apple-mail-mcp --port=3000 launchd create

# With debug logging enabled
apple-mail-mcp --debug launchd create

# The subcommand will:
# - Create the launchd plist
# - Load and start the service
# - Show you the connection URL and useful commands
```

**To remove the service:**

```bash
apple-mail-mcp launchd remove
```

Check logs: `tail -f ~/Library/Logs/com.github.dastrobu.apple-mail-mcp/apple-mail-mcp.log ~/Library/Logs/com.github.dastrobu.apple-mail-mcp/apple-mail-mcp.err`

To stop: `launchctl stop com.github.dastrobu.apple-mail-mcp`
To unload: `launchctl unload ~/Library/LaunchAgents/com.github.dastrobu.apple-mail-mcp.plist`

#### Option 2: Running from Terminal (Quick Testing)

If you launch from Terminal, **Terminal will be asked for permissions**, not the binary:

```bash
# This will prompt for Terminal's permissions (not ideal)
apple-mail-mcp --transport=http

# Custom port
apple-mail-mcp --transport=http --port=3000

# Custom host and port
apple-mail-mcp --transport=http --host=0.0.0.0 --port=3000
```

This is fine for quick testing, but for production use launchd (Option 1) or Finder (Option 2).

**Connect MCP clients to:** `http://localhost:8787`

### STDIO Transport

STDIO mode runs the server as a child process of the MCP client. Note that automation permissions will be required for the parent application (Terminal, Claude Desktop, etc.).

```bash
apple-mail-mcp
```

### Understanding Parent Process and Permissions

**Key Insight:** macOS grants automation permissions to the **process that launches the binary**, not the binary itself.

- **Launching from Terminal** (even with `--transport=http`): Terminal becomes the parent process, so Terminal gets the permission
- **Launching via launchd**: No parent process, so the `apple-mail-mcp` binary itself gets the permission
- **Launching via Finder** (double-click): Similar to launchd, the binary gets the permission

This is why **HTTP mode with launchd** is recommended - it ensures the binary (not Terminal or other applications) receives the automation permissions.

### Command-Line Options

```
--transport=[stdio|http]  Transport type (default: stdio)
--port=PORT              HTTP port (default: 8787, only used with --transport=http)
--host=HOST              HTTP host (default: localhost, only used with --transport=http)
--debug                  Enable debug logging of tool calls and results to stderr
--rich-text-styles=PATH  Path to custom rich text styles YAML file (uses embedded default if not specified)
--help                   Show help message

Commands:
  launchd create         Set up launchd service for automatic startup (HTTP mode)
                         Use --debug flag to enable debug logging in the service
  launchd remove         Remove launchd service
  completion bash        Generate bash completion script
```

Options can also be set via environment variables:

```
APPLE_MAIL_MCP_TRANSPORT=http
APPLE_MAIL_MCP_PORT=8787
APPLE_MAIL_MCP_HOST=localhost
APPLE_MAIL_MCP_DEBUG=true
APPLE_MAIL_MCP_RICH_TEXT_STYLES=/path/to/custom_styles.yaml
```

### Debug Mode

When `--debug` is enabled, the server logs all MCP protocol interactions and JXA script diagnostics to stderr, including tool calls, results, and JXA script logs. See [DEBUG_LOGGING.md](DEBUG_LOGGING.md) for details.

```bash
apple-mail-mcp --debug
```

### Bash Completion

Enable tab completion for commands and flags:

```bash
# Generate completion script
apple-mail-mcp completion bash > /usr/local/etc/bash_completion.d/apple-mail-mcp

# Or add to your ~/.bashrc or ~/.bash_profile
source <(apple-mail-mcp completion bash)
```

After sourcing, you can use tab completion:

```bash
apple-mail-mcp --transport=<TAB>    # Completes: http, stdio
apple-mail-mcp launchd <TAB>        # Completes: create, remove
```

### Configuration

#### Claude Desktop Configuration

**HTTP Transport (Recommended):**

1. Start the server using launchd (recommended) or Finder (see [HTTP Transport](#http-transport-recommended) section):

   ```bash
   apple-mail-mcp launchd create
   ```

   Or for quick testing from Terminal:

   ```bash
   apple-mail-mcp --transport=http
   ```

2. Configure Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json`):
   ```json
   {
     "mcpServers": {
       "apple-mail": {
         "url": "http://localhost:8787"
       }
     }
   }
   ```

**STDIO Transport:**

Configure Claude Desktop to launch the server as a child process:

```json
{
  "mcpServers": {
    "apple-mail": {
      "command": "/path/to/apple-mail-mcp"
    }
  }
}
```

**Note:** With STDIO, Claude Desktop will need automation permissions. With HTTP, only the `apple-mail-mcp` binary needs permissions.

## Automation Permissions

macOS requires automation permissions to control Mail.app. The permission behavior depends on which transport mode you use:

### HTTP Transport (Recommended)

When using `--transport=http`, permissions can be granted to the `apple-mail-mcp` binary itself, **but only if launched without Terminal as the parent process**.

**Using launchd (recommended):**

1. Set up the launchd service: `apple-mail-mcp launchd create`
2. macOS will prompt for automation permissions for `apple-mail-mcp` binary
3. Click **OK** to grant access
4. The server is now ready to use

**Using Finder:**

1. Double-click the `apple-mail-mcp` binary in Finder
2. macOS will prompt for automation permissions for `apple-mail-mcp` binary
3. Click **OK** to grant access

**Using Terminal (quick testing only):**

1. Run `apple-mail-mcp --transport=http` from Terminal
2. macOS will prompt for automation permissions for **Terminal.app** (not the binary)
3. Click **OK** to grant access to Terminal
4. Note: This grants permission to Terminal, not the binary

**Advantage:** With launchd or Finder launch, permissions stay with the binary and work with all MCP clients. With Terminal launch, only Terminal gets permissions.

### STDIO Transport

When using STDIO mode (default), permissions are granted to the **parent process** (Terminal, Claude Desktop, etc.) that launches the server:

1. Start the server (or let your MCP client start it)
2. macOS will prompt for automation permissions on first run
3. Click **OK** to grant access to the parent application
4. The server is now ready to use

**Note:** If you switch between different applications (e.g., Terminal vs Claude Desktop), each will need its own automation permission.

### Manual Permission Configuration

If the prompt doesn't appear or you need to change permissions:

1. Open **System Settings** → **Privacy & Security** → **Automation**
2. Find `apple-mail-mcp` (HTTP mode) or the parent application (STDIO mode)
3. Enable the checkbox next to **Mail**
4. Restart the server

### Resetting Permissions

To reset automation permissions (useful for testing or troubleshooting):

```bash
# Reset all automation permissions (will prompt again on next run)
tccutil reset AppleEvents

# Reset for a specific application (e.g., Terminal)
tccutil reset AppleEvents com.apple.Terminal
```

After resetting, the next time the server tries to control Mail.app, macOS will show the permission prompt again.

## Troubleshooting

### Automation Permission Errors

If you see:

```
Mail.app startup check failed: osascript execution failed: signal: killed
```

**Solution:** Grant automation permissions using the steps in [Automation Permissions](#automation-permissions) above.

### Mail.app Not Running

The server can start without Mail.app running. When you try to use a tool and Mail.app is not running, you'll receive a clear error message:

- **"Mail.app is not running. Please start Mail.app and try again"** - Simply open Mail.app and retry
- **"Mail.app automation permission denied..."** - Grant automation permissions in System Settings > Privacy & Security > Automation

Tool calls will automatically work once Mail.app is started and permissions are granted.

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

````json
{
  "subject": "Project Update",
  "content": "# Weekly Report\n\nThis week we:\n\n- Completed **Phase 1**\n- Started *Phase 2*\n\n## Code Changes\n\n```\nfunction example() {\n  return true;\n}\n```",
  "content_format": "markdown",
  "to_recipients": ["team@example.com"],
  "opening_window": false
}
````

**Custom Styling:**

You can customize rich text styling by providing a YAML configuration file:

```bash
apple-mail-mcp --rich-text-styles=/path/to/custom_styles.yaml
```

**Margin Support:**

Block elements (headings, code blocks, blockquotes, lists) support `margin_top` and `margin_bottom` properties (measured in font points) to add spacing:

```yaml
styles:
  h1:
    font: "Helvetica-Bold"
    size: 24
    margin_top: 12 # 12 point empty line before heading
    margin_bottom: 6 # 6 point empty line after heading

  code_block:
    margin_top: 6
    margin_bottom: 6

  blockquote:
    margin_top: 6
    margin_bottom: 6

  list:
    margin_top: 6 # Applied to entire list, not individual items
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
