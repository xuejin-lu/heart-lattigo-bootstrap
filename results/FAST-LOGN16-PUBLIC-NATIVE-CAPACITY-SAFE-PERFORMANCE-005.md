# FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005

**Classification:** `CURRENT_LOGN16_FAST_SPEED_VERIFIED`
**Handoff:** `READY_FOR_WEB_REVIEW`

## Result and scope

The original, unmodified LogN16/E32 public workload passed the zero-Bootstrap native oracle gate on both genuine Standard and current Fast, then completed the authorized cold/warm measurements. All 16 transferred calls were consumed exactly: LogN16 Standard 6, LogN16 Fast 6, LogN13 Standard 2, LogN13 Fast 2. Task 004 remains at 0/16. There were no implicit warmups, retries, or additional Bootstrap calls.

The classification is limited to the exact 32,768-slot input, profile, pinned implementations, and observed public CKKS execution below. It is not a universal capacity proof, a guarantee for arbitrary CNN activations, or a claim of security/noise equivalence.

## Provenance

| Item | Value |
|---|---|
| Primary measurement source commit | `edf68d4a161207cb43b66ad84970b248ef9c31e8` |
| Primary measurement source SHA-256 | `6f42022d458c646e01e27d17559037b63dfcecc861ad05d5a5ed3e8986b8e597` |
| Genuine Standard Lattigo | `5dbffbdea05394de2ca3a432ed5318aa832e3f40` |
| Fast Lattigo | `fast-qprefix@2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` |
| Environment | Go 1.26.4, darwin/arm64, Apple M4, GOMAXPROCS=10 |
| E / input slots / ring degree | 32 / 32,768 / 65,536 |
| Default scale / q0 / q5 | 2^45 / `36028797019488257` / `1152921504606584833` |
| Config / input / workload / QP SHA-256 | `16e62fdea2dc5ac32240a33cc256cdf60b935aa00b4a16aa004474d2c3033b02` / `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf` / `1413c0f0edc3c6b52d7a71e4714efe70280854e1208f87640b741682b65fa2a9` / `6c82ca724cbfdce2083b2ef2c18d53f81e658a6f28ff283e9a671d276710ec5f` |

Standard used native generated-secret key/evaluation-key, encryption, decryption, and decoding. Fast used its intentionally insecure zero-secret mode (`c1=0`). Both builds used the same Primary source and public operation sequence. No application-side `NewFastEvaluator`, manually forced Standard `c1`, or input/parameter adjustment was used.

## Capacity interpretation

The existing worst-case coefficient calculation remains unchanged. For each input, `encodedInputBoundExact` uses `ceil(2*Delta*max|slot|)+1`; addition sums component bounds; the negacyclic product bound is `N*B_add*B_C`; nearest-integer rescale uses `floor((B_mul+floor((q5-1)/2))/q5`. These are conservative bounds for the ideal encoded polynomial, not ciphertext residue magnitudes, encryption/key-switch noise, or arbitrary application activations.

For LogN16, `B=18169085704944493`, `q0=36028797019488257`, and `2B-q0=309374390400729`. Thus the terminal sufficient condition `2B<q0` remains **`NOT_PROVEN`**; the strict historical helper still reports the same deficit. Q0123 capacity passes with product `5986249335185303576695453761753997162624003383361537`.

This failed sufficient condition is not an observed overflow. The separate, versioned `logn16-original-e32-observed-native-oracle-v1` gate passed for both implementations: all eight full-slot decoded checkpoints were within `1e-6`, and each backend's decoded Level4-to-Level0 continuity difference was exactly zero. This is evidence for this fixed workload only; inspecting centered residues does not prove absence of an earlier wrap.

