# FIX-001-P3-INTEGRATE-LOGN13-C2S-COMPRESSION

## Purpose

Integrate the already validated LogN13 CoeffsToSlots capacity fix into the **Secondary Fast-CKKS production path**, then rerun the public Fast Bootstrap correctness boundary.

This is a production integration task, not another design or diagnosis task.

Accepted evidence before this task:

- Primary base: `365b2bbbcb29af305c310895ea788fca0b640688`
- Secondary required base: `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219` on `fast-ckks`
- C2S group 0 design accepted with `k_min_full_group=3`, safety choice `k=4`
- C2S group 1 design accepted with `k_min_full_group=1`, safety choice `k=2`
- groups 2 and 3 require no compression for the validated LogN13 workload
- final real/imag split is correct when groups 0/1 are corrected
- corrected final C2S real error vs genuine Standard: about `9.21e-14`
- corrected final C2S imag error vs genuine Standard: about `9.44e-14`
- `C2S_PRECISION_BLOCKER_RESOLVED_FOR_LOGN13=true`
- corrected matched-input Fast EvalMod no longer violates the `1e-2` local threshold:
  - real Fast vs genuine Standard EvalMod about `4.5663e-3`
  - imag about `4.4350e-3`
- Fast vs normalized full-RNS mirror is exact through the validated EvalMod path
- the historical raw genuine-Standard DoubleAngle q0/q1 projection remains invalid after centered capacity loss and must not be reintroduced as an oracle.

The remaining correctness gap is that the validated C2S compression exists only in Primary diagnostic code; Secondary production Fast DFT still uses the original uncompressed group-0/group-1 plaintext encoding scales.

---

# Architecture constraints

These are mandatory.

1. **Primary remains backend-agnostic.** Do not add Fast selectors or Fast implementation details to the public experiment frontend.
2. **Standard Lattigo behavior must remain unchanged.** Do not alter Standard `dft.NewMatrixFromLiteral`, Standard `dft.Evaluator`, or Standard bootstrap execution semantics.
3. The C2S fix belongs entirely in Secondary Fast code.
4. Do **not** change `MatrixLiteral.Scaling`. That field is the mathematical DFT scaling and is not the compression mechanism.
5. Mathematical diagonal values, diagonal indices, BSGS structure, LevelQ, LevelP, LogDimensions, matrix order, and group count must remain unchanged.
6. Compression changes only the **plaintext encoding Scale** of the selected factor matrices:
   - C2S group 0: original encoding Scale / `2^4`
   - C2S group 1: original encoding Scale / `2^2`
7. After the ordinary one-level Fast Rescale for each compressed group, restore with **both**:
   - physical coefficient multiplication by `2^k` on maintained q0/q1 limbs;
   - metadata Scale multiplication by `2^k`.
8. Restore consumes no additional level.
9. Groups 2 and 3 remain the original production matrices and ordinary Fast execution.
10. SlotsToCoeffs remains unchanged.
11. Do not add a public runtime backend selector or user-facing `--fast-c2s-compression` flag.
12. Restrict the production behavior to the exact/bounded validated LogN13 bootstrap profile. Non-matching profiles must keep the previous Fast DFT behavior.

---

# I0 — safe startup and provenance

Primary startup follows `AGENTS.md` and pulls current `main`.

Secondary modifications are explicitly authorized by this Primary task.

Before editing Secondary require:

- repository `xuejin-lu/lattigo`;
- branch `fast-ckks`;
- clean worktree;
- safe fetch/pull according to Secondary instructions;
- starting commit exactly `ed8b19fc1b51fdf37383f9e196f3a5bed2c03219` unless the only newer commits are task-instruction commits explicitly accepted by this task. If source history differs materially, stop and report instead of silently rebasing assumptions.

Read fresh Secondary instructions and relevant source before editing.

Record actual Secondary base SHA in the final artifact.

---

# I1 — Fast-only bounded LogN13 profile guard

Implement a Fast-bootstrap-side guard for this exact validated C2S profile.

The guard must be source-derived and sufficiently narrow. It should verify the relevant combination, including at minimum:

- Bootstrapping `LogN == 13`;
- effective `LogSlots == 12`;
- `CoeffsToSlots` / HomomorphicEncode;
- C2S factor schedule exactly four one-level groups `[1,1,1,1]`;
- C2S matrix count exactly 4;
- expected C2S starting level / Mod1 boundary compatible with the current LogN13 chain;
- one modulus consumed per rescale;
- current tested Q-chain shape/ranges sufficiently constrained to avoid accidentally applying the fix to an unrelated parameter set.

The guard must not depend on Primary code, environment variables, command-line flags, or a caller choosing “Fast”. It is internal to the already-selected Secondary Fast bootstrap implementation.

For non-matching profiles:

- matrix encoding is unchanged;
- no post-Rescale restore plan is applied;
- existing Fast DFT behavior remains byte-for-byte/source-semantically the default path as far as practical.

