# FIX-001-P3-DIAG-LOGN13-Q055-PLAN92-PS-DEPENDENCY-WINDOW-CLOSURE

## Goal

Correct the remaining diagnostic-harness defects from the previous PS first-divergence task, then determine the **minimal dependency-correct q012 authority window or window set** needed for LogN13 `q0=55 / planScale=2^92` to preserve the intended Paterson–Stockmeyer computation.

The final question is:

> Can the checked-in frontend `q0=55` remain unchanged if temporary q2 authority is retained across the actual PS causal window(s), or is there still a source-backed blocker after those windows are repaired?

This remains a Primary-only diagnostic/design task.

**Do not modify Secondary production code.**

---

# Required provenance

## Primary

Repository: `xuejin-lu/heart-lattigo-bootstrap`

Required ancestor:

`45db7fa141158f1c672ae0cbb86ebddb88939fb7`

Start clean on `main`, fetch + ff-only pull, re-read fresh:

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

# Accepted evidence from the previous task

The previous artifact established useful evidence, but its final `multiple_independent_blockers` classification is **not yet considered proven**.

Accepted observations:

- B control (`q0=55`, plan92) remains catastrophic:
  - EvalMod real ≈ `102.29911082927852`
  - EvalMod imag ≈ `102.2990271839068`
  - post-S2C ≈ `114.46562670803426`
  - public-like ≈ `3662.9000547146406`
- D control (`q0=56`, plan92) passes:
  - EvalMod real ≈ `6.162959141897749e-5`
  - EvalMod imag ≈ `4.7461415340529166e-5`
  - public-like `0.005554220603853743`
- B4 T2 q012 guard is arithmetically valid but is not sufficient by itself.
- A q0=55 PS giant-step region contains physical values not centered-unique in Q01 while remaining tiny in Q012.
- Example observed at `G0-multiply`:
  - B intended/Q01 ≈ `3.561156696`
  - B intended/Q012 ≈ `6.48e-12`.

These observations justify investigating a wider q012 PS window.

---

# Mandatory corrections to the previous diagnostic harness

The new runner must explicitly correct all of the following before drawing a production conclusion.

## C1 — use actual accepted generated powers

The previous `makeCase` used:

`context.Branch.Base.OraclePowerMap`

for the principal PS trace.

That is not the fixed accepted stack required by the spec.

For both real and imag, generate and use the accepted temporary-q012 multiply-first powers through the same source-backed path used by the successful generated-power feasibility work:

- obtain E2/Z from the accepted setup;
- call the genuine `nativeQ012Generate`/equivalent accepted generated-power diagnostic producer;
- use generated T2/T3/T4/T6/T8/T16 in the PS replay;
- keep their existing q012 Rescale/contraction proofs.

Oracle powers may exist only as comparison/reset oracles.

The principal B/D controls and every candidate system replay must use **generated powers**.

## C2 — do not infer multiple blockers from the unrelated B4 T2 candidate

The previous classification path effectively said:

- first causal window looks capacity-related;
- q012 is sufficient there;
- old B4 T2 candidate does not fix system;
- therefore multiple blockers.

That does not prove multiple independent causal windows because B4 T2 is not the newly identified `G0` window.

Remove this inference.

## C3 — dependency graph, not chronological adjacency

The previous `first_centered_sensitive_boundary` was selected as the first centered-sensitive operation appearing later in the flat operation list.

That is insufficient.

Build an explicit dependency graph for the PS replay.

For every stable operation ID record:

- input operation IDs / source nodes;
- output operation ID;
- block/merge identity;
- operation kind;
- whether it is centered-sensitive;
- whether q01 physical uniqueness holds;
- whether q012 uniqueness holds.

A centered-sensitive boundary is causal only if it consumes, directly or transitively, the unsafe state.

Do not call `G1-rescale` causal merely because it appears later in execution order.

## C4 — bounded cumulative reset attribution

The previous task performed only one reset attempt at `G0-add` for both branches.

The prior spec required cumulative attribution when one reset is insufficient.

This task must perform dependency-aware cumulative resets until one of these happens:

1. catastrophic EvalMod error collapses to normal/accepted regime;
2. all identified B-only causal frontiers are reset without collapse;
3. q012 width becomes insufficient;
4. a non-capacity semantic mismatch is found.

Record the minimal reset set that first collapses the catastrophic error, if one exists.

## C5 — populate the last proven Q01-unique ancestor

The previous artifact left `last_q01_unique_state` blank.

For each causal window, identify the last dependency ancestor where all authoritative physical values are proven centered-unique in Q01. This is where q2 authority may safely begin by exact lift.

---

# Fixed system configuration

Principal B case:

- LogN13;
- checked-in frontend config unchanged;
- `q0=55`;
- `planScale=2^92`;
- accepted C2S compression;
- **generated** temporary-q012 powers;
- accepted G0 local-q2 guard=2;
- accepted F0 local-q2 guard=3;
- accepted all DoubleAngle local-q2 rounds;
- accepted final-parent B4 T2 temporary-q012 guard;
- same polynomial coefficients and PS decomposition;
- same deterministic workload/reference;
- same S2C/finalization.

