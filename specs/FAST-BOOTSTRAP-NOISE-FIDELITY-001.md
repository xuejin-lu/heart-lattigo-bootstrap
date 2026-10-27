# FAST-BOOTSTRAP-NOISE-FIDELITY-001 — Three-way bootstrap precision study

**Status: ACTIVE — explicitly activated by the user on 2026-10-27.** This task supersedes `FIXED-SIGMA-DIRECT-MUL-004` as the Primary current-task pointer but does not alter its historical artifacts. CP-A and CP-B are authorized. The user also explicitly authorizes the CP-C pilot budget below: at most one successful full Bootstrap per arm (S, F, FN), 3 total, with no retry or hidden Bootstrap warmup. CP-D requires a later Web decision; CP-E remains unauthorized.

## Primary research goal
Make full Bootstrap output precision of intentionally insecure Lattigo_FAST match genuine Standard CKKS as closely as possible with **minimal Fast code changes and minimal runtime overhead**, rather than trying to reproduce every internal ciphertext noise term. First finish a complete, scientifically comparable three-arm measurement even if noise fails to improve precision. Preserve Fast zero-secret semantics and all original Q-prefix and safety invariants. No security-fidelity claim.

## Immutable sources and branch
- Control plane/measurement: `xuejin-lu/heart-lattigo-bootstrap/main`.
- Genuine pristine Standard: `xuejin-lu/lattigo@5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Frozen Fast control: `fast-qprefix@1e7828293c36f9845e30723ba4e9500d84bb50d2`.
- New development branch, already created from exact Fast control: `xuejin-lu/lattigo:experiment/fast-bootstrap-noise-fidelity`. Do not change `fast-qprefix`, genuine Standard, or old 001–004 captures. Branch-local `AGENTS.md` requires an explicit branch exception on activation, otherwise it still mandates `fast-qprefix`.
- Same Primary frontend/application logic, same effective parameters and input, three isolated pinned dependency builds. Never call a Fast-linked build “Standard”; do not change code merely to make labels agree.

## Mandatory three-arm comparison
| Arm | Library | Injection |
| --- | --- | --- |
| S | genuine Standard | none |
| F | experiment Fast | disabled, must match frozen Fast |
| FN | **same Fast binary as F** | explicit pre-ct×ct injection enabled |

For each arm independently compare decoded **full Bootstrap** output `y` with the **same original input message m**:
`RMSE=sqrt(mean(|y_i-m_i|^2))`; `precision_bits=-log2(RMSE)`, with explicit undefined/zero/non-finite handling. Compute `abs(P_F-P_S)`, `abs(P_FN-P_S)`, `P_FN-P_S`, and change in absolute bit distance. Record **also** S-vs-F, S-vs-FN and F-vs-FN output pairwise RMSE, but do not confuse those with input-relative precision. `-log2(global RMSE)` differs from paper's mean per-slot precision; do not silently equate them. Report MaxAbs, p99/p99.9 where supported.

For each arm record bootstrap-only wall time (cold/warm provenance), allocations and peak memory if natively available, speedup relative S, overhead FN vs F. Separate setup/key generation/encode/decode from bootstrap timing. Identical input fingerprint, slot order, native Standard decrypt vs documented Fast projection, LogN/E/Scale/Q/P, host, source, binary SHA and `go version -m` VCS metadata are mandatory for comparable results. Never compare unmatched host timing as speedup.

## Hypothesis H1: minimal Fast-only pre-multiplication noise
An opt-in hook before actual **degree-1 ciphertext × degree-1 ciphertext** multiplication, modifying **temporary local copies**, never caller originals:
`A'=(A.c0+eta_A,0)`, `B'=(B.c0+eta_B,0)`, then original product. Every call and side draws a fresh independent length-N signed Gaussian integer coefficient vector of the same sigma, with a stable root-seed/operation/side derivation; if operands alias (square), they still get independent local draws. This is an exploratory model and differs mathematically from reusing one noisy ciphertext twice. Start candidate `sigma=4.004280012097432` inherited **only from post-MulRelin-Rescale** experiments; it is not pre-multiplication calibrated. Do not tune it merely to achieve a pleasing bits number.

Source review candidates: `schemes/ckks/internal/fastcore/mul.go:MulWorkspace.ApplyRows`, `schemes/ckks/fast/evaluator_ntt.go`, public Fast dispatch, `circuits/ckks/polynomial/fast.go` and `circuits/ckks/mod1/fast.go`. The shared core handles ct×pt too; **never inject into ct×pt/scalar**, nor unconditional real Standard. Confirm every intended Bootstrap ct×ct path actually traverses the hook; identify uncovered paths. Reuse existing `schemes/ckks/evaluator_fast_fixed_sigma_noise.go` Gaussian/RNS/NTT and strict Q-prefix/centered-capacity validation, but first inspect existing Level/Scale restrictions. Prefer a small reusable helper, not a new noise framework. Preserve NTT/Montgomery format, active Q-prefix row sign consistency, c1=0, metadata, Scale, Level, no on-error partial mutation. OFF must be exactly backward-compatible and have negligible extra overhead. Track operation counts, injected-polynomial counts, sampled sigma/RMS, seed and injection stages. Fail closed on invalid input.

## Staged execution and experiment budgets
**CP-A, no compute:** Safely sync both repos; inventory sources/branch instructions and existing measurement support; inspect exact Bootstrap ct×ct pathways/aliasing and source noise helper; record resource and coverage plan. Read file sizes before inspecting large JSON, logs or CSV.

**CP-B, no Bootstrap:** Implement only the smallest opt-in H1 change, branch-local guard and focused regression tests: OFF equivalence; each left/right seed distinct; alias/square preserved; no ct×pt; correct RNS/NTT; correct Q-prefix rejection without mutation; all public and explicit Fast Bootstrap multiplication calls covered. Reuse established `cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, `internal/numericalmetrics`. Do not invent a new runner/comparator/statistics engine. No implicit backend substitution.

