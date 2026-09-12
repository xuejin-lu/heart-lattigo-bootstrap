# FIX-001-P3-DIAG-PS-GLOBAL-SEMANTICS — Locate the first source-polynomial semantic drift inside compressed PS

## Purpose

The attempted final-Fast-Rescale diagnostic stopped correctly with:

`FINAL_RESCALE_PRECONDITION_MISMATCH`

Observed on the canonical LogN13 `2^91` path:

- pre-final-Rescale whole-polynomial error (`R0`): `0.01144397857243229`;
- post-final-Rescale whole-polynomial error (`R7`): `0.011443985451588534`;
- the R0 maintained q0/q1 hashes exactly match the prior normalization task's `before_rescale` state;
- therefore the previously cited `~4.13e-8` checkpoint cannot be treated as the whole-polynomial pre-final-Rescale error.

The preceding normalization spec required local semantic oracles throughout the giant tree, but required a source-backed whole degree-30 polynomial oracle only at finalization. The current evidence therefore supports a narrower interpretation:

> local PS arithmetic remained self-consistent, while a source-polynomial semantic discrepancy was already present before the final Rescale.

This task must find the **first PS checkpoint where a source-backed global/subtree oracle diverges**, without changing production Lattigo.

Diagnostic only. LogN13 only.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`d6179b1a39a2a9b3e50e0432304e171021ef80f2`

Authoritative normalization result commit:

`0292d11edfa5bfee647bfd3a04d1f395c805ab96`

