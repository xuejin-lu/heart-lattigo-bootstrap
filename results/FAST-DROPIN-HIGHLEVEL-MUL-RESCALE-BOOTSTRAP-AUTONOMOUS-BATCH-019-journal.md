# Batch 019 execution journal

## Startup and frozen inputs

- Primary startup sync completed on `main`; starting Primary HEAD / `origin/main`: `a40e06a3903fa91f1c3606e6d576a2d710d80fb4`, clean.
- Secondary `fast-qprefix`: `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean and read-only.
- Genuine Standard checkout: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean and read-only.
- Frozen config and Q/P hashes matched Batch 017/018. No Secondary or Standard source was modified.
- Batch cost ceiling: Standard Bootstrap `0/1`, Fast Bootstrap `0/1` before B.

## A — exact scale, capacity, and source feasibility

The task requires Scale `2^45` after a logical-top-q5 Rescale. With both inputs at the default, the exact equation is `2^90/q5`; for `q5=1152921504606830593`, the resulting `log2` is `30.00000000000002`, not 45. No default/config change, exponent trick, extra Rescale, or `SetScale` was selected. Instead, following the accepted Batch 017 public-CKKS precedent for per-input C Scale, A/B remain at `2^45`, and C uses exact `Scale=q5` in both builds. The chosen product Scale is `40564819207302764422326571237376`; dividing by q5 returns exact `35184372088832` (`2^45`).

Capacity bounds were computed with exact rational representations of the deterministic binary64 coordinates (not a rounded `float64` ceil). Results are recorded in the summary and evidence JSON. The decisive checks were `2B_MulRelin < S_Q0123` at Level 5, `2B_Rescale < S_Q0123` at Level 4, and `2B_Rescale < q0` for the final Level-0 projection. The q0 margin is `31486525592416549`. The Q0123 product is `5986308565615587353347023386369277282933624144412673`.

The frozen full Bootstrap parameter object has `MaxLevel=16`, E32, and the canonical full Q/P arrays. Public `GenEvaluationKeys(sk)` extends the same residual secret into that full Q domain; the run checks the extended secret reaches full Q and preserves residual q0/q1. Source audit at the pinned Fast commit confirmed public MulRelin dispatches to the zero-secret shared Fast core with four authoritative rows, and Rescale dispatches to the Q-prefix core while using `ringQ.SubRings[op0.Level()]` (logical q5) as the divisor. Compact outputs leave dormant higher-row backing empty; no Standard fallback or q4+/q5 coefficient read is required. `DropLevelNew(4)` is only a Level projection and preserves Scale.

**A: PASS, zero Bootstrap calls.** The measured C encoding scale is about 21 bits above Batch 017’s accepted q1-scaled C. Its scale-limited encoding error therefore decreases; the output Scale remains exactly the frozen `2^45`. Native B measurements below validate actual precision, including Standard key-switch/encryption noise.

## Runner and preflight validation

The Primary-only runner/test commit is `9c3a705c854a3c03d7182786c9463f0f6606722c`. Both target builds passed before measurement:

```text
GOCACHE=/private/tmp/batch019-gocache go test ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -count=1
GOCACHE=/private/tmp/batch019-gocache GOWORK=/private/tmp/batch016-standard-work/go.work go test -tags lattigo_standard ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -count=1
GOCACHE=/private/tmp/batch019-gocache go vet ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019
GOCACHE=/private/tmp/batch019-gocache GOWORK=/private/tmp/batch016-standard-work/go.work go vet -tags lattigo_standard ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019
```

The first normal-cache test attempt hit a sandbox write denial in the external default Go cache; retrying with the task-local `/private/tmp/batch019-gocache` passed. No permission workaround was used.

Each backend ran the same committed Primary source and full-profile public chain:

`EncryptNew(A,B,C) → AddNew(A,B) → MulRelinNew(sum,C) → Rescale(product) → RotateNew(1) → DropLevelNew(4) → native Level-0 DecryptNew/Decode`.

Commands (raw per-backend vectors were held only under `/private/tmp/batch019-run.h4yTcR/`):

```text
GOCACHE=/private/tmp/batch019-gocache GOWORK=/private/tmp/batch016-standard-work/go.work go run -tags lattigo_standard ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -out /private/tmp/batch019-run.h4yTcR/standard.json -primary-commit 9c3a705c854a3c03d7182786c9463f0f6606722c -backend-commit 5dbffbdea05394de2ca3a432ed5318aa832e3f40 -backend-ref 5dbffbdea05394de2ca3a432ed5318aa832e3f40
GOCACHE=/private/tmp/batch019-gocache go run ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -out /private/tmp/batch019-run.h4yTcR/fast.json -primary-commit 9c3a705c854a3c03d7182786c9463f0f6606722c -backend-commit 2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac -backend-ref fast-qprefix
GOCACHE=/private/tmp/batch019-gocache go run ./tools/fast-dropin-highlevel-mul-rescale-bootstrap-batch-019 -mode combine-preflight -standard-run /private/tmp/batch019-run.h4yTcR/standard.json -fast-run /private/tmp/batch019-run.h4yTcR/fast.json -out /private/tmp/batch019-run.h4yTcR/preflight-pair.json
```

Both runs reached the one-shot wait gate with `bootstrap_calls=0`. The Fast observer passed for A/B/C, Add, MulRelin, Rescale, Rotate, and DropLevel; Fast outputs after Add/MulRelin stored four maintained rows, after DropLevel one q0 row, with c1 zero. Standard remained an ordinary genuine encrypted lifecycle with nonzero c1. Native preflight max errors were `5.86605418722453e-11` Standard and `2.1986015127059466e-12` Fast. The paired preflight gate returned `BATCH019_PREFLIGHT_PASSED_PENDING_ONE_SHOT_BOOTSTRAP`: RMSE `1.4678127606723786e-11`, max `5.903515384823777e-11`, SNR `152.5550089855535 dB`.

**B: PASS, zero Bootstrap calls.**

## C — bounded one-shot Bootstrap pair

Only after the paired B gate passed, the held Standard process received one `bootstrap` release and completed one public call. It returned Level 1 / Scale `2^45`; native oracle RMSE `4.826764480602127e-9`, max `1.424471312557334e-8`, SNR `102.21530000009507 dB`.

The held Fast process then received one `bootstrap` release and completed one public call. It returned Level 1 / Scale `2^45`, two residual rows and zero c1; native oracle RMSE `1.1795718406445398e-9`, max `5.7387742390963576e-8`, SNR `114.45393423748425 dB`.

Final paired metrics: RMSE `4.937529503163558e-9`, max `5.038821901648196e-8`, SNR `102.01822817774493 dB`; fixed `1e-6` gate PASS. Total calls Standard `1/1`, Fast `1/1`; no retry, benchmark, sweep, LogN16, or later task.

Primitive key accesses were observed through a wrapper around the same public keyset: Standard MulRelin fetched one relin key and Rotate fetched one Galois key; Fast MulRelin and Rotate made zero direct relin/Galois lookups. Both evaluator constructors enumerated the Galois-key list once. Bootstrap-internal key accesses use the existing concrete `EvaluationKeys` and were not runtime-instrumented; no zero-access claim is made.

## D — artifacts and handoff

- Compact evidence JSON: 32,095 bytes; decoded 4096-slot arrays are excluded.
- Summary and this journal report exact pins, scale/capacity proofs, checkpoints, key accesses, SNR, tests, and limitations.
- Final paired evidence status: `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`.
- Secondary remains unchanged and clean at `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`; Standard remains unchanged and clean at `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
