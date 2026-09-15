# Fast-CKKS Codex Handoff

This is the **current operational handoff for starting a fresh Codex chat**.

For historical research context, `docs/ORCHESTRATOR_HANDOFF.md` still exists, but it is stale for the present PS-wide Q012/P93 production work. New Codex chats should read this file first for the current state.

## Repositories

Primary:

`xuejin-lu/heart-lattigo-bootstrap`

Secondary:

`xuejin-lu/lattigo`

Secondary active branch:

`fast-ckks`

## Current architecture being pursued

The active LogN13 Fast-CKKS design is now:

- q0 = 56-bit profile;
- q1 approximately 39 bits;
- q2 approximately 40 bits;
- full Paterson-Stockmeyer polynomial phase is Q012-authoritative;
- planScale = `2^93`;
- q3+ are not arithmetic sources;
- Q012 -> Q01 happens once at PS exit after centered-uniqueness proof;
- downstream DoubleAngle may use bounded local-q2 windows where already validated;
- current goal is speed-first functional correctness, not noise fidelity.

Do **not** reopen the old q0=55/local-G0/F0/T2-patch design unless new evidence explicitly forces it.

## Why the architecture changed

Earlier q0/q1-only PS arithmetic repeatedly crossed centered CRT capacity. Local q2 patches became increasingly complex.

A PS-wide Q012 design sweep showed:

- P92 public-like ≈ `0.0162977678` (fails 1e-2);
- P93 public-like ≈ `0.0097053814` (passes 1e-2);
- P94 first fails at `round0.after_multiplier_capacity`;
- maximum observed Q012 capacity ratio ≈ `0.0002439022`;
- P92/P93 PS-exit Q01 contraction passed.

Therefore P93 is the current production candidate.

The historical local PS oracle threshold `1.2e-8` is now diagnostic evidence only, not the final system correctness gate.

## Important committed Primary references

PS-wide Q012 scale sweep result commit:

`41993cd03f85ec5f2de54dc21ade5ec812090d50`

Production-integration spec commit sequence began from Primary:

`8ca4d13eeaaafd6fea2d8277558d197ac224848b`

Current task should always be read from the latest remote `CURRENT_TASK.md` after synchronizing Primary.

## Current Secondary state: intentionally dirty

The previous Codex chat stopped with **uncommitted production-candidate changes intentionally preserved** in the local Secondary worktree.

Expected committed Secondary base:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Expected branch:

`fast-ckks`

The dirty changes include the current fixed-width Q012 / PS-wide P93 production implementation attempt.

### Critical handoff rule

A fresh Codex chat must **not** reset, stash, clean, discard, checkout-overwrite, rebase, or otherwise destroy these Secondary dirty changes.

The active handoff spec explicitly authorizes read-only inspection/diagnosis of this intentional dirty state.

If Secondary is unexpectedly clean, on the wrong branch, or not based on `7d05f1f3...`, stop and report the state mismatch.

## Latest correctness evidence

The current dirty production implementation does not yet reproduce the accepted P93 public Bootstrap result.

Latest normal-baseline evidence:

- production baseline EvalMod real ≈ `0.004183019582623534` (passes 1e-2);
- production baseline public-like ≈ `0.2219142519401473` (fails badly).

A previous checkpoint comparison found:

- PS input: reference vs production identical;
- T1: identical;
- T2: first row divergence, semantic difference ≈ `6.940818919609626e-9`.

However T2 is **not proven causal**.

A cumulative reference-power replacement experiment tested:

- `{2}`
- `{2,3}`
- `{2,3,4}`
- `{2,3,4,6}`
- `{2,3,4,6,8}`
- `{2,3,4,6,8,16}`

None produced a comparable downstream public result because every replacement run stopped at `final.restore_capacity` before EvalMod.

Therefore:

- do not repair T2 merely because it is the first row mismatch;
- do not assume `final.restore_capacity` is a real normal-baseline blocker until verified on the ordinary production path;
- the next localization should start from PS output and move downstream on the normal baseline path.

## Current speed status

Historical clean Fast performance before this correctness investigation was very strong:

- LogN13 roughly 12.55x Fast vs Standard in an earlier end-to-end experiment;
- stage-breakdown LogN13 roughly Fast 24.3 ms vs Standard 303.6 ms.

These are historical performance numbers only and are **not** proof that the new three-limb PS-wide P93 production candidate retains the same speedup.

The user has now explicitly requested a fresh provisional timing snapshot of the current dirty Q012 production candidate.

## Immediate next task

Read the current remote spec referenced by `CURRENT_TASK.md`.

At the time this handoff was updated, the intended task is:

`specs/FIX-001-P3-HANDOFF-RESUME-P93-PRODUCTION-LOCALIZE-AND-TIME.md`

Its two goals are:

1. provisional speed snapshot:
   - Standard Bootstrap vs current dirty Fast Bootstrap;
   - same LogN13 q0=56 config/input;
   - 1 warmup + at least 5 measured runs;
   - median/min/max + Standard/Fast speedup;
   - stage timing only if already available with low overhead;
   - clearly label Fast timing provisional because correctness is not accepted;

2. normal-baseline correctness localization:
   - compare accepted P93 reference vs dirty Secondary production path;
   - begin at final PS output and proceed through PS-exit contraction, downstream restore, DoubleAngle rounds, EvalMod, S2C, public output;
   - find the first meaningful baseline divergence;
   - explicitly determine whether `final.restore_capacity` is replacement-harness-only or also present on normal baseline;
   - stop after localization; do not fix.

## Do not do in the next chat

- do not tune q0/q1/q2;
- do not change planScale from 2^93;
- do not repair T2 before causal proof;
- do not run a new power replacement sweep;
- do not run LogN16;
- do not run Gate4/5 or EXP-003;
- do not commit or push Secondary during the handoff localization task;
- do not destroy the dirty Secondary worktree.

## How a fresh Codex chat should start

After opening a fresh Codex chat, tell it:

> Start from the current Fast-CKKS handoff. First synchronize the Primary repository safely, then read `AGENTS.md`, `docs/CODEX_HANDOFF.md`, `CURRENT_TASK.md`, and the referenced task spec. The Secondary `fast-ckks` worktree is intentionally dirty from the previous production attempt; do not reset/stash/discard it. Follow the handoff spec exactly and stop at its first stop condition.

Do not paste old conversational context unless the new chat explicitly needs something not present in these files.
