# QPREFIX-IMPL-007 — EvalMod, Paterson–Stockmeyer, and DoubleAngle

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
- QPREFIX-IMPL-005: `08b36b0b730dcc57eb98466594796c10bffbfdb9`
- QPREFIX-IMPL-006: `2aaec605951b4469cef6db10c28453422204595e`

Authoritative architecture:
- Secondary `docs/FAST_QPREFIX_SPEC.md`
- Primary `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

## Purpose

Migrate the full Fast EvalMod path to explicit Q-prefix authority, including:

- Mod1 input clone;
- polynomial workspace;
- generated powers;
- Paterson–Stockmeyer baby/giant steps;
- scale alignment;
- Mul / MulRelin / Relinearize;
- MulThenAdd and the one-bit scalar guard;
- every Rescale boundary;
- DoubleAngle;
- final restore.

For the accepted LogN13/P93 path, EvalMod enters at Level 12 and exits at Level 4. Therefore throughout the accepted production EvalMod path:

[
Level ge 3
]

and the Q-prefix policy remains:

[
q0123.
]

Hence q0123 authority must remain continuous from EvalMod input to EvalMod output.

After this milestone, production S2C may be activated with q0123 authority because its upstream producer will finally provide it.

---

# 1. Do not confuse storage width with polynomial schedule

The current implementation contains Q012-specific names/conditions such as:

- `postProductQ012Schedule`;
- normalized LogN13/P93 branches;
- fixed plan-scale exponents;
- specific PS / DoubleAngle level transitions.

These encode historically validated **polynomial/scale scheduling decisions**.

They do not imply that arithmetic must remain three-row.

Therefore:

- preserve the accepted polynomial, PS, level, scale, and DoubleAngle schedule unless a concrete correctness failure proves otherwise;
- migrate the arithmetic authority from legacy q012 to q0123;
- renaming misleading Q012 identifiers is allowed when it improves clarity, but schedule semantics must remain unchanged;
- do not redesign polynomial degree, factorization, target scales, restore exponents, or DoubleAngle count.

This milestone is an authority migration, not a new EvalMod algorithm.

---

# 2. Explicit authority contract

Add an explicit row-count contract through the polynomial / Mod1 stack.

For the production LogN13/P93 path:

[
rows=QPrefixWidth(input.Level())=4.
]

All workspace states derived from that input must preserve `rows=4` while Level remains >=3.

If a future profile crosses below Level 3, authority may contract only through an explicit Rescale result:

[
rows' = min(rows,QPrefixWidth(newLevel)).
]

Never infer authority from:
- allocated backing;
- `MaintainedLimbCount`;
- old Q012 profile gates.

The explicit authority must be visible at every evaluator boundary that mutates ciphertext contents.

---

# 3. Mod1 input clone

Current `cloneFastCiphertext` copies legacy maintained rows.

Replace/generalize it with an explicit-row clone.

For production EvalMod:
- copy q0123 at entry;
- preserve Level, Scale, NTT, Montgomery, batching/dimensions;
- do not read q4+;
- do not silently drop q3.

Add a poison regression:
- alter q3 while keeping q012 equal;
- explicit q0123 EvalMod input must produce a correspondingly different q3 result;
- q012 rows must remain mathematically independent from the q3 poison under row-wise arithmetic.

---

# 4. Polynomial workspace authority

Every polynomial workspace ciphertext/buffer must carry the caller-selected authoritative row count.

Generalize:
- power buffers;
- baby-step buffers;
- balanced left/right copies;
- scale scratch;
- result clone;
- maintained-copy helpers;
- zero helpers.

Required:
- copy/zero exactly `rows`;
- leave rows above `rows` untouched/dormant;
- workspace allocation may use Q-prefix physical width, but authority remains explicit;
- no helper may fall back to `MaintainedLimbCount` when operating inside migrated EvalMod.

The workspace may store `rows` once per evaluation if that is simpler and safer.

---

# 5. Explicit-row arithmetic surfaces

Use or add narrow explicit-row variants for every operation used by polynomial/Mod1.

At minimum support explicit `rows` for:
- Add/Sub with ciphertext operand;
- Add/Sub with scalar;
- Mul ciphertext×ciphertext;
- MulRelin;
- Relinearize/truncate;
- scalar Mul;
- MulInteger;
- MulThenAdd with ciphertext operand;
- MulThenAdd with scalar coefficient;
- one-bit scalar guard;
- Rescale / RescaleTo.

Prefer wrappers over the already row-parametric internal kernels introduced in QPREFIX-IMPL-003/006.

Do not globally change the legacy public wrappers.

---

# 6. Relinearization / degree handling

The current Fast zero-secret model discards c2 at relinearization.

Explicit-row Relinearize must:
- preserve q0..q(rows-1) of c0/c1 exactly according to the current Fast zero-secret semantics;
- drop c2 structurally;
- preserve Level/Scale/domain/metadata;
- never copy stale q3 from an input/output buffer.

Test degree-2 -> degree-1 with independently populated nonzero q0123 rows.

---

# 7. Generated powers

Every power in the polynomial power basis must remain prefix-complete.

For each generated power:
- both operands must have the same explicit authority;
- multiplication uses exactly those rows;
- Chebyshev doubling/subtraction uses exactly those rows;
- pre/post-product Rescale uses exactly those rows;
- after Rescale update:
  [
  rows:=min(rows,QPrefixWidth(level)).
  ]

For accepted P93, rows should remain 4 throughout all generated-power checkpoints above Level 3.

No power may silently revert to q012.

---

# 8. Preserve the accepted P93 schedule

The accepted normalized LogN13/P93 schedule is frozen for this milestone.

Preserve:
- polynomial degree;
- Paterson–Stockmeyer split/planner behavior;
- plan-scale selection;
- balanced-vs-post-product choice;
- normalized LogN13 plan bits;
- working/k/multiplier exponents;
- target-scale recurrence;
- three DoubleAngle rounds;
- final restore behavior.

If a function is named `postProductQ012Schedule`, it may be renamed to reflect the profile/schedule rather than row width.

Do not change the schedule merely because the storage authority is now q0123.

---

# 9. PS baby/giant accumulation

Migrate:
- `evaluateBabyStep`;
- `evaluateMonomial`;
- scale-aligned add/sub;
- final PS accumulation.

For every operation:
- both inputs and accumulator must use the same explicit authoritative row count;
- scale alignment integer multiplication must process all authoritative rows;
- no q3 can be left from an older accumulator value after q012 are updated;
- `MulThenAdd` must support both scalar and RLWE operands at explicit width.

Include q3 poison/staleness tests for at least:
- one baby step;
- one giant step;
- final PS result.

---

# 10. One-bit scalar guard

The normalized P93 path contains the explicit final-parent one-bit scalar guard.

Generalize it to explicit Q-prefix authority without changing its mathematical behavior.

The guard's:
- accumulator doubling;
- guarded scalar accumulation;
- centered rounded divide-by-two contraction

must process all authoritative rows.

If the centered divide-by-two implementation currently reconstructs only q01, generalize it to the caller's explicit Q-prefix using the fixed-width reconstruction foundation from QPREFIX-IMPL-004.

Do not reduce the guard's authority merely for convenience.

Test exact q0123 guard semantics against an independent centered integer oracle.

---

# 11. DoubleAngle

For every DoubleAngle round:

[
x mapsto 2x^2-c
]

using the existing Fast schedule.

Migrate all four steps to explicit rows:
1. square / MulRelin;
2. integer/scalar multiplier or doubling;
3. constant subtraction;
4. Rescale.

At each round record:
- input Level;
- rows;
- input Scale;
- post-multiply Level/Scale;
- post-constant state;
- post-Rescale Level/rows/Scale;
- exact observed centered bound per component.

For P93, rows must remain 4 through the final Level-4 output.

---

# 12. EvalMod capacity checkpoints

This milestone owns the missing EvalMod / PS / DoubleAngle capacity evidence.

Do not perform an unrelated full pipeline audit.

For the accepted LogN13/P93 production EvalMod path, capture checkpoints at least at:
- EvalMod entry;
- selected generated powers that dominate growth;
- each PS baby/giant accumulation boundary;
- polynomial result before final Rescale;
- polynomial result after final Rescale;
- each DoubleAngle round before and after Rescale;
- final restore / EvalMod output.

For each checkpoint:
- Level;
- authoritative rows;
- Scale;
- degree;
- exact observed max absolute centered coefficient for every ciphertext component;
- actual prefix product;
- strict capacity check:
  [
  2B<S_Q(Level).
  ]

If capacity fails, report the first failing checkpoint and stop with `NEEDS_WEB_REVIEW`.

Do not hide the failure with:
- full-RNS fallback;
- dormant q4+ rows;
- changed public error threshold;
- changed polynomial schedule.

---

# 13. Independent semantic oracle

Use two levels of reference:

## Row-wise modular oracle

For focused primitives/workspace transitions:
- populate q0123 independently;
- compare every row against the corresponding ring operation.

## EvalMod numerical/reference oracle

For fixed accepted P93 inputs:
- compare the migrated q0123 EvalMod decoded/public behavior against the accepted pre-migration Fast result and/or Standard logical comparator already used in existing tests;
- preserve public correctness/error gates.

Do not require q0123 residue equality with Standard full-RNS after operations whose Fast zero-secret semantics intentionally differ structurally; use the existing accepted semantic oracle at those boundaries.

---

# 14. EvalMod output contract

Successful production EvalMod output for accepted P93 must satisfy:

- Level = current accepted final Level (historically 4);
- Degree = 1;
- q0123 all authoritative;
- q0..q3 all N-sized;
- q4+ dormant;
- NTT + Montgomery;
- Scale restored exactly as the current Fast Bootstrap boundary expects;
- q3 is freshly produced by the same arithmetic history as q012, not copied stale.

Add an output poison/staleness test that would fail if q3 stopped being updated during PS or DoubleAngle.

---

# 15. Activate production S2C

Only after EvalMod has proven q0123 authority, update the Bootstrap `SlotsToCoeffs` production boundary to pass:

[
rows=QPrefixWidth(EvalModOutput.Level()).
]

For accepted P93 this should be 4.

Then production S2C must use the q-prefix-capable path implemented in QPREFIX-IMPL-006.

Remove/replace the temporary EvalMod->S2C poison guard that asserted q3 was non-authoritative.

Add the inverse test:
- valid q3 changes at EvalMod output must affect S2C q3;
- q012 behavior remains correct;
- q3 remains authoritative until natural Level contraction.

---

# 16. Legacy compatibility

Outside the migrated EvalMod/S2C path:
- legacy evaluator wrappers remain legacy-authority;
- do not globally replace `MaintainedLimbCount`;
- do not force unrelated callers to q0123.

The new explicit-row APIs should be boundary-oriented and reusable without redefining every public Fast method.

---

# 17. Regression expectations

Required:
- polynomial focused tests;
- Mod1/EvalMod focused tests;
- QPREFIX-IMPL-006 C2S/S2C tests;
- public Bootstrap generated-secret and zero-secret controls;
- accepted P93 correctness tests;
- `go test ./schemes/ckks/fast ./circuits/ckks/polynomial ./circuits/ckks/mod1 ./circuits/ckks/dft ./circuits/ckks/bootstrapping -count=1`;
- `go test ./... -count=1`;
- `git diff --check`;
- gofmt.

Any public correctness regression is blocking.

---

# 18. Performance guard

Record focused P93 benchmarks for:
- polynomial evaluation legacy q012 vs migrated q0123 where a comparable legacy helper is still available;
- EvalMod end-to-end;
- at least one representative DoubleAngle round if separable.

Report:
- ns/op;
- B/op;
- allocs/op.

A >10x unexplained regression requires `NEEDS_WEB_REVIEW`.

Do not start an optimization campaign beyond one bounded repair.

---

# 19. Prohibitions

Do not:
- redesign the polynomial/PS schedule;
- change polynomial degree or coefficients;
- change DoubleAngle count;
- change CKKS parameters;
- modify packing/N1-N2/public boundary migration (QPREFIX-IMPL-008);
- introduce F;
- use Standard full-RNS fallback;
- read q4+ dormant rows;
- weaken public correctness gates;
- modify `fast-ckks`.

---

# 20. Completion report

Report:
- Secondary commit;
- changed files;
- explicit authority propagation design;
- any renamed Q012 schedule identifiers and confirmation schedule semantics are unchanged;
- generated-power / PS / guard / DoubleAngle q3 evidence;
- checkpoint capacity table and first-failure status;
- EvalMod final q0123 contract;
- production S2C activation result;
- public correctness regressions;
- benchmark numbers;
- full regression results.

Successful handoff:
`READY_FOR_WEB_REVIEW`.

Use `NEEDS_WEB_REVIEW` if:
- any accepted EvalMod checkpoint fails strict Q-prefix capacity;
- the one-bit guard cannot be expressed with q0123 under current semantics;
- q3 cannot remain coherent through PS/DoubleAngle;
- activating q0123 S2C exposes a semantic mismatch;
- >10x unexplained regression remains after one bounded repair.
