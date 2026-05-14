# Changelog

All notable changes to `sarif-html` will be documented in this file.

## 1.0.0 - 2026-05-14

- Initial stable Go CLI release.
- SARIF 2.1.0 parsing and normalization.
- Self-contained HTML report renderer.
- Custom template support through the versioned `sarif-html.template.v1` data contract.
- Versioned `sarif-html.baseline.v1` baselines and CI threshold gates.
- Static HTML documentation in `docs/`.
- Unit tests for parser, normalizer, renderer, and CLI behavior.
