# FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011

**Status:** READY_FOR_CODEX
**Type:** Bounded E system integration / isolated frontend key-plan correction only.
**Review:** `results/FAST-DROPIN-PUBLIC-ROTATE-BOOTSTRAP-COMPOSITION-010-web-review.md`.
**Fast Secondary pinned:** `00ac70ba136d190fa31bbb26c2f51d003a221634`.
**Genuine Standard pinned:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Canonical config:** `configs/bootstrap_config.logN13.json`, SHA-256 `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; fixed input SHA `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.

## Root cause and authorized fix

Task 010 incorrectly supplied `bootstrapping.Parameters.GenEvaluationKeys(sk)`'s full-Q/P Bootstrap GaloisKey (P Level 4) to `ckks.NewEvaluator(residual,...)`, whose parameters have P Level -1. Native Standard `GadgetProductLazy` panicked before any Bootstrap call. Fix the **frontend pairing**, not Standard/Fast algorithm code.

**Choose in the new identical frontend:**
```go
keys, _, err := btpParams.GenEvaluationKeys(sk) // original Bootstrap key plan
rotateEval := ckks.NewEvaluator(btpParams.BootstrappingParameters, keys.MemEvaluationKeySet)
rotated, err := rotateEval.RotateNew(ctAtLevel0UnderResidualQPrefix, 1)
```
The same source compiles and runs against both dependencies. The evaluator's Q/P matches the key-generation basis; the ciphertext is still at **logical Level 0** and was encrypted using residual CKKS parameters with the identical actual q0. This is not a change to user-facing CKKS parameters, config, Scale, E=32, slot count or frontend API names, and needs no new Galois-key plan. Zero-secret Fast public Rotate remains keyless, while genuine Standard takes native key switching.

## Mandatory compatibility gates before Rotate

- Check `residual.N()==btpParams.BootstrappingParameters.N()`, both RingType Standard, `residual.Q()` is an **exact ordered prime prefix** of `btpParams.BootstrappingParameters.Q()`, and the two actual q0 primes are identical. Reject mismatch before any operation.
- Check the original Bootstrap Galois key for `residual.GaloisElement(1)` exists and its native Standard layout (under genuine Standard) has P Level compatible with evaluator's `MaxLevelP`; Fast keys may be Fast layout, but no GetGaloisKey should occur on Fast Rotate.
- Source-verify same-N Bootstrap key generation **extends** the same secret across full Q/P (not unrelated regenerated secret). Report precise source path and key parameters. Validate the resulting Rotate ciphertext still decrypts with the **original residual** secret key; don't silently replace decryption keys.
- The original Level0 encrypted input remains NTT/non-Montgomery, degree one, complete q0 backing, Scale=2^45; preserve LogSlots and layout across Rotate. Do not raise ciphertext Level or allocate fake extra active rows.

## Execution order

1. Safe-sync both Primary/Secondary under `AGENTS.md`; require clean and exact Secondary pinned source. Read 010 negative evidence, 010 Web review, 004 canonical successful bootstrap and original 010 runner. Preserve 010 source/evidence unchanged.
2. Create one **new Primary** runner `tools/fast-dropin-rotate-bootstrap-keyplan-011/main.go` with identical frontend source for both dependency builds. Minimize duplication where feasible, but do not modify historical 004/010 sources merely to share helpers.
3. Make both runners' P-level/secret/prime/config/input assertions **before** the expensive operation. Improve diagnostics so any unexpected panic from Standard (or Fast) is recovered **only at the runner boundary** to serialize `status=preflight_failed`, exact first error, no phantom Bootstrap. Never continue after recovering a panic.
4. Under pinned Standard and Fast, execute **one preflight per backend** with KeyGen → Encode/EncryptNew → public `ckks.NewEvaluator(btpParams.BootstrappingParameters, keys.MemEvaluationKeySet).RotateNew` → DecryptNew/Decode. Compare against correctly rotated complex plaintext oracle and record RMSE/max/SNR, Level/Scale/degree/NTT/Montgomery/Q row backing, c1 count, key access, and exact profile/prime/input/frontend hashes. Preflight must succeed on BOTH to proceed.
5. If and only if both pass, execute **at most one** public Bootstrap per backend on each pre-rotated ciphertext, using original public `bootstrapping.NewEvaluator(btpParams, keys)` and same canonical E32 settings. Record Bootstrap output oracle vs rotated plaintext, exact metadata/c1, finite RMSE/max/SNR, and direct decoded Fast-vs-Standard error if both arrays available. No tuning, retries or warmups. If one Bootstrap errors, stop and preserve evidence; do not attempt a compensating rerun.
6. Execution source hash must be captured **before** run in each backend including failures/panics. Prefer write compact JSON `results/FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011-evidence.json` under Primary, and preserve full decoded output arrays only if appropriate (no keys/secret values). Show exact Standard/Fast same source SHA, config SHA, input hash, Q/P SHA and backend heads.
7. Tests and limits: `go test`/compile and `go vet` new Primary runner under both workspaces; small focused auxiliary unit tests for param-key consistency are optional. **No changes to either Lattigo library**, no CKKS primitive rewrites, no production changes, no Bootstrap test suite, no benchmark, no LogN16, no CNN. Self-review one bounded pass; no unauthorized repeated bootstrap operations.

## Deliverables

- New shared 011 runner, `results/FAST-DROPIN-ROTATE-BOOTSTRAP-KEYPLAN-011-summary.md` and compact evidence JSON, committed/pushed only in Primary, Secondary unchanged.
- `FAST_DROPIN_KEYPLAN_COMPOSITION_COMPLETE_PENDING_WEB_REVIEW` / `FAST_DROPIN_KEYPLAN_COMPOSITION_PARTIAL` / `FAST_DROPIN_KEYPLAN_COMPOSITION_BLOCKED` plus `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.
- The performance baseline `FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED. Do not claim full CKKS drop-in, compact-Q public acceleration, or cryptographic security from this limited chain.
