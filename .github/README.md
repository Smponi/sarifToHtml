# GitHub Configuration

This directory contains repository automation and contribution templates for the
GitHub-hosted project.

Important files:

- `workflows/ci.yml`: required checks for formatting, linting, tests, vet, race
  tests, and vulnerability scanning.
- `workflows/release.yml`: tag-based release publishing through GoReleaser.
- `workflows/release-check.yml`: snapshot release validation for pull requests.
- `dependabot.yml`: dependency update configuration for Go modules and GitHub
  Actions.
- `CODEOWNERS`: template for protected-path ownership once the GitHub owner or
  maintainer team exists.
