# FAST-LOGN16-EVALMOD-ROOTCAUSE-001 — first divergence, not premature repair

## Authority / task class
Task class: **E with M review gate** (bounded measurement + mathematical diagnosis). Codex collects and validates evidence; GPT Web decides any CKKS semantic correction. **Primary diagnostic-only changes** are permitted, **Secondary production source must remain pinned and unmodified**. Do not implement a speculative fix or rebaseline performance in this task.

## Accepted prior result, not accepted functional correctness
- Primary accepted measurement/result commit: `d24396a76ce7773c8b70200ee0d7633de48f8126`; frozen matched timing harness: `68558a82af6faffd304fa0f14bbc4c82fa190176`.
- Fast Secondary: `5117fc57949647182f476dc5952099c706b9f869`. Genuine historical Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- `results/FAST-STANDARD-PERF-REBASELINE-002-logN16-report.md` and raw fastdiag/paired numerical artifacts are the authoritative evidence. LogN16 full Bootstrap median 3827.880 ms genuine Standard / 675.102 ms Fast; **5.6701x is invalid-output speed only**, not an accepted correct acceleration.
- LogN16 Standard/Fast decoded-domain Bootstrap SNR: 130.947666 / -0.000461910319 dB; final Fast-vs-Standard RMSE 0.020485498. Output real/imag threshold violations: 138780/30306, capacity strict `2B < S_Q`: 56/56 pass.
- At the first **public material** checkpoint `evalmod_real`, max complex diff 0.00932384135 > 0.00364073040 material threshold. **This is not the earliest internal nonzero divergence.** Internal LogN16 generated Chebyshev T2 real RMSE 3.06256424e-10 (imag 3.04223095e-10); T3 real RMSE 9.78484187e-9. Polynomial output real RMSE 9.54849068e-9, last DoubleAngle real RMSE 7.84118323e-8, real EvalMod output after public scale reset RMSE 0.00256939892.
- The raw fastdiag records **unequal exact scales** for T3 and T6 despite almost identical displayed log2(scale). Evidence only: this does **not** prove exact-scale drift is the root cause.
- LogN16 configured q0 target 55 but **effective q0 has 56 bits** (prime 36028797019488257). Preserve the effective chain and input fingerprint for primary reproduction. Do not silently change q0 or infer it is the cause.
- LogN13 at the same Fast commit remains reference PASS: Fast-vs-Standard RMSE 1.44680895e-10, SNR 144.713390 / 144.710750 dB.

## Narrow scientific question
Find the **first operation and mathematical invariant** that explains why LogN16 Fast deviates from the matched Standard in the generated-power / EvalMod path, or give bounded evidence explaining why this is not yet knowable. Separate:
1. true CKKS coefficient/rounding/rescale divergence;
2. representation/projection mismatch in `fastdiag` or comparison of unequal exact-scale states;
3. error amplification that is subsequent to an earlier divergence.

The experimental contract stays `LogN=16, LogSlots=15`, the exact effective Q/P chains, `c0=encoded-message, c1=0`, original SHA `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf`, Mod1 degree 30, DoubleAngle 3 and existing thresholds. This comparator does **not** establish secure-encryption or RLWE noise properties.

