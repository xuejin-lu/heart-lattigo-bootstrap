# FIX-001-P3-DIAG-LOGN13-K3-DOWNSTREAM-SUFFICIENCY

## Purpose

Test whether the best valid local-q2 PS candidate is sufficient for the actual downstream LogN13 public-like correctness target, despite missing the conservative local polynomial budget by ~2%.

Accepted evidence from commit `04db11434229ef97d372922cd3ee3f516d37014c`:

- q0/q1/q2 = 56/39/40 bits;
- G0 guard = 1 bit;
- local-q2 F0 guard sweep k=2..6 is fully valid for expansion, Q012 capacity, full-RNS row equality, rounded division, contraction and metadata;
- best final polynomial max error occurs at k=3:
  - real `1.223639867209414e-8`;
  - imag `1.1510743691545144e-8`;
- local planning budget is `1.2e-8`, so k=3 misses by only ~1.97%;
- increasing F0 guard beyond k=3 reduces local F0 rounding error but does not reduce final real polynomial error monotonically, proving the remaining error is not primarily F0 Rescale rounding.

The local `1.2e-8` value is a derived planning bound, not the final system requirement. The true system target remains:

`public_like_max_component_error <= 1e-2`.

This task runs the actual downstream path for k=3 without changing the candidate.

---

## Required provenance

Primary required base:

`04db11434229ef97d372922cd3ee3f516d37014c`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use validated oracle powers.

Fix the PS candidate exactly to:

- q0/q1/q2 = 56/39/40 profile;
- G0 guard = 1 bit;
- local q2 only at F0;
- F0 guard = 3 bits;
- immediate contraction to q0/q1 after F0;
- common plan scale `2^92`;
- unchanged polynomial coefficients, basis, PS decomposition and level topology.

Do not:

- retune the k=3 candidate;
- test more guard bits or combinations;
- modify Secondary production code;
- redesign generated powers;
- widen q0/q1;
- use q3+;
- change C2S/S2C/Mod1 algorithms;
- add Q levels;
- benchmark;
- run LogN16, Gate 4/5, EXP-003;
- production-integrate.

---

# D0 — reproduce k=3 control

Reproduce the accepted k=3 oracle-power PS result:

- real near `1.223639867209414e-8`;
- imag near `1.1510743691545144e-8`;
- q012 expansion/guard/rows/rounded division/contraction all valid;
- same Level/Scale metadata;
- Secondary exact/clean.

If mismatch, stop:

`logn13_k3_downstream_precondition_mismatch`.

---

# D1 — derive downstream normalization from actual k=3 scale

Do not reuse a hard-coded normalization chosen for another candidate.

From the actual k=3 polynomial output Scale:

- derive the valid normalized Mod1 K / DoubleAngle schedule using the same accepted normalized-recurrence methodology;
- preserve the same logical Mod1/DA semantics;
- require q0/q1 centered-uniqueness and Fast/full-RNS agreement at the accepted checkpoints;
- no additional Q level.

Record the derived K/schedule and resulting EvalMod output Level/Scale.

---

# D2 — run downstream stages

Run, in order:

1. normalized DoubleAngle / remaining EvalMod path;
2. unchanged Fast S2C;
3. supported one-input unpack/finalization/public-like path.

Compare against the genuine Standard widened-profile reference using the same corrected-C2S input/workload.

Record compactly:

- polynomial output real/imag error;
- final EvalMod real/imag vs Standard;
- post-S2C error;
- unpack/finalization incremental effect;
- public-like max-component error;
- final Level, Scale, Degree, N, slots and NTT/Montgomery metadata.

---

# D3 — decision rule

Primary system success criterion:

`public_like_max_component_error <= 1e-2`.

If it passes while all semantic/capacity/metadata contracts remain valid, classify:

`logn13_k3_downstream_sufficient_despite_local_budget_miss`

and record:

`PS_ARITHMETIC_SYSTEM_SUFFICIENT = true`.

Do not rewrite history by claiming the `1.2e-8` local budget was met. Explicitly state:

- local heuristic/planning budget: missed;
- actual downstream system threshold: passed.

At that point the PS arithmetic path is sufficient for LogN13 system correctness and the next blocker is the separately-proven generated-power error.

If public-like fails, classify:

`logn13_k3_downstream_insufficient`

and record the first stage where amplification makes the result exceed the system target.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DIAG-LOGN13-K3-DOWNSTREAM-SUFFICIENCY-summary.json`

Include only:

- provenance;
- D0 k=3 control;
- derived normalization/K summary;
- EvalMod/S2C/finalization/public-like metrics;
- final metadata;
- local-budget pass/fail flag;
- public-like threshold pass/fail flag;
- `PS_ARITHMETIC_SYSTEM_SUFFICIENT`;
- classification;
- first remaining blocker;
- validation flags.

Do not serialize vectors, coefficient arrays, complete RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*K3.*Downstream.*Sufficiency` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- no candidate retuning;
- no generated-power redesign;
- no production integration;
- no extra Q levels;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.