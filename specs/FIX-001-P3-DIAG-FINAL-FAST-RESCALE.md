# FIX-001-P3-DIAG-FINAL-FAST-RESCALE — Isolate the first failing sub-step inside final Fast Rescale

## Purpose

The completed LogN13 compressed-PS normalization diagnostic established:

- all candidate baby blocks pass;
- all normalized giant steps pass;
- the last pre-final-Rescale accumulation checkpoint is semantically correct at about `4.13e-8` max component error;
- the actual final Fast `Rescale` raises whole degree-30 polynomial error to about `0.0114439`, above the fixed `1e-2` threshold;
- the six compressed internal scales `2^91 ... 2^86` produce nearly identical final error;
- no obvious q0/q1 capacity overflow was observed;
- target-scale promotion is not reached;
- prior classification is `FIRST_SUPPORTED_CAUSE = compressed_finalization_failure_after_normalization`.

This task must locate the **first unsupported or incorrect sub-step inside the final Fast Rescale boundary** without changing production Lattigo behavior.

Do not assume a Fast Rescale implementation bug. The diagnostic must distinguish an implementation/sub-step mismatch from a q0/q1-only semantic insufficiency.

Diagnostic only. LogN13 only.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`fe5c483a7ecaaa2a74b3c42300f238c00050d0d0`

Authoritative normalization result commit:

`0292d11edfa5bfee647bfd3a04d1f395c805ab96`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Relevant pinned Secondary source:

- `schemes/ckks/fast/rescale.go`
- `schemes/ckks/fast/partial_ntt.go`

Both worktrees must start clean and follow their `AGENTS.md` startup rules.

Secondary must be restored to exact clean `61607bb4bb82591009ce768d9a3773bed1497565` before completion.

Parent FIX-001/P3 remains incomplete.

---

## Scope lock

Do not:

- run LogN16;
- benchmark;
- run formal Gate 4/5 production experiments;
- start EXP-003;
- modify production Lattigo code;
- change the polynomial, PS schedule, candidate scale schedule, semantic threshold, or P3 power generation;
- read or depend on stale q2+ Fast residues as if they were maintained state;
- introduce a production repair;
- declare `Rescale` buggy merely because the final semantic oracle fails.

The only question is:

> Starting from the already-correct pre-final-Rescale ciphertext, at which smallest internal boundary does the Fast final Rescale first diverge from its independently checked intended operation?

---

# Mandatory historical reproduction

Use the exact LogN13 real-branch degree-30 polynomial workload and normalization path from the preceding task.

Use `2^91` as the mandatory canonical probe candidate.

Before internal diagnosis, reproduce all of the following for that candidate:

1. all baby blocks pass;
2. all normalized giant steps pass;
3. the last pre-final-Rescale whole-polynomial error is <= `1e-2` and remains in the previously observed `~4e-8` regime;
4. actual final Fast `Rescale` produces whole-polynomial error > `1e-2` and remains in the previously observed `~0.01144` regime.

If this boundary does not reproduce, stop with:

`FINAL_RESCALE_PRECONDITION_MISMATCH`.

Do not proceed by tuning inputs until it fails.

After the first failing sub-step has been identified for `2^91`, use `2^86` only as a low-scale confirmation probe. Do not replay all six candidates unless the canonical and confirmation probes disagree in the location of the first failing sub-step.

---

# Source-backed Fast Rescale path

For the canonical failing final Rescale, record the actual:

- input `Level()` and `Degree()`;
- `IsNTT` and `IsMontgomery` state;
- `LevelsConsumedPerRescaling()` / effective `nbRescales`;
- dropped logical modulus or moduli;
- input and output Scale;
- q0/q1 moduli;
- q0/q1 residue hashes for each ciphertext component.

The current pinned source path to diagnose is:

```text
FastPartialINTT(q0,q1)
→ optional IMForm when input is Montgomery
→ centered CRT reconstruction from q0/q1
→ rounded division by dropped logical modulus/moduli
→ signed residue encoding into q0/q1
→ NTT restoration
→ optional MForm restoration
→ metadata copy and Scale division
```

The existing Level-1 ordinary non-Montgomery special path must be recorded if selected. If that path is selected, diagnose its own `DivRoundByLastModulusNTT` boundary separately and do not pretend the generic CRT path ran.

