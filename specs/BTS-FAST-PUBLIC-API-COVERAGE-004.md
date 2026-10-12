# BTS-FAST-PUBLIC-API-COVERAGE-004

Status: `BATCH_IN_PROGRESS`\
Class: bounded CKKS API integration and No-Trace experiment (I/E)\
Repositories: Primary `heart-lattigo-bootstrap`; Secondary `lattigo` `fast-qprefix`

## Authority and objective

This is one autonomous Batch charter, authorized by the Web-reviewed user direction dated 2026-10-12. Preserve Task 003, its `1/1` failed Bootstrap attempt, raw log, summary, and journal unchanged. Preserve the original Secondary worktree's five known dirty CKKS files; implementation must use a separately verified clean checkout with the real GitHub origin.

End goal: run the unchanged research-lab `external/BTS/` LogN16 program through ordinary CKKS/RLWE APIs on the existing Fast zero-secret Q0123 backend. This Batch repairs only evidenced API dispatch/representation integration gaps in the actual BTS call path. It does not redesign Q-prefix, Rescale, CRT, key switching, rotations, capacity tracking, or measurement infrastructure.

## Frozen invariants and scope

- `external/BTS/` source, `go.mod`, `go.sum`, parameters, seed, input generation, and operation order are immutable and must never be staged or published.
- Use the original BTS defaults: LogN16, SlotLimbs=15, `BTSModePlain`, amp=1, seed=0, and original Q/P and scales.
- Application calls remain ordinary public CKKS/RLWE API calls. No frontend Fast branches, `NewFastEvaluator`, manual ciphertext edits, hidden Standard/full-RNS fallback, or stale dormant-row reads.
- Fast zero-secret semantics remain mode-specific. Elide zero-operand Key Switching only where the existing API boundary validates the required zero component(s); retain required Galois permutations and output metadata/representation.
- Reuse existing Fast kernels and `schemes/ckks/internal/fastcore` where source dependencies permit. Do not duplicate scalar multiplication, Rescale, CRT, rotation, capacity, or statistics algorithms. A source-backed import-neutral extraction/shared bridge is allowed; no Go import cycle.
- No Standard rerun, warmup, Trace, benchmark campaign, or additional measurement platform.

## Checkpoint A — zero-Bootstrap API coverage audit

Trace the unchanged BTS call graph across construction, S2C, ScaleDown, ModUp, Trace, C2S, EvalMod/PS powers, DoubleAngle, and final output. Produce a compact source-backed table in the Batch summary/journal, with one row per ciphertext-processing public API and columns:

1. BTS caller and exact API / operand kind;
2. ordinary public dispatch and implementation actually executed under Fast zero-secret mode;
3. existing Fast kernel reused, or explicit gap classification;
4. Q-prefix rows versus verified complete-Q or QP representation;
5. dormant-row read/materialization risk and its guarding condition;
6. Level, Scale, NTT/Montgomery and zero-component contract relevant at that boundary;
7. existing or added test proving dispatch and output.

Cover at least the BTS-used paths among `Mul` complex/scalar/element variants, `MulThenAdd`, `Add/Sub`, `Conjugate`/automorphism, `Relinearize`, `Rescale/RescaleTo`, `Trace`, and any direct RLWE evaluator/key-switch method. Do not count unused APIs as coverage. In particular explain the Task 003 failure at `ckks.Evaluator.Mul(ctImag, -1i, ctImag)` and prove whether the current ordinary path is still the full-Q scalar routine.

## Checkpoint B — reuse-first API repair

Repair `ckks.Evaluator.Mul(ctImag, -1i, ctImag)` through the existing Fast Q-prefix complex-scalar arithmetic. First verify its scalar encoding, exact complex multiplication, in-place safety, representation preconditions, and dependency direction. Prefer extracting the actual shared low-level kernel into import-neutral `fastcore` and calling it from both existing evaluators; a minimal bridge is acceptable only if it delegates to that same kernel and does not duplicate arithmetic. Preserve ordinary public API behavior, Level/Scale, NTT/Montgomery state, and supported zero-secret c1 semantics.

Repair other same-class dispatch gaps only when Checkpoint A confirms they are in the actual BTS path, the behavior is already mathematically specified by current production APIs, and the existing Fast core can be safely reused. Preserve verified complete-Q/QP compatibility lanes where required by BTS; they may not consume unverified dormant rows. If a fix requires a new mathematical, capacity, key-layout, or representation rule, or no safe shared/import-neutral path exists, stop immediately with `NEEDS_WEB_REVIEW` and concrete source/test evidence.

