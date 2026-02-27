# Mail MCP Server

[![CI](https://github.com/dastrobu/mail-mcp/actions/workflows/ci.yaml/badge.svg)](https://github.com/dastrobu/mail-mcp/actions/workflows/ci.yaml)

A Model Context Protocol (MCP) server providing programmatic access to macOS Mail.app using JavaScript for Automation (JXA).

## Table of Contents

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

- [Overview](#overview)
- [Security & Privacy](#security--privacy)
- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
  - [Option 1: Homebrew (Recommended)](#option-1-homebrew-recommended)
  - [Option 2: Download Binary](#option-2-download-binary)
  - [Option 3: Install via Go](#option-3-install-via-go)
  - [Option 4: Build from Source](#option-4-build-from-source)
- [Usage](#usage)
  - [HTTP Transport (Recommended)](#http-transport-recommended)
  - [STDIO Transport](#stdio-transport)
  - [MCP Client Configuration](#mcp-client-configuration)
  - [Command-Line Options](#command-line-options)
- [Permissions](#permissions)
  - [Accessibility Permissions](#accessibility-permissions)
  - [Automation Permissions](#automation-permissions)
  - [Manual Permission Configuration](#manual-permission-configuration)
  - [Resetting Permissions](#resetting-permissions)
- [Troubleshooting](#troubleshooting)
  - [Automation Permission Errors](#automation-permission-errors)
  - [Mail.app Not Running](#mailapp-not-running)
  - [Debug Mode](#debug-mode)
  - [Bash Completion](#bash-completion)
- [Available Tools](#available-tools)
  - [list_accounts](#list_accounts)
  - [list_mailboxes](#list_mailboxes)
  - [get_message_content](#get_message_content)
  - [get_selected_messages](#get_selected_messages)
  - [find_messages](#find_messages)
  - [list_drafts](#list_drafts)
  - [create_reply_draft](#create_reply_draft)
  - [replace_reply_draft](#replace_reply_draft)
  - [create_outgoing_message](#create_outgoing_message)
  - [list_outgoing_messages](#list_outgoing_messages)
  - [replace_outgoing_message](#replace_outgoing_message)
- [Upgrading](#upgrading)
  - [Homebrew](#homebrew)
  - [Manual Installation](#manual-installation)
- [Uninstalling](#uninstalling)
  - [Homebrew](#homebrew-1)
  - [Manual Installation](#manual-installation-1)
- [Architecture](#architecture)
- [Development](#development)
  - [Build](#build)
  - [Git Hooks](#git-hooks)
  - [Format](#format)
  - [Update Table of Contents](#update-table-of-contents)
  - [Clean](#clean)
- [Error Handling](#error-handling)
- [Limitations](#limitations)
  - [Rich Text Limitations](#rich-text-limitations)
- [License](#license)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Overview

This MCP server enables AI assistants and other MCP clients to interact with Apple Mail on macOS. It provides read-only access to mailboxes, messages, and search functionality through a clean, typed interface.

## Security & Privacy

- **Human-in-the-loop design**: No emails are sent automatically - all drafts require manual sending. This prevents agents from sending emails without human oversight.
- No data transmitted outside of the MCP connection
- Runs locally on your machine
- Grant automation and accessibility permissions to the MCP server alone, not to the terminal or any other application like Claude Code.
- No credentials to a mail account ot SMTP server required, all interactions happen transparently with the Mail.app.

## Features

- **List Accounts**: Enumerate all configured email accounts with their properties
- **List Mailboxes**: Enumerate all available mailboxes and accounts
- **Get Message Content**: Fetch detailed content of individual messages
- **Get Selected Messages**: Retrieve currently selected message(s) in Mail.app
- **Find Messages**: Search messages with efficient filtering by subject, sender, read status, flags, and date ranges
- **Create Reply Draft**: Create a reply to a message with preserved quotes using the Accessibility API.
- **Create Outgoing Message**: Create new email drafts with Markdown rendering to rich text.
- **Replace Drafts**: Robustly update existing drafts (replies or standalone) while preserving quotes and signatures.
- **Rich Text Support**: Native support for Markdown (headings, bold, italic, links, strikethrough, lists, code blocks, and more) using native Mail.app rendering via the Accessibility API.

## Requirements

- macOS (Mail.app is macOS-only)
- Mail.app configured with at least one email account (does not need to be running at server startup)
- **Automation and Accessibility permissions** for Mail.app (see [Permissions](#permissions) below)

## Installation

### Option 1: Homebrew (Recommended)

```bash
# Add the tap
brew tap dastrobu/tap

# Install
brew install mail-mcp

# Start the service (Standard)
brew services start mail-mcp

# OR use the built-in subcommand for more customization (port, debug)
mail-mcp launchd create
```

**Important**: For proper automation permissions, you must run the server as a service (not from Terminal). Using `brew services start` is the standard way, while `mail-mcp launchd create` offers more customization.

**Note**: When you upgrade via `brew upgrade mail-mcp`, the launchd service will automatically restart with the new version if it's already running. You don't need to manually recreate the service.

➡️ See [Usage](#usage) for how to configure and use the server.

### Option 2: Download Binary

Download the latest release from [GitHub Releases](https://github.com/dastrobu/mail-mcp/releases):

- **Intel Mac**: `mail-mcp_*_darwin_amd64.tar.gz`
- **Apple Silicon**: `mail-mcp_*_darwin_arm64.tar.gz`

```bash
# Extract
tar -xzf mail-mcp_*.tar.gz

# Set up launchd service (uses full path to binary)
mail-mcp launchd create
```

➡️ See [Usage](#usage) for how to configure and use the server.

### Option 3: Install via Go

```bash
# Install directly from GitHub (requires Go 1.26+)
go install github.com/dastrobu/mail-mcp@latest

# Set up launchd service
mail-mcp launchd create
```

**Note**: Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is in your PATH, or use the full path:

```bash
~/go/bin/mail-mcp launchd create
```

➡️ See [Usage](#usage) for how to configure and use the server.

### Option 4: Build from Source

```bash
git clone https://github.com/dastrobu/mail-mcp.git
cd mail-mcp

# Build locally
go build -v -o mail-mcp .

# Set up launchd service
./mail-mcp launchd create
```

➡️ See [Usage](#usage) for how to configure and use the server.

## Usage

The server supports two transport modes: **HTTP (recommended)** and STDIO.

### HTTP Transport (Recommended)

HTTP mode runs the server as a standalone daemon, allowing automation permissions to be granted directly to the `mail-mcp` binary rather than the parent application.

⚠️ To get permissions granted to the binary (not Terminal or IDE), you must launch it without Terminal as the parent process.

#### Option 1: Using launchd (Recommended for Production)

Create a launch agent to run the server in the background.

**Quick setup using the built-in subcommand:**

```bash
# Run the setup subcommand
mail-mcp launchd create
```

➡️ See [MCP Client Configuration](#mcp-client-configuration) to connect your MCP client.

Or alternatively, create the launch agent manually:

```bash
# See available options
mail-mcp launchd create -h

# With custom port
mail-mcp --port=3000 launchd create

# With debug logging enabled
mail-mcp --debug launchd create

# Disable automatic startup on login (start manually instead)
mail-mcp launchd create --disable-run-at-load

# The subcommand will:
# - Create the launchd plist
# - Load and start the service
# - Show you the connection URL and useful commands
```

**To remove the service:**

```bash
mail-mcp launchd remove
```

Check logs: `tail -f ~/Library/Logs/com.github.dastrobu.mail-mcp/mail-mcp.log ~/Library/Logs/com.github.dastrobu.mail-mcp/mail-mcp.err`

To stop: `launchctl stop com.github.dastrobu.mail-mcp`
To unload: `launchctl unload ~/Library/LaunchAgents/com.github.dastrobu.mail-mcp.plist`

#### Option 2: Running from Terminal (Quick Testing)

If you launch from Terminal, **Terminal will be asked for permissions**, not the binary:

```bash
# This will prompt for Terminal's permissions (not ideal)
mail-mcp --transport=http

# Custom port
mail-mcp --transport=http --port=3000

# Custom host and port
mail-mcp --transport=http --host=0.0.0.0 --port=3000
```

This is fine for quick testing, but for production use launchd.

**Connect MCP clients to:** `http://localhost:8787`

➡️ See [MCP Client Configuration](#mcp-client-configuration) to connect your MCP client.

### STDIO Transport

STDIO mode runs the server as a child process of the MCP client. Note that automation permissions will be required for the parent application (Terminal, Claude Desktop, etc.).

```bash
mail-mcp
```

➡️ See [MCP Client Configuration](#mcp-client-configuration) to connect your MCP client.

### MCP Client Configuration

#### VS Code Configuration

Make sure the server is running, see [HTTP Transport](#http-transport-recommended)

Configure VS Code (`~/Library/Application Support/Code/User/mcp.json` on macOS):

```json
{
  "servers": {
    "mail-mcp": {
      "type": "http",
      "url": "http://localhost:8787"
    }
  }
}
```

#### Zed Configuration

Make sure the server is running, see [HTTP Transport](#http-transport-recommended)

Configure Zed (`~/.config/zed/settings.json`):

```json
{
  "context_servers": {
    "mail-mcp": {
      "url": "http://localhost:8787"
    }
  }
}
```

#### Claude Desktop Configuration

Make sure the server is running, see [HTTP Transport](#http-transport-recommended)

Configure Claude Desktop (`~/Library/Application Support/Claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "mail-mcp": {
      "url": "http://localhost:8787"
    }
  }
}
```

### Command-Line Options

Use `-h` or `--help` with any command to see available options:

```bash
mail-mcp -h                    # Show main help
mail-mcp launchd -h            # Show launchd subcommands
mail-mcp launchd create -h     # Show launchd create options
```

**Available options:**

```
--transport=[stdio|http]  Transport type (default: stdio)
--port=PORT              HTTP port (default: 8787, only used with --transport=http)
--host=HOST              HTTP host (default: localhost, only used with --transport=http)
--debug                  Enable debug logging of tool calls and results to stderr

-h, --help               Show help message

Commands:
  launchd create         Set up launchd service for automatic startup (HTTP mode)
                         Use --debug flag to enable debug logging in the service
                         Use --disable-run-at-load to prevent automatic startup on login
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

➡️ See [MCP Client Configuration](#mcp-client-configuration) to connect your MCP client.

## Permissions

macOS requires both **Automation** and **Accessibility** permissions for full functionality.

### Accessibility Permissions

The draft creation and replacement tools (`create_reply_draft`, `replace_reply_draft`, `create_outgoing_message`, `replace_outgoing_message`) use the macOS Accessibility API to simulate pasting content.
This is the only reliable way to support rich text (Markdown) while preserving original message quotes and signatures.
If you only want to use tools that read emails, you can skip granting the accessibility permission.

To ensure the highest level of security, grant accessibility permissions directly to the `mail-mcp` binary alone:

1. Open **System Settings** → **Privacy & Security** → **Accessibility**.
2. Click the **+** (plus) button at the bottom of the list.
3. In the file picker that appears, navigate to the path where `mail-mcp` is installed.
   - *Tip:* Press `Cmd + Shift + G` to enter the path manually (e.g. `/usr/local/bin/mail-mcp`).
4. Select the binary and click **Open**.
5. Ensure the toggle switch next to `mail-mcp` is **ON**.

If permissions are missing, these tools will return an error explaining what to do.

### Automation Permissions

macOS requires automation permissions to control Mail.app. The permission behavior depends on which transport mode you use:

#### HTTP Transport (Recommended)

When using `--transport=http`, permissions can be granted to the `mail-mcp` binary itself, **but only if launched without Terminal as the parent process**.

**Using launchd (recommended):**

1. Set up the launchd service: `mail-mcp launchd create`
2. macOS will prompt for automation permissions for `mail-mcp` binary
3. Click **OK** to grant access
4. The server is now ready to use

**Using Finder:**

1. Double-click the `mail-mcp` binary in Finder
2. macOS will prompt for automation permissions for `mail-mcp` binary
3. Click **OK** to grant access

**Using Terminal (quick testing only):**

1. Run `mail-mcp --transport=http` from Terminal
2. macOS will prompt for automation permissions for **Terminal.app** (not the binary)
3. Click **OK** to grant access to Terminal
4. Note: This grants permission to Terminal, not the binary

**Advantage:** With launchd or Finder launch, permissions stay with the binary and work with all MCP clients. With Terminal launch, only Terminal gets permissions.

#### STDIO Transport

When using STDIO mode (default), permissions are granted to the **parent process** (Terminal, Claude Desktop, etc.) that launches the server:

1. Start the server (or let your MCP client start it)
2. macOS will prompt for automation permissions on first run
3. Click **OK** to grant access to the parent application
4. The server is now ready to use

**Note:** If you switch between different applications (e.g., Terminal vs Claude Desktop), each will need its own automation permission.

### Manual Permission Configuration

If the prompt doesn't appear or you need to change permissions:

1. Open **System Settings** → **Privacy & Security** → **Automation**
2. Find `mail-mcp` (HTTP mode) or the parent application (STDIO mode)
3. Enable the checkbox next to **Mail**
4. Restart the server

### Resetting Permissions

To reset automation permissions (useful for testing or troubleshooting):

```bash
# Reset all automation permissions (will prompt again on next run)
tccutil reset AppleEvents

# Reset for a specific application (e.g., Terminal)
tccutil reset AppleEvents com.apple.Terminal

# Reset for a specific application (e.g., Mail)
tccutil reset Accessibility
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

### Debug Mode

When `--debug` is enabled, the server logs all MCP protocol interactions and JXA script diagnostics to stderr, including tool calls, results, and JXA script logs. See [DEBUG_LOGGING.md](DEBUG_LOGGING.md) for details.

```bash
mail-mcp --debug
```

### Bash Completion

Enable tab completion for commands and flags:

```bash
# Generate completion script
mail-mcp completion bash > /usr/local/etc/bash_completion.d/mail-mcp

# Or add to your ~/.bashrc or ~/.bash_profile
source <(mail-mcp completion bash)
```

After sourcing, you can use tab completion:

```bash
mail-mcp --transport=<TAB>    # Completes: http, stdio
mail-mcp launchd <TAB>        # Completes: create, remove
```

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

### find_messages

Finds messages in a mailbox using efficient bulk array property fetching. Supports filtering by subject, sender, read status, flagged status, and date ranges. Uses constant-time filtering for optimal performance.

**Important:** At least one filter criterion must be specified to prevent accidentally fetching all messages.

**Parameters:**

- `account` (string, required): Name of the email account
- `mailboxPath` (array of strings, required): Mailbox path array (e.g., `["Inbox"]` or `["Inbox", "GitHub"]`)
- `subject` (string, optional): Filter by subject (substring match)
- `sender` (string, optional): Filter by sender email address (substring match)
- `readStatus` (boolean, optional): Filter by read status (true for read, false for unread)
- `flaggedOnly` (boolean, optional): Filter for flagged messages only (default: false)
- `dateAfter` (string, optional): Filter for messages received after this ISO date (e.g., "2024-01-01T00:00:00Z")
- `dateBefore` (string, optional): Filter for messages received before this ISO date (e.g., "2024-12-31T23:59:59Z")
- `limit` (integer, optional): Maximum number of messages to return (1-1000, default: 50)

**Note:** While all filter parameters are individually optional, you must provide at least one filter criterion. The tool will return an error if no filters are specified.

**Output:**

```json
{
  "messages": [
    {
      "id": 123456,
      "subject": "Meeting Tomorrow",
      "sender": "colleague@example.com",
      "date_received": "2024-02-11T10:30:00Z",
      "date_sent": "2024-02-11T10:25:00Z",
      "read_status": true,
      "flagged_status": false,
      "message_size": 2048,
      "content_preview": "Hi team, just wanted to remind everyone about...",
      "content_length": 500,
      "to_count": 3,
      "cc_count": 1,
      "total_recipients": 4,
      "mailbox_path": ["Inbox"],
      "account": "Work"
    }
  ],
  "count": 1,
  "total_matches": 15,
  "limit": 50,
  "has_more": false,
  "filters_applied": {
    "subject": "meeting",
    "sender": null,
    "read_status": null,
    "flagged_only": false,
    "date_after": "2024-02-01T00:00:00Z",
    "date_before": null
  }
}
```

**Performance:**

The tool uses AppleScript bulk array property fetching to extract filters efficiently. This makes it efficient even for mailboxes with thousands of messages.

**Examples:**

Find unread messages from a specific sender:

```json
{
  "account": "Work",
  "mailboxPath": ["Inbox"],
  "sender": "boss@example.com",
  "readStatus": false,
  "limit": 10
}
```

Find flagged messages from last week:

```json
{
  "account": "Personal",
  "mailboxPath": ["Inbox"],
  "flaggedOnly": true,
  "dateAfter": "2024-02-04T00:00:00Z",
  "limit": 50
}
```

Find all messages with specific subject in nested mailbox:

```json
{
  "account": "Work",
  "mailboxPath": ["Inbox", "GitHub", "notifications"],
  "subject": "Pull Request",
  "limit": 100
}
```

### list_drafts

Lists persistent draft messages from the Drafts mailbox for a specific account.

**Parameters:**

- `account` (string, required): Name of the email account
- `limit` (integer, optional): Maximum number of drafts to return (1-1000, default: 50)

### create_reply_draft

Creates a reply to a specific message using the Accessibility API. This approach preserves the original message quote and signature. It requires Accessibility permissions for the mail-mcp binary. The message is NOT sent automatically.

**Parameters:**

- `account` (string, required): Name of the email account
- `mailboxPath` (array of strings, required): Path to the mailbox as an array (e.g. `["Inbox"]`). Use the `mailboxPath` field from `get_selected_messages` or `find_messages`.
- `message_id` (integer, required): The unique ID of the message to reply to
- `reply_content` (string, required): The content/body of the reply message
- `content_format` (string, optional): Content format: "plain" or "markdown". Default is "markdown"
- `reply_to_all` (boolean, optional): Whether to reply to all recipients. Default is false.

**Output:**

- `draft_id`: ID of the created draft message
- `subject`: Subject line of the reply
- `original_message_id`: ID of the message replied to
- `message`: Confirmation message

### replace_reply_draft

Replaces an existing reply draft with new content while preserving the original message quote and signature. It achieves this by deleting the old draft and creating a fresh reply to the original message before pasting the new content. Requires Accessibility permissions.

**Parameters:**

- `outgoing_id` (integer, required): The ID of the reply draft to replace (from `list_outgoing_messages`)
- `original_message_id` (integer, required): The ID of the original message being replied to
- `account` (string, required): The account name of the original message
- `mailbox_path` (array of strings, required): The mailbox path of the original message
- `content` (string, required): New email body content (supports Markdown)
- `content_format` (string, optional): Content format: "plain" or "markdown". Default is "markdown"
- `subject` (string, optional): New subject line (optional)
- `to_recipients` (array of strings, optional): New list of To recipients
- `cc_recipients` (array of strings, optional): New list of CC recipients
- `bcc_recipients` (array of strings, optional): New list of BCC recipients
- `sender` (string, optional): New sender email address

### create_outgoing_message

Creates a new outgoing email message using the Accessibility API to support rich text content. The message is saved but NOT sent automatically. Requires Accessibility permissions.

**Parameters:**

- `subject` (string, required): Subject line of the email
- `content` (string, required): Email body content (supports Markdown formatting when `content_format` is "markdown")
- `content_format` (string, optional): Content format: "plain" or "markdown". Default is "markdown"
- `to_recipients` (array of strings, required): List of To recipient email addresses
- `cc_recipients` (array of strings, optional): List of CC recipient email addresses
- `bcc_recipients` (array of strings, optional): List of BCC recipient email addresses
- `sender` (string, optional): Sender email address (uses default account if omitted)

### list_outgoing_messages

Lists all `OutgoingMessage` objects currently in memory in Mail.app. These are unsent messages that were created with `create_outgoing_message` or `create_reply_draft`. Returns `outgoing_id` for each message which can be used with replacement tools.

### replace_outgoing_message

Replaces an existing outgoing message (draft) with new content using the Accessibility API. This tool is for standalone drafts (not replies). It deletes the old draft and creates a fresh instance before pasting the new content. Requires Accessibility permissions.

**Parameters:**

- `outgoing_id` (integer, required): The ID of the outgoing message to replace
- `content` (string, required): New email body content (supports Markdown)
- `content_format` (string, optional): Content format: "plain" or "markdown". Default is "markdown"
- `subject` (string, optional): New subject line
- `to_recipients` (array of strings, optional): New list of To recipients
- `cc_recipients` (array of strings, optional): New list of CC recipients
- `bcc_recipients` (array of strings, optional): New list of BCC recipients
- `sender` (string, optional): New sender email address

**Rich Text Formatting:**

When `content_format` is set to "markdown", the content is parsed as Markdown and rendered with rich text styling:

**Supported Markdown Elements:**

- **Headings**: `# H1` through `###### H6`
- **Bold**: `**bold text**`
- **Italic**: `*italic text*`
- **Bold+Italic**: `***bold and italic text***`
- **Strikethrough**: `~~strikethrough text~~` (natively supported)
- **Inline Code**: `` `code` ``
- **Code Blocks**: ` ```code block``` `
- **Blockquotes**: `> quote`
- **Lists**: Unordered (`-`, `*`) and ordered (`1.`, `2.`)
- **Nested Lists**: Up to 4 levels deep
- **Links**: `[text](url)` (rendered as native, clickable links)
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
- For persistent drafts that survive Mail.app restart, use `create_reply_draft` or save the draft after creation
- The message is NOT sent automatically - manual sending required
- Default format is Markdown
- Plain text content works as Markdown with no special characters
- Use `content_format: "plain"` to explicitly bypass Markdown parsing

## Upgrading

**Note on Permissions & Service Restart:** After upgrading, macOS may prompt you to re-grant **Automation** and **Accessibility** permissions to the new binary. If features like "Get Selected Messages" or "Create Reply Draft" stop working, please re-enable these permissions in **System Settings > Privacy & Security**. You may also need to restart the service for the changes to take effect.

### Homebrew

When you upgrade via Homebrew, the launchd service will automatically restart with the new version:

```bash
brew upgrade mail-mcp
```

The upgrade process:

1. Downloads and installs the new version
2. Updates the symlink at `/opt/homebrew/bin/mail-mcp` to point to the new version
3. If a launchd service exists, automatically recreates it with the new binary while preserving your settings:
   - Port (if customized)
   - Host (if customized)
   - Debug flag (if enabled)
   - RunAtLoad setting (automatic startup behavior)
4. No manual intervention required

**Note**: The upgrade preserves all your custom settings by parsing the existing plist and recreating the service with the same configuration.

### Manual Installation

If you installed manually (via binary download or Go install), you'll need to restart the launchd service after upgrading:

```bash
# After upgrading the binary
mail-mcp launchd create
```

This will recreate the service with the new binary path.

## Uninstalling

### Homebrew

**⚠️ IMPORTANT:** Stop the service BEFORE uninstalling:

```bash
# Step 1: Stop the service (whichever method you used)
brew services stop mail-mcp
# OR
mail-mcp launchd remove

# Step 2: Uninstall the package
brew uninstall mail-mcp

# Step 3 (Optional): Remove logs
# If you used brew services:
rm $(brew --prefix)/var/log/mail-mcp.*
# If you used mail-mcp launchd create:
rm -r ~/Library/Logs/com.github.dastrobu.mail-mcp/
```

**Why this order matters:** The `launchd remove` command needs the binary to properly unload and remove the service. If you uninstall first, you'll need to manually remove the plist file.

**If you already uninstalled without removing the service:**

```bash
# Manually remove the plist and unload the service
launchctl unload ~/Library/LaunchAgents/com.github.dastrobu.mail-mcp.plist
rm ~/Library/LaunchAgents/com.github.dastrobu.mail-mcp.plist
```

### Manual Installation

If you installed manually, remove the launchd service first, then delete the binary:

```bash
# Remove the launchd service
mail-mcp launchd remove

# Remove the binary (adjust path as needed)
sudo rm /usr/local/bin/mail-mcp

# Optionally remove logs
rm -r ~/Library/Logs/com.github.dastrobu.mail-mcp/
```

## Architecture

- **Go**: Main server implementation using the MCP Go SDK
- **JXA (JavaScript for Automation)**: Scripts embedded in the binary for Mail.app interaction
- **Dual Transport Support**: HTTP (recommended) and STDIO transports for flexible deployment

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

### Update Table of Contents

The Table of Contents is auto-generated using [doctoc](https://github.com/thlorenz/doctoc). To update it after making changes to section headings:

```bash
# Using make
make doctoc

# Or directly with npx
npx doctoc --maxlevel 3 README.md --github --title "## Table of Contents"
```

The TOC is wrapped in special comments (`<!-- START doctoc -->` ... `<!-- END doctoc -->`) and should never be edited manually. The `--maxlevel 3` flag limits the TOC to main sections (h2 headings only) for better readability.

**Note:** The pre-commit hook automatically runs doctoc when `README.md` is staged, so the TOC stays in sync with section changes.

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

The move to the Accessibility-based pasting strategy has resolved many previous JXA-related constraints.

- **Tables**: Markdown tables are currently rendered as plain text. Native table support is planned.
- **Images**: Inline images are not currently supported; please use the standard Mail.app attachment feature.
- **Dark Mode**: Mail.app automatically adapts the colors of pasted HTML content to match your current system theme (Light or Dark).

Previously documented limitations regarding **Strikethrough** and **Links** are now resolved—they are rendered as native, functional Mail.app elements.

## License

MIT License - see LICENSE file for details
