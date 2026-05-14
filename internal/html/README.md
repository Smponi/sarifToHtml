# HTML Package

`internal/html` renders normalized report data into a single static HTML file.

The output should remain self-contained, portable, and easy to inspect as a CI
artifact. This package may format, group, and link findings, but it should not
parse SARIF or make normalization decisions.

When changing the template or styles, update renderer tests and inspect a
generated report from `examples/detekt-like.sarif`.
