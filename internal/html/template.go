package html

const pageTemplate = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{ .Title }}</title>
  <style>` + reportCSS + `</style>
</head>
<body>
  <header class="topbar">
    <div class="topbar-inner">
      <div class="title-block">
        <span class="report-kicker">SARIF analysis</span>
        <h1>{{ .Title }}</h1>
        <p>{{ .Report.Summary.Total }} findings across {{ len .Report.Sources }} source file(s) · generated {{ .GeneratedAt }}{{ if .Baseline.Enabled }} · {{ .Baseline.Unchanged }} in baseline{{ end }}</p>
      </div>
      <div class="sources" role="group" aria-label="Filter by SARIF source">
        {{ range .Sources }}<button class="source-chip" type="button" data-source-chip="{{ .Name }}" aria-pressed="false" title="Filter {{ .Name }}"><span>{{ .Name }}</span><strong>{{ .Count }}</strong></button>{{ end }}
      </div>
    </div>
  </header>

  <main class="page-main">
    <section class="summary" aria-label="Summary">
      <div class="metric metric-total">
        <span class="metric-value">{{ .Report.Summary.Total }}</span>
        <span class="metric-label">Total findings</span>
      </div>
      {{ range .Severities }}
      <div class="metric {{ severityClass .Name }}">
        <span class="metric-value">{{ .Count }}</span>
        <span class="metric-label">{{ .Name }}</span>
      </div>
      {{ end }}
    </section>

    <section class="command-bar" aria-label="Search and filters">
      <label class="search-label">
        Search findings
        <input id="search" type="search" placeholder="Search message, rule, file, source, or tool">
      </label>
      <div class="filter-row" aria-label="Facet filters">
        <label>
          Severity
          <select id="severity">
            <option value="">All severities</option>
            {{ range .Severities }}<option value="{{ .Name }}">{{ .Name }} ({{ .Count }})</option>{{ end }}
          </select>
        </label>
        <label>
          Source
          <select id="source">
            <option value="">All SARIF files</option>
            {{ range .Sources }}<option value="{{ .Name }}">{{ .Name }} ({{ .Count }})</option>{{ end }}
          </select>
        </label>
        <label>
          Rule
          <select id="rule">
            <option value="">All rules</option>
            {{ range .Rules }}<option value="{{ .Name }}">{{ .Name }} ({{ .Count }})</option>{{ end }}
          </select>
        </label>
        <label>
          File
          <select id="file">
            <option value="">All files</option>
            {{ range .Files }}<option value="{{ .Name }}">{{ .Name }} ({{ .Count }})</option>{{ end }}
          </select>
        </label>
      </div>
      {{ if .Baseline.HideUnchangedByDefault }}
      <div class="baseline-row">
        <label class="baseline-toggle">
          <input id="show-baseline" type="checkbox">
          <span>Show baseline findings</span>
          <strong>{{ .Baseline.Unchanged }}</strong>
        </label>
      </div>
      {{ end }}
      <div class="search-meta">
        <strong id="result-count">{{ .Report.Summary.Total }} shown</strong>
        <button id="reset-filters" type="button">Reset filters</button>
      </div>
      <div class="view-switch" role="group" aria-label="Report view">
        <button class="view-toggle is-active" type="button" data-view-toggle="review" aria-pressed="true">Review</button>
        <button class="view-toggle" type="button" data-view-toggle="compact" aria-pressed="false">Compact dashboard</button>
      </div>
    </section>

    <section id="review-view" class="findings-panel view-panel" data-view-panel="review" aria-label="Findings">
      <div class="findings-head">
        <div>
          <span class="panel-kicker">Findings</span>
          <h2>Review queue</h2>
        </div>
        <p>Scan the title and location first. Evidence stays nearby, details stay folded until needed.</p>
      </div>
      <div id="findings" class="findings-list">
        {{ range .Report.Findings }}
        <article class="finding{{ if .IsBaseline }} hidden{{ end }}" id="{{ .ID }}" data-id="{{ .ID }}" data-severity="{{ .Level }}" data-source="{{ .Source }}" data-tool="{{ .Tool }}" data-rule="{{ .RuleID }}" data-file="{{ .Path }}" data-baseline="{{ .IsBaseline }}" data-baseline-state="{{ .BaselineState }}" data-search="{{ .Source }} {{ .Tool }} {{ .RuleID }} {{ .RuleName }} {{ .Path }} {{ .Message }} {{ .BaselineState }}">
          <header class="finding-top">
            <div class="finding-status">
              <a class="finding-id" href="#{{ .ID }}">{{ .ID }}</a>
              <span class="badge {{ severityClass .Level }}">{{ .Level }}</span>
              {{ if .IsBaseline }}<span class="badge baseline-badge">baseline</span>{{ end }}
            </div>
            <div class="finding-copy">
              <h3>{{ .Message }}</h3>
              <p>
                {{ $link := sourceLink . }}
                {{ if $link }}<a href="{{ $link }}">{{ .Path }}:{{ line .StartLine }}</a>{{ else }}{{ .Path }}:{{ line .StartLine }}{{ end }}
              </p>
            </div>
          </header>
          <dl class="finding-facts" aria-label="Finding evidence">
            <div><dt>Rule</dt><dd><code>{{ .RuleID }}</code>{{ if .RuleName }}<span>{{ .RuleName }}</span>{{ end }}</dd></div>
            <div><dt>Tool</dt><dd>{{ .Tool }}{{ if .ToolVersion }} <span>{{ .ToolVersion }}</span>{{ end }}</dd></div>
            <div><dt>Source</dt><dd>{{ .Source }}</dd></div>
            <div><dt>Fingerprint</dt><dd><code>{{ .Fingerprint }}</code></dd></div>
          </dl>
          {{ if hasDetails . }}
          <details class="details">
            <summary>Details</summary>
            <div class="detail-grid">
              <div>
                {{ if .RuleDescription }}<p class="description">{{ .RuleDescription }}</p>{{ end }}
                <dl>
                  {{ if .BaselineState }}<dt>Baseline</dt><dd>{{ .BaselineState }}</dd>{{ end }}
                  {{ if .RuleHelpURI }}<dt>Rule docs</dt><dd><a href="{{ .RuleHelpURI }}">{{ .RuleHelpURI }}</a></dd>{{ end }}
                </dl>
              </div>
              {{ if .Snippet }}
              <pre class="snippet"><code>{{ .Snippet }}</code></pre>
              {{ end }}
            </div>
            {{ if .RelatedLocations }}
            <h3>Related locations</h3>
            <ol class="compact-list">
              {{ range .RelatedLocations }}<li><code>{{ .Path }}:{{ line .StartLine }}</code> {{ .Message }}</li>{{ end }}
            </ol>
            {{ end }}
            {{ if .CodeFlows }}
            <h3>Code flows</h3>
            {{ range .CodeFlows }}
            <ol class="compact-list">
              {{ range .Locations }}<li><code>{{ .Path }}:{{ line .StartLine }}</code> {{ .Message }}</li>{{ end }}
            </ol>
            {{ end }}
            {{ end }}
          </details>
          {{ end }}
        </article>
        {{ end }}
      </div>
      <div id="empty-state" class="empty-state hidden">No findings match the current filters.</div>
    </section>

    <section id="compact-view" class="compact-dashboard view-panel hidden" data-view-panel="compact" aria-label="Compact tool dashboard">
      <div class="compact-head">
        <div>
          <span class="panel-kicker">Compact mode</span>
          <h2>Tool dashboards</h2>
        </div>
        <p>Each tool gets its own compact table, so clusters and outliers are visible at a glance.</p>
      </div>
      <div class="tool-grid">
        {{ range .ToolGroups }}
        <section class="tool-card" data-tool-section="{{ .Name }}">
          <header class="tool-card-head">
            <div>
              <span>Tool</span>
              <h3>{{ .Name }}</h3>
            </div>
            <strong><span data-tool-visible>{{ .Count }}</span><small>/ {{ .Count }}</small></strong>
          </header>
          <div class="compact-table-wrap">
            <table class="compact-table">
              <thead>
                <tr>
                  <th>Finding</th>
                  <th>Message</th>
                  <th>Rule</th>
                  <th>Source</th>
                </tr>
              </thead>
              <tbody>
                {{ range .Findings }}
                <tr class="compact-row{{ if .IsBaseline }} hidden{{ end }}" data-id="{{ .ID }}" data-severity="{{ .Level }}" data-source="{{ .Source }}" data-tool="{{ .Tool }}" data-rule="{{ .RuleID }}" data-file="{{ .Path }}" data-baseline="{{ .IsBaseline }}" data-baseline-state="{{ .BaselineState }}" data-search="{{ .Source }} {{ .Tool }} {{ .RuleID }} {{ .RuleName }} {{ .Path }} {{ .Message }} {{ .BaselineState }}">
                  <td>
                    <a class="compact-open" href="#{{ .ID }}" data-open-finding="{{ .ID }}">{{ .ID }}</a>
                    <span class="severity-text {{ severityClass .Level }}">{{ .Level }}</span>
                    {{ if .IsBaseline }}<span class="baseline-text">baseline</span>{{ end }}
                  </td>
                  <td>
                    <strong>{{ .Message }}</strong>
                    {{ $link := sourceLink . }}
                    <span>{{ if $link }}<a href="{{ $link }}">{{ .Path }}:{{ line .StartLine }}</a>{{ else }}{{ .Path }}:{{ line .StartLine }}{{ end }}</span>
                  </td>
                  <td><code>{{ .RuleID }}</code></td>
                  <td>{{ .Source }}</td>
                </tr>
                {{ end }}
              </tbody>
            </table>
          </div>
          <div class="tool-empty hidden">No matches for this tool.</div>
        </section>
        {{ end }}
      </div>
      <div id="compact-empty-state" class="empty-state hidden">No tool tables match the current filters.</div>
    </section>
  </main>
  <script>` + reportJS + `</script>
