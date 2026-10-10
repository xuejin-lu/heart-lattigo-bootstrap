# FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001

## Objective and authority

Numerical revalidation ONLY. User requests a FRESH run of current LogN16 Fast versus Standard, with LogN13 as control, before any optimization. Previous LogN16 Fast–Standard RMSE 0.02048549798 (EvalMod real 0.00256939892) was measured on obsolete Secondary 5117fc57949647182f476dc5952099c706b9f869. Its LogN13-only Chebyshev generated-power arithmetic selection was subsequently removed. Those historical numbers MUST NOT be described as current results.

Follow both repositories' AGENTS.md; read Primary docs/MEASUREMENT_PLATFORM.md and reuse cmd/fastdiag numerical/lockstep, cmd/perfprobe as applicable, internal/perfmeasure and internal/numericalmetrics. Do not invent a new benchmark or metrics harness. No new performance benchmark and no changes to Fast production arithmetic, CKKS semantics, coefficient reconstruction, normalization, or scale compensation.

## Sources and controls

- Last known Secondary fast-qprefix HEAD 2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f; confirm latest origin/HEAD clean and pin actual SHA in report. If updated, audit differences and record actual revision.
- Primary task-planning base a857077577f78dbe0113690fcfcf84ccd782cc7b; actual SHA must be recorded.
- Genuine Standard historical reference SHA 5dbffbdea05394de2ca3a432ed5318aa832e3f40. Deep stage replay using Standard API compiled inside Fast source tree is NOT automatically identical to an independently built Standard reference; label both separately.
- Reuse source-backed configs/bootstrap_config.logN13.json and bootstrap_config.logN16.json and the exact prior numerical input generator/metadata/fingerprints from results/FAST-STANDARD-PERF-REBASELINE-002-logN13-numerical.json and -logN16-numerical.json. Preserve *effective* Q/P primes; LogN16 effective q0 is 56-bit even though the literal requests 55. Do not change it. Preserve LogN13 LogSlots12 and LogN16 LogSlots15.
- This experiment deliberately reproduces historical deterministic encoded c0 / c1=0 numerical comparator workload to isolate software-version effects. DO NOT silently replace with E32 formal public encryptor profile. Label the profile.

## P1 — shared harness unblock and source audit

After mandatory safe sync, inventory shared tools. Known blocker: Primary cmd/fastdiag/numerical.go pins correctedFastCommitSHA=5feb44917fca40c93abec6def6f26bc81a82c536 and rejects latest Fast. Repair the reusable numerical harness's version selection by adding explicit labelled current mode while preserving historical mode; do not disable clean-tree/provenance checks. Add focused tests. Check current Secondary circuits/ckks/polynomial/fast.go, circuits/ckks/mod1/fast.go, circuits/ckks/bootstrapping/fast_evalmod.go and schemes/ckks/fast/q012.go, including which Q-prefix width the actual explicit-row caller uses, and whether LogN causes any branch in the mathematical Chebyshev recurrence.

Existing Secondary TestFastBootstrapActualLogN16FullSlotProfile verifies output structure, NOT decoded RMSE; do not mistake its success for numerical correctness.

## P2 — fresh numerical experiments

Perform ONE matched profile comparison each for LogN16 and LogN13 control using latest pinned Fast. Prefer at most 2 Fast full Bootstrap calls and 1 independently genuine Standard full Bootstrap call per profile; if deep stage replay calls are additional, disclose actual counts and limit to one extra call per backend per profile. This is newly authorized numerical-only execution, NOT reusing earlier exhausted performance budgets. Do NOT perform seven-run timing loops.

For both LogN and real/imag where applicable, record RMSE, max absolute complex error, Level, exact Scale, Q-prefix rows, comparator provenance and whether states are mathematically comparable:
- input, scale_down, mod_up, c2s_real, c2s_imag, evalmod_real, evalmod_imag, s2c, final_public_output;
- each EvalMod branch: pre/post normalization, generated Chebyshev T1/T2/T3/T6, polynomial output, each of 3 DoubleAngle post-Rescale states, immediately before and after public scale reset.
- At different RNS widths, use only existing validated common-Q projection. Never invent a direct decoded comparison if not mathematically comparable.
- Also compare full output of EACH backend against original input message (RMSE, SNR, max), not only Fast–Standard. Genuine Standard fresh comparison is preferred; if unobtainable, clearly mark restricted authority.
- Check first fresh divergence and whether Chebyshev calculation ordering follows multiply → recurrence correction → Rescale in both LogN.

## P3 — report and no arithmetic changes

Create results/FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001.md and matching compact .json. Include both LogN values, fresh timestamp, exact SHAs, config/hash, input fingerprint, stage and internal RMSE table, original-vs-output checks, call counts, provenance, explicit old vs new comparisons. Suggested rigorous reporting gate: final Fast–Standard complex RMSE <=1e-7, Fast-vs-original <=1e-7, and comparable EvalMod stage RMSE <=1e-7; give actual numbers regardless of status. Distinguish threshold criterion from proof of coefficient identity.

Classify exactly one:
- LOGN16_CURRENT_NUMERICALLY_CLOSE: new LogN16 numerical closure demonstrated with comparable reference and actual outputs.
- LOGN16_CURRENT_STILL_DIVERGES: new run fails; identify earliest divergent operator and numerical invariant; DO NOT fix or disguise with pre-scale workarounds.
- LOGN16_CURRENT_REVALIDATION_BLOCKED: source/tool/environment prevents meaningful new numerical run. Explain blockers, no invented new RMSE.

Use existing tooling and focused tests, go test and git diff --check. Only Primary harness/test/report changes authorized. Secondary production code, arithmetic and benchmark configurations unchanged. Commit/push Primary results (or precise blocked report); READY_FOR_WEB_REVIEW, no automatic optimization Batch.