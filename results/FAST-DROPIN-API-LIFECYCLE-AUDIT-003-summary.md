# FAST-DROPIN-API-LIFECYCLE-AUDIT-003

**Classification:** `FAST_DROPIN_LIFECYCLE_PARTIAL`
**Handoff:** `NEEDS_WEB_REVIEW`

## Provenance and identical frontend

- Primary commit used for both builds and runtime preflights: `7a87502f4b09205fe5d711e8fc998d0a12ddc9e2`; frontend source is at `tools/fast-dropin-api-lifecycle-audit/main.go`, SHA-256 `e77a308bdd192eaa4a28bc44e32321058b0e4ef99936957145482845e913646c`.
- That one source file, with no backend selector or Fast-only import, was used for both builds. It imports only `core/rlwe`, `ring`, `schemes/ckks`, `circuits/ckks/bootstrapping`, and `circuits/ckks/mod1` public packages.
- Standard source: clean detached `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast source: clean detached `463d494627b2e9e2bfac51aefe7f4ecb3493b68e`. The synchronized Secondary branch is separately at `68c3da8ec18222f985ebb5c43f6a8495b853cfc2` (the only later change is its AGENTS.md workflow update; no Fast production source changed).
- Both used `configs/bootstrap_config.logN13.json`, config SHA-256 `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`, deterministic original-input SHA-256 `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`, E=32, LogN=13, 4,096 slots, residual scale `2^45`, Q-prime SHA-256 `30066778ef1caea959b3357b0a70ff792fe8582ba9545f6b53c8df2a1a6cb568`, and P-prime SHA-256 `70451211d27cd1e6bf092aa3c0752a63147c34842211dcd70d64d50755d318f7`.

## Call-path and runtime findings

| Stage | Standard | Fast | Evidence |
|---|---|---|---|
| Secret key / encryption / decryption | Ordinary `rlwe.NewKeyGenerator`, `GenSecretKeyNew`, secret-key `NewEncryptor` / `EncryptNew`, and matching `NewDecryptor` / `DecryptNew` | Same public RLWE lifecycle and signatures; Fast's `checkPk` difference only rejects Fast-layout public keys, which this secret-key path does not use | Both produced degree-1, two-component, Level-0, Scale `2^45`, N=8192, NTT=true, Montgomery=false ciphertexts. All 8,192 c1 coefficients were nonzero. Native decrypt/decode was finite: Standard RMSE `8.322811412451754e-12`, max complex error `2.4224945257366723e-11`; Fast RMSE `8.300463515225963e-12`, max `2.3121647318459015e-11`. |
| `Parameters.GenEvaluationKeys(sk)` | Standard key-generation path; E=32 dense/sparse keys present | The public wrapper selects `GenFastEvaluationKeys`: residual Level 1, bootstrap Level 16, `ModUpThenEncode`, and input secret `LevelP=-1` satisfy its Fast-dispatch guard. E=32 dense/sparse keys are present and marked internally Fast-compatible. | Both calls succeeded with the same Q/P hashes and configuration. |
| `bootstrapping.NewEvaluator(params, keys)` | Constructs the ordinary CKKS/DFT/Mod1 evaluator; embedded CKKS evaluator present | `fastCompatible` selects `NewFastEvaluator`; embedded CKKS evaluator is nil. The same public constructor succeeds. | Both returned a non-nil evaluator. Runtime observation: embedded CKKS evaluator present `[Standard=true, Fast=false]`. |
| `Bootstrap(ct)` | Standard path would evaluate the ordinary bootstrap circuit | Public wrapper delegates to `FastEvaluator.Bootstrap` | **Not called** in either build. Fast source states that it simulates a zero secret and skips the dense/sparse secret KeySwitch path; meanwhile ordinary `EncryptNew` produced nonzero c1. The Fast structural input validator checks degree, level, NTT/Montgomery, scale and Q-prefix backing, but has no c1-zero check. Therefore the native encrypted input's semantic suitability for the Fast zero-secret path is not established by these preflight passes. |

The first unresolved compatibility boundary is consequently semantic, not a compile or constructor failure: ordinary `EncryptNew` creates a genuine ciphertext under the generated nonzero secret, while the Fast bootstrap path documents a zero-secret interpretation. Calling Bootstrap here would assume the very equivalence the audit is meant to establish, so the run stopped before it. Bootstrap calls: **0 Standard, 0 Fast**. No Add/Mul follow-up was run. No public-key `GenPublicKeyNew` path was exercised.

Relevant Fast source locations at the pinned commit: `circuits/ckks/bootstrapping/keys.go:73-78` (`GenEvaluationKeys` guard); `fast_keys.go:10-14,54-80` (Fast evaluation keys and marker); `evaluator.go:53-64,183-199` (constructor and Bootstrap dispatch); `fast_bootstrap.go:78-81,100-155` (zero-secret comment and structural input checks). Standard/Fast RLWE sources retain the ordinary secret-key generation, `EncryptNew`, and `DecryptNew` APIs; Fast's public-key layout rejection is outside the exercised secret-key path.

## Exact build and run commands

Run from the Primary repository root. The two workspace files differ only in the local Lattigo replacement: `fast-dropin-api-audit-standard.work` points to the pinned Standard checkout and `fast-dropin-api-audit-fast.work` to the pinned Fast worktree.

```sh
GOCACHE=/private/tmp/fast-dropin-api-audit-gocache GOWORK=/private/tmp/fast-dropin-api-audit-standard.work go build -o /private/tmp/fast-dropin-api-audit-standard ./tools/fast-dropin-api-lifecycle-audit
GOCACHE=/private/tmp/fast-dropin-api-audit-gocache GOWORK=/private/tmp/fast-dropin-api-audit-fast.work go build -o /private/tmp/fast-dropin-api-audit-fast ./tools/fast-dropin-api-lifecycle-audit
/private/tmp/fast-dropin-api-audit-standard -config configs/bootstrap_config.logN13.json
/private/tmp/fast-dropin-api-audit-fast -config configs/bootstrap_config.logN13.json
```

Both builds and runtime preflights exited 0. Go emitted non-fatal module stat-cache permission messages during build; they were not Git metadata errors and did not prevent compilation. The preflight outputs matched config/input/Q/P hashes and parameter metadata. `git diff --check` passed. No production code was changed in either Lattigo source.

## Scope and next decision

This is partial evidence, not proof of the drop-in lifecycle: post-bootstrap output, `BootstrapMany`, public-key encryption, and ordinary arithmetic were not tested. The smallest next review surface is the semantic contract at the existing `GenEvaluationKeys` → `NewEvaluator` → Fast public `Bootstrap` boundary: decide whether native nonzero-c1 ciphertexts have a sound zero-secret interpretation before authorizing any Bootstrap test or implementation change. No repair is proposed here; that decision requires Web review. No security-equivalence, general API-completeness, cross-profile, or performance claim is made.

`FAST-STANDARD-PERF-REBASELINE-003` remains **BLOCKED**.
