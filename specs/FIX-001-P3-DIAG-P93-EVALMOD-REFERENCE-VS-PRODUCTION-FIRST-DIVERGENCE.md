# FIX-001-P3-DIAG-P93-EVALMOD-REFERENCE-VS-PRODUCTION-FIRST-DIVERGENCE

## Purpose

The preceding diagnostic commit `35b06d1c2ef67c1962f02971dcef225049fbc692` established strong current evidence that Fast S2C itself is not adding the observed post-S2C error:

- current dirty production F0 max component: `0.0402736269192105`;
- full-RNS/reference S2C fed with the same production EvalMod semantics gives the same output;
- Fast-S2C implementation effect: exactly `0` in the diagnostic;
- vector closure residual: `0`;
- linear-delta agreement residual: about `9.29e-12`.

Therefore do **not** modify or re-debug S2C in this task.

However, that preceding task did not satisfy its R0-A requirement literally: it loaded the already committed P93 design artifact instead of executing a fresh P93 reference replay. Its classification `P93_EVALMOD_ERROR_AMPLIFIED_BY_S2C` is directionally supported for the current production input, but the accepted P93 design path has not yet been freshly reconciled with the current production path.

The new authoritative production EvalMod comparison from that task is also materially worse than earlier rough evidence:

- real max component: `0.022317936759951508`;
- imag max component: `0.020569287028685726`.

Both already exceed the current **diagnostic/system milestone** `1e-2`.

The accepted historical P93 design path recorded:

- real EvalMod vs Standard: `0.0001635104343335875`;
- imag EvalMod vs Standard: `0.00013359983063118935`;
- post-S2C reference proxy: `0.0003032931504531125`;
- public-like: `0.009705381393898434` (passes the `1e-2` milestone).

Important: that historical P93 candidate did **not** pass every strict internal precision criterion; treat it only as a design/reference path that reached the `1e-2` system milestone, not as a proof of final Fast correctness.

This task has exactly two goals:

1. freshly replay the historical P93 reference methodology without touching the intentional dirty Secondary worktree;
2. locate the first stage-aligned semantic divergence between that freshly replayed P93 design/reference path and the current dirty production EvalMod path.

No repair in this task.

---

## Repository safety

### Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Start with the mandatory preflight in `AGENTS.md`, synchronize `main`, then read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

### Secondary production worktree

Repository: `xuejin-lu/lattigo`

Expected local production state:

- branch: `fast-ckks`
- committed HEAD: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- intentionally dirty
- expected deterministic dirty-diff SHA-256:
  `44b430bd7e148c919538c5487d95dbbc2d195054aaf15cae8b4a241d875bb46e`

Require exact match before doing anything else.

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter this dirty production worktree.

No Secondary production source modification, commit, or push is authorized.

---

## Fixed architecture / workload

Keep fixed:

- LogN13 only;
- q0 = 56-bit diagnostic/effective profile;
- q1 approximately 39 bits;
- q2 approximately 40 bits;
- PS-wide Q012 design;
- planScale = `2^93`;
- q3+ not arithmetic sources during PS;
- Q012 -> Q01 once at PS exit;
- same deterministic 4096-slot workload;
- same polynomial degree, K, DoubleAngle count, message ratio, circuit order;
- same C2S/EvalMod/S2C matrices and factorization;
- `1e-2` is the current **diagnostic system milestone only**, not the final research accuracy target.

Do not tune any parameter.

---

# R0 — fresh historical P93 reference replay

The historical design result was committed by:

- Primary reference commit:
  `41993cd03f85ec5f2de54dc21ade5ec812090d50`
