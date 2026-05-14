package html

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"sort"
	"strings"
	"time"

	"sarif-html/internal/report"
)

// Options controls report metadata and source link generation.
type Options struct {
	Title             string
	RepoURL           string
	Revision          string
	SourceURLTemplate string
}

// Render turns reportData into a self-contained HTML document.
func Render(reportData report.Report, options Options) ([]byte, error) {
	if options.Title == "" {
		options.Title = reportData.Title
	}
	if options.Title == "" {
		options.Title = "SARIF HTML Report"
	}

	view := viewModel{
		Report:      reportData,
		Title:       options.Title,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Severities:  orderedCounts(reportData.Summary.BySeverity, []string{"error", "warning", "note", "none"}),
		Sources:     orderedCounts(countBySource(reportData.Findings), nil),
		ToolGroups:  groupedFindingsByTool(reportData.Findings),
		Rules:       orderedCounts(reportData.Summary.ByRule, nil),
		Files:       orderedCounts(reportData.Summary.ByFile, nil),
	}

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"severityClass": severityClass,
		"line":          line,
		"sourceLink": func(finding report.Finding) string {
			return sourceLink(finding, options)
		},
		"hasDetails": hasDetails,
	}).Parse(pageTemplate)
	if err != nil {
		return nil, fmt.Errorf("parse HTML template: %w", err)
	}

	var output bytes.Buffer
	if err := tmpl.Execute(&output, view); err != nil {
		return nil, fmt.Errorf("render HTML report: %w", err)
	}
	return output.Bytes(), nil
}

type viewModel struct {
	Report      report.Report
	Title       string
	GeneratedAt string
	Severities  []count
	Sources     []count
	ToolGroups  []toolGroup
	Rules       []count
	Files       []count
}

type count struct {
	Name  string
	Count int
}

type toolGroup struct {
	Name     string
	Count    int
	Findings []report.Finding
}

func orderedCounts(values map[string]int, preferred []string) []count {
	result := make([]count, 0, len(values))
	seen := map[string]bool{}
	for _, name := range preferred {
		if value, ok := values[name]; ok {
			result = append(result, count{Name: name, Count: value})
			seen[name] = true
		}
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		if !seen[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, count{Name: key, Count: values[key]})
	}
	return result
}

func countBySource(findings []report.Finding) map[string]int {
	result := map[string]int{}
	for _, finding := range findings {
		source := strings.TrimSpace(finding.Source)
		if source == "" {
			source = "unknown-source"
		}
		result[source]++
	}
	return result
}

func groupedFindingsByTool(findings []report.Finding) []toolGroup {
	byTool := map[string][]report.Finding{}
	for _, finding := range findings {
		tool := strings.TrimSpace(finding.Tool)
		if tool == "" {
			tool = "unknown-tool"
		}
		byTool[tool] = append(byTool[tool], finding)
	}

	groups := make([]toolGroup, 0, len(byTool))
	for tool, toolFindings := range byTool {
		groups = append(groups, toolGroup{Name: tool, Count: len(toolFindings), Findings: toolFindings})
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Count != groups[j].Count {
			return groups[i].Count > groups[j].Count
		}
		return groups[i].Name < groups[j].Name
	})
	return groups
}

func severityClass(level string) string {
	switch strings.ToLower(level) {
	case "error":
		return "sev-error"
	case "warning":
		return "sev-warning"
	case "note":
		return "sev-note"
	case "none":
		return "sev-none"
	default:
		return "sev-unknown"
	}
}

func line(value int) string {
	if value <= 0 {
		return "-"
	}
	return fmt.Sprint(value)
}

func hasDetails(finding report.Finding) bool {
	return finding.RuleDescription != "" || finding.RuleHelpURI != "" || finding.Snippet != "" || len(finding.RelatedLocations) > 0 || len(finding.CodeFlows) > 0
}

func sourceLink(finding report.Finding, options Options) string {
	if strings.HasPrefix(finding.URI, "http://") || strings.HasPrefix(finding.URI, "https://") {
		return finding.URI
	}
	if finding.Path == "" {
		return ""
	}

	if options.SourceURLTemplate != "" {
		return sourceLinkFromTemplate(options.SourceURLTemplate, finding, options)
	}

	if options.RepoURL == "" || options.Revision == "" {
		return ""
	}

	base := strings.TrimSuffix(options.RepoURL, "/")
	base = strings.TrimSuffix(base, ".git")
	escapedPath := escapePath(finding.Path)

	if strings.Contains(base, "gitlab.") || strings.Contains(base, "gitlab.com") {
		return fmt.Sprintf("%s/-/blob/%s/%s%s", base, url.PathEscape(options.Revision), escapedPath, lineFragment(finding))
	}
	return fmt.Sprintf("%s/blob/%s/%s%s", base, url.PathEscape(options.Revision), escapedPath, lineFragment(finding))
}

func sourceLinkFromTemplate(template string, finding report.Finding, options Options) string {
	replacements := map[string]string{
		"{path}":         escapePath(finding.Path),
		"{pathRaw}":      strings.TrimLeft(finding.Path, "/"),
		"{line}":         positiveInt(finding.StartLine),
		"{startLine}":    positiveInt(finding.StartLine),
		"{endLine}":      positiveInt(finding.EndLine),
		"{column}":       positiveInt(finding.StartColumn),
		"{startColumn}":  positiveInt(finding.StartColumn),
		"{endColumn}":    positiveInt(finding.EndColumn),
		"{revision}":     url.PathEscape(options.Revision),
		"{revisionRaw}":  options.Revision,
		"{lineFragment}": lineFragment(finding),
	}

	result := template
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

func escapePath(rawPath string) string {
	escapedPath := strings.TrimLeft(rawPath, "/")
	parts := strings.Split(escapedPath, "/")
	for index, part := range parts {
		parts[index] = url.PathEscape(part)
	}
	return strings.Join(parts, "/")
}

func lineFragment(finding report.Finding) string {
	if finding.StartLine <= 0 {
		return ""
	}
	if finding.EndLine > finding.StartLine {
		return fmt.Sprintf("#L%d-L%d", finding.StartLine, finding.EndLine)
	}
	return fmt.Sprintf("#L%d", finding.StartLine)
}

func positiveInt(value int) string {
	if value <= 0 {
		return ""
	}
	return fmt.Sprint(value)
}
