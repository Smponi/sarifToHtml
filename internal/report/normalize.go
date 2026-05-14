package report

import (
	"fmt"
	"net/url"
	"path"
	"sort"
	"strings"

	"sarif-html/internal/sarif"
)

const (
	levelError   = "error"
	levelWarning = "warning"
	levelNote    = "note"
	levelNone    = "none"
)

// FromSARIF converts a parsed SARIF log into a deterministic report.
func FromSARIF(log *sarif.Log, sourceName string) Report {
	result := Report{
		Sources: []string{sourceName},
		Summary: newSummary(),
	}

	for _, run := range log.Runs {
		rulesByID, rulesByIndex := indexRules(run.Tool.Driver.Rules)
		uriBaseIDs := run.OriginalURIBaseIDs
		toolName := strings.TrimSpace(run.Tool.Driver.Name)
		if toolName == "" {
			toolName = "unknown-tool"
		}

		for _, sarifResult := range run.Results {
			rule := resolveRule(sarifResult, rulesByID, rulesByIndex)
			location := firstLocation(sarifResult.Locations)
			artifact := location.PhysicalLocation.ArtifactLocation
			region := location.PhysicalLocation.Region
			context := location.PhysicalLocation.ContextRegion
			ruleID := strings.TrimSpace(sarifResult.RuleID)
			if ruleID == "" && rule.ID != "" {
				ruleID = rule.ID
			}
			if ruleID == "" {
				ruleID = "unknown-rule"
			}

			level := normalizeLevel(sarifResult.Level)
			message := messageText(sarifResult.Message)
			filePath := resolveArtifactPath(artifact, uriBaseIDs)
			snippet := context.Snippet.Text
			if snippet == "" {
				snippet = region.Snippet.Text
			}

			finding := Finding{
				ID:              fmt.Sprintf("F%04d", len(result.Findings)+1),
				Source:          sourceName,
				Tool:            toolName,
				ToolVersion:     firstNonEmpty(run.Tool.Driver.SemanticVersion, run.Tool.Driver.Version),
				RuleID:          ruleID,
				RuleName:        firstNonEmpty(rule.Name, ruleID),
				RuleDescription: firstNonEmpty(messageText(rule.ShortDescription), messageText(rule.FullDescription)),
				RuleHelpURI:     rule.HelpURI,
				Level:           level,
				Message:         message,
				Path:            filePath,
				URI:             artifact.URI,
				URIBaseID:       artifact.URIBaseID,
				StartLine:       region.StartLine,
				StartColumn:     region.StartColumn,
				EndLine:         region.EndLine,
				EndColumn:       region.EndColumn,
				Snippet:         snippet,
				Fingerprint: firstNonEmpty(
					firstFingerprint(sarifResult.PartialFingerprints, "primaryLocationLineHash", "primaryLocationStartColumnFingerprint"),
					firstFingerprint(sarifResult.Fingerprints, "stable", "guid"),
				),
				BaselineState: sarifResult.BaselineState,
			}
			if finding.Message == "" {
				finding.Message = finding.RuleDescription
			}
			if finding.Fingerprint == "" {
				finding.Fingerprint = stableFingerprint(finding.Tool, finding.RuleID, finding.Path, fmt.Sprint(finding.StartLine), finding.Message)
			}
			finding.RelatedLocations = normalizeRelated(sarifResult.RelatedLocations, uriBaseIDs)
			finding.CodeFlows = normalizeCodeFlows(sarifResult.CodeFlows, uriBaseIDs)

			result.Findings = append(result.Findings, finding)
			addSummary(&result.Summary, finding)
		}
	}

	sortFindings(result.Findings)
	reassignIDs(result.Findings)
	return result
}

// Merge combines several reports into one sorted report and rebuilds summaries.
func Merge(title string, reports ...Report) Report {
	merged := Report{
		Title:   title,
		Summary: newSummary(),
	}
	seenSources := map[string]bool{}
	for _, report := range reports {
		for _, source := range report.Sources {
			if source != "" && !seenSources[source] {
				merged.Sources = append(merged.Sources, source)
				seenSources[source] = true
			}
		}
		merged.Findings = append(merged.Findings, report.Findings...)
	}
	sortFindings(merged.Findings)
	reassignIDs(merged.Findings)
	for _, finding := range merged.Findings {
		addSummary(&merged.Summary, finding)
	}
	return merged
}

func newSummary() Summary {
	return Summary{
		BySeverity: map[string]int{},
		ByTool:     map[string]int{},
		ByRule:     map[string]int{},
		ByFile:     map[string]int{},
	}
}

func indexRules(rules []sarif.ReportingDescriptor) (map[string]sarif.ReportingDescriptor, map[int]sarif.ReportingDescriptor) {
	byID := map[string]sarif.ReportingDescriptor{}
	byIndex := map[int]sarif.ReportingDescriptor{}
	for index, rule := range rules {
		if rule.ID != "" {
			byID[rule.ID] = rule
		}
		byIndex[index] = rule
	}
	return byID, byIndex
}

func resolveRule(result sarif.Result, byID map[string]sarif.ReportingDescriptor, byIndex map[int]sarif.ReportingDescriptor) sarif.ReportingDescriptor {
	if result.RuleID != "" {
		if rule, ok := byID[result.RuleID]; ok {
			return rule
		}
	}
	if result.RuleIndex != nil {
		if rule, ok := byIndex[*result.RuleIndex]; ok {
			return rule
		}
	}
	return sarif.ReportingDescriptor{}
}

func firstLocation(locations []sarif.Location) sarif.Location {
	if len(locations) == 0 {
		return sarif.Location{}
	}
	return locations[0]
}

