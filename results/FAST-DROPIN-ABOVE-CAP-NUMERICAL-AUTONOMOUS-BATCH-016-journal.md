# Batch 016 measurement journal

## Execution

The synchronized task required one matched public CKKS composition at logical Level 5. The Primary runner was committed as 5499629fa2071a6782dd744dd75aa336d0efc0a8 so both backend runs could record the same clean Primary revision.

- Standard ran once from the independent clean Standard checkout at 5dbffbdea05394de2ca3a432ed5318aa832e3f40.
- Fast ran once from clean fast-qprefix at 2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac.
- Both runs used the same Primary source, deterministic inputs, Q/P chain, and profile. Effective Q/P SHA-256: 55739c2e8e86aa61f750ae2f59bf00dca75ddd40d6843ae79912085a304fe46b.
- Runtime: Go 1.26.4, darwin/arm64.
- Operation path: EncryptNew for both inputs, AddNew, MulRelinNew, Rescale, then RotateNew(1). Rescale used q5 and changed Level 5 to 4.
- Each backend and direct Fast-vs-Standard comparison passed the fixed 1e-6 maximum complex-error gate at all six checkpoints.
- Bootstrap calls: 0. Benchmark, LogN16, and sweep: none.

## Independent capacity method

For the deterministic input slots, the normalized inverse embedding bounds each encoded coefficient by the scaled slot maximum; the runner uses twice that value plus one as a conservative fixed-point/Float64 guard. This is separately checked against q0 for initial encoding and against the actual q0*q1*q2*q3 product for every Fast checkpoint.

Subsequent bounds follow the authoritative Q-prefix formulas:

- Add: B_add = B_a + B_b.
- MulRelin: B_mul = N * B_add * B_b. Fast c1 is zero on all maintained rows, so the product's c0 bound is the negacyclic convolution bound.
- Rescale: B_rescale = floor((B_mul + (q5-1)/2)/q5).
- Rotate: the coefficient permutation preserves the infinity bound.

The independent bound is checked before measurement decoding. At each checkpoint, the existing Fast Q-prefix observer is also invoked, and its centered representative maxima must not exceed the independently propagated bound. A centered representative always lies within half the residue product, so the observer's StrictFit field is not used alone as a mathematical capacity proof.

## Measurement-only high-Level decode

Fast zero-secret EncryptNew stores the encoded plaintext in c0 and clears c1. The measurement adapter converts only the four authoritative q0..q3 rows to coefficient form, checks canonical residues, reconstructs the centered lift with BigInt CRT, verifies it against the independent bound, and extends that lift into a fresh plaintext across the logical Level for the common CKKS Encoder. This code is outside the evaluation path. It never reads dormant q4/q5 and does not use normal full-active-Q Fast decryption.

Source inspection of the pinned Fast implementation found:

- public Add/Sub eligibility validates and dispatches exactly QPrefixWidth(Level) rows to the shared Add/Sub core;
- public zero-secret MulRelin dispatches exactly QPrefixWidth(Level) rows to the shared Mul core and requires authoritative c1 rows to be zero;
- public Rescale passes the maintained source width to the shared transactional core, which uses logical q5 as divisor and checks result capacity;
- public Rotate validates and applies the shared automorphism core to exactly the maintained prefix.

The two EncryptNew inputs retain full logical-Q backing because Fast zero-secret EncryptNew copies the encoded plaintext rows. Their measurement decoder nevertheless reads only q0..q3, and the first public AddNew output is compact. Every subsequent Fast output remains four-row compact, with no physical q4/q5 backing.

## Checkpoint record

| Checkpoint | Level | Scale log2 | Fast bound B | Fast observed max | Fast rows | Result |
|---|---:|---:|---:|---:|---|---|
| input-a | 5 | 45.000000000 | 6953922115224 | 1154632540908 | q0..q3; q4/q5 allocated on EncryptNew input | pass |
| input-b | 5 | 45.000000000 | 9066805155832 | 1430771723322 | q0..q3; q4/q5 allocated on EncryptNew input | pass |
| AddNew | 5 | 45.000000000 | 16020727271056 | 1802151382211 | q0..q3 only | pass |
| MulRelinNew | 5 | 90.000000000 | 1189943808994417608156762865664 | 9218059374559769222516541 | q0..q3 only | pass |
| Rescale | 4 | 50.999999054 | 2164493760219204398 | 16767541329893 | q0..q3 only; q5 divisor | pass |
| RotateNew(1) | 4 | 50.999999054 | 2164493760219204398 | 16767541329893 | q0..q3 only | pass |

The actual q0*q1*q2*q3 product at all six checkpoints was 11972622661806413470138489463490221725670618395508737.

## Numerical record

| Checkpoint | Level | Scale log2 | Standard RMSE/max | Fast RMSE/max | Fast-vs-Standard RMSE/max |
|---|---:|---:|---:|---:|---:|
| input-a | 5 | 45.000000000 | 4.5184e-13 / 6.6128e-13 | 4.3074e-14 / 8.5641e-14 | 4.3662e-13 / 6.5520e-13 |
| input-b | 5 | 45.000000000 | 6.7114e-13 / 1.1983e-12 | 4.8063e-14 / 9.4999e-14 | 6.6229e-13 / 1.2494e-12 |
| AddNew | 5 | 45.000000000 | 8.3595e-13 / 1.6098e-12 | 6.3661e-14 / 9.9283e-14 | 8.2325e-13 / 1.6465e-12 |
| MulRelinNew | 5 | 90.000000000 | 1.3168e-13 / 3.8771e-13 | 1.2108e-14 / 3.4128e-14 | 1.3212e-13 / 4.0008e-13 |
| Rescale | 4 | 50.999999054 | 1.3147e-13 / 3.9444e-13 | 1.2261e-14 / 3.4527e-14 | 1.3144e-13 / 4.0689e-13 |
| RotateNew(1) | 4 | 50.999999054 | 1.3196e-13 / 3.9391e-13 | 1.2260e-14 / 3.4523e-14 | 1.3184e-13 / 4.0598e-13 |

## Verification commands

- go test ./tools/fast-dropin-above-cap-numerical-batch-016
- go vet ./tools/fast-dropin-above-cap-numerical-batch-016
- GOWORK=/private/tmp/batch016-standard-work/go.work go test -tags lattigo_standard ./tools/fast-dropin-above-cap-numerical-batch-016
- GOWORK=/private/tmp/batch016-standard-work/go.work go vet -tags lattigo_standard ./tools/fast-dropin-above-cap-numerical-batch-016
- Secondary: go test ./schemes/ckks/fast -run 'TestCheckQPrefixCapacity|TestObserveQPrefixCapacityReportsExactCenteredComponentBounds|TestPublicFastMulRelinProductPassesExactQPrefixCapacityAtLevelsOneAndThree|TestFastRescaleMatchesBigIntCenteredCRTOracleAtEveryPrefixWidth'
- git diff --cached --check

The two per-run JSON files containing raw decoded slots were kept under /private/tmp and were not added to the repository. The compact evidence JSON contains aggregate metrics, capacity checkpoints, state, and provenance without decoded slot arrays.
