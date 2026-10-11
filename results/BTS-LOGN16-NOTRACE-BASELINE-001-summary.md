# BTS LogN16 No-Trace Standard/Fast baseline — NEEDS_WEB_REVIEW

Date: 2026-10-12

Primary source commit before this task: `8b4a40ff951f458eed4ab9add2a9797bc93033b9`

Task: `BTS-LOGN16-NOTRACE-BASELINE-001`

No `fastdiag`, trace build tags, warmups, repetitions, Rescale benchmark, or extra Bootstrap calls were used.

## Outcome

The unchanged BTS program completed one full Standard Bootstrap and decoded its output. The one Fast Bootstrap invocation panicked during the BTS CoeffsToSlots path before producing an output. Therefore the experiment has a valid Standard result and a captured Fast failure, but **no paired numerical acceptance and no Standard/Fast speed ratio**. The Fast failure is a backend/API representation compatibility issue requiring Web review; no BTS or Secondary production source was changed.

## Fixed workload and provenance

The exact same `external/BTS` module and `main.go` defaults were used in both builds, with no arguments: LogN16, SlotLimbs=15, `BTSModePlain`, amp=1, seed=0, 32,768 slots, `MaxLevel=L29`, `LogQP=1290.0`, `S2CStart=L3`, expected BTS output L18. The original main creates the secret key, encoder, encryptor, and decryptor; times `NewBTS`; encrypts the deterministic input; times its single `bts.Bootstrap`; then decrypts/decodes and calculates the errors. No frontend/backend conditional code or manual ciphertext component edits were introduced.

`external/BTS` is locally excluded from Git. Its original `.go`, `go.mod`, `go.sum`, and `run.sh` files were hash-checked before and after both executions; the hashes match. The pre-existing 11 MB `external/BTS/btstest` file was not used or modified. The source-manifest SHA-256 below is over the sorted per-file SHA-256 manifest for the listed Go/module/run files (excluding `btstest`).

| BTS source file | SHA-256 |
|---|---|
| `bts.go` | `02a8f803802d74a9cc6ca42e4bcac65d32af7f71373885036b52b27f4a207bdf` |
| `evalmod.go` | `56fb499dee861316e8e417624b02ceed3cac299b0a1378127ae7e887b6f69226` |
| `main.go` | `1f0f9a550b67147c58f096f3c2e9b92cb1ada5463b07d0238ac5c147656fd402` |
| `modup.go` | `50b248defe01e8dd013a32fec754d7f1ec201c0cd1fd9df14138a6b489b8b104` |
| `my_arsin.go` | `8d4cdbfcb1abeb950749d19509aa0083a4dc0d2d0d441bb242cd4503b6deb4be` |
| `my_cosine.go` | `5a6809dd2aeb9ca1f82b3c660afd2f9ce98f70bfa70369e3668e7406da074a89` |
| `my_evalcos.go` | `2f4445a6d20e0adb24fdbbc680d91518625830db0250383fc71a0cf1370c7c8f` |
| `my_relu.go` | `fc9e65ddac80f0519750243303c3893f97e1d6284918d5bb45852535c07519d7` |
| `myc2s.go` | `3879c8244c036ceb5c48790f0887b7b1655cc556a5e914b0319818e805738402` |
| `mys2c.go` | `1fd61343d3edb055c32451a463d109b9514dd3c1a1c175d0b48d906b13a78eae` |
| `go.mod` | `80a5ca5b071ed57233fa3262afa880a882ca21f044a46d07c02293954f465cbb` |
| `go.sum` | `ba0b56d4f4d13fd01648402ad271b0c27a6722830ca2f652fcd94e4cfdcb508d` |
| `run.sh` | `3ca27493d303b8acc16c3a8153a39d8065de7621289dfc89a1b659c491e92bb2` |

