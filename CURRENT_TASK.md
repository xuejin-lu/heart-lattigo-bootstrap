# Current Task

Task: FAST-DROPIN-MEASUREMENT-PLATFORM-REPAIR-AUTONOMOUS-BATCH-021
Status: READY_FOR_CODEX
Mode: PLATFORM-FIRST AUTONOMOUS P1–P5 MIGRATION, NO EXPENSIVE BOOTSTRAP
Review: ONE Web scientific/code-architecture review at completion or genuine STOP

**Runnable and authoritative task:** `specs/FAST-DROPIN-MEASUREMENT-PLATFORM-REPAIR-AUTONOMOUS-BATCH-021.md`
**Mandatory permanent infrastructure rules:** Primary `AGENTS.md` plus `docs/MEASUREMENT_PLATFORM.md`; Secondary `AGENTS.md`; `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B.
**Predecessor accepted:** Batch020 Web review. **Supersedes previous** `specs/FAST-DROPIN-BOOTSTRAP-PERFORMANCE-ATTRIBUTION-AUTONOMOUS-BATCH-021.md` cold/warm plan, now DEFERRED until shared platform is fixed and independently reviewed.

## Why this task

The project already has `cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, `internal/numericalmetrics`, root `fast_measurement*.go` and Secondary `internal/fastdiag`. These are **primary research infrastructure**, not historical examples. Batch019/020 duplicate stage timing/provenance/metrics because AGENTS previously failed to enforce reuse. No further Batch-specific timing runner should be written before repair.

Legacy tool weaknesses established from source: E=0 hardcoded in `perfmeasure.ParametersFromConfig`; Fast `perfprobe` direct-c0 input/manual c1 zero and explicit NewFastEvaluator; restricted legacy refs/profiles, forced warmup/repetitions. `fastdiag numerical` has E0/P93 legacy assumptions. This does NOT mean Fast current zero-secret mode should have nonzero c1: distinguish E/ephemeral weight from zero-secret coefficient semantics.

## Five autonomous deliverables

P1: inspect current source and existing measurement tooling; source-backed reuse/gap inventory with migration map.
P2: repair `perfmeasure/perfprobe` in place for named current formal E32 Public CKKS mode using original 019 chain, native Standard and Fast public constructors, preserved legacy diagnostic mode and hashes.
P3: integrate `fastdiag` where genuinely compatible; adapt bounded invocation budget, measurement stages and legacy mode labeling. No new tracing system.
P4: paired zero-Bootstrap cheap preflight with genuine Standard vs Fast exact accepted 019 profile, Level/Scale/capacity/native Decrypt and numerical gates. Legacy tests unchanged.
P5: updated durable platform status, compact report and evidence, focused tests/vet, normal safe Primary commits/push, Web review.

**Freeze:** pinned true Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, original Fast production `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` (Secondary may have documentation-only AGENTS update), original LogN13 E32 Q/P, 4096 slots, A/B Scale2^45, C Scale=q5, E32 Public Bootstrap keyplan, 1e-6 numerical gate. No mathematical/backend production changes.

**Budget:** ABSOLUTE zero actual Bootstrap calls, zero LogN16, no benchmarks/sweeps. All unit tests must avoid triggering Bootstrap. If architecture/math incompatible, STOP for Web with exact evidence; never silently use legacy E0/direct-c0 as formal.
