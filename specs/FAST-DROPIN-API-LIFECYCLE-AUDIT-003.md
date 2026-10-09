# FAST-DROPIN-API-LIFECYCLE-AUDIT-003

Status: READY_FOR_CODEX. Task class: M/I evidence-driven API compatibility audit with bounded inexpensive tests. Long-term objective: `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`.

## Accepted predecessor

Web accepts `FAST-ZERO-SECRET-E-METRICS-RECOVERY-002` as a successful **bounded numerical validation** on canonical LogN13: Fast E0 and E32 outputs were identical after decode (RMSE 0 against each other); Fast-vs-original RMSE 1.1907546782784477e-9, and Fast E32-vs-genuine Standard E32 RMSE 4.914687044299207e-9. This does not prove native Fast EncryptNew/DecryptNew, full frontend API compatibility, other parameter profiles or speedup. Evidence: `results/FAST-ZERO-SECRET-E-METRICS-RECOVERY-002-summary.md` on Primary commit `9834c94a68479c167189caa205830243ec033d23`, Secondary `463d494627b2e9e2bfac51aefe7f4ecb3493b68e`.

## Goal: find the first actual obstacle to swap-library-only compatibility

Use one **identical, unchanged frontend Go program** and its exact CKKS configuration against two pinned dependency builds: genuinely native Standard Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and current Fast fork `463d494627b2e9e2bfac51aefe7f4ecb3493b68e`. The frontend MUST import only existing Standard public packages, APIs and signatures, not `schemes/ckks/fast`, `NewFastEvaluator`, custom c0 input constructors or backend-specific flags. Select dependency at harness/build orchestration only, never inside frontend or by changing CKKS parameters. Use E=32 and same exact N/Q/P, scale, slot configuration, deterministic *original message* and sequence.

### Audit before experiment (no production change)

Trace normal frontend lifecycle:

1. `rlwe.NewKeyGenerator(...).GenSecretKeyNew` or ordinary documented equivalent; `GenPublicKeyNew` if exercised.
2. `rlwe.NewEncryptor(...).EncryptNew`; ordinary CKKS encoder and ciphertext creation.
3. `params.GenEvaluationKeys(sk)`, `bootstrapping.NewEvaluator(params, keys)`, `Bootstrap(ct)`.
4. `rlwe.NewDecryptor(...).DecryptNew`; ordinary CKKS decoding.
5. If these pass, only a minimal representative ordinary CKKS operation such as Add or Mul as supported without changing profile or broadening computational scope.

For each constructor/operation record Standard signature, Fast call path, crypto/zero-secret treatment, expected ciphertext representation, actual compatibility, and the first failure. **Critical:** source currently shows `rlwe.NewKeyGenerator(...).GenSecretKeyNew`, `rlwe.NewEncryptor(...)`, `rlwe.NewDecryptor(...)` in the Fast fork remain ordinary RLWE implementations, while `GenEvaluationKeys` conditionally selects Fast material and `NewEvaluator` conditionally selects Fast evaluator. Determine whether this produces a genuine lifecycle mismatch before making any assumptions. Do not silently treat standard ciphertext as zero-secret input.

### Execution bounds

- Start with safe sync and read Primary/Secondary AGENTS and workflow instructions. For this audit use a clean pinned Standard checkout and pinned Fast source; never edit Standard production code.
- First perform static API contract/call-path audit, and compile **the exact same frontend source** against each. If Fast compile fails, record exact error; do not modify frontend, do not execute Bootstrap.
- If compilation succeeds for both, perform an inexpensive native lifecycle **preflight only** with original input and E=32; verify key, encryption, degree/c1 representation and evaluator dispatch, plus decryption consistency where valid. No Bootstrap unless static and runtime preflight already establish that Fast accepts its input under a sound zero-secret interpretation and executing Bootstrap would be meaningful. Default to **zero Bootstrap calls**; no performance benchmark or expensive sweeps.
- Do not change production internals this task. If a minimal source compatibility patch seems needed, propose it with exact file/function scope and mathematical invariants for Web review, rather than guessing a fix. No broad refactor, no bypassing guards, no claims that native ciphertext is zero-secret.
- If an environment or pinned-checkout constraint prevents execution, report precisely; static audit is still a valid partial deliverable.

### Deliverables

One compact Primary report `results/FAST-DROPIN-API-LIFECYCLE-AUDIT-003-summary.md` with concrete call-path table, the **same** frontend source path/commit/hash used in both builds, exact build/run commands, parameters and backend SHAs, first compatibility failure or PASS for each stage, smallest justified next implementation surface, tests, and explicit unsupported/unproven cases. A small reusable diagnostic frontend and build-only harness may be added to Primary **only if it uses exactly the same source for both libraries**; no application production change and no changes to Secondary production code. Do not commit secret keys, ciphertexts or long logs.

Result classifications: `FAST_DROPIN_LIFECYCLE_AUDITED` (source and compile/runtime preflight adequately captured), `FAST_DROPIN_LIFECYCLE_PARTIAL` (identified blockage), `FAST_DROPIN_LIFECYCLE_BLOCKED` (safe sync/tooling impossible), followed by `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`. Self-review, diff check, and safe authorized fast-forward push to Primary; leave Secondary unchanged. The formal Standard-vs-Fast performance rebaseline remains BLOCKED.
