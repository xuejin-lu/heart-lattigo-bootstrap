# Independent Web review: Batch025 E32 Fast offline trace

**Date:** 2026-10-11, Asia/Taipei  
**Verdict:** **ACCEPT `OFFLINE_REVALIDATED_TRACE`** as a separately labeled **post-hoc event/provenance analysis**, not as a passed original live trace or a new speedup comparison. Do not rewrite the original `TRACE_UNVERIFIED` file. No further Batch023/024 Bootstrap calls are authorized.

## Evidence independently examined

- Primary live remote `main@4827ba99d57ffeae41393a21c1fecda10ca3b1bb`; Secondary `fast-qprefix@4f2557062cb5c1ffb9a671bc3df67401fd7092b4`. Batch025 implementation `75a8563d7371970141bb6705f0e7132d8aab610f` changed existing `cmd/fastdiag` offline mode/tests, not Secondary arithmetic.
- Read the published Batch025 summary, journal, full compact evidence.json and offline replay source. Recorded immutable source raw SHA-256 `f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74` (539 events, 177696 bytes), sidecar digest, original fixture/vector SHA, original diagnostic source and 2/2 spent reservations. The replay SHA-checks original on-disk evidence, requires `TRACE_UNVERIFIED` as source, checks source/fixture/recorded output hashes, validates recursive power ancestry and tree closure, and writes a separately labeled derived artifact. The exact original Mac-local raw bytes are **not independently accessible to this Web reviewer**: acceptance is of corroborated source, submitted replay execution and published evidence, not an independent reproduction over local raw.
- Output checks are of **recorded** two Bootstrap runs, not a new decrypt. Both originally passed 1e-6 complex-max gate (RMSE `1.1795785978778852e-9`, maximum `5.738790672737443e-8`), with decoded hashes matching archived six-output Fast vectors. Batch025 spent **0 new Bootstrap**; prior Batch023 14/14 and Batch024 2/2 immutably spent.
- CPU pprof file available but **not cryptographically bound by source raw metadata**; intentionally excluded. Heap profile unavailable. Single traced warm run is not a statistical distribution, and adding per-coefficient trace timers changes runtime overhead. Historical **uninstrumented** Batch023 five warm Fast median 72.580 ms, Standard 306.636 ms (observed **4.225x**) is a **separate** experiment for intentionally insecure zero-secret Fast; do not compare those times as if identical overhead or security.

## Source-backed E32 diagnosis: rooted, non-overlapping sums only

**One traced warm Bootstrap** (not production time): `122.193458 ms`, root children `122.182875 ms`, root unexplained residual `0.010583 ms`; no reported negative exclusive event residual. Nine real stages each once. Stage **inclusive** values, all percentages relative to this same single Bootstrap root:
- `evalmod_real` **45.968333 ms / 37.6193%**
- `evalmod_imag` **45.889625 ms / 37.5549%**
- `coeffs_to_slots` **16.469166 ms / 13.4779%**
- `slots_to_coeffs` **13.362875 ms / 10.9358%**
- all other stage spans and root gap account for the balance (~0.5035 ms).
Together EvalMod real+imag **91.857958 ms / 75.1742%**. This supports **EvalMod as optimization target**, not yet one particular source kernel as the exclusive cause.

The collector also creates **two independent generated-power roots** (summed independent-root wall spans **39.056333 ms**) and **35 independent Rescale roots** (sum of their inclusive spans **108.326291 ms**). These roots can lie inside/overlap the stage and power spans, and recorded events lack sufficient globally timestamped interval evidence to establish a partition of Bootstrap total. **Never add their root sums to 122.193458 ms or present 108.326291 / 122.193458 as a proven Bootstrap Rescale share.**

Within the **35 Rescale roots alone**, the existing per-event **exclusive** times partition those 108.326291 ms without double-counting nested event descendants. Sum across the 35 roots:
- `reconstruct_center_round_capacity` c0+c1 **34.285478 ms = 31.65%**;
- `coefficient_loop` remaining own exclusive c0+c1 **38.647753 ms = 35.68%** (not a separate full-loop inclusive charge);
- `residue_materialization` c0+c1 **14.945317 ms = 13.80%**;
- `prefix_to_coefficient` c0+c1 **10.174914 ms = 9.39%**;
- `ntt_montgomery_restore` c0+c1 **9.567417 ms = 8.83%**;
- remaining root/preflight/materialization own exclusive ~0.7054 ms = 0.65%.
Thus coefficient work, CRT rounding/capacity and residue production are more promising than NTT restoration in this E32 **diagnostic trace**. This is **prioritization**, not a proven speedup of an optimization.

## Actual source hypothesis and proper next experiment

Secondary `schemes/ckks/internal/fastcore/rescale.go`, `rescaleComponent` around lines 369–436: per N coefficient `reconstructQPrefix`, `centeredQPrefix`, `roundedMagnitude192`, `cmp192` capacity, then `signedResidue192` per target row. `crtQ01`, `crtQ012Prepared`, `crtQ0123Prepared`, `mod192By64`, `roundedMagnitude192` use fixed-width arithmetic and multiple `bits.Div64`. Historical P93 source diagnostics also pointed to CRT/fixed-width operations, but **the P93 percentages must not be substituted for E32**. Accurate next direction: low-cost, strict baseline-versus-candidate **real Rescale microbenchmark** on the E32 representative row/level/configuration with a fixed correctness oracle, before assuming a particular Div64 or CRT rewrite is a winner.

## Web disposition

Batch025 scientific goal completed within evidence boundaries, final classification `BATCH_COMPLETE_ACCEPTED_WITH_LIMITATIONS`. No need to repair original Standard stale/prunable checkout to inspect Fast trace; do not prune it. Activate Batch026 as a **focused E32 Rescale hotspot reproduction and a single bounded optimization hypothesis**, with **zero new Bootstrap**, no Standard or historical Fast pins mutated, no non-matching parameter sweeps, and only source-constrained cheap local benchmarks. Do not make a general library speedup claim without a separate duly authorized representative full-workload comparison.
