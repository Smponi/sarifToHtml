package report

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// BaselineDataVersion identifies the JSON contract used to store accepted
// findings between runs.
const BaselineDataVersion = "sarif-html.baseline.v1"

// Baseline is the stable JSON document used by --baseline and --baseline-out.
type Baseline struct {
	SchemaVersion string            `json:"schemaVersion"`
	GeneratedAt   string            `json:"generatedAt"`
	Findings      []BaselineFinding `json:"findings"`
}

// BaselineFinding stores the identity and review context for one accepted
// finding. Key is the matching identity; the remaining fields make the baseline
// readable and allow a missing key to be reconstructed.
type BaselineFinding struct {
	Key           string `json:"key"`
	Source        string `json:"source,omitempty"`
	Tool          string `json:"tool"`
	RuleID        string `json:"ruleID"`
	Path          string `json:"path"`
	StartLine     int    `json:"startLine,omitempty"`
	Fingerprint   string `json:"fingerprint"`
	Level         string `json:"level,omitempty"`
	Message       string `json:"message,omitempty"`
	BaselineState string `json:"baselineState"`
}

// NewBaseline creates a baseline from the current report using the current UTC
// time as metadata.
func NewBaseline(reportData Report) Baseline {
	return NewBaselineAt(reportData, time.Now().UTC())
}

// NewBaselineAt creates a deterministic baseline when tests need to pin the
// generation timestamp.
func NewBaselineAt(reportData Report, generatedAt time.Time) Baseline {
	if generatedAt.IsZero() {
		generatedAt = time.Now().UTC()
	}

	baseline := Baseline{
		SchemaVersion: BaselineDataVersion,
		GeneratedAt:   generatedAt.UTC().Format(time.RFC3339),
		Findings:      make([]BaselineFinding, 0, len(reportData.Findings)),
	}
	for _, finding := range reportData.Findings {
		entry := BaselineFinding{
			Key:           FindingBaselineKey(finding),
			Source:        finding.Source,
			Tool:          finding.Tool,
			RuleID:        finding.RuleID,
			Path:          finding.Path,
			StartLine:     finding.StartLine,
			Fingerprint:   finding.Fingerprint,
			Level:         finding.Level,
			Message:       finding.Message,
			BaselineState: BaselineStateUnchanged,
		}
		baseline.Findings = append(baseline.Findings, entry)
	}
	return baseline
}

// ParseBaseline decodes and validates a baseline JSON document.
func ParseBaseline(r io.Reader) (Baseline, error) {
	var baseline Baseline
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&baseline); err != nil {
		return Baseline{}, fmt.Errorf("parse baseline JSON: %w", err)
	}
	if baseline.SchemaVersion == "" {
		return Baseline{}, fmt.Errorf("baseline schemaVersion is required")
	}
	if baseline.SchemaVersion != BaselineDataVersion {
		return Baseline{}, fmt.Errorf("unsupported baseline schemaVersion %q: only %s is supported", baseline.SchemaVersion, BaselineDataVersion)
	}
	for index := range baseline.Findings {
		baseline.Findings[index].BaselineState = normalizeBaselineState(baseline.Findings[index].BaselineState)
		if baseline.Findings[index].BaselineState == "" {
			baseline.Findings[index].BaselineState = BaselineStateUnchanged
		}
		if baseline.Findings[index].Key == "" {
			baseline.Findings[index].Key = baselineFindingKey(baseline.Findings[index])
		}
	}
	return baseline, nil
}

// MarshalBaseline returns a stable, human-readable JSON baseline.
func MarshalBaseline(baseline Baseline) ([]byte, error) {
	output, err := json.MarshalIndent(baseline, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal baseline: %w", err)
	}
	return append(output, '\n'), nil
}

// ApplyBaseline marks findings that are already present in baseline as
// unchanged and all other current findings as new. Duplicate keys are matched by
// count, so an additional occurrence of the same finding identity is still new.
func ApplyBaseline(reportData *Report, baseline Baseline) {
	statesByKey := map[string][]string{}
	for _, entry := range baseline.Findings {
		key := entry.Key
		if key == "" {
			key = baselineFindingKey(entry)
		}
		state := normalizeBaselineState(entry.BaselineState)
		if state == "" {
			state = BaselineStateUnchanged
		}
		statesByKey[key] = append(statesByKey[key], state)
	}

	for index := range reportData.Findings {
		key := FindingBaselineKey(reportData.Findings[index])
		states := statesByKey[key]
		if len(states) == 0 {
			reportData.Findings[index].BaselineState = BaselineStateNew
			continue
		}
		reportData.Findings[index].BaselineState = states[0]
		statesByKey[key] = states[1:]
	}
	RebuildSummary(reportData)
}

// FindingBaselineKey returns the stable identity used to match current findings
// against a persisted baseline.
func FindingBaselineKey(finding Finding) string {
	return stableFingerprint(finding.Tool, finding.RuleID, finding.Path, finding.Fingerprint)
}

func baselineFindingKey(finding BaselineFinding) string {
	return stableFingerprint(finding.Tool, finding.RuleID, finding.Path, finding.Fingerprint)
}
