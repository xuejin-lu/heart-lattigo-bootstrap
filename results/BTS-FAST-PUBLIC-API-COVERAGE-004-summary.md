# BTS Fast public API coverage — Batch 004

Status: `BATCH_IN_PROGRESS` (A–C passed; D pending). This report covers only the unchanged BTS LogN16 / `SlotLimbs=15` / `BTSModePlain` source path and the pushed Fast backend. It does not claim end-to-end success or Q0123-only execution.

## A. Ordinary API coverage and actual backend lanes

| BTS path / ordinary entry point | Fast dispatch and reused work | Representation / dormant-row boundary | Provenance or coverage |
|---|---|---|---|
| `NewBTS`, input creation: ordinary `rlwe.NewKeyGenerator`, `ckks.NewEvaluator`, `rlwe.NewEncryptor`, `EncryptNew` | Fast library chooses its mode-specific `a=0` evaluation-key/zero-secret simulation internally; BTS source makes no backend branch. Encryption uses the Fast zero-secret path, where `c1=0` is an explicit invariant. | Ordinary CKKS ciphertext metadata and public interfaces. Zero-secret behavior is intentionally insecure and is not a Standard security claim. | `external/BTS/main.go`, `bts.go`; same BTS source built against the pinned backend.
| S2C BSGS: `DecomposeNTT`, `AutomorphismHoistedLazy`, `GadgetProductLazy`, `ModDownQPtoQNTT`; direct `ringQP` pointwise accumulation / automorphism | Zero-input decomposition/key-switch portions are skipped only at validated zero-component boundaries. Required Galois index permutation and BSGS arithmetic remain. Existing Fast compatibility wrappers and ordinary Fast arithmetic are reused; direct QP arithmetic remains BTS-owned. | This is an explicit full-Q/QP lane: lazy automorphism results are P-scaled QP; BSGS accumulators and ModDown require complete Q/P rows. Dormant Q rows are admitted only after complete-Q representation validation; no unverified residues are promoted. | `external/BTS/mys2c.go`; existing QP compatibility and zero-secret automorphism tests.
| ScaleDown: ordinary `Mul(ct, integer, ct)`, then `RescaleTo` | Integer scalar multiply dispatches through the shared Fast Q-prefix scalar core; `RescaleTo` uses the existing Fast Rescale core. No second multiplication or Rescale algorithm was added. | Compact active Q-prefix rows; Level decreases only when a real rescale is required; Scale follows the public API target. | Shared-kernel tests and Fast Rescale tests.
| ModUp: `ApplyEvaluationKey`, direct coefficient/NTT/Q extension, `Trace` | `ApplyEvaluationKey` checks the required `c1=0` condition and elides only zero-operand key-switch arithmetic. `Trace` is ordinary public API; at LogN16/LogSlots15 its gap is 1 and the existing full-slot path is a copy/no-op. | ModUp explicitly constructs the complete-Q representation it needs; it is not a Q0123-only stage. Fast deferred errors are surfaced at the next error-returning boundary. | `external/BTS/modup.go`; existing ApplyEvaluationKey/Trace compatibility tests.
| C2S: `DecomposeNTT`, `AutomorphismHoistedLazy`, `Rescale`, `Conjugate`, `Sub`, `Mul(ctImag,-1i,ctImag)`, `Add` | Reuses existing Fast hoisted/Rescale/automorphism cores. The repaired ordinary complex-scalar `Mul(-1i)` now reaches the same import-neutral `fastcore` kernel as the explicit Fast evaluator; ordinary scalar/ct `Add` and `Sub` use the existing prefix arithmetic. Conjugation preserves the required automorphism; zero key-switch work is conditional on validated `c1=0`. | Q-prefix arithmetic where the input is compact; verified full-Q/QP compatibility is retained for BTS lazy outputs. No arbitrary dormant row reads. Scale/Level and NTT/Montgomery metadata are retained by the shared path. | `external/BTS/myc2s.go`; `TestPublicFastBTSLogN16...`, scalar oracle/dispatch/in-place tests, existing automorphism/QP tests.
| EvalMod PS / DoubleAngle: scalar and ct `Add/Sub`, scalar/int/ct `Mul`, `MulThenAdd`, `MulNew`, `MulRelin[New]`, `Relinearize`, `Rescale` | Ordinary API dispatch uses shared Fast scalar and multiplication kernels; key-switch/relinearization elision occurs only after zero `c1`/`c2` checks. Scalar `MulThenAdd` is covered by the shared core. Fast Rescale is reused. | Compact Q-prefix for supported Fast arithmetic; full-Q compatibility is fail-closed where API operands require complete backing. No Standard/full-RNS fallback was introduced. | `external/BTS/evalmod.go`; public BTS-shaped API-chain, Fast/core, and bootstrapping interface tests.
| Final `Step6_Relabel`, output/decode | Metadata relabel only; no additional cryptographic kernel. Decrypt/decode remains the original BTS operation. | Output correctness must be established from a completed ciphertext and the original `want` vector; unavailable until D succeeds. | `external/BTS/bts.go`, `main.go`.

