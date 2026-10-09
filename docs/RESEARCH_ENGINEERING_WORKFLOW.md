# Fast-CKKS Research Engineering Workflow

This document defines the durable division of labor between GPT Web, Codex, and the repository. It is meant to survive individual chats, model changes, and personnel handoffs.

## 1. Core principle

This project contains two different kinds of work:

1. scientific/mathematical reasoning;
2. software implementation and execution.

They must not be treated as the same review problem.

The default authority chain is:

```text
GPT Web: reason about mathematics / architecture / scientific interpretation
        ↓
GitHub spec: encode the decision as explicit invariants and acceptance criteria
        ↓
Codex: implement and test the authorized scope
        ↓
Codex: perform one bounded coding self-review + repair pass
        ↓
GPT Web: independently review scientific correctness and repository evidence
```

In short:

```text
Web thinks
→ Spec freezes the contract
→ Codex builds
→ Codex checks implementation quality
→ Web performs scientific acceptance
```

## 2. Task classes

Every nontrivial task should be treated as one of three classes.

### M — Mathematics / Architecture

Examples:

- deriving a CKKS invariant;
- deciding the correct Rescale semantics;
- proving a storage-capacity condition;
- deciding KeySwitch / Relinearize algebra;
- interpreting a failure causally;
- choosing whether two representations are mathematically equivalent.

Primary authority:

```text
GPT Web
```

Codex may:

- inspect source;
- gather runtime evidence;
- run controlled experiments;
- verify an already-defined invariant;
- report conflicts.

Codex must not independently resolve a new mathematical or cryptographic semantic ambiguity merely because one implementation choice is convenient.

### I — Implementation

Examples:

- implementing an already-approved oracle;
- wiring measurement infrastructure;
- adding typed schemas;
- implementing a previously-specified arithmetic primitive;
- adding tests for a frozen invariant.

Primary execution authority:

```text
Codex
```

Normal cycle:

```text
implementation pass
→ coding self-review
→ one bounded repair pass
→ final tests
→ commit/push
→ READY_FOR_WEB_REVIEW
```

GPT Web still performs final independent acceptance.

### E — Experiment / Measurement

Examples:

- running a parameter sweep;
- collecting scale/capacity traces;
- measuring Fast vs Standard;
- generating distribution summaries;
- reproducing a failure.

Codex owns:

- reproducible execution;
- provenance;
- artifact generation;
- mechanical validity checks.

GPT Web owns:

- interpretation;
- causal conclusions;
- deciding whether evidence supports a repair;
- choosing the next experiment.

A measurement result is evidence, not automatically a conclusion.

## 3. Review responsibilities

### Codex self-review

Purpose:

- catch coding defects before handoff.

It should check:

- spec compliance;
- complete task diff;
- acceptance criteria;
- tests;
- edge cases;
- error handling;
- unintended scope expansion;
- stale assumptions.

It is a coding-quality gate.

It is **not** independent scientific validation.

### GPT Web scientific review

Purpose:

- decide whether the result is scientifically and mathematically acceptable.

It should check:

- whether the mathematical contract itself is sound;
- whether the implementation matches that contract;
- whether the experiment measures what it claims;
- whether evidence supports the causal conclusion;
- whether there are confounders or stale provenance;
- whether the next task is justified.

GPT Web must review repository evidence independently rather than accepting a Codex completion report at face value.

## 4. Escalation rules

Codex must stop and return `NEEDS_WEB_REVIEW` when:

- a new mathematical or cryptographic semantic decision is required;
- source evidence conflicts with the active spec;
- the durable architecture and implementation assumptions disagree;
- the same unexplained failure remains after the bounded repair pass;
- the only apparent fix requires broader scope than authorized;
- a Primary-only task appears to require Secondary changes or vice versa;
- a threshold, parameter, or invariant would need to be weakened;
- available evidence cannot distinguish implementation defect from stale diagnostic/provenance debt;
- safe Git conditions fail.

Do not keep iterating autonomously past these boundaries.

## 4A. Reuse-first architecture gate (mandatory for existing capability integration)

**Why this exists:** In October 2026, Fast CKKS already had explicit implementations of Add, MulRelin, Rescale and Rotate in `schemes/ckks/fast`, but the orchestration incorrectly commissioned additional ordinary-public-`ckks.Evaluator` primitive implementations without first making a repository-wide integration decision. A local test of numerical correctness does not demonstrate that an existing optimized kernel has been reused or even invoked.

**Before GPT Web authorizes an implementation spec** for an API, primitive, backend replacement, or integration feature, it must inspect **current source** and record a compact reuse inventory:

