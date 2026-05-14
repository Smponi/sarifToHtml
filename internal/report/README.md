# Report Package

`internal/report` converts parsed SARIF logs into the stable report model used by
the HTML renderer.

Responsibilities:

- Resolve rules, levels, paths, fingerprints, related locations, and code flows.
- Merge several SARIF inputs into a single sorted report.
- Build summary counts for severities, tools, rules, and files.
- Hydrate missing source snippets from trusted source roots.

This package is the main compatibility layer between diverse scanner output and
the user-facing report. Keep behavior deterministic and well covered by tests.
