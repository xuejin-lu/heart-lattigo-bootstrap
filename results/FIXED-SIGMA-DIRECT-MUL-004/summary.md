# FIXED-SIGMA-DIRECT-MUL-004 — direct-mul experiment result summary

**BATCH_IN_PROGRESS** — CP-A passed; no ciphertext-chain numerical results have been collected yet.

**Question:** Does the σ calibrated on the historical `(A+B)×C` fixture transfer to genuinely direct `A×B` and `(A×B)×C` with no Add?

Frozen σ: `4.004280012097432`; genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; active Fast initial `1e7828293c36f9845e30723ba4e9500d84bb50d2`. New inputs `direct-base`, `direct-c4` must be fingerprinted separately from 001–003.

| Workload / metric | Standard | Fast OFF/OFF | Fast ON/OFF | Fast OFF/ON | Fast ON/ON |
|---|---|---|---|---|---|
| Direct-base Stage1 incremental RMS, variance, max | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Direct-base Stage2 incremental RMS, variance, max | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Direct-base Stage2 total oracle error | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Direct-c4 Stage1 and Stage2 metrics | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Propagated `P1`, exact residual decomposition and tails | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |
| Actual Level/Scale, strict capacity and built source | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN | NOT_RUN |

## CP-A preflight

Effective profile: LogN13/E32, 4096 slots, Δ=`2^45`, RingN=8192, Q=`q0…q16`; the experiment touches Levels 5→4→3 and four active Q-prefix rows `q0..q3`. Effective `q4=549756174337`, `q5=1152921504606830593`, Q/P SHA-256 `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`.

Both direct-base and direct-c4 passed all eight strict `2B < q0q1q2q3` checks. Their exact input hashes, workload hashes, bounds and CP-A source/test evidence are in [journal.md](journal.md). The shared `perfmeasure` focused test passed. No ciphertext chain, Bootstrap, LogN16 workload or BTS has run.

## Durable attempt counters

- Fast two-stage chains: `0/16`
- Genuine Standard two-stage chains: `0/2`
- Bootstrap: `0/0`
- LogN16 workload / BTS: `0/0`

CP-B remains in progress. Frozen σ, seed schedule, workload identities and ±10% descriptive screen are fixed before numerical capture. See the journal for the current checkpoint table and safe synchronization evidence.

`SCOPED_RMS_TRANSFER_SCREEN_PASS` only if two stagewise relative RMS discrepancies to genuine Standard are within predeclared ±10% for **both workloads** and all tested ON/ON seed arms; this is a descriptive screen, not statistical equivalence. Always separately report variance, signed bias, p95/p99/p99.9 and maximum mismatch, and any failure to meet the band. No tuning σ after viewing results.

**Authorized budgets:** Fast 0/16 used, Standard 0/2 used, Bootstrap 0/0. Raw results, source/binary provenance and hash-indexed compact evidence: NOT_RUN.
