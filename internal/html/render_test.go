package html

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Smponi/sarifToHtml/internal/report"
)

func TestRenderIncludesFindingsAndSourceLinks(t *testing.T) {
	reportData := report.Report{
		Title:   "Demo",
		Sources: []string{"demo.sarif"},
		Summary: report.Summary{
			Total:      1,
			BySeverity: map[string]int{"warning": 1},
			ByTool:     map[string]int{"detekt": 1},
			ByRule:     map[string]int{"LongMethod": 1},
			ByFile:     map[string]int{"src/App.kt": 1},
		},
		Findings: []report.Finding{
			{
				ID:              "F0001",
				Source:          "demo.sarif",
				Tool:            "detekt",
				ToolVersion:     "1.23.0",
				RuleID:          "LongMethod",
				RuleName:        "Long Method",
				RuleDescription: "Method is too long.",
				Level:           "warning",
				Message:         "The method has 80 lines.",
				Path:            "src/App.kt",
				StartLine:       12,
				Fingerprint:     "fingerprint",
			},
		},
	}

	output, err := Render(reportData, Options{
		Title:    "Demo",
		RepoURL:  "https://github.com/acme/project",
		Revision: "abc123",
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	html := string(output)
	for _, expected := range []string{
		"Demo",
		"LongMethod",
		"The method has 80 lines.",
		"https://github.com/acme/project/blob/abc123/src/App.kt#L12",
		"Method is too long.",
		"All SARIF files",
		"All rules",
		"Compact dashboard",
		"Tool dashboards",
		`data-tool-section="detekt"`,
		`class="compact-row"`,
		`id="result-count"`,
		`data-source-chip="demo.sarif"`,
		`aria-pressed="false"`,
		`data-source="demo.sarif"`,
		`data-rule="LongMethod"`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected rendered HTML to contain %q", expected)
		}
	}
}

func TestRenderMarksBaselineFindingsHiddenByDefault(t *testing.T) {
	reportData := report.Report{
		Title:   "Baseline Demo",
		Sources: []string{"demo.sarif"},
		Findings: []report.Finding{
			{
				ID:            "F0001",
				Source:        "demo.sarif",
				Tool:          "detekt",
				RuleID:        "LongMethod",
				Level:         "error",
				Message:       "Known finding",
				Path:          "src/App.kt",
				StartLine:     12,
				Fingerprint:   "fingerprint",
				BaselineState: report.BaselineStateUnchanged,
			},
		},
	}
	report.RebuildSummary(&reportData)

	output, err := Render(reportData, Options{Title: "Baseline Demo"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	html := string(output)
	for _, expected := range []string{
		"Show baseline findings",
		`id="show-baseline"`,
		`class="finding hidden"`,
		`data-baseline="true"`,
		`data-baseline-state="unchanged"`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected rendered HTML to contain %q, got:\n%s", expected, html)
		}
	}
}

func TestRenderUsesCustomTemplateFile(t *testing.T) {
	dir := t.TempDir()
	templatePath := filepath.Join(dir, "report.tmpl")
	if err := os.WriteFile(templatePath, []byte(`<!doctype html>
<html>
<body>
  <p id="schema">{{ .SchemaVersion }}</p>
  <h1>{{ .Title }}</h1>
  {{ with index .Report.Findings 0 }}
  <a href="{{ sourceLink . }}">{{ line .StartLine }}</a>
  <div id="message">{{ .Message }}</div>
  {{ end }}
</body>
</html>`), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}

	reportData := report.Report{
		Summary: report.Summary{
			Total:      1,
			BySeverity: map[string]int{"warning": 1},
			ByTool:     map[string]int{"demo": 1},
			ByRule:     map[string]int{"Rule": 1},
			ByFile:     map[string]int{"src/App.kt": 1},
		},
		Findings: []report.Finding{
			{
				ID:        "F0001",
				Tool:      "demo",
				RuleID:    "Rule",
				Level:     "warning",
				Message:   `<script>alert("owned")</script>`,
				Path:      "src/App.kt",
				StartLine: 9,
			},
		},
	}

	output, err := Render(reportData, Options{
		Title:             "Custom Report",
		TemplatePath:      templatePath,
		Revision:          "abc123",
		SourceURLTemplate: "https://example.test/{revision}/{path}{lineFragment}",
	})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	html := string(output)
	for _, expected := range []string{
		TemplateDataVersion,
		"Custom Report",
		"https://example.test/abc123/src/App.kt#L9",
		"&lt;script&gt;alert(&#34;owned&#34;)&lt;/script&gt;",
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected rendered HTML to contain %q, got:\n%s", expected, html)
		}
	}
	if strings.Contains(html, `<script>alert("owned")</script>`) {
		t.Fatalf("expected custom template data to be HTML escaped")
	}
}

func TestRenderUsesCustomTemplateDirectory(t *testing.T) {
	dir := t.TempDir()
	partialsDir := filepath.Join(dir, "partials")
	if err := os.MkdirAll(partialsDir, 0o755); err != nil {
		t.Fatalf("create partials dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.tmpl"), []byte(`<!doctype html>
<html><body>{{ template "summary" . }}</body></html>`), 0o644); err != nil {
		t.Fatalf("write main template: %v", err)
	}
	if err := os.WriteFile(filepath.Join(partialsDir, "summary.tmpl"), []byte(`{{ define "summary" }}<strong>{{ .Report.Summary.Total }} total via {{ .SchemaVersion }}</strong>{{ end }}`), 0o644); err != nil {
		t.Fatalf("write partial template: %v", err)
	}

	output, err := Render(report.Report{
		Summary: report.Summary{Total: 3},
	}, Options{TemplatePath: dir})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	expected := "3 total via " + TemplateDataVersion
	if !strings.Contains(string(output), expected) {
		t.Fatalf("expected custom directory template output to contain %q, got:\n%s", expected, string(output))
	}
}

func TestMarshalTemplateDataUsesVersionedJSON(t *testing.T) {
	data := NewTemplateData(report.Report{
		Summary: report.Summary{
			Total:      1,
			BySeverity: map[string]int{"warning": 1},
			ByTool:     map[string]int{"demo": 1},
			ByRule:     map[string]int{"Rule": 1},
			ByFile:     map[string]int{"src/App.kt": 1},
		},
		Findings: []report.Finding{
			{ID: "F0001", Tool: "demo", RuleID: "Rule", Level: "warning", Path: "src/App.kt", StartLine: 3},
		},
	}, Options{
		Title:             "JSON Contract",
		Revision:          "abc123",
		SourceURLTemplate: "https://example.test/{revision}/{path}{lineFragment}",
	})

	output, err := MarshalTemplateData(data)
	if err != nil {
		t.Fatalf("MarshalTemplateData returned error: %v", err)
	}
	json := string(output)
	for _, expected := range []string{
		`"schemaVersion": "sarif-html.template.v1"`,
		`"title": "JSON Contract"`,
		`"sourceLink": "https://example.test/abc123/src/App.kt#L3"`,
		`"baseline"`,
		`"findings"`,
	} {
		if !strings.Contains(json, expected) {
			t.Fatalf("expected template data JSON to contain %q, got:\n%s", expected, json)
		}
	}
}

func TestCountBySource(t *testing.T) {
	counts := countBySource([]TemplateFinding{
		{Source: "detekt.sarif"},
		{Source: "detekt.sarif"},
		{Source: "semgrep.sarif"},
		{Source: " "},
	})
	if counts["detekt.sarif"] != 2 {
		t.Fatalf("expected detekt.sarif count 2, got %d", counts["detekt.sarif"])
	}
	if counts["semgrep.sarif"] != 1 {
		t.Fatalf("expected semgrep.sarif count 1, got %d", counts["semgrep.sarif"])
	}
	if counts["unknown-source"] != 1 {
		t.Fatalf("expected unknown-source count 1, got %d", counts["unknown-source"])
	}
}

func TestGroupedFindingsByTool(t *testing.T) {
	groups := groupedFindingsByTool([]TemplateFinding{
		{ID: "F0001", Tool: "semgrep"},
		{ID: "F0002", Tool: "detekt"},
		{ID: "F0003", Tool: "semgrep"},
		{ID: "F0004", Tool: " "},
	})
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Name != "semgrep" || groups[0].Count != 2 {
		t.Fatalf("expected semgrep group first with count 2, got %#v", groups[0])
	}
	if groups[1].Name != "detekt" {
		t.Fatalf("expected detekt second due alphabetical tie-break, got %s", groups[1].Name)
	}
	if groups[2].Name != "unknown-tool" {
		t.Fatalf("expected unknown-tool third, got %s", groups[2].Name)
	}
}

func TestGitLabSourceLink(t *testing.T) {
	link := sourceLink(report.Finding{Path: "src/App.kt", StartLine: 7}, Options{
		RepoURL:  "https://gitlab.com/acme/project.git",
		Revision: "main",
	})
	expected := "https://gitlab.com/acme/project/-/blob/main/src/App.kt#L7"
	if link != expected {
		t.Fatalf("expected %s, got %s", expected, link)
	}
}

func TestSourceLinkTemplate(t *testing.T) {
	link := sourceLink(report.Finding{Path: "src/App.kt", StartLine: 7, EndLine: 9}, Options{
		Revision:          "feature/sarif report",
		SourceURLTemplate: "https://example.test/repo/blob/{revision}/{path}{lineFragment}",
	})
	expected := "https://example.test/repo/blob/feature%2Fsarif%20report/src/App.kt#L7-L9"
	if link != expected {
		t.Fatalf("expected %s, got %s", expected, link)
	}
}

func TestSourceLinkTemplateRawPath(t *testing.T) {
	link := sourceLink(report.Finding{Path: "src/App.kt", StartLine: 7}, Options{
		SourceURLTemplate: "editor://open?file={pathRaw}&line={line}",
	})
	expected := "editor://open?file=src/App.kt&line=7"
	if link != expected {
		t.Fatalf("expected %s, got %s", expected, link)
	}
}
