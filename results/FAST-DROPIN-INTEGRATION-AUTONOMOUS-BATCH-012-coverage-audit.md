# FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012 — Coverage Audit

## Scope and provenance

This is a read-only source audit of the completed, identical-source LogN13 E32 run. Primary runner/evidence changes are confined to the harness and reports; Secondary production code was not changed.

- Primary run commits: Standard `30a312996b5916de97b1e14b4af4438d95660ebe`; Fast `a682e8616e441f0c9fba15b3bf888187fba276d2`. The only intervening Primary change was the batch journal.
- Shared measured frontend SHA-256: `a19841d94311d95a6f0cbab9a3301e49d5820fb5660885f1037a088e1a9204ad`.
- Genuine Standard backend: clean pinned `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Fast backend: clean `fast-qprefix` `00ac70ba136d190fa31bbb26c2f51d003a221634`.
- Fixed config/input SHA-256: `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` / `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.
- Profile: N=8192, 4096 slots, residual MaxLevel 1, Bootstrap MaxLevel 16, E=32. Each backend had one Rotate preflight and exactly one Bootstrap. No benchmark or timing measurement was run.

## Operation coverage

| Operation | Same-source run evidence | Coverage conclusion |
|---|---|---|
| `EncryptNew` | Direct public `rlwe.NewEncryptor(...).EncryptNew`; Standard produced nonzero c1 (8192/8192), Fast produced zero c1. Both decrypted to the fixed input within the preflight gate. | Proven for this profile and the two selected backend builds. Fast is the explicitly insecure zero-secret simulation, not secure encryption. The Fast parameter capability and direct plaintext-to-c0/zero-c1 implementation are visible in Secondary `core/rlwe/encryptor.go:14-19,144-197`; the run records the observed c1 contract. |
| `Add` | Not invoked as a standalone public operation by the shared runner. | Not established as a drop-in public operation. The Fast Q-prefix evaluator has explicit `Add`/`AddNew` implementations (`schemes/ckks/fast/evaluator_ntt.go:22-59`), while the ordinary `ckks.Evaluator.Add` body uses generic `RingQ` operations (`schemes/ckks/evaluator.go:57-115`). No standalone oracle or dispatch counter covers it here. |
| `MulRelin` | Not invoked as a standalone public operation by the shared runner. | Not established end-to-end. The ordinary CKKS wrapper has a bounded Fast zero-secret ciphertext×ciphertext branch and fails closed for unsupported pairs (`schemes/ckks/evaluator.go:756-784`, `schemes/ckks/evaluator_fast_zero_secret.go:21-115`). The explicit Q-prefix evaluator also has `MulRelin` (`schemes/ckks/fast/evaluator_ntt.go:209-235`). This batch does not verify public operand variants or prove which MulRelin kernel an application workload would use. |
| `Rescale` | Not invoked as a standalone public operation by the shared runner. It may occur inside the Fast Bootstrap circuit, but no per-call counter or isolated oracle was recorded. | Standalone public compatibility is unproven. The generic CKKS `Rescale` body performs full logical-chain ring operations (`schemes/ckks/evaluator.go:483-520`); the explicit Fast evaluator has Q-prefix-aware `Rescale` / `RescaleQPrefixRows` (`schemes/ckks/fast/rescale.go:138-185`). This run does not prove the generic public method routes to that Q-prefix implementation. |
| `Rotate` | Direct public `ckks.NewEvaluator(BootstrapParameters, keys).RotateNew(ct, 1)`, followed by public decrypt/decode and a nonconstant rotated-plaintext oracle. Standard made one Galois-key lookup; Fast made zero key/list lookups. | Proven for the tested public path. In the Fast build, the parameter capability selects the zero-secret branch (`schemes/ckks/evaluator.go:1212-1245`; `schemes/ckks/evaluator_fast_zero_secret.go:9-18`), which calls the shared `fastcore` automorphism workspace (`schemes/ckks/evaluator_fast_zero_secret_automorphism.go:58-107`; `schemes/ckks/internal/fastcore/automorphism.go:21-24,69-87`). This proves the bounded Fast Rotate route was selected, not a speedup. |
| `Bootstrap` | Direct public `bootstrapping.Evaluator.Bootstrap` once per backend on the freshly pre-rotated ciphertext. Both outputs decoded and compared against the rotated cleartext oracle. | Proven for this exact E32 profile and the one-call run. In Fast, `GenEvaluationKeys` selects `GenFastEvaluationKeys` for the supported same-N profile (`circuits/ckks/bootstrapping/keys.go:73-78`), which marks the key object Fast-compatible and extends the same-N secret basis (`fast_keys.go:47-52,75-81`). The public constructor selects a `FastEvaluator`, and public `Bootstrap` delegates to it (`evaluator.go:53-64,182-199`). Its source path includes pack/switch, ScaleDown, ModUp/Trace, C2S, real/imag EvalMod, S2C, unpack and public finalization (`fast_bootstrap.go:191-271,274-283,301-395,398-435`). The run flag `fast_bootstrap_selected=true`, output, and invocation count support the selected Fast Bootstrap boundary. No per-arithmetic-kernel counter was collected. |
| `DecryptNew` / `Decode` | Direct public decrypt/decode before and after Rotate and Bootstrap, using the original residual secret. | Proven for this simulation/profile and recorded ciphertexts. The Fast run is not a security or noise-fidelity result: c1 is zero, so ordinary RLWE decryption reduces to the represented c0 plaintext path (`core/rlwe/decryptor.go:40-75`). |

