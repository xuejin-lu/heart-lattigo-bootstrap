# QPREFIX-IMPL-009 — Performance and Release Gate

## Status

Executable Codex validation/release task.

## Task class

`V — Validation / Release Gate`

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- orchestration/spec only.

Secondary:
- `xuejin-lu/lattigo@fast-qprefix`
- production candidate.

Accepted production candidate:
- QPREFIX-IMPL-008: `f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

Historical pre-F comparison point:
- `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- message: `fast-ckks: validate LogN13 P93 Q012 bootstrap candidate`

Authoritative architecture:
- Secondary `docs/FAST_QPREFIX_SPEC.md`
- Primary `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

## Purpose

This is the final Q-prefix v2 release gate.

Do not redesign or optimize the implementation.

The task must answer four questions:

1. **Semantic release readiness** — does the current Fast Q-prefix v2 candidate still satisfy public correctness?
2. **Structural release readiness** — does the production path truly use at most four authoritative logical-Q rows internally and public q0/q01 externally, with no dormant/full-RNS fallback?
3. **Capacity release readiness** — do the accepted C2S and EvalMod/PS/DoubleAngle strict-capacity checkpoints remain valid on the final integrated branch?
4. **Performance result** — under matched conditions, how does the final candidate compare with:
   - the pre-F Q012 baseline;
   - current Standard Lattigo;
   - the current production P93 q0=56 profile?

No production-source changes are authorized unless a release-blocking bug is discovered and reported first.

---

# 1. Candidate freeze

The production candidate is exactly:

`f9c7f21e65915bd3eafcd5b12590b570c22a7d6f`

plus only the current task-pointer commit.

Do not:
- change arithmetic;
- change schedules;
- change parameters;
- optimize;
- relax gates;
- introduce F;
- introduce full-RNS fallback;
- modify `fast-ckks`.

Benchmark-only temporary files are allowed as described below, but must not alter the final candidate.

---

# 2. Baseline isolation

Create a separate detached temporary worktree for:

`40532b4dce5c7eeae2db5b0b6f21be64801ce923`

Requirements:
- do not reset/stash/clean the current `fast-qprefix` worktree;
- do not change any branch pointer;
- do not commit anything to the baseline;
- use separate GOCACHE paths if useful;
- record:
  - commit SHA;
  - Go version;
  - GOOS/GOARCH;
  - CPU;
  - benchmark command;
  - `-count` value.

Remove the temporary worktree after measurement, or clearly report if intentionally retained for review.

---

# 3. Parameter comparability

The historical pre-F P93 helper used q0=55, while the current accepted production P93 profile uses q0=56.

Therefore do **not** present a raw baseline-vs-current q0=56 timing ratio as if only code changed.

Report three distinct comparisons:

## A. Same-code-benchmark comparison

Use benchmark entry points that exist unchanged in both commits, on their own checked-in parameters, to detect broad implementation regressions.

At minimum:
- `BenchmarkFastBootstrapEndToEnd/LogN=13/Fast`;
- `BenchmarkFastEvalModLogN13`;
- `BenchmarkFastPackingLogN13Count2` or the closest identical packing benchmark.

These are attribution/regression metrics, not the sole P93 release number.

## B. Matched q0=55 P93 comparison

Run the **same temporary public-API benchmark harness** on:
- baseline `40532b4...`;
- current candidate;

with the historical q0=55 LogN13/P93 profile.

This isolates the Q-prefix-v2 code migration cost from the q0=55 -> q0=56 parameter change.

## C. Current production q0=56 P93

Run the final current production profile:
- LogN=13;
- LogSlots=12;
- q0=56 accepted chain;
- K=16;
- Mod1Degree=30;
- DoubleAngle=3;
- accepted C2S compression/restore plan.

This is the release number.

---

# 4. Temporary matched P93 benchmark harness

If the exact same q0=55 P93 benchmark does not already exist in both commits, create one temporary `_test.go` benchmark file in both worktrees with **byte-identical benchmark logic** except for no source-API changes.

Use only public/stable package APIs available in both commits.

The harness must:
- build the same q0=55 P93 parameters;
- construct the same encoded input values;
- initialize Fast evaluator outside the timed loop;
- exclude key generation/setup from Fast timing;
- call the same public `Bootstrap` or `BootstrapMany` API in each timed iteration;
- clone input per iteration if the API contract requires ownership safety;
- report allocations.

Do not commit this temporary harness.

Before final report:
- remove it from both worktrees;
- verify current production worktree is clean except the already-authorized task pointer state.

---

# 5. Benchmark discipline

For every release benchmark:
- warm up once;
- use `go test -run '^$' -bench ... -benchmem`;
- use at least `-count=5` for short/medium benchmarks;
- use at least `-count=3` for expensive full P93 Bootstrap benchmarks;
- report median ns/op, B/op, allocs/op;
- report raw per-run numbers or retain them in the completion response;
- run baseline/current comparisons on the same machine and power state as closely as practical;
- do not interleave unrelated heavy workloads.

If variance is obviously unstable (>15% max/min spread after warmup), rerun once and report both sets.

Do not cherry-pick the fastest run.

---

# 6. Current production P93 measurements

On the final current candidate, measure at minimum:

### End to end
- P93 `Bootstrap` count 1;
- P93 `BootstrapMany` count 3.

Use the current q0=56 profile.

### Attribution
- EvalMod;
- C2S representative LinearTransform/group or existing DFT benchmark;
- ModUp;
- packing/unpacking;
- N1<->N2 only where the accepted profile exercises it or as a focused structural benchmark.

Use existing benchmark entry points where possible.

Do not add a new optimization-specific microbenchmark solely to make performance look better.

---

# 7. Standard P93 comparison

On the current q0=56 P93 profile, compare Fast against Standard using the same:
- LogN;
- LogSlots;
- parameter/config literals;
- input message;
- public operation semantics.

Setup/key generation is excluded from timed loops for both.

Record:
- Fast median latency;
- Standard median latency;
- Fast/Standard ratio or Standard/Fast speedup;
- B/op;
- allocs/op.

If Standard `BootstrapMany` does not expose an exactly equivalent batch API, compare count-1 end-to-end and clearly state the limitation rather than inventing a batch ratio.

A Fast release candidate that is slower than Standard on the matched count-1 production P93 workload is a performance-gate failure.

No stronger speedup threshold is invented here unless already documented elsewhere.

---

# 8. Pre-F baseline comparison

For matched q0=55 P93:
- record baseline latency/allocations;
- record current-candidate latency/allocations;
- compute current/baseline ratios.

Interpretation:
- this is the cost of moving from the accepted pre-F Q012 implementation to fixed-cap Q-prefix v2;
- it is not the same as the q0=56 production release number.

Hard escalation:
- any >10x latency or allocation regression relative to matched baseline is a performance-gate failure.

Smaller regressions must be reported accurately; do not optimize in this milestone.

---

# 9. Full semantic regression gate

Run on the final candidate:

- `go test ./schemes/ckks/fast ./circuits/ckks/polynomial ./circuits/ckks/mod1 ./circuits/ckks/dft ./circuits/ckks/bootstrapping -count=1`
- `go test ./... -count=1`
- `git diff --check`
- gofmt check.

Also explicitly rerun:
- public generated-secret decoded comparator;
- zero-secret ordinary decryptor/decode control;
- full-slot real/imag Bootstrap;
- sparse/repacked Bootstrap;
- `BootstrapMany` odd-count coverage;
- N1=N2 and N2=2*N1 public boundary tests.

Any failure is release-blocking.

---

# 10. Capacity regression gate

Rerun the accepted final-branch evidence for:

## C2S
All four C2S group checkpoints:
- raw LinearTransform;
- post-Rescale;
- restore where present.

Require strict:

[
2B<S_Q(Level).
]

## EvalMod / PS / DoubleAngle
Rerun the QPREFIX-IMPL-007 capacity observer test covering:
- EvalMod entry;
- generated powers;
- PS baby/giant;
- polynomial final Rescale;
- each DoubleAngle round;
- EvalMod output.

Require every authoritative component to satisfy strict capacity.

Do not rely only on old pasted numbers; execute the current tests.

Any concrete capacity failure is release-blocking.

---

# 11. Structural row-cap gate

Instrument or use existing tests to prove the production path never requires more than:

[
MaxQPrefixWidth=4.
]

Required stage map:
- public input;
- post-pack;
- ScaleDown;
- ModUp;
- C2S group outputs;
- EvalMod entry/output;
- S2C group outputs;
- pre-unpack;
- public output.

For each stage record:
- logical Level;
- authoritative rows;
- physical row count/backing;
- NTT;
- Montgomery;
- Degree;
- Scale.

Rules:
- internal Level >=3 authority <=4;
- q4+ never read;
- public input/output Level <=1 authority is q0/q01;
- no full logical-Q materialization appears in production arithmetic/packing paths.

If physical allocation temporarily has more rows due to a Standard/public type outside Fast-owned compact storage, identify it and prove dormant rows are not consumed. Do not hide it.

---

# 12. Static fallback audit

Inspect final production Fast source for forbidden fallback patterns.

At minimum check the call graph rooted at:
- `FastEvaluator.Bootstrap`;
- `FastEvaluator.BootstrapMany`;
- `bootstrapCore`.

Confirm production execution does not instantiate/use:
- Standard CKKS evaluator as a fallback;
- Standard full-RNS Rescale/LinearTransform/Mod1 path;
- `rlwe.SwitchCiphertextRingDegree` as production fallback;
- private F storage/bridge;
- q4+ dormant row arithmetic.

Standard APIs may appear in tests/oracles only.

Any hidden production fallback is release-blocking.

---

# 13. Public representation gate

Finalized outputs must be ordinary public Lattigo ciphertexts:

- ring degree = residual N;
- Level = residual MaxLevel;
- Degree = 1;
- NTT = true;
- Montgomery = false;
- Scale = residual default;
- LogDimensions/batching metadata correct;
- authoritative q0/q01 storage valid;
- ordinary decryptor + encoder can consume them without a Fast-specific conversion step.

Inputs must remain unchanged.

---

# 14. Release verdict categories

Return exactly one overall status:

## `QPREFIX_V2_RELEASE_PASS`

All semantic, capacity, structural, fallback, and public-representation hard gates pass, and Fast is not slower than Standard on the matched production P93 count-1 benchmark. No >10x matched-baseline regression.

## `QPREFIX_V2_PERFORMANCE_REVIEW`

All correctness/structure/capacity gates pass, Fast remains faster than Standard, but the matched pre-F baseline regression is material enough to warrant an explicit human decision even though it is below the 10x hard failure threshold.

Do not invent a hidden numeric threshold for "material"; report the ratios and use this status when the regression is clearly substantial and not measurement noise.

## `QPREFIX_V2_RELEASE_FAIL`

Any hard correctness/capacity/structural/fallback/public-output gate fails, Fast is slower than Standard on matched production P93 count-1, or matched-baseline latency/allocations regress by >10x.

Do not repair in this task.

---

# 15. Completion report

Report:

### Environment
- candidate SHA;
- baseline SHA;
- machine/CPU;
- Go version;
- OS/arch.

### Correctness
- focused/full test results;
- generated-secret/zero-secret results;
- public max error.

### Capacity
- C2S strict-capacity summary;
- EvalMod/PS/DA strict-capacity summary;
- first failure if any.

### Structure
- production stage row map;
- maximum authoritative rows;
- evidence of no q4+ / full-RNS fallback;
- public input/output contract.

### Performance table
At minimum:
- same-code benchmark baseline/current;
- matched q0=55 P93 baseline/current;
- current q0=56 P93 count1/count3;
- current Standard q0=56 P93 count1;
- ratios and allocation ratios.

### Verdict
One of:
- `QPREFIX_V2_RELEASE_PASS`
- `QPREFIX_V2_PERFORMANCE_REVIEW`
- `QPREFIX_V2_RELEASE_FAIL`

Do not commit production changes.

Successful handoff:
`READY_FOR_WEB_REVIEW`.

If a hard release bug is found, stop with:
`NEEDS_WEB_REVIEW`

and identify the first concrete failing gate.
