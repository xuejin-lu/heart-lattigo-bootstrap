# FAST-STORAGE-008 — C2S Stage-Local Width-2 Experiment

## Status

Executable Codex task.

## Task class

`E — Experiment / bounded implementation`

This task does not modify production Bootstrap.

Authoritative architecture:
- Secondary `docs/FAST_CKKS_SPEC.md`
- especially Sections 4.4.1, 4.4.2, 4.10.3, 4.11–4.15.

Accepted prerequisites:
- FAST-STORAGE-007 at `1a8018efda159f1dc9ab7078ce80a40c6db28e71`
- explicit stage-local width experiment rule at `04396c94a3bb063e93b00bb80a500dfcc2891e21`

## Purpose

Determine whether a width-2 private-F resident CoeffsToSlots chain is fast enough to justify a later production-integration task.

Compare three execution modes on the current supported LogN13 compressed C2S profile:

1. existing logical Fast C2S;
2. private-F width 3;
3. private-F width 2.

Production policy remains unchanged regardless of the result.

---

# 1. Storage capacities

Use the existing fixed private primes.

For width 2:

[
S_2=f_0f_1.
]

For width 3:

[
S_3=f_0f_1f_2.
]

All private states must satisfy the strict uniqueness invariant:

[
2B_j<S_w
]

for their active width (w).

---

# 2. Explicit contraction primitive

Add a standalone explicit contraction API conceptually:

`ContractStorage(targetWidth int)`

or equivalent.

Required behavior:
- only allows targetWidth < current width;
- targetWidth in {1,2};
- validate current state first;
- require every component bound to satisfy strict centered capacity in target width;
- preserve exact authoritative integer semantics;
- preserve logical Level, Scale, degree, domain, metadata, component bounds;
- non-mutating;
- copy/truncate prefix F rows only;
- no INTT;
- no CRT reconstruction;
- no NTT;
- no logical-Q materialization.

The key theorem is:

[
X leftrightarrow (Xmod f_0,ldots,Xmod f_{w-1})
]

