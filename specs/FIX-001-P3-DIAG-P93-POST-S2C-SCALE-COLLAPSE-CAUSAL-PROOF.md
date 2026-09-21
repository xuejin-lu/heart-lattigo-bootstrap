# FIX-001-P3-DIAG-P93-POST-S2C-SCALE-COLLAPSE-CAUSAL-PROOF

## Purpose

The preceding causal-proof task at Primary commit `00bb90c25137a2d570ac11c728d081fbc96fc125` established the following as accepted evidence:

1. The public Fast EvalMod wrapper performs a metadata-only scale reset:
   - internal/input-restored Scale = `2^50`
   - public wrapper Scale = residual DefaultScale `2^45`
   - active q0/q1/q2 rows unchanged
   - exact scale ratio = 32.

2. A diagnostic metadata-only correction:
   `corrected.Scale = original_C2S_input.Scale`
   reproduces the forced-P93 Fast EvalMod output exactly:
   - real corrected-vs-forced = 0
   - imag corrected-vs-forced = 0.

3. The corrected EvalMod returns to the accepted `1e-2` regime:
   - real vs matched Standard = `0.004183019582623534`
   - imag vs matched Standard = `0.004261561142162278`.

4. The correction remains effective through S2C:
   - uncorrected post-S2C/F0 vs matched Standard = `0.0402736269192105`
   - corrected post-S2C/F0 vs matched Standard = `0.006934820352192803` (passes `1e-2`).

5. However, after the remaining downstream path, corrected and uncorrected final outputs become identical:
   - corrected final public-like = `0.2219142519401473`
   - uncorrected final public-like = `0.2219142519401473`
   - corrected-minus-uncorrected final vector = exactly 0.

Therefore the EvalMod scale bug is proven causal through S2C, but a later downstream boundary collapses the corrected and uncorrected semantic paths back onto the same metadata/interpretation.

This task must identify the **first downstream boundary after S2C where corrected and uncorrected scale semantics converge**, prove whether that convergence is metadata-only, and test a metadata-only counterfactual through the final output.

No Secondary production source modification is allowed.

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
- PS-wide Q012
- planScale = `2^93`
- deterministic 4096-slot workload
- production Pack / ScaleDown / ModUp / C2S
- current public Fast EvalMod
- the already-proven diagnostic EvalMod metadata correction only
- production S2C
- production unpack/switch-back
- current finalization path
- diagnostic milestone = `1e-2`

No parameter tuning.

---

# R0 — reproduce corrected and uncorrected branches through S2C

From the same production bootstrap input, produce:

## Branch U — uncorrected

- production Fast EvalMod public outputs unchanged
- production S2C

Expected post-S2C/F0 vs matched Standard:
`~0.0402736269192105`.

## Branch C — corrected

- same public Fast EvalMod outputs
- diagnostic clone metadata-only correction:
  `Scale := original corresponding C2S input Scale`
- production S2C

Expected post-S2C/F0 vs matched Standard:
`~0.006934820352192803`.

Before continuing, record for both branch C and U S2C outputs:

- Level
- Scale
- Degree
- NTT/Montgomery flags
- q0/q1/q2 row hashes where available
- decoded semantic difference C-vs-U

If these two branches no longer reproduce the committed evidence, stop:

`P93_POST_S2C_REPLAY_CONFLICT`.

---

# R1 — trace the downstream collapse boundary

Trace Branch C and Branch U from the S2C output through every downstream operation that can affect ciphertext metadata or coefficients.

At minimum capture:

1. S2C/core output
2. input to `UnpackAndSwitchN2ToN1`
3. immediate output of `UnpackAndSwitchN2ToN1`
4. each ring-degree / switch / conversion step inside or immediately after unpack that is observable without modifying Secondary
5. input to the finalization/oracle helper used by the production-facing test
6. every explicit scale assignment or scale-normalization step in that finalization helper
7. final ciphertext before decode
8. decoded final vector

For each checkpoint record C-vs-U:

- max-component semantic difference
- max-complex difference
- Level
- Scale
- Degree
- NTT/Montgomery flags
- row-hash equality for all maintained/meaningful rows
- whether metadata is equal
- whether coefficient rows are equal.

The primary boundary is:

> the first checkpoint where C and U were semantically different immediately before, but become semantically identical or their Scale metadata becomes equal in a way that erases the correction.

Do not skip a boundary merely because coefficient rows are equal; this task is specifically about metadata semantics.

---

# R2 — classify the first convergence as metadata-only or arithmetic

At the first convergence boundary, prove one of:

## Metadata-only convergence

All of the following:
- coefficient row hashes before/after are unchanged as required by the operation;
- C and U coefficient rows are equal where expected;
- the relevant difference is Scale or other metadata only;
- changing only the relevant Scale metadata on a diagnostic clone can preserve the corrected semantic branch.

## Arithmetic convergence

