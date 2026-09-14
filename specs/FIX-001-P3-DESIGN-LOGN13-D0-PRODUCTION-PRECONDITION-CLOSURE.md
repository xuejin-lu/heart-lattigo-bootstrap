# FIX-001-P3-DESIGN-LOGN13-D0-PRODUCTION-PRECONDITION-CLOSURE

## Goal

Before modifying Secondary production code, determine the **minimal production precondition set** required to transfer the now-validated LogN13 D0 arithmetic stack into the real fixed frontend configuration.

The central question is:

> With the actual checked-in LogN13 frontend configuration (`q0 = 55`), does the complete validated arithmetic stack already pass once all accepted fixes are present together, or is `planScale = 2^92` and/or `q0 = 56` genuinely required?

This is a Primary-only diagnostic/design-closure task.

**Do not modify Secondary production code in this task.**

---

# Why this task is required

The latest successful diagnostic result is:

Primary commit:

`882721ee3f9752fc0dc44ecdc4b1d7b4bab5947b`

Classification:

`logn13_generated_power_native_q012_multiply_first_system_sufficient`

It proves that the complete accepted D0 diagnostic stack can achieve:

`public_like = 0.005554220603853743 <= 1e-2`

with generated powers implemented through a temporary authoritative q0/q1/q2 multiply-first schedule.

However, that successful stack was evaluated with diagnostic preconditions including:

- q0 profile bit size = 56;
- common polynomial planScale = `2^92`;
- accepted G0 local-q2 guard = 2;
- accepted F0 local-q2 guard = 3;
- all accepted DoubleAngle local-q2 rounds;
- accepted final-parent T2 one-bit scalar guard;
- generated-power temporary-q2 multiply-first.

The actual checked-in frontend configuration remains:

`configs/bootstrap_config.logN13.json`

with:

`"q0": [55]`

and the current Secondary production normalized LogN13 path still has an internal plan scale of `2^91` and the historical balanced generated-power schedule.

Therefore a direct single-feature production integration would repeat the earlier integration mistake: moving one validated component into a production stack whose other numeric preconditions differ.

This task isolates those remaining preconditions **before** any Secondary patch.

---

# Required provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required ancestor:

`882721ee3f9752fc0dc44ecdc4b1d7b4bab5947b`

Start from clean `main`, fetch + ff-only pull, then re-read fresh:

- `AGENTS.md`
- `CURRENT_TASK.md`
- this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required exact commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Required branch: `fast-ckks`

Required state: clean.

Secondary is read-only in this task.

---

# Fixed frontend/workload rule

The checked-in production-facing LogN13 configuration is the architectural baseline:

`configs/bootstrap_config.logN13.json`

In particular:

`q0 = 55`

Do **not** edit that file in this task.

The q0=56 cases below are diagnostic clones only. They exist solely to isolate whether the 56-bit q0 is a true numeric requirement.

Keep fixed across all matrix cases:

- LogN13;
- logDefaultScale = 45;
- same Mod1 degree/type/K/logMessageRatio;
- same DFT level structure;
- same P configuration;
- same deterministic workload/input;
- same polynomial coefficients and PS split;
- same Standard-reference methodology;
- same downstream S2C/finalization semantics;
- no frontend Fast/Standard selector.

---

# Accepted arithmetic stack — must be identical in all matrix cases

Every matrix case must use the **same complete accepted arithmetic stack**. The only allowed matrix variables are q0 bit size and polynomial planScale.

Required arithmetic fixes:

1. accepted C2S compressed groups already established by FIX-001;
2. G0 local-q2 guard = 2;
3. F0 local-q2 guard = 3;
4. all accepted DoubleAngle local-q2 rounds;
5. final-parent T2 one-bit scalar guard;
6. generated-power temporary authoritative q0/q1/q2 multiply-first schedule:
   - exact q01 parent recovery while parent Q01 is unique;
   - exact lift to q2;
   - q0/q1/q2-authoritative multiply/square;
   - Chebyshev doubling;
   - recurrence correction **before** Rescale;
   - one centered symmetric rounded Rescale by the actual logical current-level divisor `Q[L]`;
   - Q012 must remain centered-unique before rounding;
   - post-Rescale Q01 must be centered-unique before q2 contraction/drop;
   - q3+ remain dormant and must not be read as arithmetic truth.

Do not selectively remove any accepted guard from one matrix case.

This is a precondition-isolation experiment, not another arithmetic-component attribution task.

---

# P0 — source-backed current-production control

Record the current production source state at Secondary `7d05f1...`:

- normalized LogN13 planScale bits currently used by production;
- current generated-power schedule;
- current q0/q1-only Fast Rescale contract;
- current checked-in Primary q0 value.

Run one current-production-like control using the checked-in LogN13 config and current Secondary behavior, without any diagnostic overrides.

Record:

- effective full LogQ chain;
- polynomial/EvalMod/post-S2C/public-like metrics;
- Level/Scale checkpoints relevant to comparison.

This control is informational. Do not require a specific historical public-like number beyond source-consistent reproduction.

