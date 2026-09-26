# QPREFIX-IMPL-003 — Prefix-Complete Primitive Kernels

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

Authoritative architecture:
- `docs/FAST_QPREFIX_SPEC.md`
- Primary `docs/QPREFIX-V2-PRODUCTION-MIGRATION-PLAN.md`

## Purpose

Generalize the existing Fast arithmetic/domain kernels so they can operate correctly on an explicit Q-prefix width from 1 to 4 rows.

This task prepares prefix-complete kernels but must **not globally activate q2/q3 consumption in existing production wrappers yet**.

Reason:

- QPREFIX-IMPL-002 now allocates policy-width backing rows;
- some existing production producers still populate only the legacy Q01/Q012 rows;
- reading every allocated row as authoritative before those boundaries are migrated would make zero/stale q2/q3 data semantic.

Therefore this task separates:

[
	ext{kernel capability}
]

from

[
	ext{production activation}.
]

Later milestones switch each subsystem only after its inputs are proven/materialized prefix-complete.

---

# 1. Explicit row-count core

Refactor relevant kernels so the arithmetic core takes or derives an explicit validated row count.

The row count must support:

[
1le wle QPrefixWidth(level)le4.
]

Provide a small internal validation/helper layer that verifies:
- requested rows do not exceed logical Level+1;
- requested rows do not exceed Q-prefix policy width;
- every requested operand/output row has N-sized backing;
- requested q subrings exist.

Do not infer authority merely from allocated/non-nil backing.

---

# 2. Kernels to generalize

At minimum cover existing functionality used by later Bootstrap stages:

- Add/Sub row loops and maintained-row copy/negation;
- integer/scalar multiplication row loops;
- pointwise Mul / MulThenAdd;
- Mul / MulRelin / Relinearize internal row copying;
- FastPartialNTT / FastPartialINTT;
- automorphism / rotation / conjugation core;
- truncate degree-2 -> degree-1 row copying;
- scalar NTT encoding values for up to four q rows.

Where practical, refactor one shared core instead of creating separate Q01/Q012/Q0123 copies.

Do not change CKKS math or public Scale semantics.

---

# 3. Transitional production rule

Existing production-facing wrappers must continue to use their currently accepted legacy authoritative-row count unless this task can prove their complete call graph already supplies all Q-prefix rows.

Default requirement:

- public/current production wrapper behavior remains byte-for-byte compatible on currently authoritative rows;
- newly allocated but not-yet-authoritative q2/q3 rows are not read or overwritten unless the wrapper already historically owned them;
- new prefix-complete core paths are exercised explicitly by focused tests.

Do **not** globally replace every `maintainedLimbCount(...)` call with `QPrefixWidth(...)`.

Later QPREFIX-IMPL-004/005/006/007 milestones will activate prefix-complete rows at explicit boundaries.

---

# 4. Scalar encoding width

The current scalar NTT helper uses a fixed three-row result.

Generalize it safely to the architecture cap of four rows or an equivalent explicit-width representation.

For every requested row:
- reduce the same encoded scalar integer modulo the actual q_i;
- preserve current half-slot/root handling;
- preserve optional Montgomery form.

Legacy callers requesting 2/3 rows must remain unchanged.

---

# 5. Exact modular oracle tests

For synthetic ciphertexts whose q0..q3 rows are deliberately populated with independent nonzero residues, exercise the prefix-complete kernels at:

- Level 0 / width 1 where the primitive logically supports Level 0;
- Level 1 / width 2;
- Level 2 / width 3;
- Level >=3 / width 4.

For each active row compare against the corresponding `ring.SubRing` operation or independent modular oracle.

Include:
- Add/Sub;
- scalar/integer multiply;
- point Mul and MulThenAdd;
- NTT/INTT;
- automorphism in NTT domain;
- copy/neg/truncate behavior;
- degree-one × degree-one Mul component formulas.

Rows outside the explicitly requested count must remain untouched.

---

# 6. Poisoned-row transition tests

Create cases where:

- q0/q1 or q0/q1/q2 are current legacy-authoritative rows;
- the newly allocated higher prefix rows contain poison values.

Run the unchanged production wrappers.

Require:
- legacy authoritative outputs remain exact;
- poison higher rows are not read as input semantics;
- wrappers do not accidentally promote those rows to authoritative results.

Then run the explicit prefix-complete core with all rows intentionally populated and require all requested rows are transformed exactly.

This test is mandatory.

---

# 7. Alias / scratch / metadata

Preserve existing:
- allowed aliases;
- evaluator-owned scratch reuse;
- Degree semantics;
- Scale;
- Level;
- IsNTT;
- IsMontgomery;
- batching/dimension metadata.

Do not add per-call full-RNS allocation.

Scratch already allocated at QPREFIX-IMPL-002 policy width may be reused, but loops must respect the explicit requested row count.

---

# 8. Bounds/capacity

This task does not attach persistent bound metadata to ciphertexts.

Use the QPREFIX-IMPL-001 capacity helpers in tests/preflight where a kernel operation requires a bounded authoritative interpretation.

Pure modular Add/NTT/automorphism do not need an integer-lift reconstruction merely to execute row-wise.

Do not invent full-path capacity evidence here.

---

# 9. Validation

Required:
- focused prefix-complete kernel tests;
- legacy poisoned-row regression;
- `go test ./schemes/ckks/fast`;
- relevant polynomial/DFT tests if touched indirectly;
- `go test ./...`;
- `git diff --check`;
- gofmt check.

Any production regression is blocking.

---

# 10. Prohibitions

Do not:
- change Rescale implementation/semantics;
- change ModUp;
- change ScaleDown;
- change DFT/LinearTransform production routing;
- change EvalMod/PS/DoubleAngle;
- change Bootstrap orchestration;
- introduce F;
- globally activate Q0123 merely because rows are allocated;
- read dormant/stale rows as authoritative;
- modify `fast-ckks`.

---

# 11. Completion report

Report:
- Secondary commit;
- changed files;
- explicit row-count/core APIs or refactors;
- which production wrappers remain legacy-activated;
- width-4 exact-oracle tests;
- poisoned-row transition evidence;
- full regression results;
- confirmation Rescale/ModUp/DFT/EvalMod/Bootstrap were not migrated.

Successful handoff:
`READY_FOR_WEB_REVIEW`.
