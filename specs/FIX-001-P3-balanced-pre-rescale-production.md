# FIX-001-P3 — Production balanced pre-Rescale Fast power scheduling

## Goal

Replace the incomplete one-sided P2 scheduling experiment with the balanced two-sided pre-Rescale schedule that has now been validated on the exact formal LogN13 T3 path.

Authoritative diagnostic evidence at Primary commit `a5e635a5684789a93771a5c7e6d3efd063c0d1d2` established:

- current P2 gate selects `preRescale=false` for formal T3, so the old unsafe path remains active;
- safe balanced factors exist for the formal T1/T2 maintained state;
- selected factors: `m1 = m2 = 1073741824 = 2^30`;
- `m1*m2 = 2^60`, differing from the formal logical divisor by only ~`2.98e-13` relative;
- c0 and all-maintained-components feasibility checks pass;
- post-Rescale T1 preservation error ~`9.20e-8`;
- post-Rescale T2 preservation error ~`6.87e-9`;
- balanced product has zero q0/q1 capacity violations;
- actual-scale product error ~`1.84e-7`;
- exact-planner-scale interpretation error ~`1.84e-7`;
- final T3 correction error ~`1.84e-7`;
- classification: `balanced_pre_rescale_validated`.

This task must productionize that schedule generally enough for the bounded Fast polynomial surface, then run the complete formal LogN13 correctness gate chain.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`a5e635a5684789a93771a5c7e6d3efd063c0d1d2`

Secondary repository: `xuejin-lu/lattigo`

Expected Secondary starting HEAD:

`f2b89ed0ade9efb5f99fce3949ffeba87cd3d2fd`

Relevant incomplete P2 production commit in its history:

`ceb5482c1f5e5ffaccd4d09ec835ea6c1bd4f04b`

Prior FIX-001 repaired base:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Pinned Standard reference:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

Follow both repositories' `AGENTS.md` and Secondary `docs/FAST_CKKS_SPEC.md`.

If either worktree is dirty or cannot be safely synchronized, stop and report. Never reset/stash/discard unrelated work automatically.

The parent FIX-001 remains incomplete until independent review after this task.

---

# Why P2 must be replaced, not merely have its threshold removed

P2 introduced one-sided pre-Rescale:

```text
chosen' = Rescale(chosen)
out     = other * chosen'
```

but guarded it with roughly:

```text
chosen.Scale > rescaleDivisor * 2^10
```

Formal T1/T2 are both near `2^60` scale and the logical divisor is also near `2^60`, so the gate is false and the formal bug case bypasses the repair.

Removing the gate is not acceptable: one-sided pre-Rescale would reduce the chosen operand to scale near one and destroy encoding precision.

Therefore remove/replace the **one-sided scheduling concept** itself for the high-scale formal path.

---

# Balanced schedule identity

At common logical level `L`, the old planner semantics are:

```text
old = Rescale(left * right)
```

with target metadata scale:

`Delta_target = Delta_left * Delta_right / q_L`.

Balanced scheduling uses independent copies:

```text
left'  = Rescale(m1 * left)
right' = Rescale(m2 * right)
out    = left' * right'
```

so:

`Scale(out) = Delta_left * Delta_right * m1*m2 / q_L^2`.

Choose bounded positive integers with:

`m1*m2 ≈ q_L`.

Then the actual output scale is extremely close to the old planner target while neither operand is forced down to scale ~1.

Both operand Rescales happen from the same logical level `L` to `L-1`; therefore arithmetic cost includes two Rescale operations but **logical multiplicative depth still consumes one level**.

---

# Supported rescale-depth scope

The current Fast polynomial/Bootstrap profile uses:

`LevelsConsumedPerRescaling() == 1`.

For this production repair, explicitly support the balanced branch only when `LevelsConsumedPerRescaling() == 1`.

