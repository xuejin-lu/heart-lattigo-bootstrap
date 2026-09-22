# FIX-001-P3-DIAG-P93-T3-BALANCED-VS-POSTPRODUCT-SCHEDULE-CAUSAL-AB

## Purpose

The preceding q01/q012 A/B at Primary commit

`48a42226182088167eebdc822b5ffbe00d629efa`

is accepted.

It proved:

1. current production q0/q1 rows after `MulIntegerMaintained` are identical to the q01-only shadow;
2. current Q012 Rescale and clean q01-authoritative Rescale produce the same balanced-operand semantic residual;
3. q01-only scaling + q01 Rescale also produces the same residual;
4. feeding q01-balanced operands back into T3 does not improve the T3 endpoint;
5. therefore the q01-vs-q012 scaling/rescale implementation difference is **not causal** for the current T3 regression.

A more fundamental schedule difference has now been confirmed.

Historical P93 design at Primary commit

`41993cd03f85ec5f2de54dc21ade5ec812090d50`

generated powers with:

[
	ext{multiply at full input scales}
ightarrow
	ext{relinearize}
ightarrow
	ext{Chebyshev recurrence}
ightarrow
	ext{single post-product Q012 Rescale}.
]

It did **not** use the current production balanced pre-Rescale schedule.

Current production T3 instead does:

[
T_2,T_1
ightarrow
	ext{balanced operand scaling}
ightarrow
	ext{Rescale each operand to about }2^{30}
ightarrow
	ext{multiply}
ightarrow
	ext{recurrence}.
]

The balanced operand Rescales introduce stable semantic residuals of roughly:

- real left: `1.77e-8`
- real right: `1.11e-7`
- imag left: `1.61e-8`
- imag right: `9.28e-8`

and the final T3 implementation residual is roughly:

- real `2.23e-7`
- imag `1.86e-7`.

Historical direct/post-product T3, by contrast, reaches approximately `1e-16` recurrence residual.

This task must causally determine whether the **balanced pre-Rescale schedule itself** is responsible for the T3 implementation regression.

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

No Secondary source modification, commit, or push is authorized.

