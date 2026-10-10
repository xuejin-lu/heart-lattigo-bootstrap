# Current Task

Task: FAST-DROPIN-E32-TRACE-REPAIR-AUTONOMOUS-BATCH-024
Status: READY_FOR_CODEX
Mode: ROOT-CAUSE-PROVEN RESCALE TRACE REPAIR + FAIL-CLOSED E32 DIAGNOSTIC SESSION
Web review: once after Batch024, or immediately if new math/unsafe obstacle.

**Authoritative spec:** `specs/FAST-DROPIN-E32-TRACE-REPAIR-AUTONOMOUS-BATCH-024.md`
**Batch023 review:** `results/FAST-DROPIN-PUBLIC-NATIVE-REPEATABILITY-AUTONOMOUS-BATCH-023-web-review.md`
**Permanent rules:** both AGENTS.md, docs/MEASUREMENT_PLATFORM.md and workflow §4A/4B; Secondary docs/FAST_QPREFIX_SPEC.md.

## Accepted prior result and diagnosed failure

Batch023 public repeatability passed: canonical LogN13 E32 4096 slots, Standard 6 + original Fast 6 real Bootstrap calls, 6/6 matched native output numeric gates RMSE `5.0361026156396345e-9`, max complex `4.062428762032295e-8`, 5 warm Standard median `306.636ms`, Fast `72.580ms`, observed **4.225x**, not stable hardware-general speedup, Fast intentionally zero-secret/insecure. Diagnostic Fast two additional calls passed native oracle, but trace validation failed because `RescaleWorkspace.ApplyRows` records `rescale→preflight` and executes the true subsequent NTT/MForm/commit **without emitting `materialization`** or `ntt_montgomery_restore` subevents. This is proved from source, not an unknown missing event. 023 budget **14/14 SPENT**; do not retry 023.

## Batch024 autonomous S1–S5
S1: source map true producer/consumer and optional previously preserved pprof, zero Bootstrap.
S2: add actual real-Rescale cheap event regression, then guarded **diagnostic-only** span emissions in Secondary `schemes/ckks/internal/fastcore/rescale.go ApplyRows` (ONLY authorized non-test change), absolutely no math changes. Repair E32 test to always preserve RAW events before fatal tree validation; Primary existing fastdiag must preserve raw path on failure. No Bootstrap.
S3: zero-call regression for Rescale event shape, widths/alias, enabled/disabled equivalence, public E32 fixture validation and cost/provenance. STOP if not proven before calls.
S4: NEW independent maximum **two Fast E32 diagnostic Bootstrap** calls total: one cold + one warm traced, budget tokens journaled before process spawn, no Standard/formal-fast runs, no retries. Numerical oracle & raw event preservation before event-tree validation. If valid derive stage/power/rescale Pareto/closure and bounded Amdahl, otherwise report first mismatch with raw provenance and no extra calls.
S5: one compact Primary evidence/report and permanent docs update, test/vet/diff/safe commits and independent Web review, no Batch025 authority.

**Frozen:** original Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` read-only; original uninstrumented Fast production `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` read-only. Diagnostic Secondary branch `fast-qprefix` may receive precisely approved event-span-only source changes plus relevant tests, recorded as distinct diagnostic SHA. Canonical public CKKS E32 Q/P/input, Level/Scale/c1/rows/capacity/math/oracle unchanged. No new runner or trace engine.

**Budget:** absolutely zero Bootstrap before S4, at most two NEW Fast-only in S4. Never reuse 023's 14 spent calls. Avoid triggering broad/bootstrap tests; all source preflight tests cheap/fake only. No LogN16, sweeping or altering Standard.

## Web-approved S4 zero-call fixture correction (2026-10-10)

**READ FIRST:** `results/FAST-DROPIN-E32-TRACE-REPAIR-AUTONOMOUS-BATCH-024-web-review-resume.md`. The last S4 attempt was **pre-reservation validation only**, supplied `fast-standard-public-native-vectors.v1` preflight instead of the frozen **six-output** `fast-standard-public-native-repeatability-vectors.v1` artifact; no attempts were reserved and no real Bootstrap was called. This is NOT a consumed or failed Bootstrap attempt. Web **authorizes re-running ONLY the zero-call fixture preflight** after identifying and independently verifying the actual six-output vectors saved from Batch023. Search read-only only in recorded task temp paths, inspect metadata/hashes without dumping vectors. Do NOT rename or fake schema, lower gate, rerun 023, regenerate lost vectors from summaries, use wrong file, or alter arithmetic. If authentic complete artifact is absent STOP and report; if genuine and valid, continue S4 with exactly original two new Fast-only diagnostic calls and immutable journals as previously specified. Preserve original 024 blocker evidence and add new provenance; never use Bootstrap for artifact discovery. When ready and clean, re-run the mandatory AGENTS startup synchronization before acting.
