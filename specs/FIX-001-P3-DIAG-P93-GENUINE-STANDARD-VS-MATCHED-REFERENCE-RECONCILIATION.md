# FIX-001-P3-DIAG-P93-GENUINE-STANDARD-VS-MATCHED-REFERENCE-RECONCILIATION

## Purpose

The preceding diagnostic at Primary commit `28f16dbe665968a3febff33e9c107bb7c7cac9c2` established:

- EvalMod public Scale reset `2^50 -> 2^45` is a real metadata bug;
- metadata-only correction recovers forced-P93 EvalMod exactly;
- corrected EvalMod vs the **matched diagnostic Standard**:
  - real ≈ `0.004183019582623534`
  - imag ≈ `0.004261561142162278`
- corrected post-S2C/F0 vs the same matched Standard core:
  - ≈ `0.006934820352192803`, below `1e-2`;
- finalization contains a second metadata-only Scale reset that collapses corrected and uncorrected branches;
- preserving that downstream Scale still gives exact final E2E ≈ `0.1885754453731296`, far above `1e-2`.

This means the two metadata bugs are real and causal, but they do **not** explain the remaining E2E error.

A critical reference inconsistency is now visible:

[
	ext{Fast corrected S2C vs matched Standard core} approx 0.0069
]

while

[
	ext{Fast corrected exact E2E vs original message} approx 0.1886.
]

Therefore this task must determine whether the "matched Standard" objects used in P93 diagnostics are actually equivalent to a **Genuine Standard full-bootstrap execution from the same deterministic input**.

Do not repair Fast production code in this task.

---

## Repository safety

### Primary

Repository:
`xuejin-lu/heart-lattigo-bootstrap`

Follow `AGENTS.md` startup preflight, synchronize `main`, then read:

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

Use exactly:

- LogN13
- the same `BootstrapConfig`
- q0 = 56-bit effective Fast profile
- q1 ≈ 39 bits
- q2 ≈ 40 bits
- P93 / planScale = `2^93` for the Fast diagnostic branch
- deterministic 4096-slot workload
- the same reproducible input/message
- no parameter tuning
- current diagnostic milestone = `1e-2`

---

# R0 — fresh Genuine Standard baseline

Create a **genuine Standard evaluator**, not a Fast evaluator and not a matched helper.

Use the same deterministic input message and same bootstrap parameters.

Run both:

1. public Standard `Bootstrap(input)`;
2. a staged Standard pipeline:
   - PackAndSwitchN1ToN2
   - ScaleDown
   - ModUp
   - CoeffsToSlots
   - EvalMod real/imag
   - SlotsToCoeffs
   - UnpackAndSwitchN2ToN1
   - final public boundary / decoding.

Require the staged Standard pipeline and public Standard Bootstrap to agree structurally and semantically within deterministic numerical tolerance.

Record exact final E2E:

[
E_{max}=max_i{|Re(hat m_i-m_i)|, |Im(hat m_i-m_i)|}.
]

Historical context is around `5.83e-8`, but do not force this number. Fresh measurement is authoritative.

If Genuine Standard does not produce a small, self-consistent E2E baseline, stop:

`P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`.

---

# R1 — reconstruct the matched diagnostic Standard path

Reproduce exactly the Standard objects used by the current P93 matched/forced helpers, including the provenance of:

- matched/aligned C2S real/imag;
- matched Standard EvalMod real/imag;
- matched Standard S2C/core.

Do not silently substitute Genuine Standard objects.

For each object record:

- constructor/helper/function provenance;
- source input ciphertext;
- secret key used;
- Level/Scale/Degree;
- representation flags;
- decoded semantic summary.

The goal is to make the matched reference's input lineage explicit.

---

# R2 — stage-align matched Standard vs Genuine Standard

Compare only semantically equivalent stages.

Required comparisons:

1. C2S real:
   matched Standard vs Genuine Standard
2. C2S imag:
   matched Standard vs Genuine Standard
3. EvalMod real:
   matched Standard vs Genuine Standard
4. EvalMod imag:
   matched Standard vs Genuine Standard
5. S2C/core:
   matched Standard vs Genuine Standard
6. unpacked output, if a matched equivalent exists
7. final decoded message, if a matched equivalent exists.

For each comparison record:

- max-component difference;
- max-complex difference;
- mean-complex difference;
- worst slot/component;
- Level/Scale/Degree;
- whether comparison is semantically valid;
- exact input lineage for both sides.

If a stage is not semantically comparable, mark it `not_comparable` rather than forcing a metric.

---

# R3 — compare both Standard references to the original message

