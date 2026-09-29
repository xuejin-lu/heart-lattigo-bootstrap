# Web-review amendment 2 — use the proven clean EXP-002 measurement harness

The previous attempts are superseded for execution purposes.

Do **not** use current Primary source for the detached measurement worktrees.
Do **not** use `lattigo_standard` build tags.
Do **not** patch build constraints.

Use the exact Primary measurement harness commit that already produced matched Standard/Fast EXP-002 stage measurements:

`8186f50e7b591b7f76b39fb89b47c32ad1cc1410`

Why this is authoritative for this timing task:

- it already successfully measured genuine Standard against Fast using the same public bootstrap API;
- its `configs/bootstrap_config.logN16.json` is byte-for-byte semantically the requested LogN16 configuration;
- its `stage_runner.go` records the same required stage names and full-bootstrap call;
- key generation/evaluator construction/warmup remain outside measured regions;
- it has no later Fast-only diagnostic compile dependencies;
- both compared backends therefore use exactly the same detached Primary harness source.

Frozen Secondary refs remain:

- Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- current Q-prefix: `82601ea2517edc14784c9da250426169a1b221c7`

The fact that the control-plane Primary currently contains newer files is irrelevant to the timing measurement. Result artifacts are copied back and committed on current Primary only after the detached measurements finish.

## Revised exact detached-worktree commands

### Standard

```sh
BASE_STD=$(mktemp -d /tmp/logn16-standard.XXXXXX)

git -C /Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap \
  worktree add --detach "$BASE_STD/heart-lattigo-bootstrap" \
  8186f50e7b591b7f76b39fb89b47c32ad1cc1410

git -C /Users/xuejinlu/Developer/xuejin-lu/lattigo \
  worktree add --detach "$BASE_STD/lattigo" \
  5dbffbdea05394de2ca3a432ed5318aa832e3f40

cd "$BASE_STD/heart-lattigo-bootstrap"

go test ./...

go run . \
  -stages \
  -config configs/bootstrap_config.logN16.json \
  -warmup 1 \
  -repetitions 7 \
  -out "$BASE_STD/standard-logN16.json"
```

### Current Q-prefix Fast

```sh
BASE_FAST=$(mktemp -d /tmp/logn16-fast.XXXXXX)

git -C /Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap \
  worktree add --detach "$BASE_FAST/heart-lattigo-bootstrap" \
  8186f50e7b591b7f76b39fb89b47c32ad1cc1410

git -C /Users/xuejinlu/Developer/xuejin-lu/lattigo \
  worktree add --detach "$BASE_FAST/lattigo" \
  82601ea2517edc14784c9da250426169a1b221c7

cd "$BASE_FAST/heart-lattigo-bootstrap"

go test ./...

go run . \
  -stages \
  -config configs/bootstrap_config.logN16.json \
  -warmup 1 \
  -repetitions 7 \
  -out "$BASE_FAST/fast-qprefix-logN16.json"
```

## Compatibility gate

Before timing acceptance:

1. both detached Primary harnesses must report commit `8186f50e...`;
2. Standard must compile and run without Fast package dependencies;
3. current Q-prefix must compile against the old public harness without source patching;
4. both effective parameter metadata blocks must match exactly;
5. all staged/full metadata checks must pass.

If current Q-prefix cannot compile against this clean historical harness **without source modification**, stop as `LOGN16_STANDARD_FAST_TIMING_BLOCKED`; do not patch either backend.

## Result provenance

The summary must explicitly distinguish:

- **measurement harness commit**: `8186f50e...`
- **current control-plane Primary commit**: the commit that stores the results
- **Standard Secondary commit**
- **Fast Secondary commit**

This avoids pretending the old measurement harness is the current project state.

---

# QPREFIX-PERF-MEASURE-LOGN16-001 — Matched Standard vs Q-prefix Bootstrap Timing

## Status

Executable measurement-only task.

## Task class

`M — Performance Measurement`

## Suspended task

Temporarily suspend:
`FAST-STANDARD-NUMERICAL-DIAG-002`

Do not delete or modify its spec. Resume only after this LogN16 measurement is reviewed.

## Goal

Measure current Q-prefix versus genuine Standard bootstrapping speed at:

- `N = 2^16` / `LogN=16`;
- full-slot config from `configs/bootstrap_config.logN16.json`;
- identical host, primary harness, config, warmup and repetition count.

Report:

