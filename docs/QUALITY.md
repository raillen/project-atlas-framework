# Validation and Quality

Framework quality gates:

- schemas parse and validate fixtures;
- catalogs load and IDs are unique;
- resolver produces deterministic manifests;
- Goal transitions reject illegal state changes;
- `DONE` requires evidence;
- platform compilation only uses selected workforce entries;
- generated project validates from a clean temporary directory;
- documentation entrypoints and examples stay synchronized.

GitHub Actions runs the Python test suite and a CLI smoke test. Projects using Atlas should add their own domain gates rather than treating framework validation as product verification.
