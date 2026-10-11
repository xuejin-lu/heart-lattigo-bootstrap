# FAST-LOGN16-EVALMOD-BOTTLENECK-003

**Classification:** `SOURCE_BACKED_RESCALE_HOTSPOT_CANDIDATE_NOT_CAUSAL`

## Conclusion

The captured LogN16/E32 Fast Bootstrap assigns **71.4485%** of its traced root time to the two disjoint EvalMod stages. Within each of the 28 EvalMod Rescale calls, the measured `preflight` span is the largest Rescale child; source maps it to per-component NTT-to-coefficient conversion and the coefficient-wise Q-prefix reconstruction/round/capacity loop. The most defensible next optimization hypothesis is therefore the fixed-width Q0123 coefficient loop in `fastcore.RescaleWorkspace.rescaleComponent`.

This is a source-backed hotspot candidate, **not causal proof** that this kernel explains the full EvalMod cost. The power and Rescale event trees overlap each other and the direct Bootstrap stage spans. Their inclusive durations are not additive and are not used as an end-to-end share or speedup estimate. PS baby/giant-step operations also lack individual timing in this trace.

## Provenance and run separation

| Evidence | Value | Class / limitation |
|---|---|---|
| Primary source | `aff322665f6d178f159408d66eae39a2823815f5` | Captured source pin |
| Fast production source | `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` (`fast-qprefix`) | Captured pin; current Secondary was clean at this same commit |
| Diagnostic adapter | `649da85757a97dc8c60b5f9cc031634f39ad4a40` | Test-only adapter; object is absent from current Secondary clone, recorded patch digest matches |
| Workload | LogN16, `N=65,536`, 32,768 input slots, `E=32`, `DoubleAngle=3` | Captured effective parameters |
| Q-prefix | Width 4 (`q0..q3`) through EvalMod; configured q0 target 55 bits, actual q0 56 bits | Captured parameter record, not a theoretical full-Q path |
| Config / QP / workload / input SHA-256 | `16e62f...3b02` / `6c82ca...0ec5f` / `1413c0...fa2a9` / `659fc5...ce8ccf` | Full values are in the compact Task006 evidence JSON |
| Raw trace | 539 events; 178,754 bytes; SHA-256 `9bd5ba81abce75b8fd7c06ac544410246fb385bafb1585a646d7d71e2f8e872d` | Original trace file remains labelled `TRACE_UNVERIFIED`; offline recursive analysis is separately classified `OFFLINE_RECURSIVE_TRACE_REVALIDATED` |

The raw trace is at `/var/folders/dn/6p3z5ctd50v_2dzzh2y4nvyc0000gn/T/fastdiag-public-e32-2617861684/raw-trace-unverified.json`. The maintained Batch025 recursive analyzer was reused; no new analyzer or Bootstrap was run. Because the historical adapter commit is unavailable for a fresh source-tree rebuild, event attribution is based on the captured event names/metadata and current production source at the matching Fast pin. This limitation does not change the original raw trace status.

Timing populations are intentionally kept separate:

| Observation | Time | Interpretation |
|---|---:|---|
| Task006 Fast cold Bootstrap | 12,279.909583 ms | One captured cold call; includes initialization |
| Task006 traced warm call | 1,060.782625 ms wall; Bootstrap event root 1,060.517791 ms | Single instrumented Fast call used for this attribution |
| Task005 Fast uninstrumented warm median | 663.878208 ms | Historical five-sample median; not comparable to traced Task006 |
| Task005 genuine Standard uninstrumented warm median | 6,425.275750 ms | Historical Standard reference; not a new comparison in this task |

Task006 used its recorded 2/2 Fast calls. No call budget was reset or reused. Traced Task006 latency is not presented as a speed regression against Task005.

## Measured direct Bootstrap stages

These are the nine nonoverlapping direct stage spans in the captured Bootstrap root. Percentages use the 1,060.517791 ms root; `71.4485%` is the sum of the two direct EvalMod stage shares only.

| Stage | Count | Inclusive ms | Root share | Evidence |
|---|---:|---:|---:|---|
| `coeffs_to_slots` | 1 | 152.656958 | 14.39% | MEASURED-DIRECT |
| `evalmod_imag` | 1 | 382.586417 | 36.08% | MEASURED-DIRECT |
| `evalmod_real` | 1 | 375.138083 | 35.37% | MEASURED-DIRECT |
| `mod_up_trace` | 1 | 4.629292 | 0.44% | MEASURED-DIRECT |
| `pack_n1_to_n2` | 1 | 0.085291 | 0.01% | MEASURED-DIRECT |
| `public_finalization` | 1 | 0.073000 | 0.01% | MEASURED-DIRECT |
| `scale_down` | 1 | 0.062750 | 0.01% | MEASURED-DIRECT |
| `slots_to_coeffs` | 1 | 145.065667 | 13.68% | MEASURED-DIRECT |
| `unpack_n2_to_n1` | 1 | 0.172208 | 0.02% | MEASURED-DIRECT |

