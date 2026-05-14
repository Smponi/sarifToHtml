package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunAcceptsFlagsAfterInput(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.sarif")
	output := filepath.Join(dir, "report.html")

	if err := os.WriteFile(input, []byte(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "demo" } },
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Message" },
						"locations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "src/App.kt" },
									"region": { "startLine": 4 }
								}
							}
						]
					}
				]
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := run([]string{input, "--title", "Demo", "--out", output}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if _, err := os.Stat(output); err != nil {
		t.Fatalf("expected output file: %v", err)
	}
}

func TestRunReturnsExitErrorForFailOn(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.sarif")
	output := filepath.Join(dir, "report.html")

	if err := os.WriteFile(input, []byte(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "demo" } },
				"results": [
					{
						"ruleId": "Rule",
						"level": "error",
						"message": { "text": "Message" },
						"locations": []
					}
				]
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	err := run([]string{input, "--out", output, "--fail-on", "warning"})
	exitErr, ok := err.(exitError)
	if !ok {
		t.Fatalf("expected exitError, got %T: %v", err, err)
	}
	if exitErr.code != 2 {
		t.Fatalf("expected exit code 2, got %d", exitErr.code)
	}
	if _, statErr := os.Stat(output); statErr != nil {
		t.Fatalf("expected report to be written before fail-on exits: %v", statErr)
	}
}

func TestRunHydratesMissingSnippetFromSourceRoot(t *testing.T) {
	dir := t.TempDir()
	sourceDir := filepath.Join(dir, "src")
	input := filepath.Join(dir, "input.sarif")
	output := filepath.Join(dir, "report.html")

	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatalf("create source dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "App.kt"), []byte("fun main() {\n    println(\"loaded from source\")\n}\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(input, []byte(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "demo" } },
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Message" },
						"locations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "App.kt" },
									"region": { "startLine": 2 }
								}
							}
						]
					}
				]
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := run([]string{input, "--source-root", sourceDir, "--out", output}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	html, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(html), "loaded from source") {
		t.Fatalf("expected hydrated snippet in output")
	}
}

func TestRunAcceptsCustomTemplate(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.sarif")
	templatePath := filepath.Join(dir, "report.tmpl")
	output := filepath.Join(dir, "report.html")

	if err := os.WriteFile(input, []byte(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "demo" } },
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Custom template message" },
						"locations": []
					}
				]
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	if err := os.WriteFile(templatePath, []byte(`<html><body>{{ .Title }}: {{ .Report.Summary.Total }} finding(s) from {{ .SchemaVersion }}</body></html>`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	if err := run([]string{input, "--template", templatePath, "--title", "Team Report", "--out", output}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	html, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(html), "Team Report: 1 finding(s) from sarif-html.template.v1") {
		t.Fatalf("expected custom template output, got:\n%s", string(html))
	}
}

func TestSourceLinkOptionsDetectsGitHubActions(t *testing.T) {
	t.Setenv("GITHUB_SERVER_URL", "https://github.com")
	t.Setenv("GITHUB_REPOSITORY", "acme/project")
	t.Setenv("GITHUB_SHA", "abc123")

	options := sourceLinkOptions("", "", "")
	expected := "https://github.com/acme/project/blob/{revision}/{path}{lineFragment}"
	if options.SourceURLTemplate != expected {
		t.Fatalf("expected template %s, got %s", expected, options.SourceURLTemplate)
	}
	if options.Revision != "abc123" {
		t.Fatalf("expected revision abc123, got %s", options.Revision)
	}
}

func TestSourceLinkOptionsDetectsGitLabCI(t *testing.T) {
	t.Setenv("CI_PROJECT_URL", "https://gitlab.com/acme/project")
	t.Setenv("CI_COMMIT_SHA", "def456")

	options := sourceLinkOptions("", "", "")
	expected := "https://gitlab.com/acme/project/-/blob/{revision}/{path}{lineFragment}"
	if options.SourceURLTemplate != expected {
		t.Fatalf("expected template %s, got %s", expected, options.SourceURLTemplate)
	}
	if options.Revision != "def456" {
		t.Fatalf("expected revision def456, got %s", options.Revision)
	}
}

func TestSourceLinkOptionsDoesNotOverrideExplicitTemplate(t *testing.T) {
	t.Setenv("GITHUB_SERVER_URL", "https://github.com")
	t.Setenv("GITHUB_REPOSITORY", "acme/project")
	t.Setenv("GITHUB_SHA", "abc123")

	options := sourceLinkOptions("", "", "https://example.test/{path}")
	if !strings.Contains(options.SourceURLTemplate, "example.test") {
		t.Fatalf("expected explicit template, got %s", options.SourceURLTemplate)
	}
}