- Secondary clean reference commit:
  `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

The original design runner is:

`fix001_p3_design_logn13_ps_wide_q012_q056_scale_sweep_runner.go`

The historical artifact is:

`results/FIX-001-P3-DESIGN-LOGN13-PS-WIDE-Q012-Q056-SCALE-SWEEP-summary.json`

## Required isolation method

Create temporary detached reference worktrees under a fresh `/tmp` directory, for example sibling paths:

- `/tmp/.../heart-lattigo-bootstrap` at Primary `41993cd...`
- `/tmp/.../lattigo` at Secondary `7d05f1f3...`

The temporary reference pair must not reuse or overwrite the dirty production Secondary working directory.

Using `git worktree add --detach` is allowed only for these new temporary paths. Do not modify the user's existing dirty Secondary worktree.

If the original runner would sweep P92/P93/P94, make only the smallest **temporary, uncommitted reference-worktree-only** harness change needed to execute P93 alone. Record that temporary harness diff in the result summary. Do not commit it anywhere.

Run the original P93 reference methodology, not merely load the historical JSON.

Record freshly measured:

- P93 real PS polynomial vs Standard;
- P93 imag PS polynomial vs Standard;
- P93 real EvalMod vs Standard;
- P93 imag EvalMod vs Standard;
- P93 post-S2C proxy;
- P93 public-like max component;
- PS-exit contraction status;
- capacity status;
- exact temporary Primary/Secondary refs and any temporary harness-only diff.

Historical expectations for consistency only:

- real EvalMod ~ `1.6351e-4`;
- imag EvalMod ~ `1.3360e-4`;
- post-S2C ~ `3.0329e-4`;
- public-like ~ `9.7054e-3`.

Require:

- public-like remains below `1e-2`;
- replay is numerically consistent with the historical P93 artifact;
- no production worktree state changes.

If not, stop with:

`P93_REFERENCE_REPLAY_CONFLICT`

Do not localize production until this passes.

---

# R1 — reproduce current dirty production EvalMod

Using the unchanged dirty production Secondary worktree, reproduce the current production path with the same deterministic workload.

Record:

- PS/EvalMod input semantics;
- production EvalMod real vs matched Standard;
- production EvalMod imag vs matched Standard;
- production post-S2C/F0 only as context.

Expected current magnitudes from commit `35b06d1...`:

- real ~ `0.02231793676`;
- imag ~ `0.02056928703`;
- F0 ~ `0.04027362692`.

If the current production EvalMod values or F0 materially fail to reproduce while branch/HEAD/fingerprint are unchanged, stop with:

`P93_PRODUCTION_REPLAY_CONFLICT`.

Do not use earlier `~0.00418/~0.00426` rough values as authoritative.

---

# R2 — stage-aligned reference vs production checkpoint trace

Compare the **freshly replayed P93 design/reference path** against the **current dirty fixed-width production path**.

The comparison is semantic. Do not require row identity between the Big.Int/design reference arithmetic and fixed-width production arithmetic unless row identity is mathematically required at that checkpoint.

Keep full vectors only under `/tmp`. Commit compact metrics only.

At minimum compare these checkpoints, when present in both paths:

1. C2S / EvalMod polynomial input;
2. PS entry;
3. generated powers T2, T3, T4, T6, T8, T16;
4. each PS baby-block output that feeds a giant step;
5. each giant-step aligned-add output;
6. final PS parent/root before final Rescale;
7. final PS Rescale output;
8. PS-exit Q012 -> Q01 contraction;
9. EvalMod polynomial output before DoubleAngle;
10. DoubleAngle round 0 output;
11. DoubleAngle round 1 output;
12. DoubleAngle round 2 / final EvalMod output.

For every aligned checkpoint report compactly:

- max-component semantic difference: production vs fresh P93 reference;
- max-complex difference;
- worst slot/component;
- reference max-component error vs Standard where available;
- production max-component error vs Standard where available;
- Level;
- Scale;
- Degree;
- Q012/Q01 capacity status where applicable;
- metadata equality where the two paths are expected to share metadata.

Do not commit full vectors, coefficient dumps, or RNS rows.

---

# R3 — identify first observable and first material divergence

Report two separate boundaries.

## First observable divergence

The earliest stage whose production-vs-reference semantic difference exceeds the measured replay/numerical floor.

Determine the floor from the fresh reference replay and deterministic repeatability; do not assume an arbitrary zero floor.

This boundary is descriptive only. Do **not** automatically treat T2 as causal merely because an earlier task saw a T2 row difference around `6.94e-9`.

## First material divergence

The earliest stage where the production-vs-reference semantic difference becomes large enough to explain a meaningful fraction of the final EvalMod gap.

Use the final real/imag EvalMod gap as the scale of interest and report the ratio explicitly. Use 10% of the final relevant EvalMod max-component gap as the materiality criterion unless the data justify a stricter threshold.

For each transition around the first material divergence record:

- incoming difference;
- outgoing difference;
- amplification factor;
- local capacity margin;
- Level/Scale/Degree before and after.

This task localizes; it does not prove a source-code line causal unless the evidence directly supports that claim.

---

# R4 — reconcile the 1e-2 milestone

Produce a compact table with exactly these rows:

1. fresh P93 design/reference public-like;
2. fresh P93 reference EvalMod real/imag;
3. current production EvalMod real/imag;
4. current production post-S2C/F0;
5. current official public output (context only).

For each row state:

- comparison reference;
- max-component error;
- pass/fail vs `1e-2` when semantically applicable;
- provenance.

The table must clearly show that `1e-2` is not currently end-to-end passed by production.

---

# Decision classification

Choose exactly one primary classification:

## A — `P93_REFERENCE_REPLAY_CONFLICT`

Fresh historical P93 reference replay does not reproduce the documented P93 system-milestone result.

Stop. Next task must repair reference/evidence methodology.

## B — `P93_PRODUCTION_REPLAY_CONFLICT`

Dirty production EvalMod/F0 does not reproduce current committed evidence despite unchanged provenance.

Stop. Next task must repair production-evidence reproducibility.

## C — `P93_PRODUCTION_INPUT_MISMATCH`

Production already differs materially from fresh P93 reference before PS begins.

Next task should localize the upstream C2S/normalization/input construction boundary.

## D — `P93_FIXED_WIDTH_PS_FIRST_MATERIAL_DIVERGENCE`

Inputs align, but the first material divergence occurs inside PS or at PS-exit contraction.

Record the exact earliest material checkpoint. No repair yet.

## E — `P93_POST_PS_EVALMOD_FIRST_MATERIAL_DIVERGENCE`

PS output is sufficiently aligned, but the first material divergence occurs after PS during downstream EvalMod / DoubleAngle.

Record exact round/checkpoint. No repair yet.

## F — `P93_NO_MATERIAL_REFERENCE_PRODUCTION_DIVERGENCE`

The production path remains aligned with the fresh P93 design reference through final EvalMod, contradicting the current error accounting.

Stop and classify the comparison harness/reference semantics as unresolved.

## G — `P93_STAGE_ALIGNMENT_INVALID`

Required checkpoints cannot be semantically aligned or the comparison conversion itself cannot be validated.

Stop. Fix the diagnostic methodology before production code.

---

## Required artifact

Create:

`results/FIX-001-P3-DIAG-P93-EVALMOD-REFERENCE-VS-PRODUCTION-FIRST-DIVERGENCE-summary.json`

Include only:

- current Primary provenance;
- dirty production Secondary provenance + exact fingerprint;
- temporary clean reference Primary/Secondary provenance;
- temporary reference harness-only diff, if any;
- R0 fresh replay metrics;
- R1 production replay metrics;
- compact checkpoint table;
- first observable divergence;
- first material divergence;
- five-row `1e-2` reconciliation table;
- one classification A-G;
- explicit confirmation that the user's dirty Secondary worktree was not modified.

Detailed vectors/traces remain under `/tmp`.

---

## Validation

Run:

- focused test/runner for the fresh P93 replay;
- focused test/runner for current production checkpoint comparison;
- directly affected Primary tests;
- `git diff --check`.

On successful diagnostic completion, commit and push only Primary diagnostic code + compact evidence according to `AGENTS.md`.

Temporary reference worktrees may be removed only if they were created by this task and removal does not touch the user's existing worktrees.

---

## Prohibitions

- no modification of dirty Secondary production source;
- no Secondary production commit/push;
- no reset/stash/clean/discard/rebase/checkout-overwrite of the dirty Secondary worktree;
- no S2C repair or further S2C attribution;
- no q0/q1/q2 tuning;
- no planScale tuning;
- no P92/P94 sweep except unavoidable original-runner setup before constraining the temporary harness to P93;
- no polynomial degree/K/DoubleAngle tuning;
- no T2 repair merely because it is an early observable mismatch;
- no PS repair;
- no generated-power replacement sweep;
- no public-finalization repair;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no benchmark campaign.

The deliverable is the first **material** reference-vs-production EvalMod divergence, not a fix.
