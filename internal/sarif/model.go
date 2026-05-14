package sarif

// Log is the top-level SARIF 2.1.0 document.
type Log struct {
	Schema  string `json:"$schema,omitempty"`
	Version string `json:"version"`
	Runs    []Run  `json:"runs"`
}

// Run describes one scanner execution inside a SARIF log.
type Run struct {
	Tool               Tool                        `json:"tool"`
	OriginalURIBaseIDs map[string]ArtifactLocation `json:"originalUriBaseIds,omitempty"`
	Results            []Result                    `json:"results,omitempty"`
}

// Tool contains scanner metadata.
type Tool struct {
	Driver Driver `json:"driver"`
}

// Driver describes the scanner driver that produced the results.
type Driver struct {
	Name            string                `json:"name"`
	Version         string                `json:"version,omitempty"`
	SemanticVersion string                `json:"semanticVersion,omitempty"`
	InformationURI  string                `json:"informationUri,omitempty"`
	Rules           []ReportingDescriptor `json:"rules,omitempty"`
}

// ReportingDescriptor describes a rule emitted by the scanner.
type ReportingDescriptor struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name,omitempty"`
	ShortDescription Message                `json:"shortDescription,omitempty"`
	FullDescription  Message                `json:"fullDescription,omitempty"`
	Help             Message                `json:"help,omitempty"`
	HelpURI          string                 `json:"helpUri,omitempty"`
	Properties       map[string]interface{} `json:"properties,omitempty"`
}

// Result is a single SARIF finding.
type Result struct {
	RuleID              string                 `json:"ruleId,omitempty"`
	RuleIndex           *int                   `json:"ruleIndex,omitempty"`
	Level               string                 `json:"level,omitempty"`
	Message             Message                `json:"message"`
	Locations           []Location             `json:"locations,omitempty"`
	RelatedLocations    []Location             `json:"relatedLocations,omitempty"`
	CodeFlows           []CodeFlow             `json:"codeFlows,omitempty"`
	Fingerprints        map[string]string      `json:"fingerprints,omitempty"`
	PartialFingerprints map[string]string      `json:"partialFingerprints,omitempty"`
	BaselineState       string                 `json:"baselineState,omitempty"`
	Properties          map[string]interface{} `json:"properties,omitempty"`
}

// Message contains plain text or Markdown content.
type Message struct {
	Text     string `json:"text,omitempty"`
	Markdown string `json:"markdown,omitempty"`
}

// Location points to code or another relevant artifact location.
type Location struct {
	PhysicalLocation PhysicalLocation `json:"physicalLocation,omitempty"`
	Message          Message          `json:"message,omitempty"`
}

// PhysicalLocation describes the concrete artifact and region for a location.
type PhysicalLocation struct {
	ArtifactLocation ArtifactLocation `json:"artifactLocation,omitempty"`
	Region           Region           `json:"region,omitempty"`
	ContextRegion    Region           `json:"contextRegion,omitempty"`
}

// ArtifactLocation identifies a file or URI.
type ArtifactLocation struct {
	URI       string `json:"uri,omitempty"`
	URIBaseID string `json:"uriBaseId,omitempty"`
}

// Region describes a line and column range inside an artifact.
type Region struct {
	StartLine   int     `json:"startLine,omitempty"`
	StartColumn int     `json:"startColumn,omitempty"`
	EndLine     int     `json:"endLine,omitempty"`
	EndColumn   int     `json:"endColumn,omitempty"`
	Snippet     Snippet `json:"snippet,omitempty"`
}

// Snippet contains source text attached to a SARIF region.
type Snippet struct {
	Text string `json:"text,omitempty"`
}

// CodeFlow captures one or more thread flows for a result.
type CodeFlow struct {
	ThreadFlows []ThreadFlow `json:"threadFlows,omitempty"`
}

// ThreadFlow contains ordered locations that describe one execution path.
type ThreadFlow struct {
	Locations []ThreadFlowLocation `json:"locations,omitempty"`
}

// ThreadFlowLocation wraps a location inside a thread flow.
type ThreadFlowLocation struct {
	Location Location `json:"location,omitempty"`
}
