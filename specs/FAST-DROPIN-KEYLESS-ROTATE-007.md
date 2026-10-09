# FAST-DROPIN-KEYLESS-ROTATE-007

**Status:** READY_FOR_CODEX. **Class:** M-approved zero-secret algebra / bounded I implementation and inexpensive E preflight.
**Authority:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`; `results/FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A-web-review.md`.
**Previous stage:** `FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A` ACCEPTED.
**Pinned genuine Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (immutable).
**Fast starting SHA:** `93ab7ecc5a864fd9bbacea55f0af730941552691`.

## Frozen invariant

For Fast simulated CKKS zero-secret degree-one ciphertext `(c0,0)` with **all logical active Q rows correctly materialized**, slot rotation by k corresponds to ring automorphism `sigma_g`: `(c0,0) -> (sigma_g(c0),0)`; no Galois/rotation key lookup, native `GadgetProduct`, or native KeySwitch is mathematically necessary. This intentionally insecure simulation does **not** preserve Standard encryption security or noise. Keep logical Level/Scale, metadata and correct slot permutation exactly as the normal CKKS API; preserve every active Q limb, output degree one, c1=0.

## Public API and change boundary

1. In the Fast fork only, transparently implement keyless `ckks.Evaluator.Rotate(ct,k,out)` and `RotateNew(ct,k)` behind existing signatures, under the established CKKS Fast marker/capability. `ckks.NewEvaluator` must continue returning `*ckks.Evaluator`; no special Fast constructor or frontend change. Do not import sibling `schemes/ckks/fast` into `schemes/ckks` (import cycle), and do not change pinned Standard code, ordinary RLWE/BFV/BGV, original `schemes/ckks/fast` Q-prefix automorphism or Bootstrap.
2. Use existing `ring.Ring` full-active-Q NTT automorphism/valid index generation with a verified k-to-Galois mapping. Only support carefully source-verified representations; reject unsupported NTT/Montgomery/ring types before mutation. For initial pilot restrict to ring.Standard, degree-one, c1=0, correctly backed **all logically active** Q rows. Do not silently apply q0/q1-only Fast automorphism while keeping stale higher Q limbs.
3. **Fail closed** for Fast-marked Rotate/RotateNew ciphertexts with active nonzero c1, malformed backing/metadata, unsupported domain/degree/ring, invalid rotations or insufficient destination. Errors must be descriptive, independent of whether ordinary or Fast layout Galois keys exist; do not fall through to native RLWE Automorphism/KeySwitch. No automatic full-RNS fallback on compact Q-prefix or stale rows. Preserve input and output **bitwise** on rejected calls; support in-place `Rotate(ct,k,ct)` via scratch or other mathematically justified alias-safe technique.
4. Valid zero-secret input must work with `ckks.NewEvaluator(params,nil)` (no Galois keys). A tracking keyset with a real normally generated Standard-layout Galois key must show **zero GetGaloisKey lookups** for valid and invalid Fast cases. Standard dependency must still use its authentic KeySwitch path and no special Fast semantics.
5. Preserve `RotateNew` source-compatible return type, and unchanged Scale/Level/LogDimensions/IsNTT flags, zero c1. If a natural common internal keyless automorphism hook can be reused for public `Conjugate`, do **not** expand to Conjugate in this task; reserve it separately.

## Tests, experiments, handoff

- Safe sync Primary+Secondary via updated AGENTS; reread active task/spec and Secondary authoritative Fast Q-prefix spec; stop on unexpected branch/provenance conflicts. Scope code changes to Secondary ordinary `schemes/ckks/evaluator.go`, its existing `evaluator_fast_zero_secret.go` helper, and targeted tests (or a minimal independent file if necessary). No unrelated production edits.
- Secondary public API tests: keyless Rotate/RotateNew at **two logical Levels** and rotations k=0, +1, -1, with a **plaintext slot-rotation oracle** and non-default LogSlots; check c1=0, all active-Q backing, Level/Scale/dimensions, in-place aliasing, no Galois key access, out-of-domain/degree/c1/backing/metadata negative tests, transactional failure even when a valid Standard-layout Galois key exists. Tests must be bounded and not invoke Bootstrap.
- Reuse **exact same immutable shared frontend** `tools/fast-dropin-ckks-primitive-api-audit-005/main.go`, source SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`, and exact fixed profile. Run one Fast dependency numerical regression if static/preflight accepts. Existing pinned Standard output may be reused with matched hashes, no Standard rerun required. Do not change the frontend, inputs/params, or compare new profile data against mismatched Standard.
- Run `go test ./schemes/ckks ./schemes/ckks/fast ./core/rlwe -count=1`, `git diff --check`, one self-review/repair. No full suite requirement, no Bootstrap, benchmark, LogN16, parameter sweep, or claims of runtime speedup.
- Write compact Primary `results/FAST-DROPIN-KEYLESS-ROTATE-007-summary.md`, with actual SHA, source-map/diff, keylookup-free proof, numerical slot oracle and metadata, tests/command, provenance and untested cases. Keep historical reports unchanged. Safe fast-forward pushes only; if a semantic conflict is found, stop `NEEDS_WEB_REVIEW`.
- Classifications: `FAST_DROPIN_KEYLESS_ROTATE_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_KEYLESS_ROTATE_PARTIAL`, `FAST_DROPIN_KEYLESS_ROTATE_BLOCKED` plus `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
