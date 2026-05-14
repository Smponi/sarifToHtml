package report

import (
	"strings"
	"testing"

	"sarif-html/internal/sarif"
)

func TestFromSARIFNormalizesFindings(t *testing.T) {
	log, err := sarif.Parse(strings.NewReader(sampleSARIF))
	if err != nil {
		t.Fatalf("parse sample SARIF: %v", err)
	}

	reportData := FromSARIF(log, "sample.sarif")
	if reportData.Summary.Total != 2 {
		t.Fatalf("expected 2 findings, got %d", reportData.Summary.Total)
	}
	if reportData.Summary.BySeverity["error"] != 1 {
		t.Fatalf("expected one error, got %d", reportData.Summary.BySeverity["error"])
	}

	first := reportData.Findings[0]
	if first.ID != "F0001" {
		t.Fatalf("expected stable ID, got %s", first.ID)
	}
	if first.RuleID != "ComplexMethod" {
		t.Fatalf("expected ComplexMethod first due severity sort, got %s", first.RuleID)
	}
	if first.Path != "src/main/kotlin/App.kt" {
		t.Fatalf("unexpected path: %s", first.Path)
	}
	if first.Fingerprint != "abc123" {
		t.Fatalf("expected SARIF fingerprint, got %s", first.Fingerprint)
	}
	if len(first.RelatedLocations) != 1 {
		t.Fatalf("expected related location")
	}
	if len(first.CodeFlows) != 1 || len(first.CodeFlows[0].Locations) != 1 {
		t.Fatalf("expected one code flow location")
	}
}

func TestMergeRebuildsSummaryAndIDs(t *testing.T) {
	log, err := sarif.Parse(strings.NewReader(sampleSARIF))
	if err != nil {
		t.Fatalf("parse sample SARIF: %v", err)
	}
	left := FromSARIF(log, "left.sarif")
	right := FromSARIF(log, "right.sarif")

	merged := Merge("Merged", left, right)
	if merged.Title != "Merged" {
		t.Fatalf("unexpected title: %s", merged.Title)
	}
	if merged.Summary.Total != 4 {
		t.Fatalf("expected 4 findings, got %d", merged.Summary.Total)
	}
	if merged.Findings[3].ID != "F0004" {
		t.Fatalf("expected reassigned IDs, got %s", merged.Findings[3].ID)
	}
	if len(merged.Sources) != 2 {
		t.Fatalf("expected two sources, got %d", len(merged.Sources))
	}
}

func TestMeetsThreshold(t *testing.T) {
	if !MeetsThreshold("error", "warning") {
		t.Fatal("error should meet warning threshold")
	}
	if MeetsThreshold("note", "warning") {
		t.Fatal("note should not meet warning threshold")
	}
}

func TestFromSARIFResolvesRelativeURIBaseIDs(t *testing.T) {
	log, err := sarif.Parse(strings.NewReader(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "tool" } },
				"originalUriBaseIds": {
					"SRC": { "uri": "src/main/kotlin/" }
				},
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Message" },
						"locations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "App.kt", "uriBaseId": "SRC" },
									"region": { "startLine": 4 }
								}
							}
						],
						"relatedLocations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "Util.kt", "uriBaseId": "SRC" },
									"region": { "startLine": 8 }
								}
							}
						]
					}
				]
			}
		]
	}`))
	if err != nil {
		t.Fatalf("parse sample SARIF: %v", err)
	}

	reportData := FromSARIF(log, "base.sarif")
	finding := reportData.Findings[0]
	if finding.Path != "src/main/kotlin/App.kt" {
		t.Fatalf("unexpected path: %s", finding.Path)
	}
	if finding.RelatedLocations[0].Path != "src/main/kotlin/Util.kt" {
		t.Fatalf("unexpected related path: %s", finding.RelatedLocations[0].Path)
	}
}

func TestFromSARIFKeepsRelativePathWhenURIBaseIDIsAbsoluteFileURI(t *testing.T) {
	log, err := sarif.Parse(strings.NewReader(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "tool" } },
				"originalUriBaseIds": {
					"SRCROOT": { "uri": "file:///Users/ci/work/project/" }
				},
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Message" },
						"locations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "src/App.kt", "uriBaseId": "SRCROOT" },
									"region": { "startLine": 4 }
								}
							}
						]
					}
				]
			}
		]
	}`))
	if err != nil {
		t.Fatalf("parse sample SARIF: %v", err)
	}

	reportData := FromSARIF(log, "base.sarif")
	if reportData.Findings[0].Path != "src/App.kt" {
		t.Fatalf("unexpected path: %s", reportData.Findings[0].Path)
	}
}

