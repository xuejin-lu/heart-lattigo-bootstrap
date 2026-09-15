# FIX-001-P3-HANDOFF-RESUME-P93-PRODUCTION-LOCALIZE-AND-TIME

## Purpose

This is a **handoff/resume task for a new Codex chat**. The previous chat intentionally stopped with uncommitted Secondary production-candidate changes still present in the local `xuejin-lu/lattigo` worktree.

Do not reinterpret that dirty state as disposable work. Do not reset, stash, checkout-overwrite, clean, or discard it.

The immediate goals are:

1. take a provisional performance snapshot of the current fixed-width three-limb production candidate; and
2. localize the first **normal-baseline** semantic divergence between the accepted P93 reference and the dirty Secondary production path after PS.

Do not repair anything in this task. Stop after evidence is collected.

---

## Repositories / expected state

Primary:

`xuejin-lu/heart-lattigo-bootstrap`

Primary should be synchronized to current `origin/main`, then read:

- `AGENTS.md`
- `CURRENT_TASK.md`
- `docs/ORCHESTRATOR_HANDOFF.md`
- this spec

Secondary:

`xuejin-lu/lattigo`

Expected local branch/HEAD before the previous production attempt:

- branch: `fast-ckks`
- committed HEAD: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- **worktree intentionally dirty** with the uncommitted fixed-width Q012 / PS-wide P93 production candidate from the immediately preceding task.

### Dirty-worktree handoff rule

This task explicitly tells the new chat that the existing Secondary dirty changes are the object under diagnosis. Preserve them exactly.

On entering Secondary:

1. inspect branch, HEAD, and `git status --short`;
2. do **not** run reset/stash/clean/checkout-overwrite/rebase;
3. do **not** pull if doing so could overwrite or conflict with the dirty state;
4. summarize the dirty files/diff stat before running anything;
5. if the branch is not `fast-ckks`, HEAD is not `7d05f1f3...`, or the worktree is unexpectedly clean, stop and report a handoff-state mismatch;
6. otherwise treat the dirty state as intentional and continue read-only diagnosis/timing under this task.

No Secondary commit or push is authorized in this task.

---

## Current accepted architecture under test

Keep fixed:

- LogN13 only;
- q0 = 56-bit diagnostic/target profile;
- q1 ≈ 39 bits;
- q2 ≈ 40 bits;
- PS-wide Q012 authoritative arithmetic;
- planScale = `2^93`;
- q3+ are not arithmetic sources;
- Q012 -> Q01 contraction only at PS exit;
- downstream bounded local-q2 DoubleAngle behavior as required by the accepted diagnostic stack.

Do not tune q0, q1, q2, planScale, polynomial degree, decomposition, or workload in this task.

---

## Important evidence already established

### PS-wide design diagnostic

Primary diagnostic commit:

`41993cd03f85ec5f2de54dc21ade5ec812090d50`

Key results:

- P92 public-like ≈ `0.0162977678` -> fail;
- P93 public-like ≈ `0.0097053814` -> passes the `1e-2` system target;
- P94 first fails at `round0.after_multiplier_capacity`;
- Q012 max observed capacity ratio in the sweep ≈ `0.0002439022`;
- P92/P93 PS-exit Q01 contraction passed;
- the historical PS-local `1.2e-8` oracle budget is diagnostic, not the final system correctness gate.

### Dirty production-candidate status

Fixed-width Q012 / Rescale / truncate focused tests passed in the previous chat.

The previous public production attempt did **not** reproduce the accepted P93 full-system result.

Latest normal-baseline evidence reported by Codex:

- production baseline EvalMod real ≈ `0.004183019582623534` (below `1e-2`);
- production baseline public-like ≈ `0.2219142519401473` (fails badly).

A prior checkpoint comparison found:

- PS input: reference and production rows identical;
- T1: identical;
- T2: first row divergence, max semantic difference ≈ `6.940818919609626e-9`.

