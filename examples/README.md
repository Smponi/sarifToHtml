# Examples

This directory contains small SARIF files that are safe to use in local
development, tests, documentation, and screenshots. It also contains optional
custom report templates under `templates/` and checked-in rendered examples
under `outputs/`.

Examples should be compact, deterministic, and easy to understand. Larger
scanner-generated reports belong in `reports/`, while intentionally vulnerable
source projects belong in `fixtures/`.

Render the pixel console template:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif \
  --title "Pixel Console SARIF" \
  --template examples/templates/pixel-console/report.tmpl \
  --out /private/tmp/pixel-console-report.html
```

Render the minimal template:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif \
  --template examples/templates/minimal/report.tmpl \
  --out /private/tmp/minimal-report.html
```