---

# P1 — accepted D0 control

Reproduce the accepted latest D0 diagnostic control:

- q0 diagnostic bit size = 56;
- planScale = `2^92`;
- complete accepted arithmetic stack listed above.

Expected public-like:

`0.005554220603853743`

Expected classification-equivalent behavior:

- all generated powers use genuine temporary-q2 multiply-first arithmetic;
- all q012 rounded Rescale proofs pass;
- all post-Rescale Q01 contractions pass;
- EvalMod real/imag remain near the accepted values;
- public-like passes `1e-2`.

If this control does not reproduce, stop:

`logn13_d0_production_precondition_control_mismatch`.

---

# P2 — 2×2 precondition matrix

Run exactly these four complete-stack cases:

| Case | q0 bit size | planScale | Meaning |
|---|---:|---:|---|
| A | 55 | `2^91` | actual frontend q0 + current production internal plan scale |
| B | 55 | `2^92` | actual frontend q0 + accepted diagnostic plan scale |
| C | 56 | `2^91` | diagnostic q0 + current production internal plan scale |
| D | 56 | `2^92` | accepted D0 control |

### Important

For A/B, build the profile from the actual checked-in LogN13 configuration. Do not silently route through a helper that forces q0=56.

For C/D, clone the config in memory and change only q0 bit size for diagnostic construction. Do not edit the checked-in config file.

For every case record the **actual generated**:

- q0 modulus and bit length;
- q1 modulus and bit length;
- q2 modulus and bit length;
- complete bootstrap `LogQ` list;
- logical Mod1 start level;
- actual divisor modulus used for every generated-power Rescale;
- actual planScale.

Do not assume the downstream generated prime bit sizes are identical merely because the requested bit-size vectors look similar.

---

# P3 — required metrics for every matrix case

For each A/B/C/D record:

## Generated powers

For T2/T3/T4/T6/T8/T16, real and imag:

- completed local residual vs canonical;
- q012 maximum intended physical capacity ratio before Rescale;
- minimum Q012 centered-uniqueness margin;
- maximum intended physical ratio against Q01;
- post-Rescale Q01 ratio;
- q012 rounded quotient agreement;
- post-Rescale q01 contraction validity;
- output Level/Scale/degree/NTT/Montgomery contracts.

## Polynomial / EvalMod / downstream

Record:

- polynomial residual;
- EvalMod real max component;
- EvalMod imag max component;
- post-S2C max component;
- public-like max component;
- public margin `1e-2 - public_like`;
- all accepted guard-selection counts/contracts;
- final Level/Scale metadata.

System pass condition:

`public_like <= 1e-2`.

Do not substitute local thresholds for the system criterion.

---

# P4 — isolate q0 effect from planScale effect

Use the 2×2 matrix to answer these questions explicitly:

1. Does A pass?
   - If yes, neither q0=56 nor planScale92 is required for correctness once the **full arithmetic stack** is present.
2. If A fails, does B pass?
   - If yes, the checked-in q0=55 frontend config can remain unchanged; only the backend-internal LogN13 planScale must move from 91 to 92 together with the arithmetic fixes.
3. If both A and B fail but a q0=56 case passes, does evidence show a genuine q0 width dependency?
   - If yes, production integration cannot proceed under the fixed frontend config without an explicit architecture/configuration decision.
4. Is there an interaction where only one of C/D passes or where q0 and planScale effects are non-monotonic?
   - If yes, report the interaction rather than simplifying it to a single-factor claim.

For failing cases, identify the **first system-relevant divergence** among:

- generated-power q012 capacity;
- post-Rescale q01 contraction;
- generated-power local residual;
- PS/polynomial residual;
- EvalMod amplification;
- S2C amplification;
- public finalization.

Do not reopen already excluded causes without new evidence.

---

# P5 — production-readiness decision

Produce one explicit production-readiness decision from the matrix.

## Outcome A — fixed config/current plan scale sufficient

If Case A passes:

`production_ready_fixed_q055_plan91`

Next task may production-integrate the complete arithmetic stack in Secondary without any Primary frontend config change.

## Outcome B — fixed config + planScale92 sufficient

If A fails and B passes:

`production_ready_fixed_q055_requires_plan92`

Next task may production-integrate the complete arithmetic stack plus the narrow internal normalized-LogN13 planScale change to 92. Primary frontend config remains unchanged.

## Outcome C — q0=56 required

If A and B fail while C and/or D passes because of a source-backed q0-dependent numeric/capacity condition:

`production_blocked_by_q0_width_precondition`

Do **not** create or execute a production patch that edits the frontend config. Stop at the architecture boundary and report the exact evidence.

## Outcome D — q0/plan interaction

If the matrix does not support a one-factor conclusion:

`production_blocked_by_q0_plan_interaction`

Localize the interaction and stop.

## Outcome E — accepted control or complete stack mismatch

If D itself fails or the full arithmetic stack cannot be reproduced consistently:

`production_blocked_by_d0_precondition_mismatch`

---

# P6 — fix the latest artifact comparison bookkeeping bug