At least one of:
- coefficient rows change differently between C and U;
- arithmetic/switching maps both branches to the same coefficients independent of prior Scale;
- a metadata-only counterfactual cannot preserve the corrected branch.

If arithmetic convergence is observed, stop after identifying the exact operation; do not repair it.

---

# R3 — metadata-only downstream counterfactual

If R2 proves a metadata-only convergence, create diagnostic clones only.

Do not change coefficients.

At the exact first convergence boundary, preserve/restore the **corrected branch's semantically appropriate incoming Scale** across the boundary.

Important:
- do not blindly hardcode `2^50` unless the traced corrected branch mathematically retains `2^50` at that exact boundary;
- use the corrected branch's actual pre-overwrite Scale/semantic scale as the counterfactual value;
- do not alter Level, Degree, rows, or representation flags.

Then continue the normal downstream path.

Record:

- row hashes before/after correction
- metadata before/after
- corrected counterfactual vs ordinary corrected branch
- corrected counterfactual vs uncorrected branch.

---

# R4 — exact final E2E test

For the metadata-preserved downstream counterfactual, compare final decoded output directly against the original deterministic message/workload using the same exact E2E metric used for the Genuine Standard baseline:

[
E_{max} = max_i { |Re(hat m_i-m_i)|, |Im(hat m_i-m_i)| }.
]

Record:

- max-component E2E
- max-complex
- worst slot/component
- final Level/Scale/Degree
- pass/fail vs current diagnostic milestone `1e-2`.

Also record controls:

1. uncorrected current final E2E
2. EvalMod-corrected-only final E2E
3. EvalMod-corrected + downstream-metadata-preserved final E2E.

Do not call a post-S2C-vs-Standard proxy an E2E metric.

If the fully metadata-preserved branch reaches <= `1e-2`, state this clearly.

Do not yet use the later final research target (~`1e-7`) as the acceptance gate for this task.

---

# R5 — source audit

Read the relevant dirty Secondary and Primary finalization/unpack code read-only.

Record:

- exact function/file for the first convergence boundary
- exact source expression that overwrites or derives Scale
- incoming corrected Scale
- outgoing Scale
- whether Standard uses the same semantic contract at the analogous boundary
- whether the current operation intentionally treats Scale as metadata-only or recomputes it from a fixed default
- smallest candidate production repair if causality is proven.

Do not implement any repair.

---

# Decision classification

Choose exactly one:

## A — `P93_POST_S2C_REPLAY_CONFLICT`

Corrected/uncorrected S2C boundary no longer reproduces committed evidence.

## B — `P93_POST_S2C_ARITHMETIC_COLLAPSE`

The first corrected-vs-uncorrected convergence is caused by coefficient-changing arithmetic/switching, not metadata-only behavior.

## C — `P93_UNPACK_SCALE_METADATA_COLLAPSE`

The first causal convergence occurs in or immediately at `UnpackAndSwitchN2ToN1`, and is metadata-only.

## D — `P93_FINALIZATION_SCALE_METADATA_COLLAPSE`

Unpack preserves the corrected distinction; the first metadata-only convergence occurs in the subsequent finalization path.

## E — `P93_MULTIPLE_DOWNSTREAM_SCALE_RESETS`

More than one distinct metadata reset must be preserved to keep the corrected semantics through final output.

## F — `P93_DOWNSTREAM_METADATA_CAUSAL_BUT_SYSTEM_STILL_FAILS_1E2`

The first metadata convergence is proven and counterfactually corrected, but exact final E2E remains > `1e-2`.

Record the next failing boundary without repairing it.

## G — `P93_EVALMOD_PLUS_DOWNSTREAM_METADATA_SUFFICIENT_FOR_1E2`

All required metadata-only corrections are causally proven and exact final E2E reaches <= `1e-2`.

If G is selected, the next task may implement the smallest production repair(s) in Secondary plus regression validation.

## H — `P93_POST_S2C_TRACE_ALIGNMENT_INVALID`

The corrected and uncorrected paths cannot be traced/aligned reliably enough to make a causal claim.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-POST-S2C-SCALE-COLLAPSE-CAUSAL-PROOF-summary.json`

Include only:

- Primary provenance
- Secondary branch/HEAD/dirty fingerprint
- R0 replay
- compact downstream checkpoint table
- first convergence boundary
- R2 metadata-vs-arithmetic proof
- R3 counterfactual
- R4 exact E2E controls
- R5 source audit
- one classification A-H
- explicit confirmation that Secondary was not modified.

No full slot vectors or large coefficient dumps.

---

## Validation

Run:

- focused downstream-collapse causal test
- directly affected Primary tests
- `go test ./...`
- `git diff --check`

On successful completion, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification
- no Secondary commit/push
- no destructive operation on dirty Secondary
- no coefficient correction
- no q/planScale tuning
- no PS repair
- no C2S repair
- no S2C repair
- no production finalizer repair
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

Only diagnostic clone metadata preservation at the proven downstream convergence boundary is allowed.
