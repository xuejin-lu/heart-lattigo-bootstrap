# QPREFIX-AUDIT-001 — Full Bootstrap Capacity Audit Before Reimplementation

## Status

Executable Codex audit task.

## Task class

`M/E — Architecture audit + reproducible measurement`

No production algorithm change is authorized.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- owns orchestration/report.

Secondary:
- `xuejin-lu/lattigo@fast-qprefix`
- branch point: `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- authoritative architecture:
  `docs/FAST_QPREFIX_SPEC.md`
  at `c9fb900314a1c82f2cd21320554d3d47fd7dbf00`

The existing `fast-ckks` branch is comparison evidence only. Do not modify it.

## Purpose

Determine whether the independent private-F storage architecture is unnecessary for the current target LogN13/P93 Bootstrap profile.

The Q-prefix v2 policy is:

[
A(ell)={q_0,ldots,q_{min(ell,3)}}
]

with maintained-row count:

[
w_Q(ell)=min(ell+1,4).
]

At each production stage that requires an authoritative integer lift, prove:

[
2B_j<S_Q(ell)
]

where:

[
S_Q(ell)=prod_{i=0}^{min(ell,3)}q_i.
]

No F basis exists on this branch.

---

# 1. Fixed profile

Use the exact accepted LogN13/P93 profile represented by branch-point history and the current Primary reproducible workload.

Record exact:
- q0, q1, q2, q3 values;
- bit lengths;
- MaxLevel;
- ring N;
- current public Bootstrap milestone/error control.

Do not retune parameters.

---

# 2. Audit scope

At minimum inspect and report these boundaries:

1. Bootstrap input before ScaleDown.
2. ScaleDown intermediate level drops where relevant.
3. Level-0 ScaleDown result.
4. ModUp canonical Level-0 representative.
5. post-ModUp maintained q-prefix.
6. Trace input/output.
7. each CoeffsToSlots factor:
   - input;
   - LinearTransform output;
   - Rescale output;
   - restore output.
8. EvalMod:
   - normalization;
   - generated powers;
   - each critical PS multiply/rescale/add boundary;
   - polynomial output;
   - each DoubleAngle round/window.
9. SlotsToCoeffs:
   - each factor input/output/rescale/restore.
10. packing/unpacking/ring-degree boundaries.
11. finalization/residual level.

Use actual execution order from the pre-F branch. Do not infer stage order only from names.

---

# 3. Bound sources

Prefer, in order:

1. exact existing deterministic fixture coefficients where already available;
2. existing diagnostic maximum-centered-lift measurements from accepted Q012/P93 work;
3. mathematically propagated conservative bounds;
4. new read-only instrumentation in tests/diagnostics if needed.

Do not alter production semantics to obtain bounds.

For each reported bound state, identify whether the number is:
- exact observed maximum;
- proven conservative bound;
- inherited accepted diagnostic evidence.

---

# 4. Capacity table

Produce one compact table/artifact with columns:

- stage;
- substage;
- logical Level (ell);
- maintained prefix under Q-prefix v2;
- exact q-prefix product;
- component bounds ([B_0,B_1,ldots]);
- maximum (2B/S_Q(ell)) ratio;
- PASS/FAIL;
- source of bound;
- next operation;
- historical maintained representation at branch point.

For every PASS require strict:

[
2B<S_Q(ell).
]

Do not round an equality into PASS.

---

# 5. Important Level transitions

Audit these transitions explicitly.

## Level >= 4

Maintained prefix remains q0123 while logical Level decreases above 3.

## 4 -> 3

Maintained prefix still q0123.

## 3 -> 2

Maintained prefix must shrink q0123 -> q012.

Determine whether the current state satisfies same-lift contraction:

[
2B<S_{q012}.
]

If not, determine whether the logical operation explicitly permits canonical contraction modulo (Q_2).

Do not assume.

## 2 -> 1

Same for q012 -> q01.

## 1 -> 0

Same for q01 -> q0.

Report the actual production behavior required at each crossing.

---

# 6. Rescale audit

For every production Rescale:

- current logical Level;
- divisor (q_ell);
- maintained source prefix;
- source bound;
- predicted post-bound:
  [
  B'=leftlfloorrac{B+(q_ell-1)/2}{q_ell}ightfloor;
  ]
- new maintained prefix;
- strict new capacity result.

Classify whether the existing branch-point implementation already supports the needed source width:
- q01;
- q012;
- q0123;
- unsupported.

Do not implement q0123 Rescale in this audit.

---

# 7. ModUp audit

For Level-0 Bootstrap entry verify:

[
C=operatorname{Center}_{q_0}(r).
]

When raised to MaxLevel, the Q-prefix v2 hot state would materialize at most q0123 from (C).

Record:
- maximum observed/possible (|C|);
- q0 capacity;
- q0123 capacity;
- whether historical ModUp already computes q0/q1 or q012 rows needed downstream.

No code change.

---

# 8. C2S and S2C audit

Use actual matrices and current compressed/restore schedules.

For each factor report:
- matrix LevelQ;
- number of diagonals;
- matrix Scale;
- historical maintained rows;
- conservative LinearTransform bound using:
  [
  B' le Bsum_d|P_d|_1
  ]
  where available;
- post-Rescale bound;
- post-restore bound;
- Q-prefix capacity.

It is acceptable to reuse later private-F branch *numeric evidence* only as an external cross-check, but the authoritative audit must be reproducible from this qprefix branch / Primary diagnostics.

---

# 9. EvalMod/PS audit

This is the most important section because the branch-point candidate historically used Q012 selectively.

Identify every stage that historically required:
- Q01;
- Q012;
- any evidence suggesting q3.

For generated powers and PS:
- record exact/proven max centered lifts;
- record current logical Level at those moments;
- compare to q01, q012, and q0123 capacities;
- explain each widening/contraction boundary.

For DoubleAngle:
- record the bounded local-q2 windows already accepted historically;
- determine whether any round would require q3 under the v2 policy.

Do not modify polynomial scheduling.

---

# 10. Decision classification

Choose exactly one:

## A — `QPREFIX_V2_FULL_PROFILE_CAPACITY_PROVEN`

Every required authoritative state fits:

[
Q_{min(ell,3)}
]

at its actual logical Level, with all Level-crossing contraction/canonicalization semantics identified.

This authorizes a later implementation plan on `fast-qprefix`.

## B — `QPREFIX_V2_FAILS_AT_STAGE`

At least one required state cannot be represented under the capped prefix at its actual logical Level.

Report the **first causal failure**:
- stage;
- logical Level;
- required bound;
- available prefix product;
- deficit ratio;
- whether q4/full-Q would solve it;
- whether the failure is caused by an avoidable scheduling choice or fundamental current semantics.

Do not propose F automatically.

## C — `QPREFIX_V2_AUDIT_EVIDENCE_INCOMPLETE`

A required bound/semantic boundary cannot be established from available deterministic evidence.

Report exactly what evidence is missing.

---

# 11. Artifact

Create in Primary:

`results/QPREFIX-AUDIT-001-LOGN13-P93-FULL-BOOTSTRAP-CAPACITY.json`

Keep it compact.

Include:
- provenance;
- exact q0..q3;
- profile;
- capacity table;
- Level-transition table;
- Rescale table;
- C2S table;
- EvalMod/PS table;
- S2C table;
- first failure if any;
- classification.

Also create a concise markdown summary:

`results/QPREFIX-AUDIT-001-summary.md`

focused on the architectural decision.

---

# 12. Validation

Because this is an audit:
- no Secondary production behavior changes;
- read-only or test/diagnostic instrumentation only if required;
- run directly affected tests;
- run the accepted public Bootstrap control to ensure provenance;
- `git diff --check`.

If temporary instrumentation is added, keep it isolated to tests/diagnostics and commit only if it is useful/reproducible.

---

# 13. Prohibitions

Do not:
- modify `fast-ckks`;
- introduce F primes;
- implement q0123 arithmetic yet;
- rewrite Bootstrap;
- change matrices/scales/scheduling;
- tune parameters;
- weaken correctness thresholds;
- infer dormant rows as valid;
- start performance optimization;
- create production code merely to make the audit pass.

The only question is whether the capped, naturally shrinking logical-Q prefix is sufficient for the current profile.
