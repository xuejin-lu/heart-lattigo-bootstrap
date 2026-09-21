# FIX-001-P3-DIAG-P93-EVALMOD-PUBLIC-SCALE-RESET-CAUSAL-PROOF

## Purpose

The preceding diagnostic at Primary commit `5019a22b839ef832d804af9858f913ce2493fc8c` established a much narrower boundary than its broad classification name suggests.

From an identical production C2S input:

- actual public Fast EvalMod and forced P93 Fast EvalMod use the same effective `planScale = 2^93`;
- preprocessing matches;
- PS generated powers and maintained q0/q1/q2 rows match;
- PS arithmetic matches;
- all three DoubleAngle rounds match;
- at the internal `final_scale_reset`, the two paths match exactly:
  - coefficient row hashes equal;
  - Level equal;
  - Degree equal;
  - Scale equal to the original C2S input scale `2^50`;
  - decoded semantic difference = `0`.
- the first material divergence is only at the public wrapper scale reset:
  - coefficient row hashes remain equal;
  - actual public Fast EvalMod Scale becomes residual DefaultScale `2^45`;
  - forced P93 path retains the original EvalMod input Scale `2^50`;
  - direct Fast-vs-Fast decoded difference becomes:
    - real `0.018243003747759622`
    - imag `0.01657172300885934`.

Because the coefficient rows remain identical, the scale ratio is exactly the key clue:

`2^50 / 2^45 = 32`.

This task must prove or reject the causal hypothesis:

> The current production precision failure is caused by a metadata-only public EvalMod scale reset from the correct `2^50` semantic scale to `2^45`, while leaving coefficients unchanged.

This is a diagnostic causal proof only. Do **not** modify Secondary production source in this task.

---

## Repository safety

### Primary

Repository:
`xuejin-lu/heart-lattigo-bootstrap`

Follow `AGENTS.md` mandatory startup preflight, synchronize `main`, then read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

### Secondary

Repository:
`xuejin-lu/lattigo`

Required local state:

- branch: `fast-ckks`
- committed HEAD:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- exact expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty Secondary worktree.

No Secondary source modification, commit, or push is authorized.

---

## Fixed profile

Keep fixed:

- LogN13
- q0 = 56-bit effective profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- PS-wide Q012 production candidate
- planScale = `2^93`
- deterministic 4096-slot workload
- same Pack / ScaleDown / ModUp / C2S
- same EvalMod polynomial
- same three DoubleAngle rounds
- same production S2C / unpack / finalization
- current diagnostic milestone = `1e-2`

Do not tune any parameter.

---

# R0 — reproduce the exact scale-reset boundary

From one production C2S real/imag pair, run:

1. actual public `profile.Fast.EvalMod`;
2. source-faithful replay of the internal Fast EvalMod path through its internal final scale reset before the public wrapper override.

For both real and imag, prove:

## Internal pre-public state

At the internal final scale reset:

- Level equal;
- Degree equal;
- Scale = original C2S input scale;
- expected scale = `2^50 = 1125899906842624`;
- active q0/q1/q2 row hashes equal;
- decoded semantic difference = `0`.

## Public output state

After public `FastEvaluator.EvalMod` returns:

- Level unchanged;
- Degree unchanged;
- q0/q1/q2 row hashes unchanged relative to the internal state;
- only Scale metadata changes;
- actual public scale = residual DefaultScale `2^45 = 35184372088832`;
- exact scale ratio relative to internal/input scale = `32`.

If any coefficient row changes at this boundary, stop:

`P93_PUBLIC_SCALE_RESET_NOT_METADATA_ONLY`.

If the scale ratio or provenance does not reproduce, stop:

`P93_PUBLIC_SCALE_RESET_REPLAY_CONFLICT`.

---

# R1 — metadata-only counterfactual correction

Take the **actual public Fast EvalMod output** for real and imag.

Create diagnostic clones only.

On each clone perform exactly one change:

`corrected.Scale = original_C2S_input.Scale`

Expected corrected Scale:

`2^50`.

Do not:

- alter coefficients;
- multiply/divide coefficients;
- rescale;
- change Level;
- change Degree;
- change NTT/Montgomery flags;
- change q limbs.

Prove:

- before/after q0/q1/q2 row hashes are identical;
- every ciphertext coefficient is untouched;
- metadata differs only in Scale;
- corrected actual output vs forced P93 output:
  - max component <= `1e-10`, preferably exact zero if deterministic representation permits;
  - Level/Scale/Degree equal;
  - row hashes equal.

If metadata-only correction does not recover forced P93 semantics, classify:

`P93_SCALE_METADATA_CORRECTION_INSUFFICIENT`.

---

# R2 — EvalMod correctness after metadata-only correction

Compare corrected real/imag EvalMod outputs against the same matched P93 Standard objects used in the accepted diagnostic.

Record:

- real max-component error;
- imag max-component error;
- threshold status vs `1e-2`.

Historical forced-P93 expectations:

- real ≈ `0.004183019582623534`;
- imag ≈ `0.004261561142162278`.

The corrected actual public output should reproduce those values within deterministic/numerical tolerance.

Also record the uncorrected public-output comparison for contrast:

- real ≈ `0.022317936759951508`;
- imag ≈ `0.020569287028685726`.

Do not compare against a different Standard object in the same row.

---

# R3 — downstream production replay with only corrected metadata

This is the critical causal test.

Starting from the exact same production bootstrap input:

