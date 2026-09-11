# FIX-001-DIAG-Q01 — Re-localize the first numerical divergence after FIX-001

## Status

READY once routed by `CURRENT_TASK.md`.

## Purpose

FIX-001 repaired the confirmed Fast Chebyshev q0/q1 capacity bug and produced repaired Fast commit:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Secondary focused regressions now pass, including formal-scale T2/T3 and degree-30 Mod1 tests, but the authoritative full LogN13 Bootstrap correctness gate still fails with approximately the same final amplitudes as before:

- max_abs_real ≈ `0.18750034197742593`
- max_abs_imag ≈ `0.06266213402955517`

The FIX-001 parent task therefore remains incomplete.

This subtask must answer exactly one question:

> With repaired Fast `87be78ff...`, where is the **new first supported numerical divergence** in the formal LogN13 Bootstrap pipeline?

Use the already-validated q0/q1 projection oracle from `EXP-002-C-DIAG-Q01`. Do not invent a new oracle and do not modify Lattigo in this task.

---

## Fixed provenance

### Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected remote base when this spec is authored:

`0a458f4381a368907e79f47e4899a2ab288b4c8b`

The parent task `FIX-001-fast-chebyshev-capacity-safe.md` remains **INCOMPLETE** even though this diagnostic subtask becomes the active `CURRENT_TASK`.

### Standard reference

Pinned Standard backend:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

### Repaired Fast backend

Pinned Fast backend:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Do not use old Fast `ce79b861...` as the active Fast backend except when reading historical evidence for comparison.

Follow `AGENTS.md` startup/safety rules in both repositories. If either worktree is dirty or cannot be synchronized safely, stop and report. Never reset/stash/discard unrelated work automatically.

---

## Fixed workload

Use exactly the same formal LogN13 workload as the earlier Q01 diagnosis:

- LogN = 13
- all 4096 slots
- `configs/bootstrap_config.logN13.json`
- deterministic input:
  - `real = ((i % 7) - 3) / 16`
  - `imag = ((i % 5) - 2) / 32`
- existing `ModUpThenEncode` circuit order
- same frontend runner/stage API
- same threshold:
  - `max_abs_real <= 1e-2`
  - `max_abs_imag <= 1e-2`

Do not run LogN16.

---

## Historical comparison baseline

The old Q01 diagnosis at Fast `ce79b861...` established:

- Q1 ModUp: exact match
- Q2 CoeffsToSlots real: PASS, max component ≈ `3.0981e-4`
- Q2 CoeffsToSlots imag: PASS, max component ≈ `2.6368e-4`
- Q3 EvalMod(real): first failure, max component ≈ `0.12023816407863477`
- classification: `FIRST_SUPPORTED_CAUSE = Q3_eval_mod_real/real`

FIX-001 was intended to repair one concrete cause inside this old Q3 failure.

The new result must explicitly compare repaired-Fast stage metrics to these historical values, but classification must be based only on the current repaired run.

---

## Reuse the validated q0/q1 projection oracle

Reuse the existing implementation/helper from `EXP-002-C-DIAG-Q01` wherever possible.

For any source ciphertext at logical Level >= 1, construct an isolated diagnostic level-1 ciphertext that:

- preserves degree;
- copies only q0/q1 from every existing ciphertext component;
- preserves source Scale and LogSlots/batching metadata;
- preserves NTT state;
- never reads dormant q2+ rows.

### Standard

- copy q0/q1 unchanged;
- decrypt with the Standard bootstrap secret;
- ordinary CKKS decode.

### Fast

- copy q0/q1 to the diagnostic copy;
- IMForm q0/q1 on the copy only when source is Montgomery;
- mark diagnostic copy non-Montgomery;
- preserve Scale exactly;
- decrypt with authoritative zero bootstrap secret;
- ordinary CKKS decode.

Prove the projection does not mutate the source ciphertext.

---

## Mandatory Standard oracle validation

For every Standard stage/branch used for causal comparison:

1. decode full-RNS Standard normally;
2. decode its q0/q1 level-1 projection;
3. compare them under the fixed `1e-2` component threshold.

A stage may classify Fast divergence only if its Standard projection oracle is valid.

If any required Standard projection oracle that previously passed now fails under the same pinned Standard/backend/frontend, classify:

`DIAGNOSTIC_ORACLE_REGRESSION`

and stop rather than attributing a Fast cause.

---

## Required checkpoints

Run Standard and repaired Fast through the same public stage flow.

### Q1 — ModUp

Record:

- Standard full vs Standard q01 oracle validity;
- Standard q01 vs repaired Fast q01.

### Q2 — CoeffsToSlots real

Same comparisons.

### Q2I — CoeffsToSlots imag

Same comparisons.

### Q3 — EvalMod(real)

Same comparisons.

This checkpoint is especially important because it was the first failing stage before FIX-001.

Record explicitly:

- old Fast max component: `0.12023816407863477`;
- repaired Fast max component;
- absolute improvement;
- whether repaired Q3(real) now passes `1e-2`.

### Q4 — EvalMod(imag)

Same comparisons.

Do not omit Q4 merely because Q3 used to fail. FIX-001 may have moved the first divergence to the imag branch or later.

### Q5 — SlotsToCoeffs

Compare Standard q01 vs repaired Fast q01 immediately after S2C and before unpack/finalization.

Also compare this q01 projection against the repaired Fast finalization-compatible view when available to confirm internal consistency.

### Q6 — final official Bootstrap output

Use the ordinary official public Bootstrap result, authoritative secrets, and all 4096 slots.

Record:

