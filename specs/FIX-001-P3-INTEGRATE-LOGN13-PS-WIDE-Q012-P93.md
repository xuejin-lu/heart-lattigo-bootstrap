# FIX-001-P3-INTEGRATE-LOGN13-PS-WIDE-Q012-P93

## Goal

Promote the successful LogN13 PS-wide three-limb design into the Fast backend.

The production candidate is:

- LogN13;
- q0 = 56-bit profile;
- q1/q2 from the same parameter builder (currently 39/40 bits);
- PS plan scale = `2^93`;
- q0/q1/q2 authoritative from PS entry through generated powers, baby steps, giant steps, relinearization, and final PS Rescale;
- exactly one Q012 -> Q01 contraction at PS exit after proving Q01 centered uniqueness;
- no q3+ arithmetic source;
- downstream normalized DoubleAngle uses the already validated bounded local-q2 arithmetic where required;
- public Fast Bootstrap must pass `max_component_abs <= 1e-2`.

This task ends PS design exploration. Do not reopen local G0/F0/T2 patching unless production evidence contradicts the PS-wide reference result.

---

## Why P93 is the production candidate

The accepted PS-wide diagnostic at Primary commit `41993cd03f85ec5f2de54dc21ade5ec812090d50` established:

- P92 public-like: about `0.0162977678` -> fail;
- P93 public-like: about `0.0097053814` -> pass;
- P94: downstream first failure at `round0.after_multiplier_capacity`;
- P93 Q012 capacity and PS-exit Q01 contraction pass;
- maximum observed Q012 capacity ratio across the sweep is about `0.0002439022`, with large remaining headroom;
- q2 is dropped only once at PS exit;
- q3+ are not arithmetic sources.

P93 therefore passes the actual system correctness target while preserving large Q012 capacity margin.

The historical PS-local `1.2e-8` polynomial budget is a planning/diagnostic heuristic derived from an amplification estimate. It remains useful evidence but is **not a hard production correctness gate** in this task. The hard correctness gate is full-system Bootstrap semantics (`<= 1e-2`) together with capacity, metadata, contraction, and deterministic regression checks.

Do not classify P93 as failed solely because its PS-vs-plaintext-oracle error is above `1.2e-8` if all system gates pass.

---

## Provenance / startup

Primary required ancestor:

`41993cd03f85ec5f2de54dc21ade5ec812090d50`

Secondary starting point:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

branch `fast-ckks`.

Follow both repositories' `AGENTS.md` startup rules. Before editing Secondary, re-read its `docs/FAST_CKKS_SPEC.md` and inspect current source.

Normal fast-forward task-completion pushes are authorized under the standing repository rules. Never force push or rewrite history.

---

# Architecture to integrate

## A. Parameter/profile boundary

The accepted active LogN13 profile is q0=56 for this design.

During implementation, use an in-memory/test profile first. Do not change the checked-in Primary LogN13 config until the production Fast path passes the required qualification below.

After the backend passes qualification:

1. update the checked-in LogN13 q0 baseline from 55 to 56;
2. run Standard and Fast through the same final checked-in config;
3. record exact generated q0/q1/q2 values and bit lengths.

This is an intentional new experiment baseline, not a hidden Fast-only frontend branch. Standard and Fast must see the same final application config.

## B. PS-wide Q012 authority

For the bounded normalized LogN13 Mod1/PS profile only:

1. lift the actual PS input from maintained Q01 into q2 once;
2. maintain q0/q1/q2 throughout all PS arithmetic;
3. generated Chebyshev powers use Q012 multiply-first arithmetic;
4. baby-step scalar/coefficient accumulation maintains all three limbs;
5. giant-step Rescale/multiply/aligned-add maintains all three limbs;
6. preserve the established Fast zero-secret degree/relinearization semantics;
7. final PS Rescale is still Q012-authoritative;
8. prove final centered value is Q01-unique;
9. contract Q012 -> Q01 exactly once at PS exit.

Do not use intermediate Q012 -> Q01 -> Q012 transitions.

Do not read dormant q3+ rows as arithmetic truth.

## C. Scale

For this bounded profile set the internal PS plan scale to exactly `2^93`.

Do not change unrelated polynomial workloads to P93.

Keep Standard CKKS logical Level/Scale progression and actual logical Rescale divisors.

## D. Production arithmetic requirements

Do not copy the diagnostic Big.Int implementation into the production hot path.

### Per-limb arithmetic

Add/Sub/Mul/scalar operations should remain RNS-native on q0/q1/q2.

### Three-limb centered reconstruction / Rescale

