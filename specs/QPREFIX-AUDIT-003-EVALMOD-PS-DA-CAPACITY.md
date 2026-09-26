# QPREFIX-AUDIT-003 — Current-Branch EvalMod / PS / DoubleAngle Capacity Audit

## Status

Executable Codex audit task.

## Task class

`M/E — Focused evidence completion`

No production behavior change is authorized.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- owns compact evidence/report.

Secondary:
- `xuejin-lu/lattigo@fast-qprefix`
- current-branch audit target.

Accepted prerequisites:
- QPREFIX-AUDIT-001 at Primary `494e7b7b672cbedb4eb842887d464df2ea0cb7f9`
- QPREFIX-AUDIT-002 accepted at Primary `78ce1599ac9a118fdd8b246d8bb52d4d4dbaeb7f`
- Q-prefix v2 architecture `docs/FAST_QPREFIX_SPEC.md`.

## Purpose

Close the remaining high-risk capacity evidence for the current LogN13/P93 profile in:

1. EvalMod normalization;
2. Chebyshev generated powers;
3. Paterson–Stockmeyer baby/giant stages;
4. polynomial output;
5. DoubleAngle rounds.

This task must answer:

[
oxed{
	ext{Does every authoritative EvalMod state fit }
Q_{min(ell,3)}
	ext{ at its actual logical Level?}
}
]

Do not audit S2C in this task.

---

# 1. Exact current-branch execution

Use the same deterministic q0=56 LogN13/P93 profile and workload as the accepted public control.

Start from the exact current `fast-qprefix` C2S output produced by the accepted current-branch path.

Run the actual current Fast EvalMod path.

Instrumentation may be added only in tests/diagnostics to expose internal checkpoints.

Do not alter:
- polynomial planner;
- generated-power schedule;
- PS grouping;
- DoubleAngle count;
- Scale choices;
- q0/q1/q2 selection logic;
- production semantics.

---

# 2. Required checkpoints

At minimum capture for real and imag branches where both exist:

## EvalMod entry / normalization
- C2S output entering EvalMod;
- after any scalar normalization/offset preparation;
- before polynomial evaluation.

## Generated powers
Capture every power actually used by the current P93 plan, including at least:
- T2;
- T3;
- T4;
- T6;
- T8;
- T16;
and any additional stored powers used by the planner.

For each generated-power construction, capture:
- operand input states;
- pre-Rescale product/correction state;
- post-Rescale state;
- logical Level and Scale.

## Paterson–Stockmeyer
Capture every critical state used in the accepted historical diagnostics, including:
- B0..B4 constants/terms as applicable;
- G0/G1/... add/multiply/rescale boundaries;
- F0/root/final-rescale boundaries;
- final polynomial output.

Do not require naming identical to old artifacts if current source structure differs, but map each current checkpoint to its PS role.

## DoubleAngle
For every round:
- input;
- square;
- after multiplier;
- after constant;
- post-Rescale.

Also capture final internal restore/output entering S2C.

---

# 3. Exact observed bounds

At every checkpoint compute, per component:

[
B_j=max_k|X_{j,k}|.
]

Use the strongest physically authoritative maintained prefix present in the current branch:

- Q01 when only q0/q1 are authoritative;
- Q012 when q2 is maintained/authoritative.

Never read dormant q3+ rows.

Record:
- exact prefix used;
- exact product;
- centered reconstruction convention;
- whether strict uniqueness holds.

---

# 4. Q-prefix-v2 policy capacity

For logical Level (ell), policy capacity is:

[
S_Q(ell)=prod_{i=0}^{min(ell,3)}q_i.
]

For each checkpoint report:

[
ho_j=rac{2B_j}{S_Q(ell)}.
]

Require strict:

[
ho_j<1.
]

If a physically reconstructed Q012 checkpoint is strict Q012-safe at Level >= 3, that is stronger evidence than required by the Q0123 policy.

If a checkpoint is only Q01-safe but historical/current semantics require Q012, classify that distinction explicitly.

---

# 5. Arithmetic recurrence proofs

In addition to observed bounds, record conservative recurrence bounds for every critical transition.

## Add/Sub
[
B'le B_X+B_Y.
]

## Integer scalar
[
B'le |m|B.
]

## Negacyclic ciphertext multiplication
[
B'_kle Nsum_i B_{X,i}B_{Y,k-i}.
]

## Plaintext/scalar polynomial operations
Use the existing operation-appropriate exact or conservative bound.

Observed values must satisfy the carried proven bound.

If the carried proof becomes too loose to prove capacity while exact observed values fit, classify as evidence incomplete rather than claiming safety from observation alone.

---

# 6. Rescale recurrence

For every current EvalMod/PS/DoubleAngle Rescale record:

- source checkpoint;
- logical source Level (ell);
- divisor:
  [
  d=q_ell;
  ]
- exact source observed bound;
- carried proven source bound;
- predicted rounded result:
  [
  widehat B'=
  leftlfloor
  rac{B+(d-1)/2}{d}
  ightfloor;
  ]
- observed post-Rescale bound;
- new Level;
- new Q-prefix policy capacity.

Require:
[
B'_{	ext{obs}}le widehat B'_{	ext{obs-source}}
]
and
[
B'_{	ext{obs}}le widehat B'_{	ext{proven-source}}.
]

Record whether the current implementation uses Q01 or Q012 reconstruction for that Rescale.

---

# 7. Level crossing audit

This task is especially responsible for transitions near lower levels.

Explicitly identify any current EvalMod transition crossing:

## 3 -> 2
Policy:
[
Q0123	o Q012.
]

Prove the output state satisfies:
[
2B<Q012.
]

## 2 -> 1
Policy:
[
Q012	o Q01.
]

Prove:
[
2B<Q01
]
at the actual crossing, or identify the explicit canonicalization semantics if same-lift contraction is not intended.

Do not assume row truncation is valid.

If these crossings occur only later in S2C, state that and leave them for the next audit.

---

# 8. Historical evidence comparison

Old `fast-ckks` artifacts may be used only as a cross-check.

For each matched checkpoint, report:
- old observed bound/ratio;
- current `fast-qprefix` observed bound/ratio;
- whether they materially agree.

The current-branch result is authoritative.

Do not fail merely because floating/numerical values differ slightly if current public semantics remain accepted and capacity proof is valid.

---

# 9. Public semantic guard

Run the accepted q0=56 public Bootstrap control after instrumentation.

Require:
- generated-secret max component error < 1e-2;
- zero-secret max component error < 1e-2;
- final Level/Scale/NTT/Montgomery metadata unchanged.

Instrumentation must not affect public behavior.

---

# 10. Decision classification

Choose exactly one:

## A — `QPREFIX_EVALMOD_CAPACITY_PROVEN`

All required current-branch EvalMod/PS/DoubleAngle authoritative states:
- have exact observed bounds;
- have valid carried recurrence proofs;
- satisfy strict Q-prefix-v2 capacity;
- have valid Rescale recurrence;
- preserve public control.

This closes EvalMod/PS/DA as a reason for independent F storage.

## B — `QPREFIX_EVALMOD_CAPACITY_FAIL`

A required state violates:
[
2Bge S_Q(ell).
]

Report first causal failure:
- stage/checkpoint;
- Level;
- observed/proven bound;
- available policy prefix;
- deficit ratio;
- whether q3 or q4 would be sufficient.

Do not propose F automatically.

## C — `QPREFIX_EVALMOD_PROOF_TOO_LOOSE`

Observed current-branch state fits, but the required conservative recurrence proof cannot establish strict capacity.

Report the first proof gap.

## D — `QPREFIX_EVALMOD_MEASUREMENT_CONFLICT`

Observed state violates the predicted arithmetic/Rescale recurrence or instrumentation changes semantics.

## E — `QPREFIX_EVALMOD_EVIDENCE_INCOMPLETE`

A required current-branch internal state cannot be observed/reconstructed without production changes.

---

# 11. Artifacts

Create in Primary:

`results/QPREFIX-AUDIT-003-EVALMOD-PS-DA-CAPACITY.json`

and:

`results/QPREFIX-AUDIT-003-summary.md`

Include:
- provenance;
- q0..q3;
- checkpoint table;
- generated-power table;
- PS table;
- DoubleAngle table;
- every Rescale recurrence;
- level-crossing table;
- old-vs-current cross-check;
- public-control result;
- classification.

Keep compact; no full vectors.

---

# 12. Secondary instrumentation

Preferred:
- focused test/diagnostic-only hooks on `fast-qprefix`;
- no production behavior change.

If internal state is inaccessible without a small diagnostic helper, keep it test-only or clearly diagnostic and non-production.

Useful reproducible instrumentation may be committed/pushed after validation.

---

# 13. Validation

Run:
- focused EvalMod/PS/DA audit test;
- relevant Fast polynomial/Mod1/Bootstrap tests;
- accepted q0=56 public control;
- `git diff --check`;
- JSON parse.

The known unrelated Primary `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT` remains documented debt.

---

# 14. Prohibitions

Do not:
- modify production EvalMod/polynomial/DA behavior;
- modify `fast-ckks`;
- introduce F;
- implement q3 arithmetic;
- tune planScale/q values;
- change polynomial schedule;
- weaken public threshold;
- audit S2C in this task;
- repair unrelated historical test debt.

This task closes only the current-branch EvalMod/PS/DoubleAngle evidence chain.