## Numerical and structural result

All measured input/Rotate ciphertexts were Level 0, Scale log2 45, degree one, NTT/non-Montgomery, with one complete active q0 row. Both Bootstrap outputs were Level 1, Scale log2 45, degree one, and had two complete active Q rows.

| Comparison | Complex RMSE | Max complex error | SNR |
|---|---:|---:|---:|
| Standard Rotate vs rotated input | `1.342087606921074e-11` | `5.379396273494723e-11` | `183.67157992669885 dB` |
| Fast Rotate vs rotated input | `7.44997342660252e-13` | `2.1019564094557047e-12` | `208.78410277402207 dB` |
| Fast vs Standard Rotate | `1.3396909694030365e-11` | `5.3631037691780624e-11` | `183.68710464953378 dB` |
| Standard Bootstrap vs rotated input | `4.885854924910455e-9` | `1.5351218791077742e-8` | `132.44838589971747 dB` |
| Fast Bootstrap vs rotated input | `1.1814120448423705e-9` | `5.771241651805313e-8` | `144.77916936116213 dB` |
| Fast vs Standard Bootstrap | `4.954605783029611e-9` | `4.3240068697009917e-8` | `132.3270151591381 dB` |

These are single-run numerical observations for one deterministic message/profile. They are not a general correctness theorem, cryptographic security claim, noise-fidelity result, or runtime comparison.

## Q-prefix / public-boundary limits

The active Q-prefix specification defines real logical q-residues, a capped maintained prefix, and no implicit dormant-row reads or hidden full-RNS fallback (`docs/FAST_QPREFIX_SPEC.md` §§1, 3–4, 13–14, 19). This experiment exercises the P0 public zero-secret path. At input Level 0, full-active-Q and the minimal q0 representation both contain one row, so this run does not stress a higher-level public/full-Q versus explicit compact-Q boundary. The Level-1 result has two active rows. It does **not** prove implicit P0↔C0 conversion, compact-Q public Add/Mul/Rescale interoperability, or that arbitrary compact storage can be passed to ordinary Standard APIs.

The shared source calls ordinary public APIs in both builds; backend selection is at the dependency/build boundary. Fast key generation and zero-secret behavior are properties of the Fast library build. No fallback, configuration change, or Secondary edit was made in this batch.

## Remaining gaps and recommended next tasks

1. **Public Add/Sub coverage:** add a bounded same-source public Add/AddNew (and relevant Sub) oracle case that proves active-row authority and Fast dispatch/no fallback; cover supported operand categories and unchanged Standard behavior.
2. **Public MulRelin coverage:** test ciphertext×ciphertext and plaintext operands separately, recording which Fast path is selected, c1 semantics, row authority, output level/scale, and oracle delta; reject unsupported forms without native key-switch fallback.
3. **Public Rescale integration:** test `Rescale` and `RescaleTo` against Standard logical top-modulus and level/scale semantics, including prefix contraction and Level 0, with evidence that dormant rows are not consumed.
4. **Explicit compact-Q boundary contract:** define and test the public-to-compact and compact-to-public boundary cases, including rejection/materialization rules; do not introduce implicit conversion. Only after these correctness tasks should a separate matched performance task establish runtime speedup.

These are recommendations for Web task planning, not authorization to implement them in this batch.
