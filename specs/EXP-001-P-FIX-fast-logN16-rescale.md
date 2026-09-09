# EXP-001-P-FIX — Fast LogN16 Rescale Compatibility

## Purpose

EXP-001-P is blocked only on the pinned Fast backend when running the full-slot LogN16 / N=65536 profile.

Observed failure during Fast warm-up:

```text
unsupported Fast Rescale moduli:
q0 must be <= 55 bits and q1 <= 39 bits
```

The same profile succeeds on the pinned Standard backend. Its effective first two bootstrapping Q primes are:

```text
q0 bit length = 56
q1 bit length = 39
```

Do not change the experiment profile to avoid this blocker. Fix the Fast backend compatibility boundary, then finish the missing EXP-001-P measurements.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- current blocking baseline commit: `d5429b1c368a6828ef1bb8a229e3f57458ad6491`

Follow both repositories' `AGENTS.md` startup/safety rules before editing.

## Confirmed source cause

At the blocking Fast commit, `schemes/ckks/fast/rescale.go` uses:

```go
const (
    fastRescaleMaxQ0Bits  = 55
    fastRescaleMaxQ1Bits  = 39
    fastRescaleMaxQ01Bits = 94
)
```

The previous optional LogN16 Rescale benchmark explicitly avoided this by changing its local test profile from `55/39` to `54/38` for LogN16. That workaround must not be used for the formal N=65536 experiment.

## Required mathematical safety analysis

Before changing constants, verify and document why the current fixed-width implementation can safely support the actual LogN16 case:

```text
q0 <= 56 bits
q1 <= 39 bits
q0*q1 < 2^95
```

The implementation uses a two-word 128-bit representation (`hi`,`lo`), so the CRT product itself remains below 128 bits.

For centered reconstruction:

```text
|x| < q0*q1/2 < 2^94
```

The existing validator requires the rescale divisor to have bit length at least 32, hence:

```text
d >= 2^31
```

Therefore the rounded quotient magnitude is bounded below 2^63 (up to the ordinary rounding boundary), which remains representable in the existing `uint64` magnitude path.

Also inspect the preconditions of every `bits.Div64` call used by:

- `crtQ01`
- `roundedMagnitude128`

and prove that the widened 56/39 domain does not violate the required `hi < divisor` conditions.

If source inspection disproves any of these assumptions, stop and report the exact arithmetic blocker instead of merely raising constants.

## Secondary implementation scope

If the proof holds, make the smallest correct change in `xuejin-lu/lattigo` `fast-ckks` to support the real LogN16 domain.

Expected direction:

```text
q0 max:        55 -> 56 bits
q1 max:        remain 39 bits
q0*q1 max:     94 -> 95 bits
min divisor:   remain unchanged unless source proof requires otherwise
```

Do not:

- introduce arbitrary-precision `big.Int` arithmetic into the production Fast Rescale loop;
- fall back to the Standard CKKS evaluator;
- reduce q0/q1 in the experiment profile;
- alter Standard behavior;
- broaden unrelated Fast parameter support;
- weaken validation beyond the mathematically proven domain.

## Required Secondary tests

Add focused correctness coverage that exercises an **actual generated 56-bit q0 and 39-bit q1**, not only synthetic constants.

At minimum cover:

1. Range validation accepts the intended 56/39 domain.
2. Fixed-width CRT reconstruction matches a `big.Int` oracle for representative positive/negative values, including near centered-CRT boundaries.
3. Rounded division matches the Standard rule near rounding boundaries.
4. Fast `Rescale` / `RescaleTo` matches Standard on maintained q0/q1 for the widened domain.
5. LogN13 existing Rescale tests remain unchanged and passing.
6. LogN16 ScaleDown using the real profile succeeds without substituting 54/38.
7. Complete Fast Bootstrap with LogN16, `log_slots=-1`, N=65536 succeeds at least once before formal measurement.

Run at least:

```text
go test ./schemes/ckks/fast
go test ./circuits/ckks/bootstrapping
go test ./circuits/ckks/dft
go test ./circuits/ckks/mod1
go test ./circuits/ckks/polynomial
```

If a narrower package list is justified by repository structure, still run the existing Fast bootstrap acceptance tests required by `AGENTS.md` / `docs/FAST_CKKS_SPEC.md`.

## Secondary commit

Commit and push the Fast fix to `origin/fast-ckks`.

Report the new exact Fast commit SHA. This new SHA becomes the Fast backend identity for the completed LogN16 baseline; do not keep pretending the old blocking SHA produced a LogN16 result.

## Resume EXP-001-P after the backend fix

After the Secondary fix is pushed, return to the Primary experiment harness.

Use the experiment source/config state from Primary commit:

```text
1e91682e14fd9939ffb2d72c6770ed85a3870bbf
```

for formal measurement so the new Fast measurements can be compared against already-archived Standard results produced from the same Primary source/config commit.

The existing Standard raw results may be reused if all identity fields remain valid; do not rerun Standard merely to create a new file unless necessary.

### Required reruns

Because the Fast backend commit changed, rerun Fast for:

1. LogN13 clean profile
2. LogN16 clean full-slot profile

This gives one consistent new Fast backend commit across the final LogN13/LogN16 matrix.

Write measurement output outside the repository first (`/tmp/...`) so Primary records `dirty=false`, then archive only after validating metadata.

## Preserve previous evidence

Do not overwrite or delete:

```text
results/EXP-001-standard.json
results/EXP-001-fast.json
results/EXP-001-summary.json
results/EXP-001P-logN13-standard.json
results/EXP-001P-logN13-fast.json
results/EXP-001P-logN13-summary.json
results/EXP-001P-logN16-standard.json
```

The existing LogN13 Fast result from the old backend remains useful historical evidence.

Archive the new Fast reruns with distinct names, for example:

```text
results/EXP-001P2-logN13-fast.json
results/EXP-001P2-logN16-fast.json
```

Then create final summaries that explicitly identify the new Fast commit, for example:

```text
results/EXP-001P2-logN13-summary.json
results/EXP-001P2-logN16-summary.json
results/EXP-001P2-matrix-summary.json
```

Do not fabricate a LogN16 speedup until the real Fast run succeeds.

## Final identity checks

For each final Standard/Fast pair used in the new summaries:

- Primary experiment commit must be identical.
- Primary dirty=false.
- Secondary dirty=false.
- Config identical within each pair.
- Effective parameters identical within each pair.
- Environment metadata matched.
- Warmup/repetition counts matched.
- Output level expected.
- Standard backend commit remains `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast backend commit equals the new pushed Fast fix SHA.

## Completion

This fix task passes only when:

1. the 56/39 arithmetic domain is proven safe or a concrete source-backed reason is reported why it is not;
2. the Secondary implementation supports the real LogN16 profile without Standard fallback or profile reduction;
3. relevant Fast tests pass;
4. the new Fast commit is pushed;
5. clean LogN13 and LogN16 Fast formal runs are archived under distinct filenames;
6. the final LogN16 Standard-vs-Fast speedup is computed from real matching results;
7. a final matrix summary is archived;
8. previous experiment files remain untouched;
9. Secondary is restored cleanly to `fast-ckks` at the new pushed commit.

Do not proceed automatically to EXP-002 stage profiling after completion.