Direct children total 1,060.469666 ms; signed root residual is 0.048125 ms. Amdahl ceiling if both EvalMod stages became free is **3.502448×** for this one traced call. This is only an upper bound from disjoint stage timing, not an optimization forecast.

## Measured independent power roots

| Branch | Root sequence | Inclusive ms | Direct children ms | Signed residual ms | Evidence |
|---|---:|---:|---:|---:|---|
| real | 59 | 160.157292 | 160.155917 | 0.001375 | MEASURED-INDEPENDENT-ROOT |
| imag | 279 | 163.584750 | 163.583292 | 0.001458 | MEASURED-INDEPENDENT-ROOT |

These power roots overlap the respective EvalMod stages and their child Rescale instrumentation. Do not add either to the direct stage table.

## All 35 measured Rescale roots

Each row is scope/name `rescale/rescale`, 13 events in its analyzed tree. `L` is logical Level; rows are source→target authoritative Q-prefix widths; scale values are `log2(Scale)`. Root durations are inclusive. The real/imag branches have separate sequence groups; PS branch levels need not form one serial chain.

| Group | Seq | L in→out | Rows in→out | log2 Scale in→out | ms | Evidence |
|---|---:|---:|---:|---:|---:|---|
| C2S | 6 | 16→15 | 4→4 | 106→50 | 25.697333 | MEASURED-RESCALE-ROOT |
| C2S | 19 | 15→14 | 4→4 | 106→50 | 25.699459 | MEASURED-RESCALE-ROOT |
| C2S | 32 | 14→13 | 4→4 | 106→50 | 25.637750 | MEASURED-RESCALE-ROOT |
| C2S | 45 | 13→12 | 4→4 | 106→50 | 25.624291 | MEASURED-RESCALE-ROOT |
| EvalMod real | 69 | 12→11 | 4→4 | 120→60 | 25.620000 | MEASURED-RESCALE-ROOT |
| EvalMod real | 87 | 11→10 | 4→4 | 120→60 | 26.226583 | MEASURED-RESCALE-ROOT |
| EvalMod real | 105 | 10→9 | 4→4 | 120→60 | 25.576041 | MEASURED-RESCALE-ROOT |
| EvalMod real | 123 | 9→8 | 4→4 | 120→60 | 25.634750 | MEASURED-RESCALE-ROOT |
| EvalMod real | 143 | 11→10 | 4→4 | 120→60 | 25.834083 | MEASURED-RESCALE-ROOT |
| EvalMod real | 161 | 10→9 | 4→4 | 120→60 | 25.603917 | MEASURED-RESCALE-ROOT |
| EvalMod real | 174 | 9→8 | 4→4 | 120→60 | 25.624083 | MEASURED-RESCALE-ROOT |
| EvalMod real | 187 | 11→10 | 4→4 | 120→60 | 24.061250 | MEASURED-RESCALE-ROOT |
| EvalMod real | 200 | 10→9 | 4→4 | 120→60 | 24.263167 | MEASURED-RESCALE-ROOT |
| EvalMod real | 213 | 9→8 | 4→4 | 120→60 | 25.594625 | MEASURED-RESCALE-ROOT |
| EvalMod real | 226 | 8→7 | 4→4 | 120→60 | 25.621833 | MEASURED-RESCALE-ROOT |
| EvalMod real | 239 | 7→6 | 4→4 | 120→60 | 25.562750 | MEASURED-RESCALE-ROOT |
| EvalMod real | 252 | 6→5 | 4→4 | 120→60 | 25.569000 | MEASURED-RESCALE-ROOT |
| EvalMod real | 265 | 5→4 | 4→4 | 120→60 | 25.621208 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 289 | 12→11 | 4→4 | 120→60 | 25.711625 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 307 | 11→10 | 4→4 | 120→60 | 27.033583 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 325 | 10→9 | 4→4 | 120→60 | 26.161166 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 343 | 9→8 | 4→4 | 120→60 | 26.095917 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 363 | 11→10 | 4→4 | 120→60 | 25.856833 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 381 | 10→9 | 4→4 | 120→60 | 27.091375 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 394 | 9→8 | 4→4 | 120→60 | 26.302166 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 407 | 11→10 | 4→4 | 120→60 | 24.624583 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 420 | 10→9 | 4→4 | 120→60 | 25.194584 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 433 | 9→8 | 4→4 | 120→60 | 26.215500 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 446 | 8→7 | 4→4 | 120→60 | 25.790041 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 459 | 7→6 | 4→4 | 120→60 | 25.695250 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 472 | 6→5 | 4→4 | 120→60 | 25.812042 | MEASURED-RESCALE-ROOT |
| EvalMod imag | 485 | 5→4 | 4→4 | 120→60 | 25.708333 | MEASURED-RESCALE-ROOT |
| S2C | 499 | 4→3 | 4→4 | 84→45 | 25.312209 | MEASURED-RESCALE-ROOT |
| S2C | 512 | 3→2 | 4→3 | 84→45 | 24.749750 | MEASURED-RESCALE-ROOT |
| S2C | 525 | 2→1 | 3→2 | 84→45 | 20.162458 | MEASURED-RESCALE-ROOT |

