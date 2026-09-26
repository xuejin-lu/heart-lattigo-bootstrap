# FAST-STORAGE-007 — Private-F Plaintext Mirror + LinearTransform Feasibility

## Status

Executable Codex task.

## Task class

`I — Implementation / feasibility foundation`

This task does **not** modify production Bootstrap or production C2S.

Authoritative architecture:
- Secondary `docs/FAST_CKKS_SPEC.md`
- especially Sections 4.10.3 and 4.11–4.15.

If implementation appears to require changing CKKS encoding semantics, matrix Scale semantics, fixed-width-3 storage policy, or logical Rescale semantics, stop with `NEEDS_WEB_REVIEW`.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap@main`
- orchestration/spec only.

Secondary:
- `xuejin-lu/lattigo@fast-ckks`
- implementation target.

Accepted prerequisites:
- FAST-STORAGE-006 at `532319346d8235fb42c72bfd22b57a6468675c82`
- private-F plaintext-mirror architecture at `22b9f07969af38705573686fe96a873e4bd001a3`

---

# 1. Purpose

Build the minimum private-F plaintext/linear-transform foundation required to decide whether CoeffsToSlots should become the next long-lived private-F production segment.

Deliverables:

1. a private-F plaintext mirror for encoded CKKS matrix diagonals;
2. exact one-time LogicalQ -> authoritative integer -> F conversion;
3. exact plaintext coefficient bounds;
4. standalone private-F automorphism/diagonal multiplication/addition;
5. standalone private-F linear transform, including BSGS if needed by current C2S matrices;
6. feasibility measurements on the actual Fast C2S matrices:
   - per-diagonal plaintext bounds;
   - per-factor predicted ciphertext bounds;
   - whether width 3 passes or fails before each factor/group;
   - exact semantic comparison against current Fast logical LinearTransform for fixtures.

Do **not** wire this into production Bootstrap in this task.

---

# 2. Private-F plaintext type

Introduce a private type conceptually equivalent to:

`FastStoragePlaintext`

It should contain at least:
- CKKS parameter identity/value;
- logical Level at which the source plaintext was encoded;
- plaintext Scale;
- plaintext metadata relevant to batching/dimensions;
- fixed storage width 3;
- private-F NTT rows;
- exact or proven coefficient bound information.

This is immutable after construction.

Do not reuse `FastCiphertext` for plaintexts if doing so would blur degree/metadata semantics.

---

# 3. Authoritative LogicalQ -> F mirror conversion

Input source:
- the Q polynomial from an encoded CKKS linear-transformation diagonal;
- declared `LevelQ`;
- NTT=true;
- Montgomery=true;
- Standard ring.

Required algorithm:

1. validate the source has complete logical-Q rows through `LevelQ`;
2. copy source rows to scratch;
3. for every logical q row:
   - IMForm;
   - INTT;
4. CRT reconstruct each coefficient modulo:
   [
   Q_L=prod_{i=0}^{L}q_i;
   ]
5. choose the centered representative:
   [
   P_k=operatorname{Center}_{Q_L}(r_k);
   ]
6. require:
   [
   2|P_k|<S_3
   ]
   for every coefficient;
7. encode `P_k` into f0/f1/f2;
8. NTT all three F rows;
9. record:
   - exact max coefficient magnitude:
     [
     B_P=max_k|P_k|;
     ]
   - exact L1 norm:
     [
     L_P=sum_k|P_k|.
     ]

Arbitrary precision CRT is allowed here because this is one-time matrix initialization, not the evaluation hot path.

Do not infer from q0 or q0/q1 only.

---

# 4. Exact mirror oracle

The construction must be independently checked.

For sampled/all coefficients:
- reconstruct private-F mirror back to centered integer `P_k`;
- reduce to every source logical q_i;
- reapply NTT+MForm;
- require exact equality with the original encoded Q plaintext row.

This proves the mirror represents exactly the same quantized CKKS plaintext integer polynomial.

---

# 5. Private-F plaintext multiplication

Provide a standalone primitive conceptually:

`FastStorageMulPlaintext(ct *FastCiphertext, pt *FastStoragePlaintext) (*FastCiphertext, error)`

Required semantics:

[
Y_j=X_jstar P.
]

Production policy for experiments:
- width 3;
- NTT domain;
- non-Montgomery F representation.

Bound:

[
B'_{j}le B_j L_P.
]

Before execution require:

[
2B'_j<S_3.
]

Scale:

[
Delta'=DeltacdotDelta_P.
]

Logical Level unchanged.

No reconstruction in the coefficient hot loop.

Use per-row NTT coefficient multiplication modulo f_i.

Transactional failure.

---

# 6. Private-F automorphism

Provide a standalone private-F automorphism/rotation helper if not already naturally reusable.

For NTT private-F rows:
- use `ring.AutomorphismNTTIndex`;
- same Galois element semantics as CKKS logical path;
- apply independently to each F row and ciphertext component.

Automorphism preserves:
- component bounds;
- Scale;
- logical Level;
- width 3;
- metadata.

No key switching is used, matching current insecure Fast architecture.

---

# 7. Private-F LinearTransform

Implement a standalone private-F LinearTransform for the current diagonal matrix representation.

Input:
- width-3 private-F degree-one NTT ciphertext;
- logical `lintrans.LinearTransformation`;
- private-F plaintext mirrors corresponding one-to-one with its `Vec` diagonals.

Output:
- width-3 private-F ciphertext;
- same logical Level;
- Scale multiplied by matrix Scale;
- same plaintext metadata/dimensions semantics as existing Fast LinearTransform.

For each diagonal:

[
Rot_d(X)star P_d.
]

Sum result.

Use the proven bound:

[
B'_{j}
le
B_jsum_d L_{P_d}.
]

Preflight capacity before expensive execution:

[
2B'_j<S_3.
]

If BSGS is required by the matrix:
- preserve the existing matrix BSGS schedule;
- baby/giant automorphisms may be implemented over F rows;
- semantic output must equal direct diagonal summation.

Do not use logical-Q multiplication in the private-F transform.

---

# 8. C2S feasibility audit

Using the actual Fast Bootstrapping C2S matrices produced by current `buildBootstrapCircuitData` / `prepareFastLogN13C2S`, collect for LogN13:

For every factor matrix:
- factor index;
- logical LevelQ;
- matrix Scale;
- number of diagonals;
- N1 / BSGS status;
- maximum (B_P);
- maximum (L_P);
- sum:
  [
  L_{Sigma}=sum_dL_{P_d}.
  ]

Using representative bounded ciphertext inputs from the existing Fast Bootstrap tests, report predicted per-component output bound:

[
B'_{j}=B_jL_{Sigma}.
]

Report whether:

[
2B'_j<S_3
]

passes for every factor.

Also propagate the bound through the existing group schedule conceptually:
- LinearTransform multiplies bound by factor (L_Sigma);
- logical Rescale uses the accepted 004 bound transition;
- restore scalar uses (|m|B).

This audit may be a test or benchmark/report helper, but it must be reproducible.

Do not alter the C2S matrix generation or restore plan to make it pass.

---

# 9. Exact semantic tests

## T1 — plaintext mirror round trip

For real C2S diagonals:
- mirror LogicalQ -> F;
- reconstruct F -> integer -> all q rows;
- exact source equality after NTT/MForm.

Cover both q01 and q012 matrix profiles if present.

## T2 — small synthetic plaintext mirror

Use known integer encoded polynomials including:
- 0;
- +/-1;
- values above single q modulus;
- negative values;
- values near safe S3 capacity.

Verify CRT centering and F rows exactly.

## T3 — MulPlaintext exactness

Construct bounded private ciphertext + known plaintext integer polynomial.

Compare decoded private-F result against independent negacyclic integer convolution.

Verify bound and Scale.

## T4 — automorphism exactness

Private-F rotation/automorphism vs independent signed coefficient permutation.

NTT path required.

## T5 — direct LinearTransform exactness

Small matrix with several diagonals.

Compare private-F decoded output against independent integer:
[
sum_d Rot_d(X)star P_d.
]

## T6 — private-F vs current logical Fast LinearTransform

For representative C2S factor(s):
1. begin from the same logical bounded ciphertext state;
2. private route: import/mirror/transform/export compact logical;
3. current route: existing `fast.Evaluator.LinearTransform`;
4. compare maintained logical rows exactly when representation/scales match.

If current logical path is Montgomery and private F is ordinary, compare only after correct domain conversion; do not reinterpret residues.

## T7 — capacity rejection

Construct a transform where input fits but:
[
2B_jL_Sigmage S_3.
]

Require preflight failure and input immutability.

## T8 — actual LogN13 C2S feasibility

Produce the reproducible factor-by-factor capacity report described in Section 8.

This is mandatory.

## T9 — regressions

Run:
- `go test ./schemes/ckks/fast`
- relevant DFT/bootstrapping tests;
- `go test ./...`
- `git diff --check`
- gofmt check.

Production Fast Bootstrap must remain unchanged.

---

# 10. Performance measurement

This task is foundation/feasibility, but record useful costs:

For at least one representative LogN13 C2S factor:
- one-time plaintext mirror build time and bytes;
- private-F LinearTransform ns/op, B/op, allocs/op;
- existing logical Fast LinearTransform ns/op, B/op, allocs/op.

No production speed gate yet.

The decision criterion is:

1. correctness exact;
2. width-3 feasibility for the real C2S factor chain;
3. no obviously pathological (>10x) per-factor slowdown without explanation.

If real C2S capacity fails, do **not** attempt production integration. Report `NEEDS_WEB_REVIEW` with the first failing factor and exact bound numbers.

---

# 11. Prohibitions

Do not:
- modify production Bootstrap/C2S;
- change DFT matrices/scales/restore plan;
- add adaptive private storage width;
- add new storage primes;
- change CKKS encoder rounding semantics;
- infer plaintext lift from q0 only;
- add full-RNS fallback;
- implement KeySwitch/Relinearize;
- migrate EvalMod/S2C;
- weaken any existing correctness threshold.

---

# 12. Success classification

If standalone foundation is correct and real LogN13 C2S width-3 feasibility passes:

`FAST_STORAGE_007_PRIVATE_F_LINEAR_TRANSFORM_FEASIBLE_CANDIDATE`

If foundation is correct but actual C2S width-3 capacity fails:

`FAST_STORAGE_007_FOUNDATION_VALID_C2S_CAPACITY_BLOCKED`

Use `NEEDS_WEB_REVIEW` for the blocked case.

Successful feasible handoff:

`READY_FOR_WEB_REVIEW`

## Completion report

Report:
- Secondary commit if feasible and committed;
- changed files;
- plaintext mirror type/API;
- CRT mirror algorithm;
- plaintext max/L1 bounds;
- MulPlaintext/automorphism/LinearTransform APIs;
- exact semantic test results;
- LogN13 factor-by-factor C2S capacity table;
- first failing factor if any;
- representative benchmark comparison;
- confirmation production Bootstrap/C2S remained unchanged.
