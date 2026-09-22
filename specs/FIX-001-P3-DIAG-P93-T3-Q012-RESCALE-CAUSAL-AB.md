# FIX-001-P3-DIAG-P93-T3-Q012-RESCALE-CAUSAL-AB

## Purpose

The preceding T3 primitive-localization task at Primary commit

`a4872ef14b4bde43aa6c8c307b9a7d6c2b533daf`

is accepted up to the following boundary:

- the source-faithful T3 shadow reproduces production T3 exactly;
- T1/T2 inputs are semantically correct;
- capacity is safe by many orders of magnitude;
- the first material semantic residual appears at the **right balanced operand post-Rescale**.

Real branch:
- left post-Rescale residual:
  `1.771577862186291e-8`
- right post-Rescale residual:
  `1.1105472362549218e-7`
- right/left amplification:
  ~`6.27x`

Imag branch:
- left post-Rescale residual:
  `1.606223765104886e-8`
- right post-Rescale residual:
  `9.276714673864261e-8`
- right/left amplification:
  ~`5.78x`

The T3 final implementation residual remains:

- real `2.2255463619225146e-7`
- imag `1.8575101814327644e-7`.

The task classified the boundary as operand-Rescale regression, but the required bounded `q01 vs q012` A/B was **not run**.

Current dirty source is known to include:

- q0/q1/q2 maintained integer scaling;
- q0/q1/q2 maintained multiplication;
- a Q012 rescale path at levels with three maintained limbs.

The clean committed `rescale.go` implementation is q01-authoritative and reconstructs/divides from q0/q1 only.

This task must causally separate:

1. q2-aware maintained integer scaling;
2. Q012 Rescale;
3. their interaction.

No production repair in this task.

---

## Repository safety

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary production:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed HEAD:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- required dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact branch/HEAD/fingerprint match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty production worktree.

No Secondary production source modification, commit, or push is authorized.

Clean reference semantics:
- committed Secondary `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- q01-authoritative `schemes/ckks/fast/rescale.go`

Any clean-reference execution must use isolated temporary worktrees under `/tmp`.

---

## Fixed profile

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012 production candidate
- P93 / planScale `2^93`
- deterministic 4096-slot workload
- same production T1/T2
- T3 recurrence:
  [
  T_3=2T_2T_1-T_1
  ]
- balanced factor pair at T3:
  [
  (2^{30},2^{30})
  ]
- same common level and divisor
- no parameter tuning

---

# R0 — reproduce the current operand boundary

Using the current source-faithful T3 shadow, reproduce for real and imag:

1. right-copy-before-factor
2. right-after-maintained-factor
3. right-post-current-Rescale
4. left equivalents for control

Require current right post-Rescale semantic residual to reproduce approximately:

- real `1.1105472362549218e-7`
- imag `9.276714673864261e-8`

and current T3 final implementation residual to reproduce.

If not, stop:

`P93_T3_Q012_AB_REPLAY_CONFLICT`.

---

# R1 — q0/q1 scaling identity check

Starting from one identical right-copy-before-factor clone, create:

## Scaling A — current production maintained scaling

Use the current dirty production `MulIntegerMaintained(2^30)`.

## Scaling B — diagnostic q01-only scaling shadow

In Primary diagnostic code only, reproduce the clean q01 modular integer-scaling semantics on q0/q1 and do not treat q2 as authoritative.

Do not modify Secondary.

Before any Rescale, compare Scaling A vs Scaling B on q0/q1:

- q0 row hashes
- q1 row hashes
- exact coefficient equality
- Scale metadata
- Level/Degree/representation flags

The semantic decode of this pre-Rescale state is not used for causality.

Classification use:

- if q0/q1 rows already differ, maintained scaling is a candidate causal source;
- if q0/q1 rows are exactly equal, q2-aware scaling has not corrupted the authoritative q01 rows before Rescale.

Also record q2 row for production context only.

---

# R2 — isolate Rescale using the same pre-Rescale ciphertext

Take the exact production Scaling-A pre-Rescale ciphertext and clone it.

Run two bounded paths:

## Path A — current Q012 Rescale

Use the current dirty production `Rescale`.

## Path B — diagnostic q01-authoritative Rescale

Use a Primary diagnostic shadow implementing the clean committed q01 Fast Rescale semantics from:

`schemes/ckks/fast/rescale.go`

at Secondary commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`.

It must:

- use only q0/q1 authoritative residues;
- reproduce the clean centered CRT + rounded division semantics;
- preserve NTT/Montgomery representation correctly;
- produce the same target Level and Scale metadata.

Compare both post-Rescale outputs against the unchanged right-input plaintext semantics (x_1).

Record:

- max component residual
- max complex
- mean complex
- worst slot/component
- Level/Scale/Degree
- q01 row hashes
- capacity

Primary causal test:

If Path B materially recovers the right operand while Path A reproduces ~`1e-7`, Q012 Rescale is causally implicated.

Use:

[
E_B le 0.1 E_A
]

