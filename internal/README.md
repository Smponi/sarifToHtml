# Internal Packages

This directory contains implementation packages used by the CLI.

The packages are intentionally internal so the project can evolve its data model,
HTML output, and SARIF compatibility without committing to a public Go API before
the first stable release.

Package responsibilities:

- `sarif`: Minimal SARIF 2.1.0 input model and parser validation.
- `report`: Normalized finding model, summaries, snippets, and SARIF-to-report
  mapping.
- `html`: Static, self-contained HTML report rendering.
