# FIX-001-P3-DIAG-LOGN13-B4-BABY-STEP-ATTRIBUTION

## Purpose

Localize the G0 parent-branch blocker to the smallest source-backed operation inside the B4 baby-step polynomial accumulation, and determine whether the residual is caused by Fast q0/q1 arithmetic, scalar accumulation semantics, or unavoidable sequential CKKS rounding at the current schedule.

Accepted evidence at Primary commit `33ff9bd8392685e31433a4956ea36309de4fd9de`:

- actual public-like error = `0.010683237260415292`;
- G0 merge classification = `logn13_g0_merge_parent_branch_blocker`;
- correcting only the G0 parent branch (`CA`) gives public-like `0.0032351181652570046` and passes;
- correcting only the product branch (`CP`) gives `0.01326258708226942` and fails;
- both branches canonical (`CAP`) gives `0.005258205152043761` and passes;
- smallest passing parent interpolation alpha = `0.07421875`;
- current Fast merge and stage-aligned full-RNS mirror are identical, so the G0 merge is not a q0/q1 storage mismatch;
- Standard native-scale Add semantics at G0 are materially worse and are not the fix;
- G0 parent actual-vs-canonical residual:
  - `1.800410931451779e-10` max-component;
- prior PS DAG evidence shows `B4-term-2` actual-vs-canonical residual is exactly the same value:
  - `1.800410931451779e-10`;
- the G0 parent branch has the same Level/Scale/logical state as the completed B4 baby-step output.

Therefore the next question is not "which PS block?". It is:

> Which operation inside B4 leaves the final `1.800410931e-10` residual that is disproportionately important to the downstream worst direction?

Important caution:

- B4 intermediate decoded residual magnitudes can be very large before the final cancellation, because partial sums may not be uniquely interpretable in q0/q1 at their intermediate scales.
- Do **not** rank B4 operations from intermediate decoded max norms alone.
- Use exact modulo-state/full-RNS comparisons and downstream checkpoint-reset counterfactuals.

Diagnostic only. No production arithmetic changes are authorized.

---

## Required provenance

Primary required base:

`33ff9bd8392685e31433a4956ea36309de4fd9de`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Keep fixed:

- oracle powers;
- q0/q1/q2 = 56/39/40;
- common plan scale `2^92`;
- existing PS split and coefficients;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DA local-q2 behavior;
- current G0 merge semantics;
- S2C and final restore behavior;
- genuine Standard widened-profile reference.

Do not:

- alter G0 merge semantics;
- retune global plan scale or guards;
- change q parameters;
- redesign generated powers;
- modify Secondary production code;
- production-integrate;
- run LogN16, benchmark, Gate4/5 or EXP003.

Primary-only diagnostic full-RNS mirrors, canonical checkpoint construction and operation resets are allowed.

---

# B0 — reproduce accepted controls and prove B4 == G0 parent

Reproduce:

- C00 actual public-like `~0.01068323726`;
- CA parent-corrected public-like `~0.00323511817`;
- CP product-corrected public-like `~0.01326258708`;
- parent smallest passing alpha `0.07421875` within prior tolerance;
- G0 parent actual-vs-canonical residual `~1.800410931e-10`.

Then prove the completed B4 baby-step state is the G0 parent input, not merely numerically similar.

Record for actual and canonical states:

- q0/q1 coefficient-state hash;
- Level;
- Scale;
- Degree;
- NTT/Montgomery metadata;
- decoded residual metric.

Require exact q0/q1 state equality between:

- completed actual B4 output and `A_F` at G0 input;
- completed canonical B4 checkpoint and `A_Q` used in the G0 attribution.

If not exact, stop:

`logn13_b4_parent_identity_mismatch`.

---

# B1 — source-backed B4 operation sequence

Derive the exact B4 operation order from `fastPolynomialWorkspace.evaluateBabyStep` and the active B4 polynomial.

Expected logical contributions must be source-confirmed, not assumed from this spec. Based on current diagnostics they are expected to include:

- constant initialization/add;
- `T6 * c6` via scalar `MulThenAdd`;
- `T4 * c4` via scalar `MulThenAdd`;
- `T2 * c2` via scalar `MulThenAdd`.

For every actual operation record compactly:

- operation ID;
- coefficient degree;
- coefficient value (decimal summary only, no large arrays);
- input power Level/Scale;
- accumulator Level/Scale before operation;
- accumulator Level/Scale after operation;
- scalar encoding scale selected by Fast evaluator;
- whether accumulator promotion occurred;
- promotion factor if any;
- whether the coefficient is treated as integer or non-integer;
- q0/q1 intended centered-capacity ratio where meaningful.

Also record the exact source function/path used for:

- scalar conversion;
- coefficient scale selection;
- scalar rounding to integer;
- accumulator promotion.

---

# B2 — final B4 execution-mode decomposition

Evaluate B4 from the same oracle power inputs under three stage-aligned modes.

### F — current Fast B4

Current q0/q1 Fast `Add`/`MulThenAdd` behavior.

### N — full-RNS mirror of Fast scalar schedule

Reproduce the **same Fast scalar schedule and integer rounding decisions** across full RNS, including:

- same accumulator promotions;
- same scalar encoding scales;
- same rounded scalar integers;
- same operation order;
- same target metadata.

This isolates q0/q1 storage/arithmetic from the schedule itself.

### S — genuine Standard scalar accumulation where source semantics permit

Using full-RNS stage-aligned copies of the same oracle powers and the same B4 coefficient order/initial accumulator metadata, execute genuine `ckks.Evaluator.Add` / `MulThenAdd` semantics.

If genuine Standard rejects a scale relation that current Fast supports, record the exact first unsupported operation and do not invent a workaround. The final S state may then be marked unavailable.

### Q — canonical B4 checkpoint

The deterministic canonical CKKS state for the exact completed B4 partial polynomial at the native B4 output metadata.

At the completed B4 boundary record:

- `F-N`;
- `N-Q`;
- `S-Q` if S is available;
- `F-Q`;
- q0/q1 row equality F vs N;
- final q01 uniqueness/capacity.

Primary interpretations:

- `F-N != 0`: Fast q0/q1 implementation/storage effect exists inside B4;
- `F-N == 0` and `N-Q != 0`: residual comes from the scalar schedule/rounding, not q0/q1 storage;
- S materially closer to Q than N: Fast scalar semantics differ meaningfully from Standard and are a candidate design target;
- S ~= N: current sequential CKKS schedule itself sets the residual floor.

---

# B3 — canonical checkpoint resets inside B4

For every source-backed B4 operation output, construct the deterministic canonical modulo state for the exact partial expression at that operation's **actual native metadata**.

Do not require intermediate q0/q1 decoded uniqueness. The reset is a modulo-state counterfactual; interpret success only by replaying to the safe completed B4/final system outputs.

Run these reset cases independently:

- after constant;
- after each scalar `MulThenAdd` in actual order.

For each case:

1. replace only that operation output with its canonical checkpoint state;
2. replay the remaining B4 operations using current Fast arithmetic;
3. replay all remaining PS, G0 merge, DA, S2C and finalization unchanged;
4. compare against genuine Standard reference/message.

Record:

- completed B4 `F-Q` after replay;
- final PS residual vs canonical;
- EvalMod real/imag;
- post-S2C max-component;
- public-like max-component;
- pass/fail at `1e-2`;
- all final metadata and capacity contracts.

Decision rule:

- identify the **latest** B4 operation reset that passes;
- compare it to the immediately previous reset;
- the interval between previous-failing and latest-passing is the smallest operation-level causal region.

The final B4 operation reset must reproduce CA within measured reset/materialization floor. If not, stop:

`logn13_b4_reset_endpoint_mismatch`.

---

# B4 — per-operation arithmetic substitution

Only for the smallest causal operation/interval identified in B3, compare arithmetic substitutions while leaving all other B4 operations Fast/current.

Test only source-grounded variants that preserve the same polynomial and logical metadata:

### V0 current Fast operation
Control.

