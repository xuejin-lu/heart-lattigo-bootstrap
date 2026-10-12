# BTS-FAST-E2E-FIRST-006 — Full Bootstrap First, Then Numerical Comparison

Status: `BATCH_IN_PROGRESS`
Class: bounded backend implementation and experiment (I/E)
Supersedes: only the unresolved Batch 005 capacity-certificate stop for the authorized Fast zero-secret experimental mode. Batch 005 status, commits, results, and evidence remain immutable historical records.

## Goal and approved research-policy change

Run the unchanged original BTS LogN16 workload through the intentionally insecure Fast zero-secret backend, then compare its decoded output with genuine original Standard output on the same plaintext input. Prioritize an actual end-to-end result over a generic Q-prefix capacity certificate.

The user explicitly authorizes continuing Fast zero-secret API repair, local tests, and bounded end-to-end execution without first obtaining a trusted centered-coefficient-bound certificate. A missing certificate is reported as `CAPACITY_UNPROVEN_EXPERIMENTAL`, with numerical-risk caveat; it is not a stop condition for this mode. This authorization applies only to the experimental Fast mode whose secret and evaluation-key `a` are zero and whose non-`c0` ciphertext components are checked zero at the relevant API boundary. It does not apply to genuine Standard, secure CKKS, nonzero-secret paths, or any security/equivalence claim.

This is not a claim that an unproven prefix lift matches the full logical-Q result. Preserve the fixed Q-prefix policy, exact row-wise modular operations, logical Level/Scale and domain semantics; never read dormant rows or silently fall back to Standard full-RNS. Record the capacity status and risk explicitly. No captured-seed bound may be hard-coded as a generic API precondition or guarantee.

## Repositories and provenance

- Primary: `xuejin-lu/heart-lattigo-bootstrap`, branch `main`; baseline at task creation is the synchronized Primary HEAD recorded in the initial journal.
- Fast Secondary: `xuejin-lu/lattigo`, branch `fast-qprefix`; use an independent clean checkout created from the freshly fetched official `origin/fast-qprefix`. Preserve the original local Secondary worktree and its five pre-existing dirty paths exactly; do not merge, reset, stash, checkout, or implement there.
- Genuine Standard: unmodified Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, built from a separate clean checkout.
- BTS source: `external/BTS/`; do not modify, stage, commit, or push any original BTS source or its `go.mod`/`go.sum`.
- Use isolated Go `-modfile` replacements and `GOWORK=off` to compile Standard and Fast binaries from identical BTS source and measurement adapter. Verify module replacement, source/ref cleanliness, Go build info, and binary SHA before each call.

## Fast public `Sub` repair

Implement ordinary `ckks.Evaluator.Sub` for the BTS T3 shape by reusing the already-existing Fast scalar and Add/Sub kernels:

1. Compute `k=floor(ScaleHigh/ScaleLow)` from the actual Scale metadata; do not hard-code the captured value.
2. Multiply the lower-Scale ciphertext coefficients by `k` using existing `fastcore.ScalarWorkspace.MulRows` into a compact temporary.
3. Only after the coefficient multiplication succeeds, align the temporary Scale metadata to the higher Scale.
4. Complete subtraction using existing `fastcore.AddSubWorkspace.ApplyRows`.

Priority shape: Level 25, high-Scale left operand Degree 2, low-Scale right operand Degree 1, `opOut==op0`, Q0123 compact backing, NTT domain, zero `c1`/`c2`. Preserve ordinary public API use, output Level/Scale/Degree/domain/batching and alias behavior. Validate all relevant structure and zero-secret conditions before modifying output. Do not read or materialize q4+ dormant rows. Do not implement a new scalar, Add/Sub, Rescale, CRT, NTT, capacity tracker, or BTS frontend algorithm.

Unsupported Level/scale direction/alias/degree/domain/backing shapes must reject before partial mutation unless existing source proves an already-supported full-Q route. Existing equal-Scale public Add/Sub behavior must remain intact.

## Zero-Bootstrap validation gates

Add focused tests to the existing Secondary CKKS Fast API test suite (no new test platform). Cover:

- T1→T2→T3-shaped local arithmetic only; no BTS stage, full EvalMod, or Bootstrap.
- Level 25, real unequal Scales, left Degree 2/right Degree 1, in-place `opOut==op0`; include independent output if the same path supports it without widening scope.
- Runtime Standard integer ratio semantics, row-by-row modular oracle over q0..q3, output Scale/Level/Degree, NTT/Montgomery and batching.
- Actual dispatch through both existing Fast scalar and Add/Sub kernels, independently of the numeric oracle.
- Only Q0123 rows are read/allocated; q4+ remains absent/dormant.
- Negative c1/c2, malformed/missing authoritative backing, unsupported alias/direction/Level/metadata cases return errors without mutating input or preexisting output.
- Equal-Scale Add/Sub regression.

Clearly distinguish Standard modular coefficient semantics, independent oracle, observed Fast-kernel dispatch, and capacity status. Capacity status for an unproven experimental lift is `CAPACITY_UNPROVEN_EXPERIMENTAL`; it must not be represented as a proof or numerical pass.

