package cmd

import (
	"encoding/json"
	"fmt"
)

const (
	// MaxMetadataSize is the maximum size for extra_metadata in bytes (100KB)
	MaxMetadataSize = 100 * 1024
)

// ParseExtraMetadata parses and validates a JSON string for extra_metadata
func ParseExtraMetadata(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return nil, nil
	}

	// Check size limit
	if len(jsonStr) > MaxMetadataSize {
		return nil, fmt.Errorf("extra_metadata exceeds maximum size of %d bytes (got %d bytes)", MaxMetadataSize, len(jsonStr))
	}

	// Parse JSON
	var metadata map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
		return nil, fmt.Errorf("invalid JSON for extra_metadata: %w", err)
	}

	// Validate it's an object, not an array or primitive
	if metadata == nil {
		return nil, fmt.Errorf("extra_metadata must be a JSON object, not null")
	}

	return metadata, nil
}
