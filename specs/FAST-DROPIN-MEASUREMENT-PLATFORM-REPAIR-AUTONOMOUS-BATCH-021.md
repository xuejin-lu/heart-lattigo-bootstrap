# Batch 021 — Refit existing measurement platform to formal public E32 CKKS

Status: READY_FOR_CODEX. Autonomous reuse-and-repair engineering milestones P1–P5. One Web review at completion or actual architecture blocker.
Normative: Primary/Secondary AGENTS.md, Primary docs/MEASUREMENT_PLATFORM.md, docs/RESEARCH_ENGINEERING_WORKFLOW.md §4A/4B, Secondary docs/FAST_QPREFIX_SPEC.md.
Authority: Primary measurement code only; both pinned Lattigo implementations READ-ONLY. Genuine original Standard 5dbffbdea05394de2ca3a432ed5318aa832e3f40; Fast original baseline 2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac (Secondary AGENTS doc-only commit is newer; identify and pin code unambiguously).

## Goal and priorities

Repair existing reusable platform in place, NOT another Batch-specific stopwatch/runner: `cmd/perfprobe/`, `cmd/fastdiag/`, `internal/perfmeasure/`, `internal/numericalmetrics/` and their existing tests. Preserve historical legacy diagnostic modes, existing result schemas/artifacts and prior DIAG work as immutable. Enable the accepted Batch019/020 genuine Standard vs zero-secret Fast **ordinary public** CKKS E32 LogN13 pipeline with the same frontend and Q/P/input. The next cold/warm Bootstrap experiment is DEFERRED until this platform passes a Web compatibility review.

**No expensive Bootstrap calls in this repair batch**; zero benchmark sweeps, zero LogN16, no production or mathematical library modifications. Run cheap unit/static/integration preflights only. If zero Bootstrap cannot prove full public path, document the precisely remaining one-call acceptance smoke for a later Web decision; do not secretly run it.

### P1 — Audit and freeze contract, no new code

1. Safe sync repos, read durable platform docs, current source/tests and historical reports. Build explicit existing-function reuse map: perfprobe timing/warmup/paired compare/input; fastdiag stage and reference attribution; perfmeasure exact Q/P+input; numericalmetrics; existing Secondary tracing hooks.
2. Inventory obsolete paths/constraints with file/function/line evidence: E=0 hard-code in perfmeasure; Fast direct-c0, zero-a input constructor; explicit NewFastEvaluator backend; old commit pins; perfprobe `warmup>=1,repetitions>=7,standardTrials>=3` and expensive implicit calls; fastdiag P93-specific pinned E0 code; source/provenance schemas. Separate truly required Fast zero-c1 from obsolete universal zero-a input generation/E=0.
3. Produce a transition map from legacy diagnostic modes to explicit formal `public-e32` mode; preserve old modes and their tests without pretending E0 and E32 are the same semantics. No duplicate metrics or stopwatch code.

### P2 — Repair shared profile/input/backends

1. Extend `internal/perfmeasure` to accept explicit profile EphemeralSecretWeight with the documented existing default preserved for historical profiles and explicit E32 for the new mode; current accepted canonical config `configs/bootstrap_config.logN13.json`, exact Q/P, plaintext, default scale2^45, per-input C scale=q5 and true full-Q key domain, without altering existing config/hash.
2. Add explicit versioned mode separation to existing `cmd/perfprobe` CLI and adapters: default historical behavior remains labelled legacy unless user selects formal `public-e32`. In formal public mode, genuine Standard uses ordinary keygen+native EncryptNew and native DecryptNew; Fast uses ordinary zero-secret public lifecycle (not manual c0 patch) and public GenEvaluationKeys→bootstrapping.NewEvaluator→Bootstrap. Build-tag may select backend construction, not app operations. Respect original 019 chain AddNew→MulRelinNew→Rescale q5→RotateNew→DropLevelNew(4), not simplified direct Bootstrap. Fast zero-c1 remains checked as a property rather than injected artificially.
3. Reuse/refactor existing perfprobe helpers, not new parallel `tools/batch021` package. If shared code requires a small extracted common workload function to avoid duplicate 019 arithmetic, implement with source-identical Standard/Fast and regression against frozen 019 preflight. Do not change Secondary library APIs.
4. Preserve existing `perfprobe` compare, output and stage evidence, updating only schema/mode distinctions needed; make empty/length mismatch/nonfinite pairwise comparisons fail closed.

