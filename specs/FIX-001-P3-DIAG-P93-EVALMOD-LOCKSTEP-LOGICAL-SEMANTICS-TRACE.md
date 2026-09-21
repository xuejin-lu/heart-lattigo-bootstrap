# FIX-001-P3-DIAG-P93-EVALMOD-LOCKSTEP-LOGICAL-SEMANTICS-TRACE

## Purpose

The preceding task at Primary commit `0c9987fa2ebaaf924e8a6b4cd85cafe1d53a3406` correctly reconfirmed:

- Fast and Genuine Standard C2S inputs align to about `1e-13`;
- final internal Fast-vs-Genuine-Standard EvalMod error is about:
  - real `0.004183019582623534`
  - imag `0.004261561142162278`;
- the shared public DefaultScale contract maps this to exactly 32x public error.

However, its claimed first material divergence at `pre_double_angle` is **not accepted**.

The runner paired evolving Fast DoubleAngle states against a Standard object that was left at the polynomial output. In addition, Fast normalized LogN13 uses a deferred maintained scalar represented by a virtual exponent, so raw decode after metadata-only coherent-scale changes is not in the same semantic coordinate system as Standard.

This caused nonphysical trace behavior such as:

[
5e-7 ightarrow 0.779 ightarrow 7e-7
]

across adjacent checkpoints.

Therefore this task must repair the **diagnostic methodology** and produce a true lockstep logical-semantics trace.

No production source repair is allowed.

---

