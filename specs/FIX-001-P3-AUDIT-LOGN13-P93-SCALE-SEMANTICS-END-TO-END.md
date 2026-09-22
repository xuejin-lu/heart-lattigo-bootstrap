# FIX-001-P3-AUDIT-LOGN13-P93-SCALE-SEMANTICS-END-TO-END

> **STATUS: SUPERSEDED AS AN EXECUTABLE CODEX TASK**
>
> Preserve this file as the original audit design and decision record. Do **not** execute it directly.
> The scaling audit is now orchestrator-led: ChatGPT performs the static/mathematical review first. A later Codex task, if any, will be measurement-only and limited to runtime quantities that cannot be established from committed source/evidence.

## Purpose

The LogN13/P93/Q012 `1e-2` system milestone is already finalized and must remain frozen.

Authoritative Secondary candidate:

`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Authoritative exact Fast E2E:

[
E_{max}=0.009705381393898434 < 1e-2.
]

Genuine Standard control:

[
5.830057349387463e-8.
]

This task is **not a new precision phase** and must not attempt to improve the result.

The purpose is a full-system audit of every scaling-factor/Scale semantic contract in the current Fast bootstrap implementation, because several different concepts have previously been conflated:

- ciphertext metadata `Scale`;
- physical coefficient multiplication;
- CKKS Rescale modulus division;
- `ScalingFactor()` used by Mod1;
- `targetScale`;
- polynomial `planScale = 2^93`;
- maintained/deferred virtual scalar exponents;
- integer scale-alignment ratios;
- DFT restore powers;
- public `DefaultScale` resets.

The audit must prove where each scale change comes from, whether it changes physical coefficients, whether it changes only interpretation, what mathematical semantic is intended, and whether Fast matches Genuine Standard at the relevant public contract.

No Secondary modification is authorized.

---

# Repository state

## Primary

`xuejin-lu/heart-lattigo-bootstrap`
branch `main`

Read:

- `AGENTS.md`
- `CURRENT_TASK.md`
- `docs/CODEX_HANDOFF.md`
- this spec
- finalized milestone evidence

## Secondary

`xuejin-lu/lattigo`
branch `fast-ckks`

Required committed SHA:

`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Require clean worktree.

No Secondary edits, commits, or pushes.

---

# Audit taxonomy

Every scale-changing event must be classified into exactly one primary type.

## M — metadata reinterpretation

Ciphertext coefficients unchanged; only metadata `Scale` changes.

Examples to audit:
- Mod1 normalization `res.Scale = ScalingFactor()`
- coherent-scale assignment
- public DefaultScale restoration
- exact-equality normalization after `InDelta`

This changes decoded interpretation unless paired with an explicit virtual/deferred semantic contract.

## P — physical scalar with matching metadata

Coefficients are multiplied by (s), and metadata Scale is also multiplied by (s).

Expected semantic invariant:

[
rac{s c}{sDelta}=rac{c}{Delta}.
]

Examples:
- DFT restore powers
- explicit scale-alignment multipliers
- some ModUp scaling.

## R — CKKS Rescale

Coefficients are approximately divided by modulus (q_ell), and Scale is divided by the same (q_ell).

Expected semantic invariant up to rounding:

[
rac{c/q_ell}{Delta/q_ell}approxrac{c}{Delta}.
]

## V — virtual/deferred scalar

Physical coefficients and metadata do not by themselves represent the current logical value; a tracked exponent/factor is required.

Example:
- normalized LogN13 Mod1 maintained-scalar schedule.

## L — planner/logical scale

A scale controls polynomial planning or coefficient encoding but is not itself a direct ciphertext semantic mutation at assignment time.

Examples:
- `targetScale`
- `planScale = 2^93`
- PS polynomial planned scales.

## A — alignment/rounding

A real-valued scale ratio is converted to an integer multiplier or approximate metadata equality.

Examples:
- `Scale.Div(...).BigInt()`
- `math.Round(scale)`
- `InDelta` followed by metadata equality assignment.

## C — public API contract

A metadata transition deliberately establishes the Standard-compatible externally visible scale, even though it changes the numerical interpretation of the same coefficients.

Example:
- final residual `DefaultScale`.

Every event may have secondary tags, but must have one primary classification.

---

