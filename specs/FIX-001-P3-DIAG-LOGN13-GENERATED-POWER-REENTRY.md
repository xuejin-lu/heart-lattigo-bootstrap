# FIX-001-P3-DIAG-LOGN13-GENERATED-POWER-REENTRY

## Goal

Resume the deliberately deferred generated-power correctness work **after** the oracle-power PS/B4 path has been closed.

Do not treat the failed production integration as evidence that the T2 scalar guard is wrong. The previous production comparison was invalid because the successful feasibility candidate and ordinary production did not share the same prerequisites.

Accepted successful oracle-power candidate at Primary commit `3bc71929c4b1a9196088dbb20dec6b50f105e5bc`:

- oracle powers;
- effective q0/q1/q2 = 56/39/40;
- common plan scale = `2^92`;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all three DoubleAngle rounds use the accepted local-q2 behavior;
- B4 final-parent T2 one-bit scalar guard;
- unchanged G0 merge semantics;
- unchanged validated S2C/final restore;
- public-like = `0.005554220603853743`;
- all capacity/CRT/rounded-contraction/row/metadata contracts pass.

Failed ordinary production at Primary commit `cb8c1c10fc678a2c81f8f732ffdfeff5fdc050aa` used a different stack:

- generated powers;
- effective q0 = 55 bits;
- production normalized plan scale = `2^91`;
- no production G0/F0 local-q2 integration;
- no production DA local-q2 integration;
- T2 guard enabled anyway;
- public-like ~= `0.2462732498`.

Therefore the next correctness question is:

> With every already-accepted non-power prerequisite held fixed, what generated-power state first prevents the system from retaining the oracle-power passing result?

This task is diagnostic. Do not attempt final production integration.

---

# Required provenance

## Primary

Required base/ancestor:

`cb8c1c10fc678a2c81f8f732ffdfeff5fdc050aa`

Start clean `main`, fetch and ff-only pull, then re-read fresh `AGENTS.md`, `CURRENT_TASK.md`, and this spec.

## Secondary

Required starting commit:

`d074ce8b1703a73b06e1f2ee2fe9912056f54f02`

Branch `fast-ckks`, clean.

---

# S0 — contain the unsafe production activation

The Secondary guard primitive/API itself is retained. The problem is only that the current ordinary normalized LogN13 production branch invokes the guarded evaluator under the legacy q0=55 / planScale=2^91 profile, while the guard was proven only in the q0=56 / planScale=2^92 diagnostic profile.

Make the smallest Secondary safety correction:

- in `circuits/ckks/mod1/fast.go`, restore the current q0=55 normalized LogN13 production branch to ordinary `EvaluateWithPlanScale(...)` rather than `EvaluateWithPlanScaleFinalParentOneBitScalarGuard(...)`;
- do not delete the one-bit guard primitive;
- do not delete the guarded polynomial API;
- do not redesign the normalized Mod1 path;
- do not change q parameters or plan scale in Secondary in this task.

Required regression:

- q0=55 ordinary production no longer reports a guarded selection;
- numeric behavior returns to the pre-guard production baseline within deterministic tolerance;
- existing guard primitive/polynomial focused tests still pass when invoked explicitly.

Run Secondary focused tests, `go test ./...`, and `git diff --check`.

Commit and ordinary ff push `fast-ckks` under the standing authorization. Record the resulting Secondary SHA.

This containment is not a correctness solution; it only prevents an unproven guard from remaining active in legacy production.

---

# D0 — exact oracle-power passing control

Back in Primary, construct the accepted final diagnostic stack, not ordinary production:

- same workload/input as the accepted B4 T2 feasibility run;
- effective q0/q1/q2 = 56/39/40;
- plan scale `2^92`;
- G0 local-q2 guard = 2;
- F0 local-q2 guard = 3;
- all three accepted DA local-q2 rounds;
- T2 one-bit scalar guard using the production guard primitive/API from Secondary;
- oracle powers;
- validated S2C/finalization.

Require reproduction of:

- guarded B4 residual ~= `6.3683e-11`;
- public-like ~= `0.005554220603853743`;
- all accepted arithmetic contracts.

If this control does not reproduce, stop:

`logn13_generated_power_reentry_oracle_control_mismatch`.

Do not continue into generated-power attribution with a broken oracle control.

---

# D1 — generated-power all-on case under the SAME stack

Using the exact D0 parameters and downstream arithmetic, replace only oracle powers with the current real Fast generated powers produced by the Secondary polynomial power-generation code.

Everything else remains identical to D0.

Record:

- exact generated power keys required by the actual PS plan (derive from runtime/source; do not assume a fixed list);
- each power native Level/Scale/degree/NTT/Montgomery metadata;
- each generated power vs deterministic canonical/oracle CKKS state at the same native metadata;
- q01 centered capacity at each safe checkpoint;
- final polynomial residual vs canonical;
- EvalMod real/imag;
- post-S2C;
- public-like.

