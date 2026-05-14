# Contributing to sarif-html

Thanks for taking the time to improve `sarif-html`.

This project is still young, so the most useful contributions are small, focused changes that improve correctness, SARIF compatibility, report usability, or documentation clarity.

## Development Setup

Requirements:

- Go 1.26 or newer.

Run the test suite:

```sh
go test ./...
```

If the default Go build cache is not writable in your environment:

```sh
GOCACHE=/private/tmp/sarif-html-go-cache go test ./...
```

Build the CLI:

```sh
go build ./cmd/sarif-html
```

Before opening a pull request, run the local equivalents of the strict CI checks:

```sh
gofmt -w $(git ls-files '*.go')
go mod tidy
go test ./...
go vet ./...
go test -race ./...
```

GitHub Actions also runs `golangci-lint`, `govulncheck`, and a GoReleaser
snapshot check when release-related files change.

Generate the example report:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif --out report.html
```

## Contribution Guidelines

- Keep changes focused and easy to review.
- Add or update tests for parser, normalizer, renderer, or CLI behavior changes.
- Prefer real SARIF fixtures when adding support for tool-specific behavior.
- Do not add dependencies unless they clearly remove more complexity than they introduce.
- Keep the generated report UI practical and information-dense.
- Keep docs readable and project-facing; they do not need to look like the report UI.

## Good First Areas

- Add fixtures from real SARIF-producing tools.
- Improve URI normalization and `uriBaseId` support.
- Add baseline comparison support.
- Improve HTML report filtering and grouping.
- Add release automation once the module path and license are final.

## Before Opening a Pull Request

Run:

```sh
go test ./...
```

If your change affects generated HTML, also generate `report.html` from the example fixture and inspect it locally.