# R0 — frozen baseline replay

Before auditing scale semantics, rerun the finalized committed candidate without modification.

Require:

- Secondary SHA exactly `40532b4d...`
- worktree clean
- Fast exact E2E <= `1e-2`
- polynomial gates pass
- public EvalMod gates pass
- Genuine Standard baseline remains approximately `5.83e-8`.

If not:

`SCALE_AUDIT_BASELINE_REPLAY_CONFLICT`

and stop.

---

# R1 — static scale-mutation inventory

Audit the entire executed LogN13 Fast bootstrap source path and enumerate every location that can change:

1. ciphertext `Scale`;
2. coefficient magnitude through integer/real scalar multiplication;
3. level/modulus through Rescale;
4. planner scale;
5. virtual/deferred scalar state.

At minimum inspect:

- `circuits/ckks/bootstrapping/fast_bootstrap.go`
- `circuits/ckks/bootstrapping/fast_modup.go`
- `circuits/ckks/bootstrapping/fast_packing.go`
- `circuits/ckks/dft/fast.go`
- `circuits/ckks/mod1/fast.go`
- `circuits/ckks/polynomial/fast.go`
- `schemes/ckks/fast/fast_add.go`
- `schemes/ckks/fast/fast_mul.go`
- `schemes/ckks/fast/rescale.go`
- any called helper that changes metadata or maintained residues.

For every mutation record:

- file/function;
- operation;
- taxonomy M/P/R/V/L/A/C;
- coefficient effect;
- metadata effect;
- expected mathematical invariant;
- analogous Genuine Standard behavior if one exists;
- executed in current P93 path yes/no.

No source location may be omitted merely because current E2E passes.

---

# R2 — end-to-end runtime Scale ledger

Run one deterministic 4096-slot Fast bootstrap and one Genuine Standard bootstrap from the same input.

Create a compact ledger of the Fast path at all major boundaries.

Required checkpoints:

## Public / packing boundary

1. original input
2. PackAndSwitchN1ToN2 output
3. ScaleDown input/output
4. ModUp input
5. ModUp after basis raise
6. ModUp after any physical scale lift
7. ModUp after Trace/Montgomery conversion

## C2S

8. C2S input
9. after each DFT factor group Rescale
10. after each restore-plan scalar, where used
11. C2S real output
12. C2S imag output

## EvalMod entry / normalization

13. EvalMod input real/imag
14. after metadata normalization to `ScalingFactor()`
15. after cosine offset
16. derived `targetScale`
17. `planScale`
18. polynomial input

## Generated powers / PS

19. T1
20. T2
21. T3
22. T4
23. T6
24. T8
25. T16
26. each PS baby-step result
27. each PS giant-step stable result
28. polynomial root before final Rescale
29. polynomial output after final Rescale

## Normalized LogN13 DoubleAngle

30. polynomial output entering normalized path
31. coherent metadata Scale assignment
32. current virtual exponent
33. each DA round:
    - pre-square
    - after square where semantic coordinate is valid
    - physical maintained multiplier
    - constant encoding scale
    - post-Rescale
    - updated virtual exponent
34. final maintained-scalar materialization
35. internal final metadata reset to caller EvalMod input Scale
36. public EvalMod output / public Scale contract

## S2C / output

37. S2C input
38. after each S2C DFT group Rescale
39. S2C output
40. UnpackAndSwitchN2ToN1 input/output
41. pre-finalization output
42. post-finalization public output
43. decoded final output

At each checkpoint record:

- Level
- Degree
- metadata Scale (exact decimal and log2)
- current top modulus (q_ell), if relevant
- maintained limb count
- virtual exponent/factor, if any
- physical scalar applied since previous checkpoint
- Rescale divisor since previous checkpoint
- expected Scale formula
- actual/formula relative error
- semantic max magnitude
- Fast-vs-Standard metric when stages are semantically comparable.

Do not dump vectors.

---

# R3 — semantic invariant checks for every scale event

For every executed scale mutation, perform the appropriate invariant test.

## Type P

Before and after physical scalar + metadata scalar:

[
v_{before}approx v_{after}.
]

Record semantic residual.

## Type R

Before and after Rescale:

[
v_{before}approx v_{after}
]

up to measured CKKS rounding floor.

Record:

