# FAST-Q0123-RESCALE-REALISM-OPT-PILOT-005

**Outcome:** `REJECT` — the source audit supports zero `c1` across the accepted Fast EvalMod path, but the one bounded Rescale-loop prototype did not improve the matched workload and is not promoted.

## Scope and provenance

- Primary task/spec source: `2e7bec24ede5f9ac50f9508555540d05c2826552` / `specs/FAST-Q0123-RESCALE-REALISM-OPT-PILOT-005.md`.
- Secondary production pin: clean `fast-qprefix@2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`; no production file, commit, or branch in the actual Secondary checkout was changed.
- The Task004 testing.B patch was reused in a detached temporary Secondary copy. The only fixture extension added `c1=0` beside its existing `c1=-c0` case. The Rescale-loop candidate was also confined to that temporary copy and is not retained as a patch.
- Machine/runtime: Apple M4, `darwin/arm64`, Go `go1.26.4`, `GOMAXPROCS=1`, `GOGC=100`; no `fastdiag` build tag or selected tracing.
- Exact LogN16 Task006 Q/P fingerprint: `6c82ca724cbfdce2083b2ef2c18d53f81e658a6f28ff283e9a671d276710ec5f`; N=65,536; two components; Level 12→11; four Q-prefix rows; Scale `2^120`→`2^120/q12` (log2 `59.999999999986`); NTT=true, Montgomery=false; in-place public Fast `Rescale`.
- Existing `c1=-c0` fixture SHA-256: `52ce8159b612f860fd081c31a1a3a9bc7a681d34fa759e95e577c363213c93ce`. Matched `c1=0` fixture SHA-256: `4acbf50de8ed655e62b598ff7f23c6f592fc0384c3cfb1c144f79f8c8144260a`.

## EvalMod component-state audit

`FastEvaluator.BootstrapMany` calls `validateFastBootstrapPublicInputs`, which rejects any nonzero active-row `c1` at the public zero-secret simulation boundary. The reviewed code then preserves a zero `c1` through the relevant path:

- `FastEvaluator.ScaleDown` uses integer scaling and Rescale; those operations process ciphertext components independently.
- `FastEvaluator.ModUp` / `modUpBasisAtLevel` lift, NTT-transform, trace, and Montgomery-transform each component independently.
- `dft.FastEvaluator.CoeffsToSlotsNewWithRestorePlan` and `SlotsToCoeffsNewQPrefixRows` compose linear component-wise CKKS operations, so zero `c1` stays zero.
- `mod1.FastEvaluator.EvaluateNewQPrefixRows` clones the authorized rows, adds constants to the `c0` path, evaluates the polynomial, and applies DoubleAngle. Fast ciphertext multiplication forms `c1` from cross terms; when both operands have zero `c1`, those terms remain zero. Add/constant/Rescale operations preserve that state.
- `FastEvaluator.EvalMod` delegates to that Mod1 path and only resets Scale metadata.

This is a source-backed invariant, not an assertion from a newly captured live Bootstrap ciphertext. No Bootstrap or traced capture was run. The coefficient values remain deterministic synthetic data; only the `c1` component state was changed to match the accepted zero-secret mode.

## Correctness and matched A/B timing

Both fixture variants passed the existing `bigIntRescaleOracle` exactly for every coefficient in both components. The test also confirmed the intended zero/nonzero `c1` state before Rescale, zero preservation after Rescale, Level/Scale progression, representation flags, and four-row output.

Each cell below is one independent `-benchtime=1x` sample. Fixture setup/reset stayed outside the timed region. Baseline is the unmodified `2cb8bc7` kernel; candidate is the isolated prototype precomputing each Rescale transition's divisor, target Level, centered-capacity half, and prefix product once per operation rather than repeating invariant lookups/calculation inside every coefficient loop.

| c1 state | Baseline samples (ms/op) | Baseline median (range) | Candidate samples (ms/op) | Candidate median (range) | Median change |
|---|---|---:|---|---:|---:|
| zero | 16.755542, 16.443500, 16.101626, 16.234750, 16.279499 | 16.279499 (16.101626–16.755542) | 16.974874, 16.022250, 16.364667, 18.612625, 16.291209 | 16.364667 (16.022250–18.612625) | +0.52% slower |
| `-c0` | 21.151334, 21.139459, 22.020541, 21.119209, 21.136625 | 21.139459 (21.119209–22.020541) | 21.892583, 21.276417, 22.515417, 21.972291, 21.409459 | 21.892583 (21.276417–22.515417) | +3.56% slower |

Median allocations remained 22/op in both baseline and candidate for both states. Median bytes/op were 816 for each baseline; candidate was 832 for zero `c1` and 816 for `c1=-c0`. The candidate had neither a repeatable timing benefit nor an allocation benefit, so it is rejected. No second candidate was attempted. The zero-`c1` baseline was about 22.99% faster than the matched synthetic `c1=-c0` baseline; this is a component-state microbenchmark observation only, not an EvalMod or Bootstrap speedup claim.

## Validation and limitations

Passed in the isolated prototype:

- `go test ./schemes/ckks/fast ./schemes/ckks/internal/fastcore -run '^(TestFastRescale|TestFixedWidthCRT|TestFixedWidthRoundedDivision|TestQ0123FastRescale|TestRescale|TestBatch027CoreHelperFixtureUsesCanonicalQPrefix|TestTask005LogN16RescaleC1StatesMatchOracle)' -count=1`
- `go vet ./schemes/ckks/fast ./schemes/ckks/internal/fastcore`
- `git diff --check`
- Five uninstrumented samples per `c1` state, before and after the single candidate.

The broader `go test ./schemes/ckks/fast ./schemes/ckks/internal/fastcore -count=1` also exposed two unrelated historical expectation failures in `TestPublicCKKSEvaluatorFastMulRelinFailsClosedWithStandardRelinKey` (higher-degree and nil-input error text). The same test was run against the untouched, clean Secondary production checkout at the same `2cb8bc7` pin and failed identically; neither the Rescale candidate nor the Task005 fixture changes those paths. No unrelated repair was made.

No full Bootstrap, benchmark sweep, selected trace, Secondary production change/commit/push, or claim of end-to-end gain was made. Task006's original trace and Task004 evidence remain unchanged. Recommendation: reject this coefficient-loop candidate; retain the source-backed zero-`c1` realism result and the negative candidate measurement for Web review.
