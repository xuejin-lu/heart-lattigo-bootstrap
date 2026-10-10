# Batch 020 S3 — current Fast Rescale attribution

## Verdict

`NO_SAFE_OPT_CANDIDATE` — skip optional S4 and continue S5 on the frozen original Fast pin. This is the successful diagnostic outcome defined by Batch 020; no Secondary source was changed.

## Frozen provenance and measurement boundary

- Primary harness: `c486cb2b229f4cc6c0b77f9be2b85066c3b6c213` (same-source public-chain instrumentation).
- Current Primary evidence/journal base: `d3f630d1aed3a2c6b99275652fe66d79e8190a6f`.
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Original Fast: `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean.
- Host/runtime for S2: Go 1.26.4, darwin/arm64, Apple M4, `GOMAXPROCS=1`; five process-isolated public-chain trials per backend.
- S2 Fast public Rescale median: `1.730 ms`; it was the largest single timed eval-only stage. Fast eval-only stage-sum median was `3.178 ms`. These are descriptive measurements, not Bootstrap speedup claims.
- One bounded current-pin CPU profile used the existing rows4 LogN13 q0=55 Rescale benchmark (`-benchtime=8s`): 8,078 iterations, `1,278,368 ns/op`, `680 B/op`, `20 allocs/op`; profile contained 9.18 s CPU samples. This benchmark runs at the configured maximum logical level and is a source-attribution proxy, not the exact public-chain q5 Rescale input or a replacement for S2 timings.

## Public dispatch and current source

The public `ckks.Evaluator.Rescale` Fast zero-secret branch in `schemes/ckks/evaluator.go` computes the authoritative `QPrefixWidth(op0.Level())` and dispatches to the evaluator-owned `fastcore.RescaleCore.ApplyRows`. The current `RescaleWorkspace` validates all source rows and components, stages results, verifies every intermediate centered-capacity boundary, and only then commits output. Each component passes through `rescaleComponent`:

1. `prefixToCoefficientRows` performs the required per-row INTT and, for Montgomery inputs, IMForm into reusable `w.coeff` scratch.
2. The coefficient loop gathers the active residues, reconstructs the Q-prefix value, applies centered rounding for each logical rescale, checks capacity, and materializes target residues into component-local staging.
3. After every component passes, the commit phase performs the required output NTT/MForm and metadata/Level/Scale update.

Source reviewed at frozen Secondary `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`; no production diff was created.

## Profile attribution

The sampled profile's largest flat shares were `math/bits.Div64` 15.80%, `RescaleWorkspace.rescaleComponent` 15.58% (cumulative 74.18%), `ring.nttUnrolled16Lazy` 13.73%, `ring.inttLazyUnrolled16` 12.75%, and `ring.MRedLazy` 10.02%. Further arithmetic contributors included `mod192By64` 4.36%, `crtQ0123Prepared` 2.83%, `reconstructQPrefix` 2.40%, and `signedResidue192` 1.63%. Parent cumulative and child flat samples overlap and must not be summed. The profile points to the required transform and centered-CRT/rounding work, not an obviously redundant wrapper operation.

## Candidate assessment

The reviewed path does not expose a representation-neutral allocation or copy that can be removed: INTT/IMForm, CRT reconstruction and rounding/capacity checks, residue materialization, and output NTT/MForm each implement required domain or arithmetic transitions. Replacing the shared coefficient scratch with batched transforms into the existing component staging buffers would retain all transforms and coefficient work; it would instead keep more intermediate component data live and alter staging/cache behavior. The profile provides no evidence that this restructuring is a high-impact improvement, and it is not justified as a mechanical removal of redundant work under S4's narrow authorization.

No safe, source-backed candidate therefore meets S3's `MECHANICAL_OPT_CANDIDATE` bar. S4 is skipped without edits, speculative optimization, or a new benchmark candidate. The frozen original Fast implementation remains the S5 backend.

## Safety and next bounded step

- No Bootstrap was run for S3; cumulative budget remains Standard `0/1`, Fast `0/1`.
- No Standard or Secondary source, parameters, public workload, or numerical threshold changed.
- Continue with the one paired S5 Bootstrap only after all cheap gates and provenance are rechecked. Do not add another optimization task to this batch.
