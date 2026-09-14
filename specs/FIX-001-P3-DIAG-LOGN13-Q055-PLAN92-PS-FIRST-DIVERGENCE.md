# FIX-001-P3-DIAG-LOGN13-Q055-PLAN92-PS-FIRST-DIVERGENCE

## Goal

Localize the **first actual semantic divergence inside Paterson–Stockmeyer evaluation** for the fixed frontend profile:

- LogN13
- checked-in `q0 = 55`
- `planScale = 2^92`
- complete accepted D0 arithmetic stack
- generated powers using genuine temporary-q012 multiply-first
- final-parent B4 T2 guard using the now-proven temporary-q012 implementation

The previous task proved that the known B4 T2 q01 transient capacity violation is real but **not sufficient to explain** the catastrophic q0=55/plan92 EvalMod error. The q012 T2 replacement has correct q0/q1/q2 rows, correct centered divide-by-two, correct metadata and valid q01 contraction, yet the downstream result remains unchanged at approximately:

- EvalMod real: `102.29911082927852`
- EvalMod imag: `102.2990271839068`
- post-S2C: `114.46562670803426`
- public-like: `3662.9000547146406`

Therefore do **not** continue treating T2 as the sole q0/planScale blocker.

This task must find the first PS operation/window where q0=55/plan92 ceases to represent the intended physical computation, explain why q0=56/plan92 avoids that divergence, and test one minimal local-q2 repair window if the evidence supports one.

Primary-only diagnostic/design work. Do not modify Secondary production code.

---

# Required provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required ancestor:

`10302fee90b4b2a372f05eeb670c9fa2be361db6`

Start from clean `main`, fetch + ff-only pull, then read fresh:

- `AGENTS.md`
- `CURRENT_TASK.md`
- this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required exact commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Branch: `fast-ckks`

State: clean.

Secondary is read-only.

---

# Accepted facts from the previous task

Treat these as controls to reproduce, not hypotheses to reopen without contrary evidence.

## q0/planScale interaction

Previous matrix:

- A = q0=55, plan91: FAIL late, public-like `0.01076789220080715`
- B = q0=55, plan92: FAIL catastrophically, public-like `3662.9000547146406`
- C = q0=56, plan91: FAIL late, public-like `0.01076789220080715`
- D = q0=56, plan92: PASS, public-like `0.005554220603853743`

Correct interpretation:

`A=false, B=false, C=false, D=true` means a q0/planScale interaction/joint precondition until localized. It is not evidence that q0 width alone is sufficient.

## generated powers

For q0=55/plan92, generated T2/T3/T4/T6/T8/T16 temporary-q012 multiply-first proofs pass, completed local residuals remain near canonical, Q012 stays unique, rounded Rescale proofs pass, and post-Rescale Q01 contractions pass.

Generated powers are therefore not the current first blocker.

## B4 T2 temporary-q2 guard

For q0=55/plan92:

- q01-only guarded post-add intended ratio is about `1.0507925 * (Q01/2)` and is genuinely outside Q01;
- Q012 ratio is tiny (~`1.91e-12`) and unique;
- temporary-q012 guard reproduces independent q0/q1/q2 rows;
- symmetric rounded divide-by-two proof passes;
- post-divide Q01 is unique (~`0.5253963` ratio);
- contraction and metadata proofs pass;
- the q012 guard is genuinely wired into the PS suffix replay;
- nevertheless downstream EvalMod/public metrics remain the same catastrophic B values.

Therefore T2 q01 overflow is a real local-capacity defect but not the remaining first system-causal defect.

---

# Fixed experiment stack

For the principal failing case B use the checked-in frontend config unchanged:

`configs/bootstrap_config.logN13.json`

with:

`q0 = 55`

and:

`planScale = 2^92`.

Use the complete accepted stack:

- accepted C2S compression;
- generated-power temporary-q012 multiply-first;
- final-parent B4 T2 temporary-q012 one-bit scalar guard;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DoubleAngle local-q2 rounds;
- same coefficients and PS decomposition;
- same S2C/finalization;
- same deterministic workload/reference.

Control D is identical except diagnostic q0=56.

Do not vary any other parameter.

---

# L0 — reproduce B and D controls

Reproduce:

## B: q0=55 / plan92 / q012 T2 guard

Expected order of magnitude:

