# FAST-DROPIN-PUBLIC-PRIMITIVE-BOOTSTRAP-AUTONOMOUS-BATCH-017 — Summary

**Result:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` / `READY_FOR_WEB_REVIEW`.

This one-run-per-backend milestone exercised the same ordinary public CKKS frontend through `EncryptNew → AddNew → MulRelinNew → Rescale → RotateNew → bootstrapping.Evaluator.Bootstrap → DecryptNew/Decode`. It used genuine Standard and the intentionally insecure zero-secret Fast backend with the frozen LogN13 E32 profile. It does not establish security/noise equivalence or acceleration.

## Frozen provenance

- Primary runner revision: `f3148a13735f8ba3a4eb9e7dc6a952cf4c1302bd`, clean during both final held runs.
- Shared frontend SHA-256: `b2683887620529e196e68fb949b72929bf0cdeb705771fe35937c9392fd9daf3`.
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean, selected through the pinned Standard GOWORK and `lattigo_standard` build tag.
- Fast Secondary: `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, clean and unchanged.
- Canonical config SHA-256: `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`.
- Canonical 012 seed input SHA-256: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`; common derived workload SHA-256: `f14999e10eb04d7161524d4e798440c279f33d4ea0f3a75e9e68382ab12316e1`.
- Effective profile SHA-256: `16a26b4c12e06ed2c6a4774125b286fb64bbdb66ecde8678fb5c7c59a0221359`; actual Q/P SHA-256: `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`.
- E=32; N=8192 / LogN13; 4096 slots / LogSlots=12; residual MaxLevel=1; Bootstrap MaxLevel=16; default Scale=`2^45`. Full actual Q/P prime arrays are in the compact evidence JSON.

## Checkpoint A — valid Level/Scale and key plan

The residual chain has actual `q0=36028797018652673` and `q1=549755731969`. Inputs A/B use the frozen default scale `2^45`; input C uses the ordinary CKKS encoding scale equal to the actual logical `q1`. This does not change CKKS parameters, config, E, Q/P, or default Scale. It makes the ordinary scale equation close exactly:

`(2^45 × q1) / q1 = 2^45`.

Thus degree-one ciphertext `MulRelinNew` at Level 1 produces Scale `2^45 × q1`; `Rescale` consumes logical `q1`, changes Level 1→0, and returns Scale `2^45`, which is the canonical Level-0 Bootstrap input scale. The public Standard and Fast Bootstrap constructors accepted the same residual profile and input contract. No Bootstrap ran during this checkpoint.

Public `GenEvaluationKeys(sk)` preserved the original secret's q0 residues and supplied the relinearization and rotation key plan. Standard used its native implicit-Standard layout with rotation key P-level 4 matching evaluator MaxLevelP=4. Fast selected its Fast-compatible key plan; its public Bootstrap evaluator selected the Fast path while retaining E=32 in the visible parameters.

## Checkpoints B/C — native public oracles

All eight comparable checkpoints passed finite native `DecryptNew/Decode` plaintext oracles and matched Level/Scale. Direct paired complex errors are below the fixed `1e-6` maximum-error gate.

| Checkpoint | Level | Scale log2 | Fast-vs-Standard RMSE | Fast-vs-Standard max |
|---|---:|---:|---:|---:|
| EncryptNew A | 1 | 45 | 8.2003e-12 | 2.2411e-11 |
| EncryptNew B | 1 | 45 | 8.2522e-12 | 2.3856e-11 |
| EncryptNew C (q1 scale) | 1 | 39.0000 | 5.2533e-10 | 1.6452e-9 |
| AddNew | 1 | 45 | 1.1736e-11 | 3.1800e-11 |
| MulRelinNew | 1 | 84.0000 | 1.3547e-11 | 6.3544e-11 |
| Rescale | 0 | 45 | 1.7087e-11 | 6.6647e-11 |
| RotateNew | 0 | 45 | 2.0131e-11 | 6.9826e-11 |
| Bootstrap | 1 | 45 | 4.9979e-9 | 4.3264e-8 |

Bootstrap oracle metrics:

- Standard: RMSE `4.908811884776509e-9`, max complex error `1.4187430657863595e-8`, SNR `102.06889437527516 dB`.
- Fast: RMSE `1.1790593543156793e-9`, max complex error `5.7394238676200815e-8`, SNR `114.45770879941921 dB`.
- Exactly one public Bootstrap call per backend; total budget used `1 Standard + 1 Fast`, no retry, warmup, benchmark, LogN16, or sweep.

At Level 1 the Fast public outputs retained exactly q0/q1 (2 rows); after Rescale/Rotate at Level 0 they retained q0 only (1 row); the Fast Bootstrap output retained the two active residual rows. Fast c1 stayed zero; genuine Standard EncryptNew produced nonzero c1 and used ordinary native decryption. The checkpoint row hashes, state, per-backend plaintext metrics, direct pair metrics, and provenance are in [the evidence JSON](FAST-DROPIN-PUBLIC-PRIMITIVE-BOOTSTRAP-AUTONOMOUS-BATCH-017-evidence.json).

## Reuse and dispatch evidence

The unchanged public CKKS wrapper selects the existing Fast Q-prefix kernels from the frozen Fast capability: Add/Sub through `addSubFastCKKSZeroSecret` / shared AddSub core, ciphertext `MulRelin` through `mulRelinFastCKKSZeroSecret` / shared Mul core, Rescale through the Q-prefix-aware `fastcore.RescaleCore.ApplyRows`, and Rotate through the shared automorphism core. The pinned source paths are `schemes/ckks/evaluator_fast_zero_secret_add.go`, `schemes/ckks/evaluator_fast_zero_secret.go`, `schemes/ckks/evaluator.go`, `schemes/ckks/fast/rescale.go`, and `schemes/ckks/evaluator_fast_zero_secret_automorphism.go`. Fast primitive operations made zero direct relinearization/Galois-key lookups; evaluator construction enumerated the Galois-key list once. Standard MulRelin retrieved one relinearization key and Rotate retrieved its Galois key.

Public `bootstrapping.NewEvaluator` selected the pinned Standard evaluator or the existing Fast evaluator, and both calls used public `Bootstrap`. Bootstrap's internal per-call key accesses are not dynamically counted: the public Bootstrap key bundle embeds a concrete `MemEvaluationKeySet`. The committed evidence records that limitation and the source-traced path rather than claiming an unobserved zero count. Fast's source-traced zero-secret path omits Standard key-switch/relinearization operations.

## Verification and review boundary

Focused tests and `go vet` passed under both backend workspaces. The implementation was self-reviewed for identical frontend use, exact pinned provenance, public lifecycle, key-plan/Level/Scale gates, fixed Q-prefix rows, no Bootstrap retry, and compact artifacts. No Secondary or Standard source was modified. No performance or security claim follows from this one deterministic workload.

See the journal for the session-control attempt count and exact verification commands. The independent Web review remains the batch acceptance boundary; do not start Batch 018 before that review.
