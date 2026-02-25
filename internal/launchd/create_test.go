package launchd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetBinaryVersion(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantNonEmpty bool
	}{
		{
			name:         "nonexistent binary",
			path:         "/tmp/nonexistent-binary-xyz",
			wantNonEmpty: false,
		},
		{
			name:         "invalid binary",
			path:         "/dev/null",
			wantNonEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version := getBinaryVersion(tt.path)
			if tt.wantNonEmpty && version == "" {
				t.Errorf("getBinaryVersion() = empty, want non-empty")
			}
			if !tt.wantNonEmpty && version != "" {
				t.Errorf("getBinaryVersion() = %q, want empty", version)
			}
		})
	}
}

func TestVersionsMatch(t *testing.T) {
	tests := []struct {
		name  string
		path1 string
		path2 string
		want  bool
	}{
		{
			name:  "both paths are same",
			path1: "/tmp/test",
			path2: "/tmp/test",
			want:  true, // Will return true due to "dev" version
		},
		{
			name:  "nonexistent binaries default to match",
			path1: "/tmp/nonexistent1",
			path2: "/tmp/nonexistent2",
			want:  true, // Empty versions are considered matching
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := versionsMatch(tt.path1, tt.path2)
			if got != tt.want {
				t.Errorf("versionsMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig() error = %v", err)
	}

	// Verify basic fields are set
	if cfg.BinaryPath == "" {
		t.Error("BinaryPath is empty")
	}

	if cfg.Host == "" {
		t.Error("Host is empty")
	}

	if cfg.Port == 0 {
		t.Error("Port is 0")
	}

	if cfg.LogPath == "" {
		t.Error("LogPath is empty")
	}

	if cfg.ErrPath == "" {
		t.Error("ErrPath is empty")
	}

	// Verify binary path exists
	if _, err := os.Stat(cfg.BinaryPath); os.IsNotExist(err) {
		t.Errorf("BinaryPath does not exist: %s", cfg.BinaryPath)
	}

	// Verify binary path is absolute
	if !filepath.IsAbs(cfg.BinaryPath) {
		t.Errorf("BinaryPath is not absolute: %s", cfg.BinaryPath)
	}

	// Verify log paths are in expected location
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home dir: %v", err)
	}

	expectedLogDir := filepath.Join(home, "Library", "Logs", "com.github.dastrobu.mail-mcp")
	if !filepath.HasPrefix(cfg.LogPath, expectedLogDir) {
		t.Errorf("LogPath not in expected directory: got %s, want prefix %s", cfg.LogPath, expectedLogDir)
	}

	if !filepath.HasPrefix(cfg.ErrPath, expectedLogDir) {
		t.Errorf("ErrPath not in expected directory: got %s, want prefix %s", cfg.ErrPath, expectedLogDir)
	}

	// Verify default values
	if cfg.Host != DefaultHost {
		t.Errorf("Host = %s, want %s", cfg.Host, DefaultHost)
	}

	if cfg.Port != DefaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, DefaultPort)
	}
}

func TestDefaultConfig_BinaryPath(t *testing.T) {
	cfg, err := DefaultConfig()
	if err != nil {
		t.Fatalf("DefaultConfig() error = %v", err)
	}

	t.Logf("BinaryPath: %s", cfg.BinaryPath)

	// If 'which' found a binary, it should be at /opt/homebrew/bin or similar
	// If not found, it should be the current executable path
	if filepath.Base(cfg.BinaryPath) != "mail-mcp" &&
		!strings.HasSuffix(cfg.BinaryPath, ".test") {
		t.Errorf("BinaryPath basename unexpected: %s", filepath.Base(cfg.BinaryPath))
	}
}
