# FAST-STANDARD-NUMERICAL-FIX-001 — Remove Fast-only normalized EvalMod path

## Status

Executable correctness repair task.

## Task class

`C — Numerical Correctness Repair`

## Architectural authority

This task follows the permanent Fast/Q-prefix architecture rather than preserving a historical workaround.

Authoritative architecture requirements from the Secondary specifications:

- Fast should preserve Normal/Standard behavior and public structure where practical.
- Fast follows Standard CKKS Level/Scale progression.
- Fast Rescale changes how the result is computed, not what Rescale means.
- Q-prefix changes which real logical-q residue rows are maintained, not the CKKS mathematical algorithm.
- The Q-prefix v2 production policy is `w_Q(level)=min(level+1,4)`.
- A capacity failure must be reported concretely; do not silently redesign CKKS semantics around a hypothetical capacity problem.
- Current key behavior is approximately error-only / `(e,0)`: the large `a` mask contribution is removed, while sampled error is retained.

The current special `normalizedLogN13` EvalMod schedule is not part of the permanent architecture and is now considered a historical workaround from the narrower Fast residue design.

## Accepted numerical diagnosis

From clean DIAG-002 reproduction:

- C2S real/imag match Standard to approximately 2.4e-14 RMSE.
- First outer material divergence occurs in EvalMod.
- Standard Bootstrap SNR: 144.71339028083412 dB.
- Fast Bootstrap SNR: -0.17249085395397162 dB.
- Final Fast-vs-Standard complex RMSE: 0.02089221106956351.
- S2C is downstream amplification, not the origin.

DIAG-003 semantic-alignment work is superseded by this repair direction. Do not spend another task preserving or tuning the Fast-only normalized representation.

## Goal

Make Fast Q-prefix EvalMod follow the same mathematical Mod1 sequence and Scale/Level schedule as Standard Lattigo, while retaining Fast Q-prefix arithmetic/storage.

The intended conceptual structure is:

[
	ext{Normalize-to-Mod1}
ightarrow
	ext{Chebyshev offset}
ightarrow
	ext{Polynomial at Standard target scale}
ightarrow
	ext{Standard-equivalent DoubleAngle recurrence}
ightarrow
	ext{Standard-equivalent Rescale progression}
ightarrow
	ext{restore input/public scale}
]

Fast may use different low-level primitives to implement these operations, but it must not use a different mathematical recurrence merely to preserve an obsolete q0/q1 capacity workaround.

## Required production change

In Secondary `circuits/ckks/mod1/fast.go`:

1. Remove the production dispatch to the special normalized LogN13 path:
   - `normalizedLogN13Profile`;
   - `normalizedLogN13PlanScaleBitsForParams`;
   - `evaluateNormalizedLogN13`;
   - special plan bits 91/93;
   - working exponent 31/33;
   - k exponent 29/27;
   - multiplier exponent 30/28;
   - metadata-only coherent-scale reinterpretation;
   - normalized DoubleAngle multiplier/constant compensation;
   - final power-of-two normalized restore.

2. The supported LogN13 q0=55 and q0=56 profiles must use the same generic Fast Mod1 algorithmic schedule as Standard:
   - initial `res.Scale = mod1Params.ScalingFactor()`;
   - identical target-scale derivation from the logical Q schedule;
   - identical Chebyshev variable offset;
   - polynomial evaluation targeting the same `targetScale` as Standard;
   - DoubleAngle recurrence equivalent to:
     [
     y leftarrow 2y^2-c
     ]
     with the same Standard scale/level evolution;
   - Fast `RescaleQPrefixRows` for the physical implementation;
   - final `res.Scale = inputScale`, followed by the existing public Bootstrap default-scale boundary.

3. Preserve Fast/Q-prefix backend rules:
   - use only authoritative Q-prefix rows;
   - no Standard full-RNS fallback;
   - no reading stale dormant rows;
   - no QP/Gadget/key-switch reintroduction;
   - retain current zero-a/error-retaining key mode;
   - no application/frontend changes.

