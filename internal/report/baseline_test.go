package report

import (
	"strings"
	"testing"
	"time"
)

func TestBaselineRoundTripMarksSameFindingsUnchanged(t *testing.T) {
	reportData := Report{
		Findings: []Finding{
			{
				ID:          "F0001",
				Source:      "detekt.sarif",
				Tool:        "detekt",
				RuleID:      "LongMethod",
				Level:       "error",
				Message:     "Too long",
				Path:        "src/App.kt",
				StartLine:   12,
				Fingerprint: "abc123",
			},
		},
	}

	baseline := NewBaselineAt(reportData, time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC))
	raw, err := MarshalBaseline(baseline)
	if err != nil {
		t.Fatalf("MarshalBaseline returned error: %v", err)
	}
	parsed, err := ParseBaseline(strings.NewReader(string(raw)))
	if err != nil {
		t.Fatalf("ParseBaseline returned error: %v", err)
	}

	current := reportData
	current.Findings = append([]Finding(nil), reportData.Findings...)
	ApplyBaseline(&current, parsed)

	if current.Findings[0].BaselineState != BaselineStateUnchanged {
		t.Fatalf("expected finding to be unchanged, got %q", current.Findings[0].BaselineState)
	}
	if current.Summary.ByBaselineState[BaselineStateUnchanged] != 1 {
		t.Fatalf("expected unchanged summary count, got %#v", current.Summary.ByBaselineState)
	}
}

func TestApplyBaselineMarksMissingFindingsNew(t *testing.T) {
	baseline := NewBaselineAt(Report{
		Findings: []Finding{
			{Tool: "detekt", RuleID: "LongMethod", Path: "src/App.kt", Fingerprint: "abc123"},
		},
	}, time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC))
	current := Report{
		Findings: []Finding{
			{Tool: "detekt", RuleID: "LongMethod", Path: "src/App.kt", Fingerprint: "abc123", Level: "error"},
			{Tool: "detekt", RuleID: "ComplexMethod", Path: "src/App.kt", Fingerprint: "def456", Level: "error"},
		},
	}

	ApplyBaseline(&current, baseline)

	if current.Findings[0].BaselineState != BaselineStateUnchanged {
		t.Fatalf("expected first finding unchanged, got %q", current.Findings[0].BaselineState)
	}
	if current.Findings[1].BaselineState != BaselineStateNew {
		t.Fatalf("expected second finding new, got %q", current.Findings[1].BaselineState)
	}
}

func TestApplyBaselineRespectsDuplicateCounts(t *testing.T) {
	baseline := NewBaselineAt(Report{
		Findings: []Finding{
			{Tool: "semgrep", RuleID: "Rule", Path: "src/app.py", Fingerprint: "same"},
		},
	}, time.Date(2026, 5, 14, 10, 0, 0, 0, time.UTC))
	current := Report{
		Findings: []Finding{
			{Tool: "semgrep", RuleID: "Rule", Path: "src/app.py", Fingerprint: "same"},
			{Tool: "semgrep", RuleID: "Rule", Path: "src/app.py", Fingerprint: "same"},
		},
	}

	ApplyBaseline(&current, baseline)

	if current.Findings[0].BaselineState != BaselineStateUnchanged {
		t.Fatalf("expected first duplicate unchanged, got %q", current.Findings[0].BaselineState)
	}
	if current.Findings[1].BaselineState != BaselineStateNew {
		t.Fatalf("expected second duplicate new, got %q", current.Findings[1].BaselineState)
	}
}

func TestParseBaselineRejectsUnknownVersion(t *testing.T) {
	_, err := ParseBaseline(strings.NewReader(`{"schemaVersion":"other","findings":[]}`))
	if err == nil {
		t.Fatal("expected unsupported version error")
	}
}
