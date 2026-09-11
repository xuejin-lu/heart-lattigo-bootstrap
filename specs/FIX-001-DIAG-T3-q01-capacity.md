# FIX-001-DIAG-T3-CAPACITY — Prove or reject q0/q1 pre-Rescale capacity failure

## Purpose

The corrected formal diagnosis now establishes:

- repaired `T1` passes;
- repaired `T2` passes;
- formal `T3 = 2*T1*T2 - T1` uses strict `MulRelin` with split `(1,2,1)`;
- the first supported T3 failure occurs **before** the `c>0` correction term;
- specifically, the post-Rescale value after `MulRelin(T1,T2) -> double -> Rescale` has max component error about `0.031234736649481946`;
- therefore `subAligned` is not yet causal.

This task must prove or reject the strongest remaining hypothesis:

> The true coefficient-domain value of the formal T3 pre-Rescale product exceeds the centered uniqueness range of the two maintained limbs, so Fast q0/q1 CRT reconstructs a wrapped representative before dividing by the logical rescale modulus.

Do not repair production Lattigo in this task.

---

## Fixed provenance

Primary repository: `xuejin-lu/heart-lattigo-bootstrap`

Expected Primary remote base when this spec is authored:

`d95c376d290c9b33aee19673455d11c3552276b1`

Repaired Fast backend:

`87be78ff3c591932699aba63d3be46ca306a6eea`

Standard reference backend, when a full-Q evaluator is needed for the diagnostic reference:

