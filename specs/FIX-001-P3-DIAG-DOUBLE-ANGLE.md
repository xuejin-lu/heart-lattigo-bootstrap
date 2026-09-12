# FIX-001-P3-DIAG-DOUBLE-ANGLE — Validate the LogN13 Fast Mod1 DoubleAngle chain

## Purpose

The canonical LogN13 compressed polynomial path at internal scale `2^91` is now validated through exact restoration to the public polynomial target Scale.

Authoritative prior result:

`FIRST_SUPPORTED_CAUSE = compressed_ps_2p91_target_scale_restoration_validated_after_capacity_domain_fix`

with:

- corrected polynomial oracle: `raw_chebyshev_on_preprocessed_z`;
- corrected post-final polynomial error about `2.6648e-7`;
- corrected prospective target-scale promotion capacity ratio about `9.07e-11`;
- integer promotion and coefficient-domain checks passing;
- exact target-Scale metadata normalization passing;
- P4 semantic error about `2.6648126e-7`;
- P4 metadata perturbation about `7.89e-14`.

The next unresolved boundary is the Fast Mod1 **DoubleAngle** recurrence that follows polynomial evaluation.

Pinned Standard and Fast source both execute, for each DoubleAngle round:

1. square the running `sqrt2pi` constant;
2. `MulRelin(res, res, res)`;
3. `Add(res, res, res)`;
4. `Add(res, -sqrt2pi, res)`;
5. `Rescale(res, res)`.

After all rounds, both implementations perform a metadata-only reset:

`res.Scale = inputScale`.

This task must validate those operations step-by-step for LogN13, starting from the already-validated canonical `2^91` restored polynomial output.

Diagnostic only. No production Secondary changes.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`39a3e1d1cf11a1f817b6e325f7305dfa5961c8d0`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Authoritative polynomial oracle:

`raw_chebyshev_on_preprocessed_z`

Fixed semantic threshold:

`1e-2`

Canonical internal polynomial scale:

`2^91`

Do not modify Secondary production code.

---

# Scope lock

Do not:

- run lower internal polynomial scales;
- run LogN16;
- benchmark;
- run Gate 4/5 production experiments;
- start EXP-003;
- modify production Lattigo;
- change Mod1 parameters, polynomial coefficients, PS schedule, target-scale restoration, DoubleAngle count, constants, or threshold;
- use the historical double-change-of-basis O1 oracle;
- skip directly to only the final Mod1 output;
- continue beyond the final Mod1 scale reset into later bootstrap stages.

The only question is:

> Starting from the validated exact-target-scale compressed polynomial result, where—if anywhere—does the Fast DoubleAngle recurrence first cease to match its source-defined plaintext recurrence or violate q0/q1 centered-capacity requirements needed by Fast Rescale?

---

# Source-defined recurrence oracle

Let the validated P4 polynomial plaintext vector be:

`Y0 = raw_chebyshev_on_preprocessed_z(Mod1Poly, z)`.

Let:

`C0 = Mod1Parameters.Sqrt2Pi`.

For DoubleAngle round `i = 0 ... DoubleAngle-1`:

`Ci = C(previous)^2`

with the exact same update order as source, and define:

- after multiply: `M_i = Y_i * Y_i`;
- after doubling: `D_i = 2 * M_i`;
- after offset: `O_i = D_i - C_i`;
- after Rescale: `Y_{i+1} = O_i`.

Use independent plaintext/high-precision arithmetic for this recurrence. Do not derive expected values from decoded ciphertext intermediates.

Record the exact/float representation of each `C_i` used by the ciphertext operation and the plaintext oracle.

---

# Mandatory canonical starting state

Reproduce the exact `2^91` compressed polynomial path through corrected capacity-domain restoration P4.

Require before entering DoubleAngle:

1. exact target Scale equals the public polynomial target Scale;
2. P4 Level and Degree reproduce the prior validated state;
3. P4 q0/q1 hashes reproduce the prior authoritative restored state;
4. P4 corrected semantic max component error remains in the `~2.665e-7` regime and <= `1e-2`;
5. coefficient-domain q0/q1 capacity remains valid;
6. original Mod1 input Scale, before `res.Scale = ScalingFactor()`, is captured as an actual `rlwe.Scale` value for the final reset check.

If any of these fail:

`DOUBLE_ANGLE_PRECONDITION_MISMATCH`.

Do not enter round 0.

---

# Representation-correct capacity oracle

Any centered-capacity check in this task must use the corrected representation path established by the previous task:

`authoritative q0/q1 NTT/Montgomery rows`

`-> FastPartialINTT`

`-> IMForm when required`

`-> coefficient-domain centered CRT under Q01=q0*q1`.

Never use direct CRT of NTT row indices.

For every state checked, retain only compact evidence:

- maximum absolute centered coefficient;
- ratio to `Q01/2`;
- outside count;
- worst component/index/value.

---

# Per-round checkpoints

Execute all rounds one at a time on the actual Fast primitives. For every round `i`, preserve an independent input snapshot.

## D0.i — round input

Before `MulRelin` record:

- Level / Degree / Scale;
- q0/q1 hashes;
- source-backed error versus `Y_i`;
- corrected coefficient-domain capacity.

Require source-backed error <= `1e-2` and capacity outside count `0`.

If the input is already semantically wrong:

`double_angle_round_input_already_wrong`.

If capacity is already invalid:

`double_angle_round_input_capacity_failure`.

## D1.i — after MulRelin

Execute exactly:

`FastCKKS.MulRelin(res, res, res)`.

Compare decoded result with independent `M_i = Y_i^2`.

Record:

- Level / Degree / Scale;
- semantic max component error;
- error delta/ratio from D0.i where meaningful;
- q0/q1 hashes;
- corrected coefficient-domain capacity.

If first semantic failure occurs here:

`double_angle_multiply_failure`.

Do not require the post-multiply Scale to equal a hard-coded value; record and compare it to the expected algebraic scale product derived from the D0.i Scale.

## D2.i — after doubling

Execute exactly:

`FastCKKS.Add(res, res, res)`.

Compare against:

`D_i = 2*M_i`.

Record the same compact semantic/scale/capacity evidence.

If first failure occurs here:

`double_angle_doubling_failure`.

## D3.i — after constant offset

Update `sqrt2pi` in exactly the source order and execute:

`FastCKKS.Add(res, -C_i, res)`.

Compare against:

`O_i = D_i - C_i`.

Record the same compact evidence.

This checkpoint is especially important because it is the **actual input to Fast Rescale**.

Before allowing Rescale, require corrected coefficient-domain centered uniqueness:

`|x| < Q01/2`

for every maintained coefficient used by the Fast Rescale reconstruction.

If semantic failure first appears here:

`double_angle_offset_failure`.

If semantics pass but centered capacity fails:

`double_angle_pre_rescale_centered_capacity_failure`.

Stop before calling Rescale in the capacity-failure case.

## D4.i — after Fast Rescale

Execute exactly:

`FastCKKS.Rescale(res, res)`.

Compare against unchanged plaintext recurrence value:

`Y_{i+1} = O_i`.

Record:

- pre/post Level;
- expected level consumption from actual CKKS parameters;
- pre/post Scale;
- exact dropped modulus sequence used by the operation if available from parameters;
- semantic max component error;
- semantic perturbation D3.i -> D4.i;
- q0/q1 hashes;
- corrected post-rescale coefficient-domain capacity.

If first failure occurs here:

`double_angle_rescale_failure`.

Only after D4.i passes may the task enter round `i+1`.

---

# Error-growth evidence

For each round retain a compact table containing:

- `round`;
- `checkpoint`;
- expected operation;
- Level / Degree / Scale;
- source-backed max component error;
- error delta and ratio from prior checkpoint when meaningful;
- capacity ratio / outside count;
- worst slot/component for semantic error;
- worst coefficient component/index for capacity.

