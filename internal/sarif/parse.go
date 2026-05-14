package sarif

import (
	"encoding/json"
	"fmt"
	"io"
)

// Parse decodes and validates a SARIF 2.1.0 log from r.
func Parse(r io.Reader) (*Log, error) {
	var log Log
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&log); err != nil {
		return nil, fmt.Errorf("parse SARIF JSON: %w", err)
	}

	if log.Version == "" {
		return nil, fmt.Errorf("SARIF version is required")
	}
	if log.Version != "2.1.0" {
		return nil, fmt.Errorf("unsupported SARIF version %q: only 2.1.0 is supported", log.Version)
	}
	if len(log.Runs) == 0 {
		return nil, fmt.Errorf("SARIF log must contain at least one run")
	}

	return &log, nil
}
