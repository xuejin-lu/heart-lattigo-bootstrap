# FIX-001-P3-DIAG-PUBLIC-FINALIZATION

## Purpose

The immediately preceding handoff run reported that the current dirty LogN13 P93 production candidate stays within the fixed `1e-2` numerical gate through `post-S2C`, but the ordinary public-like Bootstrap output fails badly.

Reported local evidence from that run:

- fixed architecture: q0=56, PS-wide Q012, planScale=`2^93`;
- final PS output max semantic difference about `5.0346e-7`;
- EvalMod real about `0.004183019582623534`;
- EvalMod imag about `0.004261561142162278`;
- post-S2C max component about `0.006934820352192803`;
- public-like max component about `0.2219142519401473`;
- `final.restore_capacity` passed with ratio about `3.9302334609202067e-19`;
- DoubleAngle checkpoints matched;
- no Secondary production source was changed or committed.

These numbers were produced in the local Codex environment and were intentionally not committed by the previous task. Reproduce the relevant boundary before making a causal claim.

The only goal of this task is:

> Localize the first failing operation between the raw `SlotsToCoeffs`/post-S2C result and the ordinary Fast public Bootstrap output.

Do not repair anything in this task.

---

## Why this boundary is now the target

The committed Fast public path performs, after the Bootstrap core:

1. `UnpackAndSwitchN2ToN1`;
2. `finalizeFastPublicCiphertext`;
3. public return.

The committed finalizer performs two logically distinct actions on the maintained rows:

1. Montgomery `IMForm` conversion and `IsMontgomery=false`;
2. metadata assignment `ct.Scale = ResidualParameters.DefaultScale()`.

A historical Primary finalization diagnostic already contains a useful F0/F1/F2/F3/F4 decomposition. Reuse or adapt that methodology rather than inventing a new broad diagnostic.

Important: the previous reported errors have a striking magnitude ratio:

`0.2219142519401473 / 0.006934820352192803 ~= 32.0000001`.

Treat this only as a clue. It is not proof of a scale bug. Record the actual pre-finalization/default scale ratio with high precision and test the boundary directly.

---

## Repository state and safety

### Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Synchronize `main` safely according to `AGENTS.md`, then re-read:

- `AGENTS.md`
- `docs/CODEX_HANDOFF.md`
- `CURRENT_TASK.md`
- this spec

### Secondary

Repository: `xuejin-lu/lattigo`

Expected state:

