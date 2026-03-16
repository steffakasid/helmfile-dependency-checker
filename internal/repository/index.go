package repository

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseIndexYAML parses a raw Helm index.yaml into an Index.
func ParseIndexYAML(data []byte) (*Index, error) {
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index yaml: %w", err)
	}

	jsonBytes, err := reEncodeJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to re-encode index: %w", err)
	}

	var ij indexJSON
	if err := unmarshalJSON(jsonBytes, &ij); err != nil {
		return nil, fmt.Errorf("failed to unmarshal index entries: %w", err)
	}

	return fromJSON(ij)
}