---

# I2 — re-encode only C2S group 0 and group 1

In Secondary Fast bootstrap circuit initialization, after ordinary shared bootstrap circuit data has produced the original C2S matrix, construct the Fast LogN13 production C2S matrix from the **same mathematical matrices**.

Recommended source-defined construction:

1. regenerate the mathematical factor matrices from the existing `MatrixLiteral.GenMatrices(params.LogN(), params.EncodingPrecision())`;
2. start from the original encoded transformation metadata for the selected group;
3. create a replacement linear transformation with all structural parameters unchanged;
4. set only its plaintext encoding Scale to:
   - group 0: `originalMatrix.Scale / 16`;
   - group 1: `originalMatrix.Scale / 4`;
5. encode the same mathematical factor matrix values into the replacement transformation;
6. leave group 2 and group 3 transformations untouched.

Require checks/tests that for group 0 and 1:

- mathematical diagonal values are unchanged;
- diagonal index set unchanged;
- BSGS structural metadata unchanged;
- `N1` behavior unchanged;
- LevelQ/LevelP unchanged;
- LogDimensions unchanged;
- only encoded Scale changes by the intended exact power of two.

Do not mutate `MatrixLiteral.Scaling`.

---

# I3 — Fast DFT post-Rescale restore mechanism

Add the smallest clean Fast-only DFT mechanism needed to execute a per-group post-Rescale restore plan.

Preferred architectural shape:

- keep the existing ordinary `FastEvaluator.CoeffsToSlotsNew` behavior unchanged for callers without a restore plan;
- add an explicit Fast-only entry point or internal execution path that accepts a per-group restore exponent plan;
- Fast bootstrap uses it only for the bounded LogN13 C2S path;
- SlotsToCoeffs continues to use the ordinary no-restore path.

Validated production plan:

`[4, 2, 0, 0]`

For each C2S group:

1. execute all factor matrices in the group in the same source order;
2. execute exactly one ordinary Fast Rescale;
3. if restore exponent `k > 0`:
   - multiply maintained physical coefficients by `2^k` using the Fast maintained-limb integer primitive;
   - multiply metadata Scale by the same `2^k`;
   - consume no level;
4. continue to the next group.

After group 0 and group 1 restore, the ciphertext metadata contract must match the original uncompressed production contract expected by the next group.

Validate arguments defensively:

- restore plan length matches DFT group count when supplied;
- exponents are non-negative and supported;
- no restore is silently applied to ordinary callers;
- domain flags and degree remain valid.

Do not duplicate the entire CoeffsToSlots algorithm in bootstrapping if the Fast DFT implementation can support the bounded restore plan cleanly.

---

# I4 — Secondary unit tests

Add focused Secondary tests covering at least:

### Matrix preparation

- bounded LogN13 profile selects `[4,2,0,0]`;
- group 0 encoded Scale = original / 16;
- group 1 encoded Scale = original / 4;
- groups 2/3 unchanged;
- mathematical matrix structure unchanged;
- `MatrixLiteral.Scaling` unchanged.

### Fast DFT restore execution

For a deterministic compatible ciphertext/reference:

- group 0 after compressed LT remains within centered q0/q1 capacity under the already validated design expectation;
- after Rescale + restore, level and Scale equal the original contract;
- group 1 likewise;
- restore consumes no additional level;
- q0/q1 Fast result agrees with an appropriate full-RNS compressed+restore reference;
- no-restore entry point retains previous behavior.

### Guard behavior

- exact LogN13 production profile activates the fix;
- at least one deliberately changed profile characteristic does not activate it and follows ordinary Fast DFT behavior.

Run focused package tests for changed Secondary packages, then `go test ./...` in Secondary.

---

# I5 — production C2S stage validation from Primary

After Secondary implementation is committed and available locally to the Primary module replacement/dependency path, validate the **actual production Fast bootstrap C2S code path**, not the old diagnostic hand-replay.

Use deterministic `reproducibleInput.v1` and the genuine Standard control.

Capture actual production stages through the public/production evaluator path with minimal diagnostic hooks if required.

Require:

- ModUp precondition unchanged;
- production Fast C2S group 0 uses compressed encoding and post-Rescale restore `k=4`;
- production Fast C2S group 1 uses compressed encoding and post-Rescale restore `k=2`;
- groups 2/3 execute once with no compression/restore;
- final split executes once;
- final real and imag Level/Scale/degree/domain match the expected production contract;
- final Fast C2S real vs genuine Standard C2S <= derived C2S budget `1.9531249403614602e-5`;
- final Fast C2S imag vs genuine Standard C2S <= same budget.

The expected diagnostic-scale errors are around `1e-13`, but do not require exact historical floating values. The budget is authoritative.

If production C2S fails despite the design harness passing, classify:

`logn13_c2s_production_integration_mismatch`.

Do not repair by changing group2/3, split, EvalMod, or parameters in the same run unless the failure is clearly an implementation mistake in the requested integration itself.