1. The **user's end goal** (e.g. identical CKKS frontend source and parameters, only library dependency swapped), distinct from the next local test.
2. Existing implementations and call sites, with **files/functions and their verified capabilities**, including parallel or legacy variants, and which code is actually dispatched.
3. The gap classification: `MISSING_CORE`, `INTEGRATION_DISPATCH`, `REPRESENTATION_CONTRACT`, `API_ADAPTER`, `GENUINE_UNSUPPORTED_CASE`, or `UNKNOWN`. Do not call an integration issue a missing algorithm.
4. A **reuse-first design** with candidate dependency direction, Go concrete types/signatures, interface/import-cycle constraints, existing shared helpers, and Level/Scale/c1/Q-prefix/full-active-Q/NTT/Montgomery/key-layout invariants. Prefer an adapter or moving a truly shared kernel into an import-neutral lower-level package, only if mathematical and API semantics remain valid; do not require a specific architecture before source analysis.
5. A minimal execution/verification plan that **proves which kernel actually ran** (dispatch/key-lookup instrumentation or equivalent), separately from cleartext-oracle correctness, and states explicitly what remains unproved.

**Default ordering:** reuse already-proved kernel -> minimal API bridge/adapter -> source-backed extraction of shared kernel -> **new algorithm implementation only when an existing kernel demonstrably cannot satisfy the contract**. Duplicating an existing arithmetic kernel requires an explicit written technical reason and Web authorization; a failing constructor dispatch or different Go method signature is not sufficient reason by itself.

**Mandatory STOP gate:** If a source audit finds existing functionality with a plausible reuse path, a Go import cycle, or incompatible full-Q vs compact Q-prefix representations, do **not** issue a piecemeal reimplementation spec. Classify `NEEDS_WEB_ARCHITECTURE_REVIEW`, compare the feasible architectures, and approve an integration contract **before** implementation. Codex must report the discovery and stop rather than implement around it.

**Before Web accepts results:** check four separate questions: (a) numerical output correct for supported inputs; (b) proven execution of intended Fast kernel/algorithm; (c) no forbidden native fallback or representation violation; (d) compatibility with the original *end-to-end* goal. Pass on (a) alone is `NUMERICAL_ONLY`, not `FAST_BACKEND_INTEGRATED`.

**Scope discipline:** This gate is a brief source-backed check, not an excuse for unbounded audits or exhaustive historical research. If the same inventory remains valid, link to it in subsequent tasks rather than redo it. Codex self-review must flag any new implementation that duplicates a discovered existing backend primitive. This rule supplements the current M/I/E authority model and preserves the user's simple `開始` / `review` workflow.

## 4B. Autonomous milestone batches (default after architecture is frozen)

**Reason:** The former `one small task → independent Web review → next task` handoff created unnecessary human coordination and repeatedly blocked routine engineering progress. Independent Web review is required at **scientific milestones**, not after every mechanical implementation or test fix.

**Default handoff model:**

```text
Web: approve mathematical/representation architecture + bounded batch charter
  ↓
Codex: subtask A → test → coding self-review → repair → commit
  ↓
Codex: subtask B → test → coding self-review → repair → commit
  ↓
Codex: subtask C / D → test → self-review → final aggregate report → push
  ↓
Web: ONE independent scientific / architectural milestone review
```