### V1 full-RNS same-Fast-schedule operation
Execute that operation with the same Fast scalar integer/scale choice but full-RNS storage, then contract/re-enter the Fast diagnostic path only when the next state is mathematically safe/consistent.

### V2 genuine Standard operation
If Standard accepts the native operand scales, execute that one operation using genuine Standard scalar semantics, then return to the fixed diagnostic suffix at matched logical metadata.

Do not invent a Standard-compatible scale conversion if Standard rejects the native case.

For each valid variant record completed B4 residual and full downstream public-like result.

This determines whether a production fix should target:

- q0/q1 implementation;
- Fast scalar semantics;
- or precision/schedule while keeping the operation semantics.

---

# B5 — scalar-rounding error ledger

For B4 coefficients only, build a compact scalar quantization ledger from the **actual executed schedule**.

For each coefficient operation record:

- exact/high-precision coefficient value;
- scalar encoding scale used;
- rounded integer real/imag components used by `scalarNTT` (compact decimal integers only);
- effective represented coefficient after dividing by that scalar scale;
- scalar coefficient quantization error;
- sign of the error.

Because accumulator promotions can rescale earlier terms, propagate each coefficient's scalar quantization error through the exact subsequent scalar-scale transitions to the completed B4 mathematical value.

Compute an arbitrary/high-precision predicted completed-B4 residual vector from the scalar rounding ledger and compare it with `N-Q`.

Require vector agreement <= `1e-12` if numerically meaningful, or report a rigorously justified floor.

If it does not close, record the unaccounted component rather than forcing attribution.

Purpose:

> determine whether the `1.800410931e-10` parent residual is explainable by scalar coefficient rounding alone.

---

# B6 — one-coefficient correction counterfactuals

If B5 closes sufficiently, test one coefficient at a time.

For each B4 coefficient contribution independently:

- replace only that operation's scalar-rounded contribution with the canonical/high-precision checkpoint-equivalent contribution at the same logical operation boundary;
- preserve all other current Fast operations;
- replay B4 suffix and full downstream path.

Record public-like result.

If one coefficient correction alone passes `1e-2`, identify it as the minimal semantic target.

If no single coefficient passes but an operation-output reset from B3 passes, keep the blocker at operation/schedule scope.

No production fix is authorized.

---

# B7 — decision classification

Choose exactly one primary classification.

### Fast q0/q1 arithmetic inside B4 is causal

`logn13_b4_fast_storage_arithmetic_blocker`

Use only if F vs N is materially nonzero and a full-RNS substitution improves/passes downstream.

### One scalar operation / coefficient is sufficient

`logn13_b4_single_scalar_operation_blocker`

Record operation ID, coefficient degree and downstream result.

### Fast scalar semantics differ from Standard and are sufficient

`logn13_b4_fast_scalar_semantics_blocker`

Use only if a valid Standard-operation substitution is materially closer to Q and passes downstream.

### B4 sequential scalar rounding is distributed

`logn13_b4_distributed_scalar_rounding_blocker`

Use if F==N, no single coefficient/operation substitution passes, but the completed B4 canonical reset passes and the scalar-rounding ledger explains the residual.

### Residual is not explained by B4 scalar accumulation

`logn13_b4_attribution_mismatch`

Use if B4 identity is valid but neither execution-mode nor scalar-rounding decomposition explains the parent residual.

No production change is authorized in this task.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-B4-BABY-STEP-ATTRIBUTION-summary.json`

Include only:

- provenance;
- B0 controls and B4/G0-parent state identity hashes;
- compact B4 operation schedule;
- F/N/S/Q final B4 comparison;
- operation-reset downstream table;
- smallest causal interval;
- operation substitution table for that interval only;
- compact scalar-rounding ledger;
- scalar-ledger closure metric;
- one-coefficient correction table if reached;
- classification;
- first remaining blocker;
- recommended next design target;
- validation flags.

Do not serialize slot vectors, coefficient arrays, powers, full RNS rows, matrices or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*B4.*Baby.*Step.*Attribution` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no G0/DA/S2C/q/generated-power redesign;
- no production integration;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.