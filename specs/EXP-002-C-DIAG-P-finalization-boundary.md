# EXP-002-C-DIAG-P — Fast Bootstrap finalization-boundary diagnosis

## Status

READY after this spec is routed by `CURRENT_TASK.md`.

## Purpose

EXP-002-C-DIAG proved that the formal LogN13 Fast result is numerically wrong, but the first checkpoint at which the existing harness could support a normal CKKS numerical comparison was `unpack_and_switch_n2_to_n1`.

That does **not** prove that unpack/ring switching is the cause. In the formal LogN13 profile:

- `N1 == N2 == 8192`
- full slots are used (`4096` slots)
- the run contains one ciphertext
- therefore packing, unpacking and ring-degree switching are expected to be trivial boundaries.

This task must answer exactly one question:

> Is the Fast value already wrong at/before the SlotsToCoeffs output, or is the numerical error introduced by the final public-output boundary (unpack, Montgomery conversion, or scale restoration)?

Do not repair anything in this task.

---

## Fixed provenance

### Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Start from current clean `origin/main` containing EXP-002-C-DIAG results.

Expected base when this spec is authored:

`de7d7a778fb50bce60df8ba81dea10af62df4a19`

Follow `AGENTS.md` startup/safety rules. If the worktree is dirty, branch is wrong, or safe fast-forward synchronization is impossible, stop and report.

### Secondary backends

Use the same formal backend commits as EXP-002-C-DIAG:

- Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Fast: `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Do not modify Lattigo.

---

## Fixed workload

Use exactly the existing formal LogN13 correctness workload and config from EXP-002-C / EXP-002-C-DIAG.

- profile: LogN13 only
- slots: all 4096 slots
- deterministic input:
  - `real = ((i % 7) - 3) / 16`
  - `imag = ((i % 5) - 2) / 32`
- existing `configs/bootstrap_config.logN13.json`
- existing circuit order and parameter construction
- fixed numerical threshold for logical decoded values:
  - `max_abs_real <= 1e-2`
  - `max_abs_imag <= 1e-2`

Do **not** run LogN16.

---

## Scope

Instrument the existing diagnostic runner only as much as required to observe the Fast boundary after `SlotsToCoeffs` and before/through public finalization.

Do not add a Fast/Standard selector to the public workload contract. Preserve the existing backend-drop-in experiment architecture.

Do not import or call private/unexported Lattigo implementation helpers from the Primary repository. Use the same public stage APIs already used by the EXP-002-C-DIAG runner and ordinary public ring/CKKS operations for diagnostic copies.

Do not modify Secondary/Lattigo source.

---

## Required Fast checkpoints

For the formal Fast LogN13 execution, retain separate copies so each diagnostic manipulation is isolated and does not affect another checkpoint.

### F0 — `slots_to_coeffs_raw`

Capture the ciphertext immediately after the Fast `SlotsToCoeffs` stage and before `UnpackAndSwitchN2ToN1`.

Record at minimum:

- Level
- Scale (high-precision/string form if available, plus float approximation)
- Degree
- N
- LogSlots
- `IsNTT`
- `IsMontgomery`
- maintained limb count
- hashes/fingerprints of maintained q0/q1 rows for c0/c1

Do not claim an ordinary CKKS decoded-value comparison for this raw Montgomery checkpoint unless the representation is first normalized on a diagnostic copy.

### F1 — `after_unpack_raw`

Run the ordinary Fast public `UnpackAndSwitchN2ToN1` boundary used by the experiment.

Because this exact formal profile has one ciphertext and `N1 == N2`, test whether this boundary is mathematically/bitwise trivial.

Compare F1 against F0 for the maintained state:

- q0/q1 rows of c0/c1: exact equality
- Level
- Scale
- Degree
- N
- LogSlots
- NTT/Montgomery flags

If any maintained residue differs, report the exact first mismatch and classify `UNPACK_BOUNDARY_CHANGED_VALUE`.

Do not hide such a mismatch behind decoded tolerances.

### F2 — `imform_only`

From an independent copy of F1:

1. Apply ordinary ring `IMForm` to every maintained q0/q1 row of c0/c1 exactly as the Fast public finalization does.
2. Set the copy's Montgomery metadata consistently to false.
3. **Do not change the ciphertext Scale.** Preserve the scale from F1/S2C.
4. Keep NTT state otherwise unchanged.
5. Decrypt using the authoritative Fast zero secret and decode through ordinary public CKKS APIs.

Compare all 4096 decoded slots against:

- deterministic input
- Standard authoritative generated-secret decoded output from the same formal workload

Compute the same correctness metrics used by EXP-002-C:

- max_abs_complex
- mean_abs_complex
- rmse_complex
- max_abs_real
- max_abs_imag
- max_component_abs
- max-error indices

This checkpoint asks whether the logical S2C output is already wrong when represented as an ordinary non-Montgomery ciphertext **without scale restoration**.

### F2R — `imform_mform_roundtrip`

Prove that the diagnostic representation conversion itself is not silently changing the maintained residues.

From a separate F1 copy:

1. apply IMForm on maintained q0/q1 rows;
2. apply MForm back on those same rows;
3. compare the resulting maintained q0/q1 rows bit-for-bit with the original F1 rows.

Record exact equality for each c0/c1 q0/q1 row.

This is a representation round-trip invariant, not a decoded-value tolerance test.

If this round trip is not exact, classify `MONTGOMERY_ROUNDTRIP_FAILURE` and stop interpretation beyond this boundary.

### F3 — `full_finalization_view`

From another independent F1 copy:

1. apply IMForm to maintained q0/q1 rows;
2. set Montgomery metadata to false;
3. set Scale to `ResidualParameters.DefaultScale()` exactly as the Fast public finalizer does;
4. decrypt with the authoritative Fast zero secret and decode through ordinary public APIs.

Compare all 4096 slots against input and the Standard authoritative output with the fixed `1e-2` component threshold.

Also record:

- S2C/F1 scale
- residual default scale
- ratio `F1.Scale / ResidualDefaultScale`
- whether scale metadata changed only at this step

### F4 — official Fast public output

Run the existing ordinary Fast Bootstrap public path unchanged and retain the existing final decoded result.

Confirm whether F3 and the official Fast public output agree numerically. They should agree if the diagnostic copy reproduces finalization correctly.

If they do not, classify `DIAGNOSTIC_FINALIZATION_MISMATCH` and do not infer an implementation root cause.

---

## Standard reference

Run the same fixed Standard LogN13 workload at the pinned Standard backend.

Use the ordinary generated secret as the authoritative Standard decode, exactly as EXP-002-C established.

No need to force Standard's internal intermediate representations into the Fast checkpoint layout. The Standard final decoded output is the logical-value reference for F2/F3/F4.

Archive all 4096 Standard decoded slots required for Fast-vs-Standard comparisons.

---

## Required interpretation matrix

The summary must select only from the following evidence-backed classifications.

### Case A — unpack changes maintained residues

Condition:

- F0 → F1 q0/q1 maintained rows are not bit-exact.

Conclusion:

`FIRST_SUPPORTED_CAUSE = unpack_and_switch_n2_to_n1`

Do not repair it in this task.

### Case B — Montgomery round trip fails

Condition:

- F0 → F1 is exact/trivial, but
- F2R IMForm→MForm is not bit-exact.

Conclusion:

`FIRST_SUPPORTED_CAUSE = montgomery_representation_boundary`

Do not repair it in this task.

### Case C — F2 already numerically diverges, round trip is exact

Condition:

- F0 → F1 exact/trivial
- F2R exact
- F2 exceeds the fixed numerical threshold versus input and/or Standard

Conclusion:

`FIRST_SUPPORTED_CAUSE = at_or_before_slots_to_coeffs`

This explicitly rules out unpack and public Montgomery conversion as the first supported cause. It does **not** distinguish C2S vs EvalMod vs S2C yet.

### Case D — F2 passes, F3 fails

Condition:

- F0 → F1 exact
- F2R exact
- F2 passes
- F3 fails

Conclusion:

`FIRST_SUPPORTED_CAUSE = final_scale_restoration`

The next task may investigate why replacing the S2C scale with residual default scale changes the logical value.

### Case E — F2 and F3 pass but F4 fails

Conclusion:

`FIRST_SUPPORTED_CAUSE = unresolved_public_path_mismatch`

Do not guess.

### Case F — F2/F3/F4 all pass

Conclusion:

`FIRST_SUPPORTED_CAUSE = unresolved_previous_diagnostic_inconsistency`

The previous EXP-002-C-DIAG observation must be reconciled before proceeding.

---

## Required result files

Write raw outputs to `/tmp` first while both repositories remain clean, then copy final artifacts into Primary only after the diagnostic run is complete.

Create at minimum:

- `results/EXP-002C-DIAG-P-logN13-standard.json`
- `results/EXP-002C-DIAG-P-logN13-fast.json`
- `results/EXP-002C-DIAG-P-logN13-summary.json`

The Fast raw JSON must contain F0, F1, F2, F2R, F3 and F4 evidence.

Do not omit full decoded vectors for checkpoints that are used for logical-value comparisons.

---

## Tests

Add focused tests for diagnostic helper logic where practical, especially:

- maintained-row exact comparison
- IMForm-only copy does not mutate source
- IMForm→MForm round-trip comparison
- finalization-copy logic does not alias source ciphertext storage

Run:

- `go test ./...` against Fast backend
- `go test ./...` against Standard backend

Record exact backend commits and repository dirty state in raw results.

---

## Stop conditions

Stop and report without proceeding further if:

- repository synchronization is unsafe;
- either backend commit cannot be checked out cleanly;
- a required public stage cannot be reproduced without modifying Lattigo;
- the F2R representation round trip fails unexpectedly and prevents reliable interpretation;
- diagnostic output cannot be reproduced deterministically.

Do not weaken the `1e-2` logical threshold.

---

## Explicit non-goals

Do not:

- modify Lattigo;
- fix the numerical bug;
- run LogN16;
- run parameter sweeps;
- change `LogMessageRatio`, K, degree, DoubleAngle or profile parameters;
- start EXP-003;
- benchmark performance;
- make causal claims about C2S/EvalMod/S2C unless this task's evidence supports them.

---

## Completion

Mark this task COMPLETE only when:

1. F0→F1 triviality or change is established with exact maintained-row evidence;
2. F2R Montgomery round-trip evidence is recorded;
3. F2, F3 and F4 logical comparisons are recorded where interpretation remains valid;
4. one classification from the required interpretation matrix is selected;
5. all required raw/summary files are committed;
6. both backend test suites pass;
7. no Lattigo source was modified;
8. LogN16 and EXP-003 were not run.

The deliverable is a localization result, not a repair.