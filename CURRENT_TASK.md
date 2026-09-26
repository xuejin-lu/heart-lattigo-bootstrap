# Current Task

Task: QPREFIX-IMPL-001
Status: READY_FOR_CODEX

Architecture plan:
`docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

Task class:
`I — Implementation`

Repositories:
- Primary `xuejin-lu/heart-lattigo-bootstrap@main`: orchestration/spec only.
- Secondary `xuejin-lu/lattigo@fast-qprefix`: implementation target.

Goal:
Create the single authoritative Q-prefix policy and capacity contract for Q-prefix v2.

Frozen policy:
[
w_Q(ell)=min(ell+1,4)
]

Equivalent maintained prefixes:
- Level >= 3: q0,q1,q2,q3
- Level 2: q0,q1,q2
- Level 1: q0,q1
- Level 0: q0

Scope:
- one source of truth for Level -> maintained prefix width;
- exact prefix-product computation from actual q values;
- per-component bound/capacity helpers;
- strict centered-capacity predicate `2B < S_Q`;
- explicit/recognizable transactional capacity error form;
- tests for Level 0/1/2/3/high-Level policy and strict boundary behavior.

Do not yet:
- switch constructors/Resize/scratch to the new policy;
- widen arithmetic primitives;
- modify Rescale/ModUp/DFT/EvalMod/Bootstrap;
- add F;
- add adaptive width;
- modify `fast-ckks`;
- run broad production migration.

Capacity proof is an acceptance condition for each later milestone, not a global prerequisite for starting implementation.

Commit/push Secondary `fast-qprefix`, then report `READY_FOR_WEB_REVIEW`.
