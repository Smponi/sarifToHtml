package sarif

import (
	"strings"
	"testing"
)

func TestParseValidSARIF(t *testing.T) {
	log, err := Parse(strings.NewReader(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "detekt" } },
				"results": []
			}
		]
	}`))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if log.Version != "2.1.0" {
		t.Fatalf("unexpected version: %s", log.Version)
	}
	if len(log.Runs) != 1 {
		t.Fatalf("expected one run, got %d", len(log.Runs))
	}
}

func TestParseRejectsUnsupportedVersion(t *testing.T) {
	_, err := Parse(strings.NewReader(`{
		"version": "2.0.0",
		"runs": [{ "tool": { "driver": { "name": "tool" } } }]
	}`))
	if err == nil {
		t.Fatal("expected unsupported version error")
	}
}

func TestParseRequiresRuns(t *testing.T) {
	_, err := Parse(strings.NewReader(`{ "version": "2.1.0", "runs": [] }`))
	if err == nil {
		t.Fatal("expected runs validation error")
	}
}
