# Independent Web Review — FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010

**Disposition:** ACCEPT evidence as `FAST_DROPIN_COMPOSITION_BLOCKED`, **reject any claim of completed Standard/Fast composition**. This is a **frontend evaluation-key/parameter mismatch**, not a demonstrated Fast primitive or Standard library defect. Authorize an isolated identical-source frontend repair/test only, not any algorithm rewrite.

## Provenance independently inspected

- Primary remote `main@5a53943f387a968cbc1f906f0179033bd116ec45`; report and evidence `results/FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010-{summary.md,evidence.json}`, and runner `tools/fast-dropin-public-rotate-bootstrap-composition-010/main.go`.
- Secondary `fast-qprefix@00ac70ba136d190fa31bbb26c2f51d003a221634` unchanged. Genuine immutable Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Canonical fixed config SHA-256 `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`, deterministic 4096-slot input SHA-256 `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; LogN13 residual Level0, log2Scale45, E32, unchanged Q/P profile.
- Fast public preflight after correcting harness-only reflection panic passed: Rotated-vs-cleartext RMSE `7.44997342660252e-13`, max complex `2.1019564094557047e-12`, zero c1 and zero GetGaloisKey calls. The preflight was rerun after one diagnostic harness failure contrary to no-retries rule; accurately acknowledged, not concealed.
- Standard entered genuine native Rotate with nonzero c1, but panicked `eval.GadgetProductLazy: ctQP.LevelP()=-1 < gadgetCt.LevelP()=4`; **zero Bootstrap calls on either backend**. There was no Standard rotate oracle, no post-Bootstrap output and no same-run Standard/Fast numerical comparison. The source hash for the panicking Standard run was not serialized; therefore do not call both original runs exact-source-comparable.

## Independent root cause

The 010 runner builds `residual` with only its Q primes and no P, generates `keys` via `btpParams.GenEvaluationKeys(sk)`, then calls:
```go
trackingKeys := &trackingEvaluationKeySet{EvaluationKeySet: keys.MemEvaluationKeySet}
rotateEval := ckks.NewEvaluator(residual, trackingKeys)
rotated, err := rotateEval.RotateNew(ct, 1)
```
But the pinned genuine Standard's `circuits/ckks/bootstrapping/keys.go` explicitly generates Bootstrap Galois/relinearization keys using **`p.BootstrappingParameters`**, including its five P primes. `rlwe.Evaluator.Automorphism` uses a matching Galois Key and `GadgetProduct`, which allocates QP scratch using the evaluator's own parameters; `GadgetProductLazy` validates that its scratch P level covers the key's P level. For a P-free residual evaluator the comparison is `-1 < 4`, hence the panic. Merely seeing galEl=5 in the key list says **nothing** about parameter compatibility. The original 010 spec incorrectly assumed Bootstrap evaluation keys could be reused with a residual evaluator, and Web shares responsibility for this incorrect preflight plan.

## Source-grounded safe composition choice

**Preferred bounded repair: use the already-generated BootstrappingParameters to instantiate the public `ckks.NewEvaluator` for the pre-Bootstrap Rotate, while keeping the existing Level-0 ciphertext produced with the unchanged residual parameters and using the original Bootstrap key set.** This does not change parameter values or invent new keys. Conditions that must be checked *before* executing Rotate:
- `residual.N()==BootstrappingParameters.N()` and both are Standard ring; source `bootstrapping.NewParametersFromLiteral` explicitly prepends the entire residual-Q chain to the Bootstrap-Q chain (verify exact actual q0 and ideally all residual prefix equality in runner).
- In the same-N case `GenEvaluationKeys(skN1)` extends the same secret to the full Bootstrap Q/P basis as `skN2` rather than sampling a different secret. Source `bootstrapping/keys.go` verifies this; document resulting secret compatibility at q0.
- The pre-Bootstrap rotation key is Standard-layout with adequate P basis in the genuine Standard branch; use matching `BootstrappingParameters` evaluator, **not** `residual`. Its ciphertext is still logical Level 0 and retains the same Scale/slots; no automatic ModUp or domain mutation at this stage. The public Fast CKKS zero-secret route remains keyless, with P0 full-active-Q validation.
- The *same exact newly written frontend source* must run against pinned Standard and Fast dependencies. The source intentionally selects the already existing correct parameter object in **both** builds; no backend-specific branches or frontend Fast imports. Parameter hashes/visible settings remain identical between them.

This is a source-supported hypothesis that should be verified by a fresh no-Bootstrap preflight. It is **not proof** that the genuine Standard rotation or follow-on Bootstrap will succeed numerically; for example noise/basis/key/domain behavior needs execution. If preflight fails, stop with exact first failure and zero Bootstrap, and bring evidence back for review. Do not mask a mismatch by manufacturing P primes, changing E, reducing problem size, or manipulating ciphertext rows.

An additional independent residual-compatible Galois key was considered but is **not** the default because it introduces a second key plan and may have different noise/security behavior at Level0 with no P. Do not add it unless independently authorized after evidence; also do not patch Standard Lattigo's native GadgetProduct panic handling in this project.

## Execution controls for 011

Run a new fixed source no-Bootstrap preflight under both dependencies; no production source modifications. If both pass, **at most one** canonical E32 public Bootstrap invocation per backend using its *rotated* ciphertext and the same Bootstrap evaluator/key plan. Record independent cleartext oracle, cross-backend decoded comparisons when both outputs available, Q/P hashes, c1/Level/Scale, and explicit bootstrap counters. Catch unexpected panics at the **runner boundary** to serialize first-error evidence (not to continue or falsely mark success). No benchmark or retries. The previous 010 runner remains immutable as negative evidence.

**Classification:** `FAST_DROPIN_COMPOSITION_BLOCKED` / `NEEDS_WEB_REVIEW`; next approved work `FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011`.
