# FIX-001-P3-DESIGN-DOUBLE-ANGLE-NORMALIZED-RECURRENCE

## Purpose

The preceding diagnostic proved the first LogN13 DoubleAngle failure is caused by q0/q1 modular alias after physically restoring the Standard-style polynomial target Scale.

Authoritative finding:

- `FIRST_SUPPORTED_CAUSE = double_angle_mul_q01_alias_confirmed`
- `ROOT_MECHANISM = physical_target_scale_promotion_exceeds_q01_square_uniqueness_capacity`

Key evidence:

- low/compressed state L has rigorous square uniqueness bound ratio `~2.3177e-6` and Fast square error `~4.1543e-7`;
- high/restored state H fails already in `FastCKKS.Mul(H,H)` with error `~0.60758`;
- fresh-q2 witness reports 7941 mismatches;
- H c1/c2 are exactly zero;
- Relinearize(Mul(H,H)) equals direct MulRelin(H,H).

Therefore the next task is **not** another root-cause diagnostic and is **not** yet a production patch.

This task must validate a Fast-specific normalized DoubleAngle recurrence that:

1. preserves the mathematical DoubleAngle function;
2. preserves the Standard level / rescale schedule;
3. keeps the physically maintained q0/q1 coefficients near the already-safe compressed working scale instead of physically promoting them to the Standard target Scale before squaring;
4. avoids q0/q1 alias before every nonlinear multiply and before every Fast Rescale;
5. restores the final modular value only after the last nonlinear/rescale operation;
6. agrees with an independently materialized full-RNS reference.

Diagnostic/design validation only. LogN13 only. Do not modify Secondary production code.

---

## Fixed provenance

Primary repository:

`xuejin-lu/heart-lattigo-bootstrap`

Expected Primary base when authored:

`ccbd56d6b4dede0d9527f38b1149ba2a205bbc74`

Secondary repository:

`xuejin-lu/lattigo`

Exact Secondary:

`61607bb4bb82591009ce768d9a3773bed1497565`

Authoritative polynomial oracle:

`raw_chebyshev_on_preprocessed_z`

Canonical candidate:

`2^91`

Correctness threshold:

`1e-2`

Known integer promotion from compressed polynomial result:

`M = 536870912 = 2^29`

Known compressed polynomial-result Scale:

approximately `2.147483648e9` (`~2^31`)

Known Standard-style target Scale:

approximately `1.1529215046069e18` (`~2^60`)

Known relative difference between `compressedScale * M` and the exact Standard target Scale:

approximately `1.01e-13`.

---

# Scope lock

Do not:

- modify Secondary production code;
- patch `circuits/ckks/mod1/fast.go` yet;
- change Fast Mul/MulRelin/Relinearize/Rescale;
- change polynomial coefficients or PS schedule;
- change the canonical `2^91` polynomial candidate;
- use q2 or higher as maintained Fast production state;
- run LogN16;
- benchmark;
- run Gate 4/5;
- run EXP-003;
- continue to CoeffsToSlots or later bootstrap stages;
- silently use Standard/full-RNS APIs on stale Fast dormant limbs;
- assume that metadata-only Scale changes preserve plaintext semantics unless the normalization factor is tracked explicitly.

This task answers only:

> Can an algebraically equivalent normalized DoubleAngle recurrence execute all three LogN13 rounds using q0/q1-only Fast arithmetic without alias, while matching an independent full-RNS reference?

---

# Mathematical normalization

Let the desired Standard DoubleAngle recurrence be:

`y_{i+1} = 2*y_i^2 - c_i`

where the production code updates `sqrt2pi` before each round, so `c_i` is the exact constant subtracted in that round.

Introduce a positive normalization factor `K_i` and define:

`z_i = y_i / K_i`.

Then for any positive `K_{i+1}`:

`z_{i+1} = A_i * z_i^2 - C_i`

with:

`A_i = 2 * K_i^2 / K_{i+1}`

and:

`C_i = c_i / K_{i+1}`.

This identity is exact.

For this task, every `K_i` must be an exact positive power of two:

`K_i = 2^k_i`.

