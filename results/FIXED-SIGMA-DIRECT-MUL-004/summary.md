# FIXED-SIGMA-DIRECT-MUL-004 — direct-mul experiment result summary

**BATCH_IN_PROGRESS.** CP-A and CP-B passed. The authorized F5-R2 `direct-base` OFF/OFF chain passed all seven oracle and capacity checks. Fast use is 8/18 launch attempts and 6/16 actual chains; ten frozen Fast slots remain.

**Research question:** Does σ calibrated on the historical `(A+B)×C` fixture transfer to direct `A×B`, followed by `(A×B)×C`, with no Add?

Frozen σ: `4.004280012097432`. Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Fast: `1e7828293c36f9845e30723ba4e9500d84bb50d2`. The new `direct-base` and `direct-c4` input identities remain separate from 001–003.

| Captured evidence | Standard | Fast OFF/OFF | Fast ON/OFF | Fast OFF/ON | Fast ON/ON |
|---|---|---|---|---|---|
| Direct-base first seed pair (S1, F1–F4) | Captured | Captured | Captured | Captured | Captured |
| Direct-base second seed pair (F5-R2, F6–F8) | — | Captured | Not launched | Not launched | Not launched |
| Direct-c4 first seed pair (S2, F9–F12) | Captured | Captured | Not launched | Not launched | Not launched |
| Direct-c4 second seed pair (F13–F16) | — | Not launched | Not launched | Not launched | Not launched |
| Stagewise incremental RMS, variance, bias, max | Not analyzed | Not analyzed | Not analyzed | Not analyzed | Not analyzed |
| Total error, propagation decomposition and p95/p99/p99.9 tails | Not analyzed | Not analyzed | Not analyzed | Not analyzed | Not analyzed |
| Frozen ±10% transfer screen | No conclusion | No conclusion | No conclusion | No conclusion | No conclusion |
| Effective Level/Scale, capacity and backend dispatch | CP-B passed | CP-B passed | — | — | — |

## Checkpoints

- **CP-A — complete.** Effective LogN13/E32, 4096 slots, Δ=`2^45`, RingN=8192, Q=`q0…q16`; Levels 5→4→3 and four active Q-prefix rows `q0..q3`. Effective `q4=549756174337`, `q5=1152921504606830593`; strict `2B < q0q1q2q3` capacity checks passed for both workloads. Hashes and bounds are in [journal.md](journal.md).
- **CP-B — complete.** Two genuine Standard OFF/OFF references and two Fast OFF/OFF references were validated against the same frozen public workload/profile. Plaintext oracles, strict capacity, levels/scales, source dispatch, and Standard/Fast mode invariants passed.
- **CP-C — in progress.** The direct-base first seed pair completed all four Fast arms. Direct-c4 has only its first-pair OFF/OFF reference. Original F5 and authorized F5-R1 remain recorded as separate pre-chain failures with no ciphertext operation. Web authorized exactly one further linked F5-R2 launch. Its new binary provenance and corrected command have passed prelaunch validation; it remains unlaunched.
- **CP-D — not started.** The aggregate paired ablation analysis was not run because CP-C stopped at F5-R1.

## Durable attempt counters

- Fast actual two-stage chain executions completed: **6/16**
- Fast launch attempts recorded, including original F5 and F5-R1 failures: **8/18**
- Original authorized F5 pre-chain CLI parse failure: **1/1**
- F5-R1 pre-chain binary provenance failure: **1**, permanently recorded
- F5 recovery launches authorized / launch-reserved: **2 / 2**
- F5-R2 replacement slot: **PASS_VALIDATED**
- Genuine Standard completed two-stage chains: **2/2**
- Bootstrap: **0/0**
- LogN16 workload: **0**
- BTS: **0**

Eight validated capture pairs (2 Standard, 6 Fast; 63,376,794 bytes) and the durable ledger are under this directory. `captures/SHA256SUMS` indexes each capture. Existing captured binary SHA-256: Standard `2a44e3f89d055d4a70b294d70b9b265edee3e9f5451c5877a3491f1f2ae2a83c`; original Fast first-pair `a7083bc12aa1d3da062c0b094b99f5c4fbcf3ceb8d86ea92f981ad73ab9db2ff`; F5-R2 and subsequent remaining-pair Fast captures use `bde01f61c36f70dc70939eb12d72f3a42ce07a925942904939b54e516cafae17`, built from Primary `977d853541cf59b37e42cdfc1e635bc0cab7862d` with `vcs.modified=false` and `perf_fast`.

The first Web ruling preserved F5 and authorized F5-R1; the later ruling authorized F5-R2, which has now passed as the same-seed direct-base OFF/OFF baseline. Launch attempts and actual chains remain distinct: current Fast use is 8/18 launches and 6/16 actual chains, with at most 10 additional actual chains possible. Stagewise residual and oracle results for F5-R2 are recorded in [journal.md](journal.md). The full two-variant/two-pair matrix and aggregate propagation/tail analysis remain incomplete. Bootstrap, LogN16 workload, BTS, and large benchmark counts remain zero.
