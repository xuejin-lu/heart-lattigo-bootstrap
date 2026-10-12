# BTS-FAST-SCALE-ALIGNMENT-005

Status: `BATCH_IN_PROGRESS`
Class: bounded CKKS public-API implementation (I)
Primary base: `da014d695c3680e486bbf5f2c9e048fbc41d2fa1`
Secondary base: `xuejin-lu/lattigo` `fast-qprefix` at `d1b6dc9963b0df95a7c149306fb62660e1020f7b`
Genuine Standard semantic reference: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`

## Objective and authority

Implement the bounded ordinary CKKS `Add`/`Sub` different-Scale integration authorized for the BTS EvalMod T3 case. Keep the application/frontend unchanged and reuse existing Fast zero-secret Q-prefix scalar and Add/Sub kernels. This Batch supersedes only the blocked implementation boundary of `BTS-FAST-PUBLIC-API-COVERAGE-004`; it does not alter, reopen, or spend Batch 004's two BTS attempts or any older journal.

The user-approved offline review classified the fixed-path capacity proof as `T1_BOUND_PROVEN_OFFLINE`. Record that proof and its restrictions here; do not edit the original diagnostic artifacts. It applies only to the saved LogN16, seed-0, `BTSModePlain`, amp-1 source path and the explicitly analyzed operations. It is not a capacity guarantee for arbitrary CKKS ciphertexts or randomness.

## Frozen provenance and source evidence

- Primary source baseline: `da014d695c3680e486bbf5f2c9e048fbc41d2fa1`.
- Fast implementation base: `d1b6dc9963b0df95a7c149306fb62660e1020f7b` on the official GitHub `fast-qprefix` branch. The five pre-existing dirty files in the original local Secondary checkout are preserved and must not be changed, synchronized by checkout, or used for implementation. Use an independent clean checkout.
- Standard behavior reference only: original, unmodified Standard source at `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Capacity source artifacts, immutable: `results/BTS-LOGN16-C2S-T1-CAPACITY-BOUNDARY-DIAGNOSTIC-summary.md` (SHA-256 `70dda549326784fa70d90f3488964cd9d3e365f94749c6761b163c0a26a97878`), `...-evidence.json` (SHA-256 `3e674dd443a24d097ca5303014aacfe98f408b0867fb8702d4affb9fac05fa65`), and `...-journal.json` (SHA-256 `f8f7971c75ae715a311154a6dbb9300b051bb61ee26aba417a1f4fffdc07a3cc`). Batch 004 summary/journal also remain unchanged.
- The old diagnostic summary's `T1_SOURCE_EVIDENCE_MISSING` label predates the accepted offline re-derivation. Preserve that historical label in its artifact; this spec records the later accepted proof and why it is different, rather than rewriting the diagnostic.

## Accepted offline capacity proof and limits

All products and inequalities below are exact integer arithmetic; no floating-point approximation is used as a proof.

1. At S2C stage 3, logical Level 1 has full logical Q `q0*q1`, and the maintained Q-prefix contains both rows. The old raw conservative bound `232076387351619387697099559` exceeds `(q0*q1-1)/2 = 9442423748056159354880`. This is not, by itself, a Fast-vs-Standard modular Rescale error: for a representative shifted by `t*q0*q1`, rounded division by `q1` changes the quotient by `t*q0`, which is zero modulo the output modulus `q0`. This establishes modular Level-1-to-0 Rescale consistency, not closeness to an unmodded ideal integer result. It does not relax centered-uniqueness requirements for any operation that actually reconstructs an integer, performs a prefix conversion, or relies on a unique lift.
2. The exact captured ModUp source has complete-Q rows and a verified q0-canonical lift with coefficient bound `B0=4294966417`; this is evidence for this captured seed/input, not a generic ciphertext bound.
3. For the audited C2S diagonal transforms, the accepted recurrence is `B_raw <= B_input*D + 1`, then `B_post <= floor((B_raw + floor(q/2))/q)`. The source-backed diagonal L1 sums and dropped primes are:

   | C2S stage | `D` | dropped `q` | exact `B_post` |
   |---|---:|---:|---:|
   | 1 | `279980005292245` | `4398034845697` (`q29`) | `273418643178` |
   | 2 | `66151809604518` | `4398058045441` (`q28`) | `4112528265648` |
   | 3 | `11838663137549` | `4398055555073` (`q27`) | `11070082260443` |

   The accepted proof review checked the actual `myBSGS` chain: diagonal L1 bounds include all terms; rotations preserve coefficient infinity norm; lazy QP automorphisms preserve the P-scaled form; under the captured zero-secret `c1=0` condition, the relevant gadget/key-switch contribution is zero; ModDown's rounded divide-P is exact for that captured P-scaled zero-lane representation. The `Conjugate(zV)+zV` real split is checked against complete-Q residues, with the subsequent Q0123 output's centered lift justified by the derived bound. These statements are specific to the reviewed source path and its checked preconditions.
