# QPREFIX-IMPL-006 — Q-Prefix LinearTransform and DFT/C2S/S2C

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

Authoritative architecture:
- Secondary `docs/FAST_QPREFIX_SPEC.md`
- Primary `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

## Purpose

Migrate LinearTransform and the DFT factor chain to explicit Q-prefix execution.

This milestone has an asymmetric activation boundary:

### Production C2S

The accepted Bootstrap entry now returns authoritative q0123 from ModUp when Level >=3.

Therefore production C2S is allowed and required to preserve the full current Q-prefix through:
- LinearTransform;
- automorphism/rotation/conjugation;
- plaintext diagonal multiply;
- Add/Sub/scalar operations;
- group Rescale;
- restore scaling.

### Production S2C

EvalMod is not migrated until QPREFIX-IMPL-007.

Therefore this milestone must make S2C/DFT **q-prefix capable**, but must not assume the current production EvalMod output already carries q0123 authority.

Production S2C may remain legacy-activated until QPREFIX-IMPL-007 explicitly hands it a prefix-complete input.

This continues the architecture rule:

[
kernel capability 
e production activation.
]

---

# 1. Explicit-width LinearTransform core

Generalize `schemes/ckks/fast/linear_transform.go` so direct and BSGS paths accept an explicit authoritative row count.

Conceptually provide:

`LinearTransformQPrefixRows(ctIn, matrix, rows, ctOut)`

Exact API naming is a coding choice.

Requirements:
- (1 <= rows <= QPrefixWidth(level));
- validate every requested ciphertext row and plaintext diagonal row;
- direct path and BSGS path process exactly `rows`;
- baby rotations process exactly `rows`;
- giant rotations process exactly `rows`;
- plaintext multiplication processes exactly `rows`;
- inner/outer/term accumulation processes exactly `rows`;
- output copy processes exactly `rows`;
- rows above `rows` are untouched and never read.

The existing legacy `LinearTransform` wrapper must remain available and retain its current authority selection until callers migrate.

Do not infer authority from backing allocation.

---

# 2. Plaintext matrix policy

Do not introduce any private plaintext representation.

The CKKS DFT matrices already contain logical q rows.

For explicit width `rows`, consume:

[
q_0,ldots,q_{rows-1}
]

directly from every diagonal plaintext.

Validate:
- matrix LevelQ is sufficient;
- every requested plaintext row has N-sized backing;
- NTT/Montgomery flags match the Fast contract.

Do not CRT-reconstruct plaintexts merely to obtain q3.

---

# 3. Explicit Q-prefix arithmetic surfaces needed by DFT

The C2S wrapper performs more than LinearTransform.

After/between factors it may use:
- Rescale;
- integer restore multiplication;
- Conjugate/Rotate;
- Add/Sub;
- scalar/complex multiplication;
- copy.

Add or expose narrow explicit-row variants as needed so migrated DFT never falls back to a legacy wrapper while q3 is authoritative.

At minimum the migrated DFT path must be able to request explicit `rows` for:
- Rescale / RescaleTo, reusing QPREFIX-IMPL-004;
- integer multiplication, reusing QPREFIX-IMPL-005;
- automorphism/conjugation/rotation, reusing QPREFIX-IMPL-003 kernels;
- Add/Sub;
- scalar multiply used by C2S real/imag splitting/repacking.

Do not globally change the legacy wrappers.

The exact API surface should stay small and boundary-oriented rather than duplicating the entire evaluator.

---

# 4. DFT explicit-row execution

Add an explicit-row DFT execution path.

Conceptually:

`dftQPrefixRows(..., initialRows)`

At each factor/group:
1. current `rows` must equal the authoritative width of the current state;
2. LinearTransform processes exactly `rows`;
3. group Rescale uses exactly `rows` as source authority;
4. after Rescale:
   [
   rows := min(rows, QPrefixWidth(newLevel));
   ]
5. restore integer scaling processes exactly the updated `rows`.

This must correctly handle:
- Level >=3: q0123 stays q0123;
- 3->2: q0123 -> q012;
- 2->1: q012 -> q01;
- 1->0 where applicable.

Never keep q3 authoritative after a 3->2 Rescale.

---

# 5. Production C2S activation

`CoeffsToSlotsNewWithRestorePlan` in the Fast Bootstrap production path must recognize that the post-ModUp input is prefix-complete.

At production C2S entry:

[
rows=QPrefixWidth(ctIn.Level()).
]

For the accepted LogN13/P93 path at high Level this is 4.

The entire C2S path must preserve Q-prefix authority until natural Level contraction.

This includes:
- initial active copy;
- all factor LinearTransforms;
- group Rescales;
- restore multiplies;
- conjugation;
- real/imag Add/Sub;
- (-i) scalar multiply;
- optional repack rotation/add.

After this migration there must no longer be a deliberate q3-authority loss at the first C2S LinearTransform.

Remove/replace the QPREFIX-IMPL-005 transition test that documented that temporary loss.

---

# 6. C2S group-by-group oracle

Use the accepted QPREFIX-AUDIT-002 fixture shape.

For every C2S group record/verify:
- input logical Level;
- authoritative row count;
- raw LinearTransform q-prefix rows;
- post-Rescale Level;
- post-Rescale row count;
- restore output when present;
- Scale transitions.

Compare every authoritative row against an independent Standard/reference path where structurally valid.

At high Level, q3 equality is mandatory.

The test must fail if q3 is copied stale across a LinearTransform.

---

# 7. C2S capacity evidence

Do not reopen a global readiness audit.

For the fixed accepted LogN13/P93 C2S workload:
- reconstruct the authoritative centered lift at the existing group checkpoints;
- compute exact observed component maxima;
- verify strict:
  [
  2B<S_Q(level)
  ]
  at every required raw/post-Rescale/restore state.

Reuse accepted Audit-002 checkpoint structure and known C2S evidence where possible.

If a newly authoritative q3 path contradicts the existing mathematical C2S result or exposes a concrete capacity failure, report `NEEDS_WEB_REVIEW` with the first failing group.

Do not fall back to full RNS.

---

# 8. S2C q-prefix capability

Generalize the shared DFT / LinearTransform machinery so S2C can run prefix-complete when given a prefix-complete input.

Add focused synthetic or direct S2C tests where the input q0123 rows are all authoritative and nonzero.

Require:
- q3 is transformed exactly;
- natural Rescale contractions are correct;
- direct/BSGS behavior matches reference.

However, do not force the current full Bootstrap production S2C to use q0123 if its input comes from still-legacy EvalMod.

Production S2C activation will be completed by QPREFIX-IMPL-007 or its handoff to S2C.

---

# 9. S2C legacy transition guard

Until QPREFIX-IMPL-007 is complete:

- current EvalMod output authority remains whatever the accepted legacy EvalMod producer actually guarantees;
- production S2C must select that explicit authority, not backing width;
- allocated q3 must not be promoted automatically.

Add a transition test analogous to earlier poisoned-row tests:
- poison non-authoritative q3 at the current EvalMod->S2C boundary;
- prove production S2C does not consume it before 007.

Separately prove the explicit q0123 S2C-capable path does consume a deliberately valid q3.

---

# 10. Direct and BSGS coverage

Both LinearTransform paths are mandatory:
- direct `N1 == 0`;
- BSGS `N1 != 0`.

For each:
- width 3 legacy compatibility;
- width 4 exact row oracle;
- in-place and out-of-place where supported;
- baby/giant rotation counts/regressions;
- q3 poison isolation for legacy wrapper;
- q3 exactness for explicit wrapper.

Do not allow one path to remain q012-only.

---

# 11. Domain and metadata

Preserve throughout migrated C2S:
- NTT;
- Montgomery;
- Level;
- Scale;
- Degree;
- IsBatched;
- IsBitReversed;
- LogDimensions.

Every authoritative row must share the same domain state.

Do not convert q3 independently or leave mixed-domain rows.

---

# 12. Scratch/allocation

Reuse evaluator-owned four-row scratch created by QPREFIX-IMPL-002/003.

Do not create a new per-diagonal full-Q temporary.

Baby rotations, accumulators, and DFT scratch should remain capped at Q-prefix width.

Rows above q3 remain dormant.

---

# 13. Regression expectations

Required:
- all existing legacy LinearTransform tests;
- all existing DFT tests;
- accepted C2S checkpoint tests;
- QPREFIX-IMPL-005 ModUp boundary tests;
- current EvalMod tests;
- public Bootstrap generated-secret and zero-secret controls;
- `go test ./schemes/ckks/fast ./circuits/ckks/dft ./circuits/ckks/bootstrapping -count=1`;
- `go test ./... -count=1`;
- `git diff --check`;
- gofmt.

Any public correctness regression is blocking.

---

# 14. Performance guard

Record focused LogN13/P93 benchmarks for:
- legacy q012 LinearTransform versus explicit q0123 LinearTransform;
- one representative C2S group or full C2S;
- direct and BSGS if both are used by accepted matrices.

Report:
- ns/op;
- B/op;
- allocs/op.

A >10x unexplained regression requires `NEEDS_WEB_REVIEW`.

Do not optimize beyond one bounded repair.

---

# 15. Prohibitions

Do not:
- migrate EvalMod/PS/DoubleAngle;
- claim production S2C q0123 authority before its producer provides it;
- modify public packing/N1-N2 boundaries;
- change CKKS parameters, DFT factorization, polynomial schedule, or public error gates;
- introduce F;
- use Standard full-RNS fallback;
- read dormant q4+ rows;
- globally replace every legacy wrapper;
- modify `fast-ckks`.

---

# 16. Completion report

Report:
- Secondary commit;
- changed files;
- explicit-row LinearTransform/DFT APIs;
- production C2S activation behavior;
- exact q3 C2S oracle/checkpoint evidence;
- C2S capacity results;
- S2C q-prefix capability evidence;
- current production S2C authority and why it remains legacy or was safely activated;
- direct/BSGS coverage;
- benchmark numbers;
- full regressions.

Successful handoff:
`READY_FOR_WEB_REVIEW`.

Use `NEEDS_WEB_REVIEW` if:
- q3 C2S differs from the independent logical/reference result;
- a C2S capacity checkpoint fails;
- current EvalMod->S2C authority cannot be stated unambiguously;
- >10x unexplained regression remains after one bounded repair.
