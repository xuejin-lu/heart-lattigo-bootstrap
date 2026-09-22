# FIX-001-P3-MEASURE-LOGN13-P93-SCALE-RUNTIME-FACTS

## Purpose

This task is **measurement-only**.

The static and mathematical scale audit has already been performed by the orchestrator and recorded in:

- `docs/FIX-001-P3-LOGN13-P93-SCALE-STATIC-AUDIT.md`
- `results/FIX-001-P3-LOGN13-P93-SCALE-STATIC-AUDIT-summary.json`

Current static classification:

`STATIC_SCALE_AUDIT_NO_CAUSAL_DEFECT_FOUND_RUNTIME_ROUNDING_CHECKS_REQUIRED`

Codex must **not** decide whether a scale contract is correct or incorrect.

Codex must only measure the exact runtime quantities requested below and report them compactly.

No Secondary modification is authorized.

---

## Repository state

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- exact committed SHA:
  `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- worktree must remain clean

Frozen Fast exact E2E baseline:

`0.009705381393898434`

Do not modify Secondary.

---

# M0 — baseline confirmation

Run the finalized deterministic LogN13/P93 bootstrap.

Record only:

- Fast exact E2E
- Genuine Standard exact E2E
- final Fast Scale
- final Standard Scale

Require Fast <= `1e-2`.

---

# M1 — ScaleDown rounding measurement

At the exact executed ScaleDown scale-up boundary record:

- `currentMessageRatio`
- `targetMessageRatio`
- exact 128-bit `scaleUp`
- `scaleUp.BigInt()`
- absolute mismatch:
  [
  |operatorname{round}(scaleUp)-scaleUp|
  ]
- relative mismatch:
  [
  left|rac{operatorname{round}(scaleUp)}{scaleUp}-1ight|
  ]
- semantic before/after max-component residual for the physical+metadata scaling step only

Do not interpret the result.

---

# M2 — ModUp Float64/round measurement

At the exact executed ModUp message-scale lift record:

- `ScalingFactor()` exact value
- `MessageRatio()`
- input ciphertext Scale exact value
- ratio computed from exact 128-bit values if reconstructed in Primary
- ratio after the source's `Float64()` conversions
- `scalar = round(float64_ratio)`
- absolute:
  [
  |scalar-r|
  ]
- relative:
  [
  |scalar/r-1|
  ]
- semantic max-component before/after residual for the physical scalar + metadata update
- Standard path values for the same deterministic input

The Standard measurement is required because the source uses the same contract.

Do not classify it.

---

# M3 — PS scale-alignment ratio measurements

Instrument the actual P93 polynomial run in Primary diagnostic code only.

For every executed call corresponding to Fast polynomial:

- `subAligned`
- `addAligned`
- any equivalent scale-alignment helper on the active path

record one compact row containing:

- operation id
- branch real/imag
- caller checkpoint
- numerator Scale
- denominator Scale
- exact ratio
- `ratio.BigInt()`
- relative rounding mismatch:
  [
  |operatorname{round}(ratio)/ratio-1|
  ]
- semantic residual introduced by the alignment operation if directly measurable
- whether ratio is exactly integral

Return:
- total count
- max relative mismatch
- worst operation id
- max measured semantic residual

Do not dump all slot vectors.

---

# M4 — PS InDelta metadata-snap measurements

For every executed P93 path where source logic:

1. checks `Scale.InDelta(...)`;
2. then assigns one ciphertext Scale to another without coefficient change,

record:

- operation/checkpoint
- branch
- Scale A
- Scale B
- ratio A/B
- `Log2Delta`
- relative difference
- implied interpretation shift magnitude
- semantic before/after difference caused by metadata-only snap, if directly measurable

Return the worst observed snap.

Do not interpret it.

---

# M5 — normalized Mod1 scale ledger

For real and imag, record exact runtime values:

- EvalMod input Scale
- `ScalingFactor()`
- `targetScale`
- `planScale`
- polynomial output Scale
- `kIn`
- `coherentScale`
- coherent/target ratio
- coherent/target `Log2Delta`
- for each DA round:
  - input Scale
  - current virtual exponent
  - computed `nextScale`
  - workingScale
  - nextScale/workingScale ratio
  - `nextExponent`
  - `aExponent`
  - factor `2^a`
  - post-Rescale Scale
- final pre-materialization Scale
- final virtual exponent
- physical materialization factor
- Scale after caller-input reset
- Scale after public DefaultScale reset

No correctness judgment.

---

# M6 — C2S compressed restore measurements

For the actual C2S restore plan:

- group index
- matrix compression exponent (k)
- encoded matrix Scale before compression if available
- encoded compressed matrix Scale
- post-group Rescale Scale
- physical restore scalar (2^k)
- metadata restore scalar (2^k)
- resulting Scale
- semantic before/after restore residual

Expected restore plan context:
- group 0: (k=4)
- group 1: (k=2)
- remaining groups: (k=0)

Do not classify.

---

# M7 — output

Create:

`results/FIX-001-P3-MEASURE-LOGN13-P93-SCALE-RUNTIME-FACTS-summary.json`

Keep <=350 pretty-printed JSON lines.

Required fields:

- provenance
- M0 baseline
- M1 ScaleDown
- M2 ModUp
- M3 PS alignment summary + compact rows
- M4 metadata snap summary + compact rows
- M5 Mod1 ledger
- M6 C2S restore ledger
- Secondary clean-state confirmation

No classification field beyond:

`MEASUREMENT_COMPLETE`

Do not write recommendations.

---

# Validation

Run:

- focused measurement tests
- finalized exact-E2E replay
- Primary `go test ./...`
- `git diff --check`

Secondary must remain:
- SHA `40532b4d...`
- clean

Commit/push Primary measurement code + compact result only.

---

# Prohibitions

- no Secondary modification
- no scale fix
- no parameter tuning
- no correctness classification
- no threshold change
- no new precision work

This task measures facts only.