and when (2B<S_{w'}), the prefix residues already uniquely identify the same (X).

Transactional failure is required.

---

# 3. Width-parameterized private plaintext mirrors

Generalize the FAST-STORAGE-007 plaintext mirror infrastructure so the mirror may be created at explicit width 2 or 3.

Requirements:
- authoritative integer reconstruction from full LogicalQ source is unchanged;
- exact source round-trip verification remains mandatory;
- target width capacity is checked on every plaintext coefficient:
  [
  2|P_k|<S_w;
  ]
- store exactly w F rows;
- max coefficient and L1 norms are unchanged by width;
- metadata/Scale semantics unchanged.

Do not remove width-3 APIs or tests.

---

# 4. Width-parameterized private plaintext multiplication / automorphism / LinearTransform

Generalize the standalone FAST-STORAGE-007 arithmetic to support an explicit active width matching the ciphertext and mirrored plaintext.

Required:
- ciphertext width equals plaintext/matrix width;
- loops operate over active width, not hard-coded 3;
- bound semantics unchanged:
  [
  B'_{j}le B_jL_P
  ]
  and
  [
  B'_{j}le B_jsum_dL_{P_d}
  ]
  for a LinearTransform;
- strict capacity check uses the actual active width;
- direct and BSGS exactness remain valid;
- no logical-Q fallback.

Do not generalize unrelated private-F primitives unless required by the C2S chain.

---

# 5. Standalone full C2S factor-chain experiment

Use the current actual LogN13 compressed C2S matrices and restore plan.

Construct the same deterministic bounded input fixture used by FAST-STORAGE-007.

Prepare two equivalent private-F inputs from the same authoritative lift:
- width 3;
- width 2, preferably by explicit 3->2 contraction after the accepted import/ModUp boundary.

For each of the four C2S factor groups:

1. run private-F LinearTransform at the active width;
2. run accepted `FastStorageRescale`;
3. apply the existing restore scalar using `FastStorageMulInteger`;
4. propagate exact bound metadata;
5. assert strict capacity at every intermediate state.

The LogN13 profile has one factor per group, so one Rescale per factor matches current production grouping.

Do not perform C2S final conjugate/split/repack operations in private-F unless they are already required for a fair factor-chain comparison. The primary experiment boundary is the four-factor DFT chain through its rescale/restore schedule.

If width 2 fails capacity anywhere:
- identify the exact first factor/stage;
- report bound and (S_2);
- do not widen automatically;
- classify as width-2 capacity blocked.

---

# 6. Semantic equivalence

At the end of each factor group, or at minimum after the full four-factor chain, compare:

- width-2 private-F;
- width-3 private-F;
- existing logical Fast reference.

Use explicit export/domain conversion boundaries.

Require exact maintained logical-row equality when the compared states have the same logical Level and Scale.

Also require width-2 and width-3 decoded authoritative integer lifts to match exactly at equivalent pre-export states.

---

# 7. Hard capacity assertions

Correct the non-blocking weakness found in FAST-STORAGE-007:

Feasibility/capacity tests must hard-fail on unexpected capacity violation.

Do not use only:
`Logf + return`

for a path expected to fit.

For width-3, current LogN13 accepted feasibility must remain a hard assertion.

For width-2, either:
- all expected stages hard-pass, or
- the test/report intentionally classifies the first width-2 block with explicit data and stops the experiment.

---

# 8. Benchmarks

Benchmark in the same environment:

## A. Representative factor 0
- logical Fast
- private-F width 3
- private-F width 2

Report:
- ns/op
- B/op
- allocs/op

Run at least three iterations/runs per mode when practical.

## B. Full four-factor C2S chain
Include:
- four LinearTransforms;
- each Rescale;
- each restore scalar.

Compare:
- logical Fast chain;
- private-F width 3 chain;
- private-F width 2 chain.

Report:
- ns/op
- B/op
- allocs/op.

One-time matrix mirror construction is excluded from hot-path timing but reported separately for width 2 and width 3.

---

# 9. Decision gates

This task is experimental, not production acceptance.

Correctness is mandatory.

### Width-2 capacity
Width 2 must satisfy strict capacity at every required C2S factor-chain stage.

If not:
`FAST_STORAGE_008_WIDTH2_CAPACITY_BLOCKED`

### Performance signal

Use the existing factor-0 evidence as context:
- logical Fast ~0.41 ms/op;
- width-3 private-F ~1.01 ms/op.

A width-2 result is considered promising only if both are true:

1. factor-0 width-2 is at least 20% faster than width 3:
   [
   T_{F2}le0.8T_{F3};
   ]
2. full-chain width-2 is no worse than 1.5x the logical Fast full-chain time.

If correctness/capacity pass but these performance signals fail, still commit the experiment foundation and classify:
`FAST_STORAGE_008_WIDTH2_VALID_NOT_COMPETITIVE`.

If both pass:
`FAST_STORAGE_008_WIDTH2_PROMISING`.

No production integration occurs in either case without a later Primary task.

---

# 10. Allocation design

Avoid obvious benchmark distortion:
- no per-diagonal/per-component `make([]uint64,N)` inside the deepest multiplication loop if reusable scratch can be scoped locally to an evaluation;
- preserve clear ownership and no shared mutable global scratch;
- do not launch a broad pooling refactor.

A small matrix/evaluator scratch object is allowed.

If width 2 arithmetic is slower mainly because of repeated allocation rather than arithmetic, one bounded repair pass may address that.

---

# 11. Required tests

- contraction 3->2 exactness in coefficient and NTT domains;
- contraction capacity rejection transactionality;
- width-2 plaintext mirror round-trip;
- width-2 plaintext multiply exactness;
- width-2 automorphism exactness;
- width-2 direct/BSGS LinearTransform exactness;
- width-2 vs width-3 exact equivalence;
- real LogN13 C2S width-3 hard capacity assertion;
- real LogN13 width-2 factor-chain capacity classification;
- full-chain semantic comparison where width 2 is feasible;
- existing FAST-STORAGE-001..007 regressions;
- production Bootstrap unchanged.

Run:
- `go test ./schemes/ckks/fast`
- relevant DFT/bootstrapping tests
- `go test ./...`
- `git diff --check`
- gofmt check.

---

# 12. Prohibitions

Do not:
- modify production Bootstrap/C2S routing;
- change the fixed production width-3 policy;
- enable runtime adaptive width;
- implement automatic 2->3 widening in C2S;
- add storage primes;
- change DFT matrices/scales/restore plan;
- weaken capacity bounds;
- migrate EvalMod/S2C;
- change CKKS encoder semantics;
- add full-RNS fallback.

---

# 13. Completion report

Report:
- Secondary commit;
- changed files;
- contraction API;
- width-parameterized plaintext/LinearTransform APIs;
- S2 and S3 values;
- factor-by-factor width-2 and width-3 capacity table;
- exact semantic comparison results;
- factor-0 benchmark table for Logical/F3/F2;
- full four-factor chain benchmark table for Logical/F3/F2;
- mirror-build costs;
- final classification:
  - `FAST_STORAGE_008_WIDTH2_CAPACITY_BLOCKED`, or
  - `FAST_STORAGE_008_WIDTH2_VALID_NOT_COMPETITIVE`, or
  - `FAST_STORAGE_008_WIDTH2_PROMISING`;
- confirmation production remains unchanged.

If implementation/math conflict occurs, use `NEEDS_WEB_REVIEW`.
Otherwise commit/push the experiment result and report `READY_FOR_WEB_REVIEW`.
