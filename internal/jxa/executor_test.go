package jxa

import (
	"context"
	"encoding/json/v2"
	"strings"
	"testing"
)

func TestExecute_WrappedFormat(t *testing.T) {
	// Test script that returns data wrapped in data field (required format)
	script := `
function run(argv) {
	return JSON.stringify({
		success: true,
		data: {
			accounts: ["account1", "account2"],
			count: 2
		}
	});
}
`

	ctx := context.Background()
	result, err := Execute(ctx, script)

	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	// Result should be a map with accounts and count
	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Execute() result type = %T, want map[string]any", result)
	}

	// Verify accounts field exists
	accounts, ok := resultMap["accounts"]
	if !ok {
		t.Error("Execute() result missing 'accounts' field")
	}

	// Verify count field exists
	count, ok := resultMap["count"]
	if !ok {
		t.Error("Execute() result missing 'count' field")
	}

	// Verify count value
	if countFloat, ok := count.(float64); !ok || countFloat != 2 {
		t.Errorf("Execute() count = %v, want 2", count)
	}

	// Verify success field was removed
	if _, exists := resultMap["success"]; exists {
		t.Error("Execute() result should not contain 'success' field")
	}

	t.Logf("Result: %v", resultMap)
	t.Logf("Accounts: %v", accounts)
}

func TestExecute_MissingDataField(t *testing.T) {
	// Test script that returns unwrapped format (should fail)
	script := `
function run(argv) {
	return JSON.stringify({
		success: true,
		accounts: ["account1", "account2"],
		count: 2
	});
}
`

	ctx := context.Background()
	result, err := Execute(ctx, script)

	if err == nil {
		t.Fatal("Execute() error = nil, want error for missing data field")
	}

	if result != nil {
		t.Errorf("Execute() result = %v, want nil on error", result)
	}

	// Check that error message contains the key phrase
	if err == nil || !strings.Contains(err.Error(), "script output missing 'data' field") {
		t.Errorf("Execute() error should contain 'script output missing 'data' field', got: %v", err)
	}

	t.Logf("Error: %v", err)
}

func TestExecute_ScriptError(t *testing.T) {
	// Test script that returns an error
	script := `
function run(argv) {
	return JSON.stringify({
		success: false,
		error: "Something went wrong"
	});
}
`

	ctx := context.Background()
	result, err := Execute(ctx, script)

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	if result != nil {
		t.Errorf("Execute() result = %v, want nil on error", result)
	}

	// Check that error message contains the key phrase
	if err == nil || !strings.Contains(err.Error(), "JXA script error: Something went wrong") {
		t.Errorf("Execute() error should contain 'JXA script error: Something went wrong', got: %v", err)
	}

	t.Logf("Error: %v", err)
}

func TestExecute_WithArguments(t *testing.T) {
	// Test script that uses arguments
	script := `
function run(argv) {
	const arg1 = argv[0] || '';
	const arg2 = parseInt(argv[1]) || 0;

	return JSON.stringify({
		success: true,
		data: {
			arg1: arg1,
			arg2: arg2,
			argCount: argv.length
		}
	});
}
`

	ctx := context.Background()
	result, err := Execute(ctx, script, "test", "42")

	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("Execute() result type = %T, want map[string]any", result)
	}

	// Verify arg1
	if arg1, ok := resultMap["arg1"].(string); !ok || arg1 != "test" {
		t.Errorf("Execute() arg1 = %v, want 'test'", resultMap["arg1"])
	}

	// Verify arg2
	if arg2, ok := resultMap["arg2"].(float64); !ok || arg2 != 42 {
		t.Errorf("Execute() arg2 = %v, want 42", resultMap["arg2"])
	}

	// Verify argCount
	if argCount, ok := resultMap["argCount"].(float64); !ok || argCount != 2 {
		t.Errorf("Execute() argCount = %v, want 2", resultMap["argCount"])
	}

	t.Logf("Result: %v", resultMap)
}

func TestExecute_InvalidJSON(t *testing.T) {
	// Test script that returns invalid JSON
	script := `
function run(argv) {
	return "not valid json";
}
`

	ctx := context.Background()
	result, err := Execute(ctx, script)

	if err == nil {
		t.Fatal("Execute() error = nil, want JSON parse error")
	}

	if result != nil {
		t.Errorf("Execute() result = %v, want nil on error", result)
	}

	t.Logf("Error: %v", err)
}

func TestExecute_ContextCancellation(t *testing.T) {
	// Test context cancellation
	script := `
function run(argv) {
	// Simulate long-running script
	const start = Date.now();
	while (Date.now() - start < 5000) {
		// Busy wait
	}
	return JSON.stringify({success: true, done: true});
}
`

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	result, err := Execute(ctx, script)

	if err == nil {
		t.Fatal("Execute() error = nil, want context cancellation error")
	}

	if result != nil {
		t.Errorf("Execute() result = %v, want nil on error", result)
	}

	t.Logf("Error: %v", err)
}

func TestResult_JSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		wantJSON string
	}{
		{
			name: "success with data",
			result: Result{
				Success: true,
				Data: map[string]any{
					"count": 5,
					"items": []string{"a", "b"},
				},
			},
			wantJSON: `{"success":true,"data":{"count":5,"items":["a","b"]}}`,
		},
		{
			name: "success without data",
			result: Result{
				Success: true,
			},
			wantJSON: `{"success":true}`,
		},
		{
			name: "error",
			result: Result{
				Success: false,
				Error:   "something failed",
			},
			wantJSON: `{"success":false,"error":"something failed"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.result, json.Deterministic(true))
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			if string(got) != tt.wantJSON {
				t.Errorf("json.Marshal() = %s, want %s", string(got), tt.wantJSON)
			}
		})
	}
}

func TestExecute_MailAppNotRunning(t *testing.T) {
	// Test script that returns MAIL_APP_NOT_RUNNING error code
	script := `
function run(argv) {
	return JSON.stringify({
		success: false,
		error: "Mail.app is not running. Please start Mail.app and try again.",
		errorCode: "MAIL_APP_NOT_RUNNING"
	});
}
`

	ctx := context.Background()
	_, err := Execute(ctx, script)

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	// Verify the error message contains the expected text
	errMsg := err.Error()
	expectedPhrases := []string{
		"Mail.app is not running",
		"Please start Mail.app and try again",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Execute() error message missing expected phrase: %q\nGot: %s", phrase, errMsg)
		}
	}

	t.Logf("Error message: %v", err)
}

func TestExecute_MailAppNoPermissions(t *testing.T) {
	// Test script that returns MAIL_APP_NO_PERMISSIONS error code
	script := `
function run(argv) {
	return JSON.stringify({
		success: false,
		error: "Permission denied to access Mail.app. Please grant automation permissions.",
		errorCode: "MAIL_APP_NO_PERMISSIONS"
	});
}
`

	ctx := context.Background()
	_, err := Execute(ctx, script)

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}

	// Verify the error message contains the expected text
	errMsg := err.Error()
	expectedPhrases := []string{
		"Mail.app automation permission denied",
		"Please grant permission",
		"System Settings",
		"Privacy & Security",
		"Automation",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(errMsg, phrase) {
			t.Errorf("Execute() error message missing expected phrase: %q\nGot: %s", phrase, errMsg)
		}
	}

	t.Logf("Error message: %v", err)
}