Only where a signed reconstruction is mathematically required (notably Fast logical Rescale / contraction), implement a fixed-width production path supporting the actual q0/q1/q2 bound (about 135 total bits for this profile).

Requirements:

- no per-coefficient `math/big.Int` allocation in the production hot path;
- fixed-width multiword arithmetic (`math/bits` or an equivalent allocation-free representation) is acceptable;
- exact signed centered interpretation under Q012;
- symmetric nearest rounded division by the actual logical `Q[level]`;
- reduce quotient back into maintained q0/q1/q2 residues;
- Standard Level/Scale metadata semantics;
- targeted reference tests against the existing Big.Int diagnostic oracle.

If a clean fixed-width Q012 Rescale cannot be implemented correctly in this task, stop with a precise implementation blocker. Do not silently fall back to Standard full-RNS or Big.Int production arithmetic.

## E. Generated powers

For this profile, supersede the q0/q1 balanced pre-Rescale power-generation workaround.

Use the proven PS-wide Q012 multiply-first schedule:

- multiply under Q012;
- apply Chebyshev recurrence/correction under Q012;
- perform the logical Rescale under Q012;
- keep q2 authoritative afterward because the whole PS remains Q012.

Do not contract each generated power to Q01.

## F. Old local PS patches

The PS-wide mode supersedes profile-specific local-q2/guard logic inside PS (generated-power local q2, G0/F0 local q2, B4/T2 one-bit scalar guard) when those mechanisms only existed to make a q01-authoritative PS survive.

Do not stack those old q01-local fixes on top of the new PS-wide Q012 evaluator unless source evidence proves one has an independent semantic role. Prefer one coherent PS-wide arithmetic model.

## G. DoubleAngle downstream prerequisite

The P93 system result depends on the already validated normalized DoubleAngle capacity handling after the polynomial stage.

Integrate the bounded local-q2 DoubleAngle arithmetic required by the accepted LogN13 path if it is not already production-active at the synchronized Secondary HEAD.

Requirements:

- preserve the validated normalized recurrence semantics;
- use q2 only for the bounded DA regions that require it;
- contract back to Q01 only after the corresponding round proves Q01 uniqueness;
- no full-RNS fallback;
- no q3+ arithmetic source.

Do not redesign DoubleAngle in this task.

---

# Implementation staging

## I0 — reproduce reference controls before editing

Reproduce, from the synchronized Primary diagnostic/reference code:

- Standard LogN13 control passes;
- accepted D0 q0=56/plan92 control around `0.005554220603853743`;
- PS-wide P93 reference around `0.0097053814`.

If these controls materially differ, stop before modifying Secondary.

## I1 — fixed-width Q012 primitives

Implement and unit-test the minimal Q012 production primitives required by PS-wide P93.

For every reconstruction/divide primitive compare against the diagnostic Big.Int oracle over:

- zero;
- +/-1;
- values near Q01/2;
- values beyond Q01/2 but inside Q012/2;
- values near Q012/2;
- actual captured LogN13 PS checkpoint values.

Require exact residue/quotient agreement.

## I2 — PS-wide production integration

Wire the bounded normalized LogN13 polynomial evaluator to the Q012 mode.

Verify compact checkpoints for both real and imag branches:

- T2/T3/T4/T6/T8/T16 generated powers;
- representative baby outputs;
- every giant-step Rescale/multiply/add boundary;
- final root;
- final PS Rescale;
- PS-exit contraction.

Record only aggregate capacity, hashes/first mismatch, levels/scales/degrees, and semantic metrics. Do not emit huge vectors.

## I3 — downstream DA integration

Ensure the actual production Mod1 path, not a Primary-only diagnostic helper, uses the accepted bounded local-q2 DA behavior where required.

## I4 — ordinary public API qualification

The main correctness run must invoke the ordinary public Fast path:

`FastEvaluator.Bootstrap`

and one-element `BootstrapMany`.

No diagnostic ciphertext injection may be used for the final acceptance result.

Standard and Fast must use the same final q0=56 checked-in config/workload.

---

# Robustness gate

P93's observed public margin is relatively small, so one deterministic workload is not sufficient for production promotion.

After the ordinary public path passes on the existing canonical workload, run at least these deterministic, recorded LogN13 message variants under the same parameter config:

1. canonical existing reproducible workload;
2. its exact negation;
3. a `0.5x` amplitude version of the same values;
4. one additional deterministic sign/phase permutation that preserves the accepted input magnitude/domain.

Do not use unrecorded randomness for the acceptance decision.

For **every** workload require:

