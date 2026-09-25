# Current Task

Task: BUILD-ISO-001
Status: COMPLETE

Accepted Primary implementation:
`ef3c99f997d636de6b26078eb3acbb76512478e0`

Classification:
`BUILD_ISO_001_ACCEPTED`

Accepted result:
- Standard Linux amd64 build uses `../lattigo@main` with `-tags lattigo_standard`.
- Fast Linux amd64 build uses `../lattigo@fast-ckks` without that tag.
- Fast-only diagnostic compilation units are excluded only from Standard builds.
- Standard stubs preserve the formal frontend entrypoints and fail explicitly when a Fast-only diagnostic is requested.
- Existing Fast/default source-selection behavior is preserved because `!lattigo_standard` is true for ordinary Fast builds.
- No Secondary source, CKKS/Fast arithmetic, mathematics, Bootstrap parameters, or benchmark semantics were modified.
- `scripts/build-linux-amd64.sh` produces:
  - `dist/bootstrap-standard-linux-amd64`
  - `dist/bootstrap-fast-linux-amd64`

Review note:
Most touched FIX-001 files changed only by a Go build constraint (plus its required separating blank line). Shared PSGlobal types/helpers were moved without semantic changes into `fix001_p3_global_semantics_types.go` so they remain available to the Standard compilation unit set.

Non-blocking note:
The implementation report did not run `go test ./...`; this was not an acceptance gate for BUILD-ISO-001. Default Fast build behavior remains source-equivalent apart from the shared same-package type/helper relocation.
