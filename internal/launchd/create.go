package launchd

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/launchd.plist.tmpl
var plistTemplate string

// DefaultConfig returns the default launchd configuration
func DefaultConfig() (*Config, error) {
	// Get current executable path for fallback and version comparison
	currentExe, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("❌ failed to get executable path: %w", err)
	}

	// Try to find the binary using 'which' first (e.g., /opt/homebrew/bin/mail-mcp)
	// This ensures we use the symlinked version that will survive upgrades
	binaryPath, err := filepath.EvalSymlinks(currentExe)
	if err != nil {
		return nil, fmt.Errorf("❌ failed to resolve binary path: %w", err)
	}

	// Expand tilde in log paths
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("❌ failed to get user home directory: %w", err)
	}
	logPath := filepath.Join(home, "Library", "Logs", "com.github.dastrobu.mail-mcp", "mail-mcp.log")
	errPath := filepath.Join(home, "Library", "Logs", "com.github.dastrobu.mail-mcp", "mail-mcp.err")

	return &Config{
		BinaryPath: binaryPath,
		Host:       DefaultHost,
		Port:       DefaultPort,
		LogPath:    logPath,
		ErrPath:    errPath,
		RunAtLoad:  true, // Default: start service on login
	}, nil
}

// versionsMatch checks if two binary paths have the same version string
func versionsMatch(path1, path2 string) bool {
	version1 := getBinaryVersion(path1)
	version2 := getBinaryVersion(path2)

	// If either version is empty or "dev", skip version check
	// (happens during development/testing)
	if version1 == "" || version2 == "" || version1 == "dev" || version2 == "dev" {
		return true
	}

	return version1 == version2
}

// getBinaryVersion executes a binary with --version and extracts the version string
func getBinaryVersion(binaryPath string) string {
	cmd := exec.Command(binaryPath, "--version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil // Ignore stderr

	if err := cmd.Run(); err != nil {
		return ""
	}

	// Parse output like "mail-mcp version 0.1.0"
	output := stdout.String()
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		// Extract version from first line
		parts := strings.Fields(lines[0])
		if len(parts) >= 3 && parts[0] == "mail-mcp" && parts[1] == "version" {
			return parts[2]
		}
	}

	return ""
}

// createPlist creates the launchd plist file
func createPlist(cfg *Config) error {
	// Create LaunchAgents directory if it doesn't exist
	launchAgentsDir := filepath.Dir(PlistPath())
	if err := os.MkdirAll(launchAgentsDir, 0755); err != nil {
		return fmt.Errorf("❌ failed to create LaunchAgents directory: %w", err)
	}

	// Create log directory if it doesn't exist
	logDir := filepath.Dir(cfg.LogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("❌ failed to create log directory: %w", err)
	}

	// Parse and execute template
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return fmt.Errorf("❌ failed to parse plist template: %w", err)
	}

	// Create plist file
	plistPath := PlistPath()
	file, err := os.Create(plistPath)
	if err != nil {
		return fmt.Errorf("❌ failed to create plist file: %w", err)
	}
	defer file.Close()

	// Execute template with config
	data := struct {
		Label      string
		BinaryPath string
		Host       string
		Port       int
		LogPath    string
		ErrPath    string
		Debug      bool
		RunAtLoad  bool
	}{
		Label:      Label,
		BinaryPath: cfg.BinaryPath,
		Host:       cfg.Host,
		Port:       cfg.Port,
		LogPath:    cfg.LogPath,
		ErrPath:    cfg.ErrPath,
		Debug:      cfg.Debug,
		RunAtLoad:  cfg.RunAtLoad,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("❌ failed to write plist file: %w", err)
	}

	return nil
}

// load loads the service
func load() error {
	plistPath := PlistPath()
	cmd := exec.Command("launchctl", "load", plistPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("❌ failed to load service: %w (output: %s)", err, string(output))
	}
	return nil
}

// Create performs the complete launchd service creation
func Create(cfg *Config) error {
	fmt.Println("Apple Mail MCP Server - launchd Create")
	fmt.Println("=======================================")
	fmt.Println()

	// Unload existing service if present
	if IsLoaded() {
		fmt.Println("Service is already loaded. Unloading...")
		if err := unload(); err != nil {
			return fmt.Errorf("❌ failed to unload existing service: %w", err)
		}
	}

	// Create plist file
	fmt.Printf("Creating launchd plist at %s\n", PlistPath())
	if err := createPlist(cfg); err != nil {
		return err
	}

	// Load service
	fmt.Println("Loading service...")
	if err := load(); err != nil {
		return err
	}

	// Wait for service to start
	time.Sleep(2 * time.Second)

	// Check if service is running
	if IsLoaded() {
		fmt.Println()
		fmt.Println("✅ Service successfully loaded!")
		fmt.Println()
		fmt.Println("Configuration:")
		fmt.Printf("  Binary: %s\n", cfg.BinaryPath)
		fmt.Printf("  Endpoint: http://%s:%d\n", cfg.Host, cfg.Port)
		fmt.Printf("  Logs: %s\n", cfg.LogPath)
		fmt.Printf("  Errors: %s\n", cfg.ErrPath)
		fmt.Println()
		fmt.Println("On first run, macOS will prompt for automation permissions.")
		fmt.Println("Click OK to grant permission to the mail-mcp binary.")
		fmt.Println()
		fmt.Println("Useful commands:")
		fmt.Printf("  View logs:   tail -f %s %s\n", cfg.LogPath, cfg.ErrPath)
		fmt.Printf("  Stop:        launchctl stop %s\n", Label)
		fmt.Printf("  Restart:     launchctl kickstart -k gui/$(id -u)/%s\n", Label)
		fmt.Printf("  Unload:      launchctl unload %s\n", PlistPath())
		fmt.Println()
		fmt.Printf("Configure your MCP client to connect to: http://%s:%d\n", cfg.Host, cfg.Port)
		return nil
	}

	fmt.Println()
	fmt.Println("⚠️  Service loaded but may not be running.")
	fmt.Println("Check logs for errors:")
	fmt.Printf("  tail -f %s\n", cfg.ErrPath)
	fmt.Println()
	fmt.Println("💡 Hint: To enable debug logging, recreate the service with:")
	if cfg.Debug {
		fmt.Printf("  ./mail-mcp --port=%d launchd create\n", cfg.Port)
	} else {
		fmt.Printf("  ./mail-mcp --port=%d --debug launchd create\n", cfg.Port)
	}
	return fmt.Errorf("⚠️ service loaded but not running")
}
