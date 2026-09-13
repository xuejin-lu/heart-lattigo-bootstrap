# FIX-001-P3-DESIGN-LOGN13-DA-R0-LOCAL-Q2-FEASIBILITY

## Purpose

Continue the accepted LogN13 oracle-power k=3 path past the first downstream capacity failure.

Accepted evidence from commit `b4cdbb28ed8071d8d5dea3059bb0313ce73b5dc2`:

- fixed PS candidate: q0/q1/q2=56/39/40, G0 guard=1, local-q2 F0 guard k=3, immediate q2->q01 contraction after F0;
- polynomial error: real `1.223639867209414e-8`, imag `1.1510743691545144e-8`;
- normalized DA schedule: `K_in=24`, `K_out=24`, multiplier exponent `25`;
- DA round0 input and `after_square` are Q01-centered-unique and Fast/full-RNS rows match;
- first failure is `round0.after_multiplier_capacity` for both real and imag;
- `after_multiplier` rows still match the normalized full-RNS mirror, so the failure is capacity/range, not an arithmetic mismatch.

This task tests the minimal extension:

> Keep q0/q1 through DA round0 input and square. Immediately before the round0 multiplier, reconstruct q2 from the Q01-unique squared ciphertext. Execute multiplier, normalized constant operation, and the unchanged round0 Rescale under Q012-authoritative semantics. If the post-round0 output is again Q01-centered-unique, contract immediately back to q0/q1 and continue ordinary two-limb downstream execution.

Diagnostic/design only. Do not modify Secondary production code.

---

## Required provenance

Primary required base:

`b4cdbb28ed8071d8d5dea3059bb0313ce73b5dc2`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use validated oracle powers and the fixed k=3 PS candidate exactly as accepted.

Keep fixed:

- q0/q1/q2 existing 56/39/40 profile;
- PS candidate and polynomial output unchanged;
- normalized DA schedule `K_in=24`, `K_out=24`, multiplier exponent `25` unless source-derived exact values prove otherwise;
- same DA recurrence, constants, Levels, Rescale count, and metadata semantics;
- q2 only inside the bounded DA round0 multiplier/constant/Rescale region;
- immediate contraction to q0/q1 after round0 if valid;
- S2C/finalization unchanged.

Do not:

- retune PS, guard bits, K schedule, DA constants, or multiplier;
- modify Secondary production code;
- redesign generated powers;
- widen q0/q1;
- use q3+;
- add Q levels;
- benchmark;
- run LogN16, Gate4/5, EXP003;
- production-integrate.

Primary diagnostic helpers may use exact/big.Int CRT arithmetic.

---

# R0 — reproduce first capacity failure

Reproduce the accepted k=3 downstream control:

- DA round0 input Q01 capacity passes;
- round0 after-square Q01 capacity passes;
- round0 after-multiplier Q01 capacity fails;
- Fast q0/q1 rows at these checkpoints match the stage-aligned normalized full-RNS mirror.

Record exact capacity ratios for real and imag at:

- input;
- after square;
- after multiplier.

If not reproduced, stop:

`logn13_da_r0_local_q2_precondition_mismatch`.

---

# R1 — expand q01 -> q012 after square

At round0 `after_square`, before multiplier:

1. require Q01 centered uniqueness;
2. reconstruct exact centered coefficients from q0/q1;
3. materialize q2 residue from those coefficients;
4. preserve q0/q1 bit-for-bit;
5. place q2 in the correct NTT/Montgomery representation.

Require:

- semantic change <= `1e-10`;
- q2 row equals the normalized stage-aligned full-RNS q2 row;
- q0/q1 unchanged exactly.

If not, classify:

`logn13_da_r0_local_q2_expansion_mismatch`.

---

# R2 — Q012-authoritative multiplier and constant

Execute the exact existing round0 normalized recurrence operations under Q012:

- multiplier corresponding to the accepted exponent `25` / exact source-derived integer;
- normalized constant operation exactly as in the accepted recurrence.

Requirements:

- Q012 remains centered-unique after multiplier and after constant;
- Q01 is allowed to be non-unique inside this bounded region;
- q0/q1/q2 rows match normalized full-RNS at each exact checkpoint;
- Scale metadata follows unchanged mathematical semantics;
- no metadata-only relabeling.

Record Q01 and Q012 centered-capacity ratios before/after each operation.

---

# R3 — Q012-authoritative round0 Rescale

Perform the unchanged logical round0 Rescale using Q012-authoritative centered coefficients and the same divisor/Level transition as the accepted normalized DA path.

Require:

- exact rounded-division agreement with normalized full-RNS;
- q0/q1/q2 output rows match full-RNS;
- same output Scale and Level as the intended recurrence;
- no extra Q level.

The local semantic error relative to ideal real arithmetic is diagnostic only and is NOT by itself a failure if exact full-RNS agreement holds.

---

# R4 — contract after round0

After round0 Rescale:

- measure Q01 centered capacity;
- if Q01-centered-unique, discard q2 immediately and require q01 decode to equal q012/full-RNS semantics <= `1e-10`;
- preserve ordinary metadata.

If contraction is impossible, classify:

`logn13_da_r0_local_q2_cannot_contract_to_q01`.

Do not extend q2 into round1 in this task.

---

# R5 — continue downstream with ordinary q0/q1

Only if R4 contracts successfully:

1. run DA rounds 1 and 2 using the unchanged ordinary q0/q1 Fast path;
2. at every input/square/multiplier/constant/post-rescale checkpoint record capacity and row-match status;
3. if a later DA round first exceeds Q01 capacity, stop and classify `logn13_da_r0_local_q2_later_da_capacity_blocker`, naming the exact checkpoint;
4. if all DA rounds complete, run unchanged Fast S2C and supported one-input finalization/public-like path.

Record:

- final EvalMod real/imag vs genuine Standard widened-profile reference;
- post-S2C error;
- finalization incremental effect;
- public-like max-component error;
- final metadata.

System target:

`public_like_max_component_error <= 1e-2`.

Success classification:

`logn13_da_r0_local_q2_downstream_validated`

and record:

`PS_DA_ARITHMETIC_SYSTEM_SUFFICIENT = true`.

If DA completes but public-like still fails, classify:

`logn13_da_r0_local_q2_downstream_insufficient`.

---

# Static cost accounting

Do not benchmark. Record only:

- q01->q012 coefficient transforms;
- q2 NTT/INTT count;
- q2 arithmetic passes for multiplier/constant/Rescale;
- confirmation that q2 is active only from round0 after-square through round0 post-Rescale.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-DA-R0-LOCAL-Q2-FEASIBILITY-summary.json`

Include only:

- provenance;
- R0 capacity control;
- q01->q012 expansion validation;
- Q01/Q012 capacity table for round0 multiplier/constant/Rescale;
- contraction result;
- later DA checkpoint summary;
- final EvalMod/S2C/public-like metrics if reached;
- static q2 cost counts;
- classification;
- `PS_DA_ARITHMETIC_SYSTEM_SUFFICIENT`;
- first remaining blocker;
- validation flags.

No vectors, coefficient arrays, full RNS row dumps, or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*DA.*R0.*Local.*Q2` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- q2 active only in DA round0 bounded region;
- no q3+;
- no q0/q1 widening;
- no PS/generated-power redesign;
- no production integration;
- no extra Q levels;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.