# Fast CKKS: Drop-in Zero-Secret Numerical Validation

## User-authorized project objective (2026-10-08)

Fast is a comprehensive fast numerical-validation backend, **not synonymous with Q-prefix**. The user's end goal is to run the exact same frontend CKKS source, configurations and public API calls by **switching only the Lattigo library checkout/version**. The Fast library automatically uses intentionally insecure **zero-secret semantics**, elides unnecessary cryptographic key switching (including dense/sparse switching), and returns CKKS numerical results sufficiently close to correct normally encrypted Standard Lattigo results for high-confidence frontend verification.

Q-prefix is one optimization and is not the definition of Fast. Preserve the possibility of further Fast arithmetic, key and storage optimizations.

### Non-negotiable external contract

- No frontend source changes, runtime fast flags or special Fast constructors required in the **final** drop-in deliverable. Calls including `GenEvaluationKeys`, `NewEvaluator`, `EncryptNew`, `DecryptNew` and `Bootstrap` must work with the same signatures and frontend workflow.
- Frontend-supplied parameters remain visible and unchanged: N/LogN, actual Q/P primes, LogSlots, Scale, Level, operation depth, E (`EphemeralSecretWeight`), and all other CKKS profile settings. Fast cannot silently replace E=32 with E=0 in externally visible data.
- Fast internally interprets secrets as zero for numerical simulation. In that mode Dense/Sparse secret conversion may be omitted when mathematically justified; this does **not** establish equivalence of encryption security or internal noise distribution.
- Output must be measured against the expected cleartext result and a **genuinely encrypted Standard reference using the SAME original frontend E and CKKS profile**, with RMSE, maximum error and SNR reported; failures and unsupported cases must be explicit. Do not guarantee a universal 1e-8 without workload-specific evidence.
- Do not change the pinned genuine Standard source; Fast-fork changes must retain Standard public API shape, and comply with repository architectural invariants. Preserve user-visible CKKS semantics, Level/Scale and supported operations; never silently fall back to full-RNS on unmaterialized Q-prefix rows.

### Known current gap

Fast Bootstrap currently rejects E != 0 even though its zero-secret ModUp has no Dense/Sparse switching. Primary `backend_fast.go` manually builds a direct-encoded c1=0 input and invokes `NewFastEvaluator`, so it **has not proven** transparent library substitution. The Fast fork has a partial `NewEvaluator` compatibility dispatch, but other conventional call paths are not proven end-to-end.

The next task is a **bounded first step**: accept E=0 and E=32 in the Fast zero-secret Bootstrap path, omit unneeded Dense/Sparse KeySwitch, and report actual numeric outcomes, without yet claiming full drop-in API compatibility. A later explicitly authorized integration task must run exactly identical frontend source in native Standard and Fast builds.
