# sarif-html

`sarif-html` turns SARIF 2.1.0 files into a readable, self-contained HTML report.

SARIF is a good exchange format, but it is not a great reading experience by itself. Many tools can emit SARIF, while CI platforms display it inconsistently or only in specific product tiers. `sarif-html` keeps SARIF as the input format and focuses on the thing teams often need most: a compact report that can be opened locally, uploaded as a CI artifact, or shared with reviewers.

## Current Status

This repository is an early prototype. The core flow works:

- Parse SARIF 2.1.0 input.
- Normalize results into an internal finding model.
- Merge multiple SARIF files into one report.
- Render a static HTML report with filters, summaries, snippets, related locations, code flows, and source links.
- Run CI-style gates with `--fail-on`.

The public API, output design, and module path may still change before the first tagged release.

## Quick Start

Run the example report:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif --out report.html
```

Generate source links for GitHub or GitLab:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif \
  --title "Static Analysis Report" \
  --repo-url https://github.com/acme/project \
  --revision main \
  --out report.html
```

In GitHub Actions and GitLab CI, `sarif-html` can infer source links from the standard environment variables when no explicit link flags are provided.

Use a severity threshold in CI:

```sh
go run ./cmd/sarif-html reports/detekt.sarif --out report.html --fail-on error
```

When a finding at or above the threshold exists, the report is still written and the command exits with code `2`.

## CLI

```text
sarif-html [flags] <input.sarif> [more-inputs.sarif...]
```

| Flag | Description |
| --- | --- |
| `--out` | Output HTML file. Use `-` to write HTML to stdout. Default: `report.html`. |
| `--title` | Report title. Default: `SARIF HTML Report`. |
| `--template` | Custom Go `html/template` file or directory. When omitted, the built-in default template is used. |
| `--repo-url` | Repository URL used to build clickable source links. |
| `--revision` | Branch, tag, or commit used for source links. |
| `--source-url-template` | Full source link template. Overrides `--repo-url` and `--revision` link generation. |
| `--source-root` | Local source directory used to load missing snippets from SARIF locations. May be repeated. Default: `.`. |
| `--fail-on` | Exit with code `2` when a finding at or above this level exists: `error`, `warning`, `note`, or `none`. |
| `--version` | Print the CLI version. |

Flags may appear before or after input files:

```sh
sarif-html input.sarif --out report.html
sarif-html --out report.html input.sarif
```

## Supported SARIF Data

The prototype intentionally reads a focused subset of SARIF 2.1.0:

- `run.tool.driver` metadata.
- `run.tool.driver.rules` for rule names, descriptions, and help links.
- `result.ruleId`, `result.ruleIndex`, `result.level`, and `result.message`.
- Primary physical locations with file path, line, column, and snippets.
- `relatedLocations`.
- `codeFlows.threadFlows.locations`.
- `fingerprints` and `partialFingerprints`.
- `baselineState`.

Unknown or missing SARIF fields are ignored for now. Missing fingerprints are replaced with a stable hash based on tool, rule, path, line, and message.

When a SARIF result has a file path and line number but no embedded snippet, `sarif-html` attempts to load the snippet from `--source-root`. Pass the flag multiple times when merged reports use paths relative to different fixture roots:

```sh
go run ./cmd/sarif-html reports/*.sarif \
  --source-root . \
  --source-root fixtures/scan-targets/go-bad \
  --source-root fixtures/scan-targets/semgrep-bad \
  --out report.html
```

## Source Links

There are three ways to create source links.

### Automatic CI Links

When no link flags are provided, `sarif-html` checks common CI environment variables:

- GitHub Actions: `GITHUB_SERVER_URL`, `GITHUB_REPOSITORY`, and `GITHUB_SHA`.
- GitLab CI: `CI_PROJECT_URL`, `CI_MERGE_REQUEST_SOURCE_BRANCH_SHA`, `CI_COMMIT_SHA`, and `CI_COMMIT_REF_NAME`.

This usually links to the exact ref being built. In pull request or merge request pipelines, that is often the generated merge commit or the MR source SHA, depending on the CI provider and pipeline mode.

### Repository + Revision

When `--repo-url` and `--revision` are provided, file locations become clickable links.

GitHub-style URLs:

```text
https://github.com/org/repo/blob/<revision>/<path>#L<line>
```

GitLab-style URLs:

```text
https://gitlab.com/org/repo/-/blob/<revision>/<path>#L<line>
```

If a SARIF location already contains an `http://` or `https://` URI, `sarif-html` uses that URI directly.

### Source URL Templates

For pull request and merge request integrations, a fixed repository URL is often not enough. Use `--source-url-template` when the link should point at a provider-specific review ref, branch, commit, or custom code browser.

```sh
sarif-html report.sarif \
  --revision "$GITHUB_SHA" \
  --source-url-template "https://github.com/org/repo/blob/{revision}/{path}{lineFragment}" \
  --out report.html
```

Supported placeholders:

| Placeholder | Description |
| --- | --- |
| `{path}` | URL-escaped file path, preserving `/`. |
| `{pathRaw}` | Raw file path from the normalized finding. |
| `{line}` / `{startLine}` | Start line, or empty when missing. |
| `{endLine}` | End line, or empty when missing. |
| `{column}` / `{startColumn}` | Start column, or empty when missing. |
| `{endColumn}` | End column, or empty when missing. |
| `{revision}` | URL-escaped revision value. |
| `{revisionRaw}` | Raw revision value. |
| `{lineFragment}` | `#L<line>` or `#L<start>-L<end>`, empty when no line exists. |

## Custom HTML Templates

By default, `sarif-html` uses its embedded self-contained report template. Pass
`--template` to render the same normalized report data with your own Go
`html/template` file or template directory:

```sh
sarif-html reports/*.sarif \
  --template .github/sarif-html/report.tmpl \
  --out report.html \
  --fail-on error
```

The repository includes a complete example template with a very different visual
direction from the default report:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif \
  --title "Pixel Console SARIF" \
  --template examples/templates/pixel-console/report.tmpl \
  --out pixel-console-report.html
```

See [`examples/templates/pixel-console/report.tmpl`](examples/templates/pixel-console/report.tmpl)
for a self-contained pixel-art report with embedded CSS, JavaScript filtering, a
scrolling ticker, and scanline animations.

### Template File Shape

A template should render a complete HTML document. The simplest valid template
is a single file that uses the root data object:

```gotemplate
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>
</head>
<body>
  <h1>{{ .Title }}</h1>
  <p>{{ .Report.Summary.Total }} findings generated at {{ .GeneratedAt }}</p>
  <ul>
    {{ range .Report.Findings }}
    <li>
      <strong>{{ .Level }}</strong>
      {{ .RuleID }} in {{ .Path }}:{{ line .StartLine }}
    </li>
    {{ end }}
  </ul>
</body>
</html>
```

Use normal HTML, CSS, and JavaScript. Go's `html/template` will escape SARIF
values according to where they are rendered, including text, attributes, URLs,
CSS, and JavaScript contexts. Keep dynamic data inside template actions and let
the renderer escape it:

```gotemplate
<article
  class="finding {{ severityClass .Level }}"
  data-search="{{ .Tool }} {{ .RuleID }} {{ .Path }} {{ .Message }}">
  <h2>{{ .Message }}</h2>
  {{ if .SourceLink }}
    <a href="{{ .SourceLink }}">{{ .Path }}:{{ line .StartLine }}</a>
  {{ else }}
    <span>{{ .Path }}:{{ line .StartLine }}</span>
  {{ end }}
</article>
```

### CSS, Design, and Animation

Templates may include embedded CSS. Prefer self-contained styling so the HTML
artifact works offline in CI. This is a good place to create a completely custom
visual language: dashboard, terminal, newspaper, pixel UI, printable audit
sheet, team-branded report, or a dense engineering console.

```gotemplate
<style>
  :root {
    color-scheme: dark;
    --bg: #0b0f12;
    --panel: #17212b;
    --text: #e9ffcf;
    --accent: #c8ff41;
  }

  body {
    margin: 0;
    background:
      linear-gradient(90deg, rgba(200, 255, 65, 0.08) 1px, transparent 1px) 0 0 / 24px 24px,
      var(--bg);
    color: var(--text);
    font: 15px/1.55 "Courier New", monospace;
  }

  .finding {
    border: 4px solid var(--accent);
    background: var(--panel);
    box-shadow: 8px 8px 0 #050708;
    animation: card-in 420ms steps(4, end) both;
  }

  @keyframes card-in {
    from { opacity: 0; transform: translate(8px, 8px); }
    to { opacity: 1; transform: translate(0, 0); }
  }