This is the generated-power all-on control `G_ALL`.

Do not compare it to legacy q0=55 production; compare it only to D0.

---

# D2 — hybrid power attribution

Let `K` be the exact set of generated power keys required by the actual final PS plan.

For each `k in K`, run two complementary cases:

### Single-generated case SG(k)

- power k = actual generated Fast power;
- every other required power = oracle/canonical power;
- all non-power arithmetic = fixed D0 stack.

### Single-oracle rescue SO(k)

- power k = oracle/canonical power;
- every other required power = actual generated Fast power;
- all non-power arithmetic = fixed D0 stack.

For each case record:

- final polynomial residual;
- EvalMod real/imag;
- post-S2C;
- public-like;
- pass/fail at `1e-2`;
- all capacity/metadata contracts.

Primary causal rules:

- if exactly one SG(k) fails and its SO(k) rescues `G_ALL`, k is a single generated-power blocker;
- if multiple SG(k) fail or no single SO(k) rescues, classify as multi-power interaction and proceed with bounded cumulative attribution;
- do not infer causality from local power max-norm alone.

---

# D3 — bounded cumulative attribution if needed

Only if D2 does not isolate one key.

Use the actual power dependency/generation order from source/runtime. Build generated powers cumulatively from oracle baseline, one dependency-complete step at a time.

Record the first cumulative set that crosses the public threshold.

Then perform the complementary rescue from `G_ALL` by replacing dependency-complete generated subtrees with oracle states.

Do not brute-force all subsets.

Output the smallest dependency-complete causal power set supported by both directions.

---

# D4 — implementation-vs-schedule decomposition for causal power(s)

For each causal power identified in D2/D3, compare at its completed native checkpoint:

- F: current Fast q0/q1 generated state;
- N: full-RNS mirror of the exact same Fast operation schedule/integer decisions;
- Q: canonical CKKS state for the mathematically correct Chebyshev power at the same native Level/Scale/metadata.

Record:

- F-N;
- N-Q;
- F-Q;
- q0/q1 row equality F vs N;
- centered capacity / value-preservation evidence.

Interpretation:

- F != N => Fast q0/q1 implementation/storage effect;
- F == N and N != Q => generated-power arithmetic schedule/rounding effect;
- capacity failure => physical-wrap blocker and must be handled before precision tuning.

---

# D5 — operation-level localization inside the causal power

For the smallest causal power/subtree only, derive the exact source operation sequence. Depending on the source path this may include:

- operand preparation / balanced copy;
- integer scale balancing;
- multiply or square;
- Chebyshev doubling;
- Rescale;
- subtract-one or subtract-difference-power alignment.

At each source-backed operation boundary:

1. construct the canonical CKKS state for that partial expression at the operation's actual native metadata;
2. reset only that boundary to canonical;
3. replay the remainder of power generation;
4. replay the complete fixed D0 PS/DA/S2C/final path.

Record the latest failing reset and first passing reset.

Do not perform a new generic guard sweep in this task.

---

# D6 — classification

Choose exactly one primary classification:

- `logn13_generated_power_single_blocker_<key>`
- `logn13_generated_power_multi_power_blocker`
- `logn13_generated_power_fast_q01_implementation_blocker`
- `logn13_generated_power_capacity_blocker`
- `logn13_generated_power_schedule_rounding_blocker`
- `logn13_generated_power_reentry_attribution_mismatch`

Also record:

- causal power key/subtree;
- smallest operation-level causal interval if D5 reached;
- recommended next **design** target.

No production integration is authorized after classification.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-GENERATED-POWER-REENTRY-summary.json`

Include only:

- Primary/Secondary provenance;
- Secondary containment status/new SHA;
- D0 oracle control;
- actual required power-key list;
- G_ALL result;
- SG/SO table;
- cumulative table only if needed;
- causal power F/N/Q comparison;
- operation-reset table only for causal power;
- classification;
- first remaining blocker;
- recommended next design target;
- validation flags.

Do not serialize slot vectors, coefficient arrays, full RNS rows, or full traces.

---

# Validation

Secondary:

- focused guard/regression tests pass;
- `go test ./...`;
- `git diff --check`;
- clean `fast-ckks` after ordinary ff push.

Primary:

- focused test matching `TestFIX001P3.*Generated.*Power.*Reentry`;
- `go test ./...`;
- `git diff --check`;
- artifact no NaN/Inf;
- D0 exact oracle control reproduced;
- q0=56 / planScale92 / G0=2 / F0=3 / DA-local-q2 / T2-guard fixed in every D1-D5 case;
- no legacy q0=55 production result used as a causal comparator;
- no LogN16, benchmark, Gate4/5, EXP003;
- no generated-power production redesign/integration;
- Primary ordinary ff push and clean worktree on completion.
