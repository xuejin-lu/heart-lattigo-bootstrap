# FAST-STORAGE-005 — Level-0 ModUp Canonicalization over Private-F Storage

## Status

Executable Codex task.

## Task class

`I — Implementation`

The mathematical and architectural decisions for this task are already frozen by GPT Web in Secondary `docs/FAST_CKKS_SPEC.md`, especially Sections 4.4.1, 4.9, 4.11–4.15.

Codex owns implementation and coding-quality review only.

If current source evidence appears to require changing:
- the Level-0 canonicalization rule;
- the logical target-level semantics;
- the fixed-width-3 production policy;
- the Q/F separation;
- or the meaning of Scale,

stop with `NEEDS_WEB_REVIEW`.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`
- orchestration/spec only.

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- implementation target.

Accepted prerequisite:
- FAST-STORAGE-004 implementation `1b9ecd7973505cac1c4a1673a9578260a761950f`
- fixed-width-3 production policy in Secondary docs commit `d9919f9c080e0dfa731746f5c447f93633ae2f36`

## Purpose

Implement the standalone private-storage ModUp canonicalization boundary for the production-oriented Level-0 entry.

This task does **not** implement Bootstrap Trace, scale alignment, Montgomery conversion, or production wiring.

It implements only the mathematical boundary:

[
X
longmapsto
C=operatorname{Center}_{q_0}(Xmod q_0)
]

followed by an increase in logical Level metadata to an explicit target level.

The private storage remains in the same fixed width-3 F basis.

---

# 1. Frozen transition

Input:
- valid `FastCiphertext`;
- `LogicalLevel = 0`;
- `StorageWidth = 3`;
- positive Scale;
- non-Montgomery;
- coefficient or NTT private-F domain.

Let the authoritative lift be `X_j` for each component/coefficient.

Define:

[
C_j=operatorname{Center}_{q_0}(X_jmod q_0).
]

The output must represent `C_j` exactly in the same private basis:

[
(C_jmod f_0, C_jmod f_1, C_jmod f_2).
]

Metadata transition:

[
	ext{LogicalLevel}: 0ightarrow L_{target}
]

where

[
1le L_{target}le 	ext{params.MaxLevel()}.
]

Scale is unchanged.

Degree is unchanged.

Storage width remains exactly 3.

Domain is preserved.

---

# 2. Why canonicalization is mandatory

A private-F lift at Level 0 may be

[
X=c+kq_0
]

for nonzero `k`.

Simply increasing `LogicalLevel` while keeping `X` would generally represent the wrong class modulo:

[
Q_{L_{target}}=q_0q_1cdots q_{L_{target}}.
]

Therefore ModUp must first choose the canonical logical representative:

[
oxed{
C=operatorname{Center}_{q_0}(Xmod q_0)
}
]

before logical-basis growth.

This is a semantic boundary, not a storage-width operation.

---

# 3. API shape

Provide a small standalone API conceptually equivalent to:

`FastStorageModUpLevel0(op *FastCiphertext, targetLevel int) (*FastCiphertext, error)`

Requirements:
- non-mutating input;
- explicit target logical level;
- new output;
- no in-place alias API required;
- no Bootstrap evaluator wiring.

Exact naming/file split may improve.

Prefer:
- `schemes/ckks/fast/storage_modup.go`
- `schemes/ckks/fast/storage_modup_test.go`

---

# 4. Input contract

Require:

- non-nil valid Fast ciphertext;
- `LogicalLevel == 0`;
- `StorageWidth == 3`;
- targetLevel in `[1, params.MaxLevel()]`;
- positive Scale;
- valid bounds;
- valid storage rows;
- Standard ring;
- non-Montgomery.

Support:
- coefficient private-F domain;
- NTT private-F domain.

Output domain must equal input domain.

---

# 5. Exact physical algorithm

For each component/coefficient:

1. if NTT, INTT all three private rows;
2. reconstruct the unique authoritative lift `X`;
3. compute:
   [
   r=Xmod q_0
   ]
   in canonical residue form `0 <= r < q_0`;
4. center:
   [
   C=
   egin{cases}
   r,&rle q_0/2\
   r-q_0,&r>q_0/2
   end{cases}
   ]
   using the repository's chosen centered convention consistently;
5. encode `C` into all three private storage primes;
6. if input was NTT, NTT all three output rows.

Do not materialize logical `q_1...q_L` residue rows inside the private container.

Do not call ordinary logical RingQ operations on private-F rows.

---

# 6. Bound transition

After canonicalization:

[
|C|le q_0/2.
]

For each component, the preferred bound is the exact observed maximum canonical magnitude:

[
B'_j=max_k |C_{j,k}|.
]

This task is an explicit canonicalization boundary, so one coefficient scan is expected and valid.

The output bound must satisfy width-3 uniqueness.

Do not retain the old pre-canonicalization bound if it is larger.

Do not contract width.

---

# 7. Fixed width policy

This task implements the current production policy:

[
oxed{	ext{StorageWidth}=3}
]

throughout ModUp.

Do not:
- contract 3->2 or 3->1;
- call the width planner to choose a smaller output width;
- introduce adaptive-width policy.

The existing planner remains a correctness utility only.

---

# 8. Metadata transition

Preserve:
- Scale exactly;
- degree exactly;
- NTT/coefficient domain;
- plaintext batching/dimensions/bit-reversal metadata;
- CKKS parameter identity/value.

Change only:
- `LogicalLevel = targetLevel`;
- component bounds to canonicalized observed bounds.

`IsMontgomery` remains false.

---

# 9. Independent logical oracle

For each test input:

1. export the Level-0 Fast input to LogicalQ Level 0;
2. derive the centered `q0` representative independently;
3. run `FastStorageModUpLevel0`;
4. export the output at `targetLevel`;
5. verify every logical row satisfies:
   [
   row_i[k]=C_kmod q_i,quad 0le ile targetLevel.
   ]

This oracle must cover noncanonical private lifts `X=c+kq_0`.

The test should fail if an implementation merely increments logical Level without canonicalizing `X`.

---

# 10. Required tests

## T1 — canonical centered q0 mapping

Use values including:
- 0;
- +/-1;
- +/-q0;
- +/-2q0;
- q0/2-near boundary values;
- `c+kq0` for positive and negative k.

Verify exact canonical centered output.

## T2 — coefficient-domain exactness

Use multi-component synthetic private-F inputs.

Verify exact decoded output equals independently computed centered-q0 values.

## T3 — NTT-domain exactness

Repeat T2 with NTT input.

Verify output remains NTT and decodes to the same canonical values.

## T4 — target logical levels

Test multiple target levels, including:
- 1;
- an intermediate level;
- MaxLevel.

Verify `LogicalLevel` changes exactly and storage width remains 3.

## T5 — Scale/metadata preservation

Verify:
- Scale unchanged;
- degree unchanged;
- width=3 unchanged;
- domain unchanged;
- plaintext metadata unchanged;
- parameter identity/value unchanged.

## T6 — bound reset to canonical observed maximum

Use inputs whose pre-ModUp bound is much larger because `X=c+kq0`.

Verify output bounds equal exact observed `max |Center_q0(X mod q0)|`.

## T7 — noncanonical lift theorem

Construct two different private lifts:

[
X_1=c+k_1q_0,qquad X_2=c+k_2q_0
]

with different nonzero `k_1,k_2`.

After ModUp to the same target level, verify both produce identical canonical private-F outputs and identical exported logical-Q ciphertexts.

## T8 — logical export oracle

For successful outputs, export to LogicalQ and verify every row equals canonical C reduced modulo logical q_i.

Cover at least two target levels and both coefficient/NTT domains.

## T9 — no hidden logical-row materialization

Verify the private container still contains only three private-F rows per component regardless of target logical level.

## T10 — failures transactional

Reject:
- nil input;
- input LogicalLevel != 0;
- width 1 or 2;
- targetLevel <= 0;
- targetLevel > MaxLevel;
- Montgomery;
- malformed rows;
- malformed bounds;
- non-positive Scale.

Input must remain unchanged after every failure.

## T11 — Rescale -> ModUp boundary chain

Construct a valid path that reaches Level 0 by standalone private Rescale from Level 1, then call ModUp to a higher target Level.

Verify:
- Rescale result may carry a noncanonical Level-0 lift;
- ModUp canonicalizes it;
- exported logical output matches independent centered-q0 reference;
- Scale is unchanged by ModUp.

## T12 — regressions

Run:
- new storage ModUp tests;
- all `schemes/ckks/fast` tests;
- relevant bootstrapping tests;
- `go test ./...`;
- `git diff --check`.

---

# 11. Performance rule

This is a correctness boundary.

Allowed:
- coefficient-domain reconstruction/canonicalization;
- one INTT/NTT boundary round trip when required;
- exact fixed-width residue arithmetic;
- per-component observed-bound scan.

Avoid:
- per-coefficient `math/big` if fixed-width helpers already suffice;
- logical full-RNS work;
- materialization of q1...qL rows;
- adaptive width logic.

No benchmark threshold yet.

---

# 12. Prohibitions

Do not:
- modify historical production `FastEvaluator.modUpBasis` or `ModUp`;
- wire private-F ModUp into Bootstrap;
- perform Trace;
- perform Mod1 scale alignment;
- perform Montgomery conversion;
- implement key switching;
- implement relinearization;
- implement rotation;
- implement storage contraction;
- implement adaptive storage width;
- mutate frontend CKKS parameters;
- reinterpret private F rows as logical Q rows;
- add a Standard/full-RNS fallback.

---

# 13. Success classification

Candidate classification:

`FAST_STORAGE_005_LEVEL0_MODUP_CANONICALIZATION_ACCEPTED_CANDIDATE`

Successful Codex handoff:

`READY_FOR_WEB_REVIEW`

## Completion report

Report:
- Secondary commit;
- changed files;
- API;
- canonicalization algorithm;
- evidence for `X=c+kq0 -> Center_q0(c)`;
- coefficient/NTT exactness;
- bound reset evidence;
- target-level metadata evidence;
- logical export oracle;
- Rescale -> ModUp chain;
- regressions;
- confirmation production Bootstrap/ModUp path was untouched.
