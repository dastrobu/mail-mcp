package launchd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	// Label is the launchd service identifier
	Label = "com.github.dastrobu.mail-mcp"

	// PlistFilename is the name of the plist file
	PlistFilename = Label + ".plist"

	// DefaultPort is the default HTTP port
	DefaultPort = 8787

	// DefaultHost is the default HTTP host
	DefaultHost = "localhost"

	// DefaultLogPath is the default log file path
	DefaultLogPath = "~/Library/Logs/com.github.dastrobu.mail-mcp/mail-mcp.log"

	// DefaultErrPath is the default error file path
	DefaultErrPath = "~/Library/Logs/com.github.dastrobu.mail-mcp/mail-mcp.err"
)

// Config holds the launchd service configuration
type Config struct {
	BinaryPath string
	Host       string
	Port       int
	LogPath    string
	ErrPath    string
	Debug      bool
	RunAtLoad  bool
}

// PlistPath returns the full path to the plist file
func PlistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", PlistFilename)
}

// IsLoaded checks if the service is currently loaded
func IsLoaded() bool {
	cmd := exec.Command("launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), Label)
}
