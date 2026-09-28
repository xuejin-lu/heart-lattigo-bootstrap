# QPREFIX-PERF-OPT-001 Summary

**Classification:** `TRANSACTIONAL_RESCALE_STAGING_READY`

## Implementation

Secondary implementation commit: `d50ff4db757d4a2b9922937a4e7f316fd3f286b9` (`fast-qprefix`).

The Q-prefix Rescale path now uses evaluator-owned reusable staging storage. It transforms the source prefix once, reconstructs and rounds coefficients in sequence, checks capacity, and stages each component's final coefficient-domain residues. The destination is committed only after every component passes: staged rows are converted to the destination domain and written, with destination resizing and metadata/Scale updates deferred until commit. Commit does not repeat the CRT reconstruction or rounding pass. The common two-component staging area is allocated with the evaluator; additional component buffers grow lazily and are reused.

## Correctness and transactionality

- Transactionality tests cover in-place and out-of-place capacity failures. The test tightens the evaluator's test scratch target-half to inject a deterministic failure after component 0 has staged and component 1 fails. It verifies that visible coefficient data, shape/backing capacities, Level, Scale, metadata, and domain flags remain unchanged.
- A degree-2 ciphertext case (three components) matches Standard bit-for-bit and verifies reuse of the third staging buffer.
- The existing BigInt oracle tests and package tests pass.
- Secondary `go test ./...` passes.

Focused validation also passed:

```text
go test ./schemes/ckks/fast
go test ./circuits/ckks/polynomial ./circuits/ckks/bootstrapping
go test -tags fastdiag ./internal/fastdiag
go test -tags fastdiag ./circuits/ckks/bootstrapping -run '^TestFastDiagP93Q55Trace$' -count=1
git diff --check
```

## Performance evidence

Measurements compare the pre-change Secondary production baseline at `74cb73dcff6c552cda0671ed7faea897b448fbbd` with candidate `d50ff4db`. The rows4 baseline was replayed in a temporary detached worktree at the exact pre-change commit, with the same test-only benchmark fixture; no production edits were made there.

**Diagnostics-off P93 E2E** (`BenchmarkFastDiagP93Q55Count1`, one untimed warmup, five timed repetitions):

- Baseline samples: 127.154, 118.252, 121.267, 119.376, 120.331 ms/op; median **120.331 ms/op**.
- Candidate samples: 82.432, 78.983, 81.610, 82.350, 78.037 ms/op; median **81.610 ms/op**.
- Median improvement: **32.2%**.
- Allocations remained flat: observed ranges were 7,615,224–7,615,600 B/op and 15,951–15,957 allocs/op across both runs.

Command:

```text
go test ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkFastDiagP93Q55Count1$' -benchtime=1x -benchmem -count=5
```

**Dedicated four-row Rescale** (LogN13, P93, q0=55, high logical level 15; setup and evaluator warmup outside the timed region):

- Baseline samples: 1.950, 1.922, 1.897, 1.937, 1.875 ms/op; median **1.922 ms/op**.
- Candidate samples: 1.207, 1.198, 1.202, 1.218, 1.214 ms/op; median **1.207 ms/op**.
- Median improvement: **37.2%**.
- Allocations were unchanged: **664 B/op and 18 allocs/op** for all five candidate samples; baseline was also 664 B/op median and 18 allocs/op.

Command:

```text
go test ./schemes/ckks/fast -run '^$' -bench '^BenchmarkFastRescaleQPrefixRows4LogN13P93$' -benchmem -count=5
```

An intermediate candidate had two additional allocations because diagnostic fields were built with diagnostics disabled. Self-review identified this and the implementation was adjusted so those fields are built only in the `fastdiag` code path; the final measurements above show allocations held flat.

## Reusable fastdiag comparison

The reusable comparison ran with baseline `4783c641cee2df5f504b8033905a62a82971da28`, candidate `d50ff4db`, profile `p93-q55`, trace `stage,power,rescale`, one warmup, and five repetitions. Both sides reported `READY`. The trace reported median Rescale total time of 204.304 ms baseline and 131.828 ms candidate. The trace's materialization share changed from 65.31% to 10.76%; the materialization event was 3.654 ms versus 0.428 ms. This is structural diagnostic attribution, not a separate authoritative performance measurement or a causal claim.

## Primary validation and known debt

Primary `go test ./...` still has the documented, unrelated failure:

```text
TestFIX001P3GenuineStandardPublicVsStagedConsistency
P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT
```

This is existing diagnostic maintenance debt, outside this task's code path. It was not modified, bypassed, or weakened. The Primary `cmd/fastdiag` tests and wrapper trace smoke test passed; the trace recorded metadata match and decoded slots within `1e-2`.