## Checkpoint C — complete non-Bootstrap verification

Add regressions to existing CKKS/Fast suites, not a new test harness. Use the original BTS parameter shape at the relevant high Level. Exercise the ordinary `ckks.Evaluator` call chain through the repaired API and the immediately following public operation. At minimum verify:

- complex `Mul(-1i)` against an independent decoded/NTT-domain oracle;
- compact Q-prefix and the explicitly verified complete-Q path where each is supported;
- in-place behavior and no partial output mutation on invalid/missing backing;
- Level and Scale, NTT and Montgomery flags and coefficient representation;
- authoritative zero c1 and correct treatment of absent/dormant rows;
- downstream Rescale or next API compatibility after the scalar multiply;
- a test-visible proof that the intended Fast shared kernel actually executed, separate from numerical correctness.

Run focused CKKS compatibility tests, `go test ./schemes/ckks/fast -count=1`, relevant bootstrapping interface/dispatch tests that invoke no Bootstrap, affected-package `go vet`, and `git diff --check`. Run `go test ./schemes/ckks -count=1` on final code and a pristine Batch-start baseline at `049ffa782729d049e08f973b5852ee1e832a55f3`; compare exact failing test names/errors. Also retain the existing Task 003 comparison to pristine `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` for classifying historical suite failures. Fix all Batch-introduced regressions; do not delete, weaken, or repair unrelated historical tests.

## Checkpoint D — bounded complete BTS validation

Maximum: **two actual Fast `bts.Bootstrap` invocations total for this Batch**. Task 003's already-spent call is separate and remains immutable.

Checkpoint D is prohibited until A–C, self-review, and a Secondary code commit have passed and the Secondary commit is normally pushed. Before each authorized process is launched, persist a unique attempt reservation in `results/BTS-FAST-PUBLIC-API-COVERAGE-004-journal.md`, including attempt number, UTC reservation, unchanged BTS source manifest, backend commit, task-local modfile/module replacement, exact build/run command, binary SHA-256, and invocation count. Save full stdout/stderr to a unique Primary raw-log path. Do not relaunch the same binary.

Attempt 2 is allowed only if attempt 1 entered Bootstrap and revealed a **new, explicit ordinary public API dispatch gap** already within this charter; repair it, add/verify its non-Bootstrap regression, create and push a distinct Secondary commit, rebuild to a different binary SHA, and write a second reservation first. A math/capacity/representation/security conflict, unexplained numerical failure, or out-of-scope issue stops the Batch immediately. A started or failed call consumes its reservation and is never retried. Do not rerun Standard or add warmup/Trace/benchmark calls.

Use the unchanged BTS module and parameters with `GOWORK=off`, a task-local alternate `-modfile`, and an explicit replacement to the exact pushed Fast commit. Record provenance and, if a complete ciphertext is produced, `NewBTS()` and Bootstrap wall times separately, output Level/Scale, original BTS max/mean absolute error and max-error index. No speedup or correctness claim without complete output and valid decoded comparison evidence.

## Deliverables and Git boundaries

- This is the only new Batch spec. Primary active pointer and one compact journal must make progress and the 0/2 attempt budget recoverable across sessions.
- Primary task bootstrap commit: this spec, `CURRENT_TASK.md`, and initial journal only; push normally before Secondary implementation.
- Secondary: only authorized CKKS/Fast/shared-core source and tests needed for the audited API gaps. Never modify the original dirty Secondary worktree; use a clean independent checkout synced from the actual GitHub origin. Commit and normal fast-forward push after Checkpoint C gates.
- Primary final evidence: compact API coverage/result summary, journal, and any unique raw No-Trace logs; no BTS source or binary. Preserve Task 003 artifacts.
- Self-review the complete diff, verify Primary/Secondary provenance and clean status of the used independent checkout, then push Primary results normally. The original Secondary worktree is expected to remain dirty and untouched.

## Success / stop status

Successful completion requires A–C passing, the bounded D outcome recorded (successful or exhausted on failure), task artifacts pushed, no unauthorized source changes, and an honest separation of:

1. public API transparency;
2. numerical correctness;
3. proof of actual Fast kernel dispatch;
4. complete BTS execution time;
5. remaining full-Q/QP work (never relabel Key Switching omission as end-to-end Q0123-only).

At the completed Batch boundary report `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`. If a new mathematical/architecture decision or stop condition is encountered, preserve evidence and report `BATCH_BLOCKED_NEEDS_WEB_REVIEW`.
