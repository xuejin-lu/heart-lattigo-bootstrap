# FAST-DROPIN-ABOVE-CAP-NUMERICAL-AUTONOMOUS-BATCH-016

**Status:** BATCH_COMPLETE_READY_FOR_WEB_REVIEW
**Classification:** Level-5 compact-Q public primitive chain numerically validated within the frozen deterministic profile.

## Result

One matched Standard/Fast run pair passed the fixed plaintext and Fast-vs-Standard maximum complex-error gate of 1e-6 at both inputs and after every requested public operation. No first numerical divergence was observed. This is a bounded LogN13 Level-5 result, not a Bootstrap, benchmark, security-equivalence, or unrestricted-decode claim.

## Provenance and profile

- Primary runner commit during both runs: 5499629fa2071a6782dd744dd75aa336d0efc0a8; clean.
- Genuine Standard: 5dbffbdea05394de2ca3a432ed5318aa832e3f40; clean independent checkout.
- Fast: fast-qprefix at 2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac; clean.
- Go 1.26.4, darwin/arm64.
- Profile: LogN=13, logical Level=5, 16 slots, Scale=2^45, LogQ=[55,39,40,39,39,39], LogP=[60].
- Actual Q=[36028797018652673,549755731969,1099511480321,549756026881,549755486209,549756174337].
- Actual P=[1152921504606830593].
- Effective Q/P SHA-256: 55739c2e8e86aa61f750ae2f59bf00dca75ddd40d6843ae79912085a304fe46b.
- Primary runner SHA-256: 86bc7c6e3c94e0ab0d52bb0a6e52131130985f121a5adf50324abed1870c7270.
- Deterministic input SHA-256: 296447e0602fcaec7839fd302ec783dd0994418aad18757c20cfe9d38323cd24.

## Numerical checkpoints

RMSE and max columns are complex absolute-error metrics against the cleartext oracle; paired columns compare Fast directly with Standard. Each checkpoint has 16 finite decoded complex samples per backend.

| Checkpoint | Level | Scale log2 | Standard RMSE / max | Fast RMSE / max | Fast-vs-Standard RMSE / max |
|---|---:|---:|---:|---:|---:|
| input-a | 5 | 45.000000000 | 4.5184e-13 / 6.6128e-13 | 4.3074e-14 / 8.5641e-14 | 4.3662e-13 / 6.5520e-13 |
| input-b | 5 | 45.000000000 | 6.7114e-13 / 1.1983e-12 | 4.8063e-14 / 9.4999e-14 | 6.6229e-13 / 1.2494e-12 |
| AddNew | 5 | 45.000000000 | 8.3595e-13 / 1.6098e-12 | 6.3661e-14 / 9.9283e-14 | 8.2325e-13 / 1.6465e-12 |
| MulRelinNew | 5 | 90.000000000 | 1.3168e-13 / 3.8771e-13 | 1.2108e-14 / 3.4128e-14 | 1.3212e-13 / 4.0008e-13 |
| Rescale by q5 | 4 | 50.999999054 | 1.3147e-13 / 3.9444e-13 | 1.2261e-14 / 3.4527e-14 | 1.3144e-13 / 4.0689e-13 |
| RotateNew(1) | 4 | 50.999999054 | 1.3196e-13 / 3.9391e-13 | 1.2260e-14 / 3.4523e-14 | 1.3184e-13 / 4.0598e-13 |

All maximum errors are below 1e-6. Level and Scale match across the backends; Rescale consumes logical q5 and transitions Level 5 to 4.

## Capacity and row authority

At Levels 5 and 4 the Fast maintained width is four rows, q0..q3. Their exact product is 11972622661806413470138489463490221725670618395508737. Independent bounds use the Q-prefix contract: normalized CKKS embedding/fixed-point input bound; Add bound B1+B2; negacyclic MulRelin bound N·B1·B2 for the zero-c1 case; Rescale bound floor((B+(q5-1)/2)/q5); and unchanged bound under the ring permutation.

| Checkpoint | Independent c0 bound B | Observer centered max | 2B<S_Q0123 |
|---|---:|---:|---|
| input-a | 6953922115224 | 1154632540908 | pass |
| input-b | 9066805155832 | 1430771723322 | pass |
| AddNew | 16020727271056 | 1802151382211 | pass |
| MulRelinNew | 1189943808994417608156762865664 | 9218059374559769222516541 | pass |
| Rescale | 2164493760219204398 | 16767541329893 | pass |
| RotateNew(1) | 2164493760219204398 | 16767541329893 | pass |

The existing capacity observer ran at all six checkpoints; its centered representative maxima were checked against the separately propagated bounds. The observer output alone is not treated as proof: the independent bound and strict uniqueness check establish that the q0..q3 representative is the intended lift.

Fast EncryptNew copies the encoded plaintext's full logical-Q backing, so q4/q5 are physically allocated on the two initial inputs. The measurement adapter reads only q0..q3. The public AddNew output compacts to four rows; MulRelinNew, Rescale, and RotateNew outputs likewise have no allocated rows above the fixed prefix. Source audit confirms the Add/Sub and Mul cores receive QPrefixWidth(Level), Rescale receives four source rows while dividing by logical q5, and Rotate applies its shared core to exactly the prefix. Fast relinearization and per-operation Galois-key lookups were both zero; the evaluator enumerated the key list once at construction.

The Fast measurement-only decoder reconstructs the centered c0 lift from q0..q3 after the independent capacity gate and places it into a fresh logical-Level plaintext for the common CKKS Encoder. It does not call Fast full-active-Q DecryptNew and does not read q4/q5. Standard used its native keygen, EncryptNew, DecryptNew, and Decode lifecycle.

## Validation

- Primary focused tests: go test ./tools/fast-dropin-above-cap-numerical-batch-016 — pass.
- Primary focused vet: go vet ./tools/fast-dropin-above-cap-numerical-batch-016 — pass.
- Standard-tag focused tests/vet against the pinned independent checkout — pass.
- Secondary pinned capacity/CRT focused tests — pass; Secondary remained unchanged and clean.
- git diff --cached --check — pass for the runner commit; result artifacts were checked before commit.
- No Bootstrap, benchmark, LogN16, or parameter sweep was run.
