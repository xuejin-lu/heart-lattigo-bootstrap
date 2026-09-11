# EXP-002-C-DIAG-Q01 — q0/q1 intermediate projection oracle

## Goal

Localize the first numerical divergence inside the Fast Bootstrap core among:

`ModUp -> CoeffsToSlots -> EvalMod -> SlotsToCoeffs`

using only the q0/q1 residues that Fast intentionally maintains.

This task is diagnostic only. Do not modify Lattigo and do not repair the bug.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected starting remote HEAD when this spec is authored:

`4adf7b4217362c6d69c5c2413429ac9edfc2e560`

Pinned backends:

- Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Fast: `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Follow `AGENTS.md` startup and safety rules.

Do not modify Secondary/Lattigo.

---

## Fixed workload

Use exactly the existing formal LogN13 workload:

- LogN = 13
- full 4096 slots
- `configs/bootstrap_config.logN13.json`
- deterministic values:
  - `real = ((i % 7) - 3) / 16`
  - `imag = ((i % 5) - 2) / 32`
- circuit order unchanged
- threshold unchanged:
  - `max_abs_real <= 1e-2`
  - `max_abs_imag <= 1e-2`

Do not run LogN16.

---

## Why this diagnostic is needed

Earlier diagnostics established:

- input/packing/ScaleDown match;
- final public output diverges;
- final unpack is trivial for this profile;
- Montgomery IMForm/MForm round trip is exact;
- scale restoration is a no-op (`ratio = 1`);
- therefore the error is already present at or before `SlotsToCoeffs`.

The remaining obstacle is that Fast intermediate ciphertexts at logical levels >1 intentionally maintain only q0/q1 while higher RNS limbs are dormant. Ordinary full-level CKKS decrypt/decode is therefore not a valid oracle for those Fast intermediates.

This task creates a backend-neutral diagnostic projection onto q0/q1 only.

---

## Core diagnostic object: q0/q1 projection

For any intermediate ciphertext with logical level >= 1, build an isolated diagnostic ciphertext under the same Bootstrap CKKS parameters with:

- degree unchanged (expected degree 1);
- **physical/logical diagnostic level = 1**;
- only q0 and q1 copied from the source ciphertext;
- source Scale copied exactly;
- source LogSlots / batching metadata copied exactly;
- source NTT state preserved;
- no q2+ rows allocated or consulted.

### Standard source

Standard intermediates are ordinary non-Montgomery ciphertexts.

Copy q0/q1 into the level-1 diagnostic ciphertext unchanged.

Decrypt with the Standard bootstrap secret returned by `GenEvaluationKeys`.

### Fast source

Fast intermediates are Montgomery on the maintained q0/q1 rows.

On the isolated diagnostic copy only:

1. copy q0/q1;
2. apply public ring `IMForm` to q0/q1 for c0/c1;
3. mark diagnostic copy non-Montgomery;
4. do not alter Scale;
5. decrypt with the authoritative zero bootstrap secret;
6. ordinary CKKS-decode using the Bootstrap parameters at level 1.

Never read dormant q2+ rows.

Never mutate the stage source ciphertext.

---

## Mandatory oracle validation on Standard

A q0/q1 projection is only allowed to classify a stage if the projection itself is validated on Standard.

For each Standard stage/branch:

1. decode the original full-RNS Standard ciphertext normally with the Standard bootstrap secret;
2. construct the Standard q0/q1 level-1 projection;
3. decode that projection with the same Standard bootstrap secret;
4. compare full-RNS Standard decode vs Standard q0/q1 projection decode.

Required oracle-validity threshold:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

If Standard full-RNS vs Standard q0/q1 projection fails for a stage, mark that stage:

`PROJECTION_ORACLE_INVALID`

and do not use that stage to make a Fast causal classification.

Do not loosen the threshold.

---

## Required checkpoints

Run Standard and Fast through the same ordinary public stage API already used by EXP-002-C-DIAG.

Capture these checkpoints before subsequent stages mutate them:

### Q1 — `mod_up`

Single ciphertext after `ModUp`.

Compare:

- Standard full decode vs Standard q01 projection (oracle validation)
- Standard q01 projection vs Fast q01 projection

### Q2 — `coeffs_to_slots`

Two branches when present:

- real
- imag

For each branch independently compare:

- Standard full decode vs Standard q01 projection
- Standard q01 projection vs Fast q01 projection

Do not compare either branch directly to the original application input; only same-stage same-branch Standard-vs-Fast is authoritative here.

### Q3 — `eval_mod_real`

Compare Standard vs Fast q01 projections after EvalMod(real).

### Q4 — `eval_mod_imag`

Compare Standard vs Fast q01 projections after EvalMod(imag), when imag exists.

### Q5 — `slots_to_coeffs`

Compare Standard vs Fast q01 projections immediately after `SlotsToCoeffs` and before public unpack/finalization.

Because this stage is already level 1, the Fast q01 projection should also be numerically consistent with the previously established F2 IMForm-only view. Record this consistency check.

---

## Required metrics

For every vector comparison record:

- max_abs_complex
- mean_abs_complex
- rmse_complex
- max_abs_real
- max_abs_imag
- max_component_abs
- max-error indices
- sample count
- pass/fail at fixed threshold

Archive complete decoded vectors for all 4096 slots and all applicable branches.

Also record for each source/projection:

- logical source Level
- diagnostic projection Level (=1)
- Scale string + log2
- N
- LogSlots
- IsNTT
- IsMontgomery before normalization
- exact q0/q1 moduli
- q0/q1 maintained-row hashes before/after diagnostic copy

Prove the source ciphertext is not mutated by projection.

---

## Required parameter sanity

Before comparisons, verify Standard and Fast use identical exact q0 and q1 moduli for this formal profile.

If q0/q1 differ, stop and report `Q01_PARAMETER_MISMATCH`.

---

## Classification logic

Only stages whose Standard projection oracle is valid may be used.

Select the **first** valid stage in pipeline order whose Standard-q01 vs Fast-q01 comparison exceeds the fixed threshold.

Allowed outcomes:

### Case A — Q1 diverges

`FIRST_SUPPORTED_CAUSE = mod_up`

Do not inspect/fix later stages in this task beyond recording already-produced evidence.

### Case B — Q1 matches, Q2 diverges

`FIRST_SUPPORTED_CAUSE = coeffs_to_slots`

Identify real, imag, or both branches.

### Case C — Q1/Q2 match, Q3 or Q4 diverges

`FIRST_SUPPORTED_CAUSE = eval_mod`

Identify real, imag, or both branches and the first divergent branch result.

### Case D — Q1-Q4 match, Q5 diverges

`FIRST_SUPPORTED_CAUSE = slots_to_coeffs`

### Case E — all valid q01 checkpoints match but existing final F2 diverges

`FIRST_SUPPORTED_CAUSE = diagnostic_inconsistency`

Stop. Do not guess.

### Case F — one or more earlier projection oracles are invalid

If the first potential divergence cannot be established because the Standard q01 oracle is invalid, classify:

`FIRST_SUPPORTED_CAUSE = unresolved_projection_oracle`

Report exactly which stage first became unusable.

---

## Important interpretation rule

A Fast stage must **not** be called wrong merely because its q0/q1 projection differs from the original input.

Only these comparisons are causal:

- Standard full vs Standard q01 -> validates the oracle;
- Standard q01 vs Fast q01 -> tests backend equivalence at that same stage and branch.

---

## Required files

Write raw results to `/tmp` first while both repos are clean, then copy final artifacts into Primary.

Create at minimum:

- `results/EXP-002C-DIAG-Q01-logN13-standard.json`
- `results/EXP-002C-DIAG-Q01-logN13-fast.json`
- `results/EXP-002C-DIAG-Q01-logN13-summary.json`

Summary must contain:

- Standard projection-oracle validity by stage/branch;
- Standard-q01 vs Fast-q01 metrics by stage/branch;
- first supported cause;
- exact reason for any unresolved classification.

---

## Tests

Add focused tests for the diagnostic projection helper:

- q0/q1 copied exactly;
- diagnostic level exactly 1;
- source Scale/LogSlots preserved;
- Fast IMForm happens only on the copy;
- source ciphertext remains bit-exact unchanged;
- no q2+ row is read or required;
- Standard projection decode matches full Standard decode on a small deterministic case;
- zero-secret Fast projection decode path executes on a small deterministic case.

Run:

- `go test ./...` against Fast backend;
- `go test ./...` against pinned Standard backend.

---

## Stop conditions

Stop and report if:

- repository sync is unsafe;
- either pinned backend cannot be checked out cleanly;
- q0/q1 parameters differ;
- projection cannot be constructed without modifying Lattigo;
- Standard projection oracle fails before a supported causal localization can be made;
- diagnostic runs are non-deterministic.

---

## Explicit non-goals

Do not:

- modify Lattigo;
- repair the numerical bug;
- change Fast arithmetic;
- change any bootstrap parameter;
- sweep `LogMessageRatio`, K, degree, DoubleAngle, or DFT factorization;
- run LogN16;
- benchmark performance;
- start EXP-003.

---

## Completion

Mark COMPLETE only after:

1. Standard q01 oracle validity is recorded for every checkpoint used;
2. Standard-vs-Fast q01 comparisons are recorded stage-by-stage;
3. the first supported cause is selected using the fixed classification logic;
4. all raw vectors and summaries are committed;
5. both backend test suites pass;
6. no Lattigo source is modified;
7. LogN16 and EXP-003 remain untouched.

The deliverable is a stage localization result, not a fix.