Source-manifest SHA-256: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`.

| Backend | Exact linked source | Checkout status | Binary SHA-256 |
|---|---|---|---|
| Standard | Original Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Git tree `ba959dd386565d13bf1147a847e5dfe2c0361f90`; archived source tar SHA-256 `5cff4c901bca6cfd4ab873d37ebc239fcf18bc71541db3d521b6bbd6d36060cd` | Immutable source snapshot from that commit | `ad5bfdcec4ab380ebe824459ea834aa911d9effb555c38dce58cbd80b26e3823` |
| Fast | `fast-qprefix` `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` | Secondary branch checkout clean at the pinned commit | `dbd62f0dc9b275f11211ac6fa744ba50a89dfe0309545b995152d7e58a9dbd53` |

Both `go version -m` records show `github.com/tuneinsight/lattigo/v6 v6.2.0` replaced by the backend-specific source path above. Builds used separate temporary modfiles copied from the unchanged BTS `go.mod`/`go.sum`, with backend-specific replacements and `GOWORK=off`; Go's `-mod=mod` upgraded the temporary `go` directive from 1.24.0 to 1.25.0 to satisfy the dependency. The original BTS `go.mod` and `go.sum` were not changed. No build tags were specified. The first cache attempt hit the sandbox's unwritable default Go build cache; successful builds used a task-local cache under `/private/tmp`.

Environment: Go `1.26.4`, `darwin/arm64`, Darwin kernel `25.6.0`, `CGO_ENABLED=1`. The BTS source and both backend checkouts remained unchanged; the live Secondary remained clean at `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`.

## No-Trace results

| Measurement | Standard | Fast |
|---|---:|---:|
| `NewBTS` initialization | `44.316 s` | `1m0.538 s` |
| First `bts.Bootstrap` | `2m19.653 s`, completed | Panic; no elapsed line was emitted |
| Input Level / Scale | `L3 / 2^32` | `L3 / 2^32` |
| Output Level / Scale | `L18 / 2^32` | No output |
| Max absolute error | `1.708e-03` | N/A — no decoded output |
| Mean absolute error | `1.172e-04` | N/A — no decoded output |
| Max-error slot | `7606` | N/A — no decoded output |
| Bootstrap completed and decoded | Yes | No |

The ratio is withheld: Fast did not produce output, and the task specified no numerical acceptance tolerance with which to call a paired result acceptable. These are single observations, not a repeated or steady-state benchmark.

## Fast failure location

Exact panic from the captured run:

```text
panic: cannot inttCoreLazy: ensure that len(p1)=0, len(p2)=65536 and len(roots)=65536 >= N=65536
```

The stack identifies `main.(*BTS).Bootstrap` → `MyC2S_DHManual` at `external/BTS/myc2s.go:282`, which calls `ckksEval.DecomposeNTT(..., zV.Value[1], ...)`. In the Fast-linked binary, `rlwe.Evaluator.DecomposeNTT` reaches `ring.Ring.INTT` (`ring/ntt.go:478` in the pinned checkout), where `ring.inttCoreLazy` panics because its first operand has length zero. The failing stage is C2S; EvalMod and all later BTS stages were not reached. This identifies the exact failing API/call path, but this run does not further distinguish the underlying representation precondition from a backend defect. Do not modify BTS or Secondary based on this result without Web review.

## Commands and raw logs

Successful builds (task-local temporary paths):

```sh
GOWORK=off GOCACHE=/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/go-cache go build -mod=mod -buildvcs=false -modfile=/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/standard.mod -o /private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-standard .
GOWORK=off GOCACHE=/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/go-cache go build -mod=mod -buildvcs=false -modfile=/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/fast.mod -o /private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-fast .
```

The successful binaries were run with no arguments:

```sh
set -o pipefail; '/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-standard' 2>&1 | tee 'results/BTS-LOGN16-NOTRACE-STANDARD-001-resume-2.log'
test -x '/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-fast' && set -o pipefail; '/private/tmp/bts-logn16-notrace-001-resume.CT7bjX/bts-fast' 2>&1 | tee 'results/BTS-LOGN16-NOTRACE-FAST-001.log'
```

Raw logs are preserved (and contain the stock BTS console output, not profiler traces):

- [Standard partial pre-Bootstrap launch](BTS-LOGN16-NOTRACE-STANDARD-001.log), 86 bytes; SHA-256 `d910d1ebc16aec266dff07175a76fb08fd2f860171b6f6afbc2e16b9d7f53aab`.
- [Standard completed run](BTS-LOGN16-NOTRACE-STANDARD-001-resume-2.log), 1,784 bytes; SHA-256 `440524a8789cf3e6987d1bd6e66556e74cd9ccd093183742968b2cf0e90c9431`.
- [Fast failed run](BTS-LOGN16-NOTRACE-FAST-001.log), 3,592 bytes; SHA-256 `433f0b13764785d09cadd5cdaaaf31ddb7add41a735614970b2bf9b2baee2ef7`.
- [Execution journal](BTS-LOGN16-NOTRACE-BASELINE-001-journal.md) records the per-backend call budget and resume checkpoint.

No CKKS/BTS test suite, Bootstrap warmup, trace capture, or additional diagnostic was run. The AGENTS-mandated worktree-policy unit tests passed 2/2 and invoke no Bootstrap. Final status: `NEEDS_WEB_REVIEW` because Fast failed before producing a numerical output; the one-call-per-backend budget is exhausted.
