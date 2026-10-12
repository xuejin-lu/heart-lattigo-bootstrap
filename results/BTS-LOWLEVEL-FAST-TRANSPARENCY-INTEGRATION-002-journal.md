# BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002 attempt journal

## Durable budget and reservation

- Task-wide maximum: two actual Fast `bts.Bootstrap` invocations; target: one No-Trace invocation.
- Reservation 1: **RESERVED, NOT LAUNCHED**. At reservation time, no Fast Bootstrap invocation has begun in this task.
- Reservation timestamp: 2026-10-11T23:53:02Z.
- This journal is task-scoped and must be reconciled before any future invocation; a new Codex session does not reset the budget.

## Frozen workload

- Original BTS `main.go` defaults; no flags: LogN=16, SlotLimbs=15, `BTSModePlain`, amp=1, seed=0.
- Original Q/P chain, input generation, scale, API call order, and Bootstrap implementation are unchanged.
- Run mode: No Trace, no warmup, no retry planned.
- Working directory: `external/BTS/`.
- Exact run command: `/private/tmp/bts-fast-002.d1SUFW/bts-fast`.

## Source and backend provenance before launch

- BTS Go source, `go.mod`, and `go.sum` SHA-256 manifest:

  | File | SHA-256 |
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

- Primary repository: `main`, HEAD/origin `9ffebd1ed0dc18767c8c26fbaffe9fbc5d22b119`, clean before artifacts.
- Fast dependency: `xuejin-lu/lattigo`, branch `fast-qprefix`, HEAD/origin `ac497c82adf60733ef652cfcd39acf741f6cb1d8`, clean.
- Temporary module file: `/private/tmp/bts-fast-002.d1SUFW/bts-fast.mod`, copied from the original BTS `go.mod` and given only a local replacement for `github.com/tuneinsight/lattigo/v6 => /Users/xuejinlu/Developer/xuejin-lu/lattigo`. `external/BTS/go.mod` and `go.sum` were not changed.
- Exact build command: `GOWORK=off go build -mod=mod -modfile=/private/tmp/bts-fast-002.d1SUFW/bts-fast.mod -o /private/tmp/bts-fast-002.d1SUFW/bts-fast .`
- Build provenance: Go `go1.26.4`, `darwin/arm64`; `go list -m` resolves Lattigo to `/Users/xuejinlu/Developer/xuejin-lu/lattigo`; `go version -m` records the same replacement path. Exact checked-out Lattigo commit is recorded above.
- Binary SHA-256: `482c854e23125085692aba3c302954177a7263a566b8f3daa245d92968ddbe0b`.
- Primary reports `git status --short -- external/BTS` empty; no BTS source or binary is staged or included in this task.

## Post-attempt repair and compile-only verification

- The observed API mismatch was that BTS calls ordinary `ckks.Evaluator.Conjugate`, which reached the promoted `rlwe.Evaluator.Automorphism` and then the standard `MemEvaluationKeySet.GetGaloisKey`; that getter correctly rejects `KeyLayoutFast`. The Fast CKKS evaluator had no public `Automorphism` override at this boundary.
- A bounded repair was added after the consumed attempt and pushed as Secondary commit `0b922dc48315128a27db811a5f237e0a6c9813d6`. Ordinary Fast CKKS `Automorphism` now validates zero c1, reuses the shared Q-prefix automorphism core for compact inputs, and preserves complete rows only after the existing centered-lift verification. It performs the actual Galois permutation and elides only the zero-input key switch.
- Non-Bootstrap focused regression tests, Fast package tests, bootstrapping API tests, `go vet`, and `git diff --check` pass for this repair. The full `go test ./schemes/ckks -count=1` remains non-green: it reports a Fast-mode stale degree-error assertion, a Standard-relinearization expectation against Fast keys, and the historical `RescaleTo/Single` empty-row panic. No unrelated repair was made.
- Compile-only rebuild against final Secondary HEAD `0b922dc48315128a27db811a5f237e0a6c9813d6` succeeded. This binary was not executed. Its SHA-256 is `72a4d9d38e9478d4d244de03c06bda1dc8fe60a8cacd43c4024f7a4acfb5167b`.

## Invocation state

- Process launched: **yes**, reservation 1; process exited with status 1.
- Entered `bts.Bootstrap`: **yes**. The program returned an error wrapped as `Bootstrap: C2S: Conjugate...`; this invocation is consumed and must not be repeated.
- `NewBTS()` initialization: 46.41 s; BTS parameter setup completed.
- Input: Level 3, Scale `2^32`, amp 1.000. Parameters report LogN16, MaxLevel L29, LogQP 1290, 32,768 slots; original output target L18.
- Bootstrap completion/time: **failed before completion**; the original program prints elapsed time only after a successful Bootstrap, so no Bootstrap duration or output ciphertext metadata exists.
- Exact error: `Bootstrap: C2S: Conjugate: cannot Conjugate: cannot apply Automorphism: unsupported: cannot use Fast evaluation key with Standard execution: key for galEl 131071 = 5^{16384} key is missing`.
- Complete console transcript: `results/BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002-attempt-01.log`. The launcher's relative `tee` destination was invalid from `external/BTS/`; the command session output was captured and persisted afterward. The initial `tee` diagnostic is retained in that transcript and is not a BTS error.
- Attempt 2: **unused and unavailable for retry of this run**, because the first invocation entered `bts.Bootstrap`. No further Bootstrap call is authorized by the task rule.