## Repository safety

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed HEAD `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- expected dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact Secondary branch/HEAD/fingerprint match.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the dirty Secondary worktree.

No Secondary modification, commit, or push is authorized.

---

## Fixed evidence

Treat these as accepted controls:

- Genuine Standard exact E2E:
  `5.830057349387463e-8`
- C2S real/imag Fast-vs-Genuine Standard:
  approximately `1e-13`
- final internal pre-public Fast-vs-Genuine Standard:
  - real `0.004183019582623534`
  - imag `0.004261561142162278`
- public/internal error ratio:
  exactly `32`
- public diagnostic target:
  `1e-2`
- corresponding internal target:
  `3.125e-4`

No parameter tuning.

---

# R0 — prove the previous trace defect

Before building the corrected trace, record compact evidence that the preceding methodology was invalid:

1. identify the prior runner lines where Fast DoubleAngle evolves while the paired Standard object does not;
2. show that the previous `pre_double_angle` raw semantic comparison follows a metadata-only coherent-scale reassignment on Fast;
3. show the contradictory error pattern:
   - polynomial output ~`5e-7`
   - raw pre-DA ~`0.779`
   - next raw square ~`7e-7`.

Classify these old intermediate metrics as `invalid_raw_coordinate_comparison`.

Do not reuse the old classification `C`.

---

# R1 — source-faithful Standard lockstep trace

Run the Genuine Standard internal Mod1/EvalMod path source-faithfully.

Store actual Standard states for both real and imag at:

1. input
2. normalize scale
3. offset
4. polynomial input
5. polynomial output
6. DA0 after square
7. DA0 after double + constant
8. DA0 post-Rescale
9. DA1 after square
10. DA1 after double + constant
11. DA1 post-Rescale
12. DA2 after square
13. DA2 after double + constant
14. DA2 post-Rescale
15. internal final restore
16. public reset context

Every state must be produced by the actual corresponding Standard operation sequence.

Do not keep a static Standard polynomial object during later rounds.

Prove final Standard replay aligns with actual Genuine Standard internal/public EvalMod.

---

# R2 — source-faithful Fast trace with virtual-exponent state

Run the current dirty Fast normalized LogN13 path source-faithfully.

For each checkpoint record both:

- raw ciphertext metadata/decoded semantic value;
- the Fast logical virtual exponent (e) carried by the maintained-scalar schedule.

Required Fast logical-exponent semantics:

## Before coherent-scale normalization

At polynomial output:
- virtual exponent (e=0);
- raw decoded value is already logical value.

## Pre-DoubleAngle coherent scale

When Fast changes only metadata from working scale near (2^{33}) to coherent scale and sets:

[
e = 27,
]

the logical decoded value is:

[
v_{logical}=2^e v_{raw}.
]

## Square checkpoint

If incoming virtual exponent is (e), after raw square:

[
e_{square}=2e,
qquad
v_{logical}=2^{2e}v_{raw}.
]

Compare this with Standard **after square**, before Standard doubling.

## Fast multiplier + constant checkpoint

For:

[
a = 1+2e-e_{next},
]

Fast performs the maintained multiplier (2^a) and scaled constant such that the resulting raw state represents:

[
v_{logical}/2^{e_{next}}.
]

Therefore canonical logical comparison uses:

[
v_{logical}=2^{e_{next}}v_{raw}.
]

Compare it with Standard **after double + constant**.

## Post-Rescale

Virtual exponent remains (e_{next}).

Compare canonical logical Fast post-Rescale against Standard post-Rescale.

## Final maintained scalar restore

After Fast materializes the remaining (2^e) and restores caller input Scale, virtual exponent returns to:

[
e=0.
]

Raw decoded semantics are again directly comparable.

If source behavior differs from these equations, derive the correct equations from source and document them before comparison.

---

# R3 — canonical logical semantic metric

Implement a diagnostic-only helper that does **not modify ciphertext coefficients**.

Given decoded Fast raw vector (v) and virtual exponent (e), compute:

[
operatorname{CanonicalFast}(v,e)=2^e v.
]

For each aligned checkpoint compare:

[
E=|operatorname{CanonicalFast}-Standard|_{infty, component}.
]

Report only compact aggregate metrics:

- max component
- max complex
- mean complex
- worst slot/component
- Standard Scale
- Fast raw Scale
- Fast virtual exponent
- logical comparison valid yes/no
- Fast q012 capacity where relevant.

Do not commit full vectors or per-slot arrays.

Raw-coordinate metrics may be recorded only as debugging context and must be clearly labeled `not_semantic_for_causality`.

---

# R4 — true first observable and first material divergence

Use only canonical logical metrics from R3.

## First observable

Earliest aligned checkpoint above the measured deterministic/replay floor.

## First material

Earliest aligned checkpoint satisfying either:

1. >= 10% of the final internal Fast-vs-Standard gap; or
2. >= `3.125e-4`.

For the transition around the first material divergence report:

- incoming canonical error;
- outgoing canonical error;
- amplification;
- Standard operation;
- Fast operation;
- Fast virtual exponent before/after;
- Level/Scale/Degree;
- Fast capacity.

Do not classify a raw metadata-only coordinate change as a divergence.

---

# R5 — generated powers and polynomial boundary

Because the earlier production-vs-design trace observed small generated-power differences, explicitly report canonical/equivalent evidence for:

- T2
- T3
- T4
- T6
- T8
- T16
- final polynomial output

only if a Genuine Standard equivalent exists.

If generated powers are not directly semantically comparable because Standard planner uses a different internal decomposition, compare only the final polynomial output and mark powers `not_comparable`.

Do not force row/hash equivalence across different implementations.

---

# R6 — replay and endpoint invariants

The corrected lockstep trace is valid only if all of the following pass:

1. Standard replay final internal state aligns with actual Standard internal EvalMod.
2. Fast replay final internal state aligns with actual Fast internal EvalMod.
3. final canonical Fast-vs-Standard internal metric reproduces approximately:
   - real `0.004183019582623534`
   - imag `0.004261561142162278`.
4. applying the common public DefaultScale contract reproduces exactly 32x max-component error within numerical tolerance.
5. C2S precondition remains <= `1e-10`.

If any fail, classify trace as invalid and do not claim a causal stage.

---

# Decision classification

Choose exactly one:

## A — `P93_LOCKSTEP_TRACE_REPLAY_CONFLICT`

Standard or Fast source-faithful replay does not reproduce its actual implementation endpoint.

## B — `P93_LOGICAL_PREPROCESSING_DIVERGENCE`

First material canonical divergence is before polynomial evaluation.

## C — `P93_LOGICAL_POLYNOMIAL_DIVERGENCE`

First material canonical divergence is at polynomial/PS output.

## D — `P93_LOGICAL_DOUBLEANGLE_ROUND0_DIVERGENCE`

Polynomial output is within internal budget; first material canonical divergence appears in DA round 0.

## E — `P93_LOGICAL_DOUBLEANGLE_ROUND1_DIVERGENCE`

First material divergence appears in DA round 1.

## F — `P93_LOGICAL_DOUBLEANGLE_ROUND2_DIVERGENCE`

First material divergence appears in DA round 2.

## G — `P93_LOGICAL_INTERNAL_RESTORE_DIVERGENCE`

All DoubleAngle canonical checkpoints remain within budget; first material divergence appears only at final maintained scalar restore.

## H — `P93_LOGICAL_TRACE_NO_MATERIAL_LOCALIZATION`

Final internal gap reproduces, but no earlier checkpoint can be aligned finely enough to identify a first material operation.

## I — `P93_LOGICAL_TRACE_ALIGNMENT_INVALID`

Virtual-exponent/canonical semantics cannot be validated well enough for causal use.

---

## Artifact size requirement

Create:

`results/FIX-001-P3-DIAG-P93-EVALMOD-LOCKSTEP-LOGICAL-SEMANTICS-TRACE-summary.json`

The preceding summary was ~6500 lines and violated the project's compact-summary discipline.

This new summary must remain compact and human-reviewable:
- target <= 300 JSON lines;
- no full vectors;
- no coefficient arrays;
- no repeated large row-hash tables;
- at most one compact checkpoint record per logical stage.

Detailed temporary evidence belongs under `/tmp` and must not be committed.

---

## Validation

Run:

- focused lockstep logical-semantics test;
- Standard replay-vs-actual test;
- Fast replay-vs-actual test;
- canonicalization-equation unit tests for each virtual-exponent transition;
- directly affected Primary tests;
- `go test ./...`;
- `git diff --check`.

On successful completion, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

---

## Prohibitions

- no Secondary source modification
- no Secondary commit/push
- no production repair
- no metadata repair
- no coefficient correction
- no q/planScale tuning
- no P92/P94 sweep
- no C2S/S2C/finalizer work
- no threshold relaxation
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

The deliverable is a **valid causal lockstep trace in one logical semantic coordinate system**, not a fix.
