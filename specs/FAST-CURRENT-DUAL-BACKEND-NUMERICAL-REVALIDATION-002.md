# FAST-CURRENT-DUAL-BACKEND-NUMERICAL-REVALIDATION-002

## Scope and decision

Web review authorizes ONLY Primary measurement-backend isolation and fresh LogN13/LogN16 numerical validation. No Secondary mathematical arithmetic, Rescale semantics, mode-dispatch modification, pre-scale workaround, normalization compensation, benchmark, or optimization is permitted. Preserve results/FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001.md/.json as immutable blocked evidence.

The prior Standard API failure was produced by linking the Fast checkout rather than independently pinned genuine Standard. Fast checkout ckks.Parameters triggers FastCKKSZeroSecretSimulation; in-tree ordinary CKKS Rescale compacts Q rows to q0..q3, and the next C2S transform reads empty q4 and panics. This is a comparator construction error, NOT a genuine Standard benchmark result and NOT proof of current LogN16 EvalMod mathematical failure.

Read and follow both AGENTS.md, Primary docs/MEASUREMENT_PLATFORM.md and RESEARCH_ENGINEERING_WORKFLOW.md after safe git synchronization. Starting Primary main known 816cf8c040172888aecd28f2c7eb820fa633e0a6; Fast source fast-qprefix known 2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f; pin original genuine Standard 5dbffbdea05394de2ca3a432ed5318aa832e3f40. Check actual clean source/ref state and record exact hashes. If changed unexpectedly, fail closed.

## Prior experiment accounting

Earlier blocked task used **2 Fast Bootstrap attempts + 1 in-tree Standard API Bootstrap attempt** on LogN16. All three are already spent; previous task outcome remains LOGN16_CURRENT_REVALIDATION_BLOCKED. LogN13 had no run. Do not overwrite prior artifacts, retroactively relabel invalid Standard, reset prior call journals, or reuse prior obsolete RMSE as new measurements.

The obsolete Q01/P93 automatic tests were already removed. Do not restore them or chase their historical failures. Keep current Q0123 CRT/Rescale/NTT/Chebyshev tests. Historical report files stay in Git.

## P0 — zero-call dual-backend isolation and source audit

1. Reuse the existing Primary shared measurement platform: cmd/perfprobe with build-tag backend adapters, cmd/fastdiag for supported traces, internal/perfmeasure for exact config/input, internal/numericalmetrics for metrics, and existing attempt budgets/journals. No new stand-alone measurement framework. Inventory exact supported interfaces first.
2. Create two clean immutable independent Secondary checkouts/worktrees at the stated pinned SHAs, without changing the user's live Secondary branch or touching original Standard source.
3. Build the same Primary workload and metrics source into TWO isolated binaries with task-local temporary -modfile files, GOWORK=off, and exact replace github.com/tuneinsight/lattigo/v6 to the relevant isolated checkout; reuse perf_standard and perf_fast backend adapters where suitable. Do not change committed Primary go.mod or swap live ../lattigo branch. Record go list -m, go version -m, build tags, actual dependency path, binary hash and underlying SHA. A Standard evaluator compiled in Fast is NOT accepted.
4. Demonstrate at zero Bootstrap calls that genuine Standard does not have the Fast zero-secret mode dispatch and never silently compacts its full active Q limbs during C2S. Specifically verify genuine Standard Level16 input/ModUp, C2S group0 Level16->15 Rescale, full nonempty active q0..q15 limb backing and successful next C2S matrix. Record Level, exact Scale, physical row counts/length, NTT/Montgomery flags.
5. Existing cmd/fastdiag numerical current mode creates both Fast and its in-tree Standard API inside the Fast binary. DO NOT mislabel that as genuine Standard. Refactor only Primary tooling/backend adapter boundaries so true genuine Standard result is obtained from the independent Standard binary. Original Standard does not have Fast private hooks. Use its public stage API when source-supported and label replay-derived checkpoints; mark internal missing comparisons explicitly. Do not patch original Standard.
6. Preserve deterministic historical numerical workload profile for this specific comparison: original encoded c0 / zero-valued c1, diagnostic-only and NOT E32 formal public lifecycle. Do not repurpose or mutate frozen public-native E32 fixture. Use identical original message and effective Q/P parameters in both separate builds, not necessarily same key randomness. Canonical LogN13 LogSlots12 and LogN16 LogSlots15; preserve effective LogN16 q0 56-bit despite target 55. LogN16 original input fingerprint 659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf. Reuse prior LogN13 artifact fingerprints too.