For context, the original slot maxima are `|A|=0.0333750146301466`, `|B|=0.025648067192177354`, `|A+B|=0.05899745187088121`, and `|(A+B)C|=0.0019690408193320306` (`C=A`). The intuitive normalized centered capacity is `q0/(2*2^45)=512.0000000074506`; the original maximum `|A|` is about 0.00652% of that number. A slot magnitude of 10 would be about 1.953% of it. These ratios are not packed-coefficient bounds and do not establish safety for intermediate CNN activations or ciphertext noise.

### Zero-call paired checkpoints

Each row is Fast-vs-Standard decoded-vector RMSE and maximum complex difference. All rows passed the fixed `1e-6` maximum-complex gate; Level, Scale and Degree matched. Per-backend plaintext-oracle evidence and full checkpoint metadata are retained in the paired and run artifacts.

| Profile | Checkpoint | Level | log2(Scale) | Complex RMSE | Max complex difference |
|---|---|---:|---:|---:|---:|
| LogN16 | encrypt A | 5 | 45 | 2.337626420293912e-11 | 7.836700106908352e-11 |
| LogN16 | encrypt B | 5 | 45 | 2.3311935853066038e-11 | 7.041871757884207e-11 |
| LogN16 | encrypt C at q5 scale | 5 | 59.99999999999967 | 7.1128995011441e-16 | 2.3188667469835047e-15 |
| LogN16 | Add | 5 | 45 | 3.299433636582169e-11 | 1.093859727870019e-10 |
| LogN16 | MulRelin | 5 | 104.99999999999967 | 6.749091727584815e-13 | 3.224580292455998e-12 |
| LogN16 | Rescale(q5) | 4 | 45 | 2.930825398685596e-11 | 2.0763858711674715e-10 |
| LogN16 | Rotate | 4 | 45 | 4.1293643314233626e-11 | 2.0877707162721593e-10 |
| LogN16 | DropLevel to 0 | 0 | 45 | 4.1293643314233626e-11 | 2.0877707162721593e-10 |
| LogN13 control | encrypt A | 5 | 45 | 8.227532663160243e-12 | 2.4995830004055858e-11 |
| LogN13 control | encrypt B | 5 | 45 | 8.378525221734112e-12 | 2.486338441723047e-11 |
| LogN13 control | encrypt C at q5 scale | 5 | 59.99999999999998 | 2.5027158474574725e-16 | 6.933427371662636e-16 |
| LogN13 control | Add | 5 | 45 | 1.1794153356362292e-11 | 3.8432977788610515e-11 |
| LogN13 control | MulRelin | 5 | 104.99999999999999 | 2.418019943274425e-13 | 1.0239980760944924e-12 |
| LogN13 control | Rescale(q5) | 4 | 45 | 1.0246804763161703e-11 | 4.809671168665812e-11 |
| LogN13 control | Rotate | 4 | 45 | 1.4374699262726383e-11 | 5.3689325342406835e-11 |
| LogN13 control | DropLevel to 0 | 0 | 45 | 1.4374699262726383e-11 | 5.3689325342406835e-11 |

## Bootstrap performance and decoded output

All Standard and Fast self-oracles passed. Every paired cold/warm output had matched Level 1, Scale 2^45 and passed the `1e-6` gate.

| Profile/backend | Cold ms | Warm samples ms | Warm median ms | Warm median allocated bytes / allocations |
|---|---:|---|---:|---:|
| LogN16 Standard | 24079.744625 | 6425.275750, 5728.371333, 5957.288417, 26056.046958, 11325.118666 | 6425.275750 | 850,543,880 / 40,690 |
| LogN16 Fast | 8093.325666 | 680.102125, 663.878208, 659.665625, 662.523500, 668.751709 | 663.878208 | 57,063,448 / 8,429 |
| LogN13 Standard | 356.916958 | 341.894791 | 341.894791 | 92,213,336 / 35,047 |
| LogN13 Fast | 622.535167 | 73.822375 | 73.822375 | 7,505,568 / 8,093 |