as strong causal recovery, or justify a stricter numerical floor.

---

# R3 — full q01 scaling + q01 Rescale control

From the same right-copy-before-factor input:

1. apply Scaling B (q01-only diagnostic integer scaling);
2. apply Path B (q01-authoritative Rescale).

Call this Path C.

Compare Path C with Path B.

Interpretation:

## If B ≈ C and both recover

The q2-aware maintained integer scaling is not material; Q012 Rescale is the causal source.

## If B fails but C recovers

The production q2-aware maintained scaling creates a hidden interaction that only becomes semantic after Rescale.

## If A, B, C all fail similarly

The residual is not explained by q012-vs-q01 scaling/rescale selection.

No more variants are allowed.

---

# R4 — repeat the bounded A/B on the left operand

Repeat Paths A/B/C on the left balanced operand only to establish whether the same mechanism explains its smaller ~`1.6e-8..1.8e-8` residual.

Do not create additional variants.

Report whether:

- the mechanism is symmetric;
- right is worse only because of input/scale sensitivity;
- or left/right select materially different behavior.

---

# R5 — T3 endpoint counterfactual

Using the exact current T1/T2:

Construct at most these T3 shadows:

1. current production T3 control;
2. T3 with **right operand only** replaced by the best causally supported q01 post-Rescale path;
3. T3 with **both balanced operands** using that q01 path.

All later operations remain identical to production:

- MulRelin
- doubling
- subAligned(T1)

Compare each T3 endpoint against the exact plaintext oracle:

[
2T_2T_1-T_1.
]

Record implementation residual.

Key questions:

- Does right-only correction substantially reduce the `~2.2e-7` residual?
- Does both-operands correction reduce T3 implementation residual below:
  [
  3.716228023823462e-8
  ]
  or to <=10% of current residual?

This is a diagnostic counterfactual, not a production repair.

---

# R6 — dirty-source causal audit

Read the current local dirty Secondary source and compare only the executed scaling/rescale code against committed clean q01 behavior.

Record:

- exact dirty function selected for current right operand Rescale;
- whether it reconstructs from q0/q1/q2;
- divisor and rounding rule;
- how q2 participates;
- exact dirty `MulIntegerMaintained` limb behavior;
- any difference in Scale assignment.

Map source differences to R1–R5 evidence.

Do not identify a source line as causal unless the bounded A/B supports it.

---

# R7 — repair authorization criterion

Do not repair in this task.

A later minimal Secondary repair is authorized only if all hold:

1. production A reproduces the operand residual;
2. q01 B or C recovers the operand by >=90%;
3. T3 endpoint counterfactual correspondingly collapses the T3 implementation residual;
4. capacity remains safe;
5. no unrelated production parameter changes are required.

If these hold, state the smallest candidate repair scope, for example:

- select q01-authoritative Rescale for this safe balanced generated-power domain; or
- fix the proven q012 rounding/reconstruction defect.

Do not implement it yet.

---

# Decision classification

Choose exactly one:

## A — `P93_T3_Q012_AB_REPLAY_CONFLICT`

Current operand/T3 evidence does not reproduce.

## B — `P93_T3_Q012_RESCALE_CAUSAL`

Production q0/q1 scaling rows are correct, q01 Rescale recovers semantics, and T3 counterfactual improves accordingly.

## C — `P93_T3_Q2_SCALING_RESCALE_INTERACTION_CAUSAL`

Production scaling plus q01 Rescale does not recover, but q01 scaling + q01 Rescale does.

## D — `P93_T3_Q01_Q012_PATH_NOT_CAUSAL`

A/B/C do not materially change the residual.

## E — `P93_T3_Q012_CAUSAL_BUT_T3_STILL_FAILS`

q01 path fixes the operand residual but T3 implementation residual remains materially above budget due to another later primitive.

Record the next boundary.

## F — `P93_T3_Q01_Q012_AB_ALIGNMENT_INVALID`

The q01 shadow cannot be validated against clean semantics.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-T3-Q012-RESCALE-CAUSAL-AB-summary.json`

Keep <=300 pretty-printed JSON lines.

Include:

- provenance
- R0 replay
- R1 q0/q1 scaling identity
- R2 Path A/B
- R3 Path C
- R4 left control
- R5 T3 endpoint counterfactual
- R6 source audit
- R7 repair authorization status
- one classification A-F
- explicit statement Secondary production was not modified.

No full vectors or coefficient dumps.

---

## Validation

Run:

- current operand replay test
- q01 scaling identity test
- q01 clean-rescale shadow validation
- A/B/C operand test
- bounded T3 counterfactual test
- directly affected Primary tests
- `go test ./...`
- `git diff --check`

On success commit/push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no Rescale production repair
- no q2 production disablement
- no T3/T2 repair
- no generated-power replacement sweep
- no PS repair
- no q/planScale tuning
- no DoubleAngle/C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is a bounded causal A/B between current Q012 and clean q01 scaling/rescale semantics, not a fix.
