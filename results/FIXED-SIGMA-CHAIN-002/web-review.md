# Independent GPT Web review — FIXED-SIGMA-CHAIN-002

**Date:** 2026-10-22 Asia/Taipei
**Disposition:** `ACCEPTED_SCOPED_RMSE_ONLY`. Not a proof of standard CKKS noise distributions, correlation structure, tails, or general automatic Bootstrap injection.
**Primary reviewed:** `25de6e43386cc16baf9f81bf455356b2da49ad1b`
**Secondary reviewed:** `1e7828293c36f9845e30723ba4e9500d84bb50d2`
**Genuine Standard source pin:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.

## Independently verified evidence

- GitHub `main` and `fast-qprefix` both point at reported heads. Relative to last review, Primary advanced six commits and Secondary two. The Primary diff adds one maintained `cmd/perfprobe/fixed_sigma_chain.go` adapter with tests and artifacts, not a second general measurement framework. Secondary changes are isolated to the existing opt-in noise helper and its tests.
- The actual source passes `stage1` (after an explicit noise addition) as the `ctInputX` for the second real `MulRelinNew → Rescale`, followed by its own explicit noise addition; `stage2Seed=stage1Seed+1`. Output Levels 4 then 3 with Scale `2^45`, q5 and q4 rescale divisors, and q0..q3 active residues are asserted. **Two actual dependent operations** are demonstrated, not two disconnected single-operation tests.
- Frozen `σ_coefficient=4.004280012097432` from 001 was used for all injections, without level or seed retuning. Three CP-C paired artifacts each report two injections, zero Bootstrap, and Level4/Level3 incremental residuals close in RMS to matched genuine Standard (order `1.0e-11`). The source computes incremental `r_op` using actual decoded inputs at each stage and stores total-plaintext-oracle errors separately.
- CP-B attempted three matched comparisons but correctly **excluded pairs 02/03** after discovering binary-embedded VCS and runtime Primary source SHA disagreement; only pair 01 was retained as accepted baseline. The updated harness now fails provenance mismatch instead of silently accepting a wrong comparator.
- CP-D measurement documents three Fast runs (and one reused Standard reference) rather than three independent Standard captures; phase accounting shows Q-prefix capacity verification about 66–68% of summed helper phase time. Per-stage median phase-sum / Mul+Rescale is about 53.78% (stage 1) and 65.31% (stage 2). The outer timer is larger because instrumentation itself takes ~0.20–0.26ms. Alloc snapshots are process-wide diagnostics, not precise isolated allocation benchmarks.
- Source of Secondary helper still validates zero-secret shape, c1=0, Scale/Level/domain, canonical residues and strict centered Q-prefix capacity before committing c0; no code path automatically intercepts general CKKS MulRelin+Rescale calls.

## Remaining scientific gaps / important caveats

1. Stage2 dependent input is **structurally confirmed**. But since stage2 `r_op` subtracts the product of actual decoded inputs, it **removes preceding error propagation**; two-step RMS agreement does not by itself isolate how stage1 injected error propagates. A dedicated ablation/perturbation experiment is needed.
2. The same original A/B/C workload and fixed Levels/Scale are used across three CP-C captures. Held-out noise seeds are not held-out *plaintext circuits or distributions*; universal fixed σ and long-chain behavior remain unproven.
3. Tail mismatch persists: standard residual Max values are frequently ~4–7e-11 versus Fast+σ ~2.7–3.1e-11 across CP-C pairs. Variances differ and paired variance signs vary; small covariance magnitude across spatial slots does **not** establish model independence.
4. Fast helper is explicitly called in this test harness after each complete operation and supports only the bounded LogN13 Level4/3 Scale2^45 zero-secret profile. No arbitrary automatic multiplication hook, general Rescale hook, other Level/Scale profile, PK parity or Bootstrap coverage is achieved.
5. Strict Q-prefix capacity verification is scientifically necessary under the adopted compact representation; 66–68% phase dominance motivates *equivalent verified optimization*, never omitting validation to improve benchmark numbers.
6. Full Standard-tag suite has recorded historical A=0/modfile fixture failures; task-specific Standard, Fast and Secondary test/vet results are recorded as passing. This review used independently fetched **bounded** GitHub source, journal and compact artifacts, **not** a local rerun or exhaustive inspection of external raw vectors.

## Web decision

`FIXED-SIGMA-CHAIN-002 = ACCEPTED_SCOPED_RMSE_ONLY` for two dependent ct×ct operations with frozen σ, plus `PASS_SCOPED_ATTRIBUTION` for diagnostic cost phases. Historic results stay immutable. Next authorized task `FIXED-SIGMA-PROPAGATION-003` uses ablation and distinct input strengths to isolate propagation, then gated safe cost improvement and a **read-only** Bootstrap callsite map. Bootstrap execution budget remains 0.