- divisor;
- expected output Scale:
  [
  Delta_{out}=Delta_{in}/q_ell;
  ]
- actual output Scale;
- relative Scale formula error;
- semantic rounding error.

## Type M/C

Because coefficients do not change, explicitly record the interpretation ratio:

[
ho=Delta_{before}/Delta_{after}.
]

State why this reinterpretation is mathematically required.

Compare against Standard source/runtime contract.

A metadata-only transition without a Standard-equivalent or explicit virtual contract is a finding.

## Type V

Track both:

- raw decoded value;
- canonical logical value using the virtual factor.

Require canonical semantics to remain consistent across stable boundaries.

Do not use raw virtual-coordinate values as causal metrics.

---

# R4 — high-risk numerical conversion audit

This section is mandatory.

## 4.1 ModUp Float64 / Round mismatch

Current Fast ModUp contains the pattern:

- derive real `scale` using `Float64()`;
- physical scalar:
  [
  s=operatorname{round}(scale);
  ]
- metadata multiplication uses the original real `scale`.

For the exact P93 runtime, record:

- exact source quantities;
- computed floating ratio;
- rounded integer (s);
- absolute difference:
  [
  |s-scale|;
  ]
- relative mismatch:
  [
  |s/scale-1|;
  ]
- predicted semantic distortion;
- measured before/after semantic distortion.

Compare to Standard ModUp semantics.

If the ratio is mathematically intended to be an integer, prove how close it is and why.

Do not dismiss this because E2E passes.

## 4.2 addAligned/subAligned BigInt conversion

For every executed alignment in generated powers and PS:

- exact Scale ratio;
- resulting `BigInt()` multiplier;
- relative ratio mismatch;
- semantic residual introduced by alignment;
- whether ratio is exactly integral or merely approximated.

Report maximum mismatch across the whole run.

## 4.3 InDelta metadata snapping

Find every executed path that does:

1. checks `Scale.InDelta(...)`;
2. then assigns one Scale metadata value to another without coefficient change.

For each:

- pre-snap Scale ratio;
- tolerance;
- actual relative difference;
- implied semantic interpretation shift;
- measured downstream effect.

Special attention:
- PS giant-step alignment.

## 4.4 Float64 use elsewhere

Inventory every executed conversion of Scale or scale-derived values through `float64`.

State whether each is:
- harmless diagnostic/control-flow;
- used to choose an integer physical multiplier;
- used to mutate metadata;
- precision-sensitive.

---

# R5 — Mod1 scaling-factor audit

Treat these as distinct variables and never collapse them into one “scale”:

[
Delta_{input}
]

[
S_{mod1}=	ext{ScalingFactor()}
]

[
S_{target}
]

[
S_{plan}=2^{93}
]

[
S_{poly-out}
]

[
k_{virtual}
]

[
S_{coherent}
]

[
S_{DA,r}
]

[
Delta_{public}
]

For the actual run derive and record all of them.

Explicitly verify:

1. initial metadata normalization to `ScalingFactor()` matches Genuine Standard;
2. `targetScale` Q-index schedule matches Genuine Standard;
3. P93 `planScale` affects planner semantics only as intended;
4. generated-power post-product schedule produces expected ~(2^{60})-class power scales;
5. polynomial final Scale is the expected ~(2^{33}) working Scale;
6. coherent Scale assignment and (k=27) virtual factor form a consistent logical representation;
7. each DA multiplier exponent (a=28) satisfies the derived recurrence;
8. each post-Rescale Scale formula matches:
   [
   S_{next}=S_{current}^2/q_ell;
   ]
9. final virtual materialization + caller Scale reset reproduces Genuine Standard internal semantics;
10. public DefaultScale reset reproduces Genuine Standard public semantics.

The audit must explicitly revisit and document why the previous “(2^{50}	o2^{45}) Fast Scale bug” interpretation was incorrect.

---

# R6 — DFT / transform scaling audit

For C2S and S2C:

- enumerate each matrix group;
- input Scale;
- linear-transform scale effect;
- group Rescale divisor;
- restore-plan power (2^k), if any;
- physical multiplier and metadata multiplier equality;
- output Scale.

Compare group-by-group scale progression with Genuine Standard.

Require semantic stability at each group boundary within the expected transform numerical floor.

