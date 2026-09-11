# FIX-001 — Capacity-safe Fast Chebyshev power generation

## Goal

Repair the confirmed Fast CKKS numerical correctness bug caused by using Standard Chebyshev power-generation ordering with a two-limb q0/q1-only backend.

Formal diagnosis established:

- `MulRelin(x,x) -> Rescale` passes;
- `MulRelin(x,x) -> double -> Rescale` passes;
- exact `T2 = 2*x^2 - 1 -> Rescale` fails;
- `2*x^2 -> Rescale -> -1` passes;
- low-scale exact T2 passes;
- at formal Mod1 scale, encoded `-1` exceeds `q0*q1/2` by about `1.34e8`;
- centered-CRT uniqueness is therefore mathematically false before the original T2 Rescale.

The fix must preserve Fast's core architecture:

> only q0/q1 are actively maintained; do not reintroduce q2+ arithmetic or full-Q fallback.

---

## Provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when this spec is authored:

`8346cb91245b758358037b503e540f926457f7ef`

Secondary repository: `xuejin-lu/lattigo`

Expected Fast baseline:

`fast-ckks` @ `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

Standard reference remains pinned at:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md` and Secondary `docs/FAST_CKKS_SPEC.md`.

If either worktree is dirty at startup, stop and report. Do not reset/stash/discard unrelated work.

---

## Root-cause constraint

The Standard common power basis computes Chebyshev powers as:

`2*T_a*T_b - T_|a-b|`, then Rescale.

That ordering is valid for Standard because the full RNS basis carries the high-scale intermediate.

Fast cannot safely copy that ordering when the high-scale correction term or resulting centered coefficient cannot be uniquely represented under q0/q1.

The repair must be **Fast-specific**. Do not change the Standard/common polynomial implementation.

---

## Required repair design

Implement a capacity-safe Fast Chebyshev power-generation ordering in:

`circuits/ckks/polynomial/fast.go`

Conceptually, for a newly generated Chebyshev power `T_n`, with `n=a+b` and `c=|a-b|`:

1. generate/prepare `T_a` and `T_b` exactly as required by the existing lazy/relinearization planner;
2. multiply them using the existing Fast path;
3. double the product;
4. perform the level-consuming Rescale **before** applying the Chebyshev correction term;
5. after the product is at its post-Rescale scale/level:
   - if `c == 0`, subtract scalar `1` at the post-Rescale scale;
   - if `c > 0`, generate/obtain `T_c` and subtract it using capacity-safe level/scale alignment at the post-Rescale target;
6. preserve the final Level, Scale, Degree/lazy semantics and planner contract expected by the Fast polynomial evaluator.

The exact implementation may differ from this pseudocode if required by the existing recursion/lazy contract, but the following invariant is mandatory:

> No Chebyshev correction term may be injected at a scale whose required centered representation is not uniquely recoverable from the maintained q0/q1 basis and then passed into Fast centered-CRT Rescale.

Do not solve this by materializing q2+, calling Standard Rescale, using big.Int CRT in production, or silently falling back to full-Q arithmetic.

---

## Preserve planner semantics

Before editing, inspect:

- `circuits/ckks/polynomial/fast.go`
- `circuits/common/polynomial/power_basis.go`
- Fast polynomial planner/simulation tests

The repair must preserve the externally expected polynomial evaluation schedule:

- same final polynomial Level as the existing planner predicts;
- same final target Scale within existing scale precision rules;
- same lazy degree behavior required by baby/giant steps;
- no additional level consumption relative to the planned polynomial depth.

Moving correction after the already-required Rescale must **not** introduce an extra Rescale.

If preserving these contracts requires restructuring the Fast `genPower`/`genPowerInternal` rescale-return bookkeeping, do so narrowly and document it in code comments/tests.

---

## c > 0 must be handled, not only T2

Do not implement a `c == 0` / T2-only patch.

The same architectural issue can occur for:

`T_n = 2*T_a*T_b - T_c`

when `c > 0` because the correction polynomial can also be unsafe at the pre-Rescale product scale.

The production fix must make the general Chebyshev correction capacity-safe for both:

- `c == 0` scalar one;
- `c > 0` generated `T_c`.

Use the existing Fast `subAligned`/scale-alignment machinery only if its use at the **post-Rescale** scale is mathematically safe and preserves the planner contract. Add focused tests for both cases.

---

## Required Secondary regression tests

Add production regression tests in Secondary. These tests should remain committed with the fix.

### 1. Formal-scale T2 regression

Construct a representative Fast zero-secret ciphertext at the formal Mod1 scale/profile sufficient to reproduce the old failure.

Evaluate exact Chebyshev `T2` through the real Fast power-generation path.

Require decoded result vs plaintext:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

The test must fail on `ce79b861...` and pass after the repair.

### 2. Formal-scale c>0 Chebyshev regression

Exercise at least one non-power-of-two generated Chebyshev power whose recurrence has `c > 0`, preferably `T3` or `T5`, at the same high-scale regime.

Require the same `1e-2` component threshold.

### 3. Degree-30 formal Mod1 polynomial regression

