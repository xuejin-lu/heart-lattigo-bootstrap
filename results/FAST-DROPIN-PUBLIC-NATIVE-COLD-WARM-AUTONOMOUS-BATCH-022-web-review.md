# Independent Web Review — Batch 022 Public-Native E32 Cold/Warm Bootstrap

**Reviewed:** 2026-10-10 Asia/Taipei.
**Decision:** **ACCEPT** the bounded platform reuse, source/evidence-integrity repair, genuine Standard/Fast public E32 numerical correctness and precisely two cold/warm Bootstrap calls per backend. **DO NOT** treat a single warm observation as a statistically established 4.26× speedup or equate zero-secret Fast with cryptographic Standard security.

## Verified submission

- Primary remote `main@0e6baeb5f84aeb6069e828ec77bebaec3ccc9dfb` confirmed live through GitHub. A source/evidence-integrity patch `cfd9c1548bc39fc697a8c264a4bfd42d00463b61` and B implementation `37002ff3bbd97889cf0ec8108ff2cd527ab8d3d5` are present; final result commit adds docs and 022 report/evidence/journal only. No new standalone measurement project was created. Original Standard production pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; fixed Fast production pin `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`. Active Secondary branch `d463c336d511646994ba72f8dcfda5179431c430` changes AGENTS only after pinned Fast production. Remote heads verified; local pinned checkout cleanliness is journal-reported, not inspectable by Web.
- Inspected submitted `022-summary.md`, `022-journal.md`, `022-evidence.json`, frozen Primary `cmd/perfprobe/public_native.go` and `main.go`, tests and updated `docs/MEASUREMENT_PLATFORM.md`. Web has not independently executed the expensive calls or recovered the task-local 4096-slot vectors or per-invocation reservation JSON under `/private/tmp`; compact hashes, source and journal support bounded acceptance but not independent regeneration of raw per-slot results.

## Project constitution and scientific correctness

Primary safe sync, separate detached original Standard/Fast checkouts, same Primary `cmd/perfprobe` code via isolated temporary modfiles and build tags, source/config/QP/input/workload SHA matching, true Standard generated-key native EncryptNew/DecryptNew, intended insecure Fast zero-secret public constructor dispatch, preserved numerical/Scale/Level/q-authority and `2B<q0` capacity gates are documented and implemented. No Secondary production edits and no additional independent stopwatch or numericalmetrics subsystem were introduced.

The A regression-first fix now recomputes **all eight** 4096-slot decoded vector hashes against checkpoint `DecodedSHA256`, the cleartext oracle and stored error/SNR metrics, finite sample counts, row authority, capacity and metadata. The B native mode uses existing `bootstrapBudget.invoke`, preflight pair hash, exclusive attempt journals, same-held-input copies/fingerprints, and fail-closed first-call error/second reservation path. Negative and fake-call tests exist, with real Bootstrap excluded from unit tests. The report states exactly **two actual Standard plus two actual Fast** public Bootstrap calls, no extra warmup/retry; each returned Level1/Scale2^45/Degree1 with native residual DecryptNew/Decode on all 4096 slots. No `go test ./...` was run to avoid unbudgeted Bootstrap, a reasonable scoped choice.

Preflight eight-checkpoint **overall** largest paired complex error is `7.781760773684476e-11`, not the earlier mid-run `6.62e-16` individual/misinterpreted quantity. Paired post-Bootstrap (cold and warm) RMSE `4.7262101606249815e-9`, max `4.489543437647087e-8`; fixed gate `1e-6` passed, and both backends separately meet their plaintext oracles. Equal within-backend cold/warm decoded hashes demonstrate deterministic results for the same held input in the recorded runs; different Standard/Fast ciphertext SHA hashes and actual separate backend execution are preserved.

## Timings: allocation and evaluator-construction asymmetry resolved, not repeated statistically

**Environment:** Apple M4, macOS arm64, Go1.26.4, `GOMAXPROCS=10`, one cold and one warmed Bootstrap per backend, all reported timings descriptive.

| Phase | Genuine Standard | Intentional zero-secret Fast | Interpretation |
|---|---:|---:|---|
| public Bootstrap `NewEvaluator` | 350.424 ms | 2.443 ms | Standard eager, Fast lazy |
| First Bootstrap call | 312.426 ms | 536.207 ms | Fast carries lazy circuit build |
| **Evaluator constructor + first call** | **662.850 ms** | **538.650 ms** | Fast 1.23× faster *on these two phases only* |
| Second same-evaluator Bootstrap | **299.841 ms** | **70.399 ms** | Fast 4.26× faster, **one warm observation** |
| Second-call Go allocated bytes | 91,806,536 | 7,505,112 | Fast ~91.8% fewer allocated bytes in that one call |
| Second-call objects | 35,019 | 8,085 | Different allocations, not RSS |
| GenEvaluationKeys | 397.630 ms | 819.350 ms | Fast setup cost **larger** |
| GenEvalKeys + constructor + first Bootstrap (sum of shown phases) | **1,060.480 ms** | **1,358.000 ms** | Fast slower ~1.28× when including these phases; not full end-to-end timing |

The last row is a component-wise sum excluding encryption, primitive chain, native decoding, etc., NOT a measured total user request latency. Peak resident memory (RSS) is not proven by allocated bytes. Relative to Batch020's `GOMAXPROCS=1`, this Batch022 used `GOMAXPROCS=10`; the old and new timing numbers cannot serve as clean longitudinal speedups.

**Core conclusion:** The previous Batch020 379ms Standard vs 574ms Fast comparison unfairly attributed Fast's lazy first-call initialization to its steady Bootstrap cost. The new source-backed second-call numbers strongly motivate independent repetition and resource accounting, but do not prove a stable acceleration population estimate, E32 internal stage attribution, secure CKKS speedup, GPU performance or CNN application latency.

## Next research gate

Advance with **a reuse-first autonomous Batch023** rather than a new benchmark harness: extend existing `cmd/perfprobe` public-native mode to a controlled same-evaluator repeated-warm measurement policy with explicit irrevocable call budgets, e.g. 1 cold + 5 warm calls per backend at a single approved LogN13/E32 profile, reporting distribution/median, CPU/GOMAXPROCS/GC configuration and Go allocations; preserve hashes, accepted numerical gates and no extra/unaccounted Bootstrap attempts. At the same time audit the existing `cmd/fastdiag` and Secondary `internal/fastdiag` for formal E32 compatibility. Only implement instrumentation using pre-existing supported hooks/adapters, clearly separate Fast-only internal vs comparable public timing, and STOP if a new crypto library change or altered mathematics is required. No optimization/benchmark sweep or full new runner. Report initialization/keygen vs steady costs separately. Code/test before expensive calls and do not re-run completed 022 Bootstrap quota.
