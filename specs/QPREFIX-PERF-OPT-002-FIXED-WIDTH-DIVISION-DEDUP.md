# QPREFIX-PERF-OPT-002 — Fixed-Width Division Deduplication

## Status

Executable production optimization task.

## Task class

`P — Performance Repair`

## Accepted parent

Kernel attribution:
- `results/QPREFIX-PERF-DIAG-006-summary.md`
- Primary commit: `f4de1884e07fc89e51fa65e7999df94a60d72704`
- classification: `RESCALE_MIXED_KERNELS`
- top isolated phase: fixed-width reconstruct / round / capacity
- top phase share: `43.51%`
- rows4/rows2 full Rescale ratio: `2.199x`
- phase closure: `1.128`

Current production arithmetic baseline:
`d50ff4db757d4a2b9922937a4e7f316fd3f286b9`

Accepted pprof evidence:
- `math/bits.Div64`: 27.40% flat;
- INTT/NTT are also material;
- `mod192By64`, `reconstructQPrefix`, and `signedResidue192` are on the hot path.

## Problem

The fixed-width helpers currently perform avoidable division work.

### A. Redundant leading Div64 in modular reduction

Current forms are equivalent to:

```go
func mod128By64(lo, hi, modulus uint64) uint64 {
    _, remainder := bits.Div64(0, hi%modulus, modulus)
    _, remainder = bits.Div64(remainder, lo, modulus)
    return remainder
}

func mod192By64(value uint192, modulus uint64) uint64 {
    _, remainder := bits.Div64(0, value.hi%modulus, modulus)
    _, remainder = bits.Div64(remainder, value.mid, modulus)
    _, remainder = bits.Div64(remainder, value.lo, modulus)
    return remainder
}
```

After `hi % modulus` or `value.hi % modulus`, the operand is already strictly less than `modulus`. Therefore:

```go
bits.Div64(0, r, modulus)
```

must return quotient `0`, remainder `r`.

That first `Div64` is mathematically redundant.

### B. Duplicate quotient/remainder division expression

`roundedMagnitude192` currently derives the highest quotient limb and remainder from the same pair using separate `/` and `%` expressions.

The implementation may instead use one exact quotient+remainder operation when that is measurably beneficial and assembly/benchmark evidence supports it.

### C. Per-coefficient invariant reconstruction products

Rows3/rows4 Garner reconstruction currently recomputes modulus products such as q01/q012 inside coefficient reconstruction even though evaluator scratch already owns invariant prefix products.

Where source inspection confirms exact equivalence, reuse precomputed scratch constants rather than rebuilding invariant products per coefficient.

This part is secondary to division deduplication: do not broaden the task into a general CRT rewrite.

## Goal

Reduce the number and cost of fixed-width divisions in production Q-prefix Rescale while preserving bit-exact arithmetic and every existing semantic boundary.

The primary target is the current rows4 P93 Rescale hot path.

## 1. Hard semantic constraints

Do not change:

- `QPrefixWidth(Level)`;
- authoritative row policy;
- centered CRT representative;
- Garner reconstruction result;
- round-to-nearest rule currently implemented by `roundedMagnitude192`;
- tie behavior;
- sequential multi-rescale semantics;
- per-step capacity checks;
- signed residue semantics;
- transactional staging;
- failure-before-mutation;
- NTT/Montgomery behavior;
- metadata / Scale / Level behavior;
- P93 or generated-power schedule;
- public Fast APIs;
- F/full-RNS policy.

No approximate reciprocal arithmetic is authorized in this task.

No floating-point reciprocal.

No probabilistic or input-range shortcut.

## 2. Required helper repair

At minimum, replace the provably redundant leading `bits.Div64` in:

- `mod128By64`;
- `mod192By64`.

The replacement must be exactly equivalent for every valid nonzero modulus and all uint64/uint192 inputs accepted by the current helpers.

Add direct exhaustive/randomized tests against the old formula or a BigInt oracle.

## 3. roundedMagnitude192 quotient/remainder review

Inspect generated assembly and benchmark both versions of the high-limb quotient/remainder step:

Current semantic expression:

```go
q2 := value.hi / divisor
remainder := value.hi % divisor
```

Candidate exact expression:

```go
q2, remainder := bits.Div64(0, value.hi, divisor)
```

Use the candidate only if:

- it is bit-exact;
- it does not introduce a panic domain absent from the current valid call set;
- focused benchmark or assembly evidence indicates no regression.

If Go already fuses the original `/` + `%` optimally on arm64, retain the current source and document that finding.

Do not force a source change merely for stylistic symmetry.

## 4. Prepared rows3/rows4 reconstruction constants

Audit:

- `crtQ012`;
- `crtQ0123`;
- `reconstructQPrefix`;
- `fastRescaleScratch`.

The scratch already holds prefix products/inverses.

If q01/q012 products are rebuilt inside the per-coefficient loop, add a bounded prepared path that reuses the already-validated scratch products.

Preferred direction:

```text
rescale call setup
  prevalidated q/inverse/prefix-product constants
           |
           v
coefficient loop
  reconstructPrepared(rows, residues, scratch)
```

Do not change the generic mathematical helper semantics merely to optimize one call site. It is acceptable to add a Rescale-specific prepared helper.

Avoid extra allocations.

## 5. No broad reciprocal rewrite

Although `bits.Div64` is hot, this task does NOT authorize:

- Barrett reciprocal redesign;
- Montgomery reciprocal division;
- custom approximate reciprocal;
- assembly;
- unsafe;
- architecture-specific intrinsics.

Those require a separate proof/benchmark task if the simple exact deduplication leaves division dominant.

## 6. Correctness tests

Add direct tests for helper equivalence covering:

### mod128By64 / mod192By64

- boundary moduli used by current Q-prefix parameters;
- modulus near 2^39, 2^40, 2^55/56, 2^60/61 where valid;
- value 0;
- all-one limbs;
- values near modulus multiples;
- randomized fixed-seed corpus.

Compare against `big.Int.Mod`.

### roundedMagnitude192

Compare against BigInt exact round-to-nearest implementation across:

- zero;
- half-divisor boundaries;
- just below/above half;
- multi-limb magnitudes;
- current P93 divisors;
- fixed-seed randomized corpus.

### CRT reconstruction

For rows 1/2/3/4, compare prepared production reconstruction with:

- existing generic helper;
- BigInt CRT oracle where available.

No tolerance: exact equality.

## 7. Rescale regression suite

Run and pass existing tests covering:

- BigInt centered-CRT Rescale oracle;
- every prefix width;
- in-place / out-of-place;
- Montgomery false/true;
- sequential RescaleTo;
- dormant residue independence;
- transactionality and capacity failures;
- higher-degree ciphertext staging.

Run `go test ./schemes/ckks/fast -count=1`.

Run Secondary `go test ./...` if no unrelated repository blocker appears.

## 8. Focused helper benchmarks

Add or use reusable same-package benchmarks for:

- `mod128By64`;
- `mod192By64`;
- rows4 prepared reconstruction;
- `roundedMagnitude192`;
- fixed-width phase from DIAG-006.

Fixtures must be prepared outside timing.

Report before/after medians and allocation counts.

These benchmarks may remain in Secondary if they are clearly reusable and test-only.

## 9. Authoritative Rescale performance

Before production edits, obtain same-session baseline from the current production source.

After implementation rerun identically.

### A. full rows4 Rescale

`BenchmarkFastRescaleQPrefixRows4LogN13P93`

Requirements:
- diagnostics off;
- >=7 samples;
- same benchtime;
- `-benchmem`.

### B. fixed-width phase

Recreate/use the accepted DIAG-006 phase fixture.

Requirements:
- rows4;
- diagnostics off;
- >=7 samples;
- no setup inside timed loop.

### C. rows2 control

Rerun the matched rows2 full Rescale benchmark as a regression/control measurement.

Label:
`WIDTH_COUNTERFACTUAL_ONLY`.

## 10. P93 E2E performance

Rerun diagnostics-off:

`BenchmarkFastDiagP93Q55Count1`

with:
- one warmup;
- >=7 measured samples;
- `-benchmem`.

Use the same-session pre-edit baseline when possible.

Then run low-overhead:

```sh
./scripts/fastdiag trace --profile p93-q55 --trace power --warmup 1 --repetitions 7
```

to verify that generated-power Rescale time moves in the expected direction without deep Rescale tracing.

## 11. Performance gates

Use classification:

### `FIXED_WIDTH_DIVISION_DEDUP_READY`

only if all correctness gates pass and:

- fixed-width rows4 phase improves >= **10%**;
- full rows4 Rescale improves >= **5%**;
- diagnostics-off P93 Count-1 E2E improves >= **2%**;
- no material allocation regression.

### `FIXED_WIDTH_DIVISION_DEDUP_CORRECT_BUT_NO_WIN`

if correctness passes but one or more performance floors fail.

### `FIXED_WIDTH_DIVISION_DEDUP_BLOCKED`

if correctness/transactionality cannot be preserved.

The floors are deliberately modest because this is a bounded exact-arithmetic cleanup, not a new division algorithm.

Do not hide a no-win result.

## 12. Allocation gate

Production steady-state must not add:

- per-coefficient allocations;
- per-Rescale allocations proportional to N;
- new persistent allocation count in the focused Rescale benchmark.

Any new prepared constants belong in evaluator scratch or stack/local scalar state.

## 13. Source/assembly evidence

For the changed helpers, record compiler/assembly evidence sufficient to answer:

- was a redundant division actually removed?
- did `roundedMagnitude192` high-limb quotient/remainder become one divide or was the compiler already doing so?

Do not require architecture-specific source code.

## 14. Required result artifact

Write:

`results/QPREFIX-PERF-OPT-002-summary.md`

Include:

- exact Secondary implementation commit;
- helper/source changes;
- mathematical equivalence argument;
- test results;
- helper benchmark before/after;
- fixed-width phase before/after;
- full rows4 Rescale raw samples/medians;
- rows2 control;
- P93 E2E raw samples/medians;
- low-overhead power trace before/after;
- allocations;
- assembly/compiler observation;
- remaining top hotspot.

## 15. Completion

Commit/push Secondary production implementation normally if successful/correct.

Primary commits the result artifact.

Return exactly one:

- `FIXED_WIDTH_DIVISION_DEDUP_READY`
- `FIXED_WIDTH_DIVISION_DEDUP_CORRECT_BUT_NO_WIN`
- `FIXED_WIDTH_DIVISION_DEDUP_BLOCKED`

Then report:

`READY_FOR_WEB_REVIEW`.