LogN16 Fast's warm median was 9.6784x faster in this run (89.67% lower measured time); its five warm samples were much tighter than Standard's 5.73–26.06 s range. This is still a bounded run, not a broad machine-level performance guarantee. Fast's cold Bootstrap allocated 11,053,864,048 bytes and 225,945,923 allocations; the first call includes lazy circuit initialization. Standard key/evaluator construction is measured separately in each backend JSON. Do not compare constructor time alone or fold lazy setup into a warm-call claim.

| Paired output | Samples | Complex RMSE | Max complex difference | Level / log2(Scale) |
|---|---:|---:|---:|---|
| LogN16 cold + five warm | 6 | 4.09656347139061e-8 | 4.345682976228596e-7 | 1 / 45 |
| LogN13 cold + warm | 2 | 4.901501733076133e-9 | 4.6260892998958456e-8 | 1 / 45 |

The LogN13 control retains its strict sufficient-bound gate. Its single Fast cold sample was slower than Standard, while the one warm sample was 4.6313x faster; treat the two-call control as descriptive only.

## Fast path evidence and limits

The Primary harness called the ordinary public `ckks.NewEvaluator` and `bootstrapping.NewEvaluator`, then public `AddNew`, `MulRelinNew`, `Rescale`, `RotateNew`, `DropLevelNew`, and `Bootstrap`. The Standard adapter verified a genuine Standard evaluator; the Fast adapter verified ordinary-wrapper Fast dispatch for the pinned zero-secret capability.

At Fast commit `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`, source dispatch is visible in `schemes/ckks/evaluator.go`: public Add routes to `addSubFastCKKSZeroSecret` / `fastcore.AddSubCore.ApplyRows`; public MulRelin routes to `mulRelinFastCKKSZeroSecret` / `fastcore.MulCore.ApplyRows`; Rescale calls `fastcore.RescaleCore.ApplyRows`; Rotate routes to `rotateFastCKKSZeroSecret` / `fastcore.AutomorphismCore.ApplyRows`. These paths use the fixed Q-prefix policy and do not read dormant rows.

LogN16 runtime metadata independently showed the expected physical backing: native `EncryptNew` begins with full rows; after Add and MulRelin only q0–q3 are backed; after Level4 Rescale and Rotate q0–q3 are backed while q4 is dormant; after DropLevel to Level0 only q0 remains. The Fast Bootstrap output at Level1 has two Q-prefix rows. This evidence is independent of the LogN13-specific C2S compression path.

No Fast in-circuit C2S/EvalMod/S2C event spans were collected; the 16-call budget is exhausted. This report claims public end-to-end timing and source/runtime Q-prefix dispatch evidence, not internal stage attribution. Fast's zero-secret semantics are intentionally insecure and are not security- or noise-equivalent to Standard's genuine secret-key lifecycle.

## Artifacts and validation

The compact task summary is this file plus [`FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.json`](FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.json). The attempt ledger, paired preflight/bootstrap documents, per-backend run JSON, and all 16 reservation journals are committed under `results/`. Large decoded vectors and temporary modfiles remain outside the repository.

The platform was reused throughout: `cmd/perfprobe`, `internal/perfmeasure`, and `internal/numericalmetrics`; no one-off measurement framework or Secondary production change was introduced. Final validation passed:

- Default: `go test ./cmd/perfprobe ./internal/perfmeasure -count=1` and `go vet ./cmd/perfprobe ./internal/perfmeasure`.
- Genuine Standard pin: tagged `perf_standard` focused tests and vet passed using its isolated modfile.
- Fast pin: tagged `perf_fast` focused tests and vet passed using its isolated modfile.
- `jq empty` on the task summary/ledger JSON and `git diff --check` passed.

The first default test invocation could not write the system Go build cache (`operation not permitted`); the same command passed after setting `GOCACHE` to the task-owned temporary directory. No test used `ephemeral_diag`, and no test-triggered Bootstrap call was made.
