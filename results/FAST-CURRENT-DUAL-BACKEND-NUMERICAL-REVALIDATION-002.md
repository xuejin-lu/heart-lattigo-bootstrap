# FAST-CURRENT-DUAL-BACKEND-NUMERICAL-REVALIDATION-002

**Classification: `DUAL_BACKEND_NUMERICAL_PARTIAL`**

**Measurement status: `MEASURED`**

Both terminal Fast-vs-genuine-Standard comparisons and all eight public replay checkpoints per profile pass the `1e-7` RMSE gate. The result remains **PARTIAL**, not fully verified: the requested Fast generated-power/DoubleAngle internal checkpoints do not have matching private instrumentation in pristine original Standard. No Standard values were fabricated for those checkpoints.

## Source and binary provenance

- Primary source: `heart-lattigo-bootstrap`, `main` at `590bcb89f831c582de0566472daca37d2e02cdc1`, clean at build; measurement-source SHA-256 `1d33c100dfa77545334e6dc5fbff2b14d00132db73672dd48d013288e9919a3c`.
- Genuine Standard: clean detached Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. Independent `standard.mod` replaces the dependency with `/private/tmp/current-dual-numerical-002.MRXp6T/lattigo-standard` (modfile SHA-256 `48550d3b4e161be92fdce1fcdc1aad20488de6b8f2ddec7abd5fd9ae273f8dae`). The Standard source scan found no `FastCKKSZeroSecretSimulation` symbol. Binary `perfprobe-standard-590bcb8`, tag `perf_standard`, SHA-256 `89a88a424769a73b542e85d5962e8fe337b327047cee8b524cc80dc8178e1d7c`.
- Fast: clean detached Lattigo `fast-qprefix` pin `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`. Independent `fast.mod` replaces the dependency with `/private/tmp/current-dual-numerical-002.MRXp6T/lattigo-fast` (modfile SHA-256 `a3e8b6c80ad886772a6e05dd3c024f21a1a75a0d524ce2b78c6b223b4cb29edd`). Binary `perfprobe-fast-590bcb8`, tag `perf_fast`, SHA-256 `092cbf1ef203372529a3fdfc5da968d0162c19eec772155324ae6557038ee358`; it uses the ordinary public `bootstrapping.NewEvaluator` wrapper and the pinned Fast library's internal dispatch.
- Fast internal trace binary: `fastdiag-fast-590bcb8`, tag `fastdiag`, SHA-256 `0ee1e36fcbcbf9a3aa41eef403e87a012d463f1e0ad86694796c9eba79f0d7ed`.
- Both `perfprobe` binaries report `vcs.revision=590bcb89f831c582de0566472daca37d2e02cdc1` and `vcs.modified=false`. Builds used the same Primary source, separate temporary modfiles, and `GOWORK=off`; committed Primary `go.mod` was not changed. Environment: Go `1.26.4`, `darwin/arm64`.

## Workload and call budget

The two backends received the same deterministic original message and effective Q/P profile within each profile. This is the explicitly authorized diagnostic `diagnostic-encoded-c0-c1-zero-v1`: `ckks.Encode` at residual Level 0, encoded plaintext in `c0`, `c1=0`. It is **not** the formal public E32 `EncryptNew` lifecycle and supports no security-equivalence claim.

| Profile | Slots / ring N | Q / P primes | q0 target / actual bits | Config SHA-256 | Input SHA-256 | Q/P SHA-256 |
|---|---:|---:|---:|---|---|---|
| LogN13, LogSlots12 | 4,096 / 8,192 | 17 / 5 | 55 / 55 | `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` | `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` | `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b` |
| LogN16, LogSlots15 | 32,768 / 65,536 | 17 / 5 | 55 / 56 | `16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02` | `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf` | `6c82ca724cbfdce2083b2ef2c18d53f81e658a6f28ff283e9a671d276710ec5f` |

Each profile consumed exactly one genuine Standard and one Fast public Bootstrap call. Four fresh calls were reserved and completed (`4/4`); no retries, warmups, or performance runs occurred. The earlier three invalid-comparator LogN16 calls remain charged to the earlier task and were not reused.

## Public stage replay comparison

These are separate public-stage replays **without Bootstrap**, not in-circuit snapshots. At every checkpoint, Level, Scale, Degree, and dimensions matched. NTT was true on both sides. The backend-specific representation was recorded and decoded through its own adapter: both Montgomery flags were false at input/ScaleDown; from ModUp onward Standard was false and Fast was true. Prefix row counts are Standard/Fast.

### LogN13

| Checkpoint | Level / Scale | Rows S/F | Complex RMSE | Max complex difference | SNR (dB) |
|---|---:|---:|---:|---:|---:|
| input | 0 / `2^45` | 1 / 1 | 0 | 0 | +∞ |
| ScaleDown | 0 / `2^45` | 1 / 1 | 0 | 0 | +∞ |
| ModUp | 16 / `2^50` | 17 / 4 | 0 | 0 | +∞ |
| C2S real | 12 / `2^50` | 13 / 4 | 2.3964363961731316e-14 | 9.1850748469742e-14 | 175.43291024140183 |
| C2S imag | 12 / `2^50` | 13 / 4 | 2.4095676473593568e-14 | 9.768350882409342e-14 | 175.35874442697173 |
| EvalMod real | 4 / `2^45` | 5 / 4 | 5.095135038373601e-11 | 1.60298267715131e-10 | 163.06651362820782 |
| EvalMod imag | 4 / `2^45` | 5 / 4 | 5.135616506107663e-11 | 1.7541738200899104e-10 | 162.97107437864028 |
| S2C | 1 / `2^45` | 2 / 2 | 1.4468089544345387e-10 | 4.675260603879638e-10 | 163.01897348253942 |

