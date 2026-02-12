package jxa

import (
	"context"
	"testing"
	"time"
)

func TestStartupCheck(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
		wantErr bool
	}{
		{
			name:    "successful startup check",
			timeout: 10 * time.Second,
			wantErr: false, // This may fail if Mail.app is not running
		},
		{
			name:    "startup check with short timeout",
			timeout: 100 * time.Millisecond,
			wantErr: true, // Should timeout
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			data, err := StartupCheck(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("StartupCheck() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Log the result for debugging
			if err != nil {
				t.Logf("StartupCheck() returned error: %v", err)
			} else {
				t.Logf("StartupCheck() passed successfully")
				if properties, ok := data["properties"].(map[string]any); ok {
					t.Logf("Retrieved %d properties from Mail.app", len(properties))
				}
			}
		})
	}
}

// TestStartupCheckScript verifies the startup check script structure
func TestStartupCheckScript(t *testing.T) {
	// Verify the script is not empty
	if startupCheckScript == "" {
		t.Error("startupCheckScript is empty")
	}

	// Verify the script contains expected components
	expectedComponents := []string{
		"function run(argv)",
		"Application('Mail')",
		"includeStandardAdditions",
		"running()",
		"success",
		"data",
	}

	for _, component := range expectedComponents {
		if !containsString(startupCheckScript, component) {
			t.Errorf("startupCheckScript missing expected component: %s", component)
		}
	}
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || len(substr) == 0 ||
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
