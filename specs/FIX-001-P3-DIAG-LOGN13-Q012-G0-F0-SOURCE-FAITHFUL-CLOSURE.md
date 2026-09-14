# FIX-001-P3-DIAG-LOGN13-Q012-G0-F0-SOURCE-FAITHFUL-CLOSURE

## Goal

Determine whether the previous q012 PS-window failure was caused by a **diagnostic candidate implementation mismatch** or by a real architectural incompatibility.

The previous candidate must **not** be treated as evidence that q012 arithmetic is insufficient. Its capacity was ample, but its handwritten giant-step recurrence did not faithfully mirror the source PS tree and already diverged at `G0-add`.

This task has two stages:

1. make a source-faithful q012 mirror of the actual `G0` merge and prove it operation-by-operation;
2. only after `G0-add` is exact, carry q2 along the **actual dependency path** `G0-add -> G3-add -> F0-root -> F0-final-rescale` and test system sufficiency.

Primary-only diagnostic/design work. Secondary remains read-only.

---

# Required provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required ancestor:

`997caf2caa9657812cc2c8b9b9821ba9dff71bf4`

Start clean on `main`, fetch + ff-only pull, then read fresh:

- `AGENTS.md`
- `CURRENT_TASK.md`
- this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required exact commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Branch: `fast-ckks`

State: clean.

Secondary is read-only.

---

# What the previous task actually proved

Accepted:

- B (`q0=55`, planScale `2^92`) remains catastrophic:
  - EvalMod real ≈ `102.29911082927852`
  - EvalMod imag ≈ `102.2990271839068`
  - public-like ≈ `3662.9000547146406`.
- D (`q0=56`, planScale `2^92`) passes:
  - public-like `0.005554220603853743`.
- principal traces now use accepted generated temporary-q012 powers.
- one B-only capacity frontier was identified on the dependency graph:
  - unsafe node: `G0-add`
  - last Q01-unique ancestor: `G0-rescale`
  - dependency path: `G0-add -> G3-add -> F0-root -> F0-final-rescale`
  - max intended Q01 ratio ≈ `1.558947601645921`
  - max intended Q012 ratio ≈ `2.83570806943308e-12`.
- Q012 therefore has overwhelming capacity margin on this path.

Not accepted as architecture evidence:

- previous handwritten q012 candidate public-like ≈ `94423.82`;
- previous classification `first_remaining_noncapacity_mismatch` as evidence against q012 itself.

Reason: the candidate's recurrence was not source-faithful.

The previous candidate serially carried one `parent` through `G0/G1/G2/G3`, while the actual replay is a tree. Source-backed dependency examples show:

- `G2-add` consumes `B2-term-2` and `G2-multiply`;
- `G3-add` consumes `G0-add` and `G3-multiply`;
- `F0-root` consumes `G3-add`.

Therefore `G0-add` does **not** feed G1/G2 as a serial accumulator.

Also, the previous artifact recorded `minimal_causal_reset_set=[G0-add]` even though the reset did not collapse the error. That field is not a proven minimal causal set and must not be carried forward.

---

# Source semantics that must be mirrored exactly

Use source evidence, not inferred algebra.

## Actual PS giant merge

From the existing source-backed replay:

1. `b.value` is Fast-Rescaled;
2. `eval.Mul(b.value, powers[deg], b.value)`;
3. degree-1 × degree-1 Fast `Mul` produces degree 2 and retains `c2`;
4. `psRescaleGuardAddAligned(params, eval, a.value, b.value, ...)` aligns scales using its exact integer-ratio policy;
5. add result remains at max operand degree;
6. the tree is regrouped; root is relinearized only when required by the actual replay.

## Fast zero-secret semantics

At Secondary baseline:

- `Mul` degree1×degree1 retains `c0,c1,c2`;
- `MulRelin` would omit `c2`, but the source replay uses `Mul` at these giant steps;
- `Add` preserves the higher degree and requires equal scales;
- `Relinearize` is current zero-secret degree2→degree1 truncation;
- do not invent Standard key-switching semantics.

## Scale alignment

Mirror `psRescaleGuardAddAligned` exactly.

If scales differ:

- determine which operand has smaller scale;
- compute the exact scale ratio;
- use `ratio.BigInt()` exactly as source does;
- multiply the smaller-scale ciphertext by that integer;
- set its metadata scale exactly to the larger target scale;
- then add.

Record the fractional difference between exact ratio and integer ratio.

Do not replace this with a generic `nativeQ012NativeAlign` unless it is independently proven byte/row-equivalent to the source policy.

---

# Fixed experiment stack

Run both profiles:

## B

- LogN13
- diagnostic q0=55
- planScale=2^92
- accepted generated temporary-q012 powers
- accepted B4 T2 q012 guard
- accepted G0/F0/DoubleAngle local-q2 guard schedule downstream
- same polynomial decomposition/workload.

