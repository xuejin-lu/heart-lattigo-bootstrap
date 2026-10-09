# FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015 journal

## Checkpoint A — source and reuse audit

- Primary `main`: `d097a3f638431ed862e8f0da34956480e348c937`, clean after safe sync.
- Secondary `fast-qprefix`: `0c92e7ed071a1be6eb258474811db470e4463df7`, clean and equal to `origin/fast-qprefix` after safe sync.
- Genuine Standard reference: `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (must be rechecked before matched validation).
- End goal remains unchanged public CKKS frontend/profile with only the Lattigo backend changed; the bounded chain is public Add → MulRelin → Rescale → Rotate.

### Existing implementations and dispatch

| Boundary | Existing source path | Finding before this batch |
|---|---|---|
| Public `ckks.Evaluator.MulRelin` | `schemes/ckks/evaluator.go` → `mulRelinFastCKKSZeroSecret` in `evaluator_fast_zero_secret.go` | Ciphertext-pair Fast path currently requires full active-Q materialization, NTT/non-Montgomery representation, degree one, matching batching/domain, and c1=0; it computes into scratch with native multiplication `relin=false`, so no relin-key lookup/GadgetProduct. This rejects compact high-Level rows before arithmetic. |
| Explicit Fast MulRelin | `schemes/ckks/fast/evaluator.go` → `mulElementRows` in `fast/evaluator_ntt.go` → `pointMulPrefixRows` in `fast/prefix_kernels.go` | Existing exact q-prefix product kernel; handles supported NTT/Montgomery rows, uses evaluator-owned three-polynomial scratch, truncates degree two for zero-secret MulRelin, and keeps only requested rows. |
| MulRelin oracle/capacity | `fast/prefix_kernels_test.go`; `fast/qprefix_capacity.go`; `FAST_QPREFIX_SPEC.md` | Existing tests independently check each q-row modular product formula. Existing exact centered observer reconstructs coefficients from actual prefix residues and enforces `2B < S_Q`; paired numeric fixtures in this batch must make that observer a hard test assertion. |
| Public `Rescale` | `schemes/ckks/evaluator.go` → native `DivRoundByLastModulusManyNTT`, after full-active-row guards | No shared Fast rescale dispatch yet. |
| Explicit Fast `Rescale` | `schemes/ckks/fast/rescale.go` → `rescaleNQPrefix` / `rescaleQPrefixComponent` in `rescale_qprefix.go` | Existing transactional bounded centered-CRT implementation uses logical top-q divisors, validates capacity after every transition, and contracts rows to the target Level prefix. It is the implementation to reuse/migrate, not duplicate. |
| Public `Rotate` | `schemes/ckks/evaluator.go` → `rotateFastCKKSZeroSecret` | Already uses `fastcore.AutomorphismCore`, but requires every logical Q row and all rows of c1 exactly zero. |
| Explicit Fast `Rotate` | `schemes/ckks/fast/evaluator.go` → `AutomorphismQPrefixRows` → `fastcore.AutomorphismWorkspace.ApplyRows` | Already shares the same permutation kernel and accepts explicit authoritative prefix rows. |
| Public Add/Sub | `evaluator_fast_zero_secret_add.go` → `fastcore.AddSubCore` | Existing accepted Batch 014 bridge; preserve unchanged behavior and reuse shared compact allocator/core. |

### Frozen compact MulRelin contract

- Supported public path: ciphertext × ciphertext, both degree one; matching configured NTT domain and batching metadata; NTT, non-Montgomery; c1 exactly zero over authoritative rows; canonical residues and complete N-sized backing for those rows; valid metadata/dimensions/levels; output remains transactional on rejection, including aliases.
- Output remains logical `min(Level0, Level1, output Level)`, degree one, `Scale0*Scale1`, c1 exactly zero. No relin key, KeySwitch, or GadgetProduct is permitted.
- Row authority is `w_Q(Level)=min(Level+1,4)`. Level-5 structural fixtures must use six Q primes while retaining only q0..q3; q4/q5 stay dormant. This structural check is not a numeric decoding claim.
- Existing product oracle is the rowwise modular NTT product formula in `TestPrefixCompleteTruncateAndDegreeOneMulFormulas` (plus the Montgomery row oracle); it does not by itself prove centered decoding capacity. For decoded fixtures, use the existing exact q-prefix observer and strict `2B < S_Q` condition as a hard assertion.
- Existing focused baseline tests passed with `GOWORK=off GOCACHE=/private/tmp/codex-batch015-gocache.ARYgww`: `go test ./schemes/ckks/fast -run 'TestPrefixCompleteTruncateAndDegreeOneMulFormulas|TestPublicCKKSEvaluatorKeylessZeroSecretMulRelin|TestFastRescaleMatchesBigIntCenteredCRTOracleAtEveryPrefixWidth' -count=1`.

### Classification and implementation direction

The first gap is `API_ADAPTER` / `INTEGRATION_DISPATCH`, not `MISSING_CORE`: explicit Fast MulRelin and Rescale kernels already exist; Rotate and Add/Sub already use import-neutral shared cores. The `ckks` parent cannot import `ckks/fast` without a cycle, so move/share only the existing kernels at `schemes/ckks/internal/fastcore` and keep thin public/explicit adapters. No Standard source, frontend, parameter, or historical branch is in scope.

Unsupported public scalar/plaintext/mixed MulRelin overloads must be inventoried and made visibly fail closed when they would otherwise send compact backing through generic full-RNS code; this batch does not implement those overloads.

## Checkpoint B — public compact MulRelin

- Extracted the existing Fast row-product/relinearization scratch kernel into import-neutral `schemes/ckks/internal/fastcore.MulWorkspace`; both `ckks.Evaluator.MulRelin(New)` and explicit `fast.Evaluator.MulRelinElementQPrefixRows` dispatch through that core. Recording tests prove both routes call the shared core.
- Public Fast ciphertext-pair MulRelin retains degree-one, NTT/non-Montgomery, matching metadata, canonical fixed-prefix, and zero-c1 requirements. Output retains logical Level and product Scale, degree one, compact Q rows, and zero c1. No Fast relin-key/Galois-key retrieval is required by the actual operations; the RLWE evaluator constructor enumerates the supplied Galois-key list once.
- Level 1/3 exact q-prefix capacity observer checks passed; the separate six-Q-prime Level-5 case verifies only q0..q3 structure with q4/q5 absent/dormant.
- Added malformed `MulRelinNew` and `MulRelin` cases for absent components, nil metadata, empty backing, and inconsistent logical rows. Errors occur before Level dereference/output allocation for the new API; existing inputs and output remain unchanged on failure.

## Checkpoint C — Rescale / Rotate and composition

- Public CKKS and explicit Fast Rescale now share the existing bounded transactional `fastcore.RescaleWorkspace`; ordinary wrappers pass fixed `QPrefixWidth(Level)`. Top logical q remains the divisor, with Standard Level/Scale progression.
- **Review confirmation:** normal public production calls do not provide narrower `sourceRows`; they always use the fixed width. The Level-3 / one-row construction is therefore not a public production adaptive-width correctness bug. The explicitly row-parameterized API does promise exact authority, and its no-op branch previously copied policy-width rows. The repaired no-op copies only the explicit rows and clears all higher backing. A true Rescale also clears target rows beyond `min(sourceRows, QPrefixWidth(targetLevel))`. These are representation/authority fixes, not a new limb-selection algorithm.
- Level 0 true Rescale remains rejected; the tested Level-3 RescaleTo no-op preserves Level and Scale. RescaleTo Level 0 continues to match the existing Standard API error behavior.
- Public Rotate continues to reuse shared `fastcore.AutomorphismWorkspace` with the fixed Q-prefix policy. The Level-5 public Add→MulRelin→Rescale→Rotate fixture remains structural-only and keeps q4/q5 dormant.
- Updated stale Rescale `maintainedLimbCount`/legacy-width annotations and adjacent stale complete-active-Q wording to fixed Q-prefix authority.

## Checkpoint D — matched runner and validation

- Matched public runner uses identical source/profile/message and native Standard keygen/encrypt/decrypt. Profile: LogN13, Q bits `[55,39,40,39]`, P bits `[60]`, scale `2^45`, `LogSlots=4`, Levels 1 and 3. Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Fast `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
- All eight Add/MulRelin/Rescale/Rotate checkpoints passed for Standard and Fast. Fast-vs-Standard decoded RMSE and max were both zero at recorded float precision. Fast direct `GetRelinearizationKey` / `GetGaloisKey` counters were zero; each evaluator's constructor called `GetGaloisKeysList` once to build its automorphism index.
- Primary during the runs was `d097a3f638431ed862e8f0da34956480e348c937` with this batch's runner/journal uncommitted (`primary_dirty=true`); source, profile, effective Q/P, and deterministic-input hashes are persisted rather than claiming a clean Primary run. Standard and Fast backends were both clean; Secondary commit `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` was pushed normally.
- Focused `go test` allowlist passed in `./schemes/ckks/internal/fastcore`, `./schemes/ckks`, and `./schemes/ckks/fast`, including the new review regressions, q-prefix capacity/CRT oracles, public/explicit core dispatch, and the compact composition. `go vet` on the same packages passed. The allowlist excluded the LogN16 tests. `git diff --check` passed before the Secondary commit.
- No Bootstrap, benchmark, or LogN16 execution was performed. Result evidence: `FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-evidence.json`; reader-friendly summary: `FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-summary.md`.

**Batch status:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`. The authorized Primary runner and report artifacts are part of this task's commit; no further experiment is authorized by this batch.
