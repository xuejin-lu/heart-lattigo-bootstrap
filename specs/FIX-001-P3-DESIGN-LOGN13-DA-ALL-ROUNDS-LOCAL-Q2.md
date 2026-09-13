# FIX-001-P3-DESIGN-LOGN13-DA-ALL-ROUNDS-LOCAL-Q2

## Purpose

Close the recurring DoubleAngle Q01-capacity blocker in one bounded design task instead of diagnosing one round at a time.

Accepted evidence from commit `5070ca40194f29d75a0f5aa79e64905ec1af821b`:

- fixed oracle-power PS candidate remains unchanged: q0/q1/q2 = 56/39/40, G0 guard=1, F0 local-q2 guard k=3, immediate contraction after F0;
- polynomial error: real `1.223639867209414e-8`, imag `1.1510743691545144e-8`;
- normalized DA has 3 rounds with accepted schedule around `K_in=24`, `K_out=24`, multiplier exponent `25`;
- DA round0 input and square are Q01-safe;
- round0 intended full-RNS coefficient after multiplier exceeds Q01 centered capacity while q0/q1 residue rows still match modulo Q01;
- bounded local q2 from round0 multiplier through constant and Rescale is exact and contracts successfully back to q0/q1;
- after that contraction, the next first blocker is `round1.after_multiplier_capacity`, with intended Q01 capacity ratios about `2.7241485` real and `2.7241488` imag.

Therefore this task applies the same bounded local-q2 pattern to all remaining DA rounds and, if all three rounds complete, runs S2C/finalization to the real public-like system target.

---

## Required provenance

Primary required base:

`5070ca40194f29d75a0f5aa79e64905ec1af821b`

Secondary exact required commit / branch:

`25e70430b10cd4af37ad4cf94cd994912e970e3f`

`fast-ckks`, clean.

---

# Scope lock

LogN13 only.

Use validated oracle powers and the fixed k=3 PS candidate exactly as accepted.

Keep fixed:

- q0/q1/q2 = existing 56/39/40 profile;
- polynomial coefficients, PS decomposition and common plan scale `2^92`;
- G0 guard=1 and F0 local-q2 guard k=3;
- normalized DA recurrence, constants, K schedule, logical Levels and Rescale count;
- unchanged S2C and one-input finalization/public-like path.

Do not:

- retune PS, guard bits, K values, DA constants or multiplier exponents;
- modify Secondary production code;
- redesign generated powers;
- widen q0/q1;
- use q3+;
- add Q levels;
- benchmark;
- run LogN16, Gate4/5 or EXP003;
- production-integrate.

Primary diagnostic helpers may use exact/big.Int CRT arithmetic.

---

# A0 — reproduce accepted control

Reproduce:

- round0 q0/q1 input and square safe;
- round0 q0/q1 intended after-multiplier capacity failure;
- round0 bounded local-q2 multiplier/constant/Rescale exact versus normalized full-RNS;
- successful round0 q2->q01 contraction;
- next blocker `round1.after_multiplier_capacity` with rows still matching modulo q0/q1.

If not, stop:

`logn13_da_all_rounds_local_q2_precondition_mismatch`.

---

# A1 — deterministic bounded local-q2 rule for DA rounds 0..2

For each DA round `r = 0,1,2`, use the same deterministic rule.

### Preferred activation point: after square

1. Start the round in ordinary q0/q1 mode.
2. Require the round input intended full-RNS value to be Q01-centered-unique.
3. Execute the ordinary q0/q1 square.
4. Compare against the normalized full-RNS intended `after_square` value.
5. If intended `after_square` remains Q01-centered-unique, expand q01->q012 **after square** and before multiplier.

### Fallback activation point: before square

If the intended `after_square` value is not Q01-centered-unique but the round input is Q01-centered-unique:

1. expand q01->q012 at the round input;
2. execute square, multiplier, constant and Rescale under Q012-authoritative semantics;
3. record that this round required `pre_square` activation.

If the round input itself is not Q01-centered-unique, stop with:

`logn13_da_all_rounds_q01_round_input_capacity_blocker`.

This fallback exists only to avoid another one-round-at-a-time diagnostic. Do not extend q2 outside the current DA round unless contraction after Rescale is impossible.

---

# A2 — q012-authoritative DA segment

Once q2 is activated in a round, all remaining operations of that round through post-Rescale are Q012-authoritative:

- square if activation was `pre_square`;
- integer multiplier;
- normalized constant operation;
- unchanged logical Rescale.

Requirements at every exact checkpoint:

