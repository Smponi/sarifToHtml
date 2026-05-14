# sarif-html Command

`cmd/sarif-html` is the CLI entrypoint for generating self-contained HTML
reports from one or more SARIF 2.1.0 files.

Responsibilities:

- Parse command-line flags and positional SARIF inputs.
- Infer source link templates from explicit flags or common CI environments.
- Load SARIF files and pass them into the report normalization pipeline.
- Load and write versioned baseline JSON through `internal/report`.
- Render HTML output to a file or standard output.
- Pass an optional custom template path into the HTML renderer.
- Validate templates with `--dry-run` and write versioned template data JSON
  with `--template-data-out`.
- Return CI-friendly exit code `2` when `--fail-on` is matched by a
  non-baseline finding.

Keep business logic out of this package whenever possible. Parser behavior
belongs in `internal/sarif`, normalized finding behavior belongs in
`internal/report`, and presentation behavior belongs in `internal/html`.
