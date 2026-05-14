# Command Packages

This directory contains executable entrypoints for the project.

Each child directory is a standalone command that wires user-facing CLI behavior
to the reusable packages under `internal/`. Command packages should stay thin:
parse flags, handle files and exit behavior, and delegate SARIF parsing,
normalization, and rendering to internal packages.