## D semantic calibration control

Identical except diagnostic q0=56.

D is especially important because its `G0-add` is Q01-safe. Therefore a correct q012 implementation must be able to contract back to q01 and match the existing D q01 rows at `G0-add`.

If it cannot, the candidate implementation is wrong and B must not be interpreted.

Checked-in frontend config remains unchanged during this diagnostic.

---

# S0 — reproduce controls

Reproduce B and D generated-power controls.

If B is not in catastrophic regime or D is not in accepted passing regime, stop:

`logn13_q012_g0_f0_control_mismatch`.

---

# S1 — capture the exact actual G0 merge inputs

Use the existing source replay snapshots/checkpoints.

For real and imag capture, at minimum:

- `G0Merge.A` — the actual parent/addend input;
- `G0-rescale` / `G0Merge.BRaw` — the actual rescaled child before multiply;
- generated `T2` at the exact level/scale used;
- actual q01 `G0-multiply` output;
- actual q01 `G0-add` output;
- source expected semantic vectors already carried by the replay.

Record for each:

- Level;
- Scale;
- degree;
- NTT/Montgomery;
- q01 capacity;
- q012 oracle capacity when available.

Do not regenerate an algebraically equivalent but differently scheduled tree.

---

# S2 — D calibration: source-faithful q012 G0 merge

Before trusting B, implement the q012 G0 mirror on **q0=56**.

Starting from the actual Q01-unique D inputs:

1. exact-lift `A`, `BRaw`, and generated `T2` into q2;
2. perform q012 `Mul(BRaw,T2)` with exact Fast `Mul` semantics:
   - degree1×degree1;
   - produce `c0`, `c1`, `c2` separately;
   - same level and scale product;
3. compare product `c0/c1/c2` rows separately against an independent q012 oracle;
4. perform source-faithful q012 scale alignment matching `psRescaleGuardAddAligned` exactly;
5. compare aligned operand rows/metadata;
6. add `A + product` preserving degree-2 semantics exactly;
7. compare `G0-add` q012 rows against an independent oracle;
8. because D output is Q01-unique, contract q012→q01 and require exact q0/q1 row equality with the actual existing D `G0-add` output.

This D contraction is a hard calibration gate.

If D fails at any point, classify:

`logn13_q012_g0_merge_harness_mismatch`

and stop before drawing conclusions about B.

---

# S3 — component-wise first mismatch localization

For both D calibration and B, record the first mismatch among:

- `BRaw` lift;
- T2 lift;
- product `c0`;
- product `c1`;
- product `c2`;
- alignment integer factor;
- aligned A rows;
- aligned product rows;
- add `c0`;
- add `c1`;
- add `c2`;
- output metadata.

For each mismatch provide:

- expected row hash or compact equality flag;
- actual row hash/equality flag;
- first differing component/limb;
- semantic residual;
- scale metadata;
- intended physical capacity.

No full row dumps.

---

# S4 — B source-faithful q012 G0 merge

Only after D calibration passes, run the identical operation mirror under B q0=55.

Requirements:

- Q012 centered uniqueness throughout;
- exact q0/q1/q2 operation rows vs independent oracle;
- exact source scale/alignment semantics;
- G0-add semantic residual in normal numerical range, not O(1).

Do **not** require q01 contraction at B G0-add if the intended physical output is not Q01-unique. Keeping q2 authoritative is the whole point.

If the source-faithful B G0 merge itself still mismatches while D calibration passes, classify precisely whether the problem is:

- q012 multiply semantics;
- scale-alignment semantics;
- add semantics;
- representation/NTT/Montgomery transition;
- oracle materialization mismatch.

Do not label it generic capacity failure.

---

# S5 — true dependency-path q2 carry

If B G0-add is locally source-faithful, **do not run a serial G1/G2 accumulator**.

Leave unrelated G1/G2 subtrees on their accepted ordinary path.

Carry the authoritative q012 G0-add only along its actual dependency path:

`G0-add -> G3-add -> F0-root -> F0-final-rescale`

Specifically:

## At G3-add

- obtain the actual accepted `G3-multiply` sibling from the real replay;
- prove it is Q01-unique before lift;
- exact-lift it to q2;
- source-faithfully scale-align it with the q012 G0-add input;
- perform q012 add preserving degree semantics;
- prove q012 rows/metadata against oracle.

## At F0-root

Mirror the actual root handling exactly.

If the actual root is degree2, apply the Fast zero-secret relinearization semantic at the exact source point:

- truncate degree2→degree1 exactly as Fast `Relinearize` does;
- do not invent evaluation-key arithmetic.

Prove `c0/c1` rows and metadata.

## At F0-final-rescale

Perform the accepted guarded final Rescale semantics in q012 using the actual logical divisor and rounding rule.