| Group | Count | Median ms | Range ms | Evidence |
|---|---:|---:|---:|---|
| C2S | 4 | 25.667542 | 25.624291–25.699459 | MEASURED-DISTRIBUTION |
| EvalMod real | 14 | 25.611959 | 24.061250–26.226583 | MEASURED-DISTRIBUTION |
| EvalMod imag | 14 | 25.834438 | 24.624583–27.091375 | MEASURED-DISTRIBUTION |
| S2C | 3 | 24.749750 | 20.162458–25.312209 | MEASURED-DISTRIBUTION |
| All roots | 35 | 25.634750 | 20.162458–27.091375 | MEASURED-DISTRIBUTION |

## Source-to-event attribution

Source pointers refer to clean Secondary Fast `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`; source correspondence is **INFERRED from the captured names/metadata and matching production pin**, not a fresh rebuild of the unavailable diagnostic adapter.

| Existing event(s) | Source path / function | Observed operand / transition | Attribution |
|---|---|---|---|
| `stage/evalmod_real`, `stage/evalmod_imag` | `circuits/ckks/bootstrapping/fast_bootstrap.go`: `Bootstrap` stage wrappers; `circuits/ckks/bootstrapping/fast_evalmod.go`: `FastEvaluator.EvalMod` | Each stage L12→4, four authoritative rows; stage Scale log2 50→45. EvalMod temporarily interprets Scale at log2 60. | INFERRED source mapping; stage duration is directly measured. Real and imag stage spans are separate; neither is a child of the other. |
| `power/generated_powers`, `power/power`, `power/mul_relin`, `power/chebyshev_doubling`, `power/recurrence_correction`, `power/rescale` | `circuits/ckks/mod1/fast.go`: `evaluateNew`; `circuits/ckks/polynomial/fast.go`: `EvaluateQPrefixRows`, `generatePowers`, `fastPowerBasis.genPowerInternal`, `evaluatePlan` | Chebyshev generated powers use ciphertext multiply/relinearize, doubling and recurrence subtraction, then Q-prefix Rescale. PS baby steps and giant-step combines use shared Q-prefix evaluator calls. Level/Scale vary by branch; the trace records 12 each for power subevents and two branch roots. | Event categories support operation families, but no per-PS-operation event duration exists. Recursive power spans are inclusive/nested. |
| `rescale/rescale`, `rescale/preflight`, `rescale/prefix_to_coefficient`, `rescale/coefficient_loop`, `rescale/reconstruct_center_round_capacity`, `rescale/residue_materialization`, `rescale/materialization`, `rescale/ntt_montgomery_restore` | `schemes/ckks/fast/rescale.go`: `RescaleQPrefixRows`; `schemes/ckks/internal/fastcore/rescale.go`: `RescaleWorkspace.ApplyRows`, `rescaleComponent`, `prefixToCoefficientRows`, `reconstructQPrefix`, `crtQ0123Prepared`, `roundedMagnitude192`; `ring/subring_ops.go`: `SubRing.INTT`/`NTT` | EvalMod has 28 roots, each L transition by one, four source/target rows, Scale log2≈120→60. Each root has two ciphertext components and N=65,536 coefficients. | Root and child timings measured; operation correspondence inferred from source. `preflight` wraps component conversion and coefficient processing; child timing records overlap that parent span. |

The Rescale child distributions (all are nested under their parent root; per-component records may be nested under a loop) are:

| Event | Count | Median ms | Range ms | Source interpretation |
|---|---:|---:|---:|---|
| `preflight` | 35 | 22.763208 | 18.693250–24.059250 | Two components' Rescale processing wrapper |
| `prefix_to_coefficient` | 70 | 1.433230 | 1.071000–1.598750 | Four row-wise INTTs per component at width four; optional IMForm is selected by input representation |
| `coefficient_loop` | 70 | 9.030459 | 7.364125–12.203667 | One loop over N coefficients for each component |
| `reconstruct_center_round_capacity` | 70 | 2.915342 | 1.355617–5.928706 | Timed CRT/centering/round/capacity work within the coefficient loop |
| `residue_materialization` | 70 | 1.489280 | 1.006803–1.792862 | Signed output residue generation within the same loop |
| `materialization` | 35 | 2.744083 | 1.368041–3.517333 | Output staging and representation/metadata materialization |
| `ntt_montgomery_restore` | 70 | 1.372959 | 0.682083–1.765250 | Target-row NTT plus MForm when needed, per component |