</body>
</html>`

const reportCSS = `
:root {
  color-scheme: light;
  --bg: #fbf7ef;
  --canvas: #fffefa;
  --surface: rgba(255, 254, 250, 0.86);
  --surface-strong: #fffefa;
  --warm: #fff3de;
  --cool: #edf9f7;
  --ink: #201a16;
  --text: #342d27;
  --muted: #74695f;
  --line: #e2d5c6;
  --line-strong: #bfa890;
  --accent: #9b2f50;
  --accent-strong: #721b35;
  --teal: #006b70;
  --teal-strong: #004f54;
  --gold: #7a5300;
  --error: #982f24;
  --warning: #735000;
  --note: #006b70;
  --none: #5c6067;
  --code: #f6efe5;
  --shadow: 0 24px 70px rgba(67, 48, 30, 0.13);
  --shadow-soft: 0 10px 30px rgba(67, 48, 30, 0.08);
}
* { box-sizing: border-box; }
body {
  margin: 0;
  background:
    linear-gradient(180deg, #fffaf1 0%, var(--bg) 52%, #f4ecdf 100%);
  color: var(--text);
  font: 15px/1.5 "Aptos", "Segoe UI", "Helvetica Neue", sans-serif;
}
body::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  background:
    linear-gradient(115deg, rgba(155, 47, 80, 0.07), transparent 32%),
    linear-gradient(245deg, rgba(0, 107, 112, 0.07), transparent 35%),
    repeating-linear-gradient(90deg, rgba(32, 26, 22, 0.025) 0 1px, transparent 1px 92px);
}
a {
  color: var(--teal-strong);
  text-decoration-thickness: 1px;
  text-underline-offset: 3px;
}
a:hover { color: var(--accent-strong); }
.topbar,
.page-main {
  position: relative;
  z-index: 1;
}
.topbar { padding: 24px 24px 0; }
.topbar-inner {
  display: grid;
  grid-template-columns: minmax(340px, 1fr) minmax(340px, 0.76fr);
  gap: 24px;
  max-width: 1660px;
  margin: 0 auto;
  border: 1px solid var(--line);
  border-radius: 8px;
  background:
    linear-gradient(135deg, rgba(255, 254, 250, 0.96), rgba(255, 243, 222, 0.82));
  box-shadow: var(--shadow);
  padding: 26px;
}
.title-block { min-width: 0; }
.report-kicker,
.panel-kicker,
label {
  color: var(--muted);
  font-size: 11px;
  font-weight: 850;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
h1 {
  margin: 5px 0 8px;
  color: var(--ink);
  font: 820 40px/1.03 "Aptos Display", "Segoe UI", sans-serif;
  letter-spacing: 0;
}
h2,
h3,
p { margin: 0; }
p { color: var(--muted); }
.page-main {
  max-width: 1660px;
  margin: 0 auto;
  padding: 18px 24px 36px;
}
.sources {
  display: flex;
  align-content: flex-start;
  align-items: flex-start;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 9px;
  min-width: 0;
}
.source-chip {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  width: auto;
  min-height: 36px;
  max-width: 290px;
  border: 1px solid #d8c6b2;
  border-radius: 999px;
  background: #fffaf1;
  color: #372f28;
  padding: 7px 11px 7px 13px;
  font: inherit;
  font-size: 13px;
  font-weight: 780;
  cursor: pointer;
  box-shadow: var(--shadow-soft);
  transition: background-color 140ms ease, border-color 140ms ease, color 140ms ease, transform 140ms ease, box-shadow 140ms ease;
}
.source-chip:hover {
  border-color: var(--accent);
  background: #fff0df;
  color: var(--ink);
  transform: translateY(-1px);
}
.source-chip.is-active,
.source-chip[aria-pressed="true"] {
  border-color: var(--ink);
  background: var(--ink);
  color: #fffaf1;
}
.source-chip span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.source-chip strong {
  min-width: 27px;
  color: var(--teal-strong);
  text-align: right;
}
.source-chip.is-active strong,
.source-chip[aria-pressed="true"] strong { color: #ffd2a8; }
.summary {
  display: grid;
  grid-template-columns: minmax(190px, 1.12fr) repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}
.metric {
  position: relative;
  min-height: 84px;
  overflow: hidden;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--surface);
  box-shadow: var(--shadow-soft);
  padding: 15px 17px;
}
.metric::after {
  content: "";
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 10px;
  height: 4px;
  border-radius: 999px;
  background: var(--teal);
}
.metric-total {
  background: var(--ink);
  color: #fffaf1;
}
.metric-total::after { background: linear-gradient(90deg, #f5b15d, #d75f75, #2d9b92); }
.metric-total .metric-label { color: rgba(255, 250, 241, 0.74); }
.metric.sev-error { background: #fff2ef; border-color: #edb3a9; color: var(--error); }
.metric.sev-error::after { background: var(--error); }
.metric.sev-warning { background: #fff7df; border-color: #e0c26e; color: var(--warning); }
.metric.sev-warning::after { background: var(--warning); }
.metric.sev-note { background: #eef9f7; border-color: #91cac4; color: var(--note); }
.metric.sev-note::after { background: var(--note); }
.metric.sev-none { background: #f3f4f5; border-color: #cfd3d8; color: var(--none); }
.metric.sev-none::after { background: var(--none); }
.metric-value {
  display: block;
  color: inherit;
  font-size: 31px;
  font-weight: 850;
  letter-spacing: 0;
}
.metric-label {
  color: var(--muted);
  text-transform: uppercase;
  font-size: 11px;
  font-weight: 850;
  letter-spacing: 0.04em;
}
.command-bar {
  display: grid;
  grid-template-columns: minmax(360px, 1fr) auto;
  gap: 14px;
  align-items: end;
  margin-bottom: 14px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background:
    linear-gradient(135deg, rgba(255, 254, 250, 0.98), rgba(237, 249, 247, 0.68));
  box-shadow: var(--shadow);
  padding: 16px;
}
.search-label {
  gap: 8px;
  color: var(--ink);
  font-size: 12px;
}
.search-label input {
  min-height: 54px;
  border-color: var(--line-strong);
  background: #fffaf1;
  padding: 11px 15px;
  font-size: 19px;
  font-weight: 730;
}
.search-label input::placeholder {
  color: #85766a;
  font-weight: 690;
}
.filter-row {
  grid-column: 1 / -1;
  display: grid;
  grid-template-columns: repeat(4, minmax(150px, 1fr));
  gap: 10px;
}
.baseline-row {
  grid-column: 1 / -1;
  display: flex;
  justify-content: flex-start;
}
.baseline-toggle {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  min-height: 38px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: #fffaf1;
  color: var(--text);
  padding: 6px 12px;
  text-transform: none;
  font-size: 13px;
  font-weight: 780;
  letter-spacing: 0;
}
.baseline-toggle input {
  width: 17px;
  min-height: 17px;
  margin: 0;
  accent-color: var(--teal);
}
.baseline-toggle strong {
  min-width: 24px;
  color: var(--teal-strong);
  text-align: right;
}
.search-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  justify-content: flex-end;
  min-width: 238px;
}
.search-meta strong {
  white-space: nowrap;
  color: var(--ink);
  font-size: 16px;
}
.view-switch {
  grid-column: 1 / -1;
  display: inline-flex;
  width: max-content;
  max-width: 100%;
  gap: 6px;
  border: 1px solid var(--line);
  border-radius: 999px;
  background: rgba(255, 250, 241, 0.76);
  padding: 5px;
}
.view-toggle {
  min-height: 34px;
  border: 0;
  border-radius: 999px;
  background: transparent;
  color: var(--muted);
  padding: 0 13px;
  font-size: 13px;
  font-weight: 820;
}
.view-toggle:hover {
  background: var(--warm);
  color: var(--ink);
}
.view-toggle.is-active,
.view-toggle[aria-pressed="true"] {
  background: var(--ink);
  color: #fffaf1;
}
label { display: grid; gap: 5px; letter-spacing: 0.03em; }
input, select, button {
  width: 100%;
  min-height: 39px;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--surface-strong);
  color: var(--text);
  padding: 7px 10px;
  font: inherit;
}
button {
  width: auto;
  min-height: 38px;
  border-color: var(--ink);
  background: var(--ink);
  color: #fffaf1;
  cursor: pointer;
  font-weight: 800;
}
button:hover { background: var(--accent-strong); border-color: var(--accent-strong); }
.source-chip:hover {
  border-color: var(--accent);
  background: #fff0df;
  color: var(--ink);
}
.source-chip.is-active:hover,
.source-chip[aria-pressed="true"]:hover {
  border-color: var(--ink);
  background: var(--ink);
  color: #fffaf1;
}
input:focus, select:focus, button:focus {
  outline: 3px solid rgba(0, 107, 112, 0.22);
  outline-offset: 2px;
}
.findings-panel {
  border: 1px solid var(--line);
  border-radius: 8px;
  background: rgba(255, 254, 250, 0.94);
  box-shadow: var(--shadow);
  overflow: hidden;
}
.findings-head {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  border-bottom: 1px solid var(--line);
  background: linear-gradient(90deg, #fff3de, #edf9f7);
  padding: 15px 18px;
}
.findings-head h2 {
  margin-top: 2px;
  color: var(--ink);
  font-size: 22px;
  line-height: 1.15;
}
.findings-head p {
  max-width: 650px;
  align-self: end;
  font-size: 13px;
  text-align: right;
}
.findings-list { display: grid; }
.finding {
  position: relative;
  display: grid;
  gap: 13px;
  border-bottom: 1px solid var(--line);
  padding: 18px 20px 17px;
  scroll-margin-top: 14px;
}
.finding::before {
  content: "";
  position: absolute;
  inset: 18px auto 18px 0;
  width: 5px;
  border-radius: 0 999px 999px 0;
  background: var(--teal);
}
.finding[data-severity="error"]::before { background: var(--error); }
.finding[data-severity="warning"]::before { background: var(--warning); }
.finding[data-severity="note"]::before { background: var(--note); }
.finding[data-severity="none"]::before { background: var(--none); }
.finding:last-child { border-bottom: 0; }
.finding:hover { background: rgba(255, 243, 222, 0.5); }
.finding-top {
  display: grid;
  grid-template-columns: 122px minmax(0, 1fr);
  gap: 18px;
  align-items: start;
}
.finding-status {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  flex-wrap: wrap;
}
.finding-id {
  color: var(--teal-strong);
  font-weight: 850;
}
.finding-copy h3 {
  color: var(--ink);
  font: 820 21px/1.24 "Aptos Display", "Segoe UI", sans-serif;
  letter-spacing: 0;
}
.finding-copy p {
  margin-top: 6px;
  font-size: 14px;
}
.finding-facts {
  display: grid;
  grid-template-columns: minmax(180px, 1.2fr) minmax(120px, 0.8fr) minmax(140px, 0.9fr) minmax(180px, 1fr);
  gap: 0;
  margin: 0 0 0 140px;
  border-top: 1px solid var(--line);
  border-bottom: 1px solid var(--line);
}
.finding-facts div {
  min-width: 0;
  padding: 10px 12px;
  border-right: 1px solid var(--line);
}
.finding-facts div:last-child { border-right: 0; }
.finding-facts dt {
  margin-bottom: 4px;
  color: var(--muted);
  font-size: 10px;
  font-weight: 850;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.finding-facts dd {
  margin: 0;
  min-width: 0;
  color: var(--text);
  font-weight: 760;
  overflow-wrap: anywhere;
}
.finding-facts dd span {
  display: block;
  margin-top: 2px;
  color: var(--muted);
  font-weight: 650;
}
.details {
  margin-left: 140px;
  padding-top: 2px;
}
.details summary {
  width: max-content;
  cursor: pointer;
  color: var(--accent-strong);
  font-weight: 820;
}
.detail-grid {
  display: grid;
  grid-template-columns: minmax(240px, 0.8fr) minmax(320px, 1.2fr);
  gap: 14px;
  margin-top: 10px;
}
.description { color: var(--text); }
dl {
  display: grid;
  grid-template-columns: 86px minmax(0, 1fr);
  gap: 5px 10px;
  margin: 10px 0 0;
}
dt { color: var(--muted); font-weight: 760; }
dd { margin: 0; min-width: 0; overflow-wrap: anywhere; }
code {
  max-width: 100%;
  overflow-wrap: anywhere;
  font-family: "SFMono-Regular", Menlo, Consolas, monospace;
  font-size: 12px;
  border: 1px solid var(--line);
  border-radius: 6px;
  background: var(--code);
  padding: 1px 5px;
}
.snippet {
  margin: 0;
  max-height: 260px;
  overflow: auto;
  border: 1px solid var(--line);
  border-radius: 8px;
  padding: 11px;
  background: var(--code);
}
.compact-list { margin: 8px 0 0; padding-left: 20px; }
.muted, .rule-name { color: var(--muted); }
.badge {
  display: inline-flex;
  align-items: center;
  min-height: 25px;
  border: 1px solid var(--line);
  border-radius: 999px;
  padding: 0 9px;
  font-size: 12px;
  font-weight: 850;
  text-transform: lowercase;
}
.sev-error { border-color: #e9a79d; background: #fff1ee; color: var(--error); }
.sev-warning { border-color: #dfbf67; background: #fff7df; color: var(--warning); }
.sev-note { border-color: #96cec7; background: #eef8f5; color: var(--note); }
.sev-none { border-color: #cdd2d7; background: #f3f4f5; color: var(--none); }
.baseline-badge {
  border-color: #91cac4;
  background: #edf9f7;
  color: var(--teal-strong);
}
.empty-state {
  padding: 30px;
  color: var(--muted);
  text-align: center;
  font-weight: 760;
}
.compact-dashboard {
  border: 1px solid var(--line);
  border-radius: 8px;
  background: rgba(255, 254, 250, 0.9);
  box-shadow: var(--shadow);
  overflow: hidden;
}
.compact-head {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  border-bottom: 1px solid var(--line);
  background: linear-gradient(90deg, #edf9f7, #fff3de);
  padding: 15px 18px;
}
.compact-head h2 {
  margin-top: 2px;
  color: var(--ink);
  font-size: 22px;
  line-height: 1.15;
}
.compact-head p {
  max-width: 650px;
  align-self: end;
  font-size: 13px;
  text-align: right;
}
.tool-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 14px;
  padding: 14px;
}
.tool-card {
  min-width: 0;
  border: 1px solid var(--line);
  border-radius: 8px;
  background: var(--surface-strong);
  box-shadow: var(--shadow-soft);
  overflow: hidden;
}
.tool-card-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: start;
  border-bottom: 1px solid var(--line);
  background: rgba(255, 243, 222, 0.56);
  padding: 12px 13px;
}
.tool-card-head span {
  display: block;
  color: var(--muted);
  font-size: 10px;
  font-weight: 850;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.tool-card-head h3 {
  margin: 2px 0 0;
  color: var(--ink);
  font-size: 18px;
  line-height: 1.15;
}
.tool-card-head strong {
  display: inline-flex;
  align-items: baseline;
  gap: 3px;
  color: var(--teal-strong);
  font-size: 24px;
  line-height: 1;
}
.tool-card-head small {
  color: var(--muted);
  font-size: 12px;
}
.compact-table-wrap {
  overflow: auto;
}
.compact-table {
  width: 100%;
  min-width: 720px;
  border-collapse: separate;
  border-spacing: 0;
}
.compact-table th,
.compact-table td {
  border-bottom: 1px solid var(--line);
  padding: 9px 10px;
  text-align: left;
  vertical-align: top;
  font-size: 13px;
}
.compact-table th {
  background: #fffaf1;
  color: var(--muted);
  font-size: 10px;
  font-weight: 850;
  letter-spacing: 0.04em;
  text-transform: uppercase;
}
.compact-table tbody tr:hover {
  background: rgba(237, 249, 247, 0.54);
}
.compact-table tbody tr:last-child td {
  border-bottom: 0;
}
.compact-table td:first-child {
  width: 112px;
  white-space: nowrap;
}
.compact-open {
  display: block;
  margin-bottom: 4px;
  font-weight: 850;
}
.severity-text {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  font-size: 12px;
  font-weight: 850;
  text-transform: lowercase;
}
.severity-text::before {
  content: "";
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: currentColor;
}
.baseline-text {
  display: inline-flex;
  margin-top: 5px;
  color: var(--teal-strong);
  font-size: 11px;
  font-weight: 850;
  text-transform: lowercase;
}
.compact-table td:nth-child(2) strong {
  display: block;
  color: var(--ink);
  font-weight: 780;
  line-height: 1.25;
}
.compact-table td:nth-child(2) span {
  display: block;
  margin-top: 3px;
  color: var(--muted);
}
.tool-empty {
  padding: 18px;
  color: var(--muted);
  text-align: center;
  font-weight: 760;
}
.hidden { display: none; }
@media (max-width: 1180px) {
  .topbar-inner { grid-template-columns: 1fr; }
  .sources { justify-content: flex-start; }
  .command-bar { grid-template-columns: 1fr; align-items: stretch; }
  .search-meta { justify-content: space-between; min-width: 0; }
}
@media (max-width: 980px) {
  .filter-row { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .tool-grid { grid-template-columns: 1fr; }
  .finding-top { grid-template-columns: 1fr; gap: 8px; }
  .finding-facts,
  .details { margin-left: 0; }
  .finding-facts { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .finding-facts div:nth-child(2n) { border-right: 0; }
  .findings-head { display: grid; }
  .findings-head p { text-align: left; }
}
@media (max-width: 720px) {
  .topbar { padding: 12px 12px 0; }
  .topbar-inner { padding: 17px; }
  h1 { font-size: 30px; }
  .page-main { padding: 12px; }
  .summary { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .command-bar { padding: 12px; }
  .search-label input { min-height: 48px; font-size: 15px; }
  .search-meta { display: grid; grid-template-columns: 1fr; }
  .search-meta button { width: 100%; }
  .view-switch { width: 100%; }
  .view-toggle { flex: 1; }
  .filter-row,
  .finding-facts,
  .detail-grid { grid-template-columns: 1fr; }
  .finding-facts div { border-right: 0; border-bottom: 1px solid var(--line); }
  .finding-facts div:last-child { border-bottom: 0; }
}
`

const reportJS = `
(function () {
  const search = document.getElementById("search");
  const severity = document.getElementById("severity");
  const source = document.getElementById("source");
  const rule = document.getElementById("rule");
  const file = document.getElementById("file");
  const showBaseline = document.getElementById("show-baseline");
  const reset = document.getElementById("reset-filters");
  const resultCount = document.getElementById("result-count");
  const emptyState = document.getElementById("empty-state");
  const compactEmptyState = document.getElementById("compact-empty-state");
  const findingRows = Array.from(document.querySelectorAll("#findings .finding"));
  const compactRows = Array.from(document.querySelectorAll(".compact-row"));
  const toolSections = Array.from(document.querySelectorAll("[data-tool-section]"));
  const sourceChips = Array.from(document.querySelectorAll("[data-source-chip]"));
  const viewToggles = Array.from(document.querySelectorAll("[data-view-toggle]"));
  const viewPanels = Array.from(document.querySelectorAll("[data-view-panel]"));
  const controls = [search, severity, source, rule, file];
  const selectedSources = new Set();

  function setView(view) {
    viewPanels.forEach(panel => panel.classList.toggle("hidden", panel.dataset.viewPanel !== view));
    viewToggles.forEach(toggle => {
      const active = toggle.dataset.viewToggle === view;
      toggle.classList.toggle("is-active", active);
      toggle.setAttribute("aria-pressed", active ? "true" : "false");
    });
  }

  function syncSourceChips() {
    sourceChips.forEach(chip => {
      const active = selectedSources.has(chip.dataset.sourceChip);
      chip.classList.toggle("is-active", active);
      chip.setAttribute("aria-pressed", active ? "true" : "false");
    });
  }

  function matchesSource(row) {
    if (selectedSources.size > 0) {
      return selectedSources.has(row.dataset.source);
    }
    return !source.value || row.dataset.source === source.value;
  }

  function matchesBaseline(row) {
    return row.dataset.baseline !== "true" || (showBaseline && showBaseline.checked);
  }

  function matches(row) {
    const query = search.value.trim().toLowerCase();
    return (!query || row.dataset.search.toLowerCase().includes(query))
      && (!severity.value || row.dataset.severity === severity.value)
      && matchesSource(row)
      && (!rule.value || row.dataset.rule === rule.value)
      && (!file.value || row.dataset.file === file.value)
      && matchesBaseline(row);
  }

  function applyFilters() {
    let visible = 0;
    findingRows.forEach(row => {
      const rowVisible = matches(row);
      row.classList.toggle("hidden", !rowVisible);
      if (rowVisible) {
        visible++;
      }
    });
    compactRows.forEach(row => row.classList.toggle("hidden", !matches(row)));
    toolSections.forEach(section => {
      const visibleRows = Array.from(section.querySelectorAll(".compact-row")).filter(row => !row.classList.contains("hidden"));
      const visibleCount = visibleRows.length;
      section.classList.toggle("hidden", visibleCount === 0);
      const count = section.querySelector("[data-tool-visible]");
      if (count) {
        count.textContent = visibleCount;
      }
      const empty = section.querySelector(".tool-empty");
      if (empty) {
        empty.classList.toggle("hidden", visibleCount !== 0);
      }
    });
    resultCount.textContent = visible + " shown";
    emptyState.classList.toggle("hidden", visible !== 0);
    compactEmptyState.classList.toggle("hidden", visible !== 0);
  }

  controls.forEach(input => {
    input.addEventListener("input", function () {
      if (input === source) {
        selectedSources.clear();
        syncSourceChips();
      }
      applyFilters();
    });
  });
  if (showBaseline) {
    showBaseline.addEventListener("change", applyFilters);
  }
  sourceChips.forEach(chip => {
    chip.addEventListener("click", function () {
      const sourceName = chip.dataset.sourceChip;
      if (selectedSources.has(sourceName)) {
        selectedSources.delete(sourceName);
      } else {
        selectedSources.add(sourceName);
      }
      if (selectedSources.size > 0) {
        source.value = "";
      }
      syncSourceChips();
      applyFilters();
    });
  });
  viewToggles.forEach(toggle => {
    toggle.addEventListener("click", function () {
      setView(toggle.dataset.viewToggle);
    });
  });
  compactRows.forEach(row => {
    const openLink = row.querySelector("[data-open-finding]");
    if (!openLink) {
      return;
    }
    openLink.addEventListener("click", function (event) {
      event.preventDefault();
      const findingID = openLink.dataset.openFinding;
      setView("review");
      const target = document.getElementById(findingID);
      if (target) {
        target.scrollIntoView({ block: "start" });
        window.location.hash = findingID;
      }
    });
  });
  reset.addEventListener("click", function () {
    controls.forEach(input => input.value = "");
    if (showBaseline) {
      showBaseline.checked = false;
    }
    selectedSources.clear();
    syncSourceChips();
    applyFilters();
    search.focus();
  });
  syncSourceChips();
  setView("review");
  applyFilters();
})();
`
