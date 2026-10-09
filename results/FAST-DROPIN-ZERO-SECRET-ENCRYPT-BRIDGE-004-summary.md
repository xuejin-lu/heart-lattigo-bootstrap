# FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004

Classification: `FAST_DROPIN_ENCRYPT_BRIDGE_COMPLETE_PENDING_WEB_REVIEW`
Handoff: `READY_FOR_WEB_REVIEW`

The scoped Fast CKKS secret-key `EncryptNew` bridge, unchanged shared frontend, both public E32 Bootstrap calls, and decoded-output comparison completed. This is evidence for the specified simulation lifecycle only; it does not establish native RLWE security, public-key encryption compatibility, or arbitrary multi-operation CKKS compatibility.

## Implementation and validation

- Secondary implementation commit: `eb7793f1e449702590cdba3d9603261623354f80` (`fast-qprefix`), pushed normally to `origin/fast-qprefix`; worktree clean.
- Primary shared frontend commit: `9fe762063af07df92375b0a3464c22430aff906f`; source SHA-256 `f81b4222acc77cd48557fca4626606a42be50bb70bfcaae0600d39c031280e32`.
- Genuine Standard dependency: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean, pinned. Fast dependency: `eb7793f1e449702590cdba3d9603261623354f80`, clean, `fast-qprefix`.
- One identical Primary frontend source compiled against each dependency using development-workspace replacement only. Both preflights passed before either Bootstrap run. Each bootstrap-phase process confirmed its own ordinary KeyGen / `EncryptNew` / `DecryptNew` preflight before the single public Bootstrap call.
- Secondary focused tests passed:
  ```text
  GOCACHE=/private/tmp/fast-dropin-bridge-gocache go test ./core/rlwe ./schemes/ckks ./schemes/bgv ./circuits/ckks/bootstrapping -run '^(TestRLWEParameterProviderKeepsNativeEncryptNew|TestFastCKKSZeroSecretEncryptNewPreservesPlaintext|TestFastCKKSZeroSecretEncryptNewCopiesMetadataDeeply|TestBGVSecretEncryptNewRemainsNative|TestFastBootstrapRejectsNonzeroC1Input|TestFastE32GenEvaluationKeysAndPublicEvaluatorDispatch)$' -count=1
  GOCACHE=/private/tmp/fast-dropin-bridge-gocache go test ./core/rlwe ./schemes/ckks ./schemes/ckks/fast ./schemes/bgv -count=1
  ```
  Both passed. Primary frontend compilation also passed against Standard and Fast (`go test ./tools/fast-dropin-zero-secret-encrypt-bridge-004 -count=1` with the corresponding Standard/Fast `GOWORK`). No additional Bootstrap tests or full circuit suite were run, to preserve the one-call-per-backend bound.

## Frozen profile and provenance

| Field | Value |
|---|---|
| Primary commit / source SHA-256 | `9fe762063af07df92375b0a3464c22430aff906f` / `f81b4222acc77cd48557fca4626606a42be50bb70bfcaae0600d39c031280e32` |
| Config SHA-256 | `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` |
| Deterministic 4096-slot input SHA-256 | `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` |
| Residual / bootstrap | LogN13 / LogN13; residual max Level 1, Scale `2^45`; bootstrap max Level 16 |
| Prime profile | Q bits `[55,39,40,39,40,60,60,61,60,60,60,61,61,56,57,56,56]`; P bits `[61,61,62,61,62]`; Q/P prime hashes match between backends (see run evidence) |
| Bootstrap settings | 4096 slots, E=32, `ModUpThenEncode`; repetitions 1, warmups 0 |
| Environment | Go `go1.26.4`, `darwin/arm64`; both Primary/backend trees clean at execution |

Both frontends used the same source, config, original message, and effective prime profile. Public evaluator construction selected Standard for the pinned build and Fast for the Fast build. The Standard input's `c1` was naturally nonzero; the Fast bridge input had zero `c1` as required. No ciphertext component was manually cleared by the frontend.

## Results

| Check | Standard | Fast |
|---|---:|---:|
| Preflight | passed; Fast dispatch `false` | passed; Fast dispatch `true` |
| Input ciphertext `c1` | 8192 / 8192 coefficients nonzero | 0 / 8192 nonzero |
| Immediate `DecryptNew`/Decode complex RMSE | `8.247557280558041e-12` | `7.449975091875706e-13` |
| Bootstrap calls | 1 | 1 |
| Bootstrap output | Level 1, Scale `2^45`, degree 1; `c1` nonzero (16384 / 16384) | Level 1, Scale `2^45`, degree 1; `c1` zero (0 / 16384) |
| Decoded output vs original: complex RMSE | `4.697396790468602e-9` | `1.1907546782784477e-9` |
| Decoded output vs original: max complex / real / imag error | `1.3779606222747244e-8` / `1.2965157489275292e-8` / `1.2128929135352129e-8` | `5.805793207821362e-8` / `1.6480985430555872e-9` / `5.8057504432468265e-8` |
| Decoded output vs original: SNR | `132.79005231297143 dB` | `144.7107513180764 dB` |
| Decoded output SHA-256 | `5bb78d601da5983a27dd9a929377dfe4b08cb13e989fdf72ec259282bac38c2b` | `b6e6c903382544adb9ee73b04d5b44eed26ef66925c1bcc310fb7c39f717b755` |

Fast-vs-Standard decoded output: complex RMSE `4.821647086585007e-9`; max complex / real / imag difference `4.901316703636267e-8` / `1.2885846606761064e-8` / `4.900648766653637e-8`; SNR `132.56328882843212 dB`. All reported metric aggregates were finite. Exactly one Bootstrap invocation was recorded per backend.

Full decoded arrays and per-run JSON evidence remain local-only under `/private/tmp/fast-dropin-bridge-004-*`; they are not committed. This summary retains hashes and aggregate evidence only.

## Self-review and scope

Reviewed the frontend's unchanged public API path, canonical config/input hashing, Fast evaluator dispatch evidence, preflight-before-Bootstrap gate, exact one-call accounting, and Fast output `c1` guard before decrypting. The Standard and Fast output metadata and shared provenance hashes match the required comparison profile. No production code in Primary, Standard Lattigo, or unrelated schemes was changed. No benchmark, parameter tuning, warmup, extra Bootstrap, or unsupported security/compatibility claim was added.

`git diff --check` passed for the task changes.
