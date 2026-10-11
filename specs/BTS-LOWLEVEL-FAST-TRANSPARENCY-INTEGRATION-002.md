# BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002

Status: READY_FOR_CODEX
Class: bounded implementation plus one bounded No-Trace experiment (I/E)
Authority: Web-approved architecture decision in the current task handoff, Primary `AGENTS.md`, `docs/BACKEND_TRANSPARENCY_CONTRACT.md`, `docs/MEASUREMENT_PLATFORM.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md`, Secondary `AGENTS.md`, and `docs/FAST_QPREFIX_SPEC.md`.

## Objective

Integrate the unchanged research-lab `external/BTS/` frontend with the Fast `fast-qprefix` backend through its existing ordinary public API. Preserve the Fast Q-prefix and zero-secret computational simplifications while supplying the BTS-required low-level RLWE/QP API semantics. This is a bounded compatibility integration, not a ring rewrite, a return to Standard full Key Switching, or a new measurement platform.

The original BTS source, parameters, input generation, execution mode, and operation order must remain unchanged. Do not modify, stage, commit, or publish `external/BTS/`.

`FAST-Q0123-RESCALE-EXACT-DIVISION-PILOT-006` remains suspended. The historical BTS Standard/Fast baseline and its raw logs remain immutable evidence. Do not rerun Standard or count the previous failed Fast baseline attempt against this task's separately authorized budget.

## Frozen workload and attempt budget

- Run the original BTS defaults only: LogN16, SlotLimbs=15, `BTSModePlain`, amp=1, seed=0, and the original Q/P chain, scale, input, and Bootstrap sequence.
- Build the unchanged BTS module against the completed Fast `fast-qprefix` commit using an isolated temporary Go `-modfile`, `GOWORK=off`, and an explicit local module replacement. Do not edit BTS or Primary's committed `go.mod`.
- Target one complete **No-Trace** Fast BTS run. The durable task-wide journal may authorize at most two new actual `bts.Bootstrap` invocations across all Codex sessions. Reserve each invocation before entering it; record launch, process state, and whether Bootstrap began. No warmups, traces, microbenchmarks, or unrelated Bootstrap calls. Never repeat a call that began and failed, panicked, or was interrupted. A second call is permitted only if the first is conclusively shown not to have entered `bts.Bootstrap` and the journal still has budget.
- Do not rerun the completed Standard baseline. Its accepted historical values are Bootstrap 139.653 seconds, max absolute error 1.708e-3, and mean absolute error 1.172e-4. Keep these labelled as the existing Standard baseline, not a new matched run.
- Before each permitted Fast invocation, verify the original BTS source identity and the Fast module commit/build provenance. Save complete stdout/stderr for every launch. Separate `NewBTS()` initialization time from `bts.Bootstrap()` time using the original program's own reporting.
- Do not invent a numerical tolerance or claim a speedup unless both outputs satisfy an already-authorized acceptance criterion. Report the original BTS error metrics and all available output metadata even if Fast fails.

## Frozen mathematical and representation rules

1. Fast evaluation keys retain the existing `a=0` design. The Fast zero-secret mode retains the `c1=0` invariant. Verify the invariant at every relevant public/low-level API boundary; missing or unallocated storage is not, by itself, proof of a zero polynomial. Only omit Decompose/GadgetProduct/KeySwitch arithmetic when the skipped operand is proven zero under the active representation. Nonzero or malformed operands must fail closed through a safe API error path; never silently skip their contribution.
2. Rotation must still perform the required Galois automorphism. `a=0` does not authorize omitting that permutation.
3. BTS-visible lazy QP results must preserve the mathematically correct P-scaled representation, metadata, NTT/Montgomery domain, and subsequent `ModDownQPtoQNTT` semantics. Returning an ordinary Q ciphertext in place of a lazy QP result is not compatible.
4. Fast storage follows `docs/FAST_QPREFIX_SPEC.md`: at logical Level `ell`, the authoritative Q rows are exactly the policy-defined prefix `w_Q(ell)=min(ell+1,4)`. Dormant rows are never inputs. Reconstruct a dormant row only at an explicitly justified representation boundary and only after proving the required centered-lift condition `2B < S_Q`; otherwise return a precise capacity/representation failure before consuming the row.
5. Do not pass stale/unmaintained residues to Standard full-RNS code. A conditional Standard-compatible fallback is allowed only when its input rows and representation are proven complete and authoritative. It must not introduce unnecessary full-QP work into the normal BSGS, Rotation, LinearTransform, or Key Switching hot paths.
6. Do not change Standard Lattigo, other schemes, BTS source, frontend math, parameters, Q-prefix policy, security semantics, or numerical gates. No private-F storage, parameter tuning, direct ciphertext rewriting, or frontend backend branch.

## Required source/reuse audit and implementation scope

Trace the actual BTS call chain and distinguish integration/dispatch/representation gaps from missing arithmetic. Inspect the ordinary key-generation path and the BTS-used low-level operations, including `DecomposeNTT`, `AutomorphismHoistedLazy`, `GadgetProductLazy`, `ModDownQPtoQNTT`, and `ApplyEvaluationKey`.

