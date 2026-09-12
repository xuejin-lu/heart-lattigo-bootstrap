# FIX-001-P3-INTEGRATE-LOGN13-NORMALIZED-MOD1

## Purpose

The LogN13 normalized DoubleAngle design is now validated end-to-end in diagnostics.

Authoritative accepted design result:

- Primary commit: `dc1183570a2967ab3057a1305c7b18d7a7f0d61a`
- Classification: `normalized_double_angle_recurrence_validated_after_final_reference_fix`
- first failing round/checkpoint: `none / none`
- all three DoubleAngle rounds matched the normalized full-RNS reference exactly at input, square, multiplier, constant, and post-Rescale checkpoints;
- all square-capacity bounds passed;
- all pre-Rescale exact q0/q1 centered-capacity checks passed;
- final `K3` restore matched the normalized full-RNS reference exactly;
- final metadata-only scale reset preserved rows and passed semantic validation.

The production Secondary `circuits/ckks/mod1/fast.go` still uses the old Standard-style physical target-scale polynomial result followed by the ordinary DoubleAngle recurrence. That path is known to fail because physical target-scale promotion causes q0/q1 square alias at the first DoubleAngle multiplication.

This task integrates only the already-validated LogN13 normalized design into the production Fast Mod1 path.

This is a two-repository task:

1. modify and commit/push Secondary `xuejin-lu/lattigo` first;
2. then verify the actual production `FastEvaluator.EvaluateNew` from Primary against the accepted diagnostic result and commit/push Primary evidence.

No LogN16, benchmark, Gate 4/5, EXP-003, or later bootstrap stage.

---

# Fixed provenance

Primary repository:

`xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`dc1183570a2967ab3057a1305c7b18d7a7f0d61a`

Secondary repository:

`xuejin-lu/lattigo`

Required Secondary starting point:

`61607bb4bb82591009ce768d9a3773bed1497565`

Secondary branch:

`fast-ckks`

Accepted LogN13 profile:

- LogN = 13;
- bootstrap LogQ bit sizes = `[55,39,40,39,40,60,60,61,60,60,60,61,61,56,57,56,56]`;
- Mod1 degree = 30;
- DoubleAngle = 3;
- Mod1 type = `CosDiscrete`;
- K = 16;
- LogMessageRatio = 10;
- no inverse polynomial;
- levels consumed per rescaling = 1;
- accepted compressed PS plan scale = `2^91`;
- accepted normalized factor exponent at the validated run = 29 (`K_i = 2^29`);
- accepted DoubleAngle multiplier at the validated run = `2^30` for all three rounds.

The implementation must derive/check runtime values where practical, but it must not invent a new general compressed-scale search algorithm in this task.

---

# Scope lock

Allowed Secondary production files are narrowly bounded to:

- `circuits/ckks/polynomial/fast.go` and focused tests;
- `circuits/ckks/mod1/fast.go` and focused tests;
- tiny adjacent Fast-only helpers if strictly necessary.

Do not modify:

- Standard CKKS evaluator behavior;
- Standard Mod1 behavior;
- `schemes/ckks/fast` multiplication / relinearization / Rescale primitives;
- key generation;
- bootstrapping stage order;
- parameters;
- polynomial coefficients;
- threshold;
- public frontend/runtime backend selectors.

Do not add `--fast`, `--normal`, environment backend switches, or equivalent runtime selectors.

Do not generalize this change to LogN16 or arbitrary parameter sets in this task.

---

# P0 — safe two-repository startup

Primary startup follows Primary `AGENTS.md`.

Secondary modification is explicitly authorized by this Primary active task. In Secondary:

1. require clean `fast-ckks` at exact `61607bb4bb82591009ce768d9a3773bed1497565` before editing;
2. follow Secondary `AGENTS.md`, `CURRENT_TASK.md`, and `docs/FAST_CKKS_SPEC.md`;
3. do not reset/stash/discard unrelated work;
4. do not modify dormant higher Fast limbs as if they were authoritative.

If Secondary is not at the required clean starting point, stop and report.

---

# P1 — add a bounded Fast polynomial plan-scale override

The accepted diagnostic polynomial path did **not** simply call the existing production evaluator with the Standard target Scale.

It did:

1. build the ordinary Paterson–Stockmeyer plan using the source-defined Standard target Scale;
2. override every `plan.Value[i].Scale` with the accepted compressed candidate `2^91`;
3. evaluate the plan;
4. let the existing `evaluatePlan` perform its existing final Fast Rescale;
5. obtain the low-coefficient result at approximately `2^31` Scale.

Production integration must reproduce that exact boundary without duplicating polynomial arithmetic in Mod1.

## Required API shape

Preserve the existing public behavior of:

`FastEvaluator.Evaluate(input, p, targetScale)`.

Refactor internally so a second bounded entry point can evaluate the same PS plan with an explicit plan-scale override while retaining the ordinary planner target Scale.

Acceptable shape, naming may differ:

```go
EvaluateWithPlanScale(input, p, targetScale, planScale rlwe.Scale)
```

Requirements:

- planner still receives the source-defined `targetScale`;
- only `plan.Value[i].Scale` is overridden;
- power generation remains unchanged;
- PS decomposition remains unchanged;
- existing final `evaluatePlan` Rescale remains exactly once;
- no new rescale is added by Mod1 after the polynomial evaluator returns;
- existing `Evaluate` behavior and tests remain unchanged.

Add focused tests proving:

- ordinary `Evaluate` is unchanged;
- override does not alter plan decomposition/levels unexpectedly;
- override output uses the existing final Rescale path;
- input remains unchanged;
- dormant higher residues are not read as authoritative by Fast arithmetic.

---

# P2 — narrowly select the validated production path

In `circuits/ckks/mod1/fast.go`, keep the existing path for parameter profiles not covered by this task.

Use the normalized production path only when the evaluator matches the accepted LogN13 profile.

At minimum guard:

- `params.LogN() == 13`;
- `Mod1Poly.Degree() == 30`;
- `DoubleAngle == 3`;
- `Mod1Type == CosDiscrete`;
- `Mod1InvPoly == nil`;
- one level consumed per Rescale;
- expected LogQ structural profile for the accepted LogN13 bootstrap chain;
- the resulting polynomial low state and normalized schedule satisfy all runtime invariants below.

Do **not** silently apply the `2^91` override to unrelated profiles.

Do not introduce a search over candidate scales.

For the accepted profile, use:

`compressedPlanScale = 2^91`.

Keep this constant private to the bounded Fast Mod1 implementation and document that it is the currently validated LogN13 Stage-A profile value.

---

# P3 — production polynomial boundary

For the accepted profile:

1. preserve `inputScale := ct.Scale`;
2. clone q0/q1-authoritative input as today;
3. perform the same source-defined Mod1 scale reinterpretation and Chebyshev offset as today;
4. compute the Standard source-defined `targetScale` exactly as today;
5. call the new Fast polynomial plan-scale override using:
   - planner target = source-defined `targetScale`;
   - plan scale = `2^91`;
6. accept the polynomial evaluator's returned ciphertext **after its existing final Rescale** as the normalized low-coefficient state.

Runtime invariants for the validated profile:

- Degree = 1;
- Level = 7 for the accepted configuration;
- NTT = true;
- Montgomery = true;
- c1 remains zero under the current zero-secret Fast contract;
- returned Scale is near the accepted working Scale (~`2^31`);
- no physical multiplication by `targetScale / workingScale` occurs here.

A violation must return a descriptive error rather than falling back to the known-bad physical target-promotion path.

---

# P4 — metadata normalization factor

Let the returned low polynomial Scale be `W`.

Let the source-defined Standard polynomial target be `T`.

Choose the normalized scale factor as the nearest power-of-two exponent exactly as in the validated design:

```text
kExp = round(log2(T / W))
K = 2^kExp
S0 = W * K
```

For the accepted LogN13 profile the validated value is:

`kExp = 29`, `K = 536870912`.

Require at runtime:

- `kExp == 29` for this bounded profile;
- `S0` is sufficiently close to `T` using a fixed precision-safe scale comparison (do not use raw float64 equality);
- the mismatch remains within at least 32 relative bits of agreement;
- coefficients are **not** multiplied by K at this stage.

Then perform metadata-only normalization:

`res.Scale = S0`.

This intentionally reinterprets the low physical coefficients as `z0 = y0 / K` while keeping q0/q1 coefficient magnitudes small.

Do not set `res.Scale = T` unless it is exactly the coherent `W*K` value; the validated recurrence carries the coherent actual Scale schedule.

---

# P5 — normalized DoubleAngle recurrence

Replace the ordinary Fast DoubleAngle operations only inside the accepted bounded path.

Maintain:

- `workingScale = W`;
- current coherent metadata Scale `S_i`;
- current power-of-two normalization exponent `k_i`;
- source constant recurrence using `sqrt2pi *= sqrt2pi` before each round.

For each round i:

1. determine the Scale that the source-defined Fast Rescale will produce from `S_i^2` using the actual current level and actual dropped modulus;
2. compute:

```text
k_{i+1} = round(log2(S_{i+1} / W))
A_i exponent = 1 + 2*k_i - k_{i+1}
A_i = 2^(A_i exponent)
normalized constant = C_i / 2^(k_{i+1})
```

where `C_i` is the ordinary source DoubleAngle constant after squaring `sqrt2pi` for that round.

For the accepted run all `k_i` are 29 and all `A_i` are `2^30`, but derive and validate rather than blindly assume inside each round.

Execute exactly:

```text
MulRelin(res, res)
MulIntegerMaintained(res, A_i)
Add(res, -normalizedConstant)
Rescale(res)
```

Do **not** separately execute the old `Add(res,res)` doubling; the factor 2 is already included in `A_i`.

Do **not** physically restore K between rounds.

After each Rescale:

- carry the actual resulting Scale forward;
- require expected level consumption exactly one level;
- require the newly derived `k_{i+1}` remains the validated value 29 for this profile;
- preserve Degree 1, NTT, Montgomery, and q0/q1-only arithmetic.

If an invariant fails, return a descriptive error. Do not silently fall back to ordinary DoubleAngle after entering the normalized path.

---

# P6 — final K restore and source-defined metadata reset

After the third normalized DoubleAngle round:

1. let `K3 = 2^k3` (validated `2^29`);
2. execute:

`FastCKKS.MulIntegerMaintained(res, K3, res)`;

3. do not change Scale during the integer multiplication;
4. then perform the original source-defined final metadata-only reset:

`res.Scale = inputScale`.

No further physical multiplication, Rescale, or DoubleAngle operation occurs after K3 restore.

The accepted diagnostic proved this final restore has enormous q0/q1 centered-capacity margin.

---

# P7 — production safety checks

Production code should fail closed for the bounded normalized path if any of the following occurs:

- compressed polynomial output geometry differs from the validated profile;
- `kExp != 29`;
- coherent Scale is not sufficiently close to Standard target Scale;
- a DoubleAngle round would consume an unexpected number of levels;
- `A_i` exponent is negative or differs from validated bounded assumptions;
- final level differs from the expected LogN13 production level;
- NTT/Montgomery representation changes unexpectedly.

Do not add expensive full coefficient-domain CRT capacity scans to the production hot path. Those were diagnostic proofs, not production operations.

The production path relies on the accepted diagnostic capacity proof for this exact bounded profile.

---

# P8 — Secondary tests

Add focused tests in Secondary.

## Required production regression

Add a LogN13 degree-30 / DoubleAngle=3 regression matching the accepted bootstrap Mod1 profile closely enough to exercise the new bounded path.

Require:

- `EvaluateNew` succeeds;
- output Degree = 1;
- expected final Level = 4 for the accepted Mod1 segment;
- output Scale equals original Mod1 input Scale after final reset;
- NTT/Montgomery remain true;
- input unchanged;
- dormant higher limbs do not affect q0/q1 output;
- no NaN/Inf;
- semantic agreement with an appropriate Standard/reference path within `1e-2` where that comparison is meaningful.

Also preserve all existing Fast Mod1 tests for non-bounded profiles.

## Polynomial override tests

Test the new explicit plan-scale override independently enough to catch:

- accidental double Rescale;
- forgotten final Rescale;
- target/plan scale argument swap;
- input mutation.

Run Secondary:

```bash
go test ./circuits/ckks/polynomial ./circuits/ckks/mod1