---

# Independent-oracle rule

Do not validate a sub-step only by reusing the same helper being tested.

For the generic path, diagnostic code in the Primary harness may mirror the pinned formulas from `rescale.go`, but every arithmetic checkpoint must also have an **independent oracle**:

- CRT and centered reconstruction: `math/big` or equivalently exact arbitrary-precision arithmetic;
- rounded division: exact signed arbitrary-precision quotient/remainder arithmetic using the same documented nearest-integer tie rule;
- modular residue encoding: independent signed modulo arithmetic;
- NTT/INTT representation checks: independent round-trip / expected-residue comparison through public ring primitives.

Do not alter or export Secondary private helpers merely to make them callable.

Temporary diagnostic helpers are allowed only in the experiment harness or temporary tests; they must not become production Lattigo behavior.

---

# Minimal checkpoints

Diagnose in order. Stop causal interpretation at the first failed checkpoint, but still archive enough state to make that classification auditable.

## R0 — pre-final-Rescale semantic state

Snapshot the exact ciphertext immediately before actual final Fast Rescale.

Record:

- whole-polynomial semantic max component error;
- Level/Degree/Scale;
- NTT/Montgomery state;
- per-component q0/q1 residue hashes;
- existing q0/q1 centered-capacity evidence.

Require semantic error <= `1e-2`.

Failure classification:

`pre_final_rescale_state_failure`

This is a precondition failure, not a Rescale failure.

## R1 — q0/q1 coefficient-domain recovery

On an independent snapshot, apply the same public q0/q1 inverse-transform semantics as `FastPartialINTT` and, when required by the input representation, remove Montgomery form exactly as the pinned source does.

For each ciphertext component and maintained limb:

1. record coefficient-domain residue hash;
2. round-trip the recovered residues back to the original input representation;
3. compare every maintained q0/q1 residue modulo its modulus with the R0 snapshot.

Require exact residue equality.

Failure classification:

`final_rescale_partial_intt_or_representation_failure`

Do not proceed to blame CRT if R1 fails.

## R2 — centered CRT reconstruction

For every component/coefficient, reconstruct the integer represented by the recovered q0/q1 residues.

Compare the pinned `crtQ01`/centering semantics against an independent arbitrary-precision CRT oracle.

Require:

- exact agreement on the non-centered representative modulo `Q01 = q0*q1`;
- exact agreement on the centered sign and magnitude;
- re-encoding the centered integer modulo q0 and q1 reproduces the R1 residues exactly.

Also record:

- maximum absolute centered magnitude;
- `maxAbs / (Q01/2)` ratio;
- outside-count relative to the centered uniqueness interval.

Important: a passing q0/q1 centered-range check proves only self-consistency of the q0/q1 representative. It must **not** be described as proof that no aliasing relative to an unavailable higher-modulus integer occurred.

Failure classification:

`final_rescale_centered_crt_failure`

## R3 — rounded logical-modulus division

For each R2 centered integer, apply the exact dropped logical modulus sequence used by the actual final Rescale.

Compare the pinned fixed-width rounding path against an independent signed arbitrary-precision division oracle.

For each division step require exact agreement on:

- sign;
- quotient magnitude;
- remainder relation to half the divisor;
- rounded signed quotient.

Record the first mismatching component/coefficient/divisor only, plus aggregate mismatch count and maximum quotient magnitude. Do not dump every coefficient.

Failure classification:

`final_rescale_rounded_division_failure`

## R4 — signed q0/q1 residue encoding

Independently encode the R3 signed rounded quotient modulo the maintained output limbs.

Compare against the residue values produced by the pinned `signedResidue` semantics before forward transform.

Require exact residue equality for every maintained output limb.

Failure classification:

`final_rescale_signed_residue_failure`

## R5 — NTT / Montgomery output restoration

From the independently validated R4 coefficient residues:

1. apply the required forward NTT;
2. if the original input was Montgomery, apply the required Montgomery restoration;
3. compare the resulting q0/q1 residues exactly against the actual Fast `Rescale` output before any later polynomial operation.

Require exact maintained-residue equality and matching output Level/Degree.

Failure classification:

`final_rescale_output_transform_failure`

## R6 — metadata / Scale transition

Independently compute expected output metadata from R0 according to pinned source semantics.

