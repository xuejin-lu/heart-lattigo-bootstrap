# FAST-LOGN16-RESCALE-INSTRUMENTATION-BIAS-004

**Classification:** `MATCHED_SHAPE_SYNTHETIC_INSTRUMENTATION_PERTURBATION_MEASURED`

## Conclusion

Q0123 Rescale remains a reasonable next **constrained microkernel candidate**, but Task003's traced Rescale/EvalMod shares cannot be treated as uninstrumented cost. On the same source-bound synthetic LogN16 Q0123 fixture, uninstrumented A had a 21.340 ms median and tagged/mask-off B had 21.402 ms. Rescale tracing in C raised the median to 31.852 ms (+48.82% vs B). A/B ranges overlap; C's minimum is above B's maximum. This measures one in-place Rescale, not Bootstrap latency or an end-to-end speedup, and does not establish the historical ciphertext's coefficient distribution.

One next hypothesis, for a separately authorized implementation task: test a fixed-Q0123 coefficient-loop reduction while preserving centered rounding, every capacity check, authoritative-row access, compact backing, representation flags, and transactional failure behavior. Require the existing BigInt oracle, Standard differential, capacity/transactionality, and Q-prefix row-authority gates, then repeat the uninstrumented A benchmark. A kernel win alone is not proof of an EvalMod or Bootstrap gain.

## Fixture and provenance

- Primary report source: `979ce3df0f284689a98bc5ed1daca3fafd642371`.
- Secondary production source: clean `fast-qprefix` at `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`; the production worktree remained unchanged. The isolated copy was detached at this commit and received only the test extension captured in [`FAST-LOGN16-RESCALE-INSTRUMENTATION-BIAS-004-benchmark.patch`](FAST-LOGN16-RESCALE-INSTRUMENTATION-BIAS-004-benchmark.patch).
- Effective ring: LogN=16, N=65,536, exact Task006 Q/P. Canonical ordered Q/P SHA-256: `6c82ca724cbfdce2083b2ef2c18d53f81e658a6f28ff283e9a671d276710ec5f`; the fixture reconstructed and verified this fingerprint from the effective primes.
- Fixture: synthetic, deterministic, uniform over centered Q0123 coefficient capacity; not the captured Task006 ciphertext or its claimed coefficient distribution. Two components, logical Level 12, four compact authoritative rows q0..q3, Scale `2^120`, one in-place Rescale to Level 11 and Scale log2 `59.999999999986` (about `2^60`), output width 4.
- Representation: NTT=true, Montgomery=false. Task006 records `in_place=true`; its event schema does not record NTT/Montgomery, so those are explicit fixture choices, not recovered trace facts.
- Fixture ciphertext SHA-256: `52ce8159b612f860fd081c31a1a3a9bc7a681d34fa759e95e577c363213c93ce`.
- Fast Rescale source SHA-256: `38b8cfd3917e3b42282e1a63231441d7a6d203266c5c713426866301a8dd7f99`. `fastdiag` enabled/disabled source SHA-256: `0667e73f476d49b3d448c25600378c8091bcaad33da6ca0d3be08331e999b406` / `30a1672f4d86499b02d2ac2ea87b4edeac8817b8365ea14231d410e9270d0883`.
- Test extension SHA-256: `da43de3bf61cb02123f80623723e2b2364c71e841d17e58f4e9a8d6e1939956b`; saved diff SHA-256: `858b46b6bc055d896a4eb6e8acc50af0f5e51d390c2cd314066e03078b737c88`.

The unchanged production path is `fast.Evaluator.Rescale` → `fastcore.RescaleWorkspace.ApplyRows` → `rescaleComponent`; `fast.NewCiphertext` supplies compact Q-prefix backing. No full-RNS fallback was used. The existing BigInt centered-CRT oracle matched every output coefficient in both components exactly, including target-capacity assertions.

## A/B/C measurement

Go `go1.26.4 darwin/arm64`, Apple M4, `GOMAXPROCS=1`, `GOGC=100`, same warmed Go build cache, one excluded warm-up per sample, five `-benchtime=1x` samples per mode. Fixture construction/reset occurred with the benchmark timer stopped. Source commit, Rescale arithmetic source, fixture hash, public in-place call, and output checks were identical.

| Mode | Build/runtime selection | Events/call | Five samples (ms/op) | Median (range) | Median allocs/op | Median B/op (range) |
|---|---|---:|---|---:|---:|---:|
| A | No `fastdiag` tag (`Enabled=false`) | 0 | 22.110375, 21.368416, 21.312291, 21.198917, 21.340167 | 21.340167 (21.198917–22.110375) | 22 | 816 (816–832) |
| B | `-tags=fastdiag`, empty Rescale mask (`Enabled=true`, `Selected=false`) | 0 | 21.694625, 21.365292, 21.393624, 21.825458, 21.402376 | 21.402376 (21.365292–21.825458) | 28 | 880 (736–896) |
| C | `-tags=fastdiag`, `ConfigureCSV("rescale")` | 13 | 32.973333, 31.626625, 31.851834, 31.858291, 31.525458 | 31.851834 (31.525458–32.973333) | 77 | 7,680 (7,536–7,696) |

