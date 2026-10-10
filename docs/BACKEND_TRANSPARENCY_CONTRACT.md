# Current CKKS backend-transparent drop-in contract

**Updated:** 2026-10-11
**Status:** implemented API-level backend-transparent integration for the exercised public workflow; latest LogN16 per-stage numerical acceptance remains OPEN.
**Primary:** heart-lattigo-bootstrap main; **Fast Secondary:** fast-qprefix at 2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f; **genuine original Standard:** 5dbffbdea05394de2ca3a432ed5318aa832e3f40.

## Non-negotiable contract

The same **application-facing CKKS source**, ordinary constructor names and operation calls execute with EITHER Lattigo library backend. Selection is strictly by the **linked Lattigo source/commit at build time**. No application branch, runtime Fast flag, special Fast constructor, manually crafted c0 or forced zero c1 to select a backend.

The public application calls the normal `ckks.NewEvaluator` and `bootstrapping.NewEvaluator`, with ordinary `GenEvaluationKeys`, `EncryptNew`, `AddNew`, `MulRelinNew`, `Rescale`, `RotateNew`, `DropLevelNew`, `Bootstrap`, and the prescribed Decode/Decrypt lifecycle. The Fast **library internally** dispatches supported workflows to its optimized zero-secret engine, including Q0123 maintained-prefix arithmetic. `NewFastEvaluator` remains a **private/internal or explicitly historical diagnostic API**, NOT a current frontend requirement.

**Use this selection model, and only this:**

1. Genuine Standard binary: identical Primary application/harness source compiled with an isolated clean Lattigo checkout pinned to 5dbffbdea05394de2ca3a432ed5318aa832e3f40.
2. Current Fast binary: identical Primary application/harness source compiled with an isolated clean Lattigo checkout pinned to fast-qprefix@2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f (or a later explicitly accepted Fast pin).
3. Use separate temporary Go `-modfile` replacements, with `GOWORK=off`, to link either source. Do NOT change the committed go.mod or switch/rebase/reset a user's working Lattigo branch. Compile TWO executables; a Git commit is **not** switched inside an already compiled Go binary.
4. Check only the **source identity**, using pinned SHA, actual module replacement, `go list -m` / `go version -m`, build tags, binary provenance and clean-tree state. This is a routine provenance gate, **not** a new Standard algorithm verification project.
5. Compare exactly the same original message, frontend CKKS parameters, effective Q/P primes, public CKKS operation sequence and metric definitions; backend-specific key material and randomness need not be identical. Shared measurement/provenance/metric tools live in Primary. Build-tag backend adapters may affect orchestration, never application CKKS math.

## The failure mode we must never repeat

The October 11 LogN16 attempt called the ordinary Standard-looking evaluator **while the Go module still pointed to the Fast checkout**. This was NOT genuine Standard: Fast-only `FastCKKSZeroSecretSimulation` dispatch was active in the shared CKKS evaluator, compacted Standard C2S Q limbs to q0–q3 and produced `INTT len(p1)=0` when the next matrix read q4. This is a wrong comparator **build**. Never try to validate original Standard by calling a different constructor in the **same Fast-linked executable**, and never modify Secondary crypto arithmetic to compensate for incorrect backend selection.

## Achieved vs not yet demonstrated

**Achieved, source-confirmed and previously exercised:** frontend/backend-transparent ordinary public evaluator integration on the supported CKKS workload; formal LogN13/E32 public lifecycle has been tested using isolated genuine Standard and Fast pins. See results/FAST-DROPIN-PUBLIC-NATIVE-COLD-WARM-AUTONOMOUS-BATCH-022-summary.md and results/FAST-DROPIN-PUBLIC-NATIVE-REPEATABILITY-AUTONOMOUS-BATCH-023-summary.md. Latest Fast source also contains ordinary public evaluator internal dispatch.

**NOT established:** universal API support outside tested profiles; equivalence of encryption security or noise distributions (Fast is intentionally insecure zero-secret simulation); latest LogN16 stage-by-stage Standard/Fast RMSE. The older failing LogN16 EvalMod values are historical, not valid measurements of current source. No performance work or mathematical fixes until current numerical gates pass.

## Maintenance checklist — MUST preserve

- New application/frontend code MUST NOT call `NewFastEvaluator` or import the Fast package to choose backend; do not reintroduce `--fast`, `isFastMode`, conditional math, or manually rewrite frontend c1.
- A test or runner claiming genuine Standard MUST be linked to **original Standard Lattigo source** (not merely invoke `NewEvaluator` within Fast checkout). Fail closed on wrong module SHA/path/build info.
- Fast-internal trace and historical `NewFastEvaluator` diagnostics, if retained, MUST be visibly labelled *internal/historical diagnostic only*; cannot represent the production frontend.
- Retain current Q0123 numerical regression tests. Obsolete Q01/P93 experiment tests are not current acceptance gates; historic artifacts remain in version history.
- `docs/MEASUREMENT_PLATFORM.md`, `AGENTS.md`, and `CURRENT_TASK.md` must defer to this **current API architecture contract** when legacy inventory statements describe former gaps. Keep this document up-to-date when capabilities change. Never silently upgrade API success to numeric accuracy.

**Next numerical task:** specs/FAST-CURRENT-DUAL-BACKEND-NUMERICAL-REVALIDATION-002.md; it should implement independent pinned dual compilation and measure latest LogN13/16 with the source provenance gate above, without reopening the already achieved frontend integration.