If the Fast polynomial surface is invoked with a parameter set where rescaling consumes more than one level, do not silently invent a generalized factorization. Either retain a proven existing safe path when applicable or return a clear unsupported/error condition from the balanced scheduling helper.

Do not overgeneralize untested multi-level rescaling.

---

# Deterministic factor selection

For divisor `q = params.Q()[commonLevel]`, choose a deterministic near-square pair using integer arithmetic only.

Preferred algorithm:

1. `m = floor(sqrt(q))` using exact uint64/integer arithmetic, never float64;
2. form a small constant candidate set around `m` and nearest integer quotients of `q/m`;
3. choose `(m1,m2)` minimizing `abs(m1*m2 - q)`;
4. stable tie-breaker: smaller max(m1,m2), then lexicographic order.

For the formal q used by T3 this must reproduce:

`m1 = m2 = 1073741824`.

No unbounded search.

No per-coefficient `big.Int` scan in production.

Factor selection depends only on public parameter metadata.

---

# Balanced branch eligibility

Do not retain the P2 condition `chosen.Scale > q*2^10`.

Instead predict both post-Rescale operand scales:

```text
Delta_left_bal  = Delta_left  * m1 / q
Delta_right_bal = Delta_right * m2 / q
```

Use a fixed, documented minimum precision floor of:

`2^20`

for the balanced branch in this bounded Fast polynomial surface.

Balanced branch is eligible only if both predicted scales are >= `2^20`.

If either falls below the floor, use the existing low-scale schedule:

```text
Mul/MulRelin -> Chebyshev double -> Rescale -> correction
```

This preserves existing low-scale compatibility tests where balanced pre-Rescale would unnecessarily collapse operand precision.

This branch decision must depend only on public Levels/Scales/moduli, never ciphertext coefficient contents.

---

# Balanced production path

When eligible:

1. compute `commonLevel = min(left.Level(), right.Level())`;
2. create independently-owned maintained copies of both operands at exactly `commonLevel`;
3. do not mutate `pb.values[a]` or `pb.values[b]`;
4. for left copy:
   - `MulIntegerMaintained(leftCopy, m1, leftCopy)`;
   - set `leftCopy.Scale = left.Scale * m1`;
   - `Rescale(leftCopy,leftCopy)`;
5. for right copy:
   - `MulIntegerMaintained(rightCopy, m2, rightCopy)`;
   - set `rightCopy.Scale = right.Scale * m2`;
   - `Rescale(rightCopy,rightCopy)`;
6. require both copies end at the same logical level `commonLevel-1`;
7. run the same strict/lazy multiplication mode the planner previously required:
   - strict => `MulRelin`;
   - lazy => `Mul` after the same existing operand-relinearization semantics;
8. for Chebyshev, double the product;
9. do **not** perform another Rescale for this power-generation step;
10. apply the existing post-Rescale Chebyshev correction:
   - `c==0`: subtract scalar 1;
   - `c>0`: subtract generated `T_c` via the existing aligned correction path.

No second logical level may be consumed.

---

# Planner-scale normalization

Let:

`actualScale = left.Scale * right.Scale * m1*m2 / q^2`

and:

`targetScale = left.Scale * right.Scale / q`.

Before rewriting metadata, require actual and target scales to be sufficiently close using the existing CKKS Scale precision conventions. Prefer the same tolerance style already used by Fast polynomial planner/evaluateMonomial (for example `Scale.InDelta` with the existing `rlwe.ScalePrecision` margin) rather than an ad hoc float64 comparison.

If the scales are not within the accepted precision tolerance, return a clear error rather than silently relabeling.

When within tolerance, normalize the balanced product metadata Scale to the exact old planner `targetScale` before correction/planner reuse.

The formal diagnostic already proved both actual-scale and exact-planner-scale interpretations pass by a large margin.

---

# Workspace ownership

Replace the single P2 `preRescale` scratch with independently-owned left/right balanced scratch storage as needed.

Requirements:

- source powers remain unchanged;
- left/right copies cannot alias each other;
- output cannot alias either source power;
- q0/q1 rows only;
- dormant q2+ rows are never read/materialized;
- preserve NTT/Montgomery/domain metadata.

Keep `copyMaintainedAtLevel` if useful and correct.

---

# Remove obsolete P2 logic

Remove the one-sided P2 scheduling heuristic and comments that claim the one-sided pre-Rescale path is the production capacity solution.

Specifically remove/replace logic equivalent to:

- choose one operand by larger scale;
- `preRescaleThreshold = rescaleFactor * 2^10`;
- one-sided preRescale branch;
- fallback semantics expressed as a compatibility exception to that one-sided design.

Historical commits remain unchanged; only current production source/docs should describe the balanced architecture.

---

# Production constraints

Do not:

- materialize q2+;
- call Standard/full-Q polynomial arithmetic as fallback;
- add auxiliary limbs;
- use per-coefficient production CRT/capacity scans;
- use arbitrary `big.Int` coefficient reconstruction in the hot path;
- change Bootstrap/Mod1 parameters;
- lower Mod1LogScale;
- special-case T3 only;
- add another logical level;
- change the Standard implementation.

Integer factor arithmetic itself may use ordinary scalar integer types / `big.Int` only as needed by existing `MulIntegerMaintained` API; this is not coefficient CRT.

---

# Secondary regression tests

Commit tests with the production repair.

## 1. Factor selection

For the exact formal divisor used by the LogN13 T3 path, assert deterministic factors:

`m1=m2=1073741824`.

Assert factor product relative error is within the scale-normalization tolerance.

Test at least several nearby 60/61-bit prime/modulus values to avoid one-value hardcoding.

## 2. Branch selection

Add tests showing:

- formal/high-scale operands choose balanced scheduling;
- low-scale operands whose predicted post-Rescale scale is below `2^20` retain the old post-product Rescale path.

Expose branch decision through an internal helper/unit test, not production logging.

## 3. Source immutability / scratch ownership

Prove balanced generation does not mutate stored `T_a/T_b` powers and left/right scratches do not alias.

## 4. Formal-scale T2

Require corrected T2 component error <= `1e-2`.

This task must prove balanced scheduling does not regress the already-fixed `c==0` case.

## 5. Formal-scale T3

Use a realistic high-scale input sufficient to trigger balanced scheduling and require T3 <= `1e-2`.

Test must fail on current `f2b89ed...` behavior and pass after P3.

## 6. c>0 and c==0

Cover both T2 and T3.

## 7. Lazy mode

Exercise at least one lazy generated power under balanced eligibility and assert final Degree/Level/Scale contracts.

## 8. Planner metadata

For balanced strict/lazy cases, assert:

- output Level equals old post-product-Rescale Level;
- normalized output Scale equals old planner target within exact Scale semantics;
- Degree matches strict/lazy planner behavior;
- one logical level is consumed.

## 9. Dormant rows

Poison q2+ source storage where available and prove balanced result depends only on maintained q0/q1 rows.

## 10. Full suite

Run:

`go test ./...`

and require all existing Fast/Bootstrap/Rescale/polynomial tests green.

---

# Documentation update

Update `docs/FAST_CKKS_SPEC.md` to supersede the incomplete one-sided P2 text.

Document:

- multiplication itself can violate q0/q1 centered capacity;
- one-sided pre-Rescale is unsuitable when operand scale is only about the rescale divisor because it collapses precision;
- high-scale Fast Chebyshev power generation therefore uses balanced two-sided integer pre-scaling + parallel Rescale;
- factors are deterministic, modulus-derived, content-independent, and near `sqrt(q_L)`;
- low-scale cases where balanced operand scale would fall below the fixed precision floor retain the old post-product Rescale schedule;
- balanced branch consumes one logical level, although it performs two physical Rescale operations;
- output Scale is normalized to the old planner target only after an explicit Scale-closeness check;
- this remains a bounded Fast polynomial rule, not a universal arbitrary-CKKS multiplication theorem.

