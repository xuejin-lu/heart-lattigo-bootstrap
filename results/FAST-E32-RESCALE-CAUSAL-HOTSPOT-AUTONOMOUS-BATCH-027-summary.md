# Batch027 — E32 Rescale causal hotspot study

- Classification: **`HOTSPOT_NOT_RESOLVED`**
- Handoff: **`READY_FOR_WEB_REVIEW`**
- Secondary test-only commit: `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` (`fast-qprefix`, pushed and clean)
- Primary task-start commit: `83a2953aca82f38357a43b099aa8b0e5fb3e9ce8`
- Budget: **0 Bootstrap calls, 0 Standard performance calls, 0 P93 runs, 0 production candidates**.

## Fixture and correctness

The old Batch026 fixture is unrecoverable from its archived seed alone: its RNG, sampling, NTT conversion, serialization, and digest recipe were not retained. Record it as `HISTORICAL_FIXTURE_RECIPE_MISSING`; archived SHA `001855dc4aba15d1da02e71df32242cb019c8976fa2ebede386bebad43ab1247` and prior unsuccessful reconstruction SHA `1652bf4f8441aa6b0d97929efe79ab150d3d39851e06f883201956ff5054d861` are not interchangeable and neither is the Batch027 identity.

The separately versioned fixture `batch027-e32-rescale-synthetic-v1` is built from the actual LogN13/E32 bootstrapping parameter chain. It matches config SHA `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` and Q/P SHA `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`. Shape: N=8192, Level 9→8, degree 1, c0/c1 nonzero, four authoritative Q rows at both levels, in-place public Fast Rescale, NTT=true, Montgomery=false, Scale `2^45`. It is deterministic synthetic input, **not** the captured Batch024/025 Bootstrap ciphertext.

The retained SplitMix64 v1 generator, centered-Q0123 construction, row NTT conversion, and canonical `canonical-ntt-u64le-v1` digest are in Secondary test source. Two independent same-seed reconstructions matched all authoritative coefficients, metadata, and the frozen SHA-256 `1de1a696fe7b596105cf236e7799b015049ae78e63e09e2ad6314be730dc3cb8`; seed+1 changed the fixture. The real Fast `Rescale` output matched an independent `math/big` centered-CRT and signed-rounding oracle for c0/c1, in-place and out-of-place, all q0–q3 output coefficients, Level, Scale, compact backing, and strict output capacity.

## New Batch027 synthetic baseline

Environment: Go 1.26.4, darwin/arm64, Apple M4, `GOMAXPROCS=1`, `GOGC=100`, regular build. Each public Rescale sample is one second; the benchmark used the same immutable fixture and excluded construction, one evaluator-scratch warm-up, and reset copies from timing.

| Measurement | Five samples | Median | Range | B/op | allocs/op |
|---|---|---:|---:|---:|---:|
| Public Fast in-place Rescale | 2,702,420; 2,712,634; 2,719,800; 2,731,379; 2,734,809 ns/op | 2,719,800 ns/op | 2,702,420–2,734,809 ns/op | 675 | 20 |
| `fastcore` `reconstructQPrefix(4)` helper | 56.70; 56.71; 58.37; 56.89; 56.56 ns/op | 56.71 ns/op | 56.56–58.37 ns/op | 0 | 0 |
| `fastcore` `roundedMagnitude192(q9)` helper | 8.548; 8.903; 8.346; 8.092; 8.260 ns/op | 8.346 ns/op | 8.092–8.903 ns/op | 0 | 0 |
| `fastcore` `signedResidue192(q0..q3)` helper | 9.332; 9.397; 10.19; 10.16; 9.872 ns/op | 9.872 ns/op | 9.332–10.19 ns/op | 0 | 0 |

Helper measurements use the same seeded representative coefficient distribution but are **non-additive** and do not establish their share of full Rescale time. They provide no proof of an instruction-level or end-to-end causal hotspot. `go tool pprof` is unavailable (`go: no such tool "pprof"`). No safe, source-supported optimization hypothesis was strong enough to justify P3; no candidate or production arithmetic change was attempted. Batch026's 100 ms baseline/candidate values are historical context only and were not pooled with this new baseline.

## Validation

Passed focused Secondary Rescale/fixture tests, the independent real-path oracle, `go vet` on the three affected packages, `git diff --check`, and `gofmt -d` (no diff). Exact commands, sample provenance, fixture byte rules, and source SHAs are in the journal and compact evidence JSON.