The previous q012 feasibility artifact contains one known diagnostic bookkeeping defect:

`local_power_comparisons[*].post_rescale_q01_rows_equal` was reported `false` even though:

- q012 Rescale proof rows passed;
- post-Rescale q01 contraction rows passed;
- native-vs-full decoded difference was zero.

Source inspection shows the local comparison passed an already-Fast/Montgomery candidate through a helper that applies `MForm` while contracting a full-RNS source, effectively double-MForming the comparison operand.

In this task:

- do not reuse that invalid comparison pattern;
- distinguish full-RNS source -> Fast contraction from Fast-to-Fast row comparison;
- require correct post-Rescale q0/q1 row comparison in the matrix artifact;
- add a focused regression assertion that would fail if an already-Montgomery Fast ciphertext is accidentally MForm-converted again.

This bookkeeping correction must not change arithmetic results.

---

# P7 — no production implementation yet

Even if A or B passes, do not modify Secondary in this task.

Instead, emit a compact `production_patch_contract` section specifying exactly what the next Secondary task is authorized to change.

If A passes, the anticipated patch contract should be limited to the accepted arithmetic fixes, especially:

- production temporary-q2 generated-power multiply-first;
- source-backed G0/F0/DoubleAngle local-q2 primitives already required by D0;
- accepted final-parent T2 scalar guard activation;
- removal/bypass of the balanced generated-power schedule for the validated bounded LogN13 path;
- no frontend config change.

If B is the first passing fixed-q0 case, additionally authorize:

- normalized LogN13 internal planScale 91 -> 92.

The next task must still choose source-backed production primitives rather than copying Primary diagnostic `big.Int` helpers into a hot path.

---

# Production primitive constraint for future integration

The latest q012 diagnostic proves mathematics/system sufficiency but its rounded Rescale producer uses big.Int reconstruction as a diagnostic implementation.

Do not interpret that as permission to productionize per-coefficient big.Int arithmetic.

Future Secondary implementation must preserve Fast hot-path principles:

- q0/q1/q2 only at the deliberate temporary boundary;
- q3+ dormant;
- no Standard full-RNS fallback;
- no stale high-limb reads;
- no GadgetProduct/standard key switch fallback;
- no routine heap-allocated per-coefficient big.Int reconstruction.

A likely production direction is a fixed-width 3-word CRT/centered division implementation using `math/bits`, validated for the actual q0/q1/q2 bit bounds. This task does not implement it.

---

# Classification

Choose exactly one primary classification:

- `logn13_d0_fixed_q055_plan91_full_stack_sufficient`
- `logn13_d0_fixed_q055_requires_plan92`
- `logn13_d0_q0_56_width_precondition`
- `logn13_d0_q0_plan_interaction`
- `logn13_d0_production_precondition_control_mismatch`
- `logn13_d0_production_precondition_numeric_mismatch`

Also include the corresponding `production_readiness` enum from P5.

---

# Required compact artifact

Create:

`results/FIX-001-P3-DESIGN-LOGN13-D0-PRODUCTION-PRECONDITION-CLOSURE-summary.json`

Include:

- provenance;
- checked-in frontend-config fingerprint;
- source-backed current-production control;
- accepted D0 control;
- A/B/C/D profile metadata;
- A/B/C/D generated-power capacity/contracts;
- A/B/C/D local power metrics;
- A/B/C/D polynomial/EvalMod/post-S2C/public metrics;
- first divergence for each failing case;
- q0 effect conclusion;
- planScale effect conclusion;
- production readiness;
- production patch contract;
- corrected row-comparison regression evidence;
- classification;
- recommended next target;
- validation flags.

Do not serialize full slots, coefficient vectors, or complete RNS rows.

---

# Validation

Require:

- focused test matching `TestFIX001P3.*D0.*Production.*Precondition.*Closure`;
- `go test ./...` in Primary;
- `git diff --check`;
- no NaN/Inf in artifact;
- Secondary remains exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean and unmodified;
- checked-in `configs/bootstrap_config.logN13.json` remains unchanged with q0=55;
- D case reproduces accepted D0 public-like `0.005554220603853743` within existing tolerance;
- all four matrix cases use the complete identical arithmetic stack;
- only q0 bit size and planScale vary across A/B/C/D;
- q012 intended capacity uses physical intended values, never wrapped q01 representatives;
- post-Rescale q01 uniqueness checked before q2 drop;
- corrected row comparison does not double-MForm already-Fast operands;
- no Secondary production integration;
- no LogN16;
- no benchmark;
- no Gate4/5;
- no EXP003;
- final Primary worktree clean after ordinary ff push.

---

# Recommended next target rules

If A passes:

`production integration of complete D0 arithmetic stack under fixed q0=55 / planScale91`

If A fails and B passes:

`production integration of complete D0 arithmetic stack under fixed q0=55 with internal planScale92`

If q0=56 is required:

`architecture decision on fixed frontend parameter configuration; do not production-integrate yet`

If D0 control fails:

`repair the precondition closure harness before any production work`
