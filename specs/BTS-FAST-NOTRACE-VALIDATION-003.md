# BTS-FAST-NOTRACE-VALIDATION-003

Status: READY_FOR_CODEX  
Class: bounded CKKS compatibility repair and one No-Trace experiment (I/E)

## Authority and objective

This is a new, independent follow-up to `BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002`, authorized by the Web review handoff dated 2026-10-12. Preserve task 002, its source history, attempt journal, and failed-attempt accounting unchanged. `FAST-Q0123-RESCALE-EXACT-DIVISION-PILOT-006` remains suspended.

Complete the reviewed, bounded compatibility validation for the unchanged research-lab BTS frontend, then run one No-Trace Fast end-to-end BTS invocation if every preflight gate passes. This task does not claim full BTS correctness or a performance win before the new output and numerical evidence are reviewed.

## Frozen source, backend, and workload

- Primary repository: `xuejin-lu/heart-lattigo-bootstrap`, branch `main`.
- Secondary repository: `xuejin-lu/lattigo`, branch `fast-qprefix`.
- Start from the synchronized Primary and Secondary histories. The accepted Secondary Conjugate dispatch repair is `0b922dc48315128a27db811a5f237e0a6c9813d6`.
- Do not modify, stage, commit, or publish any content under Primary `external/BTS/`.
- Use the unchanged BTS defaults: LogN16, SlotLimbs=15, `BTSModePlain`, amp=1, seed=0, original Q/P chain, scale, input generation, and operation order.
- Compile the exact unchanged BTS module against the final pushed Fast `fast-qprefix` commit with a task-local alternate Go modfile, `GOWORK=off`, and explicit local module replacement. Preserve the BTS `go.mod` and Go sources.
- Do not rerun Standard, add warmups, run Trace, benchmark, or execute any extra Bootstrap call.

## Required pre-Bootstrap validation

1. Extend the existing CKKS compatibility regression tests (do not create a separate test platform) with a BTS-shaped high-Level, NTT-domain, complete-Q `Conjugate` case. Verify authoritative zero `c1`, the actual complex-conjugation Galois permutation, all logical Q output rows remain materialized and correct, and Level, Scale, NTT, and Montgomery metadata are preserved. Retain the compact Q-prefix dispatch test.
2. Ensure a deferred error from the void-returning compatibility `ModDownQPtoQNTT` is surfaced by the next compact Fast `Rescale` boundary before any input/output mutation. Test the error and transactional state explicitly. Do not alter or duplicate the shared Fast Rescale arithmetic.
3. Keep full-Q CRT row-consistency validation conceptually separate from a Q-prefix operation-capacity upper-bound proof. The full-Q check may verify canonical residues and that supplied complete rows represent the same centered prefix representative; it must not claim that this alone proves an operation's prior integer bound/no-wrap history. Reuse existing capacity gates/observers where they actually apply; do not introduce another bound tracker or change Q-prefix policy. Do not claim dormant-row reconstruction where the code only validates already supplied complete rows.
4. Run `go test ./schemes/ckks -count=1` on the final code and on an isolated, non-mutating checkout of baseline `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` (the immediate pre-task-002 Fast baseline). Compare exact failing test names and errors. Fix every regression introduced by the task-002 compatibility changes; report unrelated baseline failures without expanding scope.
5. Run focused compatibility and zero-secret automorphism tests, `go test ./schemes/ckks/fast -count=1`, relevant bootstrapping interface/dispatch tests that do not invoke Bootstrap, `go vet` on the affected CKKS/RLWE/bootstrapping packages, and `git diff --check`.
6. Audit the original BTS call path and classify remaining hot-path operations as existing Fast kernels, mathematically elided zero-operand key-switch work, QP/full-Q compatibility work, or verified Standard-compatible operations. Explicitly list full-Q/QP arithmetic still executed; omission of Key Switching is not proof of Q0123-only execution. No hidden fallback may consume stale rows.

If the audit or tests reveal a new mathematical, security, representation, or Q-prefix capacity decision, stop with `NEEDS_WEB_REVIEW`; do not improvise a new semantic rule. Keep all changes CKKS-scoped in Secondary.

## Commit gate and one-call journal

Finish the code fixes, tests, source audit, self-review, and Secondary commit/push before reserving the BTS call. Primary must record the new task and its independent attempt journal before any call. The new journal is:

`results/BTS-FAST-NOTRACE-VALIDATION-003-journal.md`

The new raw log and report are:

- `results/BTS-FAST-NOTRACE-VALIDATION-003-attempt-01.log`
- `results/BTS-FAST-NOTRACE-VALIDATION-003-summary.md`

This task authorizes **at most one actual `bts.Bootstrap` invocation total**. Record a durable exclusive reservation before entering `bts.Bootstrap`; record whether the process entered Bootstrap and never retry a call that started, failed, panicked, or was interrupted. A pre-Bootstrap build/setup failure consumes no invocation but is not permission to alter the frozen workload or invoke Bootstrap without a passing code gate. This budget is separate from, and does not replenish or edit, task 002's journal.

Immediately before the call, verify BTS source manifest/SHA-256 is unchanged, record Go/OS/architecture, exact build/run commands, Fast branch and commit, module replacement and binary module provenance. Preserve complete stdout/stderr. Report `NewBTS()` initialization time separately from Bootstrap time, input/output Level and Scale, Bootstrap completion, and the original BTS max/mean absolute error and max-error slot when available. Do not invent an acceptance tolerance or claim speedup; the existing Standard baseline (`139.653 s`, max error `1.708e-3`, mean error `1.172e-4`) is historical and must not be rerun.

If the one call fails, preserve the complete log and stop. Do not add a hidden Standard fallback, alter BTS source, retry, or weaken numerical expectations.

## Git and handoff

Commit and normally fast-forward push the new Primary task pointer/spec before Secondary implementation. Commit and push only authorized Secondary CKKS source/tests after all pre-call gates pass. After the one-call result, commit and normally push only the new Primary journal, raw log, and concise report; never include BTS source or binaries. Do not force push, rewrite history, reset, rebase, stash, or overwrite unrelated work.

Successful handoff status: `READY_FOR_WEB_REVIEW`. A failed single call remains a completed bounded experiment with failure evidence for Web review; unresolved math/representation/capacity conflict is `NEEDS_WEB_REVIEW`.