- Standard public Bootstrap `<= 1e-2`;
- Fast public Bootstrap `<= 1e-2`;
- Fast `BootstrapMany` one-element result consistent with `Bootstrap`;
- correct final Level/Degree/Scale/NTT/Montgomery metadata;
- input immutability;
- PS Q012 capacity centered-unique at every checked boundary;
- PS-exit Q01 contraction valid.

Report the worst Fast public error and minimum margin to `1e-2`.

If canonical P93 passes but any robustness workload exceeds `1e-2`, classify as margin instability and stop. Do not begin benchmark or LogN16.

---

# Meaning of the 1.2e-8 PS metric

Continue reporting the final PS polynomial error against the high-precision/canonical oracle.

However:

- it is diagnostic evidence;
- it is **not** an independent hard failure if full system semantics, capacity, contraction, and robustness gates all pass;
- do not add artificial guards merely to force this number below `1.2e-8`.

This task's production correctness target is the actual public bootstrap contract.

---

# Secondary durable documentation

If integration succeeds, update Secondary `docs/FAST_CKKS_SPEC.md` to record:

- bounded normalized LogN13 PS-wide Q012 mode;
- q0=56 active profile requirement;
- q1/q2 are still parameter-builder primes, not universal fixed bit sizes;
- P93 internal plan scale for this profile;
- q2 is authoritative only for the documented PS interval (plus separately documented bounded DA windows);
- q3+ remain dormant;
- fixed-width Q012 Rescale implementation and supported modulus-size bound;
- old q01 balanced generated-power path remains for profiles not selecting this bounded mode.

Do not claim this is a universal CKKS rule.

---

# Acceptance / classification

Choose exactly one primary classification:

- `logn13_ps_wide_q012_p93_production_integrated`
- `logn13_ps_wide_q012_p93_fixed_width_primitive_mismatch`
- `logn13_ps_wide_q012_p93_production_path_mismatch`
- `logn13_ps_wide_q012_p93_ps_exit_contraction_failure`
- `logn13_ps_wide_q012_p93_downstream_da_failure`
- `logn13_ps_wide_q012_p93_public_semantic_failure`
- `logn13_ps_wide_q012_p93_margin_instability`
- `logn13_ps_wide_q012_p93_control_mismatch`
- `logn13_ps_wide_q012_p93_provenance_mismatch`

Success requires all ordinary public API + robustness gates above. The local `1.2e-8` PS metric alone cannot force a failure classification.

---

# Required tests

Secondary at minimum:

- focused Q012 reconstruction/rounded-divide tests;
- focused polynomial Fast tests;
- focused Mod1/DoubleAngle Fast tests;
- `go test ./circuits/ckks/polynomial ./circuits/ckks/mod1 ./circuits/ckks/bootstrapping`;
- `go test ./...`;
- `git diff --check`.

Primary at minimum:

- focused `TestFIX001P3` integration/regression tests;
- `go test ./...`;
- `git diff --check`;
- ordinary LogN13 production integration command producing one compact summary artifact.

No NaN/Inf.

---

# Artifact

Create one compact summary:

`results/FIX-001-P3-INTEGRATE-LOGN13-PS-WIDE-Q012-P93-summary.json`

Include only compact evidence:

- provenance and exact commits;
- final q0/q1/q2 values/bits;
- P93 reference control;
- fixed-width oracle agreement counts/first mismatch;
- max Q012 capacity ratio / worst checkpoint;
- PS final oracle metric (informational);
- PS-exit contraction proof;
- EvalMod real/imag;
- post-S2C;
- each deterministic workload public metric;
- worst public metric and margin;
- final metadata;
- Standard control;
- BootstrapMany consistency;
- classification and production readiness.

Do not write huge coefficient arrays, slot arrays, index lists, or repeated per-operation dumps.

---

# Stop conditions

Stop immediately on the first unsupported condition:

- dirty/incorrect repository provenance;
- fixed-width Q012 primitive disagrees with oracle;
- Q012 centered uniqueness failure;
- PS-exit Q01 contraction failure;
- production path does not reproduce the P93 reference behavior within normal numerical tolerance;
- any robustness workload public error > `1e-2`;
- Standard control failure;
- unexpected need for q3+ or full-RNS fallback.

Do not execute LogN16, benchmark, Gate4/5, EXP-003, or unrelated refactors.

---

# Push scope

This task explicitly authorizes the minimum required Secondary production changes and the corresponding Primary integration/config/regression changes.

After all required validation succeeds (or after producing a bounded failure artifact/diagnostic commit as required by the project workflow), ordinary fast-forward pushes are authorized under both repositories' standing rules.