D control is identical except in-memory diagnostic `q0=56`.

No frontend file edits.

---

# P0 — reproduce B and D using generated powers

Re-run B and D with the corrected generated-power stack.

Required B regime:

- EvalMod real around `102.2991`;
- EvalMod imag around `102.2990`;
- public-like around `3662.9001`.

Required D regime:

- public-like `0.005554220603853743` within existing tolerance;
- accepted EvalMod regime around `1e-5`–`1e-4`.

If these controls do not reproduce:

`logn13_q055_plan92_ps_dependency_control_mismatch`.

---

# P1 — build the PS dependency graph

Instrument the actual PS replay and construct a compact DAG/tree of dependencies.

Each node must have:

- stable ID;
- branch (`real`/`imag`);
- operation type;
- input IDs;
- output ID;
- block/merge number;
- Level/Scale/degree/NTT/Montgomery;
- centered-sensitive flag;
- q01 physical capacity ratio;
- q012 physical capacity ratio when available;
- q01 centered uniqueness;
- q012 centered uniqueness;
- semantic residual against canonical oracle.

At minimum cover:

- baby-step accumulations;
- each giant-step input;
- Rescale;
- multiply / MulRelin / relinearize as actually executed;
- aligned add/sub merge;
- final PS rescale/output.

The graph must reflect dataflow, not only execution order.

---

# P2 — identify all B-only causal capacity frontiers

Compare B and D by stable operation ID and dependency role.

A **B-only causal capacity frontier** is a node/window where:

1. B intended physical value is not centered-unique in Q01;
2. Q012 remains centered-unique;
3. D remains semantically/oracle-correct on the corresponding dependency path, or the B/D difference becomes causal at a downstream consumer;
4. that unsafe state reaches a centered-sensitive consumer before it is safely contracted/reset.

Important:

- an intermediate multiply may be modular and may wrap harmlessly for a time;
- do not declare the multiply itself causal unless a dependency-aware consumer converts the wrong centered representative into a semantic error;
- degree-2 intermediates under zero-secret semantics must be interpreted according to the actual Fast arithmetic contract, not generic full-RNS assumptions.

For every frontier report:

- first unsafe node;
- last Q01-unique ancestor;
- first dependency-descendant centered-sensitive consumer;
- safe contraction point, if any;
- maximum intended/Q01 ratio;
- maximum intended/Q012 ratio;
- whether q012 is sufficient over the whole proposed window.

---

# P3 — dependency-aware cumulative reset attribution

Use canonical/oracle ciphertexts only for attribution.

Perform reset experiments in causal/dependency order.

Start with the earliest B-only causal frontier on real and imag.

If the system remains catastrophic, cumulatively add the next independent frontier/window while preserving earlier resets.

Continue until the minimal set that collapses the catastrophic error is found or all frontiers are exhausted.

Record after every cumulative reset set:

- PS final real/imag residual;
- EvalMod real/imag;
- post-S2C;
- public-like;
- contracts/metadata.

Define collapse as:

- EvalMod real and imag each `<= 1e-2`, and preferably returned to accepted `~1e-4` regime;
- public behavior meaningfully returns toward the accepted path.

Do not call blockers independent merely because a single reset did not solve the system.

---

# P4 — derive minimal q012 authority window candidates

From the dependency graph and reset attribution derive the smallest real/imag q012 authority window(s).

For each proposed window:

1. start at the last proven Q01-unique ancestor;
2. exact-lift authoritative q0/q1 state into q2;
3. keep q0/q1/q2 authoritative through every dependent modular operation that may carry a wrapped physical value;
4. execute centered-sensitive Rescale/contraction using q012 physical interpretation;
5. contract back to q01 only when the result is again proven Q01-unique;
6. keep q3+ dormant.

A window may contain:

- multiply;
- add/sub merge;
- relinearization as required by actual Fast semantics;
- scale alignment;
- one or more Rescales.

Do not contract q2 between two operations if the first output is not Q01-unique and the second depends on it.

If two causal paths are disjoint, represent them as two explicit windows rather than inventing a broad global-q2 mode.

---

# P5 — implement bounded q012 candidate expansion

Implement Primary-only genuine q012 diagnostic candidates.

Start with the minimal window from P4.

If it is correct locally but the system remains failing, expand only along the next proven causal dependency edge/frontier.

Continue in bounded steps until:

- system passes;
- all identified frontiers are covered;
- q012 becomes insufficient;
- a non-capacity semantic mismatch is exposed.

For q012 modular arithmetic:

- q0/q1/q2 are authoritative;
- q3+ dormant;
- no Standard full-RNS producer;
- no stale high-limb reads;
- exact same scalar coefficients and scale semantics;
- exact same logical divisors.