---

# I6 — public Bootstrap end-to-end validation

Once production C2S passes, run the real public Fast bootstrap API on the deterministic input.

Validate both:

1. `Bootstrap`
2. `BootstrapMany` for the existing supported multi-input contract.

For the single public Bootstrap result require:

- public input unchanged;
- output ring degree correct;
- Degree 1;
- output residual Level correct;
- NTT/domain flags correct at public boundary;
- public output non-Montgomery after finalization;
- final public Scale equals residual DefaultScale;
- plaintext semantic max-component error <= `1e-2` against the original deterministic message.

Also run the genuine Standard control and record its semantic error.

For `BootstrapMany` require:

- no panic/error;
- output count preserved;
- each output satisfies public metadata contract;
- deterministic semantic threshold <= `1e-2` for the checked items;
- input immutability preserved according to current public contract.

If production C2S passes but public Bootstrap still exceeds `1e-2`, classify:

`logn13_public_e2e_semantic_failure_after_c2s_integration`

and record the final error. Do not immediately start another diagnosis in this task.

If public Bootstrap and BootstrapMany pass, authoritative classification:

`logn13_c2s_compression_production_integrated_e2e_validated`.

---

# I7 — preserve EvalMod conclusion

Do not modify Fast EvalMod in this task.

Record from accepted evidence and fresh production E2E result:

- `C2S_PRECISION_BLOCKER_RESOLVED_FOR_LOGN13=true` if I5 passes;
- `HISTORICAL_EVALMOD_MATCHED_INPUT_DIVERGENCE_DISPOSITION=not_a_current_1e-2_blocker_after_c2s_correction`;
- Fast EvalMod implementation unchanged from Secondary base except for unrelated source movement if unavoidable;
- no raw Standard DoubleAngle q0/q1 oracle is used.

A public E2E pass closes the current LogN13 correctness blocker set under the present `1e-2` acceptance threshold. It does **not** authorize LogN16, benchmarking, Gate 4/5, or EXP-003 in this task.

---

# I8 — documentation

Because this task changes durable Secondary production behavior, update the relevant Secondary Fast documentation (prefer `docs/FAST_CKKS_SPEC.md` or the existing canonical Fast design document) with a concise bounded statement:

- current LogN13 Fast C2S uses group-specific plaintext-encoding compression;
- group 0 exponent 4, group 1 exponent 2, later groups 0;
- restore is physical+metadata after ordinary Rescale;
- mathematical DFT scaling and Standard DFT behavior are unchanged;
- this is a bounded LogN13 capacity workaround for q0/q1 exact centered representability.

Do not add thesis prose or historical diagnostic detail to production docs.

---

# Required artifact

Create one compact Primary artifact:

`results/FIX-001-P3-INTEGRATE-LOGN13-C2S-COMPRESSION-summary.json`

Include only:

- Primary/Secondary provenance;
- Secondary changed-file summary and production guard disposition;
- matrix compression plan actually installed;
- production C2S real/imag final semantic metrics;
- production C2S metadata checks;
- public Bootstrap metadata + semantic result;
- BootstrapMany result;
- genuine Standard control result;
- tests run and pass/fail;
- classification;
- first failing checkpoint or `none`;
- worktree/branch cleanliness.

Do not serialize full slots, coefficient arrays, matrices, diagonals, traces, or long per-checkpoint lists.

---

# Allowed classifications

Choose exactly one primary classification:

- `logn13_c2s_integration_precondition_mismatch`
- `logn13_c2s_matrix_compression_integration_mismatch`
- `logn13_c2s_restore_integration_mismatch`
- `logn13_c2s_non_guarded_profile_regression`
- `logn13_c2s_production_integration_mismatch`
- `logn13_public_e2e_semantic_failure_after_c2s_integration`
- `logn13_c2s_compression_production_integrated_e2e_validated`

Record first failing checkpoint or `none`.

---

# Validation and Git workflow

Before completion require:

## Secondary

- focused tests for changed Fast DFT / bootstrap packages pass;
- `go test ./...` passes;
- branch remains `fast-ckks`;
- only task-relevant changes;
- commit the Secondary production implementation and docs;
- because this Primary spec explicitly authorizes the Secondary production change, a normal fast-forward push of `fast-ckks` is permitted under the standing safe-push rule when all Secondary safety conditions hold;
- no force push/rebase/amend/reset/history rewrite.

## Primary

- focused integration/E2E tests pass;
- `go test ./...` passes;
- compact summary artifact committed;
- update task-support code only as needed to validate actual production behavior;
- normal fast-forward push to `origin/main` under standing safe-push authorization;
- Primary worktree clean and synchronized.

Out of scope:

- LogN16;
- benchmark/performance measurement;
- Gate 4/5;
- EXP-003;
- parameter retuning;
- EvalMod redesign;
- S2C changes;
- new noise-fidelity mechanisms.