Require:

- metadata copy semantics are preserved except fields intentionally changed by Rescale/Resize;
- output Level is exactly `inputLevel - nbRescales`;
- output Scale equals input Scale divided by the exact dropped logical modulus sequence using the same `rlwe.Scale` arithmetic;
- no coefficient data changes during this metadata-only checkpoint.

Failure classification:

`final_rescale_metadata_scale_failure`

## R7 — post-Rescale semantic oracle

Only after R1-R6 pass, evaluate the actual Fast output against the same degree-30 polynomial oracle used at R0.

Record the whole-polynomial max component error.

If R1-R6 all pass exactly but R7 still exceeds `1e-2`, classify:

`final_rescale_q01_semantic_gap_without_internal_mismatch`

This classification means:

- the pinned implementation performed its q0/q1 algorithm consistently with the independently checked local arithmetic;
- the final semantic result is nevertheless outside tolerance;
- the evidence does **not yet prove** which stronger information/invariant is missing (for example, whether a centered-representative assumption is insufficient for this state).

Do not call this an implementation bug and do not invent a repair in this task.

---

# Low-scale confirmation

After classifying the canonical `2^91` probe, replay the same minimal checkpoint sequence for `2^86`.

Required output:

- whether the first failing checkpoint is identical;
- whether mismatch counts / first mismatch identity materially differ;
- R0 and R7 semantic errors.

If `2^91` and `2^86` disagree on the first failing checkpoint, classify overall result as:

`FINAL_RESCALE_SCALE_DEPENDENT_DIAGNOSTIC_DISAGREEMENT`

and stop. Do not expand automatically to all intermediate scales.

---

# Overall classification

Choose exactly one:

- `FIRST_SUPPORTED_CAUSE = final_rescale_partial_intt_or_representation_failure`
- `FIRST_SUPPORTED_CAUSE = final_rescale_centered_crt_failure`
- `FIRST_SUPPORTED_CAUSE = final_rescale_rounded_division_failure`
- `FIRST_SUPPORTED_CAUSE = final_rescale_signed_residue_failure`
- `FIRST_SUPPORTED_CAUSE = final_rescale_output_transform_failure`
- `FIRST_SUPPORTED_CAUSE = final_rescale_metadata_scale_failure`
- `FIRST_SUPPORTED_CAUSE = final_rescale_q01_semantic_gap_without_internal_mismatch`
- `FIRST_SUPPORTED_CAUSE = final_rescale_scale_dependent_diagnostic_disagreement`
- `FIRST_SUPPORTED_CAUSE = final_rescale_precondition_mismatch`

Use the first failing checkpoint only. Later consequences must not replace the earlier causal boundary.

---

# Artifact size discipline

This task must not create another multi-million-line JSON artifact.

Write compact diagnostic artifacts only:

- `results/FIX-001-P3-DIAG-FINAL-FAST-RESCALE-logN13.json`
- `results/FIX-001-P3-DIAG-FINAL-FAST-RESCALE-logN13-summary.json`

Do not serialize full per-slot vectors, all polynomial coefficients, full operation traces, or repeated index arrays.

For large vectors/polynomials store only:

- deterministic hash;
- length/count;
- max/min aggregate where relevant;
- mismatch count;
- first mismatching component/index/value tuple when a mismatch exists;
- worst-error component/index/value tuple where relevant.

The summary must remain human-reviewable and include:

- exact provenance;
- canonical `2^91` checkpoint table R0-R7;
- confirmation `2^86` checkpoint table;
- first failing checkpoint;
- independent-oracle method for each arithmetic checkpoint;
- q0/q1 centered-range evidence with the aliasing caveat;
- final overall classification;
- tests and clean-state evidence.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- any focused diagnostic tests pass;
- Secondary `go test ./...` passes if any temporary Secondary test/helper was used;
- all temporary Secondary changes are removed;
- Secondary is exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Lattigo changes are committed;
- Primary diagnostic artifacts are committed and pushed normally;
- both worktrees are clean;
- no LogN16;
- no benchmark;
- no Gate 4/5 production run;
- no EXP-003.

The deliverable is the smallest source-backed Fast final-Rescale boundary at which independent checking first fails, or a defensible result that all internal q0/q1 steps are locally correct while the final semantics still fail.