`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

The parent FIX-001 remains incomplete.

Follow both repositories' `AGENTS.md`. Both worktrees must be clean at startup. Do not reset/stash/discard unrelated work.

---

## Formal workload

Reproduce exactly the current corrected formal path:

- LogN = 13
- all 4096 slots
- `configs/bootstrap_config.logN13.json`
- same deterministic application input
- same Q2 real branch
- same E1 scale reinterpretation
- same E2 offset
- corrected Chebyshev oracle with `T0=1`
- repaired Fast exactly `87be78ff...`

Use the **actual repaired production T1 and T2 ciphertexts** generated from formal E2.

Require before interpretation:

- E2 matches historical corrected formal E2;
- T1 passes corrected oracle;
- T2 passes corrected oracle.

Do not run LogN16.

---

## Why coefficient-domain evidence is required

Fast `Rescale` at levels >1 does not possess q2...qL residues. It:

1. Partial-INTTs the maintained q0/q1 residues;
2. centered-CRT reconstructs one integer from q0/q1;
3. divides/rounds that centered representative by the logical last modulus `q_level`;
4. writes the quotient back to q0/q1.

Therefore Fast semantics are only unambiguous when the intended pre-Rescale coefficient integer lies within:

`|x| < q0*q1/2`.

Slot-level decoding of the pre-Rescale ciphertext is not a valid way to test this bound. This task must inspect coefficient-domain integers with a higher-capacity diagnostic reference.

---

## Diagnostic policy

Do not modify existing production Lattigo files.

Temporary, uncommitted, build-tagged Secondary helpers are allowed only when required to:

- snapshot actual repaired T1/T2 maintained rows;
- access existing Fast PartialINTT / representation conversion paths;
- reproduce exact production T3 metadata.

Prefer to perform the high-capacity lift/reference logic in Primary using public ring/CKKS APIs.

After raw evidence is written to `/tmp`, remove all temporary Secondary files and restore exact clean `87be78ff...`.

---

# Part 1 — Recover the actual T1/T2 coefficient representatives

For T1 and T2 separately:

1. deep-copy the actual repaired production ciphertext;
2. for **c0** at minimum, convert maintained q0/q1 from NTT to coefficient domain using the same representation conventions as Fast Rescale;
3. if source is Montgomery, IMForm q0/q1 on the copy;
4. centered-CRT q0/q1 into signed integers in `[-Q01/2, Q01/2)` for every coefficient;
5. record all N coefficients as signed integer strings.

Where practical also perform the same for c1 for completeness, but c0 is mandatory because the authoritative Fast zero secret means c0 determines the decoded message.

Record:

- q0, q1, Q01=q0*q1, Q01/2;
- T1/T2 Level and exact Scale;
- max absolute centered coefficient for T1 c0;
- max absolute centered coefficient for T2 c0.

These source coefficients must individually be representable under q0/q1 by construction. Do not infer anything yet about the product.

---

# Part 2 — Full-Q lift of the recovered signed coefficients

Construct diagnostic full-RNS polynomials at:

`commonLevel = min(T1.Level(), T2.Level())`

For each recovered signed c0 coefficient of T1/T2:

- reduce the signed integer modulo every q_i for `i=0..commonLevel`;
- populate every RNS limb;
- transform to the ordinary NTT representation expected by public ring multiplication.

This lift must preserve exactly the chosen centered integer representatives recovered from q0/q1.

Do not use CKKS Encode here; the goal is coefficient-exact lifting, not re-encoding approximate slot values.

Validate the lift by projecting the lifted polynomial back to q0/q1 and requiring bit-exact equality with the recovered ordinary coefficient residues before multiplication.

---

# Part 3 — High-capacity T3 product reference

Using the full-Q lifted **c0** polynomials, compute with ordinary full-RNS ring arithmetic:

### P1 — product before doubling

`P1 = T1_c0 * T2_c0`

using negacyclic ring multiplication at `commonLevel`.

### P2 — doubled product

`P2 = 2 * P1`

This corresponds to the c0 message-bearing component of the production strict `MulRelin -> Add(out,out)` path under the authoritative zero-secret semantics.

Transform P1/P2 back to coefficient domain.

Use an existing source-backed RNS-to-bigint / centered reconstruction helper if available. Inspect Lattigo ring APIs before implementing custom CRT.

For every coefficient of P1 and P2 recover the centered full-Q integer and record:

- absolute magnitude;
- sign;
- whether `abs(x) < Q01/2`;
- if not, the integer wrap count `k` such that the q0/q1 centered representative equals `x - k*Q01` and lies in `[-Q01/2,Q01/2)`.

Also verify:

`abs(x) < Q_full/2`

for every reconstructed P1/P2 coefficient used for interpretation. If full-Q itself wraps, stop with `FULL_Q_REFERENCE_CAPACITY_INSUFFICIENT`.

---

# Part 4 — Capacity statistics

For P1 and P2 separately record:

- coefficient count N;
- max absolute coefficient;
- max ratio `abs(x)/(Q01/2)`;
- number and percentage of coefficients outside q0/q1 centered range;
- min/max nonzero wrap count k;
- histogram of small wrap counts where practical;
- indices of maximum magnitude and representative failing coefficients.

Classification granularity:

- if P1 already exceeds q0/q1/2: candidate crossing point = multiplication;
- if P1 is entirely inside but P2 exceeds q0/q1/2: candidate crossing point = doubling;
- if both remain inside: reject pre-Rescale q0/q1 capacity hypothesis.

---

# Part 5 — Full-Q Rescale oracle

Create a diagnostic full-RNS degree-1 ciphertext whose:

- c0 = P2 full-Q polynomial;
- c1 = zero;
- Level = production T3 pre-Rescale Level;
- Scale = exact production doubled T3 Scale;
- NTT/domain metadata are valid for Standard CKKS Rescale.

Run the pinned Standard/public CKKS Rescale semantics using the **same logical divisor modulus** as the production T3 Rescale.

Decrypt/decode with zero secret.

Compare all 4096 slots against the local semantic oracle:

`2 * T1_actual * T2_actual`

Require fixed threshold:

- max_abs_real <= 1e-2
- max_abs_imag <= 1e-2

If the full-Q Rescale reference itself fails, stop with:

`FULL_Q_RESCALE_ORACLE_INVALID`.

Do not use a failed full-Q reference to blame q0/q1 capacity.

---

# Part 6 — Fast q0/q1 wrapped-reference equivalence

For each full-Q P2 coefficient `x`, compute its q0/q1 centered representative:

`x_wrapped = centered_mod_Q01(x)`

Then reproduce the exact Fast division/rounding semantics from `schemes/ckks/fast/rescale.go` using:

- the production T3 Rescale divisor;
- correct signed rounding behavior.

Create the resulting q0/q1 post-Rescale diagnostic polynomial/ciphertext and decode it at the exact production post-Rescale Scale.

Compare this simulated q0/q1-wrapped result against the actual repaired Fast T3-3 post-Rescale ciphertext.

Required evidence:

- coefficient-level q0/q1 rows bit-exact if the simulation exactly matches implementation;
- otherwise decoded max component <= numerical noise with an explanation for any representational difference.

This step is what turns a capacity correlation into a causal model.

---

# Part 7 — T2 control

Run the same coefficient-capacity accounting for the repaired formal T2 pre-Rescale doubled product:

`2 * T1 * T1`

Do not need to archive the same volume of raw data if large; summary statistics plus representative coefficients are sufficient.

Record:

- whether T2 P1/P2 exceed Q01/2;
- max capacity ratio;
- overflow coefficient count;
- actual repaired T2 post-Rescale error (already expected to be ~4.88e-4).

This control is mandatory because T2 uses the same Fast Rescale implementation and passes.

---

# Predicted slot-error sanity

If P2 wraps, compute a diagnostic estimate of the logical error introduced by one or more `k*Q01` coefficient wraps after division by the actual rescale modulus and interpretation at the output Scale.

Do not rely only on a scalar back-of-envelope estimate because CKKS slots mix polynomial coefficients.

However record the rough scalar quantity:

`Q01 / (rescale_divisor * post_rescale_scale)`

and its log2.

Compare its order of magnitude with the observed T3 max component error `~0.03123473665`.

The full wrapped-reference simulation remains authoritative.

---

# Classification

Choose exactly one primary result.

## Case A — P1 already violates Q01/2, full-Q oracle passes, wrapped simulation matches Fast

`FIRST_SUPPORTED_CAUSE = t3_multiply_crosses_q01_capacity`

## Case B — P1 fits, P2 first violates Q01/2, full-Q oracle passes, wrapped simulation matches Fast

`FIRST_SUPPORTED_CAUSE = t3_doubling_crosses_q01_capacity`

## Case C — P1/P2 violate Q01/2 but wrapped simulation does not explain actual Fast failure

`FIRST_SUPPORTED_CAUSE = q01_capacity_correlated_but_not_causal`

Do not repair yet.

## Case D — P1/P2 stay entirely within Q01/2

`FIRST_SUPPORTED_CAUSE = q01_capacity_rejected`

Then the next task must distinguish MulRelin arithmetic from Rescale implementation by another oracle.

## Case E — full-Q diagnostic reference cannot be validated

`FIRST_SUPPORTED_CAUSE = unresolved_full_q_reference`

State exactly which validation failed.

---

# Required artifacts

Write temporary output to `/tmp` while any Secondary helper exists.

After Secondary is restored clean, create:

- `results/FIX-001-DIAG-T3-CAPACITY-logN13-fast.json`
- `results/FIX-001-DIAG-T3-CAPACITY-logN13-summary.json`

Raw artifact should include sufficient coefficient evidence to reproduce the classification. It does not needlessly need every full-Q RNS limb if signed centered big integers and hashes are archived.

Summary must include:

- exact provenance;
- T1/T2 Level/Scale;
- production T3 pre/post-Rescale Level/Scale;
- exact q0/q1/Q01/Q01_half;
- exact Rescale divisor;
- P1/P2 max magnitude and overflow counts;
- wrap-count evidence;
- T2 control capacity statistics;
- full-Q Rescale metrics;
- q0/q1 wrapped simulation vs actual Fast metrics;
- rough `Q01/(d*scale_out)` magnitude;
- selected first supported cause;
- final clean/test state.

---

# Validation

Before completion:

- Primary `go test ./...` passes;
- repaired Secondary clean `go test ./...` passes;
- any temporary Secondary helper is removed;
- Secondary remains exact `87be78ff...` and clean;
- Primary artifacts committed/pushed normally;
- no Lattigo production source changes;
- no LogN16;
- no benchmark;
- no EXP-003.

---

# Non-goals

Do not:

- modify Fast MulRelin or Rescale;
- add active q2+;
- change Mod1 parameters/scales;
- implement a new power ordering;
- fix subAligned;
- loosen correctness threshold;
- run LogN16;
- benchmark performance;
- start EXP-003.

The deliverable is a coefficient-domain proof or rejection of q0/q1 capacity as the cause of the repaired formal T3 failure.