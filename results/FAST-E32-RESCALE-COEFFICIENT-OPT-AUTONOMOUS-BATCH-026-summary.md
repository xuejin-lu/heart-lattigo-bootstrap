# Batch026 — E32 Rescale coefficient hotspot

- Classification: **`NO_BENEFICIAL_CANDIDATE`**
- Handoff: **`READY_FOR_WEB_REVIEW`**
- Budget used: **0 Bootstrap calls, 0 Standard performance calls, 0 P93 benchmarks**.
- Primary baseline: `78f8ab9daadf4522af9dd693f3d952ec9982e8c2` (`main`, clean at task start).
- Secondary baseline/final: `4f2557062cb5c1ffb9a671bc3df67401fd7092b4` (`fast-qprefix`, clean and equal to `origin/fast-qprefix`). No Secondary commit or net source change was retained.
- Original Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and original Fast production pin `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` were not changed.

## Fixture and provenance

Batch025's offline-revalidated E32 trace contains 539 events and 35 Rescale roots. The repeated `Level 9→8`, `rows 4→4`, in-place shape occurs 6 times and has both c0/c1 coefficient loops. The source raw remains labelled `TRACE_UNVERIFIED`; its SHA-256 is `f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74`. The E32 parameter/Q/P fingerprint is `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`.

The bounded microbenchmark used a deterministic fixed-seed, valid NTT-domain synthetic ciphertext at LogN13/E32, Level 9, degree 1, two nonzero components, four authoritative source rows, in-place Rescale, `IsMontgomery=false`. It exercised the real public Fast evaluator path into `RescaleWorkspace.ApplyRows`; reset/copy was outside the timed region. Input fingerprint: `001855dc4aba15d1da02e71df32242cb019c8976fa2ebede386bebad43ab1247`. This fixture is explicitly **not** the captured Bootstrap ciphertext: the trace supplies shape and parameter provenance, not its coefficient arrays.

Environment: Apple M4, macOS 26.6.2, Go 1.26.4 (`darwin/arm64`), `GOMAXPROCS=1`, `GOGC=100`, regular build (no diagnostic build tags). Command for both runs:

```sh
GOMAXPROCS=1 GOGC=100 go test ./circuits/ckks/bootstrapping \
  -run '^$' \
  -bench '^BenchmarkFastRescaleE32LogN13Level9Rows4InPlace$' \
  -benchmem -benchtime=100ms -count=5 -v
```

## Candidate and outcome

The single candidate added a guarded `Uint192.Hi == 0` path to `mod192By64`, replacing the top-limb modulo with exact `bits.Div64(0, Mid, modulus)` reduction. For the captured exact Q chain, `Q0123=5986308565615587353347023386369277282933624144412673` and Level-9 `q9=1152921504606601217`; the conservative rounded-magnitude upper bound `floor(Q0123/(2*q9))+1 = 2596147500804154870871593067152517 < 2^111`. Thus the output's high 64-bit limb is zero for this selected path. The branch was source-grounded and preserved the existing arithmetic contract.

The candidate passed its temporary big.Int differential test (zero and edge values plus 512 fixed-seed values, across seven moduli), the existing focused Rescale/Q-prefix tests, and focused `go vet`. However, its repeated measurement did not show a robust speedup:

| Source state | ns/op samples (five) | Median | Range | B/op | allocs/op |
|---|---|---:|---:|---:|---:|
| Baseline `4f255706` | 2,676,129; 2,663,219; 2,712,441; 2,692,458; 2,755,351 | 2,692,458 | 2,663,219–2,755,351 | 681–682 | 20 |
| Temporary candidate | 2,688,291; 2,734,854; 2,740,962; 2,709,171; 2,702,554 | 2,709,171 | 2,688,291–2,740,962 | 681–682 | 20 |

The candidate median was about **0.62% slower** and the sample ranges substantially overlap. It was removed; no further optimization candidate was tried. The uncommitted candidate source-file SHA-256 was `7dc49c53c27212dc05ee3916aef845ae3c734c0fe676ff6b32babf0e92498094`; candidate commit SHA is intentionally absent because the candidate was not retained. Secondary is clean at its unchanged baseline SHA.

## Validation

Passed on the temporary candidate state:

- `GOMAXPROCS=1 GOGC=100 go test ./schemes/ckks/internal/fastcore -count=1`
- `GOMAXPROCS=1 GOGC=100 go test ./schemes/ckks/fast -run 'Rescale|QPrefix' -count=1`
- `GOMAXPROCS=1 GOGC=100 go vet ./schemes/ckks/internal/fastcore ./schemes/ckks/fast ./circuits/ckks/bootstrapping`
- `git diff --check`

No Bootstrap, key-generation, decrypt, Standard timing, full CKKS benchmark, or P93 benchmark ran. The single failed initial benchmark setup was a fixture assertion that confused logical Level headers with actually backed compact rows; source inspection corrected the test-only assertion before any Rescale operation was timed.