- branch: `fast-ckks`;
- committed HEAD: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`;
- intentionally dirty worktree containing the fixed-width Q012 / PS-wide P93 production candidate from the prior production attempt.

The dirty state is the implementation under diagnosis.

On entry:

1. inspect branch, HEAD, `git status --short`, and `git diff --stat`;
2. compute a compact deterministic fingerprint of the current dirty diff, e.g. SHA-256 of `git diff --no-ext-diff`, and record it in the result summary;
3. do not reset, stash, clean, discard, rebase, checkout-overwrite, or pull across the dirty worktree;
4. if branch/HEAD differ from the expected state, or the worktree is unexpectedly clean, stop and report the mismatch.

No Secondary source modification, commit, or push is authorized in this task.

---

## Fixed experiment conditions

Keep exactly the accepted current candidate conditions:

- LogN13 only;
- q0 = 56-bit profile;
- q1 approximately 39 bits;
- q2 approximately 40 bits;
- PS-wide Q012 authoritative arithmetic;
- planScale = `2^93`;
- q3+ are not arithmetic sources;
- Q012 -> Q01 only at PS exit;
- existing downstream bounded local-q2 behavior;
- same deterministic full-slot workload used by the preceding handoff task;
- all 4096 slots;
- numerical component threshold = `1e-2`.

Do not tune parameters, polynomial degree, decomposition, workload, or threshold.

---

## Step 0 — reproduce the reported boundary

Before interpreting finalization, reproduce on the normal baseline path:

- pre-public-finalization/post-S2C logical max component is <= `1e-2`;
- ordinary Fast public-like output max component is > `1e-2`.

Record both values.

If this boundary does not reproduce deterministically, stop and report `NON_REPRODUCIBLE_HANDOFF_BOUNDARY`. Do not continue by using stale `/tmp` values as authoritative evidence.

The previous `/tmp/fix001_p3_handoff_localization.json` may be consulted only as auxiliary evidence if still present; the current run must establish its own boundary result.

---

## Required checkpoints

Use independent ciphertext copies so a diagnostic transform cannot contaminate another checkpoint.

### F0 — raw post-S2C / pre-unpack

Capture the ciphertext immediately after the Fast `SlotsToCoeffs` stage / Bootstrap core and before `UnpackAndSwitchN2ToN1`.

Record compact evidence:

- Level;
- Scale in high-precision string form and log2 form;
- Degree;
- N;
- LogSlots;
- `IsNTT`;
- `IsMontgomery`;
- q0/q1 row hashes for c0/c1;
- maintained-limb count.

Also construct a diagnostic logical view if needed by applying only representation conversion on a copy while preserving the original Scale. Compare all 4096 decoded slots against the same Standard/reference semantics used in the previous handoff.

### F1 — after `UnpackAndSwitchN2ToN1`, before public finalizer

Run the ordinary current dirty Fast unpack path.

Compare F0 -> F1:

- exact maintained q0/q1 row hashes / first mismatch if any;
- Level / Scale / Degree / N / LogSlots;
- NTT/Montgomery flags.

For this LogN13 full-slot single-ciphertext profile, do not assume the boundary is trivial merely because historical runs were trivial. Prove it for the current dirty P93 candidate.

If F0 passes logically but F1 introduces a meaningful value change, classify the exact unpack/ring-switch operation and stop before proposing a repair.

### F2 — `IMForm` only, original Scale preserved

From an independent F1 copy:

1. apply the same maintained-row `IMForm` conversion used by `finalizeFastPublicCiphertext`;
2. set `IsMontgomery=false` consistently;
3. preserve the exact F1 Scale;
4. do not assign residual default scale.

Decode all 4096 slots and compare against the Standard/reference output and deterministic input.

Also perform an independent `IMForm -> MForm` round-trip on maintained q0/q1 rows and require bit-exact restoration. If the round trip fails, classify `MONTGOMERY_ROUNDTRIP_FAILURE` and stop causal interpretation beyond this point.

### F3 — current public scale restoration

From another independent F1 copy:

1. perform the same `IMForm` conversion;
2. set `IsMontgomery=false`;
3. assign `Scale = ResidualParameters.DefaultScale()` exactly as current production does.

Decode all slots and compare against Standard/reference.

Record with high precision:

- `F1.Scale`;
- `ResidualParameters.DefaultScale()`;
- ratio `F1.Scale / DefaultScale`;
- inverse ratio;
- whether F2 passes `1e-2`;
- whether F3 passes `1e-2`.

If F2 passes and F3 fails, additionally check the metadata-only scaling prediction without altering production:

`decoded_F3 ~= decoded_F2 * (F1.Scale / DefaultScale)`.

Report a compact max residual for that prediction. This is diagnostic evidence only; do not implement a correction.

### F4 — official Fast public output

Run the ordinary current dirty `Bootstrap` path unchanged.

Compare F4 against F3:

- logical max component difference;
- metadata equality;
- q0/q1 hashes where representations are directly comparable.

If F3 reproduces F4, the diagnostic copy is faithful.

If F3 and F4 disagree materially, classify `DIAGNOSTIC_PUBLIC_PATH_MISMATCH` and stop without guessing.

---

## Interpretation / stop rule

Stop at the first supported failing boundary. Use one of these classifications:

### A — `UNPACK_BOUNDARY_CHANGED_VALUE`

F0 passes, but F0 -> F1 changes maintained value/semantics materially.

Conclusion: first supported cause is `UnpackAndSwitchN2ToN1` or a sub-operation within it.

### B — `MONTGOMERY_REPRESENTATION_BOUNDARY`

F0/F1 are sound, but `IMForm -> MForm` is not exact or F2 newly fails because of representation conversion.

Conclusion: first supported cause is the Montgomery representation boundary.

### C — `FINAL_SCALE_RESTORATION`

F2 passes at the original post-S2C Scale, F3 fails after metadata-only assignment to residual default scale, and F3 matches F4.

Conclusion: first supported cause is the public final scale restoration boundary.

Do not yet decide whether the eventual repair belongs in S2C scale planning, coefficient compensation, or finalizer logic. That is the next task only after this classification is proven.

### D — `UNRESOLVED_PUBLIC_PATH_MISMATCH`

F2 and F3 pass but F4 fails, or F3 does not reproduce F4.

Do not guess. Report the mismatch.

### E — `PRE_FINALIZATION_REPRODUCTION_CONFLICT`

The current run cannot reproduce the prior claim that post-S2C passes while public output fails.

Stop and report the exact new evidence.

---

## `final.restore_capacity` is not this boundary

Do not conflate the earlier `final.restore_capacity` checkpoint inside the EvalMod/restore diagnostic path with `finalizeFastPublicCiphertext` at the public Bootstrap boundary.

The preceding run reported the former as passing. This task is specifically about the latter public finalization path.

---

## Output and evidence discipline

Create a compact, human-reviewable Primary artifact:

`results/FIX-001-P3-DIAG-PUBLIC-FINALIZATION-summary.json`

It must include at minimum:

- Primary commit used for the run;
- Secondary branch and committed HEAD;
- Secondary dirty status;
- Secondary `git diff --stat` summary;
- SHA-256 fingerprint of the dirty diff;
- fixed parameter/profile identity;
- Step-0 reproduced pre-finalization and public-output errors;
- F0/F1 row/metadata comparison;
- F2 logical error and round-trip result;
- F3 logical error and exact scale ratio;
- metadata-only scaling prediction residual if Case C applies;
- F4 logical error and F3-vs-F4 comparison;
- one classification from A-E;
- explicit statement that Secondary production code was not modified.

Do not commit full per-slot vectors or huge raw artifacts. Temporary detailed output may remain under `/tmp`.

Primary-only diagnostic helper/test changes are allowed if required and narrowly scoped. After successful completion, commit and push the compact Primary evidence and any required Primary diagnostic code according to `AGENTS.md` standing safe-push rules.

Do not commit or push Secondary.

---

## Tests

Run at least the focused diagnostic test for this task and any directly affected Primary tests.

Do not run LogN16, Gate4/5, EXP-003, or a new benchmark campaign.

No new performance timing is required in this task; the prior `5.0182x` timing remains provisional local evidence and is not the target here.

---

## Prohibitions

- no Secondary production source changes;
- no parameter tuning;
- no T2 repair;
- no PS repair;
- no new power replacement sweep;
- no DoubleAngle redesign;
- no threshold relaxation;
- no LogN16;
- no Gate4/5;
- no EXP-003;
- no Secondary commit/push;
- no reset/stash/clean/discard of the intentional dirty Secondary worktree.

The deliverable is localization evidence only, not a repair.