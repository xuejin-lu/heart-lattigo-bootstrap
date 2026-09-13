# FIX-001-P3-DESIGN-LOGN13-PS-MIXED-SCALE-PRECISION

## Purpose

Continue the accepted LogN13 PS-arithmetic investigation after the common-scale sweep failed.

This task has two tightly coupled goals:

1. prove why common plan scale `2^93` suddenly diverges although maintained q0/q1 ciphertext checkpoints remain centered-unique and Fast/full-RNS rows match;
2. test whether a **mixed/per-block PS scale schedule** can keep risky coefficient injections at a safe scale while raising precision in other blocks enough to satisfy the polynomial budget.

Use validated oracle powers throughout candidate qualification so generated-power error remains excluded.

This is diagnostic/design only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`fa4757cd70a62120d8d82988b235259c9306fdb0`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

## Accepted evidence

From the polynomial decomposition task:

- generated powers and Fast PS arithmetic are both independent blockers;
- oracle-power Fast PS at common `2^91` is ~`6.27e-8` real error;
- polynomial design budget is `1.2e-8`.

From `FIX-001-P3-DESIGN-LOGN13-PS-ORACLE-SCALE-PRECISION`:

- `2^91` real error ~`6.2675e-8`;
- `2^92` real error improves to ~`3.3025e-8`;
- `2^93` suddenly diverges to ~`1.56064e-2`;
- `2^93` first cumulative budget crossing is `real:B3-term-2`;
- `2^93` final-Rescale input is already wrong by ~`1.56064e-2`;
- the `2^93` final Rescale itself adds only ~`1.16e-8` local error;
- all candidate maintained ciphertext checkpoints were reported centered-unique and Fast/full-RNS q0/q1 rows matched;
- common scale classification: `logn13_ps_oracle_common_scale_insufficient_precision`.

Source facts that must guide the audit:

- `EvaluateWithPlanScale` overrides every `plan.Value[i].Scale`, while the planner still receives the original target scale;
- scalar `MulThenAdd` derives a scalar encoding scale from `opOut.Scale/op0.Scale` when the accumulator is higher scale;
- `scalarNTT` rounds `coefficient * scale` to an integer and then reduces it modulo q0 and q1 independently;
- therefore a scalar injection can lose centered-Q01 uniqueness even when the *resulting ciphertext checkpoint* later remains centered-unique.

Do not assume that this is the proven cause until M1 measures it.

---

# Scope lock

LogN13 only.

Do not:

- modify Secondary production code;
- change generated-power construction;
- use imperfect production powers for PS candidate qualification;
- change polynomial coefficients, degree, basis or PS decomposition;
- change C2S or S2C;
- change Mod1 degree/DoubleAngle/bootstrap parameters/Q/P chain;
- add q2+ maintained arithmetic;
- consume extra Q levels;
- benchmark;
- run LogN16;
- run Gate 4/5;
- start EXP-003;
- integrate a candidate into production.

Primary diagnostic/design helpers are allowed.

---

# M0 — baseline reproduction

Using the same validated oracle-power injection path, reproduce:

- common `2^92` final real polynomial error near `3.3025e-8`;
- common `2^93` final real polynomial error near `1.56064e-2`;
- `2^93` first cumulative crossing `B3-term-2`;
- `2^93` final-Rescale pre-error already near `1.56064e-2`;
- Fast/full-RNS q0/q1 rows match at maintained checkpoints;
- Secondary exact/clean.

If not, stop:

`logn13_ps_mixed_scale_precondition_mismatch`.

---

# M1 — prove the 2^92 -> 2^93 semantic boundary

Instrument every baby-step scalar coefficient injection for common `2^92` and `2^93`.

For each term, including constants and every `MulThenAdd` coefficient, record internally:

- block/index ID (`B#-constant`, `B#-term-k`);
- coefficient value at high precision;
- power Level and Scale;
- accumulator Scale immediately before the operation;
- exact scalar encoding scale chosen by the Fast evaluator;
- exact signed rounded scalar integer(s) before modular reduction;
- `Q01 = q0*q1` and `Q01/2`;
- scalar centered-capacity ratio `abs(encodedInteger)/(Q01/2)`;
- whether the scalar integer is centered-unique modulo Q01;
- local semantic error of the actual Fast operation against the exact numerical operation on decoded actual inputs;
- resulting ciphertext centered-capacity ratio.

Also audit every scale-alignment/promotion site used before the first divergence:

- exact real-valued scale ratio;
- integer ratio actually used (`BigInt()` where applicable);
- fractional/truncation difference;
- any metadata-only scale relabel;
- local semantic error introduced there.

The goal is to classify the first demonstrated cause of the `2^93` jump as exactly one of:

- `ps_common_93_scalar_encoding_alias`
- `ps_common_93_integer_scale_alignment_loss`
- `ps_common_93_metadata_scale_relabel_loss`
- `ps_common_93_other_semantic_boundary`

The classification must point to the first concrete operation/checkpoint, expected to be at or before `B3-term-2`.

Do not call maintained-ciphertext centered uniqueness proof that scalar injection itself was centered-unique.

---

# M2 — derive per-block safe scale ceilings

For every PS baby block produced by the real plan, derive the highest common power-of-two block scale exponent in `91..96` that satisfies all of the following for that block:

1. every scalar coefficient injection is centered-unique modulo Q01;
2. every required integer scale-alignment/promotion operation is semantically valid within measured `1e-10` local error;
3. every maintained ciphertext checkpoint remains centered-unique;
4. Fast/full-RNS q0/q1 rows match where exact;
5. no extra level is consumed.

Call this the block's `safe_scale_ceiling`.

If no exponent above 92 is safe for a block, record 92.

Do not invent a fixed safety margin. Use strict exact uniqueness plus measured local semantic correctness. Report the actual headroom ratio so a later production task can choose margin if needed.

---

# M3 — deterministic mixed-scale candidate set

Keep oracle powers fixed.

The control candidate is:

- `C0`: all blocks at `2^92`.

Then evaluate deterministic candidates without an exhaustive combinatorial search.

## M3a — single-block lifts

For each block whose `safe_scale_ceiling >= 93`, test one candidate with only that block raised from `2^92` to `2^93` and all others held at `2^92`.

Record final real/imag polynomial error and all scale-alignment/capacity validity.

This measures which blocks actually benefit from one extra bit.

## M3b — safe-93 aggregate

Test one candidate with every block that safely supports `2^93` set to `2^93`, and all remaining blocks at `2^92`.

## M3c — greedy precision schedule

Starting from `C0`, repeatedly apply the single safe one-bit block increment that gives the largest reduction in final max(real, imag) polynomial error.

Rules:

- candidate increments are block exponent `e -> e+1`;
- never exceed that block's measured `safe_scale_ceiling`;
- accept an increment only if it strictly improves final max error and preserves all semantic/capacity contracts;
- deterministic tie-break: lower block index first;
- stop when no safe increment improves error or all blocks reach ceilings;
- cap the total number of accepted increments at 12 to keep scope bounded.

Record only the accepted schedule sequence, not every rejected internal vector.

This is a diagnostic design search, not production tuning.

---

# M4 — mixed-plan scale plumbing discipline

Primary diagnostic code may construct a plan with per-`plan.Value[i]` Scale overrides.

Do not simply relabel ciphertext Scale metadata to make giant-step additions pass.

At each giant merge require one of:

- scales already exactly compatible;
- a measured integer physical alignment that preserves represented value;
- an explicit diagnostic failure.

Any non-integer/fractional alignment that would require metadata-only relabeling is not acceptable as a precision-qualified candidate.

Keep:

- same polynomial coefficients;
- same PS decomposition;
- same Levels;
- same relinearization points;
- same single final Rescale boundary;
- no extra Q level.

---

# M5 — precision qualification

For every valid mixed candidate compare final Fast oracle-power polynomial output against:

1. validated plaintext/source polynomial oracle;
2. genuine Standard polynomial output on the same corrected-C2S input;
3. stage-aligned full-RNS mirror where exact.

A mixed candidate qualifies only if both:

- real error <= `1.2e-8`;
- imag error <= `1.2e-8`;

and all semantic/capacity/level contracts pass.

Choose the qualifying candidate with:

1. fewest blocks above `2^92`;
2. then smallest sum of block exponents;
3. then lowest final max(real, imag) error;
4. deterministic lexicographic block-exponent tie-break.

If one qualifies, classify:

`logn13_ps_oracle_mixed_scale_candidate_validated`.

Do not integrate it yet.

---

# M6 — downstream oracle-power sufficiency

Only for the selected precision-qualified mixed candidate:

- derive the normalized Mod1 K schedule from the actual final compressed-polynomial working Scale;
- run valid normalized DoubleAngle;
- run unchanged Fast S2C;
- run supported one-input unpack/finalization/public-like path.

Require valid q0/q1 centered-capacity discipline throughout.

Record:

- EvalMod real/imag vs genuine Standard;
- post-S2C semantic error;
- public-like max-component error;
- metadata.

Target:

`public_like_max_component_error <= 1e-2`.

Remember this still uses oracle powers and therefore proves only that the PS arithmetic blocker can be closed independently.

---

# M7 — if no mixed schedule qualifies

Classify one supported result:

- `logn13_ps_oracle_mixed_scale_blocked_by_scalar_encoding`
- `logn13_ps_oracle_mixed_scale_blocked_by_alignment`
- `logn13_ps_oracle_mixed_scale_insufficient_precision`

Record:

- proven `2^93` semantic-boundary cause;
- per-block safe ceilings;
- best valid mixed schedule;
- best real/imag error;
- first remaining precision blocker/checkpoint;
- whether final-Rescale local error or pre-final accumulated error dominates.

Only a later task may redesign a specific primitive or introduce a third maintained limb.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-PS-MIXED-SCALE-PRECISION-summary.json`

Include only:

- provenance;
- M0 baseline;
- M1 first semantic-boundary cause and compact 2^92/2^93 scalar/alignment evidence for the causal checkpoint;
- one row per block with safe scale ceiling and limiting operation;
- C0 control;
- single-block lift final-error table;
- safe-93 aggregate result;
- accepted greedy schedule steps;
- selected/best mixed schedule;
- final real/imag polynomial errors;
- downstream metrics if reached;
- classification;
- first remaining blocker;
- validation flags.

Do not serialize full vectors, coefficient arrays, RNS rows, complete operation traces or large rejected-candidate tables.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*PS.*Mixed.*Scale.*Precision` pass;
- Primary `go test ./...` passes;
- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no generated-power redesign;
- no q2+ arithmetic;
- no production integration;
- no C2S/S2C change;
- no bootstrap parameter retuning;
- no extra Q level;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.