func TestFromSARIFResolvesNestedRelativeURIBaseIDs(t *testing.T) {
	log, err := sarif.Parse(strings.NewReader(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "tool" } },
				"originalUriBaseIds": {
					"ROOT": { "uri": "src/" },
					"KOTLIN": { "uri": "main/kotlin/", "uriBaseId": "ROOT" }
				},
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Message" },
						"locations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "App.kt", "uriBaseId": "KOTLIN" },
									"region": { "startLine": 4 }
								}
							}
						]
					}
				]
			}
		]
	}`))
	if err != nil {
		t.Fatalf("parse sample SARIF: %v", err)
	}

	reportData := FromSARIF(log, "base.sarif")
	if reportData.Findings[0].Path != "src/main/kotlin/App.kt" {
		t.Fatalf("unexpected path: %s", reportData.Findings[0].Path)
	}
}

func TestFromSARIFKeepsRelativeURIBaseTailWhenRootIsAbsolute(t *testing.T) {
	log, err := sarif.Parse(strings.NewReader(`{
		"version": "2.1.0",
		"runs": [
			{
				"tool": { "driver": { "name": "tool" } },
				"originalUriBaseIds": {
					"SRCROOT": { "uri": "file:///Users/ci/work/project/" },
					"KOTLIN": { "uri": "src/main/kotlin/", "uriBaseId": "SRCROOT" }
				},
				"results": [
					{
						"ruleId": "Rule",
						"level": "warning",
						"message": { "text": "Message" },
						"locations": [
							{
								"physicalLocation": {
									"artifactLocation": { "uri": "App.kt", "uriBaseId": "KOTLIN" },
									"region": { "startLine": 4 }
								}
							}
						]
					}
				]
			}
		]
	}`))
	if err != nil {
		t.Fatalf("parse sample SARIF: %v", err)
	}

	reportData := FromSARIF(log, "base.sarif")
	if reportData.Findings[0].Path != "src/main/kotlin/App.kt" {
		t.Fatalf("unexpected path: %s", reportData.Findings[0].Path)
	}
}

const sampleSARIF = `{
  "version": "2.1.0",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "detekt",
          "semanticVersion": "1.23.0",
          "rules": [
            {
              "id": "LongMethod",
              "name": "Long Method",
              "shortDescription": { "text": "Method is too long." },
              "helpUri": "https://detekt.dev/docs/rules/complexity"
            },
            {
              "id": "ComplexMethod",
              "name": "Complex Method",
              "shortDescription": { "text": "Method is too complex." },
              "helpUri": "https://detekt.dev/docs/rules/complexity"
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "LongMethod",
          "ruleIndex": 0,
          "level": "warning",
          "message": { "text": "The method has 80 lines." },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "./src/main/kotlin/App.kt" },
                "region": { "startLine": 12, "startColumn": 3 }
              }
            }
          ]
        },
        {
          "ruleId": "ComplexMethod",
          "ruleIndex": 1,
          "level": "error",
          "message": { "text": "The method has high cyclomatic complexity." },
          "partialFingerprints": { "primaryLocationLineHash": "abc123" },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": { "uri": "src/main/kotlin/App.kt" },
                "region": {
                  "startLine": 40,
                  "startColumn": 5,
                  "snippet": { "text": "fun calculate() {" }
                }
              }
            }
          ],
          "relatedLocations": [
            {
              "message": { "text": "Nested branch contributes complexity." },
              "physicalLocation": {
                "artifactLocation": { "uri": "src/main/kotlin/App.kt" },
                "region": { "startLine": 45, "startColumn": 9 }
              }
            }
          ],
          "codeFlows": [
            {
              "threadFlows": [
                {
                  "locations": [
                    {
                      "location": {
                        "message": { "text": "Flow starts here." },
                        "physicalLocation": {
                          "artifactLocation": { "uri": "src/main/kotlin/App.kt" },
                          "region": { "startLine": 41 }
                        }
                      }
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  ]
}`