Diagnostic big.Int centered reconstruction is allowed here only as an oracle/reference implementation.

Then prove the result is Q01-centered-unique and contract q012→q01.

Require exact contraction rows and metadata.

---

# S6 — full downstream replay

Inject only the correctly contracted F0 output into the existing accepted downstream path.

Do not rewrite DoubleAngle/S2C/finalization.

Record:

- polynomial output residual real/imag;
- EvalMod real/imag;
- post-S2C;
- public-like;
- margin to `1e-2`;
- metadata/contracts.

System pass iff:

`public_like <= 1e-2`.

Run D with the same q012 G0→F0 path as semantic control. D must remain equivalent to accepted D within source-backed tolerance.

---

# S7 — reset attribution semantics correction

Do not write a `minimal_causal_reset_set` unless a reset set actually collapses the target error.

If no reset collapses:

- record `minimal_causal_reset_set: []`;
- record `minimal_causal_reset_set_proven: false`.

A last-attempted reset is not a minimal causal set.

---

# Architecture interpretation

Do not assume q0=55 is mandatory.

The user explicitly permits q0=56 and parameter redesign. q0=55 remains in this task only because it is useful for determining whether q012 can preserve a higher-scale Fast computation.

If D q0=56 remains clean while B requires a complicated q012 path, report that as an engineering tradeoff rather than treating q0=56 as a forbidden workaround.

Likewise, do not assume q2 must be used only in tiny windows. If source-faithful evidence shows that a broader PS-wide q012 mode would be simpler and still much cheaper than full-RNS, say so in the recommendation.

However, this task itself does not change production parameters or run a scale sweep.

---

# Classification

Choose exactly one:

- `logn13_q012_g0_merge_harness_mismatch`
- `logn13_q012_g0_merge_source_faithful_but_downstream_blocked`
- `logn13_q055_plan92_g0_to_f0_q012_path_system_sufficient`
- `logn13_q055_plan92_g0_to_f0_q012_path_operation_mismatch`
- `logn13_q055_plan92_g0_to_f0_q012_path_numeric_insufficient`
- `logn13_q012_g0_f0_control_mismatch`

If local G0 fails before D calibration, use harness mismatch, not an architectural classification.

---

# Required compact artifact

Create:

`results/FIX-001-P3-DIAG-LOGN13-Q012-G0-F0-SOURCE-FAITHFUL-CLOSURE-summary.json`

Include:

- provenance;
- B/D control metrics;
- D calibration gate;
- exact G0 source snapshots metadata;
- component-wise q012 product proof (`c0/c1/c2`);
- exact alignment factor and fractional difference;
- component-wise G0-add proof;
- B G0 q012 capacity;
- actual dependency path proof;
- G3-add proof;
- F0 relinearization proof;
- F0 final q012 Rescale/contraction proof;
- downstream B/D metrics;
- reset-set semantics correction;
- classification;
- first remaining blocker;
- production readiness;
- architecture recommendation.

No full slots/coefficient/RNS arrays.

---

# Validation

Require:

- focused test matching `TestFIX001P3.*Q012.*G0.*F0.*Source.*Faithful`;
- `go test ./...` Primary;
- `git diff --check`;
- no NaN/Inf;
- Secondary exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean, unchanged;
- B/D principal controls use generated temporary-q012 powers;
- no OraclePowerMap as candidate producer;
- no serial G0→G1→G2→G3 recurrence;
- actual G0 snapshots used;
- D q0=56 contraction calibration passes before B interpretation;
- scale alignment matches `psRescaleGuardAddAligned` integer-ratio semantics exactly;
- degree2 `c0/c1/c2` proved separately;
- q2 carried only along the real dependency path when testing the bounded path candidate;
- root relinearization occurs at the actual source point;
- actual logical divisor used for final Rescale;
- q2 dropped only after Q01 uniqueness is proven;
- no false minimal reset set;
- checked-in q0 config unchanged;
- no Secondary production integration;
- no LogN16;
- no benchmark;
- no Gate4/5;
- no EXP003;
- Primary clean after ordinary ff push.

---

# Recommended next target rules

## If B passes with source-faithful G0→F0 q012 path

Do **not** automatically commit to q0=55 production.

Next target should compare architecture choices:

1. q0=56 + planScale92 accepted D0 path;
2. q0=55 + source-faithful q012 PS path;
3. optionally a small higher-scale sweep with sufficient q012 capacity.

Choose based on correctness, code complexity, and expected speed.

## If D calibration fails

Fix the q012 diagnostic primitive/semantics first. Do not infer anything about q0=55.

## If D calibration passes but B source-faithful G0 still fails

Localize that exact operation mismatch; do not widen the q2 window blindly.

## If local/path q012 works but remains too fragmented

Recommend evaluating a clean PS-wide q012 authoritative mode rather than accumulating many special-case windows.
