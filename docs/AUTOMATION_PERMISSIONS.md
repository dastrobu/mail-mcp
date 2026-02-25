# Automation Permissions Guide

This document explains how macOS automation permissions work with the Apple Mail MCP server and how to choose the best transport mode for your use case.

## Overview

macOS requires explicit permission for applications to control other applications (like Mail.app) via automation. The MCP server executes JXA (JavaScript for Automation) scripts using `osascript`, which triggers macOS's automation permission checks.

**Key Insight:** The permission is granted to the **process that executes `osascript`**, not to the script itself. This means the parent process that launched the binary determines which application gets the permission.

> **Note: Automation vs. Accessibility Permissions**
> This document covers **Automation** permissions, which are required for most tools that interact with Mail.app.
>
> A few specific tools that perform rich text editing (`create_reply_draft`, `replace_reply_draft`, etc.) require a different, more sensitive permission: **Accessibility**. The workflow for granting Accessibility permissions is different. See the "Accessibility Permissions Workflow" section below for details.

## HTTP Transport (Recommended)

### How It Works

When you run the server with `--transport=http`, the server runs as a daemon process. However, **the parent process that launches the binary determines the permission**.

**Important:** If you launch from Terminal (`./mail-mcp --transport=http`), Terminal is still the parent process, so macOS will ask for **Terminal's** permission, not the binary's.

To get permissions granted to the `mail-mcp` binary itself, you must launch it **without Terminal as the parent**.

### Permission Behavior

- **Permission granted to:** The `mail-mcp` binary itself
- **Appears in System Settings as:** `mail-mcp`
- **Prompt timing:** First time the binary executes `osascript`
- **Persistence:** Permission stays with the binary across restarts

### Advantages

1. **One-time setup:** Grant permission once, works with all MCP clients
2. **No client permissions needed:** Terminal, Claude Desktop, etc. don't need automation permissions
3. **Cleaner permission model:** Only one entry in System Settings → Privacy & Security → Automation
4. **Better for production:** Server can run as a background service without shell/terminal permissions

### Setup Steps

#### Option 1: Using launchd (Recommended)

Use the built-in `launchd create` subcommand:

```bash
# Set up the launchd service (default port 8787)
./mail-mcp launchd create

# Or with custom port
./mail-mcp --port=3000 launchd create

# Or with custom host and port
./mail-mcp --host=0.0.0.0 --port=3000 launchd create
```

The subcommand will automatically:
- Create the launchd plist file
- Load and start the service
- Display configuration and useful commands

On first run, macOS shows permission prompt:
```
"mail-mcp" wants to control "Mail.app"
```

Click **OK** to grant permission. Server is ready - connect MCP clients to `http://localhost:8787`.

#### Option 2: Double-Click in Finder

1. Open Finder and navigate to the `mail-mcp` binary
2. Double-click the binary
3. macOS will prompt for permission for `mail-mcp` (not Terminal)
4. Click **OK** to grant access

#### Option 3: From Terminal (Quick Testing Only)

⚠️ **Note:** This will prompt for **Terminal's** permission, not the binary's:

```bash
./mail-mcp --transport=http
```

On first run, macOS shows:
```
"Terminal" wants to control "Mail.app"
```

This grants permission to Terminal, not the binary. Use Option 1 or 2 for production.

### MCP Client Configuration

Configure clients to connect to the HTTP endpoint:

**Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "mail-mcp": {
      "url": "http://localhost:8787"
    }
  }
}
```

**Other MCP Clients:** Point to `http://localhost:8787`

## STDIO Transport

### How It Works

When you run the server with STDIO transport (default):

```bash
./mail-mcp
```

The server runs as a **child process** of whatever launched it (Terminal, Claude Desktop, etc.). When it executes `osascript`, macOS sees the **parent process** as the requesting process.

### Permission Behavior

- **Permission granted to:** The parent process (Terminal.app, Claude.app, etc.)
- **Appears in System Settings as:** The parent application name
- **Prompt timing:** First time the parent process launches the server
- **Persistence:** Permission tied to the parent application

### Disadvantages

1. **Multiple permissions needed:** Each parent application needs its own permission
2. **Confusing permission entries:** Multiple "Terminal", "Claude", etc. entries in System Settings
3. **Permission doesn't follow binary:** Moving or renaming the parent app requires re-granting
4. **Client application needs permissions:** Not all MCP clients may be willing/able to request automation permissions

### Setup Steps

1. Launch the server (or let your MCP client launch it):
   ```bash
   ./mail-mcp
   ```

2. On first run, macOS shows permission prompt:
   ```
   "Terminal" wants to control "Mail.app"
   ```
   (or "Claude" or whichever application launched the server)

3. Click **OK** to grant permission to the parent application

4. Server is ready for that parent application

### MCP Client Configuration

**Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
      "mail-mcp": {
          "command": "/path/to/mail-mcp"
      }
  }
}
```

**Note:** Claude Desktop will need automation permissions granted to it.

## Comparison Table

| Aspect | HTTP Transport | STDIO Transport |
|--------|---------------|-----------------|
| Permission granted to | `mail-mcp` binary (if launched via launchd/Finder) OR parent process (if launched from Terminal) | Parent process (Terminal, Claude, etc.) |
| Number of permissions | One (for the binary) | One per parent application |
| Client needs permissions | No | Yes |
| Permission portability | Follows binary | Tied to parent app |
| Setup complexity | Lower (one-time) | Higher (per client) |
| Production deployment | Easier (background service) | Harder (requires shell/terminal) |
| Recommended for | General use, multiple clients | Single-client embedding |

## Manual Permission Configuration

If the prompt doesn't appear or you need to modify permissions:

1. Open **System Settings**
2. Navigate to **Privacy & Security** → **Automation**
3. Find the entry:
   - **HTTP mode:** Look for `mail-mcp`
   - **STDIO mode:** Look for the parent application (Terminal, Claude, etc.)
4. Toggle the **Mail** checkbox to enable/disable
5. Restart the server or parent application

## Accessibility Permissions Workflow (for Rich Text Tools)

Certain tools that modify email content in an open window (e.g., `create_reply_draft`, `replace_reply_draft`, `create_outgoing_message`) require a higher level of permission called **Accessibility**. This permission is more sensitive than Automation and has a specific workflow to enable it correctly, especially when running the server as a `launchd` service.

If permission is not granted, you will receive an error message that directs you to the correct settings pane.

### The Grant-and-Restart Cycle

Due to how macOS caches permissions for running processes, you **must restart the server** after granting Accessibility permission for the first time.

1.  **Trigger the Prompt:** The first time you run a tool that requires Accessibility, it will fail, but it will open the System Settings app for you.
2.  **Grant Permission:** In `System Settings > Privacy & Security > Accessibility`, find `mail-mcp` in the list and **enable the toggle switch**.
3.  **Restart the Service:** This is the most important step. The running service is not aware of the permission change yet. You must restart it.

    ```bash
    # Restart the launchd service to apply the new permission
    ./mail-mcp launchd restart
    # (Or 'brew services restart mail-mcp' if installed via Homebrew)
    ```
4.  **Try Again:** Run the tool a second time. The new process will have the correct permissions, and the tool should now succeed.

This "grant-and-restart" cycle only needs to be done once.

## Troubleshooting

### "osascript execution failed: signal: killed"

This error means automation permissions are denied or not granted.

**Solution:**
1. Check System Settings → Privacy & Security → Automation
2. Verify the correct application/binary has Mail.app permission enabled
3. Restart the server

**Alternative - Reset and re-prompt:**
```bash
# Reset all automation permissions
tccutil reset AppleEvents

# Or reset for specific app (e.g., Terminal)
tccutil reset AppleEvents com.apple.Terminal
```

Then restart the server - macOS will prompt for permission again.

### Permission prompt doesn't appear

**For HTTP mode:**
- Try running the server from Terminal manually first
- The prompt appears when `osascript` is first executed, not at server startup

**For STDIO mode:**
- The parent application may need to be added to the Full Disk Access list first
- Try running from Terminal to trigger the prompt, then configure your MCP client

### Multiple permission entries

If you see many entries in System Settings → Automation:
- These likely came from using STDIO mode with different parent applications
- Switch to HTTP mode and remove the old entries
- Only the `mail-mcp` entry will be needed going forward

**Clean up old entries:**
```bash
# Reset all automation permissions
tccutil reset AppleEvents
```

Then run the server in HTTP mode to grant permission only to the binary.

### Resetting for Testing

To test permission prompts from scratch:

```bash
# 1. Reset all automation permissions
tccutil reset AppleEvents

# 2. Start server via launchd (will prompt for permission)
./mail-mcp launchd create

# 3. Click OK when prompted
# 4. Server is now ready with fresh permissions
```

## Recommendations

### For Most Users
**Use HTTP transport with launchd** for best results:

```bash
# Set up launchd service using the built-in subcommand
./mail-mcp launchd create

# Check it's running
launchctl list | grep com.github.dastrobu.mail-mcp

# View logs
tail -f /tmp/mail-mcp.log

# Restart the service (useful after granting Accessibility permissions)
./mail-mcp launchd restart

# To remove the service later
./mail-mcp launchd remove
```

**For quick testing from Terminal:**
```bash
# This will prompt for Terminal's permissions
./mail-mcp --transport=http --port=3000
```

### For Development/Testing
HTTP transport is still recommended, but STDIO can be useful for:
- Quick command-line testing
- Integration testing where the test runner is already permitted
- Debugging with direct stdin/stdout interaction

### For Production Deployments
**Always use HTTP transport:**
- Run as a background service (launchd, systemd)
- No dependency on terminal/shell permissions
- Easier to manage and monitor
- Better security isolation

## Security Considerations

### HTTP Transport
- Server listens on `localhost` by default (only local connections)
- Use `--host=0.0.0.0` with caution (allows network access)
- Consider firewall rules if binding to network interfaces
- No authentication built-in (design assumes localhost-only usage)

### STDIO Transport  
- Server only accessible via stdin/stdout pipes
- Parent process controls all access
- More restrictive but requires parent to have permissions

### Both Modes
- All operations are read-only
- Mail.app's own security and permissions still apply
- Server cannot send emails or modify message content
- Runs with user's Mail.app permissions (no privilege escalation)

## See Also

- [README.md](../README.md) - General usage and configuration
- [MCP Protocol Specification](https://modelcontextprotocol.io/) - Understanding MCP transports
- [Apple's JXA Documentation](https://developer.apple.com/library/archive/releasenotes/InterapplicationCommunication/RN-JavaScriptForAutomation/) - JXA and automation