4. The real split gives `B_ctReal <= 22140164520886`. `Step1_Relabel` changes metadata only. `Step2_Offset` encodes `-1/64` at scale `2^42` as coefficient `-68719476736`, so `B_T1 <= 22208883997622`. This is below the previously fixed T1 gate `354567853947127`.
5. Using `q26=4398054899713`, `N=65536`, `ScaleT1=2^42`, and Standard's exact integer ratio `k=floor(2^84/q26)=4398038122510`, the reviewed bounds are:
   - `B_T2_pre <= 64649255451857143382418104254464`
   - `B_T2_post <= 14699510789662016947`
   - `B_T3_preSub <= 42789729724024999377720936179651527668`

   Each is below the strict Q0123 centered capacity `(Q0123-1)/2 = 174123457892293883773376914685791456788480`, where `Q0123=348246915784587767546753829371582913576961`.

The derivation is accepted as a property of that evidenced path. It must not be used to authorize arbitrary scale-alignment multiplications without an applicable, trusted bound. If the production API has no safe generic bound/provenance for a broader input, reject that case or preserve a previously valid full-Q route; do not infer historical no-wrap from centered residues, hard-code these seed-specific bounds, read dormant rows, reconstruct q4+, or silently fall back. If the T3 implementation cannot remain within the existing representation/capacity contract, stop with `NEEDS_WEB_REVIEW`.

## Source and semantic contract

Before editing, inspect the current Fast source at the pinned Secondary base and compare it with genuine Standard's `schemes/ckks/evaluator.go` `evaluateInPlace` behavior. Verify scale ratio truncation, output Scale, Level, Degree, aliasing, and operand mutation rather than relying on this prose alone.

The target supported case is the reviewed T3 `Sub` shape: same Level 25, left/higher-Scale Degree 2, right/lower-Scale Degree 1, ordinary public `ckks.Evaluator.Sub`, with `opOut == op0`. Compute `k=floor(ScaleA/ScaleB)` from actual metadata; multiply the lower-scale operand's actual Q-prefix coefficients using existing `fastcore.ScalarWorkspace.MulRows`; update the temporary Scale metadata only after that coefficient multiplication; then call existing `fastcore.AddSubWorkspace.ApplyRows` for subtraction. Preserve output Scale as Standard does, Level, Degree, NTT/Montgomery, batching, Q0123 backing, and alias semantics. Do not hard-code `k`.

Share the minimum helper with ordinary `Add` only if the same source-proven semantics and safety gate apply without scope expansion. Equal-Scale Add/Sub must keep their existing dispatch. Unsupported alias directions, left-lower-scale alignment, unequal Levels, incompatible Degrees/metadata, nonzero c1/c2, invalid/noncanonical backing, or unproven capacity must fail explicitly before mutating inputs/outputs; do not claim general Standard equivalence for unsupported cases. Never read or materialize q4 or higher dormant rows.

No new multiplication, Rescale, CRT, NTT, capacity-tracking, measurement, runner, or bootstrap algorithm is authorized. No production/frontend change in Primary and no BTS source/config/input change is allowed.

## Implementation plan and acceptance tasks