`BTSModePlain` does not execute the separate ReLU helper (`my_relu.go`) or its polynomial/cosine path. `GadgetProductHoisted` is only in the optional encapsulated ModUp branch; the original configuration has `EphemeralSecretWeight=32`, so that branch is active and remains a full-Q/QP operation. BTS direct ring arithmetic is not silently relabeled as the Fast Q-prefix core.

### What the repair changed

Secondary commit [`dcb46a0028af56e6caaa6f3daf8c3b49b53349bb`](https://github.com/xuejin-lu/lattigo/commit/dcb46a0028af56e6caaa6f3daf8c3b49b53349bb) extracts the existing Fast complex/scalar arithmetic into import-neutral `schemes/ckks/internal/fastcore` and routes the ordinary evaluator through it. It fixes the Task 003 `Mul(ctImag,-1i,ctImag)` full-Q dispatch and covers the same-class scalar `Add/Sub` and scalar `MulThenAdd` calls present in the BTS path. It does not modify BTS, Standard arithmetic, Q-prefix mathematics, key semantics, or Rescale.

## C. Non-Bootstrap verification

Passed on the final Secondary code:

- `go test ./schemes/ckks -run 'TestPublicFastBTSLogN16' -count=1`
- `go test ./schemes/ckks -run 'TestCKKS/Evaluator/(Add/Scalar|Sub/Scalar|Mul/Ct/Scalar|MulThenAdd/Scalar|MulThenAdd/Pt|MulThenAdd/Vector)' -count=1`
- `go test ./schemes/ckks/fast ./schemes/ckks/internal/fastcore -count=1`
- `go test ./circuits/ckks/bootstrapping -run 'TestFastBootstrapCircuitInitializationAndInterface|TestFastE32GenEvaluationKeysAndPublicEvaluatorDispatch|TestNewEvaluatorFastDispatchDoesNotRequireDenseSparseKeys|TestFastEvalMod' -count=1` (no Bootstrap)
- `go vet ./schemes/ckks ./schemes/ckks/fast ./schemes/ckks/internal/fastcore ./circuits/ckks/bootstrapping`
- `git diff --check`

The complete `go test ./schemes/ckks -count=1` is not green. Failures were compared to pristine Batch-start commit `049ffa782729d049e08f973b5852ee1e832a55f3` and Task 003's pristine baseline `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`; they match documented historical failure classes: dormant-row decryptor panics in `RescaleTo/Single`, `RescaleTo/Many` precision/capacity cases, and the `Mul/Ct/Ct/Degree0` error-text mismatch. Task-specific scalar/API gates pass; no new failure attributable to this Batch was found. The original failing `MulThenAdd/Pt` nested full-Q route was corrected to call a private exact full-Q scalar routine, and its focused regression now passes.

## D. Bounded No-Trace run and remaining limitations

No Batch 004 Bootstrap has been launched: **0/2**. A–C and the Secondary push are complete. A Fast executable may be built without consuming an attempt, but the one No-Trace invocation must not start until the exact source manifest, backend commit, modfile replacement, build/run commands, executable SHA-256, and unique raw log path are durably reserved in the journal and that reservation is pushed to Primary `origin/main`.

The existing Standard result is historical and will not be rerun: Bootstrap 139.653 s, max absolute error `1.708e-3`, mean absolute error `1.172e-4`. Do not compute a speedup unless this Batch produces a complete Fast output with comparable metrics.

If D completes, report initialization and Bootstrap separately; output Level/Scale, max/mean absolute error and slot; backend/binary/source provenance; and the exact failure or success classification. Retained full-Q/QP hot paths include S2C's QP accumulations and ModDown, required rotations/lazy automorphism representation, and ModUp's complete-Q coefficient lift / optional encapsulation GadgetProduct. Therefore zero-key switching elimination is not evidence that the whole BTS is Q0123-only.
