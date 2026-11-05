# go-db-compare

Exact scalar comparison and deterministic keyed inventory alignment. Use `value` and `keyed` functional packages. The module has no dependencies beyond Go 1.27.3's standard library. Collection, SQL, identifiers, normalization choices, display and acceptance policy remain caller-owned.

`value.Cell` distinguishes known text/integer, SQL NULL, authoritative absence and unavailable data. Empty text and zero are real values. `value.Compare` preserves both inputs, classifies equality/case/one-sided differences and calculates arbitrary-precision integer target-minus-source deltas. No floats, trimming, sentinel strings or assumed missing zeros are involved. Two absences remain unavailable.

`value.Analyze` supports caller-selected trimming for existing textual consumers. Unequal malformed UTF-8 bytes stay different and never become case-only. `keyed.Align` joins a sorted union, rejects duplicates without a prefix and uses snapshot completeness to distinguish absence from unavailable enumeration. Values and rows own their returned data.

Limits: one MiB exact cell text, 4,096 integer digits/reason bytes/key bytes and 100,000 entries per snapshot. These are comparison contracts, not database completeness or migration certification.

Run `scripts/validate.ps1 -Race -Scan` or `scripts/validate.sh --race --scan` with approved external storage. `examples/compare` demonstrates a delta beyond binary-float precision. Read CONTRIBUTING.md and docs/code-management.md for comment, validation and release rules.

License: UPL 1.0.