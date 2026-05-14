package html

import (
	"strings"
	"testing"

	"sarif-html/internal/report"
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

func TestCountBySource(t *testing.T) {
	counts := countBySource([]report.Finding{
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
	groups := groupedFindingsByTool([]report.Finding{
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
