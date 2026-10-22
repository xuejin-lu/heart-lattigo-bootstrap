# GPT Web scientific review — FIXED-SIGMA-PROPAGATION-003

**Reviewed:** 2026-10-22 (Asia/Taipei)
**Disposition:** `ACCEPTED_SCOPED_PROPAGATION_ABLATION` / `INSUFFICIENT_EVIDENCE` for frozen-σ transferability. This is not an acceptance of distributional, security, or Bootstrap equivalence.
**Reviewed Primary:** `574a1aee5ce5b0e329096095de3eda1b2df7bd1b` (`main`).
**Reviewed Secondary:** `1e7828293c36f9845e30723ba4e9500d84bb50d2` (`fast-qprefix`).
**Genuine Standard reference:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (unaltered original).

## Independent verification

- Remote Primary and Secondary branch HEADs and Primary task diff verified. Primary advanced three commits: CP-B ablation harness, fingerprint representation repair, and bounded JSON/journal/summary. Secondary code unchanged.
- Inspected maintained `cmd/perfprobe/fixed_sigma_ablation.go`, `fixed_sigma_chain.go`, and journal/summary. Four Fast ablation arms `OFF/OFF`, `ON/OFF`, `OFF/ON`, `ON/ON` are selected by explicit helper calls only after complete `MulRelinNew→Rescale`; Stage2 consumes the actual Stage1 ciphertext. Both operands and C q4 scale are provenance-checked. No global Rescale interception.
- Two deterministic workload variants (legacy canonical C, C×4), two seed pairs each: 16 Fast chains; 2 genuine Standard references; 0 Bootstrap. Matching Fast arm inputs/row hashes, actual helper decoded deltas, injected coefficient sigma and Q-prefix strict capacity checks are reported as passed; `compare-public-ablation` returned `PASS_VALIDATED_DIAGNOSTIC_ANALYSIS`. Reviewed bounded `cp-b-analysis.json` metadata, source and compact numerical summaries, not all local raw vector files.
- Frozen `σ=4.004280012097432` unchanged. Stage2 OFF/OFF total RMS ~7.39e-13 (canonical), ~7.58e-13 (C×4); ON/OFF ~7.69–7.72e-13, ~1.13–1.14e-12. P1 RMS ~6.60–6.62e-13 and ~1.326–1.327e-12 respectively. OFF/ON and ON/ON final errors ~1e-11, with matched genuine Standard ~1e-11. These are descriptive data from only two workloads/two seed pairs, and spatial slots are not independent runs.
- CP-C deliberately did not alter the capacity checker; CP-D source-only Bootstrap boundary audit found ct×ct, ct×pt and delayed/lazy Rescale semantics inconsistent with a single safe global hook. No Bootstrap was run. The incidental LogN16 parameter-construction test was previously reviewed and logged as `NON_BLOCKING_INCIDENTAL_TEST_SCOPE`; this does not authorize new LogN16 workload runs.

## The mathematical propagation attribution to preserve

Let `x0` and `x1` denote the decoded Stage1 outputs with and without first injection; `delta1=x1-x0`. Let `c_tilde` be the decoded Stage2 input ciphertext (identical across matched arms). Define `r2^a=decoded(Stage2^a)-decoded(Stage1^a)*c_tilde`, for each arm `a`. Then for the Stage2 output ablation difference,

`P1 = decoded(Stage2_ON/OFF) - decoded(Stage2_OFF/OFF) = delta1*c_tilde + (r2_ON/OFF-r2_OFF/OFF)`.

This identity is algebraic under the matched second operand and identical slot alignment, **not** a claim that the increment difference has been attributed to a specific underlying CKKS primitive. In the current runs `RMS(P1)` exceeds `RMS(delta1*c_tilde)` and the remaining term is material. Correlation between the terms means their RMS values **must not be added**. Existing source records vectors/hashes and sufficient per-step decoded data to test the identity in a future targeted analysis if retained raw vectors remain accessible; this review did not independently reconstruct all slot vectors.

The near-zero `ON/ON−OFF/ON−P1` check demonstrates a consistent paired result at these inputs, but does not establish generic linearity, higher-depth behavior, tail or Gaussian equivalence.

## New experiment decision

Prior AddNew(A,B) is a **historical fixture artifact** of the accepted E32 public lifecycle `Encrypt→Add→MulRelin→Rescale→Rotate→Bootstrap`. It was retained for repeatability, not because Add is required by the new-noise scientific hypothesis. Historical 001–003 data/fixtures MUST NOT be overwritten.

The user approved a **new, explicitly versioned** pure ct×ct `A×B→(A×B)×C` task. Proceed under `specs/FIXED-SIGMA-DIRECT-MUL-004.md`, keeping frozen sigma and its previous fit untouched, while building an independent fingerprint and capacity proof. No new mathematical commitment to unconditional sigma generality is implied.
