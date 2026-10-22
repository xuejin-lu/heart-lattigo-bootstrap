# FIXED-SIGMA-DIRECT-MUL-004 — direct-mul experiment summary

**BATCH_BLOCKED_NEEDS_WEB_REVIEW.** All authorized captures are complete and validated. The full matched comparator finds that all four ±10% ON/ON RMS screens pass, but its propagation decomposition roundoff gate fails for every seed pair. No threshold or parameter was changed. Further work requires Web review of the residual/tolerance interpretation.

## Experiment and provenance

- Workload: two public `MulRelinNew→Rescale` operations, direct `A×B` then output×`C`; `direct-base` and `direct-c4`; LogN13/E32, 4096 slots, Level5→4→3, frozen `σ=4.004280012097432`.
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; 2/2 chains. Fast source: `fast-qprefix@1e7828293c36f9845e30723ba4e9500d84bb50d2`; Fast capture build source: Primary `977d853541cf59b37e42cdfc1e635bc0cab7862d`.
- Fast counts: 18/18 launch attempts, including the two preserved F5 pre-chain failures; 16/16 actual chains. Bootstrap 0, LogN16 0, BTS 0, multiparty 0. No successful Standard or Fast capture was repeated.
- F10–F12 used the unchanged F13-qualified Fast binary SHA-256 `bde01f61c36f70dc70939eb12d72f3a42ce07a925942904939b54e516cafae17`, with `vcs.revision=977d853...`, `vcs.modified=false`, and `perf_fast`.

## Authorized baseline reuse

F9 was preserved at its actual seeds `20261105/20261106` and original binary `a7083bc12aa1d3da062c0b094b99f5c4fbcf3ceb8d86ea92f981ad73ab9db2ff`. For the first `direct-c4` pair's OFF/OFF analysis input only, the comparator references F13 at its actual seeds `20261107/20261108` and binary `bde01f61…`. The original F9 artifact is recorded as the equivalence witness. No record is relabeled as a new capture.

F9/F13 both have zero injections and matching Primary/Secondary source, Q/P, inputs, workload, effective parameters, capacity, environment, all 24 active input row hashes and all seven checkpoint records. Their decoded vector JSON is byte-identical: SHA-256 `b1abc549359dbb3c3b293c9e5de56b1e76c7888fd35e9c973357f62de3d61f4e`; Git blob SHA-1 `db809ceeb183112fa35ae5305006ff2533d41814`. The comparator output explicitly records `baseline_origin=F13`, both true source seed pairs, both binaries and all four result/vector hashes.

## ON/ON stagewise Standard RMS comparison

Relative difference is `abs(Fast−Standard)/Standard`. The frozen screen is ±10% per stage.

| Workload / seeds | Stage 1 Standard → Fast RMS | Relative | Stage 2 Standard → Fast RMS | Relative |
|---|---:|---:|---:|---:|
| direct-base 20261101/02 | 1.03075e-11 → 1.02537e-11 | 0.522% | 1.03438e-11 → 1.04213e-11 | 0.749% |
| direct-base 20261103/04 | 1.03075e-11 → 1.04125e-11 | 1.019% | 1.03438e-11 → 1.03857e-11 | 0.404% |
| direct-c4 20261105/06 | 1.04983e-11 → 1.03294e-11 | 1.609% | 1.03111e-11 → 1.04039e-11 | 0.900% |
| direct-c4 20261107/08 | 1.04983e-11 → 1.04009e-11 | 0.928% | 1.03111e-11 → 1.04200e-11 | 1.057% |

All four ON/ON pairs pass the descriptive RMS screen. OFF/OFF stage RMS is about `7.38e-13` to `7.47e-13`; enabled-stage residual RMS is about `1.03e-11` to `1.04e-11`. Tail, bias, variance, total-oracle and propagation measurements for every arm are preserved in `analysis.json`; this RMS screen does not establish distribution or tail equivalence.

## Propagation decomposition gate

The comparator reports 4,059 of 4,096 slots outside its current `gamma_64` tolerance in each pair:

| Workload / seeds | Maximum residual | Maximum tolerance | Slots outside tolerance |
|---|---:|---:|---:|
| direct-base 20261101/02 | 3.85665e-21 | 6.92488e-26 | 4059/4096 |
| direct-base 20261103/04 | 4.93320e-21 | 7.15608e-26 | 4059/4096 |
| direct-c4 20261105/06 | 1.11992e-20 | 9.89717e-26 | 4059/4096 |
| direct-c4 20261107/08 | 1.25714e-20 | 1.15355e-25 | 4059/4096 |

The current tolerance formula was left unchanged. The comparator returned `FAIL_PROPAGATION_DECOMPOSITION_ROUNDOFF_CHECK` / `BATCH_BLOCKED_NEEDS_WEB_REVIEW`; this summary does not infer whether that gate needs a different roundoff model or identifies a real numerical issue.

## Durable evidence and verification

- Full comparator output: [analysis.json](analysis.json), 127,788 bytes, SHA-256 `dc45035b37a50ec35d97cbe9a4d747b75ed85012a663dba7f28022e47e027243`.
- Final attempt ledger SHA-256: `6e155374b8a70995dfb2c12e1b58680e414ca1d934a1815540418b9e379e79df`; it preserves all prior attempts and records F10–F12 as validated, with 18 Fast launches / 16 actual chains and the comparator stop reason.
- All 36 archived capture files pass the task capture checksum index. Final `captures/SHA256SUMS` SHA-256: `b267019d0af79ac7d19c3e888814d97f08eebc88b0ed020fdd1da7e5d1bd2102`.
- Targeted comparator regression tests passed for F9/F13 equivalence, rejected mutations, and unique-capture accounting. All F10–F12 original and archived captures passed the existing seven-oracle / strict-capacity / provenance validator.
- F10 result/vector SHA-256: `96186cdd8a240612a156f993b8d9be53bec3326cb3f123a57b4e79466ec6c34e` / `8cd8218ea7bee1e324fa3f79a4a6c701768267a04feb9f978372b53360f5f405`.
- F11 result/vector SHA-256: `1ad1f55b137033be46dfa5cb9be2559d2ab9907decd186cb3364b52e95b6510d` / `d509b055c02c0a79e3de4ca5b08f9e51f3d3425162e1493bbd685f1b73f6eb2e`.
- F12 result/vector SHA-256: `320608ecfeb06b671572c15f6c280a67d1df7b930a0f8bf0bb3e380f8e643333` / `49ea2e0a6023277d119b0ff7a45299ff7ff8e11346d7d55de471295526d1bae5`.

See [journal.md](journal.md) for the ruling, seed/binary provenance, per-capture metrics, original failed launches, counters and stop condition. Primary has no new Standard or Secondary source changes; secondary remains clean.