## P1 — tests and zero-call prerequisite

Focused existing platform tests, binary/module source identity assertions and no-Bootstrap C2S preflight MUST pass before any numerical full Bootstrap. Do not run full go test ./... if it implicitly invokes cryptographic Bootstrap. Use go test on targeted lightweight packages, targeted go vet, git diff --check; document pre-existing unrelated root vet self-assignment as historical debt. Any Standard source confusion, missing high limbs, mismatched parameters, or build blockage => stop without spending calls.

## P2 — FRESH bounded numerical executions ONLY after P1 passes

THIS task explicitly authorizes NEW, separately journaled numerical calls:
- LogN13: at most **one genuine Standard public Bootstrap and one current Fast public Bootstrap**.
- LogN16: at most **one genuine Standard public Bootstrap and one current Fast public Bootstrap**.
- Maximum **four fresh full Bootstrap calls total**. Each failed/panicked call consumes reservation. No retries, implicit warmups, other tests invoking Bootstrap, or performance campaigns. Journal pre-reservation, fail closed, use new immutable output paths. Historical three calls remain charged to earlier task.
- Match exact parameter chain, original input values/scale, algorithm and result metric; note that Standard keys/randomness can differ. Compare decoded Fast-vs-genuine Standard and each backend output-vs-original input, both RMSE and max complex error and SNR, finite values and complete slot coverage.
- Compare all mathematically comparable public stages: input, ScaleDown, ModUp, C2S real and imag, EvalMod real and imag, SlotsToCoeffs, and complete Bootstrap. Stage-level public API replays WITHOUT Bootstrap are allowed if source-supported, clearly labeled replay-derived rather than in-circuit. Check exact Level/Scale and validity of RNS/common representation. Never claim false comparability.
- Reuse existing Fast internal tracing for generated Chebyshev T1/T2/T3/T6, polynomial output, three DoubleAngle rounds and before/after public scale reset. Genuine original Standard may lack equivalent private hooks. When an independently justified genuine Standard reference for a given internal checkpoint is unavailable, mark NOT_COMPARABLE_TO_GENUINE_STANDARD, report Fast trace only and do not fabricate small RMSE. An in-tree Standard trace from Fast checkout is auxiliary diagnostic only.
- Suggested numerical gate: final Fast-vs-Standard RMSE <=1e-7; both output-vs-original RMSE <=1e-7; independently comparable EvalMod stage RMSE <=1e-7. Also report max, exact scales and threshold counts. Noncomparable checkpoints prevent claim of fully validated per-stage accuracy; do not relax thresholds or treat missing as PASS.
- If new run remains wrong, identify earliest actual comparable difference; do not fix CKKS math in this task.

## P3 — results, classification and push

Create two NEW files under results/ with stem FAST-CURRENT-DUAL-BACKEND-NUMERICAL-REVALIDATION-002 (.md and .json). Include complete source and binary provenance, tool reuse, exact config/input SHA, separate Standard/Fast identity, fresh call budgets, per-profile staged/terminal RMSE/max/SNR, replay-vs-in-circuit distinctions, Q-prefix capacity evidence where appropriate, missing checkpoint coverage, old/new contrast and explicit limitations.

Classify:
- DUAL_BACKEND_NUMERICAL_VERIFIED: genuine independent Standard comparison, all required comparable public stages and final numerical gates pass, with no falsely claimed internal coverage;
- DUAL_BACKEND_NUMERICAL_DIVERGES: a genuine paired numerical gate fails;
- DUAL_BACKEND_NUMERICAL_PARTIAL: terminal comparison succeeds but some requested staged comparisons are unavailable;
- DUAL_BACKEND_NUMERICAL_BLOCKED: source separation, compile, preflight or execution blocks reliable results.

Only Primary harness/test/report modifications allowed. Preserve both Secondary checkouts and all historical reports; no Secondary commits. Run focused tests/vet and diff checks. Safe ordinary fast-forward push of authorized Primary task output when repo rules allow, then READY_FOR_WEB_REVIEW. Unexpected representation/security/math choices must stop NEEDS_WEB_REVIEW with evidence. No optimization authorized.
