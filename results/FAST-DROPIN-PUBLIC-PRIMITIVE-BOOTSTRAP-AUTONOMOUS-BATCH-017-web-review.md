# Independent Web Review — FAST-DROPIN-PUBLIC-PRIMITIVE-BOOTSTRAP-AUTONOMOUS-BATCH-017

**Date:** 2026-10-10 (Asia/Taipei).
**Verdict:** **ACCEPT BOUNDED LOW-RESIDUAL-LEVEL PUBLIC PRIMITIVE-CHAIN → E32 BOOTSTRAP MILESTONE**, not universal drop-in, high-Level compact-to-Bootstrap, measured speedup or cryptographic security equivalence.
**Primary runner commit:** `f3148a13735f8ba3a4eb9e7dc6a952cf4c1302bd`.
**Primary submitted evidence commit / remote main observed:** `637df5999ab422c8391cd82b6a12232854709b9b`.
**Fast Secondary remote observed:** `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, unchanged.
**Untouched genuine Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, developer-reported clean independent checkout.

## Scope and independent checks

Inspected GitHub Primary 017 committed summary, journal, compact combined evidence, runner `tools/fast-dropin-public-primitives-bootstrap-batch-017/main.go`, its commit and existing Secondary public dispatch, Fast key-plan and public Bootstrap evaluator source. Both GitHub branch HEAD values and Primary changed-file set were verified. The public frontend is single source for both module checkouts; `compareRunFiles` checks pinned commits, non-dirty profiles, shared frontend/config/input/QP hashes, eight finite nonempty decoded checkpoints and exact one Bootstrap call per backend. `stripDecoded` copies the checkpoint slice before nil-ing decoded vectors: the old 015 shared-slice false-zero bug is **not present** in this 017 combiner. Web has not independently run the Go tests, expensive Bootstraps or retrieved developer-local raw 4096-slot arrays; local workspace cleanliness is developer-reported.

## Accepted result

- Canonical LogN13, 4096 slots, E32, original Q/P and public `btpParams.GenEvaluationKeys(sk)`, `ckks.NewEvaluator`, `bootstrapping.NewEvaluator`, `Bootstrap`, native `DecryptNew/Decode` all execute in ordinary public source. Input A/B Scale=2^45, C's legal per-ciphertext Scale is **exact actual residual q1** (≈2^39), in **both** backends; the Q/P/config/default Scale/E parameter settings did not change.
- Arithmetic chain `AddNew(A,B)→MulRelinNew(sum,C)→Rescale(q1)→RotateNew(1)` passes independent plaintext, Level/Scale and native decryption checks and becomes the **actual Bootstrap input**. MulRelin output Level1 Scale=2^45·q1; Rescale yields Level0 Scale=2^45.
- Genuine Standard used nonzero c1 and actual relin/Galois key accesses. Fast zero-secret c1 remained zero, shared Fast primitive core was source-selected without primitive relin/Galois lookups, and public Bootstrap selected existing Fast implementation with E32 intact. Fast evaluator creation once enumerated Galois-key list. Bootstrap **internal per-key access counts were not instrumented**; its avoidance of normal KeySwitch is source-backed, not independently runtime-counted.
- Bootstrap output for each backend: Degree1 Level1 Scale=2^45, native `DecryptNew/Decode`. Standard RMSE/max = `4.908811884776509e-9 / 1.4187430657863595e-8`; Fast = `1.1790593543156793e-9 / 5.7394238676200815e-8`; direct decoded Fast-vs-Standard RMSE/max = `4.997862453389243e-9 / 4.326414527528976e-8`. Eight paired input/primitive/Bootstrap gates passed the preset `1e-6` max error criterion and matching Level/Scale, with one public Bootstrap invoked per backend.
- The *first* noninteractive preflight pair ended at an explicit stdin authorization gate with EOF **before** any Bootstrap. Codex repeated only cheap preflight in held interactive processes; those processes supplied the two actual Bootstrap outputs. No expensive retry. The source clearly records attempt/call before Bootstrap; combined evidence records two total. An intentional wait gate is an orchestration detail, not a CKKS algorithm failure.
- Focused `go test`, `go vet` in pinned Standard and Fast workspaces and formatting/diff checks are reported passed, but not independently rerun by Web.

## Scientific limitations and next gate

At residual Levels 1→0 the maintained Q-prefix `w_Q(ell)=min(ell+1,4)` equals the **full logical active Q width**. Thus this is convincing real Public API/Bootstrap integration but **does not demonstrate execution of a genuinely truncated high-Level q4/q5 ciphertext immediately before the Bootstrap pipeline**. Batch 016 separately proved bounded logical-Level5 compact computation using a measurement-only decoder. The two proofs must not be described as already composing into one high-Level full public workflow.

Outstanding: (1) source-backed and mathematically valid high-Level compact `Level>3` arithmetic **to actual canonical residual Level0 Bootstrap input** under the existing E32 public profile; (2) arbitrary high-Level compact native `DecryptNew` boundary, for which original RLWE decryptor still expects full active Q rows; (3) unsupported scalar/plaintext/mixed operand overloads; (4) measured acceleration, no security-equivalence claims.

**Decision:** accept Batch017; authorize bounded **Batch018** focusing high-Level compact→Level0 bridge via ordinary public Level/Scale operations and first preflight; only after proven and pinned evidence may a maximum of one Standard and one Fast Bootstrap be attempted. Do **not** silently resize/DropLevel without a proven capacity condition `2B<q0` at the target Level0 or modify original E/Q/P/default profile. Treat a new mathematical condition or unavoidable public API incompatibility as immediate Web stop, not as license for arbitrary algorithm design. A High-Level-native-DecryptNew adaptation is a separate workstream and is NOT implicitly solved in 018.
