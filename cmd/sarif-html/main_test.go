package main

import (
	"encoding/json"
	"fmt"
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

func TestRunWritesTemplateDataOut(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.sarif")
	output := filepath.Join(dir, "report.html")
	templateDataOut := filepath.Join(dir, "template-data.json")

	if err := os.WriteFile(input, []byte(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "demo", "semanticVersion": "1.2.3" } },
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Template data message" },
						"locations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "src/App.kt" },
									"region": { "startLine": 11 }
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

	if err := run([]string{
		input,
		"--title", "Template Data",
		"--revision", "abc123",
		"--source-url-template", "https://example.test/{revision}/{path}{lineFragment}",
		"--template-data-out", templateDataOut,
		"--out", output,
	}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	raw, err := os.ReadFile(templateDataOut)
	if err != nil {
		t.Fatalf("read template data output: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatalf("unmarshal template data output: %v", err)
	}
	if data["schemaVersion"] != "sarif-html.template.v1" {
		t.Fatalf("expected schemaVersion sarif-html.template.v1, got %#v", data["schemaVersion"])
	}
	if data["title"] != "Template Data" {
		t.Fatalf("expected title Template Data, got %#v", data["title"])
	}
	reportData := data["report"].(map[string]any)
	findings := reportData["findings"].([]any)
	firstFinding := findings[0].(map[string]any)
	expectedLink := "https://example.test/abc123/src/App.kt#L11"
	if firstFinding["sourceLink"] != expectedLink {
		t.Fatalf("expected sourceLink %s, got %#v", expectedLink, firstFinding["sourceLink"])
	}
}

func TestRunDryRunValidatesTemplateWithoutWritingReportOrFailingThreshold(t *testing.T) {
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
						"level": "error",
						"message": { "text": "Dry run message" },
						"locations": []
					}
				]
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	if err := os.WriteFile(templatePath, []byte(`<html><body>{{ .SchemaVersion }} {{ .Report.Summary.Total }}</body></html>`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	if err := run([]string{input, "--template", templatePath, "--dry-run", "--out", output, "--fail-on", "warning"}); err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("expected dry-run not to write output, stat error: %v", err)
	}
}

func TestRunDryRunReturnsTemplateErrors(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.sarif")
	templatePath := filepath.Join(dir, "report.tmpl")

	if err := os.WriteFile(input, []byte(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "demo" } },
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Dry run message" },
						"locations": []
					}
				]
			}
		]
	}`), 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	if err := os.WriteFile(templatePath, []byte(`<html><body>{{ .Report.DoesNotExist }}</body></html>`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	err := run([]string{input, "--template", templatePath, "--dry-run"})
	if err == nil {
		t.Fatalf("expected dry-run to return template execution error")
	}
	if !strings.Contains(err.Error(), "render HTML report") {
		t.Fatalf("expected render error, got %v", err)
	}
}

func TestRunWritesAndReadsBaselineForFailOn(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.sarif")
	baselinePath := filepath.Join(dir, "baseline.json")
	output := filepath.Join(dir, "report.html")
	writeTestSARIF(t, input, testSARIFResult("KnownRule", "error", "known-fingerprint"))

	if err := run([]string{input, "--baseline-out", baselinePath, "--dry-run"}); err != nil {
		t.Fatalf("write baseline returned error: %v", err)
	}
	raw, err := os.ReadFile(baselinePath)
	if err != nil {
		t.Fatalf("read baseline: %v", err)
	}
	if !strings.Contains(string(raw), `"schemaVersion": "sarif-html.baseline.v1"`) {
		t.Fatalf("expected baseline schema version, got:\n%s", string(raw))
	}
	if !strings.Contains(string(raw), `"baselineState": "unchanged"`) {
		t.Fatalf("expected generated baseline entries to be unchanged, got:\n%s", string(raw))
	}

	if err := run([]string{input, "--baseline", baselinePath, "--out", output, "--fail-on", "error"}); err != nil {
		t.Fatalf("expected baseline finding not to fail threshold, got: %v", err)
	}
	html, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !strings.Contains(string(html), `data-baseline="true"`) {
		t.Fatalf("expected baseline marker in HTML")
	}
}

func TestRunFailsOnNewFindingWhenBaselineIsLoaded(t *testing.T) {
	dir := t.TempDir()
	baselineInput := filepath.Join(dir, "baseline-input.sarif")
	currentInput := filepath.Join(dir, "current.sarif")
	baselinePath := filepath.Join(dir, "baseline.json")
	output := filepath.Join(dir, "report.html")
	writeTestSARIF(t, baselineInput, testSARIFResult("KnownRule", "error", "known-fingerprint"))
	writeTestSARIF(t, currentInput,
		testSARIFResult("KnownRule", "error", "known-fingerprint")+","+
			testSARIFResult("NewRule", "error", "new-fingerprint"))

	if err := run([]string{baselineInput, "--baseline-out", baselinePath, "--dry-run"}); err != nil {
		t.Fatalf("write baseline returned error: %v", err)
	}

	err := run([]string{currentInput, "--baseline", baselinePath, "--out", output, "--fail-on", "error"})
	exitErr, ok := err.(exitError)
	if !ok {
		t.Fatalf("expected exitError for new finding, got %T: %v", err, err)
	}
	if exitErr.code != 2 {
		t.Fatalf("expected exit code 2, got %d", exitErr.code)
	}
}

func TestConfirmOverwriteDefaultsToNo(t *testing.T) {
	var prompt strings.Builder

	confirmed, err := confirmOverwrite("baseline.json", strings.NewReader("\n"), &prompt)
	if err != nil {
		t.Fatalf("confirmOverwrite returned error: %v", err)
	}
	if confirmed {
		t.Fatal("expected empty answer to keep existing baseline")
	}
	if !strings.Contains(prompt.String(), "[y/N]") {
		t.Fatalf("expected default-no prompt, got %q", prompt.String())
	}
}

func TestConfirmOverwriteAcceptsYes(t *testing.T) {
	var prompt strings.Builder

	confirmed, err := confirmOverwrite("baseline.json", strings.NewReader("yes\n"), &prompt)
	if err != nil {
		t.Fatalf("confirmOverwrite returned error: %v", err)
	}
	if !confirmed {
		t.Fatal("expected yes answer to overwrite baseline")
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

func writeTestSARIF(t *testing.T, path string, results string) {
	t.Helper()
	content := fmt.Sprintf(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "demo" } },
				"results": [%s]
			}
		]
	}`, results)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write SARIF: %v", err)
	}
}

func testSARIFResult(ruleID, level, fingerprint string) string {
	return fmt.Sprintf(`{
		"ruleId": %q,
		"level": %q,
		"message": { "text": %q },
		"locations": [
			{
				"physicalLocation": {
					"artifactLocation": { "uri": "src/App.kt" },
					"region": { "startLine": 4 }
				}
			}
		],
		"partialFingerprints": {
			"primaryLocationLineHash": %q
		}
	}`, ruleID, level, ruleID+" message", fingerprint)
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