**CP-C, separately authorized pilot:** ONE successful full Bootstrap call per arm (3 total), one predeclared canonical input, preferably accepted LogN13/E32/Scale 2^45 profile if its current support and exact matching are verified. No hidden warmups, repeats, Bootstrap triggered by tests, LogN16 or BTS. Freeze input SHA, root noise seed, parameter profile, source and binaries in advance. Record every attempt and any pre-compute failures. If a failure or mismatch occurs, stop; do not silently retry or relax capacity gate.

**CP-D, separately authorized:** After Web review of CP-C, optionally add two other preregistered input/noise seeds (6 more Bootstrap calls, at most 9 cumulative) for repeatability, or a matched warm-speed campaign with independently specified bounded budget. Do not automatically run CP-D.

**CP-E, not authorized:** If H1 fails, analyze whether one alternative single hook (e.g. an aggregate noise compensation after CoeffsToSlots) is appropriate. Requires separate explicit design and evaluation; do not stack additional noise sources automatically. Do not claim that one fitted end-output RMSE proves a faithful CKKS noise model.

## Reporting and preservation
Required table, each entry observed only:

| Arm | Input fingerprint | Bootstrap success/slots | RMSE vs input | Precision bits vs input | MaxAbs/p99 | Bootstrap time | Memory |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S | | | | | | | |
| F | | | | | | | |
| FN | | | | | | | |

After execution, also report pairwise output errors, bit-distance improvement, runtime overhead, injected polynomial counts and validation status. Captures, journals and byte hashes belong under new `results/FAST-BOOTSTRAP-NOISE-FIDELITY-001/` only after authorization; old data immutable. Keep large vectors outside Git where appropriate, with source fingerprints.

**Research acceptance:** `TRIPLET_COMPLETE_EXPLORATORY` means valid three-way comparable measurements, **not** successful sigma match. A target of ±2 bits may be shown as a loose descriptive screen only. `NOT_COMPARABLE` or `SCOPE_REVIEW_REQUIRED` if any correctness/provenance or calculation-budget gate fails.

## Existing authoritative rules
Before any implementation read Primary `AGENTS.md`, `docs/MEASUREMENT_PLATFORM.md`, `docs/BACKEND_TRANSPARENCY_CONTRACT.md` and Secondary `AGENTS.md`, `docs/FAST_QPREFIX_SPEC.md`. Never use `pull`, reset, rebase, stash, force-push, or destroy unknown local changes; do not repeat spent calls. Any Codex chat prompt is only a pointer to the versioned specification.

## Research context
Bossuat et al., *Accurate and Composable Noise Estimates for CKKS with Application to Exact HE Computation*, especially component-wise noise, squaring covariance, Rescale/KeySwitch, and simulated Bootstrap. `https://github.com/tuneinsight/ckks-noise-estimator`. It motivates H1's evaluation, not a claim H1 reproduces their estimator or Standard CKKS noise.

## Activation record — 2026-10-27

The user explicitly activated this versioned task and instructed the work to use `experiment/fast-bootstrap-noise-fidelity`, never `fast-qprefix`, then run the authorized three-arm pilot budget. The Primary `CURRENT_TASK.md` pointer was updated accordingly. This activation authorizes only CP-A, CP-B, and CP-C's single three-arm pilot; it does not activate CP-D or CP-E.

## Web Review recovery ruling — 2026-10-27

`AUTHORIZED_SCOPED_REPAIR_AND_CP_C_RESUME` authorizes the following bounded repair and continuation without changing the original CP-C experiment budget:

- Only `public-native-bootstrap-once` / frozen LogN13 / E32 Standard preflight may bind to the exact task Fast commit `b81f54c2e0f98dc2b4a7474fb333fbae79be8c63`. Existing public Bootstrap and repeatability modes retain `publicCurrentFastSHA`; the new mode must reject every other Fast pin.
- In the one-call mode only, final output MaxAbs versus `1e-6` is observational. Preserve the measured metrics and accurately record the gate result; all provenance, preflight, input, capacity, representation, decode, slot-count, finite-value, and data-integrity checks remain fail-closed. Historical mode numerical-gate semantics remain strict. A completed one-call observation must not use a status label that claims the numerical gate passed.
- The new Bootstrap fixed-sigma helper may omit per-phase `runtime.ReadMemStats`; its sampling, capacity, Level/Scale, and event evidence remain. The historical post-Mul/Rescale helper retains its allocation-stat behavior.
- Rebuild clean source-pinned binaries and establish a fresh zero-call preflight pair after this repair. Earlier `d3d5b0a` preflight and binaries remain historical and are not proof of the repaired source. The precompute pin mismatch remains in the attempt ledger. CP-C remains at most one successful call per arm and three total; no retry, warmup, LogN16, BTS, CP-D, or CP-E is authorized by this ruling.

After every fresh source, binary, preflight, parameter, input, and capacity gate passes, resume S, F, FN sequentially using the already frozen input, sigma, and root seed. Stop on any new failure; do not consume another call.