- repaired Fast zero-secret vs input;
- repaired Fast zero-secret vs Standard generated-secret final output;
- Standard generated-secret vs input.

This must reproduce the currently observed full-output failure unless an earlier rerun shows the prior failure was non-reproducible.

---

## Metrics

For every vector comparison record:

- sample_count
- max_abs_complex
- mean_abs_complex
- rmse_complex
- max_abs_real
- max_abs_imag
- max_component_abs
- max-error indices
- pass/fail at threshold `1e-2`

Archive all 4096 decoded values for every checkpoint/branch used in a causal classification.

Also record metadata:

- source logical Level
- projection Level (=1 where applicable)
- exact Scale string + log2
- Degree
- N
- LogSlots
- IsNTT
- IsMontgomery before normalization
- exact q0 and q1 moduli
- source/projection q0/q1 hashes for every component

---

## Parameter sanity

Before stage comparisons verify:

- Standard and repaired Fast exact q0/q1 moduli match;
- formal config is unchanged;
- repaired Fast commit is exactly `87be78ff...`;
- Standard commit is exactly `5dbffb...`.

If not, stop with the appropriate provenance/parameter mismatch instead of continuing.

---

## Classification logic

Select the **first pipeline checkpoint** whose Standard projection oracle is valid and whose Standard-q01 vs repaired-Fast-q01 comparison exceeds `1e-2`.

Allowed outcomes:

### Case A — Q1 fails

`FIRST_SUPPORTED_CAUSE = mod_up`

### Case B — Q1 passes, Q2/Q2I first fail

`FIRST_SUPPORTED_CAUSE = coeffs_to_slots`

Record whether real, imag, or both.

### Case C — Q1/Q2 pass, Q3 first fails

`FIRST_SUPPORTED_CAUSE = eval_mod_real`

Also report whether FIX-001 improved Q3 relative to old Fast and by how much.

### Case D — Q1-Q3 pass, Q4 first fails

`FIRST_SUPPORTED_CAUSE = eval_mod_imag`

### Case E — Q1-Q4 pass, Q5 first fails

`FIRST_SUPPORTED_CAUSE = slots_to_coeffs`

### Case F — Q1-Q5 all pass but Q6 official final fails

`FIRST_SUPPORTED_CAUSE = post_slots_to_coeffs_or_public_boundary`

This would contradict prior finalization-boundary evidence and must trigger a new narrow boundary diagnosis; do not guess.

### Case G — Q1-Q6 all pass

`FIRST_SUPPORTED_CAUSE = prior_fix_gate_nonreproducible`

Do not mark parent FIX-001 complete automatically. First reconcile the previously committed failed correctness evidence.

### Case H — Standard q01 oracle becomes invalid before localization

`FIRST_SUPPORTED_CAUSE = unresolved_projection_oracle`

Report the first unusable checkpoint.

---

## Interpretation requirement

The result must clearly distinguish:

1. **bug repaired by FIX-001** — old high-scale pre-Rescale Chebyshev correction capacity failure;
2. **new first divergence after FIX-001** — whatever this task identifies.

Do not roll back FIX-001 merely because another later bug remains.

If repaired Q3(real) now passes, explicitly state that FIX-001 successfully moved the first supported divergence downstream.

If repaired Q3(real) still fails but with a materially smaller error, state that FIX-001 removed one contribution but Q3 still contains another independent failure.

---

## Required artifacts

Write temporary run output to `/tmp` first while both repositories are clean.

Then create new artifacts without overwriting historical EXP or FIX evidence:

- `results/FIX-001-DIAG-Q01-logN13-standard.json`
- `results/FIX-001-DIAG-Q01-logN13-fast.json`
- `results/FIX-001-DIAG-Q01-logN13-summary.json`

The summary must contain:

- exact Standard/Fast commits;
- Standard q01 oracle validity by checkpoint/branch;
- repaired Standard-vs-Fast metrics Q1-Q5;
- authoritative Q6 final metrics;
- old-vs-repaired Q3(real) comparison;
- first supported cause;
- whether the first divergence moved downstream after FIX-001;
- clean-state/test evidence.

---

## Tests

Do not modify Secondary/Lattigo.

Run:

- Secondary repaired Fast `go test ./...` at `87be78ff...`;
- pinned Standard `go test ./...` at `5dbffb...`;
- Primary `go test ./...`.

If an existing Q01 helper needs a Primary-only adaptation to accept the repaired backend provenance, keep it backend-neutral and minimal. Do not import Fast internals into Primary.

---

## Stop conditions

Stop without repair if:

- repository state is unsafe;
- exact backend commits cannot be obtained cleanly;
- q0/q1 moduli/config differ unexpectedly;
- Standard projection oracle regresses;
- diagnostic result is nondeterministic;
- a required stage cannot be inspected without modifying Lattigo.

---

## Non-goals

Do not:

- modify or revert FIX-001 production code;
- implement a second fix;
- change parameters;
- lower Mod1 scale;
- add q2+;
- loosen the `1e-2` threshold;
- benchmark performance;
- run LogN16;
- start EXP-003.

---

## Completion

Mark this diagnostic subtask COMPLETE only when:

1. repaired Fast `87be78ff...` is used;
2. Standard q01 oracle is validated at every checkpoint used;
3. Q1-Q5 repaired stage comparisons are recorded;
4. Q6 authoritative final failure/pass is recorded;
5. old-vs-repaired Q3(real) is explicitly compared;
6. one classification is selected;
7. artifacts are committed/pushed normally;
8. Primary/Secondary end clean;
9. no Lattigo source changes, LogN16, performance work or EXP-003 occur.

The parent FIX-001 remains incomplete until a later task closes full LogN13 correctness.