# FIXED-SIGMA-DIRECT-MUL-004 — direct-mul experiment result summary

**BATCH_IN_PROGRESS.** CP-A and CP-B passed. Independent Web review authorized exactly one same-seed replacement launch, F5-R1, for the known pre-chain F5 CLI parse failure. F5 remains in history; F5-R1 has not yet launched.

**Research question:** Does σ calibrated on the historical `(A+B)×C` fixture transfer to direct `A×B`, followed by `(A×B)×C`, with no Add?

Frozen σ: `4.004280012097432`. Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Fast: `1e7828293c36f9845e30723ba4e9500d84bb50d2`. The new `direct-base` and `direct-c4` input identities remain separate from 001–003.

| Captured evidence | Standard | Fast OFF/OFF | Fast ON/OFF | Fast OFF/ON | Fast ON/ON |
|---|---|---|---|---|---|
| Direct-base first seed pair (S1, F1–F4) | Captured | Captured | Captured | Captured | Captured |
| Direct-c4 baseline first seed pair (S2, F9) | Captured | Captured | — | — | — |
| Stagewise incremental RMS, variance, bias, max | Not analyzed | Not analyzed | Not analyzed | Not analyzed | Not analyzed |
| Total error, propagation decomposition and p95/p99/p99.9 tails | Not analyzed | Not analyzed | Not analyzed | Not analyzed | Not analyzed |
| Frozen ±10% transfer screen | No conclusion | No conclusion | No conclusion | No conclusion | No conclusion |
| Effective Level/Scale, capacity and backend dispatch | CP-B passed | CP-B passed | — | — | — |

## Checkpoints

- **CP-A — complete.** Effective LogN13/E32, 4096 slots, Δ=`2^45`, RingN=8192, Q=`q0…q16`; Levels 5→4→3 and four active Q-prefix rows `q0..q3`. Effective `q4=549756174337`, `q5=1152921504606830593`; strict `2B < q0q1q2q3` capacity checks passed for both workloads. Hashes and bounds are in [journal.md](journal.md).
- **CP-B — complete.** Two genuine Standard OFF/OFF references and two Fast OFF/OFF references were validated against the same frozen public workload/profile. Plaintext oracles, strict capacity, levels/scales, source dispatch, and Standard/Fast mode invariants passed.
- **CP-C — in progress.** The direct-base first seed pair completed all four Fast arms. Direct-c4 has only its first-pair OFF/OFF reference. F5 failed during argument parsing because `mul-rescale-chain` was passed as a positional argument; the runner expects the boolean `--mul-rescale-chain` flag. It returned before any chain and created no F5 result/vector files. Web authorized one linked same-seed launch, F5-R1; no other replacement is allowed.
- **CP-D — not started.** The aggregate paired ablation analysis waits until CP-C completes.

## Durable attempt counters

- Fast actual two-stage chain executions completed: **5/16**
- Fast launch attempts recorded, including original F5 parser failure: **6/17**
- Known pre-chain CLI failures / authorized maximum: **1/1**
- F5-R1 replacement slot: **authorized, not launched**
- Genuine Standard completed two-stage chains: **2/2**
- Bootstrap: **0/0**
- LogN16 workload: **0**
- BTS: **0**

Seven validated capture pairs (2 Standard, 5 Fast; 55,915,277 bytes) and the durable ledger are under this directory. `captures/SHA256SUMS` indexes each capture. Binary SHA-256: Standard `2a44e3f89d055d4a70b294d70b9b265edee3e9f5451c5877a3491f1f2ae2a83c`; Fast `a7083bc12aa1d3da062c0b094b99f5c4fbcf3ceb8d86ea92f981ad73ab9db2ff`. Both were built from Primary source commit `977d853541cf59b37e42cdfc1e635bc0cab7862d` with isolated Standard/Fast module sources.

The user-relayed independent Web ruling is `AUTHORIZED_ONE_PRECHAIN_REPLACEMENT`: it preserves original F5 as a launch attempt, classifies it as zero actual chain executions, and authorizes exactly one same-seed replacement F5-R1. The matched two-pair matrix is still incomplete, so no RMS transfer conclusion, propagation decomposition, or tail-statistics claim is available yet. σ and all seeds remain unchanged. Bootstrap, LogN16 workload, BTS, and large benchmark counts remain zero. A second unexpected failure has no replacement authorization.