Require `A_i` to be an exact positive integer power of two as well. Therefore:

`a_i = 1 + 2*k_i - k_{i+1}`

must satisfy:

`a_i >= 0`

and:

`A_i = 2^a_i`.

Integer `A_i` multiplication must not change ciphertext Scale metadata.

---

# Core design principle: keep an approximately constant physical working scale

The already-correct low/compressed polynomial result L is the working-scale anchor.

Let:

`W = L.Scale`.

Do not physically multiply L by `M` before the first DoubleAngle square.

Instead define a coherent initial DoubleAngle metadata Scale:

`S_0 = W * M`.

Set L's Scale metadata to `S_0` **without changing its q0/q1 coefficients**.

Because the coefficients are unchanged while Scale is multiplied by M, the represented normalized variable becomes:

`z_0 = y_0 / M`.

Thus:

`K_0 = M = 2^29`.

Important:

- use `S_0 = W*M` as the internally coherent normalized schedule;
- do **not** immediately normalize `S_0` to the exact Standard target Scale;
- separately record the tiny relative difference from the exact Standard target Scale;
- later quantify the final semantic impact of this difference with a full-RNS exact-target reference.

---

# Construct the K schedule

For each DoubleAngle round i, let `S_i` be the actual ciphertext Scale before the square.

After the normal square and the normal Fast Rescale by the actual dropped modulus `q_drop_i`, the next metadata Scale is:

`S_{i+1} = S_i^2 / q_drop_i`.

Choose `K_{i+1}` as the positive power of two nearest to:

`S_{i+1} / W`.

Equivalently choose integer:

`k_{i+1} = round(log2(S_{i+1}/W))`

with deterministic nearest-power-of-two tie handling.

Require:

- `K_{i+1} > 0`;
- `a_i = 1 + 2*k_i - k_{i+1} >= 0`;
- `A_i = 2^a_i` fits the supported integer-scalar path;
- `S_{i+1}/K_{i+1}` remains close to W.

Record for each round:

- `S_i`;
- `q_drop_i`;
- `K_i` and `k_i`;
- `K_{i+1}` and `k_{i+1}`;
- `A_i` and `a_i`;
- `C_i`;
- effective physical working scale `S_i/K_i`;
- next effective physical working scale `S_{i+1}/K_{i+1}`;
- log2 deviation from W.

Do not hard-code guessed k values if they can be derived from the actual scale/modulus schedule.

---

# Canonical starting state

Reproduce the exact canonical `2^91` polynomial path through the validated post-final-Fast-Rescale low/compressed state L.

Require:

- L Level and Degree match the prior authoritative result;
- L q0/q1 hashes match prior authoritative low-state hashes;
- L Scale matches prior evidence;
- L semantic error vs authoritative polynomial oracle is <= `1e-2` and approximately the prior `~2.6648e-7`;
- maintained c1 is exactly zero in coefficient and NTT/Montgomery domain.

If not:

`NORMALIZED_DOUBLE_ANGLE_PRECONDITION_MISMATCH`.

---

# Independent full-RNS references

The design validation requires independent references that never rely on stale Fast dormant limbs.

## Reference R-coherent

Build a fresh full-RNS degree-one ciphertext from the exact coefficient-domain L state:

1. reconstruct authoritative centered L c0 coefficients from q0/q1;
2. verify L uniqueness is valid;
3. redistribute those exact coefficients into every q limb through the current logical level;
4. set c1 to exact zero in every limb;
5. convert to the required NTT/Montgomery representation;
6. preserve L metadata;
7. physically multiply the full-RNS coefficients by `M` using a valid Standard/full-RNS integer operation;
8. set Scale consistently to `S_0 = W*M`.

This produces a full-RNS ciphertext representing the same unnormalized `y_0` that the normalized Fast path represents as `z_0 = y_0/M`.

Run the ordinary DoubleAngle recurrence on R-coherent using full-RNS arithmetic:

`y <- y*y`

`y <- 2*y`

`y <- y - c_i`

`Rescale`

for all three rounds.