func normalizeRelated(locations []sarif.Location, uriBaseIDs map[string]sarif.ArtifactLocation) []RelatedLocation {
	related := make([]RelatedLocation, 0, len(locations))
	for _, location := range locations {
		artifact := location.PhysicalLocation.ArtifactLocation
		region := location.PhysicalLocation.Region
		related = append(related, RelatedLocation{
			Message:     messageText(location.Message),
			Path:        resolveArtifactPath(artifact, uriBaseIDs),
			URI:         artifact.URI,
			StartLine:   region.StartLine,
			StartColumn: region.StartColumn,
		})
	}
	return related
}

func normalizeCodeFlows(codeFlows []sarif.CodeFlow, uriBaseIDs map[string]sarif.ArtifactLocation) []CodeFlow {
	flows := make([]CodeFlow, 0, len(codeFlows))
	for _, codeFlow := range codeFlows {
		for _, thread := range codeFlow.ThreadFlows {
			flow := CodeFlow{Locations: make([]RelatedLocation, 0, len(thread.Locations))}
			for _, threadLocation := range thread.Locations {
				location := threadLocation.Location
				artifact := location.PhysicalLocation.ArtifactLocation
				region := location.PhysicalLocation.Region
				flow.Locations = append(flow.Locations, RelatedLocation{
					Message:     messageText(location.Message),
					Path:        resolveArtifactPath(artifact, uriBaseIDs),
					URI:         artifact.URI,
					StartLine:   region.StartLine,
					StartColumn: region.StartColumn,
				})
			}
			if len(flow.Locations) > 0 {
				flows = append(flows, flow)
			}
		}
	}
	return flows
}

func addSummary(summary *Summary, finding Finding) {
	summary.Total++
	summary.BySeverity[finding.Level]++
	summary.ByTool[finding.Tool]++
	summary.ByRule[finding.RuleID]++
	if finding.Path != "" {
		summary.ByFile[finding.Path]++
	}
}

func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		left, right := findings[i], findings[j]
		if severityRank(left.Level) != severityRank(right.Level) {
			return severityRank(left.Level) > severityRank(right.Level)
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.StartLine != right.StartLine {
			return left.StartLine < right.StartLine
		}
		return left.RuleID < right.RuleID
	})
}

func reassignIDs(findings []Finding) {
	for index := range findings {
		findings[index].ID = fmt.Sprintf("F%04d", index+1)
	}
}

func normalizeLevel(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case levelError:
		return levelError
	case levelWarning:
		return levelWarning
	case levelNote:
		return levelNote
	case levelNone:
		return levelNone
	default:
		return levelWarning
	}
}

func severityRank(level string) int {
	switch level {
	case levelError:
		return 4
	case levelWarning:
		return 3
	case levelNote:
		return 2
	case levelNone:
		return 1
	default:
		return 0
	}
}

func messageText(message sarif.Message) string {
	if strings.TrimSpace(message.Text) != "" {
		return strings.TrimSpace(message.Text)
	}
	return strings.TrimSpace(message.Markdown)
}

func normalizePath(rawURI string) string {
	if rawURI == "" {
		return ""
	}
	if strings.HasPrefix(rawURI, "file://") {
		parsed, err := url.Parse(rawURI)
		if err == nil {
			rawURI = parsed.Path
		}
	}
	if unescaped, err := url.PathUnescape(rawURI); err == nil {
		rawURI = unescaped
	}
	rawURI = strings.TrimPrefix(rawURI, "./")
	rawURI = path.Clean(strings.ReplaceAll(rawURI, "\\", "/"))
	if rawURI == "." {
		return ""
	}
	return rawURI
}

func resolveArtifactPath(artifact sarif.ArtifactLocation, uriBaseIDs map[string]sarif.ArtifactLocation) string {
	artifactPath := normalizePath(artifact.URI)
	if artifact.URIBaseID == "" || len(uriBaseIDs) == 0 {
		return artifactPath
	}
	if isAbsolutePath(artifactPath) {
		return artifactPath
	}

	basePath, ok := resolveURIBasePath(artifact.URIBaseID, uriBaseIDs, map[string]bool{})
	if !ok || basePath == "" || isAbsolutePath(basePath) {
		return artifactPath
	}
	if artifactPath == "" {
		return basePath
	}
	return normalizePath(path.Join(basePath, artifactPath))
}

func resolveURIBasePath(id string, uriBaseIDs map[string]sarif.ArtifactLocation, seen map[string]bool) (string, bool) {
	if seen[id] {
		return "", false
	}
	base, ok := uriBaseIDs[id]
	if !ok {
		return "", false
	}
	seen[id] = true

	basePath := normalizePath(base.URI)
	if base.URIBaseID == "" {
		return basePath, true
	}

	parentPath, ok := resolveURIBasePath(base.URIBaseID, uriBaseIDs, seen)
	if !ok || parentPath == "" || isAbsolutePath(basePath) {
		return basePath, ok
	}
	if isAbsolutePath(parentPath) {
		return basePath, true
	}
	if basePath == "" {
		return parentPath, true
	}
	return normalizePath(path.Join(parentPath, basePath)), true
}

func isAbsolutePath(value string) bool {
	if strings.HasPrefix(value, "/") {
		return true
	}
	if len(value) >= 3 && ((value[0] >= 'A' && value[0] <= 'Z') || (value[0] >= 'a' && value[0] <= 'z')) && value[1] == ':' && value[2] == '/' {
		return true
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// MeetsThreshold reports whether level is at least as severe as threshold.
func MeetsThreshold(level, threshold string) bool {
	threshold = normalizeLevel(threshold)
	return severityRank(normalizeLevel(level)) >= severityRank(threshold)
}
