# FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020

**Result:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`

**Classification:** `NO_SAFE_OPT_CANDIDATE` (S4 skipped; original Fast pin retained)

## Frozen comparison

The same committed Primary public-chain frontend, LogN13/E32 profile, Q/P, input/workload, and measurement code ran against genuine Standard and the original Fast implementation. No Standard or Secondary production source, parameters, arithmetic, or numerical threshold was changed.

- Primary source/runtime provenance at S5: `2bf9893a266d188d28197a2aef5848057d7ac9af`, clean.
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, detached and clean.
- Original Fast: `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean.
- Harness instrumentation: Primary commit `c486cb2b229f4cc6c0b77f9be2b85066c3b6c213`.
- Runtime: Go 1.26.4, darwin/arm64, Apple M4, `GOMAXPROCS=1`; `GOGC` and `GOMEMLIMIT` unset. Builds occurred before timing. Peak RSS was unavailable from the sandboxed macOS measurement command, so no RSS estimate is inferred from allocations or q-backed bytes.
- Shared frontend SHA-256: `fc059c1fb0b724f6f77734b5ba5e10e177c5fa6db85a7a188f61abb791f08d30`.
- Config/input/workload/profile/QP SHA-256: `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` / `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` / `00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74` / `862d233af63b18a46e0a2cb5e7f74cda93a9dcab6c04fc2f4b1e425c424a7044` / `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`.
- Fixed maximum-error gate: `1e-6`.

## S1–S3 and optional S4

S1 added same-source stage wall-time and allocation instrumentation, separating setup from evaluation, with observer work outside timed kernels. Focused Standard/Fast tests and vet passed. S2 completed five process-isolated cheap-chain samples per backend with zero Bootstrap calls. All paired pre-Bootstrap checks passed the unchanged gate. Eval-only stage-sum medians were Standard `6.102 ms` (samples `5.941, 6.102, 7.541, 5.755, 17.018 ms`) and Fast `3.178 ms` (samples `2.449, 3.187, 3.213, 3.178, 3.160 ms`). Fast Rescale was the largest single eval-only stage at `1.730 ms` median. These are descriptive public-chain measurements, not Bootstrap speedup claims. Full stage samples, allocations, and per-pair errors are in [the S2 baseline](FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-S2-original-baseline.md).

S3 traced public Fast Rescale through `ckks.Evaluator.Rescale` to `fastcore.RescaleWorkspace.ApplyRows`. The bounded profile attributed substantial flat CPU to the required `NTT`/`INTT`, `math/bits.Div64`, and centered-CRT/round/capacity path; no redundant representation-neutral transform/copy was found. A batched-staging rewrite would retain the required arithmetic and change intermediate liveness/cache behavior without profile-backed evidence of a high-impact mechanical gain. Therefore the successful S3 classification is `NO_SAFE_OPT_CANDIDATE`, and optional S4 was skipped without modifying Secondary. See [S3 attribution](FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-S3-attribution.md).

## S5 one-shot paired Bootstrap

The two zero-call held preflights passed before either Bootstrap was released: paired pre-Bootstrap RMSE `1.4749744522248e-11`, max `8.062206332467442e-11`, `pass=true`. Shared provenance hashes matched. Then exactly one public Bootstrap ran per backend; the combined artifact records two total calls and both native output gates passed.

| Backend | Bootstrap time | Allocated bytes / objects | Native output RMSE / max error | Output SNR |
|---|---:|---:|---:|---:|
| Standard | `379.015 ms` | `143,878,040 / 36,648` | `4.8303e-9 / 1.5016e-8` | `102.209 dB` |
| Fast | `574.403 ms` | `960,703,496 / 19,282,450` | `1.1796e-9 / 5.7388e-8` | `114.454 dB` |

The paired Bootstrap output comparison was RMSE `4.914355209707142e-9`, maximum complex difference `4.2372353589552495e-8`, SNR `102.059 dB`, and `pass=true` against `1e-6`. Both outputs were at Level 1 and Scale `2^45`; Fast retained compact Q-prefix storage and zero `c1`. The one-shot Fast wall time and allocation count were higher in this run, but one cold observation is not a statistically stable performance comparison. Do not infer a repeatable Bootstrap slowdown or speedup from it.

The first Standard process was inadvertently launched with closed stdin and exited at the gate with `bootstrap_calls=0` and `bootstrap_attempted=false`; it consumed no Bootstrap budget and is not included in the paired result. The accepted S5 runs were held until the combined preflight passed, and neither backend was retried after its single Bootstrap.

## Validation and handoff

- Fast focused test: `go test ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -count=1` — PASS.
- Standard focused test: `go test -tags lattigo_standard ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -count=1` against the pinned Standard workspace — PASS.
- Fast and Standard focused `go vet` — PASS.
- S2 and S5 fixed `1e-6` gates — PASS; total actual Bootstrap calls: Standard `1/1`, Fast `1/1`.
- No LogN16, extra Bootstrap, warmup, retry, source optimization, or follow-up batch was started.
- Detailed compact machine-readable aggregates: [evidence.json](FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-evidence.json). Execution history: [journal](FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020-journal.md).

Recommended future research for Web review: isolate the centered-CRT/rounding kernel as a distinct mathematical design question before considering any arithmetic change; this batch authorizes no such implementation.
