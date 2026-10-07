# FAST-STANDARD-OUTPUT-PREFLIGHT-001 — Genuine Standard/Fast public Bootstrap output preflight

## Scope and authority

**Task class: bounded I + E (research harness output-smoke mode and one-operation numerical preflight).**
**Primary-only repository edits** on `xuejin-lu/heart-lattigo-bootstrap/main`.
**Status when activated: READY_FOR_CODEX.**

Accepted prerequisite: `FAST-STANDARD-INPUT-PROVENANCE-001` and `FAST-STANDARD-COMPARISON-EVIDENCE-002`. They establish backend-specific input provenance and separate formal paired results from historical synthetic fastdiag evidence.

This task is deliberately **not** the seven-repetition Standard/Fast performance rebaseline. The previously authored `FAST-STANDARD-PERF-REBASELINE-003` remains blocked and must never be executed as written.

## Scientific objective

For `configs/bootstrap_config.logN13.json` and `configs/bootstrap_config.logN16.json`, verify that **one genuinely encrypted Standard input** and **one explicitly insecure Fast zero-secret/direct-encoded simulation input**, constructed from the same original vector under identical effective CKKS parameters, can each pass a full public `Bootstrap` operation and be decoded using their *own* backend semantics. Measure each output **against the original message**, not merely against the other backend.

A successful operation and finite output measurement establish *operability and evidence integrity*, not an accepted accuracy threshold. A nonzero output RMSE is normal; do not require Standard and Fast to be bitwise or numerically identical.

## Frozen implementations and constraints

- Standard: clean, detached, **unaltered** `xuejin-lu/lattigo@5dbffbdea05394de2ca3a432ed5318aa832e3f40`, build `perf_standard`, true `NewKeyGenerator/GenSecretKeyNew`, `rlwe.NewEncryptor(...).EncryptNew`, public `Bootstrap`, same-key `NewDecryptor/DecryptNew` and CKKS decode.
- Fast: clean, detached Q-prefix production `xuejin-lu/lattigo@5feb44917fca40c93abec6def6f26bc81a82c536`, build `perf_fast`, declared `fast_zero_secret_direct_encoded_v1` simulated input, public `Bootstrap`, validated Fast c0-direct decode. This is **not** secure Fast encryption. Do not use generic RLWE Encryptor with a Fast public key.
- Same original message SHA and actual effective Q/P primes, LogN, LogSlots, scale, Mod1 parameters and circuit order for a given profile. Do not tune parameters, alter Fast Q-prefix rules, relax input quality, or change any cryptographic source.
- Keep accepted `1e-6` **pre-Bootstrap** maximum complex-slot deviation gate. An output cutoff such as `1e-2` may be reported only as **descriptive**, not as an unreviewed quality pass/fail threshold.
- Results must not depend on `cmd/fastdiag numerical` or inherited synthetic Standard stage traces. It is not a formal reference.

## Authorized minimum implementation

1. Safe-sync Primary/Secondary, check current instructions/spec, and obtain clean detached source builds at both pinned SHAs. Verify the actual compiled dependency source, not only CLI strings. Record Primary harness SHA and clean tree state.
2. Extend the existing Primary `cmd/perfprobe` research harness with a **bounded `--output-smoke` mode** (or an equivalently small independently tested helper) that reuses the **already accepted** backend-specific `PrepareInput`, input provenance/quality gate, `Bootstrap`, and `Decode` functions. The existing `--input-smoke` may remain unchanged. Do not duplicate or fork the encryption path, create a third measurement engine, or introduce a runtime Fast switch to production/application code. Distinct research build tags remain mandatory.
3. In output-smoke mode, construct exactly **one input and perform at most one public Bootstrap per backend/profile**; do not run seven repetitions, timing loops, stage replay, key sweeps, or historical `fastdiag`. Output smoke may record wall time **only to troubleshoot a failure**, clearly not as benchmark latency. Do not claim speedup.
4. After Bootstrap, record output metadata (Level, Scale, Degree, domain, slot count), perform backend-correct decode, and assert: non-nil output, proper finite decoded values, expected slot count, no metadata corruption, no unexpected panic. Compute complex RMSE, maximum complex-slot error, maximum real/imag error and decoded-domain SNR both **output vs original** and **pre-Bootstrap vs post-Bootstrap** with existing `internal/numericalmetrics` where suitable; preserve finite/infinity status instead of manufacturing numbers. Do not require a preset output accuracy gate to pass.
5. Record the *unmodified* Standard constructor/evaluator/decryptor path and Fast simulation/decode path in a compact proof of provenance. Do not serialize secret keys, raw ciphertexts or whole decoded vectors. Retain input vector/config hashes, build tags, exact Standard/Fast/Primary SHAs, full effective-parameter equality within each profile and the accepted pre-Bootstrap max-error evidence.
6. One compact human-readable evidence file `results/FAST-STANDARD-OUTPUT-PREFLIGHT-001-summary.md` and, if useful, one small machine-readable JSON summary. Clearly distinguish **input quality passed** vs **public Bootstrap operation succeeded** vs **output accuracy still unassessed**. If a profile fails, include the exact stage, error and last verified metadata, rather than silently substituting an artificial input or changing Standard.
7. Add focused unit tests for the new CLI mode: no seven-repetition precondition, exactly one public Bootstrap call per invocation, output metadata/decoding validation and non-finite output rejection, refusal of any legacy Standard synthetic input. Avoid expensive real Bootstrap operations in ordinary unit tests; run the four bounded integration executions separately on the pinned builds. Run relevant untagged and both tagged `go test`, `git diff --check`, and one bounded coding self-review.
8. **Out of scope:** Secondary changes; Standard cryptography modifications; production arithmetic, key semantics, parameter tuning, full benchmark, formal speedups, numerical quality certification, deep stage/PS/DoubleAngle attribution, CI overhaul, removal or rewriting of old results, broad root `runner.go` or `cmd/fastdiag` rewrites. If the full native-Standard Bootstrap cannot run, report that actual failure and stop; do not change Standard to make it pass.

## Classification and handoff

- `OUTPUT_PREFLIGHT_OPERABLE_UNASSESSED`: all four bounded backend/profile runs complete, inputs meet `1e-6`, outputs have valid metadata and finite decoded vectors, all metrics recorded. **Quality remains `UNASSESSED`** pending independent Web review; no performance campaign is authorized by this classification.
- `OUTPUT_PREFLIGHT_NUMERICAL_CONCERN`: operation completes but decoded-vs-original output error is conspicuously larger than the descriptive `1e-2` cutoff (per real or imag component), or otherwise warrants Web scientific interpretation. Preserve numbers and do not relabel as an algorithm failure solely from an unaccepted cutoff.
- `OUTPUT_PREFLIGHT_PARTIAL`: only some profile/backend combinations complete.
- `OUTPUT_PREFLIGHT_BLOCKED`: provenance/preflight failure, nonfinite decode, unsupported API, crash, or other blocker; report precise evidence and stop without workarounds.

Make the smallest coherent Primary-only commit(s) and normal push if authorized/safe, report exact changed files, source pins, tests, measured output metrics and `READY_FOR_WEB_REVIEW` (or `NEEDS_WEB_REVIEW`). Web review will decide whether an independent subsequent seven-repetition latency/numerical experiment is justified. Never unblock or execute the old REBASELINE-003.
