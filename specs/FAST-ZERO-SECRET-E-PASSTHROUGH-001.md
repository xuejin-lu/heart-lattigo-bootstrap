# FAST-ZERO-SECRET-E-PASSTHROUGH-001

Task class: bounded implementation (Secondary) and one-shot numerical experiment. Status: READY_FOR_CODEX. Long-term objective: `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`. This task does NOT claim full drop-in replacement.

## Frozen scientific contract

1. Safe FF-sync Primary main and Secondary fast-qprefix, read their AGENTS, this spec and authoritative `docs/FAST_QPREFIX_SPEC.md`. If a durable rule conflicts, stop NEEDS_WEB_REVIEW before editing.
2. **Only Secondary Fast-fork internal production code** may change. Do not edit Standard Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, application/frontend source, Primary harness production code, existing configs or historical results. A short factual result report in Primary is authorized.
3. Keep all existing Standard public API signatures and user-passed CKKS parameter values (including E, actual Q/P primes, LogN, LogSlots, Scale and circuit order) unchanged. In the Fast **zero-secret simulation** path, accept E=0 and E=32 and omit the unnecessary Dense/Sparse secret KeySwitch. Do **not** force public-visible E to zero or route Fast keys through native Standard KeySwitch.
4. This is a numerical-validation simulation, NOT normal encryption or a security-equivalent alternative. No secret-key material may appear in reports. No production-mode runtime Fast selector or Q-prefix full-RNS fallback.

## Source-backed investigation and implementation

- `circuits/ckks/bootstrapping/fast_bootstrap.go` currently rejects any `EphemeralSecretWeight != 0`.
- `fast_modup.go` implements Fast ModUp without Dense/Sparse KeySwitch.
- `fast_keys.go` has an E-positive branch for special `(error,0)` Fast-layout keys; generating these does not mean the Fast Bootstrap executes real KeySwitch.
- Verify the relevant actual call paths and Q-prefix invariants, then make the **smallest sound change** so Fast zero-secret Bootstrap accepts E=0 and E=32 with the supplied E preserved. Inspect shared circuit planning and metadata. Removing the guard alone is not proof of correctness. Preserve all unrelated input parameter checks and failure modes.
- Focused inexpensive tests: constructor accepts E0/E32, original E remains visible, no Dense/Sparse KeySwitch is performed by the Fast bootstrap call path, and genuinely unsupported settings still error. Do not broaden into generalized native-encryption compatibility.

## One-shot measurement, not performance benchmark

- Use existing canonical **LogN13** input and CKKS profile. E0 and E32 must each run **exactly one Fast public Bootstrap** with the same original deterministic encoded message, **without modifying Q/P primes, Scale, K or other frontend parameters**. Record decoded output vs original: complex RMSE, maximum complex / real / imaginary error and SNR, output metadata; compare Fast E0 vs Fast E32. No output accuracy cutoff is assumed in advance; show numbers regardless of size.
- Obtain one **genuine unmodified Standard E32** Encrypt -> Bootstrap -> Decrypt/Decode output on the same profile and original message from a demonstrably matched source-backed prior run, or one bounded new run from a separate clean Standard checkout. Compare Fast E32 to Standard E32 and both against original. If exact source/effective primes or workload differ, explicitly mark the comparison NONCOMPARABLE; do not pretend the old Standard E0 result is the E32 reference.
- This authorizes at most 3 public bootstrap invocations on LogN13 in this task (Fast E0, Fast E32, native Standard E32). Do not launch LogN16, repeat trials, calibration, time benchmarks or fallback/reruns to pick favorable results.
- If any step cannot be completed safely, report exactly what failed and preserve the partial measurements. No false success status or silent parameter changes.

## Tests, outputs, classification

- Run targeted Secondary tests, confirm existing Fast regression suite where feasible, and do a bounded code self-review with `git diff --check`. Commit/push Secondary fast-qprefix safely if the implementation passes. Do not change Secondary task pointer as Primary is authoritative.
- Write a compact `results/FAST-ZERO-SECRET-E-PASSTHROUGH-001-summary.md` to Primary, reporting backend SHAs, run input and effective-parameter hashes, constructor/key semantics, E values, 1-run metrics, test commands, discovered limitations and commit/provenance; safe Primary push allowed. Do not serialize raw ciphertext, private keys or full slot arrays.
- `FAST_E_PASSTHROUGH_RUN_COMPLETE`: both E Fast runs and numeric outputs complete, and Standard reference comparable or explicitly pending.
- `FAST_E_PASSTHROUGH_PARTIAL` / `FAST_E_PASSTHROUGH_BLOCKED`: source, invariant, runtime or provenance blocker with exact evidence.
- Always report `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`; no numerical accuracy or performance acceptance solely because an API accepted E=32.

## Deferred full objective (NOT part of this task)

Prove actual transparent substitution using the **same unchanged frontend source** for both library checkouts, including normal GenEvaluationKeys, encrypt/decrypt, NewEvaluator, bootstrap and multi-operation CKKS chains, along with meaningful numeric acceptance gates. Keep this separate from the immediate E-passthrough implementation.