For centered q012 Rescale/division in this **diagnostic** task, existing exact q012 reconstruction/oracle helpers may be reused if necessary, but artifact must clearly mark any per-coefficient big.Int path as diagnostic-only and forbidden for future production hot path.

Full-RNS is oracle only, never candidate producer.

---

# P6 — local proof for each candidate window

At every candidate boundary prove:

- exact lift from a Q01-unique parent;
- q0/q1/q2 row agreement against independent oracle;
- metadata equality;
- intended Q012 centered uniqueness throughout;
- exact centered-sensitive rounded quotient where applicable;
- post-window Q01 centered uniqueness before q2 drop;
- exact q0/q1 contraction rows.

For degree-2 intermediates, include the actual zero-secret/Fast relevance of each component and prove candidate semantics against the source-backed oracle.

---

# P7 — full system replay

For every candidate window set, replay the full fixed B stack using:

- generated q012 powers;
- candidate PS q012 window(s);
- accepted T2 q012 guard;
- accepted downstream local-q2 DA stack.

Record:

- PS final residual real/imag;
- EvalMod real/imag;
- post-S2C;
- public-like;
- margin to `1e-2`;
- final metadata/contracts.

System pass iff:

`public_like <= 1e-2`.

Also run D with the same window logic as a semantic control. D must remain equivalent to the accepted D path within source-backed tolerance.

---

# P8 — production-readiness decision

## If one q012 window makes B pass

Set:

`production_ready_fixed_q055_plan92_single_ps_q012_window`.

## If multiple disjoint q012 windows are required and together make B pass

Set:

`production_ready_fixed_q055_plan92_multiple_ps_q012_windows`.

## If all identified capacity frontiers are repaired but B still fails

Report the first remaining non-capacity semantic divergence. Do not claim q0=56 is required yet.

## If Q012 is insufficient inside a required window

Stop at limb-width architecture decision; do not silently enable q3 or change q0.

---

# Classification

Choose exactly one:

- `logn13_q055_plan92_ps_single_q012_window_system_sufficient`
- `logn13_q055_plan92_ps_multiple_q012_windows_system_sufficient`
- `logn13_q055_plan92_ps_multiple_capacity_frontiers_not_yet_sufficient`
- `logn13_q055_plan92_ps_first_remaining_noncapacity_mismatch`
- `logn13_q055_plan92_ps_q012_window_width_insufficient`
- `logn13_q055_plan92_ps_dependency_control_mismatch`
- `logn13_q055_plan92_ps_dependency_harness_mismatch`

Do not emit `multiple_*` merely because a single unrelated candidate failed.

---

# Required compact artifact

Create:

`results/FIX-001-P3-DIAG-LOGN13-Q055-PLAN92-PS-DEPENDENCY-WINDOW-CLOSURE-summary.json`

Include:

- provenance;
- explicit correction flags for C1–C5;
- B/D generated-power controls;
- compact dependency graph;
- B-only capacity frontier table;
- last-Q01-unique ancestor for each frontier;
- first causal centered-sensitive consumer per frontier;
- cumulative reset table;
- minimal causal reset set;
- candidate q012 window definitions;
- local row/rounding/contraction proofs;
- candidate expansion history;
- B/D system metrics for each window set;
- diagnostic-only big.Int usage flag;
- classification;
- production readiness;
- first remaining blocker;
- recommended next target;
- validation flags.

Do not serialize full slot arrays, coefficient vectors, or full RNS rows.

---

# Validation

Require:

- focused test matching `TestFIX001P3.*Q055.*Plan92.*PS.*Dependency.*Window.*Closure`;
- `go test ./...` in Primary;
- `git diff --check`;
- no NaN/Inf in artifact;
- checked-in LogN13 q0 remains 55;
- Secondary exact `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`, clean, unmodified;
- principal B/D traces use generated q012 powers, not `OraclePowerMap`;
- oracle powers used only for comparison/reset attribution;
- dependency edges explicitly recorded;
- causal centered-sensitive consumer selected by dependency, not list order;
- `last_q01_unique_state` populated for every candidate window;
- cumulative reset attribution executed when a single reset is insufficient;
- no classification based on unrelated B4 T2 candidate failure;
- q012 candidate spans complete unsafe dependency path until Q01-safe contraction;
- post-window Q01 uniqueness proven before dropping q2;
- D semantic control preserved;
- no frontend q change;
- no Secondary production integration;
- no LogN16;
- no benchmark;
- no Gate4/5;
- no EXP003;
- Primary clean after ordinary ff push.

---

# Recommended next target

If B passes with bounded q012 PS window(s):

`production integration of the complete accepted LogN13 D0 stack under fixed q0=55 / internal planScale92, including generated-power q012 arithmetic, the proven PS q012 window set, T2 q012 guard, and accepted downstream local-q2 guards`

If B remains blocked after complete capacity-window closure:

`localize the first remaining non-capacity q0=55/plan92 semantic divergence before any frontend parameter change`

If q012 width is insufficient:

`architecture decision on temporary authoritative limb width for the proven PS window; do not change frontend q0 implicitly`