## Minimal executable sequence
1. **Audit the existing evidence first; no harness rewrite.** Verify LogN16 original input, effective Q/P, scales, Fast/Standard output and baseline SNR directly from committed reports and current code. Quantify Fast post-Bootstrap output RMS and normalized correlation with pre-Bootstrap values to discriminate near-zero, scaled, and corrupted outputs; do not assume output is zero merely because SNR is near 0 dB.
2. **Check diagnostic comparability.** The deep fastdiag "Standard" evaluator comes from the pinned Fast-source tree, while the authoritative genuine Standard timing/numerical result uses a separate historic Standard commit. Establish whether the relevant Standard generated-power checkpoints and final output can be treated as the same reference (or mark any unverified comparison as conditional). Respect the number and identity of Q limbs, coefficient representation (NTT/Montgomery), Level, logical modulus, exact scale, and common-Q projection. Never convert an uncomparable state into a claimed match by padding/cropping residues without algebraic justification.
3. **Trace the earliest T2 divergence, then T3/T6 only if needed.** Reuse `cmd/fastdiag/numerical.go`, `numerical_lockstep.go`, `numerical_evalmod.go` and their existing generated-power trace hooks. Compare T1 input; T2 multiplication before recurrence correction; corrected pre-Rescale T2; post-Rescale T2; where supported, matching Standard states. For each meaningful common representation capture **exact** scale (not only rounded log2), logical level, actual Q-prefix width, denominator prime used, centered coefficient bound/capacity for the intermediate (including pre-Rescale), RMSE/max difference, and first/worst differing coefficient without emitting full coefficient dumps. If existing API does not expose one substep, implement only the smallest Primary diagnostic hook; do not make any Secondary implementation change.
4. **Prove or refute a precise candidate invariant.** For Rescale, verify the logical divisor `q_l`, the intended level transition `l -> l-1`, and scale transition `Delta' = Delta/q_l` against exact Q residues and an independent centered-integer or reference arithmetic oracle where feasible. Distinguish expected approximate CKKS rounding error from incorrect residue reconstruction, wrong source limb, truncation, or double-rounding; do not declare a bug based solely on nonzero numeric RMSE. Explicitly test whether T3/T6 exact scale differences cause measurable decoded differences rather than guessing. Audit whether the 56 capacity checkpoints omit the suspect *pre-Rescale intermediate* and say so if they do.
5. **Only if necessary, run a narrow controlled discriminator**, e.g. one alternative effective-q0 or Q-prefix-width experiment. Keep it distinct from the canonical LogN16 benchmark, re-create both matched Standard/Fast references under the same modified parameters, and do not call the result an updated speedup. Such a probe is optional; prefer operator-level proof. Do not change production, original benchmark configurations or tolerances.
6. **Protect LogN13.** Run focused existing regression and narrow LogN16 reproduction; reuse raw committed measurements, no repeat of 7x timing campaigns, no additional parallel measurement implementation.

## Evidence and acceptance
Write one compact `results/FAST-LOGN16-EVALMOD-ROOTCAUSE-001.md` with:
- source SHAs/clean states and explicit reused vs fresh observations;
- an ordered table of **comparable** T1 -> T2 substeps -> T3/T6 -> polynomial -> DoubleAngle -> scale reset -> S2C, with earliest discrepancy, exact-scale/level/representation qualifications;
- a concrete proposed causal mechanism **only if its predicted effect and arithmetic invariant are tested**, otherwise competing hypotheses + discriminator;
- whether the difference is production arithmetic, diagnostic representation artifact, or not yet isolated; output RMS/correlation and counterevidence;
- LogN13 unchanged check; tests; the existing unrelated `TestFIX001P3GenuineStandardPublicVsStagedConsistency` failure must not be concealed or treated as new.

Exit status exactly one:
- `LOGN16_ROOTCAUSE_ISOLATED`: precise operator + violated/incorrect invariant demonstrated with reproducible evidence;
- `LOGN16_DIAGNOSTIC_ARTIFACT`: source/comparison evidence shows reported divergence is a measurement artifact;
- `LOGN16_ROOTCAUSE_UNRESOLVED`: nontrivial evidence collected, causal invariant not established;
- `LOGN16_DIAG_BLOCKED`: reproducible execution/semantic comparison is blocked.

If an arithmetic/cryptographic design choice is needed, **stop at `NEEDS_WEB_REVIEW`** rather than coding a guessed fix. Run relevant focused tests and `git diff --check`; no regression to LogN13. Commit/push only authorized Primary diagnostics and report; end `READY_FOR_WEB_REVIEW`. The existing build-tag backend-adapter vs `AGENTS.md` architecture concern is **recorded review debt**, not a license for an unrelated refactor in this task.
