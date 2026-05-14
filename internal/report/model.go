package report

// Report is the normalized, renderer-facing representation of one or more
// SARIF logs.
type Report struct {
	Title string
	// Sources lists the SARIF input files that contributed findings.
	Sources []string
	// Findings is sorted into review order and uses report-local IDs.
	Findings []Finding
	Summary  Summary
}

// Summary stores aggregate counts used by the report UI and CI decisions.
type Summary struct {
	Total int
	// BySeverity, ByTool, ByRule, and ByFile back the report facets.
	BySeverity map[string]int
	ByTool     map[string]int
	ByRule     map[string]int
	ByFile     map[string]int
}

// Finding is one normalized scanner result with source, rule, location, and
// evidence metadata.
type Finding struct {
	ID          string
	Source      string
	Tool        string
	ToolVersion string

	RuleID          string
	RuleName        string
	RuleDescription string
	RuleHelpURI     string

	Level   string
	Message string

	// Path is the normalized repository-relative path used for grouping and
	// source links. URI and URIBaseID preserve the original SARIF location data.
	Path      string
	URI       string
	URIBaseID string

	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int

	Snippet       string
	Fingerprint   string
	BaselineState string

	RelatedLocations []RelatedLocation
	CodeFlows        []CodeFlow
}

// RelatedLocation is secondary evidence associated with a finding.
type RelatedLocation struct {
	Message     string
	Path        string
	URI         string
	StartLine   int
	StartColumn int
}

// CodeFlow is a normalized execution path associated with a finding.
type CodeFlow struct {
	Locations []RelatedLocation
}