### LogN16

| Checkpoint | Level / Scale | Rows S/F | Complex RMSE | Max complex difference | SNR (dB) |
|---|---:|---:|---:|---:|---:|
| input | 0 / `2^45` | 1 / 1 | 0 | 0 | +∞ |
| ScaleDown | 0 / `2^45` | 1 / 1 | 0 | 0 | +∞ |
| ModUp | 16 / `2^50` | 17 / 4 | 0 | 0 | +∞ |
| C2S real | 12 / `2^50` | 13 / 4 | 0 | 0 | +∞ |
| C2S imag | 12 / `2^50` | 13 / 4 | 0 | 0 | +∞ |
| EvalMod real | 4 / `2^45` | 5 / 4 | 0 | 0 | +∞ |
| EvalMod imag | 4 / `2^45` | 5 / 4 | 0 | 0 | +∞ |
| S2C | 1 / `2^45` | 2 / 2 | 0 | 0 | +∞ |

## Terminal public Bootstrap results

| Profile | Fast vs Standard RMSE / max / SNR | Standard vs input RMSE / max / SNR | Fast vs input RMSE / max / SNR |
|---|---|---|---|
| LogN13 | `1.4468089544345387e-10` / `4.675260603879638e-10` / `163.01897348253942 dB` | `1.1903936621664996e-9` / `5.805495741185792e-8` / `144.71338512805514 dB` | `1.1907546782784477e-9` / `5.805793207821362e-8` / `144.7107513180764 dB` |
| LogN16 | `0` / `0` / `+∞` | `5.808193349077692e-9` / `4.7088049150256786e-7` / `130.94764728940336 dB` | `5.808193349077692e-9` / `4.7088049150256786e-7` / `130.94764728940336 dB` |

All listed RMSE values pass the `1e-7` RMSE gate. LogN16 output-vs-input max complex difference is reported separately and is not the gate; each backend has 0 real and 2 imaginary components above `1e-7`. Paired Fast-vs-Standard comparison has no components above that threshold.

## LogN16 genuine Standard full-Q capacity audit

The independent Standard zero-call preflight passed the previously failing C2S path: ModUp at Level 16 had 17 active `q0..q16` rows per component, each length 65,536. After group-0 Rescale, Level 15 retained all 16 active `q0..q15` rows per component, each length 65,536; the next C2S matrix succeeded with those rows. NTT was true and Montgomery false throughout the Standard checkpoints. This confirms the earlier `len(p1)=0` panic was caused by the non-genuine Standard comparator build, not by the pinned original Standard C2S path.

## Fast-only internal trace and limitations

For each profile, `cmd/fastdiag numerical-trace-only` produced real and imag traces with 20 checkpoints per branch, generated powers T1/T2/T3/T4/T6/T8/T16, three DoubleAngle rounds, and before/after public scale reset. Each branch recorded 73 Q-prefix capacity observations; all 146 satisfied strict `2B < S_Q`. Fast internal replay matched public EvalMod with RMSE 0. Both branches are explicitly `NOT_COMPARABLE_TO_GENUINE_STANDARD`; pristine Standard has no corresponding Fast private hooks. The traces are Fast-only attribution, not paired internal numerical evidence.

The previous `FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001` blocked artifact remains unchanged. It is not used as a new measurement; this run used an independently pinned Standard binary and passed the full-Q zero-call audit.

## Validation

Passed:

- `GOWORK=off GOCACHE=/private/tmp/current-dual-numerical-002.MRXp6T/gocache-primary go test ./cmd/perfprobe ./cmd/fastdiag -count=1`
- `GOWORK=off GOCACHE=/private/tmp/current-dual-numerical-002.MRXp6T/gocache-standard go test -modfile=/private/tmp/current-dual-numerical-002.MRXp6T/standard.mod -tags=perf_standard ./cmd/perfprobe -run '^TestDualCurrent' -count=1`
- `GOWORK=off GOCACHE=/private/tmp/current-dual-numerical-002.MRXp6T/gocache-fast go test -modfile=/private/tmp/current-dual-numerical-002.MRXp6T/fast.mod -tags=perf_fast ./cmd/perfprobe -run '^TestDualCurrent' -count=1`
- `GOWORK=off GOCACHE=/private/tmp/current-dual-numerical-002.MRXp6T/gocache-fast go test -modfile=/private/tmp/current-dual-numerical-002.MRXp6T/fast.mod -tags=fastdiag ./cmd/fastdiag -run '^$' -count=1` (compile-only)
- `GOWORK=off GOCACHE=/private/tmp/current-dual-numerical-002.MRXp6T/gocache-primary go vet ./cmd/perfprobe ./cmd/fastdiag`

`go test ./...` was intentionally not run because this task explicitly forbids accidental tests that may invoke cryptographic Bootstrap beyond its four-call budget. No Secondary source or production arithmetic was changed.
