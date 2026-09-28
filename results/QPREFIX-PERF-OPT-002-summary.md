# QPREFIX-PERF-OPT-002 — Fixed-Width Division Deduplication

**Classification:** `FIXED_WIDTH_DIVISION_DEDUP_READY`

## Scope and provenance

- Primary `main` at measurement: `706e43857c6ae6e7ecbe1bc890f378736e200c21`.
- Secondary `fast-qprefix` measurement base: `40c8b041a310f335a9475cf0c63d300e164bce61`, with only this task's uncommitted patch applied; committed implementation: `6930cf6cb3c71ce139a1eb42eede7be335b7174c`.
- Go `go1.26.4`, `darwin/arm64`, Apple M4.
- The same-session before/after benchmark commands used `-benchtime=100ms -benchmem -count=7`; medians are over all seven samples. No sample was discarded.

## Implementation and exactness

Changed only the fixed-width Fast-CKKS Rescale path and test-only support:

- `schemes/ckks/fast/q012.go`: removed the redundant initial `bits.Div64(0, hi%modulus, modulus)` in `mod128By64` and `mod192By64`. The remainder of a limb already reduced modulo `modulus` is exactly that reduced limb. Subsequent carry-propagating `bits.Div64` steps are unchanged; the zero-modulus panic domain is unchanged because `% modulus` remains.
- `schemes/ckks/fast/q0123.go` and `rescale_qprefix.go`: added Rescale-specific prepared CRT helpers and pass the already-validated q01/q012 products from scratch into rows3/rows4 reconstruction. Generic `crtQ012` and `crtQ0123` behavior remains unchanged.
- Added fixed-seed `math/big` oracles for 128-/192-bit modulo and rounded magnitude, plus exact all-width comparisons of prepared reconstruction against both generic CRT and BigInt. The fixed-width benchmark fixture checks reconstructed production output against Fast and Standard Rescale before timing.
- Audited the proposed `bits.Div64` rewrite for `roundedMagnitude192` but did not adopt it: its high-limb `/` and `%` already compile to one `UDIV` plus `MSUB`, and the exact `bits.Div64` candidate was slower.

## Correctness and regression tests

- Focused BigInt / prepared-CRT / production-and-Standard fixture tests — **PASS**.
- `go test ./schemes/ckks/fast -count=1` — **PASS**.
- `go test ./...` (Secondary) — **PASS**.
- Existing Rescale suite covers every prefix width, centered-CRT oracle behavior, sequential `RescaleTo`, NTT representations, transactionality/capacity failures, and higher-degree staging; no failures.

## Helper benchmarks

Median `ns/op` (before → after); helper/phase benchmarks reported 0 B/op and 0 allocs/op:

| Helper | Before | After | Change |
|---|---:|---:|---:|
| `mod128By64` | 3.966 | 2.648 | −33.23% |
| `mod192By64` | 7.406 | 6.463 | −12.73% |
| rows3 CRT reconstruction | 15.00 | 13.23 | −11.80% |
| rows4 CRT reconstruction | 39.89 | 35.04 | −12.16% |
| `roundedMagnitude192`, source `/` + `%` | 6.514 | 6.477 | −0.57% (unchanged implementation) |
| exact `bits.Div64` rounded candidate | 7.510 | 7.435 | still 14.8% slower than source form after change |

The rounded candidate numbers are measurements only; production retains the source `/` and `%` form.

## Fixed-width phase and full Rescale

Fixed-width phase median (`µs/op`):

| Width | Before | After | Change |
|---|---:|---:|---:|
| rows2 (`WIDTH_COUNTERFACTUAL_ONLY`) | 222.700 | 219.964 | −1.23% |
| rows4 | 532.672 | 467.548 | **−12.23%** |

Authoritative diagnostics-off `BenchmarkFastRescaleQPrefixRows4LogN13P93` raw samples (`ns/op`):

- Before: `1225045, 1206153, 1232024, 1221811, 1205501, 1202913, 1202359`; median **1,206,153 ns/op**.
- After: `1144718, 1144521, 1119985, 1122688, 1120525, 1124711, 1144522`; median **1,124,711 ns/op**, **6.75% faster**.

