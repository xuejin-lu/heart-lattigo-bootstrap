# FIX-001-P3-DESIGN-LOGN13-GENERATED-POWER-NATIVE-Q01-MULTIPLY-FIRST-FEASIBILITY

## Goal

Prove or reject the **minimal** generated-power fix before any production integration:

> Use the real Fast q0/q1 arithmetic directly for generated-power `multiply -> Chebyshev double -> recurrence correction -> one Rescale`, with no balanced pre-Rescale operands and no temporary q2.

The previous design task established that the multiply-first schedule is system-sufficient when evaluated through a full-RNS diagnostic mirror and then contracted back to Fast q0/q1. It also established that every intended q01 multiply-first checkpoint is centered-unique under the accepted q0=56 profile.

Therefore the next question is not whether multiply-first is mathematically useful. It is:

> Does native Fast q0/q1 arithmetic reproduce the already-passing full-RNS multiply-first candidate exactly enough that q2 is unnecessary?

This task is design-feasibility only. Do not modify Secondary production code.

---

# Required provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required base/ancestor:

`904f0460e36b69add5491f7c7006c72153970789`

Start from clean `main`, fetch + ff-only pull, then re-read fresh `AGENTS.md`, `CURRENT_TASK.md`, and this spec.

## Secondary

Repository: `xuejin-lu/lattigo`

Required exact commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Branch: `fast-ckks`, clean.

**No Secondary production changes are authorized.**

Primary-only diagnostic helpers are allowed.

---

# Accepted evidence from the previous task

Primary result:

`904f0460e36b69add5491f7c7006c72153970789`

Classification:

`logn13_generated_power_local_q2_multiply_first_system_sufficient`

Important correction to interpretation:

- the successful multiply-first candidate was implemented by materializing a full-RNS diagnostic state, executing the multiply-first arithmetic there, then contracting to q0/q1;
- q012 was used as capacity/oracle evidence;
- the previous task did **not** prove a native 3-limb Fast implementation;
- it also did **not** execute a native q0/q1 multiply-first candidate.

Accepted controls:

- D0 oracle-power public-like: `0.005554220603853743`;
- old all-generated balanced schedule: `0.22247498019233428`;
- full-RNS diagnostic all-generated multiply-first candidate: `0.005554220603853743`;
- T4 local residual improved approximately `5.73e-8 -> 2.77e-8`;
- T6 local residual improved approximately `6.29e-8 -> 4.17e-8`.

Accepted capacity evidence:

- q0/q1/q2 profile: 56/39/40;
- Q01 domain bit length: 95;
- Q012 domain bit length: 135;
- **all q01 multiply-first checkpoints centered-unique**;
- maximum observed q01 capacity ratio: approximately `0.8774491930085215` at the T3 recurrence-corrected checkpoint;
- outside_count = 0 there;
- minimum required Q01 bits = 95;
- q012 max ratio ~`0.0002439022`.

This means q01 is tight but sufficient for the accepted deterministic LogN13 workload. Do not widen q0/q1 in this task.

---

# Fixed D0 stack

Every system replay in this task must preserve:

- LogN13 only;
- effective q0/q1/q2 = 56/39/40;
- plan scale `2^92`;
- same polynomial and PS split;
- same final-parent T2 one-bit scalar guard;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all accepted DoubleAngle local-q2 rounds;
- same G0 merge semantics;
- same S2C;
- same final restoration;
- same widened Standard reference;
- same deterministic workload/input.

Do not change:

- q parameters;
- coefficients;
- power dependency graph;
- plan scale;
- PS/G0/DA/S2C arithmetic;
- frontend/workload semantics.

Do not run LogN16, benchmark, Gate4/5, or EXP003.

---

# Q0 — reproduce both accepted controls

Reproduce:

1. D0 oracle control public-like:
   `0.005554220603853743`;
2. previous full-RNS all-generated multiply-first candidate public-like:
   `0.005554220603853743`;
3. old balanced G_ALL:
   `0.22247498019233428`;
4. previous q01 max capacity ratio near `0.8774491930085215`;
5. dependency graph:
   - T2=(1,1)
   - T3=(1,2)
   - T4=(2,2)
   - T6=(3,3)
   - T8=(4,4)
   - T16=(8,8).

