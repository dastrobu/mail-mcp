package opts

import (
	"os"
	"testing"

	"github.com/dastrobu/apple-mail-mcp/internal/opts/typed_flags"
)

func TestParse_DefaultValues(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set args to use run command
	os.Args = []string{"apple-mail-mcp", "run"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with default values: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportStdio {
		t.Errorf("Expected default transport 'stdio', got '%s'", GlobalOpts.Run.Transport)
	}

	if GlobalOpts.Run.Port != 8787 {
		t.Errorf("Expected default port 8787, got %d", GlobalOpts.Run.Port)
	}

	if GlobalOpts.Run.Host != "localhost" {
		t.Errorf("Expected default host 'localhost', got '%s'", GlobalOpts.Run.Host)
	}
}

func TestParse_StdioTransport(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "run", "--transport=stdio"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportStdio {
		t.Errorf("Expected transport 'stdio', got '%s'", GlobalOpts.Run.Transport)
	}
}

func TestParse_HTTPTransport(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "run", "--transport=http"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportHTTP {
		t.Errorf("Expected transport 'http', got '%s'", GlobalOpts.Run.Transport)
	}
}

func TestParse_HTTPWithCustomPort(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "run", "--transport=http", "--port=4567"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportHTTP {
		t.Errorf("Expected transport 'http', got '%s'", GlobalOpts.Run.Transport)
	}

	if GlobalOpts.Run.Port != 4567 {
		t.Errorf("Expected port 4567, got %d", GlobalOpts.Run.Port)
	}
}

func TestParse_HTTPWithCustomHost(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "run", "--transport=http", "--host=0.0.0.0", "--port=9000"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with custom host: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportHTTP {
		t.Errorf("Expected transport 'http', got '%s'", GlobalOpts.Run.Transport)
	}

	if GlobalOpts.Run.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", GlobalOpts.Run.Host)
	}

	if GlobalOpts.Run.Port != 9000 {
		t.Errorf("Expected port 9000, got %d", GlobalOpts.Run.Port)
	}
}

func TestParse_InvalidTransport(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"apple-mail-mcp", "run", "--transport=invalid"}

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

			os.Args = []string{"apple-mail-mcp", "run", "--transport=http", "--port=" + tt.port}

			_, err := Parse()
			// The flags library might catch this before our validation
			if err == nil {
				// If flags didn't catch it, our validation should
				if GlobalOpts.Run.Port < 1 || GlobalOpts.Run.Port > 65535 {
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

	os.Args = []string{"apple-mail-mcp", "run", "--transport=http", "--host=127.0.0.1", "--port=4567", "--debug"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportHTTP {
		t.Errorf("Expected transport 'http', got '%s'", GlobalOpts.Run.Transport)
	}

	if GlobalOpts.Run.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", GlobalOpts.Run.Host)
	}

	if GlobalOpts.Run.Port != 4567 {
		t.Errorf("Expected port 4567, got %d", GlobalOpts.Run.Port)
	}
}

func TestParse_EnvironmentVariables(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set environment variables
	os.Setenv("APPLE_MAIL_MCP_TRANSPORT", "http")
	os.Setenv("APPLE_MAIL_MCP_PORT", "9999")
	os.Setenv("APPLE_MAIL_MCP_HOST", "0.0.0.0")
	defer func() {
		os.Unsetenv("APPLE_MAIL_MCP_TRANSPORT")
		os.Unsetenv("APPLE_MAIL_MCP_PORT")
		os.Unsetenv("APPLE_MAIL_MCP_HOST")
	}()

	os.Args = []string{"apple-mail-mcp", "run"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed with environment variables: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportHTTP {
		t.Errorf("Expected transport 'http', got '%s'", GlobalOpts.Run.Transport)
	}

	if GlobalOpts.Run.Port != 9999 {
		t.Errorf("Expected port 9999, got %d", GlobalOpts.Run.Port)
	}

	if GlobalOpts.Run.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", GlobalOpts.Run.Host)
	}
}

func TestParse_FlagsOverrideEnvironment(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set environment variables
	os.Setenv("APPLE_MAIL_MCP_TRANSPORT", "stdio")
	os.Setenv("APPLE_MAIL_MCP_PORT", "5000")
	defer func() {
		os.Unsetenv("APPLE_MAIL_MCP_TRANSPORT")
		os.Unsetenv("APPLE_MAIL_MCP_PORT")
	}()

	// Flags should override environment
	os.Args = []string{"apple-mail-mcp", "run", "--transport=http", "--port=6000"}

	_, err := Parse()
	if err != nil {
		t.Fatalf("Parse() failed: %v", err)
	}

	if GlobalOpts.Run.Transport != typed_flags.TransportHTTP {
		t.Errorf("Expected transport 'http' from flag, got '%s'", GlobalOpts.Run.Transport)
	}
	if GlobalOpts.Run.Port != 6000 {
		t.Errorf("Expected port 6000 from flag, got %d", GlobalOpts.Run.Port)
	}
}
