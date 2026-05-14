# Minimal Template

This template is the smallest complete custom report example. It is useful for
testing the template data contract or as a starting point for custom reports.

Render it with:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif \
  --template examples/templates/minimal/report.tmpl \
  --out /private/tmp/minimal-report.html
```

