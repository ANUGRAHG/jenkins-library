package opentofu

import (
	"encoding/json"
	"fmt"
)

// Output struct represents the structure of an OpenTofu output.
type Output struct {
	Sensitive bool        `json:"sensitive"`
	Type      interface{} `json:"type"`
	Value     interface{} `json:"value"`
}

// ReadOutputs parses a JSON string of OpenTofu outputs and returns a map of variable names to their values.
func ReadOutputs(outputJSON string) (map[string]interface{}, error) {
	var rawOutputs map[string]Output
	if err := json.Unmarshal([]byte(outputJSON), &rawOutputs); err != nil {
		return nil, fmt.Errorf("failed to parse OpenTofu output JSON: %w", err)
	}

	outputs := make(map[string]interface{}, len(rawOutputs))
	for name, data := range rawOutputs {
		outputs[name] = data.Value
	}

	return outputs, nil
}
