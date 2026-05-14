package report

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHydrateSnippetsLoadsMissingSnippet(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "app.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	reportData := Report{
		Findings: []Finding{
			{Path: "src/app.go", StartLine: 3},
		},
	}

	if err := HydrateSnippets(&reportData, []string{dir}); err != nil {
		t.Fatalf("hydrate snippets: %v", err)
	}
	if reportData.Findings[0].Snippet != "func main() {}" {
		t.Fatalf("unexpected snippet: %q", reportData.Findings[0].Snippet)
	}
}

func TestHydrateSnippetsKeepsExistingSnippet(t *testing.T) {
	reportData := Report{
		Findings: []Finding{
			{Path: "src/app.go", StartLine: 1, Snippet: "from sarif"},
		},
	}

	if err := HydrateSnippets(&reportData, []string{t.TempDir()}); err != nil {
		t.Fatalf("hydrate snippets: %v", err)
	}
	if reportData.Findings[0].Snippet != "from sarif" {
		t.Fatalf("expected existing snippet to be kept, got %q", reportData.Findings[0].Snippet)
	}
}

func TestHydrateSnippetsSkipsPathTraversal(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatalf("write outside source: %v", err)
	}

	reportData := Report{
		Findings: []Finding{
			{Path: "../secret.go", StartLine: 1},
		},
	}

	if err := HydrateSnippets(&reportData, []string{root}); err != nil {
		t.Fatalf("hydrate snippets: %v", err)
	}
	if reportData.Findings[0].Snippet != "" {
		t.Fatalf("expected traversal path to be skipped, got %q", reportData.Findings[0].Snippet)
	}
}