Because c1 is zero, c2 must remain zero. Do not invoke an evaluation-key fallback merely to truncate an exactly-zero c2. Use an explicit diagnostic-only zero-c2 degree reduction if necessary.

This reference must use the same dropped-modulus / level schedule as Fast.

## Reference R-exact-target

Independently build the same full-RNS physically promoted state, but normalize its initial Scale metadata to the exact Standard target Scale used by current Mod1.

Run the same three-round full-RNS DoubleAngle recurrence.

Purpose:

- R-coherent is the algebraic reference for the normalized Fast construction;
- R-exact-target quantifies the semantic effect of using coherent `W*M` rather than the exact Standard target Scale.

Do not use either reference as Fast production code.

---

# Normalized Fast execution

Starting from L:

1. leave q0/q1 coefficients unchanged;
2. set metadata Scale to `S_0 = W*M`;
3. define `K_0=M`;
4. decode and verify the represented value agrees with plaintext `y_0/K_0`.

For each round i = 0..2:

## N0 — pre-square safety

Construct the corrected coefficient-domain q0/q1 view of current normalized Fast `z_i`.

Let:

`B_i = max |coefficient(c0)|`.

Compute the rigorous negacyclic square bound:

`N * B_i^2`.

Require:

`N*B_i^2 < Q01/2`.

If this fails, stop:

`normalized_double_angle_square_capacity_failure`.

## N1 — square

Execute:

`FastCKKS.MulRelin(z_i, z_i)`.

At this point the semantic target is:

`z_i^2`.

Verify normalized semantic error against the high-precision plaintext oracle.

## N2 — normalized recurrence multiplier

Multiply by exact integer:

`A_i = 2*K_i^2/K_{i+1}`.

This is a value multiplication; because A_i is integer, the ciphertext Scale must remain unchanged.

Semantic target becomes:

`A_i*z_i^2`.

## N3 — normalized constant

Subtract:

`C_i = c_i/K_{i+1}`.

Semantic target becomes:

`z_{i+1} = y_{i+1}/K_{i+1}`.

## N4 — pre-Rescale exact-capacity proof

Before Fast Rescale, compare against the independently computed full-RNS normalized reference for the same normalized expression.

The full-RNS reference must provide the exact coefficient-domain integer representative at the current level.

Require every exact c0 coefficient to satisfy:

`|x| < Q01/2`.

Record:

- max exact coefficient magnitude;
- ratio to `Q01/2`;
- outside count.

This is the authoritative pre-Rescale no-alias proof.

Also require Fast q0/q1 residues to match the reference reduced modulo q0/q1.

If exact capacity fails, stop:

`normalized_double_angle_pre_rescale_capacity_failure`.

If residues disagree while exact capacity passes, stop:

`normalized_double_angle_arithmetic_mismatch`.

## N5 — Rescale

Execute the same actual Fast Rescale as production.

Require:

- expected level drop;
- Scale equals `S_{i+1}` from actual dropped modulus;
- semantic result still agrees with `y_{i+1}/K_{i+1}`;
- effective physical working Scale `S_{i+1}/K_{i+1}` remains near W;
- q0/q1 rows agree with the corresponding normalized full-RNS reference after Rescale.

Proceed to the next round only if all checks pass.

---

# Final normalization restore

After round 2 Rescale, the Fast ciphertext represents:

`z_3 = y_3/K_3`.

No nonlinear multiplication or Rescale is allowed after this point in this task.

Recover the original DoubleAngle value with one exact integer scalar multiplication:

`res <- K_3 * res`.

Requirements:

- integer multiplication leaves Scale unchanged;
- q0/q1 rows exactly match R-coherent reduced to q0/q1 at the same point;
- do not require the post-restore centered q0/q1 representative to be a unique exact integer if no later operation requires centered reconstruction;
- clearly label any post-restore capacity metric as modular-only and not a uniqueness proof.

Then perform the same final metadata reset that Mod1 would perform:

`res.Scale = original Mod1 input Scale`.

Compare q0/q1 rows and decoded final output against R-coherent with the same final Scale reset.

They should match exactly in q0/q1 rows and semantically within numerical decoding tolerance.

---

# Exact-target compatibility check