- EvalMod real ~`102.29911082927852`
- EvalMod imag ~`102.2990271839068`
- post-S2C ~`114.46562670803426`
- public-like ~`3662.9000547146406`

## D: q0=56 / plan92 / same q012 T2 guard

Expected:

- public-like `0.005554220603853743` within existing tolerance;
- normal accepted EvalMod regime (~`6.16e-5` real, ~`4.75e-5` imag).

If either control fails to reproduce, stop:

`logn13_q055_plan92_ps_control_mismatch`.

---

# L1 — instrument the complete PS operation sequence

Instrument the **actual accepted PS replay**, not a simplified synthetic polynomial evaluator.

For both real and imag, assign a stable source-backed operation ID to every numerically relevant transition in execution order.

At minimum include:

## Baby-step construction

For every PS block and every nonzero coefficient contribution:

- accumulator before coefficient;
- selected power `T_k`;
- scalar coefficient;
- scalar encoding scale / rounded integer representation;
- scalar product contribution;
- accumulator immediately after `MulThenAdd` or guarded equivalent.

Include the final-parent B4 T2 operation but do not stop there.

## Giant-step / merge operations

For every merge:

- child inputs before alignment;
- any relinearization/truncation boundary;
- each Rescale input and output;
- multiplication by giant-step power;
- scale relabel/alignment;
- add/sub merge output.

Include the already accepted G0/F0 local-q2 windows explicitly.

## Final polynomial output

Record the final real/imag polynomial ciphertext immediately before downstream EvalMod/DoubleAngle processing.

Do not serialize full coefficient arrays.

---

# L2 — exact intended-capacity and centered-sensitive analysis

For every PS checkpoint in L1, record separately:

1. **reduced q01 representative**;
2. **intended physical integer**, reconstructed from a sufficiently wide independent oracle.

Prefer q012 whenever Q012 is proven unique. If Q012 is not sufficient, use full-RNS/big.Int only as an oracle and report that q012 is insufficient for that checkpoint.

For every checkpoint record:

- max intended `|x|`;
- Q01/2;
- Q012/2 when available;
- intended/Q01 ratio;
- intended/Q012 ratio;
- reduced-q01 ratio;
- q01 intended outside count;
- q012 intended outside count;
- minimum required centered-uniqueness bits;
- Level, Scale, degree, NTT, Montgomery metadata.

Mark whether the **next operation is centered-sensitive**. Centered-sensitive operations include at least:

- Fast Rescale / rounded division;
- explicit centered divide/contraction;
- any operation whose scalar encoding or alignment derives an intended integer from centered interpretation rather than pure modular arithmetic.

A pure modular add/multiply may carry wrapped residues without immediately changing q0/q1 rows; the causal failure occurs when a later centered-sensitive boundary interprets the wrapped representative as the physical integer.

This distinction is mandatory.

---

# L3 — q0=55 vs q0=56 row and semantic comparison

Run the identical trace for D (q0=56/plan92).

For every stable operation ID compare B vs D after normalizing for their different q0 modulus:

- decoded semantic value/error vs the same canonical PS oracle;
- scale/metadata;
- intended physical coefficient magnitude;
- capacity ratio under each profile;
- whether q01 is unique before centered-sensitive boundaries;
- local q0/q1 rows against that profile's own independent oracle.

Do **not** compare raw q0 residues directly across different moduli as if they should be equal.

Find the earliest operation ID where:

- D remains oracle-consistent;
- B becomes semantically inconsistent, or
- B enters a q01-invalid state that is later consumed by the first centered-sensitive boundary producing semantic divergence.

Call this the **first causal PS window**.

---

# L4 — row-level oracle at the first causal window

For the first causal PS window, provide a detailed local proof.

Record:

- input rows vs independent oracle;
- intended physical values before each operation;
- reduced q01 values;
- q012/full-RNS rows when used as oracle;
- scalar quantization/encoding if applicable;
- exact Rescale or centered contraction quotient if applicable;
- output rows;
- output decoded semantic residual.

The task must answer:

> Is the B failure caused by insufficient Q01 centered capacity before a centered-sensitive operation, by scale/scalar-encoding semantics coupled to q0, by an operation mismatch unrelated to capacity, or by something else?

Do not classify based only on final EvalMod magnitude.

---

# L5 — dependency-aware reset attribution

Use the existing oracle/reset methodology to prove causality.

At the first causal window:

1. replace/reset only the earliest bad output with the corresponding oracle-correct state while keeping all later B operations unchanged;
2. replay through PS and downstream EvalMod/S2C/public;
3. record whether the catastrophic ~102 EvalMod error collapses to the accepted numeric regime.

If a single reset is insufficient because multiple siblings are independently bad, perform a bounded cumulative reset in execution/dependency order and report the minimal causal set.

Do not reset generated powers globally; those are already accepted.

Do not use decoded slot injection as the production candidate. Oracle replacement is diagnostic attribution only.

---

# L6 — one minimal local-q2 repair candidate when supported

Only if L2–L5 show that the first causal window is a Q01-capacity/centered-interpretation problem and Q012 is sufficient, implement one Primary-only minimal local-q2 candidate for that **whole causal window**.

Important:

- do not patch only the final operation if earlier pure-modular steps already wrapped;
- q2 authority must begin at the last proven-Q01-unique state before the unsafe physical growth;
- keep q0/q1/q2 authoritative across all required modular operations;
- perform the centered-sensitive operation in q012;
- contract back to q01 only after the result is again proven Q01-unique;
- q3+ stay dormant;
- no Standard full-RNS producer;
- big.Int/full-RNS oracle remains diagnostic only.

Then replay the entire B system.

If the evidence instead points to scale/scalar encoding rather than capacity, do not force a local-q2 repair; implement only the minimal source-equivalent diagnostic correction needed to test that hypothesis.

---

# L7 — required conclusions

Choose exactly one primary classification:

- `logn13_q055_plan92_ps_local_q2_window_system_sufficient`
- `logn13_q055_plan92_ps_first_divergence_q01_capacity`
- `logn13_q055_plan92_ps_first_divergence_scale_encoding`
- `logn13_q055_plan92_ps_first_divergence_operation_mismatch`
- `logn13_q055_plan92_ps_multiple_independent_blockers`
- `logn13_q055_plan92_ps_q012_insufficient`
- `logn13_q055_plan92_ps_control_mismatch`

If a minimal repair makes B pass `public_like <= 1e-2`, set production readiness accordingly and recommend production integration under fixed q0=55 only if all accepted stack prerequisites are represented.

If B remains failing, report the next earliest causal divergence. Do not recommend changing frontend q0 to 56 merely because D passes unless the trace proves no supported local-q2/scale-correct path can preserve q0=55.

---

# Required compact artifact

Create:

`results/FIX-001-P3-DIAG-LOGN13-Q055-PLAN92-PS-FIRST-DIVERGENCE-summary.json`

Include:

- provenance;
- B/D controls;
- ordered PS operation map;
- compact per-operation capacity/metadata table;
- B-vs-D semantic comparison;
- first causal window;
- detailed row/oracle proof for that window;
- reset attribution results;
- minimal causal set;
- local-q2 or scale-correction candidate result if applicable;
- EvalMod/post-S2C/public metrics before/after candidate;
- classification;
- first remaining blocker;
- production readiness;
- recommended next target;
- validation flags.

No full slots, full coefficient vectors, or full RNS rows.

---

# Validation

Require:

- focused test matching `TestFIX001P3.*Q055.*Plan92.*PS.*First.*Divergence`;
- `go test ./...` in Primary;
- `git diff --check`;
- artifact has no NaN/Inf;
- checked-in q0=55 config unchanged;
- Secondary exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean, unmodified;
- B and D controls reproduced;
- generated-power q012 proofs remain accepted;
- T2 q012 guard remains accepted and genuinely wired;
- every PS operation receives a stable ID;
- intended physical capacity is never inferred from wrapped q01 residue;
- centered-sensitive boundaries explicitly identified;
- first causal divergence established before any system-level classification;
- reset attribution performed;
- q012 candidate, if used, spans the full causal window rather than only its final step;
- no frontend q change;
- no production integration;
- no LogN16;
- no benchmark;
- no Gate4/5;
- no EXP003;
- final Primary worktree clean after ordinary ff push.

---

# Recommended next target rules

If a bounded PS local-q2/scale repair makes q0=55/plan92 pass:

`production integration of the complete accepted LogN13 D0 stack under fixed q0=55, including the newly proven PS window`

If the first causal window is identified but not yet repaired:

`design the narrow production-compatible fix for the identified q0=55/plan92 PS window`

If q012 itself is insufficient:

`reassess temporary-limb width for the identified PS window; do not change frontend q0 yet`