If controls materially mismatch, stop:

`logn13_generated_power_native_q01_precondition_mismatch`.

---

# Q1 — native q01 multiply-first primitive in Primary diagnostic code

Implement a Primary-only diagnostic helper using the actual Secondary Fast evaluator operations and only authoritative q0/q1 residues.

For each generated power node:

1. start from the actual Fast q0/q1 parent states at their native metadata;
2. use the actual dependency split from `commonpolynomial.SplitDegree`;
3. perform native Fast `MulRelin(left,right,out)` at the common native level;
4. for Chebyshev basis, perform native Fast doubling `Add(out,out,out)`;
5. apply the recurrence correction **before the single Rescale**, matching the accepted full-RNS multiply-first semantic schedule:
   - if `c=0`, subtract constant `1` with the correct high-scale scalar semantics;
   - otherwise subtract the difference power using source-equivalent scale alignment, without metadata-only semantic corruption;
6. perform exactly **one** native Fast `Rescale` using the actual current-level divisor `Q[L]`;
7. require output Level/Scale/degree/NTT/Montgomery metadata to match the accepted full-RNS multiply-first candidate for that power.

Do not use:

- full-RNS arithmetic to produce the native candidate;
- q2 arithmetic;
- decoded-slot replacement;
- canonical/oracle output injection;
- balanced integer factors;
- pre-Rescale of either parent.

The full-RNS candidate may be used only as an independent oracle/mirror.

---

# Q2 — pre-execution q01 capacity gate

Before each native high-scale operation, independently compute the intended centered physical coefficients from the full-RNS/big.Int oracle.

Check at minimum:

- raw product;
- after Chebyshev doubling;
- after recurrence scale alignment, if any;
- after recurrence correction;
- immediately before Fast Rescale.

For every checkpoint record:

- max `|x|`;
- Q01/2;
- ratio;
- outside_count;
- minimum required Q01 bits;
- remaining fractional headroom `1-ratio`.

### Hard rule

If any intended checkpoint has `|x| >= Q01/2`, do **not** execute the corresponding native q01 operation. Stop with:

`logn13_generated_power_native_q01_capacity_regression`.

Expected control from the previous task is max ratio about `0.8774491930`, outside_count 0.

Do not reinterpret a wrapped centered representative as safe.

---

# Q3 — row-by-row native-vs-full mirror proof

For each generated power T2/T3/T4/T6/T8/T16, compare native q01 Fast arithmetic against the accepted full-RNS multiply-first mirror at matching source checkpoints.

At each checkpoint compare q0/q1 coefficient rows after representation normalization as needed:

1. parent inputs;
2. raw product;
3. doubled state;
4. recurrence-aligned state;
5. recurrence-corrected state;
6. post-Rescale completed power.

Record:

- q0 row equality;
- q1 row equality;
- first mismatching coefficient index if any;
- decoded max component difference;
- metadata equality.

Primary desired contract:

- native q01 rows == full-RNS mirror reduced to q0/q1 at every capacity-safe checkpoint;
- post-Rescale completed power rows match exactly.

If rows first diverge while the intended q01 state is centered-unique, classify the exact source operation as an implementation/semantic mismatch rather than a capacity failure.

---

# Q4 — Rescale proof

The most important primitive boundary is the one native Fast Rescale after high-scale recurrence correction.

For every generated power:

- independently recover the exact centered q01 integer immediately before Rescale;
- perform big.Int symmetric nearest rounded division by the actual divisor `Q[L]`;
- compare the expected q0/q1 residues against native Fast Rescale output;
- compare native output against the full-RNS multiply-first mirror after its Rescale;
- require output Scale equality and Level decrement by exactly one.

Record:

- rounded quotient row agreement;
- maximum integer rounding discrepancy (must be 0 for exact row agreement);
- q01 centered uniqueness before Rescale;
- q01 centered uniqueness after Rescale.

If Fast Rescale disagrees with the independent q01 quotient despite a unique preimage, stop:

`logn13_generated_power_native_q01_rescale_mismatch`.

---

# Q5 — local power accuracy

At identical planned output metadata compare for each T2/T3/T4/T6/T8/T16:

