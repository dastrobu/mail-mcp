package tools

import (
	"reflect"

	"github.com/google/jsonschema-go/jsonschema"
)

// GenerateSchema generates a JSON schema for the given type T and fixes nullable fields
// to be compatible with Gemini API (using "nullable": true instead of type arrays).
func GenerateSchema[T any]() *jsonschema.Schema {
	schema, err := jsonschema.ForType(reflect.TypeFor[T](), &jsonschema.ForOptions{})
	if err != nil {
		// Should not happen for valid Go types used in tools
		panic(err)
	}
	fixSchema(schema)
	return schema
}

// fixSchema recursively modifies the schema to replace ["type", "null"] with "type" and nullable: true
func fixSchema(s *jsonschema.Schema) {
	if s == nil {
		return
	}

	// Check if this node needs fixing (type is array with "null")
	if len(s.Types) > 0 {
		hasNull := false
		var otherTypes []string
		for _, t := range s.Types {
			if t == "null" {
				hasNull = true
			} else {
				otherTypes = append(otherTypes, t)
			}
		}

		// If we have ["someType", "null"], convert to "someType" with nullable: true
		if hasNull && len(otherTypes) == 1 {
			s.Type = otherTypes[0]
			s.Types = nil
			if s.Extra == nil {
				s.Extra = make(map[string]any)
			}
			s.Extra["nullable"] = true
		}
	}

	// Recursively traverse children
	for _, prop := range s.Properties {
		fixSchema(prop)
	}
	if s.Items != nil {
		fixSchema(s.Items)
	}
	for _, item := range s.ItemsArray {
		fixSchema(item)
	}
	for _, def := range s.Definitions {
		fixSchema(def)
	}
	for _, def := range s.Defs {
		fixSchema(def)
	}
	if s.AdditionalProperties != nil {
		fixSchema(s.AdditionalProperties)
	}
	for _, sub := range s.OneOf {
		fixSchema(sub)
	}
	for _, sub := range s.AnyOf {
		fixSchema(sub)
	}
	for _, sub := range s.AllOf {
		fixSchema(sub)
	}
	if s.Not != nil {
		fixSchema(s.Not)
	}
	if s.If != nil {
		fixSchema(s.If)
	}
	if s.Then != nil {
		fixSchema(s.Then)
	}
	if s.Else != nil {
		fixSchema(s.Else)
	}
}
