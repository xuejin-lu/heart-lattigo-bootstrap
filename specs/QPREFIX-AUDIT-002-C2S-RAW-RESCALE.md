# QPREFIX-AUDIT-002 — Current-Branch C2S Raw-Bound and Rescale Recurrence

## Status

Executable Codex audit task.

## Task class

`M/E — Focused evidence completion`

No production behavior change is authorized.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- owns compact evidence/report.

Secondary:
- `xuejin-lu/lattigo@fast-qprefix`
- audit target only.

Accepted prerequisite:
- QPREFIX-AUDIT-001 at Primary commit `494e7b7b672cbedb4eb842887d464df2ea0cb7f9`
- Q-prefix architecture at Secondary `c9fb900314a1c82f2cd21320554d3d47fd7dbf00`.

## Purpose

Close the first evidence gap from QPREFIX-AUDIT-001:

**C2S group 0 raw LinearTransform output before Rescale at logical Level 16.**

Then complete the same source -> raw transform -> Rescale -> restore evidence chain for all four LogN13/P93 C2S groups on the current `fast-qprefix` branch.

This task must not audit EvalMod/PS or S2C beyond preserving the existing public control. Those become later focused audits.

---

# 1. Exact execution to reproduce

Use the same deterministic q0=56 LogN13/P93 profile and C2S restore plan as the accepted public Bootstrap control.

For each C2S group `g=0..3`:

1. start from the exact current group input;
2. call the current Fast `LinearTransform` directly;
3. capture the raw output **before Rescale**;
4. call the current Fast `Rescale`;
5. capture the post-Rescale output;
6. if the restore exponent is nonzero:
   - apply the same maintained integer restore;
   - apply the same Scale metadata update;
7. capture the post-restore output;
8. feed that exact state into the next group.

Do not call the combined DFT helper for the audited path except as an end-to-end equivalence control.

---

# 2. Representation/domain handling

For every captured ciphertext:
- record logical Level;
- Scale;
- degree;
- IsNTT;
- IsMontgomery;
- maintained row count;
- exact q values used.

To obtain coefficient-domain bounds:
- copy maintained authoritative rows only;
- undo Montgomery form when required;
- INTT maintained rows;
- reconstruct centered integer coefficients from the strongest currently authoritative prefix available in the branch-point implementation.

Historical branch-point Q012 support is sufficient for this audit when strict centered uniqueness under Q012 is proven.

Do not use dormant q3+ rows.

---

# 3. Exact raw bounds

For each component (j) and every checkpoint, compute the exact observed maximum:

[
B_j=max_k |X_{j,k}|.
]

Required checkpoints per group:

- `group_g_input`
- `group_g_raw_linear_transform`
- `group_g_post_rescale`
- `group_g_post_restore` when restore is nonzero.

At minimum use Q012 reconstruction where the current branch maintains q2 and proves uniqueness.

If a checkpoint is only Q01-authoritative, use Q01 and state that explicitly.

No inferred bound may be labeled exact.

---

# 4. Q-prefix v2 capacity checks

For each checkpoint compute strict capacity against the architecture policy:

[
S_Q(ell)=prod_{i=0}^{min(ell,3)}q_i.
]

Report:

[
ho_j=rac{2B_j}{S_Q(ell)}.
]

Require:

[
ho_j<1.
]

Because current branch-point code may not physically maintain q3, it is acceptable to prove a stronger condition:

[
2B_j<q_0q_1q_2
]

at a Level >= 3 checkpoint. If Q012 already uniquely contains the lift, Q0123 capacity follows immediately.

Record both:
- strongest physically reconstructed prefix;
- Q-prefix-v2 policy prefix.

---

# 5. Rescale recurrence audit

For each C2S group record:

- source checkpoint = raw LinearTransform output;
- source logical Level (ell);
- exact logical divisor:
  [
  d=q_ell;
  ]
- exact source bounds (B_j);
- conservative predicted post-bound:
  [
  widehat B'_j=
  leftlfloor
  rac{B_j+(d-1)/2}{d}
  ightfloor;
  ]
- exact observed post-Rescale bounds (B'_j);
- target logical Level (ell-1);
- target Q-prefix capacity.

Require:

[
B'_jle widehat B'_j
]

for the observed deterministic fixture.

If this inequality fails, stop with `NEEDS_WEB_REVIEW`; it indicates a mismatch in the authoritative-lift interpretation or measurement.

---

# 6. Restore recurrence

For restore exponent (k):

[
m=2^k.
]

Record:
- pre-restore bound;
- predicted bound:
  [
  widehat B^{rest}_j=m B_j;
  ]
- observed exact post-restore bound;
- Scale before/after.

Require:

[
B^{rest}_jle widehat B^{rest}_j.
]

The restore operation must not consume a logical Level.

---

# 7. Equivalence with production C2S helper

Run the ordinary current `CoeffsToSlotsNewWithRestorePlan` from the same group-0 input.

After the four-group DFT chain, before split/conjugate if practical, compare the audited manually stepped state with the combined helper:
- maintained rows exact;
- Level exact;
- Scale exact;
- metadata/domain exact.

This proves the instrumentation reproduces production order rather than an invented schedule.

Do not alter production code to make the comparison pass.

---

# 8. C2S decision classification

Choose exactly one:

## A — `QPREFIX_C2S_CAPACITY_PROVEN`

All four current-branch C2S groups have:
- exact raw pre-Rescale bounds;
- exact Rescale source/divisor/output recurrence;
- exact restore recurrence;
- strict Q-prefix-v2 capacity;
- production-helper equivalence.

This closes the C2S evidence gap from QPREFIX-AUDIT-001.

## B — `QPREFIX_C2S_CAPACITY_FAIL`

A required checkpoint has:

[
2Bge S_Q(ell).
]

Report first failing group/checkpoint and deficit.

## C — `QPREFIX_C2S_MEASUREMENT_CONFLICT`

The manually stepped path does not reproduce the combined helper, or the observed Rescale/restore result violates its proven recurrence.

## D — `QPREFIX_C2S_EVIDENCE_INCOMPLETE`

A required current-branch raw state cannot be reconstructed/observed without production changes.

Report the exact blocker.

---

# 9. Artifacts

Create in Primary:

`results/QPREFIX-AUDIT-002-C2S-RAW-RESCALE.json`

and concise:

`results/QPREFIX-AUDIT-002-summary.md`.

Include:
- exact provenance;
- q0..q3;
- four-group checkpoint table;
- Rescale recurrence table;
- restore table;
- production-helper equivalence;
- classification.

Do not dump full coefficient vectors.

---

# 10. Secondary changes

Preferred: add only a focused `*_test.go` diagnostic on `fast-qprefix` if necessary to expose exact internal states cleanly.

Production source must not change.

If no Secondary modification is needed and Primary can reproduce evidence safely, leave Secondary untouched.

Any useful diagnostic test may be committed/pushed to `fast-qprefix` after validation.

---

# 11. Validation

Required:
- focused C2S audit test;
- relevant Fast DFT/Bootstrap tests;
- accepted q0=56 public Bootstrap control;
- `git diff --check`;
- JSON parse.

The known unrelated Primary `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT` remains documented debt and must not be repaired in this task.

---

# 12. Prohibitions

Do not:
- modify production DFT/Bootstrap;
- modify `fast-ckks`;
- introduce F;
- implement q3 arithmetic;
- audit/fix EvalMod or S2C;
- retune compression/restore schedules;
- alter parameters;
- weaken the 1e-2 public control;
- use old fast-ckks C2S numbers as the authoritative answer.

This task closes only the current-branch C2S evidence chain.
