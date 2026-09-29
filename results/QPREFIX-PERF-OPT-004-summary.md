# QPREFIX-PERF-OPT-004 — Exact Barrett-Horner Reduction

## Decision

`BARRETT_FIXED_WIDTH_REDUCTION_CORRECT_BUT_NO_WIN`

The exact Barrett-Horner candidates passed the equality/oracle tests, but the production-representative `mod192By64` at q3 improved only 9.8%, below the mandatory 15% feasibility gate. Neither signed-residue staging nor the complete Barrett-backed rows4 CRT reached the alternative 8% gate. All temporary candidate code was removed; no Secondary production change or commit was made.

## Provenance and unchanged implementation

- Primary task: `QPREFIX-PERF-OPT-004`; starting Primary HEAD `d00d4504d7bac9fa2f1f54f3c08e13f94fcaa083`.
- Secondary: `fast-qprefix`, HEAD `e3a7e04dc90f06128268510bcd31b08c2fbca176`, clean.
- Secondary HEAD only updates `CURRENT_TASK.md`; Rescale production source remains at accepted implementation commit `6930cf6cb3c71ce139a1eb42eede7be335b7174c`.
- No Secondary production code, parameters, APIs, or task schedules were changed. The test-only candidate file was deleted after feasibility evaluation.

## Candidate math and exactness

For `R64 = 2^64 mod q`, the test-only candidates used exact Horner reduction:

```text
mod128: (((hi mod q) * R64) + (lo mod q)) mod q
mod192: ((((hi mod q) * R64 + (mid mod q)) mod q) * R64 + (lo mod q)) mod q
```

Each digit reduction used `ring.BRedAdd`, each multiply used `ring.BRed`, and each sum used `ring.CRed`, with `R64` prepared outside timed loops. The complete test-only CRT candidate also replaced the q1 Garner digit reduction rather than calling the current `crtQ01`. Candidate hot helpers contained no `bits.Div64`. The supported-modulus domain was checked: `q < 2^63`; `BRed` operands were canonical residues in `[0,q)`; each addition stayed below `2q` without uint64 overflow, satisfying `CRed`'s input range.

The temporary exactness tests passed:

- `TestFastRescaleBarrettFixedWidthReductionsExact`: all LogN13 P93 Q moduli and generated LogN16 q0=56/q1=39 fixture; boundary values, near-multiples, max-width patterns, and 512 fixed-seed random values per modulus. Candidate `mod128`/`mod192` matched both existing helpers and `math/big.Int.Mod`; signed residues matched for positive and negative cases.
- `TestFastRescaleBarrettPreparedCRTQ0123Exact`: boundary Cartesian set plus 1,024 fixed-seed random rows4 residues; candidate matched current prepared CRT and a BigInt CRT oracle exactly.

## Feasibility benchmark

Command (diagnostics disabled; seven samples; `-benchtime=150ms`):

```bash
go test ./schemes/ckks/fast -run '^$' \
  -bench '^BenchmarkFastRescaleBarrett' \
  -benchmem -count=7 -benchtime=150ms
```

The final reported samples use the established OPT-002 Rescale operands, prepared outside timing: `q0*q1` reduced by q2, `q0*q1*q2` reduced by q3, negative signed-residue staging from the q0123 half-product for q0–q3, and fixed rows4 residues for prepared CRT. An initial exploratory run used synthetic modular-reduction operands and a CRT candidate that still called the existing q1 `crtQ01`; it was excluded after source review. The table below is the rerun with production-representative operands and complete q1–q3 Barrett replacement.

| Operation / modulus | Current samples (ns/op) | Barrett samples (ns/op) | Current median | Barrett median | Change | Allocs |
|---|---|---|---:|---:|---:|---:|
| mod128 / q2 | 2.522, 2.530, 3.453, 2.544, 2.832, 2.539, 3.050 | 3.193, 3.161, 3.159, 3.163, 3.171, 3.177, 3.168 | 2.544 | 3.168 | 24.5% slower | 0 B/op, 0 allocs/op |
| mod192 / q3 | 6.218, 6.160, 6.211, 6.241, 6.303, 6.226, 6.235 | 5.600, 5.622, 5.603, 5.618, 5.615, 5.869, 5.626 | 6.226 | 5.618 | 9.8% faster | 0 B/op, 0 allocs/op |
| signedResidue192 / q0 | 6.330, 6.369, 6.398, 6.405, 6.379, 6.333, 6.365 | 6.214, 6.255, 6.265, 6.230, 6.206, 6.233, 6.202 | 6.369 | 6.230 | 2.2% faster | 0 B/op, 0 allocs/op |
| signedResidue192 / q1 | 6.421, 6.404, 6.430, 6.462, 6.698, 6.415, 6.407 | 6.279, 6.295, 6.299, 6.275, 6.321, 6.412, 6.270 | 6.421 | 6.295 | 2.0% faster | 0 B/op, 0 allocs/op |
| signedResidue192 / q2 | 6.415, 6.421, 6.417, 6.430, 6.389, 6.415, 6.393 | 6.214, 6.249, 6.262, 6.223, 6.237, 6.218, 6.215 | 6.415 | 6.223 | 3.0% faster | 0 B/op, 0 allocs/op |
| signedResidue192 / q3 | 6.385, 6.393, 6.368, 6.390, 6.384, 6.397, 6.404 | 6.198, 6.213, 6.232, 6.633, 6.183, 6.209, 6.214 | 6.390 | 6.213 | 2.8% faster | 0 B/op, 0 allocs/op |
| prepared rows4 CRT (complete q1–q3 Barrett) | 34.03, 34.22, 33.51, 33.76, 34.96, 63.57, 48.76 | 33.27, 33.12, 35.75, 33.66, 35.03, 34.76, 42.33 | 34.22 | 34.76 | 1.6% slower | 0 B/op, 0 allocs/op |

The required gates are conjunctive: q3 `mod192` improved 9.8%, below 15%; signed-residue improved only 2.0–3.0%; and complete rows4 CRT was 1.6% slower instead of improving at least 8%. Therefore no candidate is retained.

## Production measurements and profile

The conditional authoritative before/after residue-staging, fixed-width phase, full rows4 Rescale, P93 E2E, low-overhead power trace, and post-change CPU profile were not run: the initial feasibility gate failed and no production implementation was kept. Accepted parent baseline evidence remains: full rows4 Rescale about `1.130242 ms`; fixed-width phase `36.55%`; signed-residue staging `21.56%`; `math/bits.Div64` largest flat symbol at `20.55%`.

Benchmark environment: `go1.26.4`, `darwin/arm64`, Apple M4.

The Barrett CRT candidate is locally faster in this isolated microbenchmark, but it cannot justify retaining a partial production rewrite while the required q3 `mod192` gate fails. No production hotspot changed; the existing Rescale reduction path remains the baseline.

## Validation and final state

- Both temporary Barrett exactness/oracle tests: passed before the candidate file was removed.
- `go test ./schemes/ckks/fast -count=1`: passed.
- Secondary `go test ./...`: passed.
- Secondary `git diff --check`: passed.
- Secondary branch `fast-qprefix`: clean at `e3a7e04dc90f06128268510bcd31b08c2fbca176`; no Secondary commit/push.