However, a later cumulative reference-power replacement experiment did **not** prove T2 causal. All replacement variants `{2}`, `{2,3}`, ..., `{2,3,4,6,8,16}` stopped at `final.restore_capacity` before a comparable downstream result. Therefore:

> Do not repair T2 merely because it is the first row difference.

The replacement-harness `final.restore_capacity` failure is also **not yet proven to be a normal production-baseline blocker**.

---

# Part A — provisional performance snapshot

The user now explicitly requires speed evidence. This task may run a **bounded provisional timing snapshot** even though correctness is not yet complete.

Compare on the same machine/config/input:

1. Standard `Bootstrap`;
2. current dirty Secondary Fast `Bootstrap`.

Requirements:

- same LogN13 q0=56 frontend/config and same deterministic input;
- 1 warmup;
- at least 5 measured runs per implementation;
- report median, min, max;
- report speedup = `Standard median / Fast median`;
- if existing low-overhead stage timing is already available, also report ScaleDown / ModUp / C2S / EvalMod / S2C / finalization and specifically PS/EvalMod time;
- do not add invasive instrumentation or refactor the hot path merely to get stage timings;
- do not benchmark diagnostic Big.Int reference code;
- clearly label all Fast timing as **provisional because correctness is not yet accepted**.

This timing snapshot must not be presented as correctness-preserving speedup.

---

# Part B — normal-baseline correctness localization

Do not use replacement ciphertexts or oracle power injection in the principal run.

Use the same matched C2S real/imag inputs and compare:

- accepted P93 diagnostic/reference semantics;
- current dirty Secondary production baseline.

Start at the end of PS; do not re-open T1/T2 unless evidence points back there.

Required checkpoints, in order:

1. final PS output while Q012-authoritative;
2. PS-exit Q012 -> Q01 contraction;
3. input to downstream restore/normalization;
4. each arithmetic sub-step of the relevant restore boundary;
5. each DoubleAngle round, at least:
   - round input/multiplier state;
   - square/multiply output;
   - add/sub recurrence correction;
   - Rescale/restore boundary;
6. EvalMod real;
7. EvalMod imag;
8. post-S2C;
9. public-like output.

At every checkpoint record only compact evidence:

- Level;
- Scale;
- Degree;
- q0/q1 hashes and q2 hash only while q2 is authoritative;
- max semantic difference reference vs production;
- centered-capacity ratio for the authoritative domain;
- metadata equality.

No full slot/coefficient arrays.

## `final.restore_capacity` question

Explicitly determine:

- does `final.restore_capacity` occur only in the earlier replacement harness because injected state changes the arithmetic path/capacity?
- or does the **normal production baseline** approach or cross the same capacity boundary?

Do not promote replacement-harness failure into a production blocker without this baseline evidence.

## Stop rule

Find the first meaningful reference/production divergence on the normal baseline path after PS and stop.

Do not fix it in this task.

If final PS output already differs materially, report that and stop before downstream analysis.

If PS output/contraction match but the first divergence appears in restore/DoubleAngle, identify the exact first sub-operation.

---

# Output

Return a compact report with exactly two main sections.

## A. Performance snapshot

- machine / Go version;
- Standard median/min/max;
- Fast median/min/max;
- Standard/Fast speedup;
- EvalMod/PS timing if available;
- note that Fast correctness is still provisional.

## B. Correctness localization

- first normal-baseline divergence;
- reference semantic value/error;
- production semantic value/error;
- Level / Scale / Degree;
- capacity ratio;
- row-hash/metadata evidence;
- whether `final.restore_capacity` is replacement-only or present in normal baseline.

No huge JSON artifact is required. A small temporary/compact evidence file is acceptable, but do not commit/push in this task.

---

# Prohibitions

- no parameter tuning;
- no T2 repair;
- no new power replacement sweep;
- no LogN16;
- no formal benchmark campaign beyond the bounded timing snapshot above;
- no Gate4/5;
- no EXP-003;
- no unrelated refactor;
- no commit/push;
- no reset/stash/discard of Secondary dirty changes.
