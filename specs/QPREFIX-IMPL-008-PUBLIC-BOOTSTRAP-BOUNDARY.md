# QPREFIX-IMPL-008 — Public Bootstrap Boundary Integration

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
- QPREFIX-IMPL-001 through 006 as previously accepted;
- QPREFIX-IMPL-007 final production handoff commit:
  `74c058ad59655f2a47efcb4faf1cf38324bd6137`.

Authoritative architecture:
- Secondary `docs/FAST_QPREFIX_SPEC.md`
- Primary `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

## Purpose

Close the remaining structural/public boundary of the Q-prefix v2 Bootstrap path:

[
	ext{public inputs}
ightarrow
	ext{packing}
ightarrow
N_1leftrightarrow N_2
ightarrow
	ext{Fast Bootstrap core}
ightarrow
	ext{unpacking}
ightarrow
	ext{public finalization}.
]

This milestone does **not** add new circuit mathematics.

Its job is to prove that compact Q-prefix authority crosses:
- packing/unpacking;
- ring-degree conversion;
- `BootstrapMany`;
- finalization/public output;

without stale-row reads, silent width truncation, or hidden full-RNS fallback.

---

# 1. Public versus internal authority

The current Fast Bootstrap public contract requires:

[
ResidualParameters.MaxLevel() le 1.
]

Therefore a valid public input/output has at most:

- Level 0: q0;
- Level 1: q01.

This is intentional.

Do **not** widen public ciphertexts to q0123 merely because the internal Bootstrap circuit uses q0123.

The boundary contract is:

[
	ext{public q0/q01}
ightarrow
	ext{internal ModUp q0123}
ightarrow
	ext{internal C2S/EvalMod/S2C}
ightarrow
	ext{public q0/q01}.
]

For ordinary public ciphertexts at Level 0 or 1, the public authoritative width is:

[
rows=QPrefixWidth(Level)=Level+1.
]

At these levels this is also the complete logical residue set.

---

# 2. Ring-degree conversion explicit-row capability

Current `ring_degree.go` has a structural bug:
- the conversion derives a maintained row count;
- but `mapRingDegreeCoefficient` silently caps work to two rows.

Remove that mismatch.

Add explicit-row ring-degree APIs, conceptually:

- `FastN1ToN2QPrefixRows(..., rows)`
- `FastN2ToN1QPrefixRows(..., rows)`

Exact names are a coding choice.

Requirements:
- support rows 1..4;
- validate:
  - equal logical Level;
  - rows <= min(Level+1, 4);
  - matching q_i moduli across N1/N2 for every requested row;
  - N2 = 2*N1 for expansion;
  - distinct input/output ciphertexts;
  - degree 1;
  - matching NTT/Montgomery state;
  - N-sized backing for every requested row;
- coefficient-domain mapping processes exactly `rows`;
- NTT-domain mapping uses `FastPartialINTTRows` / `FastPartialNTTRows`;
- rows above `rows` are untouched and never read;
- metadata is copied exactly from input.

The existing `FastN1ToN2` / `FastN2ToN1` wrappers must retain legacy authority selection for compatibility.

Do not silently promote allocated rows.

---

# 3. Ring-degree coefficient oracle

For rows = 1, 2, 3, 4:

## N1 -> N2

In coefficient domain, for every requested row and coefficient i:

[
out[2i]=in[i],qquad out[2i+1]=0.
]

## N2 -> N1

For every requested row:

[
out[i]=in[2i].
]

Compare against:
- direct coefficient oracle;
- Standard `rlwe.SwitchCiphertextRingDegree` / NTT counterpart where structurally applicable.

Cover:
- coefficient domain;
- NTT domain;
- Montgomery false/true where supported;
- q3 nonzero independent data;
- rows outside authority poisoned and unchanged.

A width-4 test must fail against the old hard-coded two-row mapping.

---

# 4. Packing/unpacking explicit-row core

Generalize the internal Fast packing primitives so authority is explicit.

Conceptually:

- `fastPackRows(..., rows)`
- `fastUnpackRows(..., rows)`
- explicit-row copy helper.

Requirements:
- process exactly `rows`;
- monomial multiply/add runs on every authoritative row;
- no row above `rows` is read/written;
- preserve Degree, Level, Scale, NTT/Montgomery, batching, bit-reversal, LogDimensions;
- packed outputs own their storage and do not alias inputs or each other.

Legacy packing helpers/stage APIs may retain their existing authority contract if still used outside the production public path.

Do not infer authority from physical backing.

---

# 5. Packing monomial tables

Packing monomial tables must be usable by explicit Q-prefix packing through width 4.

Current table generation is based on the legacy maintained width.

Generalize table materialization so the table contains at least the requested Q-prefix rows up to the fixed cap:

[
min(MaxLevel+1,4).
]

For current public Bootstrap only q0/q01 will be consumed, but the shared table/helper must not prevent explicit width-3/4 tests.

Do not generate full-Q tables beyond q3.

No private-F plaintext/table mirror.

---

# 6. Public PackAndSwitch activation

Add or use an explicit-authority packing/switch path for production `BootstrapMany`.

At entry:
1. validate all public ciphertexts;
2. require equal Level / LogSlots / representation;
3. compute:
   [
   rows=QPrefixWidth(publicLevel).
   ]
4. because public Level <=1, rows is 1 or 2;
5. pack N1 using exactly rows;
6. if N1 != N2, ring-degree convert exactly rows;
7. pack N2 using exactly rows.

The internal Bootstrap core must not see q2/q3 from the public input boundary.

Those rows are created later only by Level-0 ModUp.

Do not use backing width as authority.

---

# 7. Production unpack/switch activation

After `bootstrapCore`, S2C returns to the residual/public output level.

Before unpack:
- compute authority from the actual core output Level:
  [
  rows=QPrefixWidth(coreOutput.Level()).
  ]

For the current accepted public contract this must be at most 2.

Then:
- unpack N2 over exactly rows;
- N2->N1 convert exactly rows when required;
- unpack N1 over exactly rows.

No q2/q3 may be read merely because a temporary buffer once had wider internal storage.

---

# 8. Packing contexts

`packingContext` is primarily geometry/state:
- parameters;
- LogSlots;
- LogMaxDimensions;
- count.

If authority rows are stored in the context, they must be validated against each ciphertext Level at use time and must not become stale after the Bootstrap core changes Level.

It is acceptable—and often safer—to derive rows explicitly from the ciphertext Level at each public boundary rather than persist authority in the context.

Do not let an input-side row count be reused blindly for output-side unpacking.

---

# 9. Public input validation

Strengthen `validateFastBootstrapPublicInputs` / packing validation as needed.

For every authoritative public row require:
- correct N-sized backing;
- degree 1;
- valid Level;
- NTT=true;
- Montgomery=false at the public entry;
- positive Scale;
- consistent batching/bit-reversal/LogSlots across the batch.

Inputs must remain unchanged after `Bootstrap` / `BootstrapMany`.

Rows outside public authority, when synthetically present in a compatibility fixture, must not affect production output.

---

# 10. Finalization contract

`finalizeFastPublicCiphertext` must operate on exactly:

[
rows=QPrefixWidth(output.Level()).
]

For current residual Level <=1 this is q0/q01.

Requirements:
- IMForm exactly authoritative rows;
- leave no mixed Montgomery state;
- set `IsMontgomery=false`;
- preserve NTT=true;
- set public Scale to residual default scale;
- preserve Degree=1 and expected LogDimensions/batching metadata;
- never read q2+;
- reject malformed authoritative backing transactionally before conversion where practical.

The finalized output must be consumable by ordinary Lattigo public APIs/decryptor/encoder without Fast-specific materialization.

---

# 11. BootstrapMany structural coverage

Exercise real `BootstrapMany` with:

- count 1;
- count 2;
- odd counts such as 3 and 5;
- N1=N2;
- N2=2*N1;
- Level-0 public inputs;
- Level-1 public inputs where allowed;
- sparse/repacked slot layouts;
- full-slot layout.

Verify:
- number/order of outputs;
- no aliasing between outputs;
- inputs unmodified;
- LogSlots/LogDimensions expected;
- Degree=1;
- output N = residual N;
- output Level = residual MaxLevel;
- NTT=true;
- Montgomery=false;
- Scale=ResidualParameters.DefaultScale();
- public row backing is exactly valid for the output Level;
- no dormant-row dependence.

---

# 12. Stage API equivalence

For the same fixture compare:

A. one call to `BootstrapMany`;

B. explicit stage sequence:
- PackAndSwitchN1ToN2;
- ScaleDown;
- ModUp;
- C2S;
- EvalMod;
- S2C;
- UnpackAndSwitchN2ToN1;
- finalization.

Require the same public structural contract and decoded semantics.

If legacy stage wrappers intentionally retain legacy authority, use the explicit Q-prefix stage variants needed to mirror the production path.

Do not make a comparison that accidentally routes one side through a legacy q012 circuit helper.

---

# 13. Standard/public correctness

For fixed accepted workloads, use the same frontend/configuration style for Standard and Fast.

Required controls:
- encoded nonzero messages;
- zero-secret control;
- generated-secret comparator where the existing branch supports it;
- ordinary decryptor + encoder decode of finalized Fast output;
- current public max-component / decoded error gate remains unchanged.

Do not relax thresholds.

No Standard full-RNS execution may be used as a hidden fallback inside the Fast path.

Standard is an oracle only.

---

# 14. Dormant-row isolation

Mandatory poison tests:

## Ring-degree explicit width
- rows=4: q3 must propagate correctly;
- row 4+ poison has no effect.

## Ring-degree legacy wrapper
- non-authoritative q3 poison must not be promoted.

## Public packing
For a synthetic higher-level compatibility fixture:
- legacy rows=2 packing ignores q2/q3 poison;
- explicit rows=4 packing consumes valid q0123 independently.

## Real Bootstrap public boundary
At Level 0/1, synthetic extra backing/dormant poison must not affect result.

The production public boundary must never depend on q2+ before ModUp.

---

# 15. No full-Q allocation/fallback

Inspect the complete `BootstrapMany` path.

Forbidden:
- allocating a full logical-Q polynomial/ciphertext merely to cross packing or N1/N2;
- using Standard ring-degree conversion as production fallback;
- reading q4+;
- regenerating dormant rows from another hidden basis.

Temporary polynomials for ring-degree conversion must be capped at the explicit authority width.

---

# 16. Performance guard

This is still correctness-first, but record focused structural costs for:
- N1->N2 rows=2 vs rows=4 explicit conversion;
- N2->N1 rows=2 vs rows=4;
- packing/unpacking rows=2 vs rows=4 on the small benchmark fixture;
- `BootstrapMany` count 1 and an odd batch count on the accepted profile.

Report:
- ns/op;
- B/op;
- allocs/op.

A >10x unexplained regression compared with the closest legacy path requires `NEEDS_WEB_REVIEW`.

Do not perform the full release/performance campaign; that is QPREFIX-IMPL-009.

---

# 17. Prohibitions

Do not:
- change CKKS/public parameter chains;
- change security/noise model;
- widen public Residual MaxLevel beyond the existing <=1 contract;
- modify C2S/EvalMod/S2C mathematics;
- modify polynomial/DFT schedules;
- introduce F;
- use Standard full-RNS fallback;
- read q4+;
- globally promote legacy stage APIs unless their authority contract is explicitly migrated;
- modify `fast-ckks`.

---

# 18. Completion report

Report:
- Secondary commit;
- changed files;
- explicit-row ring-degree APIs and legacy compatibility behavior;
- width-4 ring-degree oracle evidence;
- explicit packing/unpacking authority design;
- production public input/output row counts;
- N1=N2 and N2=2N1 BootstrapMany evidence;
- finalization/public output contract;
- dormant-row poison results;
- Standard/generated/zero-secret correctness results;
- benchmark numbers;
- full regression results.

Successful handoff:
`READY_FOR_WEB_REVIEW`.

Use `NEEDS_WEB_REVIEW` if:
- public packing requires q2/q3 before ModUp;
- N1/N2 conversion cannot preserve requested q-prefix rows;
- finalization needs dormant rows;
- the ordinary public decryptor/encoder cannot consume Fast output;
- a >10x unexplained structural regression remains after one bounded repair.