</style>
```

### JavaScript

Templates may also include JavaScript for local interactions such as filtering,
sorting, expanding sections, charts, keyboard navigation, or animations. Keep it
offline-friendly and avoid remote scripts unless your organization explicitly
allows them for generated CI artifacts.

```gotemplate
<input id="search" type="search" placeholder="Search findings">

<script>
  (function () {
    const search = document.getElementById("search");
    const cards = Array.from(document.querySelectorAll("[data-search]"));

    search.addEventListener("input", () => {
      const query = search.value.trim().toLowerCase();
      cards.forEach((card) => {
        const haystack = card.dataset.search.toLowerCase();
        card.hidden = query && !haystack.includes(query);
      });
    });
  })();
</script>
```

### Modular Template Directories

When the template path is a directory, `sarif-html` recursively loads files with
`.tmpl`, `.gotmpl`, and `.html` extensions in lexical order. This allows a
modular layout with partials:

```gotemplate
{{ define "finding-row" }}
<tr>
  <td>{{ .ID }}</td>
  <td>{{ .Level }}</td>
  <td>{{ .Path }}:{{ line .StartLine }}</td>
  <td>{{ .Message }}</td>
</tr>
{{ end }}
```

```gotemplate
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>{{ .Title }}</title>
</head>
<body>
  <h1>{{ .Title }}</h1>
  <p>{{ .Report.Summary.Total }} findings generated at {{ .GeneratedAt }}</p>
  <table>
    {{ range .Report.Findings }}
      {{ template "finding-row" . }}
    {{ end }}
  </table>
</body>
</html>
```

### Template Authoring Checklist

- Start with `<!doctype html>` and include a viewport meta tag.
- Use `.SchemaVersion` when logging or diagnosing template compatibility.
- Iterate over `.Report.Findings` for finding-level output.
- Prefer `.SourceLink` or `sourceLink .` for clickable source locations.
- Use `line .StartLine` so missing line numbers render cleanly.
- Put values used by JavaScript into `data-*` attributes and let
  `html/template` escape them.
- Keep CSS and JavaScript local when the report must work as a CI artifact.
- Treat SARIF data as untrusted; never add helpers that mark SARIF content as
  safe HTML.
- Treat the template file itself as trusted input; do not run templates from
  unknown sources.

Custom templates receive the versioned `sarif-html.template.v1` data contract:

| Field | Description |
| --- | --- |
| `.SchemaVersion` | Template data contract identifier. Currently `sarif-html.template.v1`. |
| `.Title` | Effective report title. |
| `.GeneratedAt` | UTC render timestamp in RFC3339 format. |
| `.Report` | Normalized report with `Sources`, `Findings`, and `Summary`. |
| `.Severities` | Ordered severity counts. |
| `.Sources` | Ordered SARIF source counts derived from findings. |
| `.ToolGroups` | Findings grouped by tool name. |
| `.Rules` | Ordered rule counts. |
| `.Files` | Ordered file counts. |

Each `.Report.Findings` entry exposes the normalized finding fields used by the
default report, including `ID`, `Source`, `Tool`, `ToolVersion`, `RuleID`,
`RuleName`, `RuleDescription`, `RuleHelpURI`, `Level`, `Message`, `Path`,
`URI`, `StartLine`, `StartColumn`, `EndLine`, `EndColumn`, `Snippet`,
`Fingerprint`, `BaselineState`, `RelatedLocations`, `CodeFlows`, and
`SourceLink`.

Available template helpers:

| Helper | Description |
| --- | --- |
| `severityClass <level>` | Maps a severity to the CSS class used by the default template. |
| `line <number>` | Renders missing line numbers as `-`. |
| `sourceLink <finding>` | Builds the same source URL used by the default template. |
| `hasDetails <finding>` | Reports whether a finding has rule details, snippets, related locations, or code flows. |

SARIF data is untrusted input and is escaped by Go's `html/template` when it is
rendered. A custom template itself is trusted input: do not run templates from
untrusted sources, because template authors can intentionally add arbitrary
HTML, JavaScript, external assets, or links to the generated report.

## Project Layout

```text
.github/            CI workflow, issue templates, and pull request template
cmd/sarif-html/      CLI entrypoint
internal/sarif/      SARIF 2.1.0 parsing model
internal/report/     Normalized finding model and summary logic
internal/html/       Self-contained HTML renderer
docs/                Static project documentation
examples/            Small SARIF examples for local development
fixtures/            Intentionally vulnerable scanner targets
```

## Development

Run tests:

```sh
go test ./...
```

In restricted environments where Go cannot write to the default build cache, set `GOCACHE` explicitly:

```sh
GOCACHE=/private/tmp/sarif-html-go-cache go test ./...
```

Build the CLI:

```sh
go build ./cmd/sarif-html
```

Run the same core checks used by CI:

```sh
gofmt -w $(git ls-files '*.go')
go mod tidy
go test ./...
go vet ./...
go test -race ./...
```

GitHub Actions also runs `golangci-lint` and `govulncheck` on pull requests.

Generate the example report:

```sh
go run ./cmd/sarif-html examples/detekt-like.sarif \
  --title "sarif-html Example" \
  --repo-url https://github.com/acme/project \
  --revision main \
  --out report.html
