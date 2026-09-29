# QPREFIX-PERF-OPT-004 — Exact Barrett-Horner Fixed-Width Reduction

## Status

Executable bounded production optimization task.

## Task class

`P — Performance Repair`

## Accepted parent

- `results/QPREFIX-PERF-OPT-003-summary.md`
- Primary commit: `16918ec5851f9c901ac4b8e266357bf3db757f47`
- OPT-003 decision: `SOURCE_INTT_BATCHING_CORRECT_BUT_NO_WIN`
- accepted production implementation remains Secondary `6930cf6cb3c71ce139a1eb42eede7be335b7174c`
- current control-plane HEAD: `2018ae7c74151d7b4de8e0c0bf393d96e18b8181`

Fresh accepted DIAG-007 evidence:
- `math/bits.Div64` remains the largest single flat symbol at 20.55%;
- fixed-width phase: 36.55%;
- signed residue staging: 21.56%;
- source+restore transforms: 41.89%;
- full rows4 Rescale about 1.130242 ms.

OPT-003 showed simple INTT loop reordering has no performance value, so do not revisit that approach.

## Opportunity

Lattigo already provides exact Barrett primitives in `ring/modular_reduction.go`:

- `ring.BRed`;
- `ring.BRedAdd`;
- `ring.CRed`;
- per-SubRing `BRedConstant`.

These avoid hardware/software integer division and are already used throughout the ring package.

Current Fast helpers `mod128By64` / `mod192By64` still depend on `bits.Div64`.

For a 192-bit value:

[
x = hcdot 2^{128}+mcdot2^{64}+l
]

and a modulus (q), define the exact precomputed scalar

[
R_q = 2^{64}mod q.
]

Then compute exactly:

[
r_0=hmod q
]
[
r_1=(r_0R_q+m)mod q
]
[
r_2=(r_1R_q+l)mod q.
]

Each 64-bit reduction/multiplication can use Lattigo's existing exact Barrett primitives.

## Goal

Determine whether an exact Barrett-Horner implementation can replace hot `Div64`-based modular reductions in the Q-prefix Rescale path and yield a measurable production win.

No approximate reciprocal arithmetic is allowed.

## 1. Hard constraints

Do not change:

- CRT semantics;
- centered representative;
- rounding;
- capacity checks;
- Q-prefix width/policy;
- transactionality;
- active rows;
- NTT/Montgomery representation;
- P93/generated-power schedule;
- public APIs;
- F/full-RNS policy.

Do not introduce:

- floating point;
- approximate reciprocal;
- unsafe;
- assembly;
- goroutines;
- architecture-specific code.

Use existing Lattigo exact Barrett primitives only.

## 2. Test-only feasibility candidate first

Before touching production behavior, implement same-package test/benchmark-only helpers:

- `mod128BarrettPrepared`;
- `mod192BarrettPrepared`.

Each candidate receives:
- modulus q;
- q's existing `BRedConstant`;
- precomputed `R64 = 2^64 mod q`.

The hot candidate must contain no `bits.Div64`.

Computing `R64` during evaluator/scratch setup may use a one-time exact `bits.Div64(1,0,q)` or equivalent; setup cost is not in the coefficient loop.

## 3. Exactness proof/tests

Compare candidate reductions against:
- current `mod128By64` / `mod192By64`;
- `math/big.Int.Mod`.

Cover:
- all active P93 q moduli;
- q0 around 55/56 bits;
- q1+ around current smaller limbs;
- values 0 and all-one limbs;
- values near modulus multiples;
- maximum uint128 / uint192 patterns;
- fixed-seed randomized corpus.

No tolerance: exact equality.

Also prove all intermediate Barrett inputs satisfy the documented primitive domains.

## 4. Feasibility benchmarks

Diagnostics disabled, fixtures outside timing, >=7 samples.

Benchmark:

- current `mod128By64` vs Barrett candidate;
- current `mod192By64` vs Barrett candidate;
- `signedResidue192` current vs Barrett-backed candidate for q0,q1,q2,q3;
- rows4 prepared CRT current vs Barrett-backed candidate.