B vs A is +0.29% at the median with overlapping ranges. C vs B is +48.82% with non-overlapping ranges; C adds 49 median allocations and 6,800 median bytes/op. C−B is the **combined selected-instrumentation effect** (clock reads, selected-scope branches, span/event collection), not isolated `time.Now` or collector-only cost.

In `rescaleComponent`, the selected path makes two `StartTimer` and two `ElapsedNS` calls per coefficient: four clock-related calls plus scope checks. Each EvalMod Rescale processes 2 × 65,536 = 131,072 iterations, or up to 524,288 clock-related calls; across Task003's 28 roots that is 14,680,064 calls. Each C operation emitted the expected 13-event Rescale tree. Do not multiply this microbenchmark delta into a Bootstrap estimate: trace scopes overlap and this synthetic coefficient distribution is not the captured workload.

## Commands and validation

The extension ran in a detached isolated copy. Every listed Go command used `GOCACHE=<isolated-copy>/gocache` after the host-default Go cache returned a sandbox write denial; no system permissions or production repo were changed.

```sh
TASK004_MODE=A GOMAXPROCS=1 go test ./schemes/ckks/fast -run '^TestTask004LogN16RescaleFixtureMatchesOracle$' -count=1 -v
TASK004_MODE=B GOMAXPROCS=1 go test -tags=fastdiag ./schemes/ckks/fast -run '^TestTask004LogN16RescaleFixtureMatchesOracle$' -count=1 -v
TASK004_MODE=C GOMAXPROCS=1 go test -tags=fastdiag ./schemes/ckks/fast -run '^TestTask004LogN16RescaleFixtureMatchesOracle$' -count=1 -v

TASK004_MODE=A GOMAXPROCS=1 GOGC=100 go test ./schemes/ckks/fast -run '^$' -bench '^BenchmarkTask004LogN16Rescale$' -count=5 -benchtime=1x -benchmem -cpu=1 -v
TASK004_MODE=B GOMAXPROCS=1 GOGC=100 go test -tags=fastdiag ./schemes/ckks/fast -run '^$' -bench '^BenchmarkTask004LogN16Rescale$' -count=5 -benchtime=1x -benchmem -cpu=1 -v
TASK004_MODE=C GOMAXPROCS=1 GOGC=100 go test -tags=fastdiag ./schemes/ckks/fast -run '^$' -bench '^BenchmarkTask004LogN16Rescale$' -count=5 -benchtime=1x -benchmem -cpu=1 -v

go test ./schemes/ckks/fast ./schemes/ckks/internal/fastcore -run 'Test(FastRescaleMatchesBigIntCenteredCRTOracleAtEveryPrefixWidth|FastRescaleToMatchesSequentialBigIntOracleAcrossContractions|FastRescaleMatchesStandardGeneratedLogN16|FastRescaleSupportsBothNTTRepresentations|Q0123FastRescaleMatchesStandardForBoundedCoefficients|RescaleCapacityFailureAfterEarlierComponentStagingIsTransactional|RescaleApplyRowsFastDiagDoesNotFabricateMaterializationOnCapacityFailure)$' -count=1
go test -tags=fastdiag ./schemes/ckks/fast ./schemes/ckks/internal/fastcore -run 'Test(FastRescaleMatchesBigIntCenteredCRTOracleAtEveryPrefixWidth|FastRescaleToMatchesSequentialBigIntOracleAcrossContractions|FastRescaleMatchesStandardGeneratedLogN16|FastRescaleSupportsBothNTTRepresentations|Q0123FastRescaleMatchesStandardForBoundedCoefficients|RescaleCapacityFailureAfterEarlierComponentStagingIsTransactional|RescaleApplyRowsFastDiagDoesNotFabricateMaterializationOnCapacityFailure)$' -count=1
go vet ./schemes/ckks/fast ./schemes/ckks/internal/fastcore
go vet -tags=fastdiag ./schemes/ckks/fast ./schemes/ckks/internal/fastcore
```

All A/B/C fixture-oracle tests passed; A/B emitted zero events, C emitted 13. Both focused regression-test variants passed, including BigInt/Standard differential, NTT-representation, Q0123, and capacity-transactionality gates. Both `go vet` variants passed; `gofmt -d` on the extension was empty. No Bootstrap, full experiment, parameter sweep, or Secondary production edit was run. Task006's original raw trace remains `TRACE_UNVERIFIED`; this measurement neither mutates nor upgrades its classification.