For every Standard reference that can legally be decoded in the original message domain, report exact E2E vs the deterministic original workload.

At minimum:

- Genuine Standard public final output vs original message;
- Genuine Standard staged final output vs original message;
- matched Standard final-like output if such an object actually exists;
- matched Standard S2C/core must **not** be called E2E unless it has been transformed back into the residual/public message domain correctly.

Explicitly distinguish:

- stage fidelity metric;
- proxy/system metric;
- exact E2E metric.

---

# R4 — current Fast corrected branch vs Genuine Standard

Use the already-proven diagnostic-only Fast metadata corrections:

1. EvalMod public output Scale restored to the original C2S input Scale;
2. downstream finalization Scale preserved across the proven metadata reset.

Do not change coefficients or Secondary source.

Compare this Fast corrected branch against **Genuine Standard**, not matched Standard, at:

- C2S input/output;
- EvalMod real/imag;
- S2C/core;
- unpack output;
- final decoded message.

Record first observable and first material Fast-vs-Genuine-Standard divergence.

Materiality criterion:
10% of the final Fast-corrected-vs-Genuine-Standard E2E gap, unless a stricter evidence-grounded criterion is justified.

Do not reopen PS internals unless the first material boundary proves that EvalMod arithmetic must be revisited.

---

# R5 — reference-lineage audit

Read the relevant Primary helper code and report:

- where `profile.Inputs.OrdinaryReal/OrdinaryImag` (or equivalent matched objects) originate;
- whether they originate from the exact same original production input used by the Genuine Standard staged run;
- whether normalization, compressed C2S, synthetic splitting, direct object reuse, or another helper changes the lineage;
- why the matched reference was valid for its original diagnostic purpose;
- whether it is valid or invalid as an exact E2E oracle.

This source audit is required before classification.

---

# Decision classification

Choose exactly one:

## A — `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`

Fresh Genuine Standard baseline is not self-consistent or does not reproduce a small exact E2E error.

Stop. Repair baseline methodology first.

## B — `P93_MATCHED_STANDARD_INPUT_LINEAGE_MISMATCH`

Matched Standard and Genuine Standard differ materially already at the C2S input/output lineage because they do not originate from the same exact Standard full-bootstrap input path.

## C — `P93_MATCHED_STANDARD_EVALMOD_REFERENCE_MISMATCH`

C2S semantics align, but matched vs Genuine Standard first diverge materially in EvalMod/reference normalization.

## D — `P93_MATCHED_STANDARD_S2C_REFERENCE_MISMATCH`

Matched and Genuine Standard align through EvalMod but diverge materially at S2C/core construction.

## E — `P93_MATCHED_REFERENCE_VALID_BUT_FAST_DIVERGES_FROM_GENUINE_STANDARD`

Matched Standard is a valid equivalent reference, but the metadata-corrected Fast branch still materially diverges from Genuine Standard at a specifically identified Fast stage.

Record the first material Fast stage.

## F — `P93_MATCHED_REFERENCE_PROXY_NOT_VALID_FOR_E2E`

Matched Standard is valid for local stage fidelity diagnostics but is not a valid exact E2E oracle. The apparent `0.0069` pass must be retained only as a proxy/stage metric.

## G — `P93_REFERENCE_RECONCILIATION_UNRESOLVED`

Required Standard stages cannot be aligned reliably enough to support a causal classification.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-GENUINE-STANDARD-VS-MATCHED-REFERENCE-RECONCILIATION-summary.json`

Keep it compact. Include only:

- Primary provenance;
- Secondary branch/HEAD/dirty fingerprint;
- fresh Genuine Standard exact E2E;
- staged-vs-public Standard agreement;
- matched-reference lineage;
- compact matched-vs-Genuine checkpoint table;
- exact-E2E/proxy metric distinction;
- corrected Fast-vs-Genuine checkpoint table;
- first material divergence;
- source-lineage audit;
- one classification A-G;
- explicit confirmation Secondary was not modified.

No full vectors or large coefficient dumps.

---

## Validation

Run:

- focused reference-reconciliation test;
- Genuine Standard public-vs-staged consistency test;
- directly affected Primary tests;
- `go test ./...`;
- `git diff --check`.

On successful completion, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification;
- no Secondary commit/push;
- no destructive Git operation on dirty Secondary;
- no new Fast production repair;
- no coefficient correction;
- no q/planScale tuning;
- no PS repair;
- no C2S repair;
- no S2C repair;
- no finalizer repair;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign.

The deliverable is a trustworthy Standard reference hierarchy and the first valid Fast-vs-Genuine-Standard divergence, not a repair.