Run focused CKKS tests, relevant Fast/core tests and affected-package `go vet`, after inspecting their call graphs to ensure they do not invoke Bootstrap. Run `git diff --check`. If broader suites are needed, compare failures against pristine task-base source before classifying historical failures.

## BTS run contract

Keep `external/BTS/` Go source, formal module files, parameters, workload, and operation order unchanged. The fixed workload is LogN16, `SlotLimbs=15`, `BTSModePlain`, `amp=1`, `seed=0`, 32,768 slots. Use ordinary public CKKS APIs; Fast zero-secret behavior is internal to the linked Fast backend.

First compile both backends independently from the same original source. A minimal temporary measurement adapter/Go overlay is allowed only if needed to preserve the existing plaintext and decoded output vectors and their hashes; it must not change the CKKS operation sequence, parameters, input, or backend behavior. Keep original-source hashes and the adapter delta in the manifest. Prefer reusing `cmd/perfprobe`, `internal/perfmeasure`, `internal/numericalmetrics`, and supported `cmd/fastdiag` utilities; do not create a duplicate runner or aggregate-metrics implementation.

### Durable call budgets

- Fast: at most **3** complete `bts.Bootstrap()` invocations for this Batch across all sessions.
- Genuine Standard: at most **1** new BTS invocation, only if preflight confirms that no reusable historical artifact contains the full decoded Standard vector for the same original 32,768-slot input and compatible source/parameter provenance.
- No warmup, trace, repeated timing campaign, extra Bootstrap, Standard rerun when an adequate vector exists, or reuse of any historical Batch 002/003/004/005 reservation.
- Before each authorized invocation, persist a new exclusive reservation in the Batch 006 journal with backend commit/ref, binary SHA-256, BTS source manifest SHA-256, command, timestamp, and attempt index. Record entry into and return/failure of `bts.Bootstrap`, exact failure position, output provenance, and budget accounting. Never re-run an identical binary/input after failure. A repaired backend requires focused tests, a new committed backend SHA, a new binary SHA, and its own reservation before another Fast attempt.
- Stop after three Fast invocations. Do not reset or reinterpret counters on session resume.

Run Fast first. Once it completes successfully, inspect prior Standard artifacts for a full valid decoded vector. If absent, use the single newly authorized genuine Standard run. Standard must link to the exact pristine Standard commit, not a Standard-named API in the Fast binary. No Standard source modifications.

## Numerical comparison

For aligned same-type decoded complex vectors `S`, `F`, and original plaintext `X`, require exactly 32,768 finite slots and valid/expected output metadata before comparing. Save SHA-256 of canonical input, Standard vector, Fast vector, source manifest, and binaries. Reuse `internal/numericalmetrics` (and relevant existing comparison helpers) for:

- `MaxAbsDiff = max |F[i]-S[i]|`
- `RMSE = sqrt(mean(|F[i]-S[i]|^2))`
- `MeanAbsDiff = mean(|F[i]-S[i]|)`
- `StandardMaxError = max |S[i]-X[i]|`
- `FastMaxError = max |F[i]-X[i]|`
- when supported by the shared metrics, separately named real, imaginary and complex-norm comparisons; never mix metric definitions.

For every nonzero reported error, include scientific notation, `log10(value)`, and nearest base-10 order-of-magnitude notation. For exact zero, report zero and omit logarithms. No `1e-6` pass threshold is introduced. Non-finite values, wrong vector length, invalid Level/Scale, or failed decode invalidate the comparison; preserve the actual evidence and classify the result as failed, not numerically successful.

No speedup, security-equivalence, or correctness-preserving FHE claim follows from this Batch. State clearly that Fast is an intentionally insecure zero-secret simulation and capacity is unproven where applicable.

## Required Primary artifacts and delivery

Create a compact human-readable summary, a durable append-only attempt journal, and necessary raw/log/vector artifacts under task-specific `results/` paths. Check sizes before reading any prior large artifact; do not print full vectors or large JSON into chat. Do not modify Batch 005 or prior BTS artifacts/journals.

Commit and push the Primary task spec/pointer and final evidence normally to `origin/main`; commit and push the Secondary Q-prefix policy amendment first, then the tested Fast repair, normally to `origin/fast-qprefix`. Preserve the original five-file dirty Secondary worktree. Never force push, reset, rebase, stash, overwrite history, modify BTS, or push BTS source.

## Stop conditions and handoff

Stop immediately for a required BTS source/parameter/operation-order change, an unapproved security-mode change, unresolved arithmetic/API semantics, invalid/absent ciphertext representation, a non-fast-forward Git state, or a new mathematical contradiction. The capacity-proof waiver alone is not a stop condition. Report exact evidence and remaining call budget.

Successful terminal labels, in order: `FAST_BTS_COMPLETED`, then `STANDARD_FAST_OUTPUT_COMPARED`, and finally `READY_FOR_WEB_REVIEW`. If Fast cannot complete within the authorized three calls, report the concrete first unresolved API failure, completed repairs/tests, and unused/consumed calls; do not claim the scientific comparison is complete.
