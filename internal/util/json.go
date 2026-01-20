package util

import (
	"bytes"
	"encoding/json"
	"os"
)

// ReadJSONFile reads a JSON file, strips BOM if present, and unmarshals into the target
func ReadJSONFile(path string, target interface{}) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// strip BOM if present
	contentBytes = bytes.TrimPrefix(contentBytes, []byte("\xEF\xBB\xBF"))

	return json.Unmarshal(contentBytes, target)
}
