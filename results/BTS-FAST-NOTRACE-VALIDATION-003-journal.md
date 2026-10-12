# BTS-FAST-NOTRACE-VALIDATION-003 attempt journal

## Durable invocation budget

- Task-wide maximum: **one** actual `bts.Bootstrap` invocation.
- Attempt 01: **RESERVED, NOT LAUNCHED**. At reservation time, actual invocations = 0; Bootstrap has not been entered.
- Exclusive reservation timestamp: `2026-10-12T02:36:10Z`.
- Do not retry if the process starts and then fails, panics, or is interrupted. This journal is independent of BTS task 002; task 002's journal and consumed attempt remain unchanged.

## Frozen source and workload

- Original module: `external/BTS/`; source, `go.mod`, and `go.sum` are unchanged and must not be staged or published.
- Defaults: LogN16, SlotLimbs=15, `BTSModePlain`, amp=1, seed=0; original Q/P chain, scale, input generation, and operation order.
- Input uses the original secret-key Encryptor. No warmup, Trace, benchmark, Standard rerun, or extra Bootstrap call.
- Source manifest SHA-256:

  | File | SHA-256 |
  |---|---|
  | `bts.go` | `02a8f803802d74a9cc6ca42e4bcac65d32af7f71373885036b52b27f4a207bdf` |
  | `evalmod.go` | `56fb499dee861316e8e417624b02ceed3cac299b0a1378127ae7e887b6f69226` |
  | `go.mod` | `80a5ca5b071ed57233fa3262afa880a882ca21f044a46d07c02293954f465cbb` |
  | `go.sum` | `ba0b56d4f4d13fd01648402ad271b0c27a6722830ca2f652fcd94e4cfdcb508d` |
  | `main.go` | `1f0f9a550b67147c58f096f3c2e9b92cb1ada5463b07d0238ac5c147656fd402` |
  | `modup.go` | `50b248defe01e8dd013a32fec754d7f1ec201c0cd1fd9df14138a6b489b8b104` |
  | `my_arsin.go` | `8d4cdbfcb1abeb950749d19509aa0083a4dc0d2d0d441bb242cd4503b6deb4be` |
  | `my_cosine.go` | `5a6809dd2aeb9ca1f82b3c660afd2f9ce98f70bfa70369e3668e7406da074a89` |
  | `my_evalcos.go` | `2f4445a6d20e0adb24fdbbc680d91518625830db0250383fc71a0cf1370c7c8f` |
  | `my_relu.go` | `fc9e65ddac80f0519750243303c3893f97e1d6284918d5bb45852535c07519d7` |
  | `myc2s.go` | `3879c8244c036ceb5c48790f0887b7b1655cc556a5e914b0319818e805738402` |
  | `mys2c.go` | `1fd61343d3edb055c32451a463d109b9514dd3c1a1c175d0b48d906b13a78eae` |

## Build and backend provenance

- Primary: `main`, clean at build time, commit `8ee8f7392f17a4d03b4c218f8a2c5b2dc85e38da`.
- Fast Secondary: `fast-qprefix`, clean independent checkout, local and `origin/fast-qprefix` both `049ffa782729d049e08f973b5852ee1e832a55f3`; official origin is `git@github-xuejin:xuejin-lu/lattigo.git` (SSH host resolves to `github.com`).
- Go: `go1.26.4 darwin/arm64`; OS/architecture: `Darwin 25.6.0 arm64`; `CGO_ENABLED=1`.
- Build directory: `/private/tmp/bts-fast-003-build.XJZ4gy`.
- Task-local modfile: `/private/tmp/bts-fast-003-build.XJZ4gy/bts-fast.mod`, copied from the original `external/BTS/go.mod`; the only addition is `replace github.com/tuneinsight/lattigo/v6 => /private/tmp/lattigo-bts-003-working.HnjCjy/repo`.
- Exact build command: `env GOCACHE=/private/tmp/go-cache-bts-validation-003 GOWORK=off go build -mod=mod -modfile=/private/tmp/bts-fast-003-build.XJZ4gy/bts-fast.mod -o /private/tmp/bts-fast-003-build.XJZ4gy/bts-fast .`
- Build succeeded. `go list -m` and `go version -m` resolve `github.com/tuneinsight/lattigo/v6` to `/private/tmp/lattigo-bts-003-working.HnjCjy/repo`; that clean checkout is commit `049ffa782729d049e08f973b5852ee1e832a55f3`.
- Binary SHA-256: `5574e7e1876678e5243bcf6a6b653c9a4d6c1a7bc2abdf8989d7d154883398e2`.
- `go version -m` VCS metadata identifies the Primary build tree as `8ee8f7392f17a4d03b4c218f8a2c5b2dc85e38da`, `vcs.modified=false`; the Lattigo replacement path and its separately verified Git commit above identify the Fast backend.

## Invocation record

- Exact run command, from `external/BTS/`: `set -o pipefail; /private/tmp/bts-fast-003-build.XJZ4gy/bts-fast 2>&1 | tee /Users/xuejinlu/Developer/xuejin-lu/heart-lattigo-bootstrap/results/BTS-FAST-NOTRACE-VALIDATION-003-attempt-01.log`
- Process launched: **no**.
- Entered `bts.Bootstrap`: **no**.
- `NewBTS()` result/time: not run.
- Bootstrap result/time: not run.
- Output Level/Scale and numerical errors: not available.
- Raw log: will be created only by the single reserved invocation above.
- Existing historical Standard baseline (not rerun): 139.653 s; max absolute error 1.708e-3; mean absolute error 1.172e-4.