R-coherent starts from coherent Scale `W*M`.

R-exact-target starts from the exact Standard target Scale.

After all three DoubleAngle rounds and final reset to the same original Mod1 input Scale, compare their decoded q0/q1-projected outputs.

Require max component difference <= `1e-2`.

Prefer recording the much tighter observed error.

Also record:

- initial relative Scale delta;
- final semantic delta;
- whether the tiny initial target normalization error is amplified materially by the three DoubleAngle rounds.

If exact-target compatibility fails, do not propose production implementation.

---

# Required checkpoints

For each round store compact checkpoints:

- input normalized semantic error;
- `k_i`, `K_i`;
- Scale and effective Scale `S_i/K_i`;
- square bound ratio;
- post-square semantic error;
- `a_i`, `A_i`;
- normalized constant `C_i`;
- pre-Rescale exact full-RNS coefficient max / Q01 ratio;
- Fast-vs-reference q0/q1 row match;
- post-Rescale semantic error;
- post-Rescale row match.

Final checkpoint:

- `K_3` restore;
- Fast-vs-R-coherent final q0/q1 row match;
- final semantic error;
- R-coherent vs R-exact-target semantic delta.

No full coefficient or slot vectors in artifacts.

---

# Classification

Choose exactly one.

## Case A — design validated

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_recurrence_validated`

Requirements:

- all three pre-square rigorous q0/q1 bounds pass;
- all three pre-Rescale exact full-RNS capacity checks pass;
- every normalized Fast q0/q1 checkpoint matches the independent normalized full-RNS reference;
- every round normalized semantic error <= `1e-2`;
- final K3 restore matches R-coherent q0/q1 rows;
- final decoded Fast output matches R-coherent;
- R-coherent vs R-exact-target final semantic difference <= `1e-2`.

Also record:

`DESIGN_MECHANISM = metadata_normalized_double_angle_with_power_of_two_value_factors`

This classification authorizes a subsequent narrow production task in Secondary `circuits/ckks/mod1/fast.go` plus focused tests.

## Case B — square still exceeds capacity

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_square_capacity_failure`

## Case C — pre-Rescale value exceeds capacity

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_pre_rescale_capacity_failure`

## Case D — Fast arithmetic disagrees with safe reference

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_arithmetic_mismatch`

## Case E — normalized recurrence loses semantic precision

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_precision_failure`

## Case F — coherent target differs too much from exact Standard target

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_exact_target_compatibility_failure`

## Case G — precondition mismatch

`FIRST_SUPPORTED_CAUSE = normalized_double_angle_precondition_mismatch`

---

# Artifacts

Create compact artifacts only:

- `results/FIX-001-P3-DESIGN-DOUBLE-ANGLE-NORMALIZED-RECURRENCE-logN13.json`
- `results/FIX-001-P3-DESIGN-DOUBLE-ANGLE-NORMALIZED-RECURRENCE-logN13-summary.json`

Summary must remain human-reviewable.

Store only:

- provenance;
- W / coherent target / exact target;
- K schedule and A schedule;
- dropped moduli;
- round-level Scale and effective working Scale;
- bound/capacity aggregates;
- semantic aggregates;
- compact q0/q1 hashes / row-match booleans;
- final restore evidence;
- exact-target compatibility evidence;
- classification;
- validation / clean-state evidence.

Do not serialize:

- full coefficient vectors;
- full slot vectors;
- repeated index arrays;
- full operation traces;
- huge hash collections.

---

# Validation

Before completion:

- focused normalized-DoubleAngle tests pass;
- Primary `go test ./...` passes;
- Secondary `go test ./...` passes if invoked;
- Secondary remains exact clean `61607bb4bb82591009ce768d9a3773bed1497565`;
- no Secondary production changes;
- compact artifacts committed and pushed normally;
- Primary and Secondary worktrees clean;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003;
- no CoeffsToSlots or later bootstrap stage.

The deliverable is proof or disproof that a power-of-two normalized DoubleAngle recurrence can retain the Standard level/rescale structure while keeping q0/q1 physical coefficients within the Fast centered-CRT capacity through all three LogN13 rounds.