1. PackAndSwitchN1ToN2
2. ScaleDown
3. ModUp
4. production Fast C2S
5. production public Fast EvalMod real/imag
6. apply **only** the R1 metadata correction to each EvalMod output:
   `Scale := original corresponding C2S input Scale`
7. production Fast S2C
8. production unpack/switch back
9. existing finalization path

No other change is allowed.

Run in the same test/runner:

### Uncorrected control

Current production path unchanged.

Record:
- post-S2C/F0 error;
- final/public-facing error;
- pass/fail vs `1e-2`.

Expected current context:
- post-S2C/F0 ≈ `0.0402736269`;
- official public-facing max component ≈ `0.2219142513`.

### Metadata-corrected counterfactual

Record:
- EvalMod real/imag;
- post-S2C/F0;
- final/public-like max component;
- max complex;
- worst slot/component;
- Level/Scale/Degree at final output;
- pass/fail vs `1e-2`.

Primary acceptance question:

`Does metadata-only correction restore the full current LogN13 production pipeline to <= 1e-2?`

Also compare the corrected final result with the historical fresh P93 system proxy:

- historical P93 public-like ≈ `0.009705381393898434`.

The corrected path need not reproduce this value bit-for-bit unless the reference semantics are identical, but the difference must be explained and the same `1e-2` milestone applied consistently.

---

# R4 — vector causal closure

Use actual decoded vectors, not subtraction of max norms.

Let:

- `U` = uncorrected public EvalMod output vector;
- `C` = metadata-corrected EvalMod output vector;
- `F_U` = downstream production output from `U`;
- `F_C` = downstream production output from `C`.

Record:

- `C - U` at EvalMod output;
- `F_C - F_U` at post-S2C;
- `F_C - F_U` at final public output.

Because S2C is linear and prior work showed its implementation effect zero for the relevant path, record empirical amplification and verify the observed downstream change is consistent with the input semantic scale correction.

Do not claim exact linearity across nonlinear stages that are not present after EvalMod. The downstream path after EvalMod should be treated according to the actual operations performed.

---

# R5 — source audit of the public scale reset

Read the current dirty Secondary source, read-only.

Record:

- exact function/file containing the public EvalMod wrapper scale assignment;
- exact source expression used for the returned Scale;
- the input Scale captured before Mod1 evaluation;
- the internal Mod1 evaluator's final Scale before wrapper reset;
- residual DefaultScale;
- whether the public wrapper currently overwrites the internal restored/input Scale unconditionally;
- whether Standard/public API behavior provides any documented reason for this assignment.

Do not modify the source.

The artifact should state the smallest source-level candidate repair if the hypothesis is proven, but must not implement it.

---

# Decision classification

Choose exactly one:

## A — `P93_PUBLIC_SCALE_RESET_REPLAY_CONFLICT`

The previously observed `2^50 -> 2^45` public scale reset does not reproduce.

## B — `P93_PUBLIC_SCALE_RESET_NOT_METADATA_ONLY`

Rows or non-Scale metadata also change at the public boundary.

## C — `P93_SCALE_METADATA_CORRECTION_INSUFFICIENT`

Changing only Scale back to the original C2S/EvalMod input scale does not reproduce forced P93 semantics.

## D — `P93_SCALE_METADATA_CAUSAL_BUT_SYSTEM_STILL_FAILS_1E2`

Metadata correction reproduces forced P93 EvalMod semantics, but the corrected full downstream production pipeline still exceeds `1e-2`.

Record the next failing boundary; do not repair it.

## E — `P93_PUBLIC_EVALMOD_SCALE_RESET_CAUSAL_AND_1E2_SUFFICIENT`

All required conditions hold:

1. public boundary changes only Scale metadata;
2. q0/q1/q2 rows are unchanged;
3. ratio is `2^50 / 2^45 = 32`;
4. metadata-only correction reproduces forced P93 Fast output within <= `1e-10` or justified deterministic floor;
5. corrected EvalMod returns to the established ~`0.0042` regime;
6. corrected full production pipeline final/public-like metric is <= `1e-2`.

If E is selected, the next task should be a minimal Secondary repair of this scale-reset behavior plus regression validation.

## F — `P93_PUBLIC_SCALE_RESET_SOURCE_SEMANTICS_UNRESOLVED`

The numerical hypothesis is supported, but source/API semantics are ambiguous enough that production repair is not yet justified.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-EVALMOD-PUBLIC-SCALE-RESET-CAUSAL-PROOF-summary.json`

Include only:

- Primary provenance;
- Secondary branch/HEAD/dirty fingerprint;
- R0 boundary proof;
- R1 metadata-only correction proof;
- R2 corrected vs uncorrected EvalMod metrics;
- R3 downstream uncorrected/corrected comparison;
- R4 compact vector causal evidence;
- R5 source audit;
- one classification A-F;
- explicit confirmation that Secondary was not modified.

No full vectors or huge RNS dumps.

---

## Validation

Run:

- focused causal-proof test;
- directly affected Primary tests;
- `git diff --check`.

On completion, commit and push only Primary diagnostic code + compact evidence under `AGENTS.md`.

---

## Prohibitions

- no Secondary production source modification;
- no Secondary commit/push;
- no destructive operation on dirty Secondary;
- no coefficient correction;
- no q tuning;
- no planScale tuning;
- no PS repair;
- no C2S repair;
- no S2C repair;
- no finalizer repair;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign.

The only intervention allowed is a **diagnostic clone metadata-only Scale assignment** used to prove or reject causality.
