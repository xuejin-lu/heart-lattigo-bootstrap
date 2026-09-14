# FIX-001-P3-INTEGRATE-LOGN13-B4-T2-ONE-BIT-SCALAR-GUARD

## Goal

Integrate the already validated LogN13 B4 final-parent one-bit scalar guard into the **Secondary Fast CKKS production path**, then prove from the Primary repository that the ordinary production bootstrap reproduces the successful diagnostic behavior.

This task is no longer exploratory. The accepted feasibility evidence at Primary commit

`3bc71929c4b1a9196088dbb20dec6b50f105e5bc`

is:

- classification: `logn13_b4_t2_one_bit_scalar_guard_system_sufficient`;
- native public-like: `0.010683237260415292`;
- guarded public-like: `0.005554220603853743`;
- public margin: `0.004445779396146257`;
- native T2 scalar encoding scale: `4.2949672960014038e+09`;
- guarded encoding scale: `8.5899345920028076e+09`;
- guarded post-add Q01 capacity ratio: about `0.525396256`;
- contracted Q01 capacity ratio: about `0.262698128`;
- rounded divide-by-two, independent CRT, row, capacity and metadata contracts all pass;
- guarded B4 residual vs canonical is about `6.3683e-11`, improved from about `1.8004e-10` (~2.83x);
- x4 is invalid (`capacity ratio ~1.05079`, outside_count=1) and must not be implemented;
- T6+T2 is unnecessary and must not be implemented.

The production implementation must preserve this exact arithmetic mechanism:

1. one-bit physical/metadata promotion immediately before the final scalar accumulation of the G0 parent baby-step;
2. unchanged scalar `MulThenAdd` at the doubled accumulator scale;
3. centered symmetric rounded divide-by-two contraction in coefficient domain;
4. restore original Level/Scale/degree/NTT/Montgomery metadata;
5. continue the existing PS / G0 / DA / S2C / finalization path unchanged.

---

# Repository provenance and authorization

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required accepted-result ancestor:

`3bc71929c4b1a9196088dbb20dec6b50f105e5bc`

Start only from clean `main`, fetch and fast-forward pull only, then re-read fresh `AGENTS.md`, `CURRENT_TASK.md`, and this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required starting commit:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

Branch: `fast-ckks`

Worktree must be clean before modification.

**This Primary task explicitly authorizes the required minimal Secondary production-code changes and an ordinary fast-forward push to `origin/fast-ckks` after all Secondary validation passes.**

No history rewrite, force push, reset, stash, unrelated cleanup, or branch substitution.

---

# Scope lock

LogN13 production correctness only.

Keep unchanged:

- q0/q1/q2 profile 56/39/40;
- polynomial coefficients;
- PS decomposition;
- common plan scale and current normalized LogN13 schedule;
- generated powers;
- G0 merge policy;
- G0/F0/DA accepted local-q2 behavior;
- S2C;
- final restoration;
- Standard reference behavior;
- frontend/workload/config.

Forbidden:

- x4 scalar guard;
- T6 guard;
- coefficient-specific `rounded_integer +/- 1` hacks;
- changing the mathematical T2 coefficient;
- global guard policy for all scalar MulThenAdd operations;
- changing ordinary `FastEvaluator.Evaluate` semantics;
- changing ordinary unguarded `EvaluateWithPlanScale` semantics for unrelated callers;
- q changes;
- plan-scale retuning;
- generated-power redesign;
- LogN16;
- benchmarks / Gate4 / Gate5 / EXP003.

---

# P1 — Secondary production API shape

The existing production path already gates the special normalized workload with:

`circuits/ckks/mod1/fast.go: normalizedLogN13Profile(...)`.

Use that existing profile gate. Do **not** make the generic Fast polynomial evaluator silently guard every workload.

Add the smallest explicit polynomial capability needed by the normalized LogN13 Mod1 path, for example an API semantically equivalent to:

`EvaluateWithPlanScaleFinalParentOneBitScalarGuard(...)`

