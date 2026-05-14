# HTML Package

`internal/html` renders normalized report data into a single static HTML file.

The output should remain self-contained, portable, and easy to inspect as a CI
artifact. This package may format, group, and link findings, but it should not
parse SARIF or make normalization decisions.

Custom templates use Go's `html/template` package and receive the exported
`TemplateData` contract. Keep `TemplateDataVersion` stable for compatible
changes and increment it only when template authors need to update their files.
The internal report model may evolve, but any field documented as part of
`TemplateData` should be treated as user-facing template API.

The default embedded template is the compatibility baseline. When adding helper
functions, keep them deterministic and side-effect free; do not expose file
system, environment, process, network, or "safe HTML" helpers to custom
templates.

When changing the template or styles, update renderer tests and inspect a
generated report from `examples/detekt-like.sarif`.