- old balanced Fast generated power vs canonical;
- previous full-RNS multiply-first candidate vs canonical;
- native q01 multiply-first candidate vs canonical.

Record max component, mean, worst index, and improvement factors.

Expected qualitative target:

- native q01 result should equal or numerically reproduce the full-RNS multiply-first candidate;
- T4/T6 should retain the previously observed improvement;
- descendants T8/T16 should no longer carry the old balanced-pre-Rescale amplification.

Do not impose a local `1e-10` perfection threshold. System public-like threshold remains authoritative.

---

# Q6 — all-generated native q01 system replay

Regenerate the complete required power map using **only native q01 multiply-first** for generated nodes:

T2, T3, T4, T6, T8, T16.

No oracle power replacement is allowed except T1/input as defined by the normal power basis.

Replay the entire fixed D0 downstream:

- PS polynomial;
- accepted T2 scalar guard;
- G0/F0 local-q2 stack;
- all accepted DA local-q2 rounds;
- S2C;
- finalization.

Record:

- final power residuals;
- polynomial residual;
- EvalMod real/imag;
- post-S2C;
- public-like;
- margin to `1e-2`;
- difference to previous full-RNS multiply-first public-like;
- all metadata/capacity contracts.

Primary success condition:

`public_like <= 1e-2`.

Preferred equivalence condition:

`abs(native_q01_public_like - full_rns_multiply_first_public_like) <= 1e-10`.

If the system passes but differs beyond that tolerance, preserve the pass result but classify a native/full-RNS numeric mismatch and localize the first divergent checkpoint before recommending production integration.

---

# Q7 — relation to q0/q1 architecture study

Record compact parameter evidence:

- actual q0 bits;
- actual q1 bits;
- Q01 modulus bit length;
- maximum native q01 multiply-first ratio;
- headroom percentage;
- minimum required Q01 bits;
- checkpoint that sets the minimum;
- whether any native/full-RNS mismatch occurs despite adequate capacity.

Use precise interpretation:

- if native q01 succeeds, current 56/39 is **sufficient but tight** for this LogN13 workload;
- do not claim it is generally optimal;
- the independent future “Fast q0/q1 parameter design” study remains valid because current max ratio ~0.877 leaves limited margin;
- do not change q values here.

---

# Classification

Choose exactly one primary classification:

- `logn13_generated_power_native_q01_multiply_first_system_sufficient`
- `logn13_generated_power_native_q01_capacity_regression`
- `logn13_generated_power_native_q01_operation_mismatch`
- `logn13_generated_power_native_q01_rescale_mismatch`
- `logn13_generated_power_native_q01_numeric_insufficient`
- `logn13_generated_power_native_q01_precondition_mismatch`

If system sufficient, recommended next target:

`production integration of native q01 generated-power multiply-first schedule`.

If q01 fails for a non-capacity arithmetic reason, recommended next target must identify the first exact divergent primitive.

If q01 fails due capacity, only then recommend a true q012/3-limb Fast primitive design.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-GENERATED-POWER-NATIVE-Q01-MULTIPLY-FIRST-FEASIBILITY-summary.json`

Include:

- provenance;
- Q0 controls;
- dependency graph;
- q01 capacity table;
- native-vs-full checkpoint row table;
- Rescale quotient/oracle table;
- local power comparison table;
- all-generated native system metrics;
- q0/q1 headroom evidence;
- classification;
- first remaining blocker;
- recommended next target;
- validation flags.

Do not serialize full slot vectors, full coefficient arrays, or full RNS rows.

---

# Validation

Require:

- focused test matching `TestFIX001P3.*Generated.*Power.*Native.*Q01.*Multiply.*First`;
- `go test ./...`;
- `git diff --check`;
- artifact has no NaN/Inf;
- D0 exact control reproduced;
- previous full-RNS multiply-first control reproduced;
- q01 capacity checked before native high-scale operations;
- no unsafe q01 operation executed;
- all-generated candidate uses genuine native Fast q0/q1 arithmetic, not full-RNS output materialization;
- Secondary remains exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean and unmodified;
- no q parameter change;
- no temporary q2 arithmetic candidate in this task;
- no production integration;
- no LogN16, benchmark, Gate4/5, EXP003;
- final Primary worktree clean after ordinary ff push.