Matched full-API rows4 fixture: `1,223,509 → 1,117,591 ns/op` (**8.66% faster**). Matched rows2 counterfactual control: `554,485 → 520,989 ns/op` (**6.04% faster**, `WIDTH_COUNTERFACTUAL_ONLY`; not a policy-eligibility claim).

Full Rescale allocation result: before **664–665 B/op, 18 allocs/op**; after **664–679 B/op, 18 allocs/op**. The fixed-width and helper loops allocate zero; no new per-coefficient or persistent Rescale allocations were introduced.

## Diagnostics-off P93 Count-1 E2E

Command: `go test ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkFastDiagP93Q55Count1$' -benchtime=100ms -benchmem -count=7`. The benchmark performs one untimed warmup before resetting its timer.

- Before raw samples (`ns/op`): `78706125, 78642500, 79169688, 78645333, 79291188, 79219521, 80671958`; median **79,169,688 ns/op**.
- After raw samples (`ns/op`): `74330000, 74068375, 74000208, 74881979, 75121896, 74357312, 73976750`; median **74,330,000 ns/op**, **6.11% faster**.
- Allocation samples stayed approximately **15,951–15,956 allocs/op** and **7,615,216–7,616,888 B/op**; no material allocation regression.

## Low-overhead power trace

Exact command before and after: `./scripts/fastdiag trace --profile p93-q55 --trace power --warmup 1 --repetitions 7`. For each run, the listed Rescale value is the sum of power-scope Rescale events; generated-power is the corresponding parent event. Timings are derived from the compact trace, not deep Rescale tracing.

| Per-bootstrap aggregate | Before raw (ns) | Before median | After raw (ns) | After median | Change |
|---|---|---:|---|---:|---:|
| Power-scope Rescale sum | `33890584, 34056291, 34124542, 34174915, 34196041, 34218709, 34431668` | 34.174915 ms | `31781000, 32489292, 32576625, 32611831, 32847457, 32855834, 32957166` | 32.611831 ms | −4.57% |
| `generated_powers` parent | `36243250, 36389167, 36519333, 36569000, 36578374, 36635459, 36855209` | 36.569000 ms | `34064000, 34858874, 34997417, 35004875, 35239875, 35249166, 35374209` | 35.004875 ms | −4.28% |

The post trace retained numerical replay agreement with diagnostics-off and decoded slots within the existing `1e-2` tolerance.

## Compiler evidence and remaining hotspot

Compiler listing was collected on `darwin/arm64` with `go test -c -gcflags='github.com/tuneinsight/lattigo/v6/schemes/ckks/fast=-S'`. In `mod128By64` the redundant leading `Div64` call is gone; the high-limb remainder is emitted as `UDIV` + `MSUB`, followed by one required `math/bits.Div64` call. `mod192By64` likewise has one fewer `Div64` call (two required lower-limb calls remain). For `roundedMagnitude192`, `/` and `%` already share a single `UDIV` + `MSUB`; the measured `bits.Div64` rewrite was slower, so no source change was warranted. `go tool objdump` is unavailable in this Go installation; the compiler `-S` listing supplied the observation above.

No new CPU profile was taken for this bounded optimization. The accepted DIAG-006 profile identified the fixed-width reconstruct/round/capacity phase as the largest isolated Rescale phase (43.51% baseline); it also showed `math/bits.Div64` (27.40% flat), `ring.inttLazyUnrolled16` (16.33%), and `ring.nttUnrolled16Lazy` (11.43%) among notable samples. This task reduced the fixed-width phase by 12.23%, but those older profile shares are not claimed as post-change attribution. Domain transforms and the remaining exact lower-limb divisions are therefore the evidence-backed residual hotspots; a fresh profile would be needed to rank them after this change.

## Classification

All correctness gates passed. Rows4 fixed-width phase improved **12.23%** (gate: 10%), authoritative full rows4 Rescale improved **6.75%** (gate: 5%), and diagnostics-off P93 Count-1 improved **6.11%** (gate: 2%). Allocation counts remained stable. Classification: **`FIXED_WIDTH_DIVISION_DEDUP_READY`**.
