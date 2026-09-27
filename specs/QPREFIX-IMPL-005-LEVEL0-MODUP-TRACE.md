# QPREFIX-IMPL-005 — Level-0 ModUp, Scale Alignment, and Trace

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
- QPREFIX-IMPL-004: `18f4da53f03acd065d18550a8f2106725462896f`

Authoritative architecture:
- Secondary `docs/FAST_QPREFIX_SPEC.md`
- Primary `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

## Purpose

Migrate the Bootstrap entry boundary

`ScaleDown -> Level-0 ModUp -> scale alignment -> Trace`

to the Q-prefix v2 architecture without activating q0123 inside DFT/C2S yet.

The key semantic boundary is Level-0 ModUp:

[
C = Center_{q_0}(X mod q_0).
]

For target logical Level (L), ModUp must materialize exactly

[
q_0,ldots,q_{min(L,3)}
]

from the same canonical integer (C).

For the production LogN13 path with (Lge3), ModUp therefore creates authoritative q0123.

After q3 is materialized, every operation inside the same ModUp boundary that mutates the ciphertext before returning — scale alignment and Trace — must process the same explicit authoritative width. Otherwise q3 becomes stale immediately.

This milestone ends at the output of `FastEvaluator.ModUp`.

C2S/DFT activation remains QPREFIX-IMPL-006.

---

# 1. Frozen Level-0 canonicalization rule

Let q0 be odd:

[
q_0 = 2m+1.
]

For canonical residue (rin[0,q_0)):

[
Center_{q_0}(r)=
egin{cases}
r,&0le rle m,\
r-q_0,&m+1le r<q_0.
end{cases}
]

Therefore:

[
r=q_0>>1=m
]

is **positive**.

The current production comparison using `coeff >= q0>>1` is inconsistent with this frozen rule and must be corrected.

Do not approximate this rule through signed casts or floating point.

Provide one small helper/owner for this canonical mapping so q1/q2/q3 materialization cannot implement different midpoint behavior.

---

# 2. ModUp materialization policy

For Level-0 input and target Level (L):

[
w = QPrefixWidth(L)=min(L+1,4).
]

Materialize exactly rows:

[
q_0,ldots,q_{w-1}.
]

Requirements:
- q0 input is the only source authority;
- INTT q0 if required by the current boundary;
- obtain canonical centered (C) using the rule above;
- set each target row to (C mod q_i);
- return target rows to the required NTT domain;
- do not materialize q_w and above;
- preserve logical Level (L), Scale, degree, batching/dimension metadata;
- do not introduce any F basis.

For (Lge3), q3 must be a real N-sized row and must be independently correct.

Do not use `MaintainedLimbCount` to decide the ModUp output width.

---

# 3. Compact physical storage

Replace the current hard-coded restore loop that allocates only rows below index 3.

The result must follow QPREFIX-IMPL-002:

- Level 0 -> q0;
- Level 1 -> q01;
- Level 2 -> q012;
- Level >=3 -> q0123;
- all higher logical rows dormant/nil.

Prefer the shared Q-prefix lifecycle helpers rather than a second storage-width policy.

---

# 4. Explicit-width integer scale alignment

Current ModUp may multiply by an integer alignment factor after basis raise.

Once ModUp has produced q0123, that multiplication must process exactly the same authoritative width.

Add/reuse an explicit-row integer-multiply entry point backed by the QPREFIX-IMPL-003 kernel, conceptually:

`MulIntegerQPrefixRows(op0, scalar, rows, opOut)`

Exact name/API is a coding choice.

Requirements:
- validate (1 le rows le QPrefixWidth(level));
- preserve Scale metadata unless caller updates it explicitly, matching current `MulIntegerMaintained` semantics;
- support in-place operation;
- touch no rows above `rows`;
- existing legacy `MulIntegerMaintained` behavior outside this migrated boundary remains unchanged.

Inside ModUp, use the explicit ModUp width, not legacy `maintainedLimbCount`.

---

# 5. Explicit-width Trace

Generalize Trace so its core can operate on an explicit Q-prefix width.

Conceptually separate:
- legacy production wrapper `Trace(...)`, if other callers still require legacy authority;
- explicit-width Trace core used by migrated ModUp.

For the ModUp path at Level >=3, Trace must process q0123 for:
- initial copy;
- inverse normalization scalar;
- every automorphism;
- every row-wise accumulation.

Use the QPREFIX-IMPL-003 explicit-row automorphism kernel.

Do not infer authority from backing allocation.

Do not globally activate width 4 for unrelated callers.

Preserve:
- Level;
- Scale;
- NTT/Montgomery domain;
- degree;
- batching/dimension metadata.

---

# 6. Montgomery boundary

Current `FastEvaluator.ModUp` performs:
1. basis raise;
2. Trace;
3. maintained-row Montgomery conversion.

After this milestone, when target Level >=3, the final Montgomery conversion must include q0123.

All rows that ModUp declares authoritative must end in the same representation.

Do not leave q3 ordinary while q012 are Montgomery.

Rows above q3 remain dormant.

---

# 7. ScaleDown ownership

ScaleDown remains the **pre-ModUp legacy-authority path** in this milestone.

Do not force q0123 activation before Level-0 canonical ModUp.

Required:
- preserve existing accepted ScaleDown numerical behavior;
- `RescaleTo` continues to use its current explicitly selected legacy authority;
- final ScaleDown result is Level 0 / q0 authoritative;
- do not let allocated q3 backing become authoritative during ScaleDown merely because QPrefixWidth at a high logical Level is four.

The current cheap logical DropLevel loop must be documented/tested as a pre-ModUp logical residue drop, not as q0123 activation.

If a concrete prefix-crossing case exposes an actual representative ambiguity in accepted ScaleDown tests, stop with `NEEDS_WEB_REVIEW`; do not silently canonicalize or read dormant rows.

---

# 8. ModUp-to-C2S transition

At the output of `FastEvaluator.ModUp`:
- q0123 is authoritative for Level >=3;
- all four rows have matching NTT/Montgomery representation.

However QPREFIX-IMPL-006 has not yet migrated LinearTransform/DFT.

Therefore:
- this milestone must not change LinearTransform/DFT;
- current C2S is still allowed to consume only its legacy q012 authority and thereby stop preserving q3;
- current production Rescale remains legacy-width activated there;
- no downstream code may assume q3 remains authoritative after the first legacy C2S producer.

Add a focused test that makes this transition explicit so future work knows exactly where q3 authority is currently lost.

---

# 9. Independent ModUp oracle

For each tested q0 coefficient residue (r):

1. compute
   [
   C=Center_{q_0}(r)
   ]
   independently using `math/big` or simple signed test arithmetic;
2. for every active target row (i):
   [
   want_i=Cmod q_i;
   ]
3. compare after the same NTT/Montgomery transformations used by production.

Cover target Levels:
- 1;
- 2;
- 3;
- a high Bootstrap Level.

For Level >=3 require an explicit q3 equality check.

---

# 10. Mandatory midpoint tests

At minimum test q0 residues:

- 0;
- 1;
- (m-1);
- (m=q_0>>1);
- (m+1);
- (q_0-2);
- (q_0-1).

The test must prove:
- (m) maps to (+m);
- (m+1) maps to (-m).

This must fail against the old `>= q0>>1` implementation.

---

# 11. Trace row oracle

Construct a synthetic fully populated q0123 NTT ciphertext.

Run explicit-width Trace with rows=4.

Independently apply the Trace normalization/automorphism schedule row-by-row using the corresponding `ring.SubRing` operations.

Require exact equality of q0,q1,q2,q3.

Also:
- poison q4+ and prove no effect;
- test legacy Trace wrapper separately, if retained, to prove it does not accidentally promote q3 authority.

---

# 12. End-to-end ModUp boundary oracle

For a real accepted LogN13/P93 Level-0 input:

`Level0 q0 -> modUpBasis -> scale alignment -> Trace -> Montgomery`

verify at the returned `ModUp` boundary:

- Level equals target Bootstrap Level;
- physical active width is `QPrefixWidth(Level)`;
- q0123 all have N-sized backing;
- q4+ are nil/dormant;
- q0123 correspond to one consistent bounded integer state through the row-wise ModUp/Trace semantics;
- all active rows are NTT + Montgomery consistently;
- Scale equals the accepted pre-005 value;
- degree and metadata are unchanged.

Use q3-specific hashes/oracles so a q012-only implementation cannot pass.

---

# 13. Existing regressions

Required regressions:
- existing Fast ScaleDown tests;
- existing Fast ModUp tests;
- existing Fast Trace tests;
- accepted QPREFIX-AUDIT-002 C2S checkpoints;
- DFT tests;
- the existing public q0=56 Bootstrap generated-secret and zero-secret correctness controls where available.

The public C2S numerical behavior should remain unchanged until QPREFIX-IMPL-006 activates q3 there.

---

# 14. Transactionality / mutation discipline

Perform all structural/domain preflight checks before destructive INTT or resize work where practical.

On validation failure before arithmetic:
- do not leave a partially widened ciphertext;
- do not leave q0 in coefficient domain while metadata still claims NTT;
- do not leave mixed Montgomery state.

If an existing in-place API makes full rollback impractical for an impossible-after-preflight internal error, document that invariant and ensure all fallible validations occur before mutation.

---

# 15. Performance guard

This milestone is correctness-first, but record focused timings/allocation counts for:
- old/legacy q012 ModUp basis path if still reproducible in a test helper;
- new q0123 ModUp basis;
- Trace rows=3 versus rows=4.

Report:
- ns/op;
- B/op;
- allocs/op.

Do not start an optimization campaign.

A pathological >10x regression requires `NEEDS_WEB_REVIEW`.

---

# 16. Prohibitions

Do not:
- modify DFT/LinearTransform production routing;
- activate q0123 inside C2S/S2C;
- modify EvalMod/PS/DoubleAngle;
- modify Bootstrap packing/N1-N2 boundaries;
- change CKKS parameters or schedule;
- introduce F;
- use Standard full-RNS fallback;
- read q4+ dormant rows;
- globally replace every legacy row selector;
- modify `fast-ckks`.

---

# 17. Completion report

Report:
- Secondary commit;
- changed files;
- midpoint fix and exact boundary evidence;
- ModUp target-width policy;
- explicit-width integer-scale and Trace APIs;
- q3 ModUp/Trace oracle evidence;
- final active-row domain state;
- where q3 authority is intentionally lost when entering still-legacy C2S;
- ScaleDown regression status;
- public Bootstrap/C2S regression status;
- benchmark numbers;
- full regression results.

Successful handoff:
`READY_FOR_WEB_REVIEW`.

Use `NEEDS_WEB_REVIEW` if:
- accepted ScaleDown semantics become ambiguous at a prefix crossing;
- a ModUp/Trace production path requires dormant q4+ data;
- q3 cannot be kept coherent through the complete ModUp boundary;
- >10x unexplained focused regression remains after one bounded repair.