Use the actual formal Mod1 polynomial parameters/coefficients or a source-backed construction equivalent to the formal LogN13 Mod1 polynomial, not a synthetic arbitrary degree-30 polynomial.

Run the real Fast polynomial evaluator and require numerical correctness at `1e-2` component threshold.

### 4. Planner metadata regression

Confirm the repaired result preserves expected:

- final Level;
- final Scale;
- degree/lazy requirements;
- no additional rescale depth.

### 5. Existing tests

All existing Fast polynomial, evaluator, Rescale and bootstrap tests must continue to pass.

---

## Update Fast architecture documentation

Update `docs/FAST_CKKS_SPEC.md` with a concise invariant derived from the bug:

- q0/q1 centered reconstruction only determines a unique centered integer while magnitude is `< q0*q1/2`;
- high-scale Standard operation ordering is not automatically valid under the two-limb Fast backend;
- operations that create a high-scale value beyond this range must reduce scale before adding/subtracting correction terms when algebraically valid;
- Chebyshev power generation uses capacity-safe post-Rescale correction ordering.

Do not generalize beyond what this repair proves.

---

## Secondary implementation workflow

Implement on `fast-ckks` starting from clean `ce79b861...`.

Run focused tests first, then:

`go test ./...`

When all Secondary tests pass:

- inspect diff;
- commit the production fix + regression tests + spec documentation in Secondary;
- push `fast-ckks` by normal fast-forward;
- record exact new Secondary commit SHA.

No force push.

---

## Primary formal validation after Secondary fix

After the repaired Secondary commit exists, use that exact clean commit for Primary validation.

Do not change the formal workload/config.

### Gate 1 — T2 micro-regression

Re-run the exact formal diagnostic semantics from EXP-002-C-DIAG-T2.

Require:

- A still passes;
- B still passes;
- exact C T2 now passes;
- D still passes;
- T2 max component <= `1e-2`.

### Gate 2 — generated Chebyshev powers

Re-run the formal power diagnostic from EXP-002-C-DIAG-POLY.

Require every generated formal power used by the degree-30 polynomial to pass the fixed `1e-2` threshold.

Previously failing set was:

`2, 3, 4, 6, 8, 16`

Do not special-case only this list; validate the entire actual generated power set.

### Gate 3 — EvalMod(real)

Re-run the formal internal/whole-polynomial evidence.

Require real Fast Mod1 polynomial output and public `EvalMod(real)` to pass the existing fixed `1e-2` logical threshold against the same source-backed plaintext/Standard oracle.

### Gate 4 — full formal LogN13 Bootstrap correctness

Re-run the authoritative dual-decryption formal correctness gate:

- Standard pinned `5dbffb...`: generated-secret vs input;
- repaired Fast: zero-secret vs input;
- repaired Fast zero-secret vs Standard generated-secret.

All formal comparisons must satisfy:

- `max_abs_real <= 1e-2`
- `max_abs_imag <= 1e-2`

Archive all 4096 decoded slots.

If any gate fails, stop. Do not run LogN16.

---

## Primary result artifacts

Create new fix-validation artifacts rather than overwriting historical diagnostic evidence:

- `results/FIX-001-T2-logN13-fast.json`
- `results/FIX-001-POLY-logN13-fast.json`
- `results/FIX-001-EVALMOD-logN13-fast.json`
- `results/FIX-001-CORRECTNESS-logN13-standard.json`
- `results/FIX-001-CORRECTNESS-logN13-fast.json`
- `results/FIX-001-summary.json`

The summary must record:

- old Fast baseline `ce79b861...`;
- repaired Fast commit;
- exact Secondary changed files;
- T2/power/EvalMod/full-bootstrap metrics;
- Standard reference commit;
- tests run;
- Primary and Secondary clean state;
- whether all four gates pass.

Do not overwrite EXP-002-C diagnostic files.

---

## Completion criteria

Mark FIX-001 COMPLETE only if all are true:

1. production Fast Chebyshev correction ordering is capacity-safe for both `c==0` and `c>0`;
2. no q2+ or full-Q fallback is introduced;
3. formal-scale T2 regression passes;
4. formal c>0 regression passes;
5. degree-30 formal Mod1 polynomial regression passes;
6. Secondary `go test ./...` passes;
7. all formal generated powers pass;
8. formal EvalMod(real) passes;
9. full formal LogN13 Bootstrap passes authoritative correctness comparisons;
10. Secondary fix is committed/pushed normally;
11. Primary evidence is committed/pushed normally;
12. both worktrees end clean.

If any condition fails, leave the task incomplete and report the first failing gate with evidence.

---

## Explicit non-goals

Do not in this task:

- run LogN16;
- rerun performance benchmarks;
- start EXP-003;
- alter Mod1 parameters;
- lower Mod1 scale;
- add active q2/q3 limbs;
- weaken `1e-2` thresholds;
- change Standard Lattigo behavior.

After FIX-001 passes LogN13, the next task will be independent review followed by LogN16 correctness validation, then performance baselines must be refreshed because production Fast code has changed.