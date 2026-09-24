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
