# QPREFIX-IMPL-004 — Q0123 Rescale and Explicit Level Transitions

## Status

Executable Codex implementation task.

## Task class

`I — Implementation`

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- orchestration/spec only.

Secondary:
- `xuejin-lu/lattigo@fast-qprefix`
- implementation target.

Accepted prerequisites:
- QPREFIX-IMPL-001: `6553491f9fb9b964c8fd0d743f3a302de83d0b54`
- QPREFIX-IMPL-002: `91baa6a4655e10fe2460a399633fa54a03a35318`
- QPREFIX-IMPL-003: `c8b591a30c05a2261de8d0181d7b3f64bc58169b`

Authoritative architecture:
- `docs/FAST_QPREFIX_SPEC.md`
- Primary `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

## Purpose

Generalize Fast Rescale and explicit Level transitions to the Q-prefix v2 policy:

[
w_Q(ell)=min(ell+1,4).
]

For logical Level (ell), the authoritative reconstruction basis is:

[
S_Q(ell)=prod_{i=0}^{min(ell,3)}q_i.
]

Rescale always divides by the logical top modulus:

[
d=q_ell,
]

even when (ell>3) and (q_ell) is not physically maintained.

This task is the first production subsystem allowed to consume a fully authoritative q0123 prefix.

Do not migrate ModUp, DFT, EvalMod, or Bootstrap orchestration yet.

---

# 1. Fixed-width q0123 reconstruction

Extend the existing q01/q012 fixed-width reconstruction foundation to support q0123.

Current target profile has approximately:

[
56+39+40+39 < 192
]

bits for q0123.

Use a fixed-width representation sufficient for the actual supported q0..q3 product. Reuse/generalize the existing `uint192` machinery where safe.

Required:
- exact q0123 modulus/product words;
- exact floor-half words;
- Garner/CRT reconstruction from q0,q1,q2,q3;
- centered representative with the accepted odd-modulus convention:
  [
  xlelfloor S/2floor Rightarrow +x,
  qquad
  x>lfloor S/2floor Rightarrow x-S;
  ]
- exact reduction of a signed fixed-width magnitude to q0..q3;
- explicit overflow/range validation for the actual q0..q3 values.

No per-coefficient `math/big` in the Rescale hot loop.

Using `math/big` in tests/oracles is encouraged.

---

# 2. Q-prefix Rescale semantics

For each coefficient component authoritative lift (X):

[
Y=operatorname{Round}(X/q_ell).
]

Preserve the existing centered round-to-nearest convention.

For one or more rescale steps:

[
X_0=X,
]

[
X_{k+1}=operatorname{Round}(X_k/q_{ell-k}).
]

Do not replace sequential rounding by division by a product unless proven exactly equivalent to the existing Standard semantics.

Output:
- Level:
  [
  ell'=ell-r
  ]
  for (r) rescale steps;
- Scale:
  [
  Delta'=Delta/prod_{k=0}^{r-1}q_{ell-k};
  ]
- physical maintained width:
  [
  w_Q(ell').
  ]

The output rows must contain:

[
Ymod q_i,quad i=0,ldots,w_Q(ell')-1.
]

Rows above the target prefix remain dormant/nil according to compact storage policy.

---

# 3. Source width rules

For production `Rescale` after this milestone:

- Level 0: invalid;
- Level 1: q01 source;
- Level 2: q012 source;
- Level >= 3: q0123 source.

The operation must validate all source rows required by:

[
QPrefixWidth(ell).
]

This is a deliberate production activation boundary for Rescale only.

Do not use legacy `maintainedLimbCount` to select Rescale source width after this task.

Do not read rows above the Q-prefix policy.

---

# 4. Logical divisor rule

The divisor is always:

[
q_ell
]

from the full configured CKKS parameter chain.

Examples:

- Level 16 with maintained q0123:
  reconstruct (X) from q0123, divide by logical q16, output still q0123 at Level 15;
- 4 -> 3:
  reconstruct q0123, divide by q4, output q0123;
- 3 -> 2:
  reconstruct q0123, divide by q3, output q012;
- 2 -> 1:
  reconstruct q012, divide by q2, output q01;
- 1 -> 0:
  reconstruct q01, divide by q1, output q0.

Never substitute q3 merely because q3 is the highest maintained row.

---

# 5. Result-capacity gate

For a source proven/observed bound (B), use:

[
widehat B'=
leftlfloor
rac{B+(q_ell-1)/2}{q_ell}
ightfloor.
]

For multiple steps, propagate this recurrence sequentially.

The implementation itself may obtain the exact centered source magnitude during reconstruction and must reject if the resulting output magnitude cannot fit the target Q-prefix centered interval.

At every target Level require:

[
2|Y|<S_Q(ell').
]

This runtime exact-result check is mandatory at prefix-shrinking transitions.

Use `QPrefixCapacityError` or a clearly wrapped recognizable equivalent.

Failure must be transactional:
- input unchanged;
- output coefficients/metadata unchanged.

---

# 6. Explicit Level-transition APIs

Do not let structural `Resize` decide representative semantics.

Provide explicit helpers/APIs for the two distinct Level-drop meanings.

## A. Same-lift contraction

Conceptually:

`DropLevelSameLift(op0, targetLevel, opOut)`

Semantics:

[
Y=X
]

as an integer lift.

Allowed only if the current authoritative (X) fits the target prefix:

[
2|X|<S_Q(targetLevel).
]

Then target residues are simply (Xmod q_i) for the new prefix.

For prefix boundary crossings:
- 3 -> 2: q0123 -> q012;
- 2 -> 1: q012 -> q01;
- 1 -> 0: q01 -> q0.

If capacity fails, reject transactionally.

## B. Canonical logical contraction

Conceptually:

`DropLevelCanonical(op0, targetLevel, opOut)`

Semantics:

[
C=operatorname{Center}_{Q_{targetLevel}}(Xmod Q_{targetLevel}).
]

For targetLevel <= 3:

[
Q_{targetLevel}=S_Q(targetLevel).
]

Materialize the target prefix from (C).

This is an explicit representative-change boundary.

Do not silently use canonical contraction where same-lift semantics are required.

If current production has no caller for one of these APIs yet, keep it standalone and tested; later milestones select the correct owner.

---

# 7. Domain behavior

Production Rescale input remains:
- NTT;
- ordinary or Montgomery.

Required conversion:
- INTT only maintained source rows;
- IMForm source rows if Montgomery;
- reconstruct/divide in authoritative integer domain;
- reduce into target prefix;
- NTT target rows;
- restore Montgomery form if input was Montgomery.

Preserve:
- degree;
- batching/dimensions metadata;
- IsNTT;
- IsMontgomery.

No dormant/full-Q materialization.

---

# 8. Scratch

Generalize `fastRescaleScratch` for q0123:
- coefficient scratch through four rows;
- result scratch through four rows;
- q0..q3 moduli;
- CRT inverses/products/half values needed by the fixed-width implementation.

Scratch remains evaluator-owned and reusable.

Do not allocate `math/big` per coefficient.

---

# 9. Standard oracle tests

Use independent `math/big` centered CRT + signed rounding as the primary oracle.

Test exact coefficients for:

## Source widths
- Level 1 / q01;
- Level 2 / q012;
- Level 3 / q0123;
- high Level (e.g. Level >= 4) / q0123.

## Sign/boundary values
- 0;
- +1 / -1;
- positive/negative values;
- floor(S/2)-1;
- floor(S/2);
- floor(S/2)+1 modulo S;
- values near target-prefix contraction capacity.

## Divisors
- q1;
- q2;
- q3;
- a logical q_l with l > 3.

Require exact equality of every maintained output row.

---

# 10. Multi-step Rescale tests

For `RescaleTo` / multiple rescale steps:
- compare against sequential independent oracle;
- cover high-Level q0123 staying q0123;
- cover a sequence crossing 4->3->2;
- cover 3->2->1;
- cover 2->1->0 where valid.

Verify Level and Scale exactly.

---

# 11. Explicit DropLevel tests

For SameLift:
- fitting positive/negative lift;
- exact strict capacity boundary;
- first non-fitting value;
- 3->2, 2->1, 1->0;
- coefficient and NTT/Montgomery forms if APIs support them;
- transactionality on failure.

For Canonical:
- source lift larger than target half-range;
- prove output equals:
  [
  Center_{Q_t}(Xmod Q_t);
  ]
- midpoint convention at target odd modulus/product;
- distinct test showing canonical output may differ from same-lift output.

Structural `Resize` must remain semantics-free.

---

# 12. Existing profile regression

Run current LogN13/P93 Q-prefix workload tests.

Specifically reproduce accepted QPREFIX-AUDIT-002 C2S Rescale checkpoints and require:
- same Level/Scale;
- maintained q rows exact;
- no public correctness regression.

This task does not need to re-audit C2S capacity from scratch; reuse the accepted fixture as regression evidence.

---

# 13. Transitional isolation

Only Rescale and explicit DropLevel helpers become Q-prefix-policy authoritative in this milestone.

Other production wrappers may still use legacy row policy until their own milestones.

Do not globally replace `maintainedLimbCount`.

Poison q3 in a legacy non-Rescale path and confirm this task does not accidentally activate it there.

---

# 14. Validation

Required:
- focused q0123 CRT/centered-rounding tests;
- Rescale/RescaleTo oracle tests;
- DropLevel SameLift/Canonical tests;
- accepted C2S Rescale regression;
- `go test ./schemes/ckks/fast`;
- relevant DFT/Bootstrap tests;
- `go test ./...`;
- `git diff --check`;
- gofmt check.

Any regression is blocking.

---

# 15. Prohibitions

Do not:
- modify ModUp/Trace production behavior;
- modify DFT/LinearTransform routing;
- modify EvalMod/PS/DoubleAngle;
- modify Bootstrap orchestration;
- introduce F;
- use dormant q rows;
- use Standard full-RNS fallback;
- change q values/parameters;
- globally activate q0123 outside Rescale/explicit Level-transition APIs;
- modify `fast-ckks`.

---

# 16. Completion report

Report:
- Secondary commit;
- changed files;
- q0123 fixed-width reconstruction design;
- Rescale source-width policy;
- logical-divisor evidence for Level >3;
- 3->2 / 2->1 / 1->0 results;
- SameLift vs Canonical DropLevel API/semantics;
- exact Standard-oracle results;
- accepted C2S regression result;
- full regressions;
- confirmation no other production subsystem was migrated.

Successful handoff:
`READY_FOR_WEB_REVIEW`.

Use `NEEDS_WEB_REVIEW` for any discovered semantic ambiguity in Level-transition ownership or fixed-width overflow/range conflict.
