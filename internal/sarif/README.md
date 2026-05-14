# SARIF Package

`internal/sarif` owns the subset of the SARIF 2.1.0 schema that the renderer
currently understands.

The package should stay close to the wire format:

- Struct fields map directly to SARIF JSON properties.
- Validation should cover basic input integrity, such as version and runs.
- Higher-level interpretation belongs in `internal/report`.

When adding SARIF fields, prefer extending the model with focused tests that use
real or representative scanner output.
