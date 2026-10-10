# Batch 017 execution journal

## Startup and pins

- Primary startup sync: `main` was clean at `0cbd552662e085ac152d17588c82ba9f74b56ffe`; fetched and fast-forwarded to `origin/main=e1c381df0bd19de1780f0fbd1812dd94101bcc5d`.
- Secondary startup sync: `fast-qprefix` clean and equal to `origin/fast-qprefix=2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`; no Secondary source was modified.
- Genuine Standard checkout: clean at `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; read-only.
- Primary runner/test commit before measurements: `f3148a13735f8ba3a4eb9e7dc6a952cf4c1302bd`; both final held runs recorded this same clean revision and frontend SHA.
- Frozen config/profile/workload/QP hashes and complete Q/P values are recorded in the final compact evidence JSON.

## Checkpoint log

| Checkpoint | Status | Bootstrap calls | Evidence / next action |
|---|---|---:|---|
| A — source/depth/key-plan/scale preflight | PASS | 0 | Residual Level 1 product followed by logical-q1 Rescale reaches Level 0 at frozen Scale `2^45`; E32 public key plan and original secret extension verified. Proceeded to B. |
| B — matched cheap public primitive chain | PASS | 0 | Standard and Fast native public lifecycles passed input, Add, ciphertext MulRelin, Rescale, Rotate and decrypt/Decode gates. All checkpoint Level/Scale values matched; direct pair differences passed. Proceeded to C. |
| C — one-shot Bootstrap composition | PASS | 2 (1 each) | Standard public Bootstrap completed first; then Fast public Bootstrap completed on its held preflight chain output. Both outputs passed native decode/oracle and paired gates. No retry. |
| D — evidence/tests/self-review | PASS | 2 total | Compact summary/evidence and this journal prepared; focused tests/vet and final diff review passed. Web review is the only next action. |

## Process-control attempt accounting

The first Standard and Fast cheap preflight pair completed all seven pre-Bootstrap checkpoints successfully and used **zero** Bootstrap calls, but the initial non-interactive execution gave the process EOF while it waited for the explicit C gate. Its temporary raw files record that wait error, not a CKKS failure. To retain the exact in-memory rotated ciphertext through the B→C gate, the preflight pair was rerun in interactive sessions. Both held sessions again passed their preflight gates; only those held sessions were released for Bootstrap. The Standard session was released first and completed successfully before the Fast session was released. Each released session called Bootstrap exactly once. Total actual Bootstrap calls: one Standard plus one Fast, within the batch-wide ceiling. No retries, warmups, benchmarks, or parameter changes occurred.

The first preflight pair's direct metric calculation covered all seven checkpoints and passed. The final combined evidence independently recomputes all eight direct pairs, including the Bootstrap outputs, from the held-run raw vectors. Temporary per-slot JSON remained under `/private/tmp`; only compact aggregate evidence and row hashes were committed.

## Frozen profile and chain

- Genuine Standard pin: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast pin: `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
- LogN13, N=8192, 4096 slots, E=32; residual MaxLevel 1, Bootstrap MaxLevel 16; default scale `2^45`.
- Actual residual `q1=549755731969`; ciphertext C uses this actual prime as its ordinary encoding scale, so `Scale(MulRelin)/q1=2^45`. This is a per-ciphertext CKKS Scale choice in the identical shared frontend, not a parameter/config change.
- Chain: `EncryptNew(A/B/C) → AddNew(A,B) → MulRelinNew(sum,C) → Rescale(q1) → RotateNew(1) → public Bootstrap`.
- Final Bootstrap outputs: Standard RMSE/max `4.908811884776509e-9 / 1.4187430657863595e-8`; Fast `1.1790593543156793e-9 / 5.7394238676200815e-8`; direct Fast-vs-Standard `4.997862453389243e-9 / 4.326414527528976e-8`.

## Runtime dispatch and row authority

- Fast public Add, ciphertext MulRelin, Rescale, and Rotate dispatch to the existing shared Fast cores; source paths and exact per-operation key lookup counters are documented in the summary/evidence.
- Fast primitive outputs use exactly 2 active rows at Level 1 and 1 row at Level 0. Fast c1 remains zero. Standard inputs have nonzero c1 and are decoded through the native original secret lifecycle.
- Both calls used ordinary public `GenEvaluationKeys`, `bootstrapping.NewEvaluator`, `Bootstrap`, `DecryptNew`, and `Decode`. Standard Bootstrap per-call key retrievals are source-traced but not dynamically counted because the public key bundle embeds a concrete `MemEvaluationKeySet`; this limitation is explicit in the evidence.
- No dormant Fast Q row was needed in this low-Level profile; no full-RNS fallback, high-Level compact-decrypt claim, security claim, or acceleration claim is made.

## Verification

Passed under the Fast workspace:

```text
GOCACHE=/private/tmp/batch017-gocache go test ./tools/fast-dropin-public-primitives-bootstrap-batch-017 -count=1
GOCACHE=/private/tmp/batch017-gocache go vet ./tools/fast-dropin-public-primitives-bootstrap-batch-017
```

Passed under the pinned Standard workspace:

```text
GOWORK=/private/tmp/batch016-standard-work/go.work GOCACHE=/private/tmp/batch017-gocache go test -tags lattigo_standard ./tools/fast-dropin-public-primitives-bootstrap-batch-017 -count=1
GOWORK=/private/tmp/batch016-standard-work/go.work GOCACHE=/private/tmp/batch017-gocache go vet -tags lattigo_standard ./tools/fast-dropin-public-primitives-bootstrap-batch-017
```

`gofmt -d` was empty; `git diff --check` is part of the final D gate. No Secondary/Standard files, configs, task pointers, production code, benchmarks, or LogN16 artifacts were changed.

## Handoff

Batch status: `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`. The runner commit is `f3148a13735f8ba3a4eb9e7dc6a952cf4c1302bd`; the summary, journal and compact evidence are committed together in the Primary artifact commit that contains this journal. No Batch 018 work is authorized before independent Web review.
