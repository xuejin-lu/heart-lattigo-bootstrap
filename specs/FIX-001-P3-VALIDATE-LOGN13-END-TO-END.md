# FIX-001-P3-VALIDATE-LOGN13-END-TO-END

## Purpose

The LogN13 Fast bootstrap core is now validated through production SlotsToCoeffs.

Accepted predecessor:

- Primary: `9e98a1a9a2a52c4a40ad8107582f7a3e966cc0bf`
- Secondary `fast-ckks`: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`
- classification: `logn13_post_mod1_s2c_core_output_validated`
- core output: Level 1, Degree 1, NTT/Montgomery, Scale `2^45`
- Fast vs independently lifted Standard full-RNS S2C output: exact q0/q1 match after representation normalization
- semantic max-component error: 0

The remaining unvalidated production boundary is:

`core output -> UnpackAndSwitchN2ToN1 -> finalizeFastPublicCiphertext -> public Bootstrap output`

This task validates the **real public `FastEvaluator.Bootstrap` end-to-end LogN13 path**. It does not modify Secondary production code.

---

# Fixed provenance

Primary required base:

`9e98a1a9a2a52c4a40ad8107582f7a3e966cc0bf`

Secondary required exact clean commit:

`ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`

Configuration:

- `configs/bootstrap_config.logN13.json`
- LogN = 13
- LogSlots = 12
- Residual MaxLevel = 1
- bootstrap N = residual N = 8192
- one full-slot ciphertext
- CircuitOrder = `ModUpThenEncode`
- threshold = `1e-2`

Do not change workload or parameter construction.

---

# E0 — disposition of predecessor replay flag

The predecessor summary reported:

`fast_replay_matches_production = false`

This is **not** an arithmetic failure.

The helper `postMod1S2CRowsEqual(params, fast, reference)` normalizes Montgomery representation only on its first argument. In the predecessor comparison both `fastReplay` and `productionOutput` were Montgomery, so only one side was IMForm-normalized before comparison.

Record:

`PREVIOUS_FAST_REPLAY_DISPOSITION = invalid_comparison_due_to_asymmetric_montgomery_normalization`

If the predecessor helper is reused in this task, fix the Primary diagnostic helper or use a new symmetric representation-normalized comparison. Do not modify Secondary for this issue.

The predecessor's authoritative evidence remains:

- every S2C group Fast-vs-full-RNS checkpoint exact after proper normalization;
- production S2C output vs Standard full-RNS public output exact after proper normalization;
- final semantic error 0.

---

# Scope lock

Do not:

- modify Secondary production code;
- modify Fast Mod1, DFT, packing, arithmetic, or key code;
- change parameters/workload;
- run LogN16;
- benchmark;
- run Gate 4/5;
- start EXP-003;
- generalize the validated profile;
- add backend runtime selectors.

This task is correctness validation of the fixed LogN13 public path only.

---

# E1 — reproduce accepted core boundary

Using the same reproducible LogN13 input, reproduce the actual production prefix through the accepted core output:

1. `PackAndSwitchN1ToN2`
2. `ScaleDown`
3. `ModUp`
4. production `CoeffsToSlots`
5. production `EvalMod` on real and imag
6. production `SlotsToCoeffs`

Require the resulting core output to match the accepted predecessor evidence:

- Level = 1
- Degree = 1
- NTT = true
- Montgomery = true
- Scale = `2^45`
- q0/q1 hashes equal the accepted core-output hashes after the same representation convention

If not, classify:

`logn13_e2e_core_precondition_mismatch`.

Preserve the returned `ctxtN1` / `ctxtN2` packing contexts from the same forward packing call. Do not reconstruct them by guesswork.

---

# E2 — validate production UnpackAndSwitchN2ToN1

Call exactly:

`eval.UnpackAndSwitchN2ToN1([]rlwe.Ciphertext{*coreOut}, ctxtN1, ctxtN2)`

on an independent copy of the accepted core output.

Record:

- number of outputs;
- N before/after;
- Level / Degree / Scale;
- LogDimensions before/after;
- NTT/Montgomery state;
- q0/q1 row hashes;
- input core-output immutability.

For this exact profile:

- residual N equals bootstrap N;
- one full-slot ciphertext is used;
- no N2->N1 ring-degree conversion should be required;
- exactly one output is expected.

Do not merely assume identity. Verify exact maintained q0/q1 rows against the source-defined expected unpack result.

If the production call errors:

`logn13_e2e_unpack_failure`

If metadata or exact rows differ from the source-defined expected result:

`logn13_e2e_unpack_mismatch`.

---

# E3 — independent public-finalization oracle

Secondary source defines finalization as:

1. require output geometry matches residual parameters;
2. require NTT=true and Montgomery=true;
3. IMForm each maintained q0/q1 row for c0/c1;
4. set `IsMontgomery=false`;
5. set `Scale = residual.DefaultScale()`.

Build an independent Primary oracle from the unpacked ciphertext using exactly this source-defined transformation on a copy.

Do not read dormant limbs as authoritative.

Record before/after row hashes and verify:

- rows after finalization equal direct IMForm of the pre-finalized q0/q1 rows;
- Level remains residual MaxLevel = 1;
- Degree remains 1;
- N remains 8192;
- NTT remains true;
- Montgomery changes true -> false;
- Scale becomes residual default Scale = `2^45`;
- LogDimensions remain correct.

This oracle is structural only; it does not replace the public Bootstrap call.

---

# E4 — real public Bootstrap is the system under test

Call exactly once on an independent copy of the original reproducible input:

`eval.Bootstrap(input)`

Do not replace this with manual stage replay.

Require:

- call succeeds;
- output N = residual N = 8192;
- output Level = residual MaxLevel = 1;
- Degree = 1;
- NTT = true;
- Montgomery = false;
- Scale equals residual default Scale `2^45`;
- LogDimensions correspond to the original full-slot input;
- original input is unchanged.

Compare public Bootstrap output against the independent E3 manually finalized oracle:

- exact q0/q1 row equality;
- exact Level / Degree / Scale / LogDimensions equality;
- same representation flags.

If public Bootstrap fails:

`logn13_e2e_public_bootstrap_failure`.

If it succeeds but differs structurally from the manual source-defined boundary oracle:

`logn13_e2e_public_structural_mismatch`.

---

# E5 — end-to-end semantic oracle

Decode the final public Fast Bootstrap output using the residual parameters and the current zero-secret/plaintext-compatible Fast contract.

Compare against the **original reproducible plaintext workload**, not merely against another Fast intermediate.

Record max-component, max-complex, mean-complex, and worst index/component.

Require:

`max_component_abs <= 1e-2`.

Also confirm no NaN/Inf.

If structural checks pass but end-to-end plaintext semantics fail:

`logn13_e2e_semantic_failure`.

Where practical, also compare the manual E3 finalized oracle semantics to the public Bootstrap output; they should be identical because E4 requires exact q0/q1 rows.

---

# E6 — BootstrapMany one-element consistency

As a lightweight public-boundary consistency check, call:

`eval.BootstrapMany([]rlwe.Ciphertext{inputCopy})`

with exactly one input.

Require:

- one output;
- output exact q0/q1 rows equal `Bootstrap(input)`;
- metadata/flags equal;
- input unchanged.

Do not expand this task into multi-ciphertext packing tests.

If one-element `BootstrapMany` differs:

`logn13_e2e_bootstrap_many_consistency_failure`.

---

# E7 — success boundary

On success classify exactly:

`FIRST_SUPPORTED_CAUSE = logn13_fast_bootstrap_end_to_end_validated`

Record:

- predecessor core-output provenance;
- Secondary exact SHA;
- unpack evidence;
- finalization-oracle evidence;
- public Bootstrap evidence;
- public BootstrapMany consistency;
- final q0/q1 hashes;
- final semantic metric;
- original-input immutability;
- `PREVIOUS_FAST_REPLAY_DISPOSITION`;
- first failing checkpoint = `none`.

A successful result means the **LogN13 correctness path is complete from public input through public Fast Bootstrap output** for the currently bounded profile.

It does **not** authorize LogN16 or benchmarking in this task. Those require a separate next decision/spec.

---

# Required classifications

Choose exactly one:

- `logn13_fast_bootstrap_end_to_end_validated`
- `logn13_e2e_core_precondition_mismatch`
- `logn13_e2e_unpack_failure`
- `logn13_e2e_unpack_mismatch`
- `logn13_e2e_finalization_oracle_failure`
- `logn13_e2e_public_bootstrap_failure`
- `logn13_e2e_public_structural_mismatch`
- `logn13_e2e_semantic_failure`
- `logn13_e2e_bootstrap_many_consistency_failure`

Also record first failing checkpoint or `none`.

---

# Artifact

Create one compact human-reviewable artifact:

`results/FIX-001-P3-VALIDATE-LOGN13-END-TO-END-summary.json`

No raw slot vectors, coefficient arrays, operation traces, or large repeated hash lists.

---

# Validation

Before completion require:

- focused `TestFIX001P3...` end-to-end test passes;
- Primary `go test ./...` passes;
- Secondary remains exact clean `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219`;
- no Secondary production changes;
- actual public `FastEvaluator.Bootstrap` is the system under test;
- one-element `BootstrapMany` consistency passes;
- Primary artifact/supporting code committed and pushed;
- Primary worktree clean and remote synchronized;
- no LogN16;
- no benchmark;
- no Gate 4/5;
- no EXP-003.