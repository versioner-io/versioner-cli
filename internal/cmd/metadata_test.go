package cmd

import (
	"strings"
	"testing"
)

func TestParseExtraMetadata(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		shouldErr bool
		errMsg    string
	}{
		{
			name:      "empty string returns nil",
			input:     "",
			shouldErr: false,
		},
		{
			name:      "valid JSON object",
			input:     `{"key": "value", "number": 42}`,
			shouldErr: false,
		},
		{
			name:      "valid nested JSON",
			input:     `{"docker_image": "myorg/api:1.2.3", "artifacts": ["binary", "docker"]}`,
			shouldErr: false,
		},
		{
			name:      "invalid JSON",
			input:     `{invalid json}`,
			shouldErr: true,
			errMsg:    "invalid JSON",
		},
		{
			name:      "JSON array not allowed",
			input:     `["item1", "item2"]`,
			shouldErr: true,
			errMsg:    "cannot unmarshal array",
		},
		{
			name:      "JSON null not allowed",
			input:     `null`,
			shouldErr: true,
			errMsg:    "must be a JSON object",
		},
		{
			name:      "JSON string not allowed",
			input:     `"just a string"`,
			shouldErr: true,
			errMsg:    "cannot unmarshal string",
		},
		{
			name:      "exceeds size limit",
			input:     `{"data": "` + strings.Repeat("x", MaxMetadataSize) + `"}`,
			shouldErr: true,
			errMsg:    "exceeds maximum size",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseExtraMetadata(tt.input)

			if tt.shouldErr {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errMsg)
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if tt.input == "" && result != nil {
					t.Errorf("Expected nil result for empty input, got: %v", result)
				}
				if tt.input != "" && result == nil {
					t.Errorf("Expected non-nil result for valid input, got nil")
				}
			}
		})
	}
}

func TestParseExtraMetadataValues(t *testing.T) {
	input := `{"string": "value", "number": 42, "bool": true, "nested": {"key": "val"}}`
	result, err := ParseExtraMetadata(input)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result["string"] != "value" {
		t.Errorf("Expected string value 'value', got %v", result["string"])
	}

	if result["number"] != float64(42) {
		t.Errorf("Expected number 42, got %v", result["number"])
	}

	if result["bool"] != true {
		t.Errorf("Expected bool true, got %v", result["bool"])
	}

	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Errorf("Expected nested to be a map, got %T", result["nested"])
	} else if nested["key"] != "val" {
		t.Errorf("Expected nested.key to be 'val', got %v", nested["key"])
	}
}