```

## Real-World SARIF Fixtures

The prototype currently ships with a small synthetic Detekt-like SARIF example. The next testing step is to add scanner-generated reports from real tools.

To avoid installing scanner CLIs locally, use the throwaway Docker workflow:

```sh
scripts/generate-sarif-fixtures.sh all
```

This builds a local scanner image, runs it with the repository mounted at `/work`, writes generated reports to ignored `reports/`, and removes each scanner container after it exits. Use `CLEAN_IMAGE=1 scripts/generate-sarif-fixtures.sh all` to remove the image after the run too.

The sample targets live under `fixtures/scan-targets/`: insecure Docker files, bad Kotlin for Detekt, Python/JS for Semgrep, and a small vulnerable Go module for Go/security scanners.

See [`docs/tested-tools.md`](docs/tested-tools.md) for the living compatibility matrix, fixture workflow, and candidate CLIs that can generate SARIF directly. When a new fixture is added, update that page with the tool version, command, input project, fixture path, and current status.

## Documentation

Project documentation lives in `docs/` as static HTML with a shared stylesheet, plus living Markdown notes for evolving compatibility data:

- `docs/index.html`
- `docs/architecture.html`
- `docs/cli.html`
- `docs/templates.html`
- `docs/sarif-mapping.html`
- `docs/testing.html`
- `docs/tested-tools.md`
- `docs/open-source.html`

The generated report UI and the documentation UI are intentionally separate. The report should feel like a professional analysis artifact; the docs can be more editorial and project-facing.

## Roadmap

- Add real-world fixtures from the candidate tools in `docs/tested-tools.md`.
- Support SARIF URI base IDs more completely, especially monorepo and Windows path edge cases.
- Add baseline comparison: new, unchanged, and resolved findings.
- Improve large-report performance and navigation.
- Add optional grouping modes by file, rule, severity, and tool.
- Add release builds for macOS, Linux, and Windows.
- Decide and add the final open-source license before publishing.

## Open Source Readiness

Already included:

- `README.md`
- `CONTRIBUTING.md`
- `CHANGELOG.md`
- GitHub issue templates
- GitHub pull request template
- GitHub Actions CI workflow
- GoReleaser release workflow
- Dependabot configuration
- `.editorconfig`
- `.golangci.yml`

Before publishing the repository, we should still decide:

- Repository host and final module path, for example `github.com/<owner>/sarif-html`.
- License, for example MIT or Apache-2.0.
- Branch protection rules for `main`.
- Real `CODEOWNERS` entries after the GitHub owner or maintainer team exists.

## Releases

Releases are published by pushing a semantic version tag:

```sh
git tag v0.1.0
git push origin v0.1.0
```

The release workflow uses GoReleaser to build Linux, macOS, and Windows binaries
for `amd64` and `arm64`, attach archives to the GitHub Release, and publish
SHA-256 checksums. Pull requests that touch release configuration run a snapshot
build through `release-check.yml`.

## License

No license has been selected yet. Choose and add a `LICENSE` file before publishing this project as open source.
