# FAST-DUAL-LOGN16-EXACT-ZERO-EVIDENCE-AUDIT-003

**Audit classification: `EXACT_ZERO_INDEPENDENTLY_CONFIRMED`**
**Prior numerical task classification remains: `DUAL_BACKEND_NUMERICAL_PARTIAL`**
**Bootstrap calls during this audit: 0**

The LogN16 exact-zero comparisons are independently supported by the original Standard and Fast artifacts. Every stored decoded vector hash was recomputed from its own vector; all vectors have the canonical slot count and finite values. Every LogN16 Standard/Fast decoded value pair is bit-for-bit identical as IEEE-754 float64 real and imaginary components at all eight public replay checkpoints and at terminal output. Recomputed RMSE, max difference, and SNR match both the prior pair artifacts and the committed result summary. No prior result was rewritten.

## Provenance and independent collection

Both builds used Primary measurement source `590bcb89f831c582de0566472daca37d2e02cdc1` (measurement-source SHA-256 `1d33c100dfa77545334e6dc5fbff2b14d00132db73672dd48d013288e9919a3c`). The clean pinned Standard checkout is `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, built with `perf_standard`; the clean pinned Fast checkout is `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`, built with `perf_fast`. The two executable SHA-256 values are respectively `89a88a424769a73b542e85d5962e8fe337b327047cee8b524cc80dc8178e1d7c` and `092cbf1ef203372529a3fdfc5da968d0162c19eec772155324ae6557038ee358`. Each raw document's module path matches its separately pinned checkout. The original independent `.mod` files and executables were present and their hashes match the prior report.

For each profile, Standard/Fast preflight and terminal artifacts are distinct files with distinct whole-file SHA-256 values. Their provenance names different backend commits, build tags, module paths, and binary hashes. Each backend/profile attempt journal contains exactly `RESERVED` then `COMPLETED`, with one call recorded. Terminal decoded hashes match that same backend's S2C preflight hash. The different intermediate row hashes described below and the separate binary/run provenance rule out a same-file or compare-time vector alias as the cause of LogN16 equality.

The workload is still the prior diagnostic `diagnostic-encoded-c0-c1-zero-v1`: the same deterministically encoded message is placed in `c0`, with `c1=0`, for both implementations. Standard used its genuine pinned Standard evaluator binary, but this was **not** native RLWE E32 encryption and supports no security-equivalence claim.

## Independently recomputed checkpoints

The table lists each vector's recomputed fingerprint and decoded-domain comparison against the other backend. `+∞` is represented by the shared metric's `POSITIVE_INFINITY` status and a null `snr_db`; it is not a missing measurement. Hashes are over the decoded complex float64 vector, not ciphertext bytes.

### LogN13 — 4,096 slots

Config/input/metadata/QP SHA-256: `919a2d94…769c98` / `d9151964…dcc285` / `9171dac4…2764f6` / `1f045e60…f6a46b`. Effective profile matched: LogN=13, LogSlots=12, RingN=8192, 17 Q primes, 5 P primes, q0 target/actual 55/55 bits, Q-prefix width 4.

| Checkpoint | Standard decoded SHA-256 | Fast decoded SHA-256 | Bitwise equal | Complex RMSE | Max complex difference | SNR (dB) |
|---|---|---|---:|---:|---:|---:|
| input | `d107a20b…a4bdc5f` | `d107a20b…a4bdc5f` | yes | 0 | 0 | +∞ |
| ScaleDown | `d107a20b…a4bdc5f` | `d107a20b…a4bdc5f` | yes | 0 | 0 | +∞ |
| ModUp | `d107a20b…a4bdc5f` | `d107a20b…a4bdc5f` | yes | 0 | 0 | +∞ |
| C2S real | `24a8f8da…8d54745` | `c9cd0604…8b69ff` | no | 2.3964363961731316e-14 | 9.1850748469742e-14 | 175.43291024140183 |
| C2S imag | `239a95d5…fd1b329` | `c4e53745…166717` | no | 2.4095676473593568e-14 | 9.768350882409342e-14 | 175.35874442697173 |
| EvalMod real | `f410ac3d…2df9988` | `43cccb39…078027a` | no | 5.095135038373601e-11 | 1.60298267715131e-10 | 163.06651362820782 |
| EvalMod imag | `9ad6aa37…386baf03` | `e1879f5b…4a7dd2b2` | no | 5.135616506107663e-11 | 1.7541738200899104e-10 | 162.97107437864028 |
| S2C | `0af5cdd5…0479df0` | `b6e6c903…f717b755` | no | 1.4468089544345387e-10 | 4.675260603879638e-10 | 163.01897348253942 |

Terminal Standard-vs-Fast: RMSE `1.4468089544345387e-10`, max `4.675260603879638e-10`, SNR `163.01897348253942 dB`. Recomputed Standard-vs-input and Fast-vs-input RMSE/SNR also match the prior report exactly.

### LogN16 — 32,768 slots

Config/input/metadata/QP SHA-256: `16e62fde…d3033b02` / `659fc59c…0ce8ccf` / `58b69911…a6cf44f` / `6c82ca72…6710ec5f`. Effective profile matched: LogN=16, LogSlots=15, RingN=65536, 17 Q primes, 5 P primes, q0 target/actual 55/56 bits, Q-prefix width 4.

| Checkpoint | Standard decoded SHA-256 | Fast decoded SHA-256 | Bitwise equal | Complex RMSE | Max complex difference | SNR (dB) |
|---|---|---|---:|---:|---:|---:|
| input | `740e6f32…00f175b` | `740e6f32…00f175b` | yes | 0 | 0 | +∞ |
| ScaleDown | `740e6f32…00f175b` | `740e6f32…00f175b` | yes | 0 | 0 | +∞ |
| ModUp | `740e6f32…00f175b` | `740e6f32…00f175b` | yes | 0 | 0 | +∞ |
| C2S real | `024a890e…801c385a` | `024a890e…801c385a` | yes | 0 | 0 | +∞ |
| C2S imag | `1cb32983…d03a670fc` | `1cb32983…d03a670fc` | yes | 0 | 0 | +∞ |
| EvalMod real | `5876e1a6…8e8508bc` | `5876e1a6…8e8508bc` | yes | 0 | 0 | +∞ |
| EvalMod imag | `b3059002…39ab644ba` | `b3059002…39ab644ba` | yes | 0 | 0 | +∞ |
| S2C | `84f85bb7…8f8d825e` | `84f85bb7…8f8d825e` | yes | 0 | 0 | +∞ |

Terminal Standard-vs-Fast: RMSE `0`, max `0`, SNR `POSITIVE_INFINITY`. Both terminal vectors have SHA-256 `84f85bb7…8f8d825e`, equal to their own S2C preflight vector. Both backends' independently recomputed output-vs-input metrics are RMSE `5.808193349077692e-9`, max `4.7088049150256786e-7`, SNR `130.94764728940336 dB`; this is not the paired Standard-vs-Fast error.

## Physical ciphertext representation

Both backends have two ciphertext components. On LogN16 preflight, Standard physically stores all active rows at N=65,536 (17 rows at ModUp, 13 at C2S, 5 at EvalMod, and 2 at S2C); Fast materializes its Q-prefix rows (4 at ModUp through EvalMod and 2 at S2C), with inactive backing rows empty. Both report NTT=true. Standard Montgomery remains false; Fast Montgomery is true from ModUp through S2C.

Comparing the shared active q rows by row SHA-256: input and ScaleDown are 2/2 equal; ModUp, both C2S branches and both EvalMod branches each have 4/8 equal and 4/8 different; S2C has 2/4 equal and 2/4 different. Thus distinct intermediate ciphertext bytes coexist with identical decoded values under the recorded backend-specific Montgomery/Q-prefix representations. The independently collected terminal outputs have two components and two active rows each; their four row hashes are equal, and both terminal Montgomery flags are false. This mixed intermediate/terminal evidence is consistent with independent runs and does not rely on physical ciphertext equality as a proxy for decoded equality.

## Comparator integrity fix and validation

The old dual-current comparators accepted stored decoded vectors without recomputing `decoded_sha256`; terminal comparison also lacked a same-file/inode guard and trusted stored output-vs-input metrics. The shared comparator now fails closed on canonical profile/source identity, backend pin/build tag/module and executable identity, exact checkpoint names/slot counts/finite values, decoded-vector hashes, same-file terminal inputs, and stale terminal metrics. It emits explicit float64 bitwise-equality evidence. Regression tests reject wrong hashes, wrong backend source, wrong slot count, swapped/reused terminal sources, stale metrics, and distinguish IEEE-754 signed zero.

Passed:

- `GOCACHE=/private/tmp/fast-dual-audit-gocache.toi2n4 go test ./cmd/perfprobe -run '^TestDualCurrent' -count=1`
- `GOWORK=off GOCACHE=/private/tmp/fast-dual-audit-gocache.toi2n4 go test -modfile=/private/tmp/current-dual-numerical-002.MRXp6T/standard.mod -tags=perf_standard ./cmd/perfprobe -run '^TestDualCurrent' -count=1`
- `GOWORK=off GOCACHE=/private/tmp/fast-dual-audit-gocache.toi2n4 go test -modfile=/private/tmp/current-dual-numerical-002.MRXp6T/fast.mod -tags=perf_fast ./cmd/perfprobe -run '^TestDualCurrent' -count=1`
- `GOWORK=off GOCACHE=/private/tmp/fast-dual-audit-gocache.toi2n4 go test -modfile=/private/tmp/current-dual-numerical-002.MRXp6T/fast.mod -tags=fastdiag ./cmd/fastdiag -run '^$' -count=1` (compile-only)
- `GOCACHE=/private/tmp/fast-dual-audit-gocache.toi2n4 go vet ./cmd/perfprobe`
- `git diff --check`

The original four authorized Bootstrap calls are only evidenced by the preserved source journals; this audit invoked none. No Secondary source, CKKS arithmetic, parameters, thresholds, benchmark, or full test suite was changed or run. Fast private internal traces remain `NOT_COMPARABLE_TO_GENUINE_STANDARD`, so this audit confirms the exact-zero decoded evidence but does **not** change the previous task's `DUAL_BACKEND_NUMERICAL_PARTIAL` classification.
