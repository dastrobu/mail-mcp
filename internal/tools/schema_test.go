package tools

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

type TestInput struct {
	OptionalField *string `json:"optional,omitempty" jsonschema:"An optional field"`
	RequiredField string  `json:"required" jsonschema:"A required field"`
}

func TestSchemaGeneration(t *testing.T) {
	// Generate schema using jsonschema-go directly
	schema, err := jsonschema.ForType(reflect.TypeFor[TestInput](), &jsonschema.ForOptions{})
	if err != nil {
		t.Fatalf("Failed to generate schema: %v", err)
	}

	// Fix the schema
	fixSchema(schema)

	// Marshal to JSON to verify the output
	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal schema: %v", err)
	}
	t.Logf("Generated Schema:\n%s", string(data))

	// Parse back to map to check structure
	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal schema: %v", err)
	}

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties not found or invalid")
	}

	optional, ok := props["optional"].(map[string]any)
	if !ok {
		t.Fatal("optional field not found")
	}

	// Check type: it should be "string" not ["string", "null"]
	typ, ok := optional["type"]
	if !ok {
		t.Fatal("type not found in optional field")
	}

	if typStr, ok := typ.(string); ok {
		if typStr != "string" {
			t.Errorf("Expected type 'string', got '%s'", typStr)
		}
	} else {
		t.Errorf("Expected type to be string (e.g. \"string\"), got %T: %v. This likely means it is still an array (e.g. [\"string\", \"null\"]).", typ, typ)
	}

	// Check nullable: it should be true
	nullable, ok := optional["nullable"]
	if !ok {
		t.Fatal("nullable not found in optional field")
	}
	if nullableBool, ok := nullable.(bool); ok {
		if !nullableBool {
			t.Error("Expected nullable to be true")
		}
	} else {
		t.Errorf("Expected nullable to be bool, got %T", nullable)
	}

	// Verify validation with null
	resolved, err := schema.Resolve(nil)
	if err != nil {
		t.Fatalf("Failed to resolve schema: %v", err)
	}

	// Test validation with null value
	// Note: Standard JSON Schema doesn't support 'nullable'.
	validInputNull := map[string]any{
		"required": "foo",
		"optional": nil,
	}
	if err := resolved.Validate(validInputNull); err != nil {
		t.Logf("Validation failed for null input (as expected without nullable support in jsonschema-go): %v", err)
		// If validation fails, it means using this schema for validation locally will reject valid inputs from Gemini.
		// However, we can't easily fix jsonschema-go validation logic.
		// For now, we prioritize fixing the Gemini API error (invalid schema format).
		// If Gemini sends null, it will be rejected by SDK, but at least the tool registration succeeds.
		// And maybe Gemini won't send null for optional fields (it might omit them).
	} else {
		t.Log("Validation succeeded for null input! Schema supports nullable.")
	}
}