Historical design reference:
- Primary commit `41993cd03f85ec5f2de54dc21ade5ec812090d50`
- Secondary clean commit `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Historical tooling may be replayed only in isolated temporary worktrees under `/tmp`.

---

## Fixed profile

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- P93 / planScale = `2^93`
- deterministic 4096-slot workload
- exact current production T1/T2 ciphertext inputs
- exact recurrence:
  [
  T_3=2T_2T_1-T_1
  ]
- polynomial-output budget:
  `3.716228023823462e-8`

No tuning.

---

# R0 — historical schedule provenance correction

Record explicitly:

## Historical P93 design generated-power schedule

From `fix001_p3_design_logn13_ps_wide_q012_q056_scale_sweep_runner.go` at commit `41993cd...`:

1. direct Q012 multiply;
2. relinearize;
3. Chebyshev doubling;
4. subtract difference power;
5. one Q012 Rescale at the end.

No balanced operand pre-Rescale.

## Current production schedule

From current dirty `circuits/ckks/polynomial/fast.go:fastPowerBasis.genPowerInternal`:

1. balanced factor pair;
2. scale each operand;
3. Rescale each operand;
4. Mul/MulRelin;
5. recurrence;
6. no final generated-power Rescale in the balanced branch.

State clearly:

> Historical P93 design accuracy is a valid design target, but its generated-power arithmetic is not the same implementation schedule as current production.

This corrects the earlier shorthand “production regression vs historical implementation”.

---

# R1 — reproduce current balanced T3 control

Using exact current production T1/T2:

Run the source-faithful current balanced T3 shadow.

Require:

- shadow == production T3 within deterministic floor;
- real implementation residual ≈ `2.2255e-7`;
- imag implementation residual ≈ `1.8575e-7`;
- balanced left/right residuals reproduce prior evidence.

Call this:

`A_balanced_current`.

---

# R2 — current-primitives post-product T3 shadow

From the **same exact current T1/T2 ciphertexts**, build a diagnostic-only T3 path that follows the historical schedule but uses current production Fast/Q012 primitives wherever possible.

Do not change Secondary.

Required logical sequence:

1. align T1/T2 to the same common level without balanced operand Rescale;
2. direct current Q012-capable Mul/MulRelin at full operand scales;
3. Chebyshev doubling;
4. subtract/alignment of T1 using the same recurrence semantics;
5. one current Q012 Rescale at the end.

Call this:

`B_postproduct_current_primitives`.

Record every stable boundary:

- direct product
- after relinearization if separate
- after doubling
- after T1 subtraction
- final post-product Rescale

For each record:

- Level/Scale/Degree
- semantic residual vs plaintext recurrence stage
- q01 centered-capacity ratio
- q012 centered-capacity ratio
- remaining q012 bits
- representation flags.

Important:
- q01 may approach its centered bound before final Rescale;
- Q012 is authoritative in this diagnostic;
- do not reject solely because q01 margin is small if Q012 centered uniqueness holds.

---

# R3 — historical big-int Q012 oracle shadow on the same current T1/T2 semantics

Build or reuse the historical diagnostic Q012 arithmetic semantics:

- Q012 multiply
- zero-secret relinearization semantics as applicable
- recurrence
- big-int centered Q012 rounded Rescale

Apply them to the same current T1/T2 semantic/input state as closely as the historical diagnostic representation allows.

Call this:

`C_postproduct_bigint_oracle`.

Purpose:

- validate the mathematical post-product schedule independent of fixed-width implementation;
- provide a reference if B fails.

Compare B vs C final T3 semantic output.

If exact ciphertext reuse across the old diagnostic harness is impossible, use the same current decoded T1/T2 values plus an equivalent Q012 diagnostic construction and state the representation limitation explicitly.

---

# R4 — schedule causal comparison

Compare final T3 implementation residuals:

[
E_A = |A-T_3^{oracle}|
]

[
E_B = |B-T_3^{oracle}|
]

[
E_C = |C-T_3^{oracle}|.
]

Strong schedule-causality criterion:

[
E_B le 0.1 E_A
]

and B remains q012-capacity safe.

Also report whether:

[
E_B le 3.716228023823462e-8.
]

Interpretation:

## If B and C both recover

Balanced pre-Rescale schedule is causal; current fixed-width Q012 primitives can implement the post-product schedule adequately.

## If C recovers but B does not

Post-product schedule is mathematically correct, but a current fixed-width Q012 primitive on the high-scale path remains defective.

## If neither B nor C recovers

Balanced schedule is not the main explanation.

---

# R5 — explain the precision mechanism

Quantify why the balanced schedule loses precision.

For current production, record:

- T1/T2 input scales (~(2^{60}));
- balanced operand output scales (~(2^{30}));
- right/left post-Rescale quantization residuals;
- product/output scale.

For post-product path, record:

- direct product scale (~(2^{120}));
- single final divisor (~(2^{60}));
- final T3 scale (~(2^{60})).

Explain numerically whether the observed `~1e-7` balanced-operand residual is consistent with reducing each operand to ~30 scale bits before multiplication.

Do not rely only on qualitative language.

---

# R6 — bounded T6 relevance check

Only if B recovers T3 by >=90%.

Using no production source changes, perform one bounded downstream relevance check:

- feed B's corrected T3 into the T6 recurrence/dependency branch;
- keep all other current production inputs/operations unchanged;
- compare resulting T6 semantic residual to current production T6.

Do not regenerate all powers or the full PS polynomial.

Purpose:
- confirm that T3 schedule correction propagates into its direct consumer;
- not to prove full-system sufficiency.

No replacement sweep.

---

# R7 — production-repair authorization

A later Secondary repair is authorized only if:

1. A reproduces current T3 residual;
2. B recovers T3 by >=90%;
3. B is Q012 centered-capacity safe;
4. C supports the same post-product semantics;
5. if R6 runs, T6 improves in the expected direction;
6. no parameter changes are required.

If all hold, state the smallest candidate repair scope:

> change generated-power scheduling in the bounded Q012-safe domain from balanced pre-Rescale to direct multiply / recurrence / post-product Rescale.

Do not implement in this task.

---

# Decision classification

Choose exactly one:

## A — `P93_T3_SCHEDULE_REPLAY_CONFLICT`

Current balanced or historical schedule evidence does not reproduce.

## B — `P93_T3_BALANCED_PRERESCALE_SCHEDULE_CAUSAL`

Current-primitives post-product path recovers >=90%, stays Q012-safe, and agrees with historical oracle.

## C — `P93_T3_POSTPRODUCT_SCHEDULE_VALID_BUT_FIXED_Q012_PRIMITIVE_FAILS`

Historical/big-int post-product path recovers, but current fixed-width post-product path does not.

Record first failing fixed-width primitive.

## D — `P93_T3_SCHEDULE_NOT_CAUSAL`

Balanced vs post-product schedule does not materially change T3 residual.

## E — `P93_T3_SCHEDULE_CAUSAL_BUT_T6_NOT_IMPROVED`

T3 recovers, but bounded T6 relevance check does not improve as expected.

## F — `P93_T3_POSTPRODUCT_ALIGNMENT_INVALID`

The post-product shadow cannot be aligned reliably enough for a causal claim.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-T3-BALANCED-VS-POSTPRODUCT-SCHEDULE-CAUSAL-AB-summary.json`

Keep <=300 pretty-printed JSON lines.

Include:

- provenance
- R0 schedule provenance correction
- R1 balanced control
- R2 current-primitives post-product path
- R3 historical/big-int post-product oracle
- R4 A/B/C residual comparison
- R5 precision mechanism
- conditional R6 T6 relevance
- R7 repair authorization
- one classification A-F
- explicit confirmation Secondary production was not modified.

No full vectors or coefficient dumps.

---

## Validation

Run:

- balanced T3 replay test
- current-primitives post-product T3 test
- Q012 capacity assertions
- historical/big-int post-product oracle test
- A/B/C causal comparison
- conditional bounded T6 relevance test
- directly affected Primary tests
- `go test ./...`
- `git diff --check`

On success commit/push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary production source modification
- no Secondary production commit/push
- no destructive operation on dirty Secondary
- no production schedule repair
- no generated-power sweep
- no full PS replacement
- no q/planScale tuning
- no DoubleAngle/C2S/S2C/finalizer work
- no P92/P94 sweep
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is a causal A/B of generated-power scheduling, not a fix.
