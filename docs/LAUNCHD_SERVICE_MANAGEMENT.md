# Launchd Service Management

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

**Implementation Pattern:**
```go
// options is *opts.LaunchdCreateCmd
cfg, err := launchd.DefaultConfig()
if err != nil {
    return err
}

// Override defaults from command-line flags
if options.Host != launchd.DefaultHost {
    cfg.Host = options.Host
}
if options.Port != launchd.DefaultPort {
    cfg.Port = options.Port
}
if options.Debug {
    cfg.Debug = options.Debug
}
if options.DisableRunAtLoad {
    cfg.RunAtLoad = false
}

// Create the service
return launchd.Create(cfg)
```

**Command Structure:**
- Main command: `launchd`
- Subcommands: `create`, `remove`, `restart`
- Examples: 
  - `./apple-mail-mcp launchd create` - Create service with automatic startup on login
  - `./apple-mail-mcp launchd create --port 9000` - Create service on custom port
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
The `.goreleaser.yaml
` includes a `post_install` script that automatically recreates the launchd service after `brew upgrade`, preserving all user settings:
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
    cmd << "launchd"
    cmd << "create"
    cmd << port_flag unless port_flag.empty?
    cmd << host_flag unless host_flag.empty?
    cmd << debug_flag unless debug_flag.empty?
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