Report medians and allocation counts.

### Feasibility gate

Do not retain a production change unless:

- `mod192` candidate improves >= 15%, and
- either signed-residue staging or rows4 prepared CRT improves >= 8%.

If these fail, remove temporary candidate code and return no-win.

## 5. Production prepared constants

If feasibility passes, extend `fastRescaleScratch` with only the scalar constants needed for the accepted helper:

- existing per-row `BRedConstant` values may be copied/referenced;
- per-row `R64 mod q`.

Initialize once with evaluator scratch.

No per-call or per-coefficient allocation.

## 6. Production scope

Replace only the Rescale hot-path modular reductions where the candidate benchmark demonstrates a win.

Preferred targets:

- residue emission in `signedResidue192`;
- rows3/rows4 prepared CRT internal reductions.

Do not automatically replace generic helpers used elsewhere unless independently validated.

It is acceptable to add Rescale-specific prepared helpers and leave generic division-based helpers untouched.

## 7. Correctness regression

Run exact tests covering:

- rows1/2/3/4 reconstruction;
- centered CRT oracle;
- sequential RescaleTo;
- in-place/out-of-place;
- Montgomery false/true;
- dormant residue independence;
- capacity failures and failure-before-mutation;
- higher-degree staged ciphertexts.

Run:
- `go test ./schemes/ckks/fast -count=1`;
- Secondary `go test ./...`.

## 8. Authoritative performance

If production change is retained, collect same-session before/after:

### A. residue staging phase
>=7 samples, diagnostics off.

### B. fixed-width phase
>=7 samples, diagnostics off.

### C. full rows4 Rescale
`BenchmarkFastRescaleQPrefixRows4LogN13P93`
>=7 samples, `-benchmem`.

### D. q0=55 P93 E2E
`BenchmarkFastDiagP93Q55Count1`
>=7 samples, `-benchmem`.

### E. low-overhead power trace
`./scripts/fastdiag trace --profile p93-q55 --trace power --warmup 1 --repetitions 7`.

Do not use deep Rescale tracing for acceptance.

## 9. Performance gates

Return `BARRETT_FIXED_WIDTH_REDUCTION_READY` only if:

- residue staging or fixed-width phase improves >= 10%;
- full rows4 Rescale improves >= 4%;
- P93 E2E improves >= 2%;
- allocation count does not materially regress;
- all exactness/transactionality tests pass.

Return `BARRETT_FIXED_WIDTH_REDUCTION_CORRECT_BUT_NO_WIN` if correctness passes but feasibility or production performance floors fail.

Return `BARRETT_FIXED_WIDTH_REDUCTION_BLOCKED` if exactness or semantics cannot be preserved.

Do not keep production churn on a no-win result.

## 10. Fresh profile after success

If the production candidate passes all gates, take one fresh focused rows4 Rescale CPU profile (>=5s) and report:

- new `math/bits.Div64` flat share;
- transform flat/cumulative evidence;
- new largest flat symbol.

This is descriptive only; do not start another optimization in this task.

## 11. Required artifact

Write:

`results/QPREFIX-PERF-OPT-004-summary.md`

Include:
- exact candidate math;
- feasibility raw samples;
- exactness/oracle tests;
- production commit if any;
- phase/full/E2E before/after;
- allocations;
- low-overhead power trace;
- fresh profile if successful;
- remaining hotspot;
- final classification.

## 12. Completion

If successful, commit/push Secondary normally.
If no-win, leave Secondary production unchanged and clean.

Primary commits/pushes the result artifact.

Return exactly one:
- `BARRETT_FIXED_WIDTH_REDUCTION_READY`
- `BARRETT_FIXED_WIDTH_REDUCTION_CORRECT_BUT_NO_WIN`
- `BARRETT_FIXED_WIDTH_REDUCTION_BLOCKED`

Then report:

`READY_FOR_WEB_REVIEW`.
