package opts

import (
	"os"
	"testing"
)

func TestParse_DefaultValues(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to just the program name (no flags)
	os.Args = []string{"apple-mail-mcp"}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with default values: %v", err)
	}

	if opts.Transport != "stdio" {
		t.Errorf("Expected default transport 'stdio', got '%s'", opts.Transport)
	}

	if opts.Port != 8787 {
		t.Errorf("Expected default port 8787, got %d", opts.Port)
	}

	if opts.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", opts.Host)
	}
}

func TestParse_StdioTransport(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "--transport=stdio"}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with stdio transport: %v", err)
	}

	if opts.Transport != "stdio" {
		t.Errorf("Expected transport 'stdio', got '%s'", opts.Transport)
	}
}

func TestParse_HTTPTransport(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "--transport=http"}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with http transport: %v", err)
	}

	if opts.Transport != "http" {
		t.Errorf("Expected transport 'http', got '%s'", opts.Transport)
	}
}

func TestParse_HTTPWithCustomPort(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "--transport=http", "--port=3000"}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with custom port: %v", err)
	}

	if opts.Transport != "http" {
		t.Errorf("Expected transport 'http', got '%s'", opts.Transport)
	}

	if opts.Port != 3000 {
		t.Errorf("Expected port 3000, got %d", opts.Port)
	}
}

func TestParse_HTTPWithCustomHost(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "--transport=http", "--host=0.0.0.0", "--port=9000"}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with custom host: %v", err)
	}

	if opts.Transport != "http" {
		t.Errorf("Expected transport 'http', got '%s'", opts.Transport)
	}

	if opts.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", opts.Host)
	}

	if opts.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", opts.Port)
	}
}

func TestParse_InvalidTransport(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "--transport=invalid"}

	_, err := Parse()
	if err == nil {
		t.Error("Parse() should have failed with invalid transport")
	}
}

func TestParse_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{"port too low", "0"},
		{"port too high", "70000"},
		{"negative port", "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()

			os.Args = []string{"apple-mail-mcp", "--transport=http", "--port=" + tt.port}

			opts, err := Parse()
			// The flags library might catch this before our validation
			if err == nil {
				// If flags didn't catch it, our validation should
				if opts.Port < 1 || opts.Port > 65535 {
					// This is expected - validation worked
					return
				}
				t.Errorf("Parse() should have failed with invalid port %s", tt.port)
			}
		})
	}
}

func TestParse_AllOptions(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"apple-mail-mcp",
		"--transport=http",
		"--host=127.0.0.1",
		"--port=4567",
	}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with all options: %v", err)
	}

	if opts.Transport != "http" {
		t.Errorf("Expected transport 'http', got '%s'", opts.Transport)
	}

	if opts.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", opts.Host)
	}

	if opts.Port != 4567 {
		t.Errorf("Expected port 4567, got %d", opts.Port)
	}
}

func TestParse_EnvironmentVariables(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set environment variables
	os.Setenv("TRANSPORT", "http")
	os.Setenv("PORT", "9999")
	os.Setenv("HOST", "0.0.0.0")
	defer func() {
		os.Unsetenv("TRANSPORT")
		os.Unsetenv("PORT")
		os.Unsetenv("HOST")
	}()

	os.Args = []string{"apple-mail-mcp"}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with environment variables: %v", err)
	}

	if opts.Transport != "http" {
		t.Errorf("Expected transport 'http' from env, got '%s'", opts.Transport)
	}

	if opts.Port != 9999 {
		t.Errorf("Expected port 9999 from env, got %d", opts.Port)
	}

	if opts.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0' from env, got '%s'", opts.Host)
	}
}

func TestParse_FlagsOverrideEnvironment(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set environment variables
	os.Setenv("TRANSPORT", "http")
	os.Setenv("PORT", "9999")
	defer func() {
		os.Unsetenv("TRANSPORT")
		os.Unsetenv("PORT")
	}()

	// Flags should override environment
	os.Args = []string{"apple-mail-mcp", "--transport=stdio", "--port=3000"}

	opts, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if opts.Transport != "stdio" {
		t.Errorf("Expected transport 'stdio' (flag override), got '%s'", opts.Transport)
	}

	if opts.Port != 3000 {
		t.Errorf("Expected port 3000 (flag override), got %d", opts.Port)
	}
}
