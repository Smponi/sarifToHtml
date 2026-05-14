# GitHub Actions Workflows

This directory contains the automated CI and release workflows.

`ci.yml` runs on pushes, pull requests, and manual dispatch. `release.yml`
publishes GitHub Releases for semantic version tags such as `v0.1.0`.
`release-check.yml` validates the GoReleaser configuration with a snapshot build
before release-related changes are merged.