1. complete Bootstrap latency;
2. per-stage latency;
3. Standard/Fast speedup for full Bootstrap;
4. Standard/Fast speedup per stage;
5. allocation bytes and allocation count for full Bootstrap;
6. effective generated Q/P parameters and LogSlots to prove the two runs are matched.

No production code change.

## Frozen provenance

Use Primary harness source exactly:
`2b1022e1820e612e9209094dd672be15d06a8d13`

Use genuine Standard Secondary:
`5dbffbdea05394de2ca3a432ed5318aa832e3f40` (`main`)

Use current Q-prefix Secondary:
`82601ea2517edc14784c9da250426169a1b221c7` (`fast-qprefix`)

The Fast ref contains only control-plane pointer changes after the accepted production arithmetic; record the exact ref nonetheless.

## Config

Use exactly:

`configs/bootstrap_config.logN16.json`

Expected top-level values include:

- `log_n = 16`
- `log_default_scale = 45`
- `q0 = [55]`
- `q_slots_to_coeffs = [39,39,39]`
- `q_coeffs_to_slots = [56,56,56,56]`
- `p = [61,61,61,61,61]`
- `log_slots = -1` (use generated max slots)
- Mod1 degree 30
- DoubleAngle 3
- K 16
- LogMessageRatio 10.

Do not edit the config.

## Measurement protocol

Use:

- warmup = 1
- measured repetitions = 7
- `-stages`

The stage harness measures a staged pipeline and a separate full `eval.Bootstrap` call per repetition.

Key generation, evaluation-key generation, evaluator construction, encoding setup and warmup are outside timed measurement regions.

## Required stage names

Collect medians for:

- `pack_and_switch_n1_to_n2`
- `scale_down`
- `mod_up`
- `coeffs_to_slots`
- `eval_mod_real`
- `eval_mod_imag`
- `slots_to_coeffs`
- `unpack_and_switch_n2_to_n1`
- `full_bootstrap`

If a stage is absent, report it explicitly rather than assigning zero.

## Safe worktree requirement

Do **not** checkout `main` in the authoritative Secondary worktree.

Create two temporary sibling worktree pairs so the Primary's local Go replace `../lattigo` resolves naturally.

Example layout:

```text
/tmp/logn16-standard-XXXX/
  heart-lattigo-bootstrap/
  lattigo/

/tmp/logn16-fast-XXXX/
  heart-lattigo-bootstrap/
  lattigo/
```

Each Primary detached worktree uses the same frozen Primary commit.

Each Secondary detached worktree uses its respective frozen Standard/Fast commit.

Remove only task-created clean worktrees at the end.

## Exact commands

Adapt only the temporary root paths.

### Standard

```sh
BASE_STD=$(mktemp -d /tmp/logn16-standard.XXXXXX)

git -C /Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap \
  worktree add --detach "$BASE_STD/heart-lattigo-bootstrap" \
  2b1022e1820e612e9209094dd672be15d06a8d13

git -C /Users/xuejinlu/Developer/xuejin-lu/lattigo \
  worktree add --detach "$BASE_STD/lattigo" \
  5dbffbdea05394de2ca3a432ed5318aa832e3f40

cd "$BASE_STD/heart-lattigo-bootstrap"

# Standard-only compile preflight. The Primary harness intentionally uses
# lattigo_standard to exclude Fast-only diagnostic runners while keeping
# the ordinary Standard bootstrap measurement path unchanged.
go test -tags lattigo_standard ./...

go run -tags lattigo_standard . \
  -stages \
  -config configs/bootstrap_config.logN16.json \
  -warmup 1 \
  -repetitions 7 \
  -out "$BASE_STD/standard-logN16.json"
```

### Current Q-prefix Fast

```sh
BASE_FAST=$(mktemp -d /tmp/logn16-fast.XXXXXX)

git -C /Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap \
  worktree add --detach "$BASE_FAST/heart-lattigo-bootstrap" \
  2b1022e1820e612e9209094dd672be15d06a8d13

git -C /Users/xuejinlu/Developer/xuejin-lu/lattigo \
  worktree add --detach "$BASE_FAST/lattigo" \
  82601ea2517edc14784c9da250426169a1b221c7

cd "$BASE_FAST/heart-lattigo-bootstrap"

go run . \
  -stages \
  -config configs/bootstrap_config.logN16.json \
  -warmup 1 \
  -repetitions 7 \
  -out "$BASE_FAST/fast-qprefix-logN16.json"
```

## Backend identity check

Before accepting results:

### Standard detached Secondary
Confirm:
- exact commit `5dbffbde...`;
- `bootstrapping.NewEvaluator` does not select a Fast compatibility evaluator.