If all individual operations pass but error increases gradually, do not invent a primitive failure. Record the accumulated error trajectory.

---

# DA-F0 — end-of-recurrence state before final metadata reset

After the last D4 checkpoint, compare the ciphertext against the independently iterated final recurrence vector `Y_final`.

Require:

- semantic error <= `1e-2`;
- corrected coefficient-domain capacity outside count `0`;
- expected final Level;
- Degree 1;
- q0/q1 hashes recorded.

Also compare the final actual Scale `S_DA` with `Mod1Parameters.ScalingFactor()`.

Record:

`Log2Delta(S_DA, ScalingFactor)`.

Do not silently normalize this Scale before recording it.

If recurrence semantics fail only at the final aggregate check despite all individual checkpoints passing:

`double_angle_accumulated_drift_failure`.

---

# DA-F1 — final `res.Scale = inputScale` metadata reset

This operation changes **metadata only** and therefore changes the decoded plaintext interpretation.

Let:

- `S_DA` = Scale immediately before reset;
- `S_in` = original Mod1 input ciphertext Scale captured before the initial Mod1 scale reinterpretation;
- `Y_final` = source-defined DoubleAngle recurrence plaintext before reset.

Because coefficients do not change, the exact expected plaintext after metadata reset is:

`Y_reset = Y_final * (S_DA / S_in)`.

Then perform exactly:

`res.Scale = S_in`.

Require:

1. q0/q1 hashes are unchanged;
2. Level and Degree are unchanged;
3. Scale equals exact `S_in`;
4. decoded ciphertext matches `Y_reset` within `1e-2`;
5. the measured decoded change agrees with the metadata scale ratio.

Record the semantic perturbation and scale ratio explicitly.

If this fails:

`double_angle_final_scale_reset_failure`.

This checkpoint validates the source-defined metadata reinterpretation only; do not proceed to later bootstrap stages.

---

# Required classification

Choose exactly one:

- `FIRST_SUPPORTED_CAUSE = logn13_double_angle_and_scale_reset_validated`
- `FIRST_SUPPORTED_CAUSE = double_angle_round_input_already_wrong`
- `FIRST_SUPPORTED_CAUSE = double_angle_round_input_capacity_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_multiply_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_doubling_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_offset_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_pre_rescale_centered_capacity_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_rescale_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_accumulated_drift_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_final_scale_reset_failure`
- `FIRST_SUPPORTED_CAUSE = double_angle_precondition_mismatch`

Also record:

- first failing round, or `none`;
- first failing checkpoint, or `none`;
- last passing round/checkpoint;
- final source-backed semantic error before reset;
- final metadata-reset semantic error.

---

# Artifact discipline

Create compact artifacts only:

- `results/FIX-001-P3-DIAG-DOUBLE-ANGLE-logN13.json`
- `results/FIX-001-P3-DIAG-DOUBLE-ANGLE-logN13-summary.json`

The summary must remain human-reviewable.

Store only:

- exact provenance;
- canonical P4 precondition evidence;
- DoubleAngle count and constants;
- compact per-round checkpoint table;
- corrected capacity evidence;
- end-of-recurrence scale comparison;
- final metadata-reset evidence;
- classification;
- tests / clean-state evidence.

Do not serialize full slot vectors, coefficient arrays, repeated indices, operation traces, or large hash collections.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- focused DoubleAngle diagnostic tests pass;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Secondary changes;
- compact Primary artifacts are committed and pushed normally;
- both worktrees clean;
- no lower-scale sweep;
- no LogN16;
- no benchmark;
- no Gate 4/5 production run;
- no EXP-003;
- no later bootstrap stage after final Mod1 scale reset.

The deliverable is whether the validated `2^91` compressed polynomial output remains correct through the complete LogN13 Fast DoubleAngle recurrence and its final source-defined metadata Scale reset.