go test ./...
```

Commit and push Secondary before generating Primary production evidence.

Record the new Secondary commit SHA.

---

# P9 — Primary production verification

After Secondary is committed and pushed, Primary must verify the **actual production** `bootstrapping.FastEvaluator.Mod1Evaluator.EvaluateNew` using the new Secondary SHA.

Do not validate by replaying the diagnostic recurrence as the system under test.

Use the existing reproducible LogN13 frontend up through the real Mod1 input, then call production Fast Mod1 `EvaluateNew` once.

Compare production output against the accepted normalized diagnostic oracle/evidence.

Required checks:

- production call succeeds;
- final Level = 4;
- Degree = 1;
- NTT/Montgomery true;
- final Scale exactly equals original Mod1 input Scale;
- q0/q1 final row hashes equal the accepted normalized final production expectation if the harness can obtain the identical deterministic state;
- otherwise, at minimum compare production output against a freshly generated normalized full-RNS stage-aligned oracle with exact q0/q1 equality;
- final semantic max-component error <= `1e-2`;
- ordinary coherent/exact-target semantic compatibility <= `1e-2`;
- input ciphertext unchanged;
- Secondary provenance is the new production commit and clean.

The production verification must not directly call private diagnostic arithmetic instead of production `EvaluateNew`.

---

# P10 — compact Primary artifact

Create only compact evidence:

- `results/FIX-001-P3-INTEGRATE-LOGN13-NORMALIZED-MOD1-summary.json`

No large raw artifact is required unless a failure needs a bounded debugging record.

Summary should include:

- Primary generation commit/provenance;
- Secondary production commit SHA;
- profile guard disposition;
- compressed plan scale (`2^91`);
- polynomial returned low Scale;
- derived K exponent/value;
- per-round Level/Scale/K/A/constant compact schedule;
- final Level/Degree/Scale;
- final q0/q1 hashes;
- semantic error;
- Standard/coherent compatibility error;
- input immutability;
- test results;
- forbidden-scope confirmations.

---

# Required final classification

Choose exactly one:

- `FIRST_SUPPORTED_CAUSE = logn13_normalized_fast_mod1_production_integrated`
- `FIRST_SUPPORTED_CAUSE = production_polynomial_plan_scale_integration_failure`
- `FIRST_SUPPORTED_CAUSE = production_normalized_schedule_invariant_failure`
- `FIRST_SUPPORTED_CAUSE = production_normalized_double_angle_arithmetic_failure`
- `FIRST_SUPPORTED_CAUSE = production_final_restore_or_scale_reset_failure`
- `FIRST_SUPPORTED_CAUSE = production_output_semantic_failure`
- `FIRST_SUPPORTED_CAUSE = production_integration_precondition_mismatch`

On success, this authorizes a **separate next task** to resume the real bootstrap path after Mod1 in LogN13. It does not authorize LogN16 or benchmarking in this task.

---

# Commit / push order

1. Secondary implementation + tests;
2. Secondary full test suite;
3. commit and push Secondary `fast-ckks`;
4. record new Secondary SHA;
5. Primary production verification against that exact SHA;
6. Primary tests;
7. commit and push Primary artifact/supporting verification code;
8. verify both worktrees clean and remote heads synchronized.

Do not leave Primary pointing at an uncommitted Secondary working tree.

---

# Final validation

Before completion require:

- Secondary focused polynomial/Mod1 tests pass;
- Secondary `go test ./...` passes;
- Primary focused production integration test passes;
- Primary `go test ./...` passes;
- production Fast Mod1 `EvaluateNew` is the system under test;
- Secondary production code actually contains the normalized LogN13 path;
- Standard behavior unchanged;
- no backend runtime selector introduced;
- both worktrees clean;
- both remote heads synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no CoeffsToSlots or later bootstrap stage after the verified Mod1 output.