- q0/q1/q2 rows match the normalized stage-aligned full-RNS mirror;
- Q012 remains centered-unique;
- Q01 is allowed to alias inside the bounded q2 region;
- rounded division at Rescale exactly matches full-RNS;
- Scale and Level semantics are unchanged;
- no metadata-only relabeling;
- no extra Q level.

Record both intended full-RNS Q01 capacity ratio and Q012 capacity ratio at input/square/multiplier/constant/post-Rescale.

Do not mistake a centered q0/q1 residue representative for proof that the intended coefficient is Q01-unique.

---

# A3 — contract after every DA round

Immediately after each round's Rescale:

1. test the intended full-RNS output against Q01 centered capacity;
2. if Q01-centered-unique, discard q2 and require q01 decode to equal q012/full-RNS semantics <= `1e-10`;
3. continue the next round in ordinary q0/q1 mode.

If post-Rescale output is not Q01-centered-unique, stop with:

`logn13_da_all_rounds_local_q2_cannot_contract_to_q01`

and name the round.

Do not keep q2 across round boundaries in this task when contraction is valid.

---

# A4 — complete EvalMod and downstream

If all three DA rounds complete and contract successfully:

1. record final EvalMod real/imag error versus genuine Standard widened-profile reference;
2. run unchanged Fast S2C;
3. run supported one-input unpack/finalization/public-like path.

Record:

- final EvalMod real/imag max-component error;
- post-S2C max-component error;
- finalization incremental effect;
- public-like max-component error;
- final Level, Scale, Degree, N, slots, NTT/Montgomery metadata.

Primary system criterion:

`public_like_max_component_error <= 1e-2`.

If passed, classify:

`logn13_da_all_rounds_local_q2_downstream_validated`

and record:

`PS_DA_ARITHMETIC_SYSTEM_SUFFICIENT = true`.

Explicitly preserve the distinction:

- local polynomial heuristic budget `1.2e-8`: slightly missed by the fixed k=3 real branch;
- actual LogN13 public-like correctness threshold: passed.

The next blocker after success is the separately-proven generated-power semantic error; do not production-integrate yet.

If all DA rounds complete but public-like fails, classify:

`logn13_da_all_rounds_local_q2_downstream_insufficient`.

---

# Failure classifications

Use the first applicable exact classification:

- `logn13_da_all_rounds_local_q2_precondition_mismatch`
- `logn13_da_all_rounds_q01_round_input_capacity_blocker`
- `logn13_da_all_rounds_local_q2_expansion_mismatch`
- `logn13_da_all_rounds_local_q2_q012_capacity_failure`
- `logn13_da_all_rounds_local_q2_rescale_mismatch`
- `logn13_da_all_rounds_local_q2_cannot_contract_to_q01`
- `logn13_da_all_rounds_local_q2_downstream_insufficient`
- success: `logn13_da_all_rounds_local_q2_downstream_validated`

Always record exact round and checkpoint for the first blocker.

---

# Static cost accounting

Do not benchmark.

Record compactly for all 3 DA rounds:

- q2 activation point (`post_square` or `pre_square`);
- q01->q012 coefficient-transform count;
- q2 NTT/INTT count;
- q2 limb arithmetic passes;
- number of q2-active squares, multipliers, constant operations and Rescales;
- confirmation that q2 is discarded after every round where Q01 contraction is valid.

---

# Required artifact

Create compact:

`results/FIX-001-P3-DESIGN-LOGN13-DA-ALL-ROUNDS-LOCAL-Q2-summary.json`

Include only:

- provenance;
- A0 control;
- one compact per-round table for activation point and Q01/Q012 capacities;
- expansion/row/rescale/contraction pass flags;
- final EvalMod/S2C/public-like metrics if reached;
- final metadata if reached;
- static q2 cost counts;
- classification;
- `PS_DA_ARITHMETIC_SYSTEM_SUFFICIENT`;
- first remaining blocker;
- validation flags.

Do not serialize vectors, coefficient arrays, complete RNS rows or large traces.

---

# Validation

Require:

- focused tests matching `TestFIX001P3.*DA.*All.*Rounds.*Local.*Q2` pass;
- Primary `go test ./...` passes;
- `git diff --check` passes;
- Secondary exact clean at `25e70430b10cd4af37ad4cf94cd994912e970e3f`;
- no Secondary production changes;
- q2 confined to bounded DA round regions and contracted after each valid round;
- no q3+;
- no q0/q1 widening;
- no PS/generated-power redesign;
- no production integration;
- no extra Q levels;
- no LogN16/benchmark/Gate4/5/EXP003;
- compact artifact committed;
- Primary ordinary fast-forward pushed under standing safe-push authorization;
- Primary worktree clean and synchronized.