Reuse existing Fast DFT, BSGS, Rotation, LinearTransform, Rescale, Q-prefix arithmetic, evaluator scratch, and zero-secret validation wherever their contracts apply. Keep changes CKKS-scoped in Secondary `fast-qprefix`. Implement only the smallest compatibility bridge needed so the unchanged BTS calls receive correct semantics. Do not create a second Fast arithmetic implementation when an existing Fast kernel can be safely dispatched or adapted.

For every adapted operation, document in code/tests or the final evidence which path is used and why. Prove the intended Fast kernel/shortcut actually executed; an output-only pass is insufficient to establish reuse or absence of fallback.

If a BTS operation cannot be represented correctly within the approved Q-prefix policy and existing capacity proof, stop that mathematical fix. Preserve evidence with the exact API/checkpoint, required representation, authoritative rows available, bound/capacity if applicable, and a minimal counterexample. Do not invent a new Q-prefix rule or hidden full-RNS workaround.

## Required targeted validation before Bootstrap

Add or extend focused Secondary regression tests for the contracts actually touched. Tests must avoid calling Bootstrap and must cover, as applicable:

- ordinary public key generation/evaluator dispatch produces the Fast `a=0` evaluation-key form;
- zero-secret validation accepts a structurally valid zero `c1`, rejects nonzero/malformed `c1`, and does not infer zero from absent storage;
- zero-operand Decompose/GadgetProduct elision preserves output shape, Level, Scale, and domain metadata;
- Galois automorphism is still applied and the lazy QP output retains P scaling;
- the lazy QP output followed by `ModDownQPtoQNTT` agrees with an independently constructed valid reference;
- no operation reads dormant Q rows, and any required reconstruction is guarded by the authoritative `2B < S_Q` proof;
- no unverified Standard full-RNS fallback is reached on the normal Fast path.

Use existing Secondary package tests and helpers; do not add a standalone benchmark, trace platform, or parallel oracle framework. Run the focused CKKS Fast/key-switch/bootstrapping tests selected by the implementation, plus the relevant broader Secondary test suite required by its `AGENTS.md`. Record exact commands and results. Run Primary tests only where the Primary task artifacts or existing shared harness code require them; do not use tests that implicitly invoke extra Bootstrap calls without checking their call budget.

The implementation must pass targeted validation and a coding self-review before any BTS Bootstrap call is reserved. If a test exposes a new mathematical or representation ambiguity beyond this approved contract, stop with `NEEDS_WEB_REVIEW` and precise evidence.

## No-Trace end-to-end validation and artifacts

After the backend changes are committed and pushed to `origin/fast-qprefix`, compile the unchanged BTS source against that exact commit. Verify the module replacement and binary provenance; do not claim Fast if the binary does not link the repaired Fast commit.

Run only the authorized No-Trace Fast BTS attempt(s). Preserve complete raw logs and a compact task journal/report under Primary `results/`, excluding all BTS source and binaries. Use the existing original BTS initialization, Bootstrap timing, decryption, and error calculations; do not build a replacement runner or duplicate metric implementation.

Record at minimum:

- BTS source file manifest/hash and confirmation the source was untouched;
- Go version, OS/architecture, exact build/run commands, Fast branch/commit, module replacement, and binary module provenance;
- `NewBTS()` initialization result/time and first complete Bootstrap result/time;
- input/output Level and Scale;
- max and mean absolute error and max-error slot when an output exists;
- exact failure/API and attempt-journal state otherwise;
- whether output completed and whether it meets the original BTS expectation, without inventing a threshold;
- a categorized operation map: (a) mathematically elided, (b) existing Fast implementation reused, (c) QP representation conversion/adaptation, and (d) verified Standard-compatible fallback, including explicit “none” where applicable;
- comparison against the historical Standard baseline, clearly noting it is not a rerun and withholding speedup unless the authorized correctness criteria are met.

## Git and completion

- Primary first commits/pushes this specification and its `CURRENT_TASK.md` pointer using a normal fast-forward only. Do not begin Secondary implementation until Primary's new task authority is present on `origin/main`.
- Secondary work is limited to authorized CKKS Fast compatibility changes and tests on `fast-qprefix`; commit and normally push to `origin/fast-qprefix` only after focused validation and self-review. Never force push or rewrite history.
- Primary completion artifacts may include the compact summary, durable attempt journal, and complete Fast raw logs. Never include or publish `external/BTS/` source, original BTS binaries, or unrelated work.
- Preserve existing local work, all historical results, and Pilot006's suspended status. If either repository has unknown dirty changes, divergent history, an unsafe push condition, or a real capacity/math blocker, stop without destructive Git operations and report exact evidence.
- Successful completion status: `READY_FOR_WEB_REVIEW`. A mathematical/Q-prefix capacity blocker or unresolved API semantic conflict: `NEEDS_WEB_REVIEW`.