4. Do not change Standard Lattigo production code.

## Capacity policy

Do **not** introduce another Scale workaround proactively.

The Q-prefix architecture already keeps q0123 at high logical levels specifically to provide a larger authoritative centered range.

Use the existing capacity observer/guards as assertions.

For every existing EvalMod capacity checkpoint, record:

- logical Level;
- active Q-prefix row count;
- exact prefix product;
- exact/observed per-component bound;
- strict `2B < S_Q(level)` result.

If the Standard-equivalent schedule violates the Q0123 capacity contract, stop at the first concrete failing operation and report:

- checkpoint;
- component;
- bound (B);
- available prefix capacity (S_Q/2);
- exact deficit.

Do not add a substitute normalization scheme in this task.

A real capacity failure is a blocking result to be reviewed separately.

## Key/noise semantics

Do not describe Fast as "noise-free" or as using a zero public key.

For the current implementation:

- the large random `a` / mask contribution is zeroed or elided;
- the public-key target is approximately `(e_pk, 0)`;
- sampled error/noise is retained;
- other intentional error terms documented by the current Fast key path remain part of the numerical execution.

This repair must not remove those error terms.

## Required tests

### A. Source/architecture

Prove there is no production `normalizedLogN13` special dispatch remaining in Fast Mod1.

If helper code becomes dead, delete it rather than leaving an unused alternative production algorithm.

### B. Stage equivalence

On the canonical LogN13 q0=55 workload, compare Fast and genuine Standard after:

- EvalMod input normalization;
- Chebyshev offset;
- polynomial output;
- every DoubleAngle round after multiply/add/constant where semantically comparable;
- every Rescale;
- EvalMod output;
- final Bootstrap output.

At checkpoints with the same mathematical contract, compare directly without representation-specific power-of-two alignment.

### C. Numerical gates

The repair is successful only if all of the following hold:

1. Fast EvalMod output no longer has the previous ~1e-2-scale divergence.
2. Final Fast-vs-Standard RMSE improves by at least 1000x from:
   [
   0.02089221106956351.
   ]
3. Fast Bootstrap SNR becomes positive and materially closer to Standard.
4. Public output Level/Degree/Scale/NTT contract remains unchanged.
5. Repeated Fast execution remains deterministic for the canonical workload.

Do not weaken existing correctness thresholds merely to pass this task.

### D. Regression

Run:

- Secondary `go test ./schemes/ckks/fast -count=1`;
- relevant `circuits/ckks/mod1` Fast tests;
- relevant `circuits/ckks/bootstrapping` Fast tests;
- Primary `go test ./cmd/fastdiag ./internal/numericalmetrics`;
- canonical numerical harness;
- `git diff --check`.

If unrelated known tests fail, report them without modifying or bypassing them.

## Required result artifact

Write:

`results/FAST-STANDARD-NUMERICAL-FIX-001-summary.md`

Report:

- before/after final RMSE;
- before/after Fast Bootstrap SNR;
- Standard Bootstrap SNR;
- per-stage Fast-vs-Standard RMSE;
- capacity audit for the Standard-equivalent Fast EvalMod schedule;
- source files changed;
- proof that sampled error/noise behavior was not intentionally removed;
- all tests;
- clean commit provenance.

## Classification

Return exactly one:

- `FAST_STANDARD_EVALMOD_PARITY_RESTORED`
- `FAST_STANDARD_EVALMOD_IMPROVED_BUT_NOT_RESTORED`
- `FAST_STANDARD_SCHEDULE_BLOCKED_BY_QPREFIX_CAPACITY`
- `FAST_STANDARD_EVALMOD_REPAIR_UNCLOSED`

If capacity blocks the direct Standard-equivalent schedule, do not invent a workaround in this task.

## Completion

Return:

`FAST_STANDARD_EVALMOD_FIX_READY`

then the classification and key metrics.

Then:

`READY_FOR_WEB_REVIEW`.