- A batch should normally include **3–5 closely related implementation/experiment subtasks**, or one coherent scientific integration milestone with 3–5 checkpoints. This is a guideline, **not a requirement to create filler tasks**, not authorization for unlimited autonomous changes, and not an excuse to delay a genuine stop condition.
- GPT Web owns the **batch charter**: original end goal, frozen M invariants, candidate scope, pre-approved I/E actions, test oracles, permitted repos/paths, upper bounds on expensive experiments, task dependencies, and precise STOP conditions. Codex may choose the ordering and detailed next I/E implementation **inside that charter** based on test results, including routine local fixes and follow-up tests. It must not invent new cryptographic semantics or broaden the charters itself.
- The first local subtask can be an existing independently approved spec. Codex should read it but follow any **explicitly specified** batch-level extension/supersession from a new charter. Preserve old specs and results as immutable history.
- **Per-subtask Codex responsibility:** implement or execute; verify the test oracle; inspect source/diff and edge cases; at most one *local repair pass* for the same issue; rerun affected tests; commit as an atomic milestone only if valid. It may continue autonomously to the next authorized subtask **without waiting for GPT Web**. Use repository-safe fast-forward pushes at coherent checkpoints or sprint end; never change refs forcefully.
- **Scientific review boundaries:** local coding self-review cannot accept new math, Q-prefix/full-Q representation contracts, key-layout semantics, crypto/security claims, changed parameters or failed numerical oracles. A bounded batch cannot silently hide a failure and move on to unrelated work. If blocked and no safe approved fallback, stop with `NEEDS_WEB_REVIEW`, comprehensive evidence and the exact decision sought. Web also reviews at batch completion, even if all Codex tests pass.
- **No counterfeit provenance:** record separate backend commits, exact frontend/config/input hashes, Standard genuine encryption, Fast zero-secret semantics, measurement methods and the **kernel that actually executed**. A passing plaintext oracle does not prove optimized Fast dispatch. Preserve the distinction between numerical correctness, architecture integration and runtime speedup.
- **Expensive work:** batch specs explicitly cap Bootstrap counts, heavyweight tests, benchmarks, parameter sweeps and retries **across the entire sprint**, not per subtask by accident. Do not rerun Bootstrap only to obtain more favorable metrics. Ordinary cheap tests may be rerun for genuine implementation repairs within the approved charter.
- **Recoverability:** Codex writes a compact batch journal under `results/` with each subtask's status, commit, verification, attempt counts and next action. If the Codex session ends before the milestone is finished, report `BATCH_IN_PROGRESS` and journal a safe resumable checkpoint; on the next user `開始`, safely sync and continue **without repeating completed heavy work**. Never claim to run asynchronously between turns.
- **Web review cadence:** after a completed batch/milestone, or immediately for an unresolved M/architecture/cross-repository decision, incompatible Standard baseline, persistent numeric failure, unsafe Git operation, or scope overrun. Web should perform ONE integrated review covering architecture, re-use, numerical correctness, true Fast dispatch, cost and next research direction, and should avoid redundant line-by-line reviews of routine source changes.

**Status vocabulary:** `BATCH_IN_PROGRESS`, `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`, `BATCH_BLOCKED_NEEDS_WEB_REVIEW`. Existing one-task statuses remain valid for explicitly single-task exceptions. In the batch mode, `READY_FOR_WEB_REVIEW` applies at the **batch boundary**, not after each successful subtask.

## 5. Executable mathematics

Whenever possible, GPT Web should convert mathematical decisions into executable contracts before implementation.

Example:

Mathematical statement:

[
Y=\operatorname{Round}(X/q_\ell)
]

with

[
\Delta'=\Delta/q_\ell.
]

Executable contract:

```text
expected_divisor = logical_q[level]
expected_level_delta = -1
expected_scale_rule = / logical_q[level]
storage_width_change = independent
```

This reduces the amount of mathematical judgment required from the implementation agent.

The preferred progression is:

```text
theorem / invariant
→ spec
→ oracle / assertion
→ production implementation
```

not:

```text
implementation guess
→ debug
→ infer mathematics afterward
```

## 6. Repository as durable memory

Chat history is not the project authority.

Durable knowledge belongs in GitHub:

- `AGENTS.md`: permanent execution rules;
- `docs/FAST_CKKS_SPEC.md`: durable Fast-CKKS mathematical/architectural constitution;
- this document: division of labor and review workflow;
- `specs/*.md`: one bounded task contract;
- `CURRENT_TASK.md`: pointer to the active task only;
- `docs/ORCHESTRATOR_HANDOFF.md`: current scientific/project handoff;
- `results/`: experiment evidence.

A new agent should be able to recover the workflow without reading old chat transcripts.

## 7. Default command workflow

The user-facing workflow should remain minimal.

### Start

The user tells Codex:

`開始`

Codex then:

1. safely syncs;
2. reads durable instructions and current task;
3. executes the bounded implementation/review cycle;
4. commits and pushes if authorized and safe;
5. returns either:
   - `READY_FOR_WEB_REVIEW`, or
   - `NEEDS_WEB_REVIEW`.

### Review

The user tells GPT Web:

`review`

GPT Web then:

1. independently reads remote GitHub evidence;
2. verifies implementation/results;
3. performs scientific review;
4. either accepts the task or writes the next bounded spec.

The user should not need to manually relay large technical state between the two agents.

## 8. Workflow freeze rule

Do not redesign this agent workflow preemptively.

Change it only when a concrete execution failure demonstrates a real problem, such as:

- repeated unnecessary handoffs;
- scientific decisions being made at the wrong layer;
- stale state causing incorrect work;
- self-review repeatedly missing the same implementation defect;
- artifacts being too large or inconsistent for independent review.

Until such evidence exists, invest effort in the research itself rather than further orchestration design.