Flag any restore factor that is not an exact power-of-two paired physical+metadata operation.

---

# R7 — public boundary audit

Compare Fast vs Genuine Standard at:

- EvalMod public output Scale;
- S2C output Scale;
- unpack output Scale;
- final public residual ciphertext Scale.

Verify:

[
Delta_{final}=	ext{ResidualParameters.DefaultScale()}.
]

For every metadata-only public reset:

- prove the Standard contract uses the same externally visible Scale;
- prove the preceding internal coefficients are in the expected coordinate system;
- record semantic interpretation change.

No public Scale reset may be accepted solely because exact E2E happens to pass.

---

# R8 — invariant ledger and findings

Produce a final table where every scale event has status:

- `PASS_EXACT`
- `PASS_NUMERICAL_FLOOR`
- `PASS_STANDARD_CONTRACT`
- `WARNING_ROUNDING_DEPENDENT`
- `WARNING_UNPROVEN_CONTRACT`
- `FAIL_SEMANTIC_MISMATCH`

For every warning/failure include:

- exact source location;
- runtime values;
- maximum possible semantic impact;
- observed semantic impact;
- whether it can threaten:
  - current `1e-2` milestone;
  - future `1e-7` precision work.

Do **not** fix anything in this task.

---

# R9 — adversarial local scale sanity checks

Without changing production parameters, add Primary-only diagnostic checks around the current committed path.

At minimum test:

1. scale invariance of exact power-of-two physical+metadata multiplication;
2. Rescale formula at q01 and Q012 maintained domains;
3. integer alignment ratios used in actual P93 PS;
4. ModUp rounded scalar mismatch;
5. metadata-only Scale reset interpretation ratios;
6. virtual exponent canonicalization formulas.

These are diagnostic assertions only.

No parameter sweep.

---

# Required artifacts

Create both:

## JSON

`results/FIX-001-P3-AUDIT-LOGN13-P93-SCALE-SEMANTICS-END-TO-END-summary.json`

Keep <=500 pretty-printed lines.

Include:

- provenance;
- baseline replay;
- count of all executed scale mutation events;
- taxonomy counts;
- high-risk numeric conversions;
- Mod1 scale ledger;
- DFT scale ledger;
- public-boundary audit;
- warnings/failures;
- exact E2E control;
- one final classification.

## Human-readable ledger

`docs/FIX-001-P3-LOGN13-P93-SCALE-LEDGER.md`

Keep it compact and table-driven.

It must clearly distinguish:

- metadata Scale;
- physical coefficient multiplier;
- Rescale divisor;
- virtual factor;
- planner scale;
- public contract.

Do not write a long narrative.

---

# Decision classification

Choose exactly one:

## A — `SCALE_AUDIT_BASELINE_REPLAY_CONFLICT`

Finalized milestone cannot be reproduced from committed Secondary.

## B — `SCALE_AUDIT_ALL_CONTRACTS_VERIFIED`

Every executed scale mutation is exact, numerical-floor-safe, or proven Standard-compatible; no unresolved warning remains.

## C — `SCALE_AUDIT_NUMERICAL_ROUNDING_WARNINGS_ONLY`

No current semantic failure, but one or more rounding/float/integer-conversion contracts are non-exact and relevant to future precision tightening.

## D — `SCALE_AUDIT_METADATA_CONTRACT_WARNING`

One or more metadata-only Scale changes are not sufficiently justified against Standard or virtual semantics, though current E2E still passes.

## E — `SCALE_AUDIT_CAUSAL_SCALE_DEFECT_FOUND`

A scale/scalar contract creates a reproducible semantic error beyond its justified numerical floor.

Do not repair it; record the first causal boundary.

## F — `SCALE_AUDIT_ALIGNMENT_INVALID`

The full scale lineage cannot be reconstructed reliably enough.

---

# Validation

Run:

- finalized exact-E2E replay;
- focused scale-ledger diagnostics;
- invariant tests from R9;
- Primary `go test ./...`;
- `git diff --check`.

Secondary must remain at:

`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

with clean worktree throughout.

---

# Prohibitions

- no Secondary modification
- no new algorithmic work
- no precision tightening
- no parameter tuning
- no q/planScale changes
- no Scale “fix”
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

This is a post-milestone integrity audit only.
