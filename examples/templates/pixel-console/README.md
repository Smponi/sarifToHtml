# Pixel Console Template

This directory contains a complete custom `sarif-html` template for the
`sarif-html.template.v1` data contract.

It deliberately does not look like the built-in report. The style is a
pixelated arcade debug console with embedded CSS, JavaScript filters, a ticker,
scanline animation, and severity cards.

Run it locally:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif \
  --title "Pixel Console SARIF" \
  --template examples/templates/pixel-console/report.tmpl \
  --out /private/tmp/pixel-console-report.html
```

The template is self-contained: no external fonts, images, scripts, or network
assets are required.