Medians from distinct nested rows must not be added. `preflight` is a broad measured parent, while the CRT and residue events are instrumented portions of its child coefficient loops. This trace includes no separate timing for individual PS multiplies/adds or for each CRT instruction.

At this workload, each of the 28 EvalMod Rescale calls processes two components × 65,536 coefficients. The source loop gathers four authoritative residues, reconstructs/centers a fixed-width 192-bit value, performs the logical rounded division and target-capacity check, and materializes up to four target residues: **131,072 coefficient iterations per call** (3,670,016 across those 28 calls). This is an operation-count model, not a timing sum.

## Prior art and one prioritized hypothesis

Batch026's LogN13 synthetic modulo candidate did not show a robust speedup and its historical fixture recipe is unrecoverable. Batch027 restored a reproducible LogN13/E32 synthetic Level9, four-row public Rescale fixture and exact big.Int oracle; its median public Rescale was 2.719800 ms. Neither is a LogN16 Bootstrap ciphertext or a direct performance comparison to this trace. This report does not repeat either benchmark.

### H1 — reduce fixed-width Q0123 work per coefficient (HYPOTHESIS; recommend for a later authorized task)

- **Affected scope:** 28 observed Rescale roots within the two EvalMod stages. Their root distributions bound per-call cost, but trace-tree overlap prevents attributing a reliable fraction of the 71.4485% EvalMod wall time to this kernel. No end-to-end gain is forecast.
- **Cost model:** O(components × N × Q-prefix width) residue access plus fixed-width reconstruction/round/capacity per coefficient. Here that is two components, N=65,536, and four source rows per call. Source supports a specialized four-row arithmetic path already; any optimization must be proven against that exact Q0123 path, not a Q0/Q1 model or full-RNS fallback.
- **Required invariants:** exact centered CRT and Standard-compatible rounded division; every intermediate target-capacity check; only the authorized q0..q3 rows are read; unchanged logical Level/Scale, NTT/Montgomery representation and compact backing; transactional output on failure; no changes to retained Fast zero-secret semantics.
- **Minimal focused falsification gate for a future implementation:** run existing `TestFastRescaleMatchesBigIntCenteredCRTOracleAtEveryPrefixWidth`, `TestFastRescaleMatchesStandardGeneratedLogN16`, `TestRescaleCapacityFailureAfterEarlierComponentStagingIsTransactional`, and Q0123 Rescale differential coverage. Then compare finite samples using the existing `BenchmarkFastRescaleLogN16` on a source-bound, valid two-component/four-row input with matching NTT/Montgomery and Level/Scale. Reject the hypothesis if the oracle/capacity/atomicity gates fail or the repeated benchmark shows no robust improvement. A benchmark win alone does not prove EvalMod or Bootstrap benefit; that requires a separately authorized numerical/end-to-end validation.
- **Pitfalls:** a fast path can accidentally read dormant Q rows, skip a rounding/capacity boundary, alter source authority, or benchmark a synthetic shape unlike the measured path. The old Batch026 result specifically warns against accepting a single noisy microbenchmark win.

No other optimization is prioritized: PS evaluation remains an unattributed source-level workload because its individual operations were not timed. Resolving that uncertainty would require a separately scoped instrumentation/measurement decision, not new timers in this task.

## Work performed and validation

- Reused committed Task006 compact evidence and the Batch025 recursive/root-partition analyzer; no independent runner or stats implementation was added.
- Inspected the 178,754-byte raw trace by hash and targeted event selections only; no full trace/vector dump.
- Secondary remained read-only at clean `fast-qprefix` commit `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`.
- No Bootstrap, benchmark, CKKS production change, test-only adapter rebuild, or new measurement was run.
- Report validation: raw/evidence SHA and event-count checks; source symbol/path checks; `git diff --check`. No Go tests were run because this is a report-only source audit and Go tests may execute unrelated experiment paths.

## Evidence links

- `results/CKKS-MEASUREMENT-PLATFORM-ACCEPTANCE-REPAIR-002-task006-summary.md`
- `results/CKKS-MEASUREMENT-PLATFORM-ACCEPTANCE-REPAIR-002-task006-evidence.json`
- `results/FAST-LOGN16-INTEGRATED-PERFPROBE-FASTDIAG-006.md`
- `results/FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.md`
- `results/FAST-E32-RESCALE-COEFFICIENT-OPT-AUTONOMOUS-BATCH-026-summary.md`
- `results/FAST-E32-RESCALE-CAUSAL-HOTSPOT-AUTONOMOUS-BATCH-027-summary.md`