Secondary repository: `xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Both worktrees must start clean and follow their `AGENTS.md` startup rules.

Do not modify Secondary production code.

Parent FIX-001/P3 remains incomplete.

---

# Scope lock

Do not:

- run LogN16;
- benchmark;
- run formal Gate 4/5 production experiments;
- start EXP-003;
- inspect final Fast Rescale internals further in this task;
- modify production Lattigo;
- change polynomial coefficients, basis, PS schedule, candidate scales, semantic threshold, generated powers, or metadata-normalization rule;
- treat local arithmetic self-consistency as proof of source-polynomial correctness;
- use decoded parent/child ciphertext values as the only expected value for a node;
- read full large `results/*.json` artifacts.

The only question is:

> At which baby/giant/final-accumulation checkpoint does the ciphertext first cease to represent the exact source-polynomial subtree that this PS node is supposed to represent?

---

# Mandatory reproduction

Use the exact LogN13 real-branch formal degree-30 polynomial workload, corrected `T0=1` oracle, compressed PS schedule, and giant-step metadata normalization from the completed normalization task.

Use `2^91` as the canonical probe.

Before deeper localization, reproduce:

1. P3 generated powers `T1,T2,T3,T4,T6,T8,T16` all pass the existing `1e-2` source-backed power oracles;
2. all baby blocks and normalized giant steps still pass their historical local semantic checks;
3. the exact pre-final-Rescale q0/q1 hashes match the authoritative normalization `before_rescale` hashes;
4. a direct whole degree-30 source-polynomial oracle at that exact pre-final state gives approximately `0.0114439786` and therefore fails `1e-2`;
5. no final Rescale is needed to reproduce the failure.

If these do not reproduce, stop with:

`PS_GLOBAL_SEMANTICS_PRECONDITION_MISMATCH`.

Do not tune the workload to force reproduction.

---

# Oracle independence rule

Every checkpoint in this task must have two conceptually different measurements:

## A. Local consistency oracle

This is the historical operation-local check, such as:

- baby operation result versus decoded operands under the operation just executed;
- `Rescale(b)` versus local rescale expectation;
- `Mul(b,xpow)` versus decoded local operands;
- `A + B*X` versus decoded local operands.

Preserve these only to show whether the arithmetic step is locally self-consistent.

## B. Source-backed subtree oracle

This is the authoritative new check.

For each PS node, construct the exact cleartext polynomial subtree that node is intended to represent from:

- the original degree-30 source coefficients/basis;
- the original input slots;
- the exact PS decomposition / block-to-term mapping.

Evaluate this subtree independently in plaintext/high precision.

The expected value must **not** be derived from decoded ciphertext children or from the ciphertext node being tested.

For every node, retain a compact identity of the expected subtree:

- basis/type;
- degree or term-index range/set;
- coefficient count;
- deterministic coefficient/subtree hash.

Do not serialize full per-slot vectors or full coefficient arrays.

For the final root, additionally evaluate the full source polynomial through the existing independent direct whole-polynomial oracle and require the two plaintext oracle constructions to agree tightly. If the PS-subtree oracle and direct full-polynomial oracle disagree materially in plaintext, stop with:

`PS_PLAINTEXT_ORACLE_CONSTRUCTION_MISMATCH`.

This is a harness/oracle problem, not a Fast arithmetic result.

---

# Checkpoints

Diagnose the canonical `2^91` path in PS execution order. Do not jump directly to the final root.

For every checkpoint record:

- checkpoint ID / round / block or node indices;
- operation that created the node;
- Level/Degree/Scale;
- local consistency max component error;
- source-backed subtree max component error;
- q0/q1 hashes;
- expected subtree identity/hash;
- metadata normalization applied yes/no;
- before/after normalization source-backed subtree error when applicable.

Use fixed semantic threshold `1e-2`.

## G0 — generated powers

Reconfirm source-backed oracles for:

`T1,T2,T3,T4,T6,T8,T16`.

If one fails first, classify:

`ps_global_generated_power_failure`

and stop causal interpretation there.

## B — baby blocks

For each of the five baby blocks, compare the ciphertext block directly against the exact source-polynomial block it is intended to represent.

A block is globally passing only if its source-backed subtree error <= `1e-2`, regardless of whether its historical local oracle passes.

If the first global failure appears here, identify whether it is already present:

1. immediately after coefficient/baby-block construction;
2. after a baby multiplication;
3. after a baby add;
4. after any baby rescale/relinearize used by the actual schedule.

Classify the first failed operation boundary as:

- `ps_global_baby_construction_failure`
- `ps_global_baby_multiply_failure`
- `ps_global_baby_add_failure`
- `ps_global_baby_rescale_failure`

Do not proceed to giant-step causal claims if a baby block is already globally wrong.

## G — giant tree

If all baby blocks are source-backed correct, follow the exact historical reversal/pairing schedule.

At every giant node evaluate source-backed semantics at these boundaries when present:

1. input accumulator `a`;
2. input branch `b`;
3. after `Relinearize(b)`;
4. after `Rescale(b)`;
5. after `Mul(b,xpow)`;
6. immediately before metadata normalization;
7. immediately after metadata-only normalization;
8. after final Add / `addAligned` creating the parent node.

The source-backed oracle for each boundary must correspond to the exact subtree represented at that boundary. For metadata-only normalization, the intended source subtree is unchanged.

If local consistency stays <= `1e-2` but source-backed error first exceeds `1e-2`, that is precisely the boundary this task is intended to expose.

Classify the first failing boundary as one of:

- `ps_global_giant_input_already_wrong`
- `ps_global_giant_relinearize_failure`
- `ps_global_giant_rescale_failure`
- `ps_global_giant_multiply_failure`
- `ps_global_giant_metadata_normalization_failure`
- `ps_global_giant_add_failure`

## F0 — pre-final-Rescale root

If no earlier checkpoint exceeds `1e-2`, evaluate the exact root immediately before final Rescale against:

1. the independently constructed PS-root plaintext oracle;
2. the direct whole degree-30 source-polynomial oracle.

These plaintext expected values must agree tightly with each other.

Record the exact pre-final error and q0/q1 hashes.

If the root is the first checkpoint above `1e-2` despite all immediately preceding child nodes passing, classify:

`ps_global_root_assembly_failure`

Do not enter final Rescale internals in this task.

---

# Error-growth evidence

Because the final error is only slightly above the `1e-2` threshold, do not report only pass/fail.

For every source-backed checkpoint also record compact error-growth information:

- absolute max component error;
- ratio to the preceding source-backed checkpoint error when mathematically meaningful;
- delta in max error across the creating operation;
- worst slot/component index and expected/actual values;
- whether the worst slot identity changes.

If no single operation jumps directly above `1e-2` and instead error grows gradually, report the **first checkpoint at which the source-backed error materially separates from the local oracle**, plus the first threshold-crossing checkpoint.

Do not manufacture a single-operation bug classification when the evidence supports accumulated drift.

Allowed accumulated-drift classification:

`ps_global_accumulated_drift_without_single_step_failure`

Use it only if all individual operation-local comparisons remain correct, source-backed error grows across multiple nodes, and no single operation boundary can be supported as the first incorrect transformation.

---

# Canonical-vs-confirmation probe

After the `2^91` first global failure boundary is located, run only `2^86` as a confirmation probe through that same boundary.

Record:

- whether the same checkpoint is the first source-backed failure;
- source-backed error at the checkpoint immediately before and at the failing boundary;
- whether the same worst slot/component is implicated;
- whether the failure magnitude materially depends on internal scale.

If `2^91` and `2^86` disagree on the first global boundary, classify:

`ps_global_scale_dependent_boundary_disagreement`

and stop. Do not automatically expand to all six candidates.

---

# Overall classification

Choose exactly one evidence-supported result:

- `FIRST_SUPPORTED_CAUSE = ps_global_generated_power_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_baby_construction_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_baby_multiply_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_baby_add_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_baby_rescale_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_giant_input_already_wrong`
- `FIRST_SUPPORTED_CAUSE = ps_global_giant_relinearize_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_giant_rescale_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_giant_multiply_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_giant_metadata_normalization_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_giant_add_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_root_assembly_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_accumulated_drift_without_single_step_failure`
- `FIRST_SUPPORTED_CAUSE = ps_global_scale_dependent_boundary_disagreement`
- `FIRST_SUPPORTED_CAUSE = ps_plaintext_oracle_construction_mismatch`
- `FIRST_SUPPORTED_CAUSE = ps_global_semantics_precondition_mismatch`

Do not classify final Fast Rescale as the first cause unless a later task establishes a correct pre-final whole-polynomial state first.

---

# Artifact size discipline

Create compact artifacts only:

- `results/FIX-001-P3-DIAG-PS-GLOBAL-SEMANTICS-logN13.json`
- `results/FIX-001-P3-DIAG-PS-GLOBAL-SEMANTICS-logN13-summary.json`

Do not serialize:

- full slot vectors;
- all coefficients;
- repeated indices;
- full operation traces;
- large hash collections.

For each checkpoint store only compact evidence:

- IDs and operation;
- subtree identity/hash;
- local/global errors;
- relevant Scale/Level/Degree;
- q0/q1 aggregate hashes;
- worst slot/component tuple;
- mismatch or threshold status.

The summary must remain human-reviewable and include:

- exact provenance;
- explicit explanation that historical `~4.13e-8` was a local checkpoint and is not established as pre-final whole-polynomial error;
- canonical `2^91` checkpoint table;
- `2^86` confirmation table;
- first source-backed divergence;
- first `1e-2` threshold crossing;
- local-vs-source-backed error comparison;
- final overall classification;
- test and clean-state evidence.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- focused diagnostic tests pass;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no production Lattigo changes;
- Primary compact diagnostic artifacts are committed and pushed normally;
- both worktrees are clean;
- no LogN16;
- no benchmark;
- no Gate 4/5 production run;
- no EXP-003.

The deliverable is the earliest source-backed PS semantic divergence, not another local-consistency proof.