Exact naming is implementation choice, but requirements are:

- ordinary `Evaluate(...)` remains unchanged;
- ordinary `EvaluateWithPlanScale(...)` remains unchanged;
- only the normalized LogN13 Mod1 branch calls the guarded variant;
- the polynomial implementation identifies the target **structurally** from the PS plan:
  - the last planned baby-step that becomes the first G0 parent branch;
  - its final non-zero non-integer scalar accumulation;
- do not hardcode diagnostic names such as `B4` in production arithmetic;
- do not hardcode the coefficient decimal value;
- the current normalized LogN13 production test must prove the selected scalar degree is 2.

If the guarded API cannot unambiguously find exactly one final scalar operation, return an error rather than silently falling back.

---

# P2 — one-bit guard arithmetic

At the selected final-parent scalar operation, let the native accumulator be `A` with Scale `S`.

Perform:

1. `A <- 2*A` using authoritative q0/q1 maintained arithmetic;
2. `A.Scale <- 2*S`;
3. execute the existing unchanged scalar `MulThenAdd(power, coefficient, A)`;
4. contract by symmetric centered rounded divide-by-two;
5. restore Scale to exactly `S` and preserve Level/degree/dimensions/NTT/Montgomery state.

Do not alter the coefficient or scalar-rounding rule.

The current production implementation must therefore obtain the extra precision solely because the scalar encoding scale doubles.

For the accepted LogN13 path, a focused test must observe the guarded scalar integer corresponding to the doubled grid (currently `3704477843` for the real scalar), without hardcoding that integer into the arithmetic itself.

---

# P3 — production rounded divide-by-two primitive

The feasibility implementation proves correctness using coefficient-domain q0/q1 recovery, signed CRT, symmetric rounding and NTT/Montgomery restoration.

Production must preserve those semantics.

Required sequence per ciphertext component:

1. partial INTT of authoritative q0/q1;
2. normalize Montgomery form as needed;
3. recover the unique centered integer represented by q0/q1;
4. symmetric nearest-integer divide-by-two:
   - positive odd: `(x+1)/2`;
   - negative odd: `(x-1)/2`;
   - even: exact `x/2`;
5. reconstruct q0/q1 residues;
6. NTT;
7. restore Montgomery form;
8. preserve all non-scale metadata;
9. set Scale to exactly half the guarded Scale.

### Hot-path implementation requirement

Do not copy the diagnostic per-coefficient heap-heavy `big.Int` implementation into production if avoidable.

For q0<=56 bits, q1<=39 bits, Q01<2^95, implement the signed two-limb CRT/sign/parity arithmetic using bounded integer arithmetic (`math/bits` / reusable scratch / equivalent allocation-free representation) where practical.

A test-only independent `big.Int` oracle is required to validate the production primitive across:

- positive/negative values;
- even/odd values;
- values near zero;
- values near but strictly inside +/-Q01/2;
- both ciphertext components;
- NTT/Montgomery round trip.

If a compact allocation-free implementation cannot be made safely, correctness wins: use a carefully scratch-reused implementation rather than an unverified shortcut. Do not substitute modular inverse-of-two for rounded centered division.

---

# P4 — production safety / invariant checks

This feature is enabled only under `normalizedLogN13Profile(...)`.

The guarded polynomial path must assert structural invariants sufficient to ensure it is executing the intended plan location, including at minimum:

- plan-scale override is active;
- selected baby-step is the final planned parent step;
- selected operation is the final non-zero scalar accumulation;
- selected coefficient is non-integer;
- selected scalar degree is 2 for the current normalized LogN13 profile;
- input/output Level remains the native baby-step Level;
- contraction returns the native Scale;
- result Degree/NTT/Montgomery metadata match the unguarded operation contract.

If these invariants change, fail loudly rather than applying the guard to another operation.

Do not add expensive debug-only full capacity scans to the ordinary hot path unless required for correctness. Capacity safety for this fixed production profile must instead be covered by focused deterministic tests reproducing the accepted feasibility values and by the existing profile gate.