- [ ] Inspect Fast evaluator dispatch, `evaluator_fast_zero_secret_add.go`, `evaluator_fast_zero_secret_scalar.go`, `fastcore/addsub.go`, `fastcore/scalar.go`, capacity guards, and the pinned genuine Standard `evaluateInPlace`; record the verified reusable call chain and any narrower cases.
- [ ] Implement only the safe equal-Level, higher-Scale-left ordinary Add/Sub alignment case described above, reusing the existing scalar and Add/Sub kernels. If the production capacity precondition cannot be represented safely, stop before broadening semantics.
- [ ] Extend existing CKKS Fast API tests with a local T1→T2→T3-shaped, no-Bootstrap regression: Level 25; real distinct scales; Sub left Degree 2/right Degree 1; `k` computed from metadata; both `opOut==op0` and independent output; per-row modular oracle for q0–q3; output metadata/domain; dormant rows absent; Fast-kernel dispatch proof separate from numeric oracle; and post-T3 local arithmetic only if needed.
- [ ] Cover invalid/noncanonical prefix backing and nonzero c1/c2 as negative cases with no partial mutation. Cover equal-Scale Add/Sub regression. Explicitly assert unsupported scale direction, alias, Level, or Degree combinations reject without mutation.
- [ ] Run focused tests, relevant Fast/core package tests that do not invoke any Bootstrap, affected-package `go vet`, and `git diff --check`. Do not run a package/test command until checking that it cannot invoke Bootstrap. If a broader CKKS suite is run and fails, compare exact failures against pristine `d1b6dc9963b0df95a7c149306fb62660e1020f7b`; do not label an un-compared failure pre-existing.
- [ ] Self-review the full Secondary diff and verify four distinct gates: Standard modular coefficient semantics, independent numeric oracle, actual Fast-kernel dispatch, and the fixed-path capacity premise. A pass at one gate does not imply the others.
- [ ] If all gates pass, commit and normal fast-forward push the Secondary repair to `origin/fast-qprefix`. Preserve the original five-file dirty Secondary worktree exactly.
- [ ] Record concise results/evidence and update this journal and Primary summary. Run Primary `git diff --check`; commit and normal fast-forward push only authorized Batch 005 artifacts to `origin/main`.

## Commands and test policy

Run from the clean isolated Secondary checkout, selecting focused existing package tests after inspecting their call graphs. Likely targets include `go test ./schemes/ckks -run '<focused test names>' -count=1`, `go test ./schemes/ckks/fast -count=1`, and `go test ./schemes/ckks/internal/fastcore -count=1`; run `go vet` for affected packages. These are examples to narrow after source review, not permission to run an uninspected Bootstrap-capable test suite. No new test platform or measurement tool is allowed.

Forbidden for this Batch: any `bts.Bootstrap`, any CKKS Bootstrap, Standard BTS, S2C/C2S execution, complete EvalMod, Trace, benchmark, warmup, or change to BTS files/parameters. The durable task journal starts at 0 calls and must remain at 0.

## Boundaries and STOP conditions

- Always preserve the five original dirty Secondary files and all old Task 002/003/004/capacity evidence; use a separately created clean checkout from the verified official GitHub origin.
- Always fail closed on a non-fast-forward remote, unknown changes, failed tests/regressions, dormant-row access, missing zero-secret invariants, or inability to verify dispatch.
- Never reset, stash, discard, overwrite, rebase, force-push, modify BTS, change Standard, spend Bootstrap budget, relax capacity policy, or assert that the accepted captured-path bound applies generically.
- Stop with `NEEDS_WEB_REVIEW` if safely supporting the public API requires a new general representation/capacity rule, new cryptographic semantics, or any contract change beyond this charter.

## Completion status

Batch completion requires the supported scope to be explicit, all mandatory tests/self-review to pass, the Secondary commit pushed normally, Primary evidence/journal/pointer pushed normally, both used clean checkouts verified, and the original Secondary worktree unchanged. Completion reports implementation coverage and numerical evidence separately. No subsequent BTS attempt is authorized by this Batch.
