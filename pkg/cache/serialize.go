package cache

import (
	"encoding/json"
	"fmt"
)

// serialize converts an interface to JSON bytes
func serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	// Check if already bytes
	if bytes, ok := value.([]byte); ok {
		return bytes, nil
	}

	// Check if already a string
	if str, ok := value.(string); ok {
		return []byte(str), nil
	}

	// Marshal to JSON
	return json.Marshal(value)
}

// deserialize converts JSON bytes to the target type
func deserialize(data []byte, target interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data")
	}
	return json.Unmarshal(data, target)
}