---

# P5 — Secondary tests

Add focused Secondary tests covering at least:

### Primitive correctness

Production centered rounded divide-by-two vs independent big.Int oracle.

### Polynomial selection

The guarded plan selects exactly one final-parent scalar operation and, for normalized LogN13 Mod1, it is degree 2.

### Guard metadata

- Scale doubles before guarded scalar add;
- Level unchanged;
- Degree unchanged;
- contraction restores native Scale;
- NTT/Montgomery state preserved.

### Guard numerical behavior

On the accepted LogN13 setup:

- native scalar encoding scale near `4.2949672960014038e+09`;
- guarded scale near `8.5899345920028076e+09`;
- post-guard Q01 capacity remains unique with ratio near `0.525396256`;
- contracted ratio near `0.262698128`;
- x4 remains rejected by evidence/tests and is not executed by production.

### Regression

- unguarded `EvaluateWithPlanScale` output is unchanged;
- non-normalized Mod1 path is unchanged;
- existing Fast polynomial / Mod1 tests pass.

Run full Secondary:

`go test ./...`

and:

`git diff --check`.

Only after all pass, commit and ordinary fast-forward push `fast-ckks`.

Record the resulting Secondary production commit SHA for Primary validation.

---

# P6 — Primary production integration validation

After Secondary is committed/pushed, return to Primary and ensure it is still clean except for intentional current-task work.

Primary validation must exercise the **ordinary production Fast bootstrap path** using the updated Secondary. It must not inject the diagnostic guarded B4 state or manually call the feasibility helper.

Create a compact production integration artifact:

`results/FIX-001-P3-INTEGRATE-LOGN13-B4-T2-ONE-BIT-SCALAR-GUARD-summary.json`

Include:

- Primary base/provenance;
- new Secondary production commit SHA;
- production LogN13 config;
- production final Level/Scale/degree;
- production EvalMod real/imag vs Standard;
- production post-S2C error;
- production public-like error;
- threshold `1e-2`;
- public margin;
- feasibility guarded public-like reference `0.005554220603853743`;
- difference production vs feasibility reference;
- guard structural-selection evidence (final parent / scalar degree 2);
- contraction/metadata test status;
- validation flags.

Do not serialize slot vectors or RNS rows.

### Required production result

`public_like <= 1e-2`.

Expected result should reproduce the feasibility guarded candidate closely, approximately:

`0.005554220603853743`.

Require production-vs-feasibility public-like difference <= `1e-10` unless a documented deterministic source of smaller harmless variation is demonstrated.

If production passes `1e-2` but differs materially from the feasibility candidate, classify as integration mismatch and stop; do not accept merely because it passes.

Required classification on success:

`logn13_b4_t2_one_bit_scalar_guard_production_integrated`

Failure classifications:

- `logn13_b4_t2_guard_production_selection_mismatch`
- `logn13_b4_t2_guard_production_contraction_mismatch`
- `logn13_b4_t2_guard_production_numeric_mismatch`
- `logn13_b4_t2_guard_production_public_threshold_failure`

---

# P7 — Primary validation

Require:

- a focused production integration test matching `TestFIX001P3.*B4.*T2.*Production.*Integration`;
- `go test ./...`;
- `git diff --check`;
- ordinary production LogN13 command / test path, not diagnostic injection;
- compact artifact has no NaN/Inf;
- Primary records the actual new Secondary SHA;
- Secondary `fast-ckks` is clean after push;
- Primary is clean after its final commit;
- both pushes are ordinary fast-forward only.

Primary may add only the minimal runner/test/artifact plumbing required to prove production integration. Do not alter frontend workload semantics.

---

# Completion state

If successful, there should be no remaining LogN13 correctness blocker from this chain.

Stop after production integration and validation.

Do **not** automatically start:

- generated-power redesign;
- LogN16;
- performance benchmarks;
- Gate4/Gate5;
- EXP003.

Report both final commit SHAs:

- Secondary Fast CKKS production commit;
- Primary integration/validation commit.
