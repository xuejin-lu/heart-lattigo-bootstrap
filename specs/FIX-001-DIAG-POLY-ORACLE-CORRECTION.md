# FIX-001-DIAG-POLY-ORACLE-CORRECTION

## Purpose

The completed `FIX-001-DIAG-POLY-formal-replay` result cannot currently support its conclusion that repaired formal `T2` still fails.

Independent review found a bug in Primary diagnostic code `fix001_diag_poly_runner.go`:

```go
vc := make([]complex128, len(x))
if c != 0 {
    vc = fix001ExpectedPower(c, x, memo)
}
out[i] = 2*va[i]*vb[i] - vc[i]
```

For `T2`, `a=1`, `b=1`, `c=0`. The current oracle therefore uses `vc=0` and evaluates:

`2*x^2`

instead of the Chebyshev identity:

`T2(x) = 2*x^2 - T0(x) = 2*x^2 - 1`.

The observed repaired `T2` apparent error of about `1.0004882811772378` is therefore consistent with an oracle missing the constant one.

This task must correct **only the Primary diagnostic oracle**, prove it with focused tests, and replay the same formal repaired-Fast power diagnosis.

Do not modify Lattigo.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected remote base when this spec is authored:

`57a55ffaf33b1aab0a55cfc35cef90245e61298`

Repaired Fast backend:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Standard reference, if needed for existing Q3 comparison:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Both worktrees must be clean before starting. Follow `AGENTS.md`.

---

## Required code correction

Modify only the Primary diagnostic implementation required to make the Chebyshev plaintext recurrence mathematically correct.

Preferred invariant:

- `T0(x) = 1`
- `T1(x) = x`
- `Tn(x) = 2*Ta(x)*Tb(x) - T|a-b|(x)` using the same `SplitDegree(n)` dependency structure as production.

Implement an explicit `n == 0` base case returning an all-ones vector, or an equivalently clear implementation.

Do not special-case only `T2`.

Do not modify the production Fast polynomial implementation.

---

## Mandatory focused tests

Add Primary tests for the corrected diagnostic oracle.

At minimum:

1. `T0` on arbitrary complex inputs returns exactly `1+0i` for every slot.
2. `T1` returns the input exactly.
3. `T2` matches `2*x*x - 1` for representative real and complex values.
4. `T3` matches `4*x^3 - 3*x` within floating-point tolerance.
5. Recurrence results for several n values agree with a source-backed Lattigo/bignum Chebyshev polynomial evaluation where practical.

These are diagnostic tests; they must not depend on Fast ciphertext arithmetic.

---

## Replay unchanged formal experiment

After the oracle fix, rerun the exact same `FIX-001-DIAG-POLY` formal replay:

- LogN13 only
- exact same config
- exact same deterministic input
- exact same formal E2 real branch
- exact same E1/E2 preprocessing
- repaired Fast exactly `87be78ff...`
- same generated power snapshots
- same q0/q1 projection/decode procedure
- same threshold `1e-2`

Require E2 to remain bit/numerically identical to the historical E2 vector before interpreting the replay.

---

## Required classification after corrected replay

### Case A — corrected T2 still fails

If corrected `T2 = 2*x^2-1` still exceeds `1e-2`, classify:

`FIRST_SUPPORTED_CAUSE = repaired_power_generation_t2_still_fails`

Then report the corrected actual-vs-expected metrics and stop further repair work.

### Case B — T2 passes but another dependency-supported generated power fails

Classify:

`FIRST_SUPPORTED_CAUSE = repaired_power_generation`

Record the first dependency-supported failing power.

### Case C — all generated powers pass but whole formal degree-30 polynomial still fails

Classify:

`FIRST_SUPPORTED_CAUSE = paterson_stockmeyer_accumulation_or_scale_alignment`

This is only allowed if:

- every actual generated power passes corrected oracle;
- whole polynomial still reproduces its formal failure;
- public Q3(real) still reproduces the known stage failure.

### Case D — all powers and whole polynomial pass but public Q3(real) fails

Classify:

`FIRST_SUPPORTED_CAUSE = post_polynomial_evalmod_logic`

### Case E — powers, whole polynomial, and public Q3 all pass

Classify:

`FIRST_SUPPORTED_CAUSE = prior_diagnostic_oracle_artifact`

Do not automatically mark parent FIX-001 complete; rerun the authoritative full LogN13 correctness gate first.

---

## Historical evidence handling

Do not delete or overwrite:

- `results/FIX-001-DIAG-POLY-logN13-fast.json`
- `results/FIX-001-DIAG-POLY-logN13-summary.json`

They are historical evidence of a diagnostic run with an invalid T0 oracle.

Create new artifacts:

- `results/FIX-001-DIAG-POLY-CORRECTED-logN13-fast.json`
- `results/FIX-001-DIAG-POLY-CORRECTED-logN13-summary.json`

The corrected summary must explicitly state:

- previous replay's power classification is superseded because its plaintext Chebyshev oracle treated `T0` as zero;
- production Lattigo evidence was not changed by this correction;
- the independent `EXP-002-C-DIAG-T2` capacity experiment remains separate evidence and is not retroactively rewritten.

---

## Validation

Before completion:

- Primary focused oracle tests pass;
- Primary `go test ./...` passes;
- repaired Secondary `go test ./...` passes without source modification;
- Secondary remains exact `87be78ff...` and clean;
- formal replay completes;
- corrected artifacts are committed/pushed normally;
- both worktrees end clean;
- no LogN16;
- no benchmark;
- no EXP-003.

---

## Non-goals

Do not:

- modify/revert FIX-001 production code;
- change Fast scalar Add/Rescale/Polynomial code;
- implement a PS fix yet;
- change Mod1 parameters;
- change thresholds;
- run LogN16;
- start EXP-003.

The deliverable is a corrected, trustworthy formal power classification.