### P3 — Upgrade diagnostics and safe run budgeting

1. Route public profile through existing `cmd/fastdiag` stage/power/rescale tracing where supported; adapt hooks/CLI only at harness level and label unsupported stages explicitly. Do not falsely label legacy P93 trace as current E32. Keep existing `fastdiag numerical` legacy outputs as such; add current-profile support via reusable helpers where viable. If full stage tracing requires Secondary math or unsupported hooks, report an explicit nonblocking diagnostic gap rather than inventing a new tracer.
2. Add hard limit mechanism for *total actual Bootstrap invocations*, including warmups, error calls and acceptance smokes: zero-call preflight must be possible despite historical `warmup>=1, repetitions>=7` defaults; explicit bounded policies for later 1/2-call cases; fail-close on implicit warmups/trials. Record immutable consumed attempts and stop rather than retesting expensive operations.
3. Instrument proper evaluation phase labels: evaluator construction, first cold bootstrap (possibly includes Fast lazy circuit initialization), later warm bootstrap; preserve actual elapsed wall and Go alloc deltas using existing timing engine. NEVER claim comparisons under different init boundaries as steady-state evidence. Do not move lazy Fast initialization with a backend-specific production special-case.
4. Add package unit tests for no-call policies, mode/provenance, correct sample/statistics and legacy mode. All tests in this batch MUST require zero actual Bootstrap calls.

### P4 — Cheap current-E32 native integration preflight

1. Run current `public-e32` same-source pre-Bootstrap native chains (no actual Bootstrap) in genuine pinned Standard and public Fast. Assert 4096-slot deterministic input Q/P/backend/source/workload SHA identity, E=32, Scale q5 on C and 2^45 A/B, public calls, q4/q5 dormant after Fast compact outputs, strict capacity 2B<q0, native Level0 decrypt, finite nonempty numerical samples under 1e-6. These are the evidence gates from accepted 019/020.
2. Verify old legacy E0/direct-c0 mode still works as old-style *diagnostic* and is never silently promoted to formal Standard/Fast comparison. The formal E32 mode must not call explicit NewFastEvaluator. Do not edit the old 019/020 reports.
3. If precise compatibility needs changes to mathematical Fast backend, or pinned refs cannot satisfy native public lifecycle, STOP for Web with source-backed gap. Do not backfill by modifying app code based on Fast backend identity.

### P5 — Document and handoff

1. Summarize original platform reuse vs changed functions, E0→E32 semantics/profile migration, backend adapters, input lifecycle and stages, CLI use examples (carefully with cost flags), tests, exact SHAs and caveats. Update docs/MEASUREMENT_PLATFORM.md status matrix to reflect reality and remaining gaps. Add durable cross-chat instructions only when new source evidence warrants.
2. Publish `results/FAST-DROPIN-MEASUREMENT-PLATFORM-REPAIR-AUTONOMOUS-BATCH-021-{summary.md,journal.md,evidence.json}`; atomic safe Primary commits/push, focused tests/vet/diff, both pinned backend code unchanged and clean. Secondary AGENTS doc-only change is already authorized and must not be conflated with production change.
3. `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only when public E32 cheap preflight+formal lifecycle+budget gates and preserved legacy tests pass. Otherwise `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with smallest exact unsupported feature. Do not auto-start delayed cold/warm Bootstrap experiment.

**STOP:** any crypto math/params/QP/Scale/Level/secret mode change; changing accepted numerical tolerance; silent historical schema rewrite; new parallel measurement subsystem; >0 Bootstrap calls; hidden zero-a input in formal mode; unsupported backend shape/parameter; unsafe Git.