### Fast detached Secondary
Confirm:
- exact commit `82601ea...`;
- evaluation keys select the Fast-compatible public constructor path;
- `bootstrapping.NewEvaluator(...)` delegates to current `FastEvaluator`.

Do not assume branch names alone prove the backend path.

## Matched-run validation

Before comparing timing, verify:

- same Primary commit;
- same host / OS / architecture / Go version;
- same config;
- same effective residual LogN;
- same effective bootstrap LogN;
- same effective bootstrap LogSlots;
- same effective Q-chain bit lengths;
- same effective P-chain bit lengths;
- same warmup/repetition count;
- each staged/full metadata correctness check passes.

If any identity condition fails, classify as unmatched and do not compute a headline speedup.

## Summary calculations

For each backend and each stage:

- report all 7 raw elapsed times;
- median;
- mean;
- min;
- max.

For allocations:
- median bytes/op;
- median allocs/op.

For matched stages:

[
S_{stage} = T_{standard,median}/T_{fast,median}.
]

For full Bootstrap:

[
S_{BTS} = T_{standard,median}/T_{fast,median}.
]

Also report absolute saved time:

[
Delta T = T_{standard,median}-T_{fast,median}.
]

For EvalMod, additionally aggregate per repetition:

[
T_{EvalMod,total}=T_{real}+T_{imag}
]

and report its median for each backend and its speedup.

Do not sum independent stage medians and call that the full Bootstrap time.

## Numerical-quality caveat

This task is timing-only.

The accepted LogN13 result currently classifies Fast numerical quality as degraded. Do not infer LogN16 correctness from successful timing execution.

Stage harness metadata equality is not a numerical-precision proof.

## Output artifacts

Copy raw JSONs into Primary results:

- `results/QPREFIX-PERF-MEASURE-LOGN16-001-standard.json`
- `results/QPREFIX-PERF-MEASURE-LOGN16-001-fast.json`

Write:

- `results/QPREFIX-PERF-MEASURE-LOGN16-001-summary.md`

The summary must contain:
- frozen refs;
- host/environment;
- effective parameter match;
- raw full-bootstrap times;
- full median comparison/speedup;
- stage table with Standard median, Fast median, speedup and saved milliseconds;
- allocation comparison;
- EvalMod aggregate;
- limitations.

## Classification

Return exactly one:

- `LOGN16_STANDARD_FAST_TIMING_READY`
- `LOGN16_STANDARD_FAST_TIMING_UNMATCHED`
- `LOGN16_STANDARD_FAST_TIMING_BLOCKED`

No performance threshold is required; this is measurement.

## Completion

Commit/push only the Primary measurement artifacts.

Do not commit Secondary.

Return:

`LOGN16_STANDARD_FAST_TIMING_READY`

plus:
- `STANDARD_FULL_MEDIAN_MS=...`
- `FAST_FULL_MEDIAN_MS=...`
- `FULL_SPEEDUP=...`
- `STANDARD_EVALMOD_TOTAL_MEDIAN_MS=...`
- `FAST_EVALMOD_TOTAL_MEDIAN_MS=...`

Then:

`READY_FOR_WEB_REVIEW`.


## Web-review amendment — Standard build tag authorization

The first execution was correctly blocked before timing because the frozen Primary harness contains Fast-only diagnostic runner files guarded by:

`//go:build !lattigo_standard`

while `fast_diagnostics_standard_stub.go` provides the corresponding Standard stubs under:

`//go:build lattigo_standard`

Therefore the Standard measurement is explicitly authorized to compile with:

`-tags lattigo_standard`

This is a harness compile-selection tag only. It must not:
- change Standard bootstrap arithmetic;
- change parameters/config;
- alter the stage measurement implementation;
- be used for the Fast measurement.

Required Standard preflight:

```sh
go test -tags lattigo_standard ./...
```

Required Standard measurement command:

```sh
go run -tags lattigo_standard . \
  -stages \
  -config configs/bootstrap_config.logN16.json \
  -warmup 1 \
  -repetitions 7 \
  -out "$BASE_STD/standard-logN16.json"
```

Fast remains:

```sh
go run . \
  -stages \
  -config configs/bootstrap_config.logN16.json \
  -warmup 1 \
  -repetitions 7 \
  -out "$BASE_FAST/fast-qprefix-logN16.json"
```

Before accepting results, record that Standard was built with `lattigo_standard` and Fast was built without it.
