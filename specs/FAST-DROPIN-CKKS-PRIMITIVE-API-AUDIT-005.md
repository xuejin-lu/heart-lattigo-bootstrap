# FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005

**Status:** READY_FOR_CODEX. **Task class:** bounded source/API/numerical preflight audit (M/E), no production modification.

## Authority

Previous `FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004` is **accepted solely for its tested, single-bootstrap, secret-key input lifecycle**. Web assessment: `results/FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004-web-review.md`. Same frontend E32/LogN13 yielded Fast-vs-original complex RMSE 1.1907546782784477e-9, Standard-vs-original 4.697396790468602e-9 and Fast-vs-Standard 4.821647086585007e-9. Do not broaden this into arbitrary CKKS API compatibility, cryptographic equivalence, or measured speedup.

Long-term user requirement: `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`: same frontend CKKS code/API/parameters, switch only Lattigo library checkout, Fast runs zero-secret numerical validation. Fast is broader than Q-prefix.

## Goal

Determine the **first concrete API/semantic incompatibility in ordinary CKKS arithmetic**, before authorizing another implementation. Audit `ckks.NewEvaluator(params, evk)` against `schemes/ckks/fast.NewEvaluator`, with representative `Add`, `MulRelin`, `Rescale` and `Rotate` or `RotateNew`. Inspect sources before choosing exact call sequence and key set. Specifically establish how native `ckks.NewEvaluator` constructs `rlwe.Evaluator`, how Fast-layout evaluation keys are created or accessed, and whether/where `KeyLayoutFast` is rejected by native KeySwitch. A key failure must be recorded, not bypassed.

## Scope

1. Safe sync Primary main and Secondary fast-qprefix per updated AGENTS; inspect source and pinned genuine Standard dependency `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Fast current baseline is `eb7793f1e449702590cdba3d9603261623354f80`. Do not mutate Standard or Secondary production.
2. Create **one same-source frontend program** in Primary under `tools/` with exclusively ordinary public CKKS APIs; compile the same program against both dependency versions by external build-workspace selection only. No Fast imports, Fast-specific API, manual direct-input building, clearing c1, or different frontend parameters.
3. Use a small fixed, source-verifiable CKKS input/parameter profile adequate for all selected operations, preferably existing canonical settings unless a smaller **identically fixed** profile is safer for inexpensive testing. Treat this as an *independent primitive preflight*, not a direct performance/numerical comparison to the previously frozen LogN13 Bootstrap profile if changed. Document exact Q/P primes, scale, levels, slots and source hashes. Generate independent encrypted operands by the public API for each operation.
4. Perform static source mapping before any execution; avoid invoking unsupported native KeySwitch on Fast-layout keys. Compare `Add`, `MulRelin`, `Rescale` and `Rotate` sequentially **only as safe**, with immediate decoded plaintext oracle per operation under respective backend, check Level/Scale/c1 semantics and exactly which public constructor was selected. On first KeySwitch or representation-invariant uncertainty, stop that operation and report specific gap; do not force full test run at cost of corruption. No Bootstrap calls, GPU tests, high-cost loop, benchmarking or tuning. No automatic fallback to Standard full-RNS on compact stale Fast rows.
5. Public API and parameter equality are necessary but not sufficient. Do not claim numerical equivalence unless both backend results are valid and independently compared to a cleartext arithmetic reference; note CKKS rescale precision and rotations require correct slot-level reference. Keep zero-secret interpretability explicit; avoid incorrect claims that native nonzero-c1 output is valid Fast.
6. Test only relevant no-Bootstrap packages, keep proof compact. Report first mismatched operation with source-level call tree, observed outcome or preflight reason to stop, algebraic reason, and the **smallest justified next internal library implementation**, not an unsupported spontaneous compatibility patch. If only limited subset (e.g., Add) passes, record partial progress honestly.

## Deliverables

- Primary `results/FAST-DROPIN-CKKS-PRIMITIVE-API-AUDIT-005-summary.md` with source locations, identical frontend SHA, build/runtime commands, input/provenance, stage-by-stage results, and isolated recommended next repair. Optional new frontend under `tools/`. No Secondary source or production modifications.
- Run `git diff --check` and self-review; safe FF commit/push Primary only when allowed. Do not repeat Standard E0 experiments, perform Bootstrap, or unfreeze the blocked `FAST-STANDARD-PERF-REBASELINE-003` benchmark.
- Return `FAST_DROPIN_PRIMITIVE_AUDITED`, `FAST_DROPIN_PRIMITIVE_PARTIAL` or `FAST_DROPIN_PRIMITIVE_BLOCKED`, plus `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`. Web will decide the next primitive-specific implementation.