---

# Secondary workflow

Starting clean at `f2b89ed...`:

1. inspect current `fast.go`, P2 tests, diagnostic surface and docs;
2. replace incomplete one-sided P2 production scheduling with balanced scheduling;
3. run focused tests;
4. run `go test ./...`;
5. inspect diff for no Standard/full-Q/q2+ fallback;
6. commit and push `fast-ckks` by ordinary fast-forward;
7. record exact new Secondary SHA.

Do not force push.

---

# Primary formal validation gates

Use the exact new Secondary commit for all gates.

Formal workload remains unchanged: LogN13, all 4096 slots, same config/input, corrected `T0=1` oracle.

Stop on the first failing gate.

## Gate 1 — branch/provenance + T2/T3

Record that formal T2/T3 actually select the balanced production branch.

Require:

- E2 exact;
- T1 pass;
- T2 pass <= `1e-2`;
- T3 pass <= `1e-2`.

Also record formal selected `(m1,m2)` and output Level/Scale.

## Gate 2 — all generated powers

Re-run actual formal power generation.

Validate every generated power used by the degree-30 run, historically:

`T1, T2, T3, T4, T6, T8, T16`.

Use actual generated set if synchronized code differs.

Every power must pass <= `1e-2`.

## Gate 3 — whole degree-30 polynomial

Actual Fast polynomial evaluator on formal E2 versus source-backed plaintext polynomial oracle.

Require <= `1e-2`.

## Gate 4 — public EvalMod real + imag

Against the already-validated Standard q0/q1 stage oracle.

Both branches must pass <= `1e-2`.

## Gate 5 — full formal LogN13 Bootstrap correctness

All 4096 slots:

- Standard generated-secret vs input;
- Fast zero-secret vs input;
- Fast zero-secret vs Standard output.

Require both max real and max imag <= `1e-2` for every required comparison.

If any gate fails, stop. No LogN16 or benchmark.

---

# Required Primary artifacts

Do not overwrite historical FIX-001/P2 diagnostics.

Create:

- `results/FIX-001-P3-POWERS-logN13-fast.json`
- `results/FIX-001-P3-EVALMOD-logN13-fast.json`
- `results/FIX-001-P3-CORRECTNESS-logN13-standard.json`
- `results/FIX-001-P3-CORRECTNESS-logN13-fast.json`
- `results/FIX-001-P3-summary.json`

Summary must include:

- old Secondary `f2b89ed...`;
- new Secondary commit;
- exact changed Secondary files;
- formal balanced factor pair;
- branch-selection evidence;
- T2/T3 metrics;
- all-power metrics;
- whole polynomial metrics;
- EvalMod real/imag metrics;
- full Bootstrap metrics;
- planner Level/Scale evidence;
- tests run;
- first failing gate or `ALL_LOGN13_GATES_PASS`;
- Primary/Secondary clean-state evidence.

---

# Completion criteria

Mark P3 COMPLETE only if:

1. obsolete one-sided P2 scheduling is replaced by balanced production scheduling;
2. formal high-scale path demonstrably takes balanced branch;
3. no q2+/full-Q/auxiliary-limb production fallback is introduced;
4. T2 passes;
5. T3 passes;
6. lazy/strict metadata contracts pass;
7. all formal generated powers pass;
8. whole formal degree-30 polynomial passes;
9. EvalMod real and imag pass;
10. full formal LogN13 Bootstrap correctness passes;
11. Secondary `go test ./...` passes;
12. both repositories end clean and pushed normally.

If all pass, report parent FIX-001 as **candidate for closure pending independent review**. Do not silently close it.

---

# Non-goals

Do not:

- run LogN16;
- benchmark performance;
- start EXP-003;
- change parameters;
- add active q2+;
- add auxiliary modulus architecture;
- change Standard Lattigo;
- broaden Fast CKKS guarantees beyond the tested bounded polynomial surface.
