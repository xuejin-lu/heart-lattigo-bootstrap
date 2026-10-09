# FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013

**Classification:** FAST_DROPIN_P0_PUBLIC_PRIMITIVE_NUMERICAL_COVERAGE_COMPLETE_PENDING_WEB_REVIEW
**Batch status:** BATCH_COMPLETE_READY_FOR_WEB_REVIEW
**Scope:** Primary-only harness, tests, and evidence. Standard and Secondary production repositories were not modified.

## Result boundary

The unchanged public CKKS Evaluator frontend completed the authorized LogN13 P0 full-active-Q checks against both pinned implementations. The result demonstrates bounded numerical/public-API coverage for the tested overloads and operation chain. It does **not** establish general drop-in compatibility, C0 compact-Q interoperability, security equivalence, or a performance result.

The shared frontend and test source hashes match between dependency builds. The genuine Standard run used native ckks.NewKeyGenerator keys and ordinary EncryptNew/DecryptNew; Fast used the same public calls with its intentional zero-secret mode. Both Lattigo worktrees were clean at the pinned commits during the runs.

## Provenance and frozen profile

- Primary during runs: e7dbe218c53b7f92df26da80f9a099c8b24e434a, primary_dirty=true (the authorized runner/test changes were uncommitted during measurement).
- Genuine Standard: 5dbffbdea05394de2ca3a432ed5318aa832e3f40, clean; resolved module directory /private/tmp/fast-standard-rebaseline-002/lattigo-standard through /private/tmp/fast-dropin-api-audit-standard.work.
- Fast: 00ac70ba136d190fa31bbb26c2f51d003a221634 on fast-qprefix, clean; resolved module directory /Users/xuejinlu/Developer/xuejin-lu/lattigo with GOWORK=off.
- Go/runtime: go1.26.4, darwin/arm64.
- Profile: LogN 13; LogQ=[55,39,40,39]; LogP=[60]; default scale 2^45; 16 deterministic complex slots; secret Hamming weight 192; test Levels 3 and 2; rotation left by 1.
- Actual Q: [36028797018652673,549755731969,1099511480321,549756026881]; actual P: [1152921504606830593].
- Profile SHA-256: bb1150b29a9826279c1977047fc015c8d532549a7ef9bed763bbf4302abea0bb.
- Effective Q/P SHA-256: a5c89c81dc72d0a3a0481e926430ac6960bf0db2c72d60627d55e162e2734128.
- Deterministic plaintext workload SHA-256: af17a0326e27bdb93b3d02bcb3e653dcb8e04d8e0e4437cdcf8151b2d2f2b6aa.
- Shared main.go SHA-256: 09cf4453d5fb2b638ed2b04af61e36d4d2e069be0dd6ebbb1b983d8ffbd03bc2; shared main_test.go SHA-256: c2f6e3ca5cba95550b94f74a5c733aaec20207210d54d80c415801782f474535.
- The combined evidence contains aggregate checkpoint metadata only: no coefficient/slot arrays and no secret or evaluation-key bytes. Size: 97,679 bytes.

## Reuse map and source-backed dispatch

The pinned public signatures are the same in Standard and Fast schemes/ckks/evaluator.go:

    Add(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) (err error)
    AddNew(op0 *rlwe.Ciphertext, op1 rlwe.Operand) (opOut *rlwe.Ciphertext, err error)
    Sub(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) (err error)
    SubNew(op0 *rlwe.Ciphertext, op1 rlwe.Operand) (opOut *rlwe.Ciphertext, err error)
    MulRelin(op0 *rlwe.Ciphertext, op1 rlwe.Operand, opOut *rlwe.Ciphertext) (err error)
    MulRelinNew(op0 *rlwe.Ciphertext, op1 rlwe.Operand) (opOut *rlwe.Ciphertext, err error)
    Rescale(op0, opOut *rlwe.Ciphertext) (err error)
    RescaleTo(op0 *rlwe.Ciphertext, minScale rlwe.Scale, opOut *rlwe.Ciphertext) (err error)
    Rotate(op0 *rlwe.Ciphertext, k int, opOut *rlwe.Ciphertext) (err error)
    RotateNew(op0 *rlwe.Ciphertext, k int) (opOut *rlwe.Ciphertext, err error)

| Public method | Signature shape | Reused implementation / observed route |
|---|---|---|
| Add, AddNew | ciphertext plus rlwe.Operand | Existing public CKKS evaluator; ciphertext, scalar, and encoded-vector branches. Fast P0 uses generic full-active-Q operations, not fast.Evaluator. |
| Sub, SubNew | ciphertext minus rlwe.Operand | Same generic public path; tested at Levels 3/2, with scalar/vector and alias cases. |
| MulRelin, MulRelinNew | ciphertext plus rlwe.Operand | Standard ciphertext pairs use the native relin key and GadgetProduct. Fast P0 ciphertext pairs enter mulRelinFastCKKSZeroSecret, reuse the existing product kernel with relin=false, and perform no native key lookup. Scalar/vector operands delegate to generic Mul; no new kernel was added. |
| Rescale, RescaleTo | ciphertext, output; RescaleTo also takes minScale | Existing public CKKS logical full-active-Q path. Rescale consumes the actual top logical q3; RescaleTo is present in both pins and was exercised with a target selecting one rescale. |
| Rotate, RotateNew | ciphertext, rotation, output | Standard uses native Automorphism/Galois key. Fast P0 uses the already-shared rotateFastCKKSZeroSecret → fastcore.ApplyRows route, with no Galois-key lookup. |

The explicit schemes/ckks/fast.Evaluator remains a separate API and was not constructed by this ordinary-public-API test. Thus numeric success for generic Add/Sub, scalar/vector multiplication, and Rescale is not evidence that those operations select an optimized explicit Fast kernel. This preserves the distinction required by the reuse-first audit.

## Coverage and numerical results

All 22 numeric operation checkpoints passed their decoded cleartext oracle, state, and backing checks; both negative controls behaved as specified (the Fast invalid-c1 control is not applicable to genuine Standard ciphertexts). Metrics are decoded complex RMSE / maximum complex error; all values were finite. The inherited 1.0 catastrophic mismatch value is a fail-stop guard only, not a newly declared acceptance threshold.

| Coverage group | Representative coverage | Standard RMSE / max error | Fast RMSE / max error | State / dispatch evidence |
|---|---|---:|---:|---|
| Add / Sub | c/c at Levels 3 and 2; AddNew/SubNew; scalar Add; vector Sub; opOut=op0 aliases | Worst RMSE 8.243743971e-13; worst max 1.580593460e-12 | Worst RMSE 6.739607437e-14; worst max 1.147158705e-13 | Degree 1, Level and Scale preserved, dimensions/domain checked, full active rows, non-aliased inputs byte-identical. Fast remains generic P0 full-Q. |
| c/c MulRelin | MulRelinNew, preallocated MulRelin, square, opOut=op0, Levels 3/2 | Worst RMSE 1.015397572e-13; worst max 2.298199806e-13 | Worst RMSE 5.704500215e-15; worst max 1.211759260e-14 | Standard total: 7 native relin lookups. Fast total: 0; Fast output c1 stayed zero. |
| MulRelin scalar/vector operands | integer scalar 2; []complex128 plaintext vector | Scalar RMSE/max 1.216050878e-12 / 2.325604954e-12; vector 2.060070768e-13 / 3.691746245e-13 | Scalar 8.614893072e-14 / 1.712827274e-13; vector 2.053161097e-13 / 3.626816342e-13 | Existing public overloads delegate to Mul; no relin lookup. An unrecognized struct{} operand was rejected in both builds without input mutation. |
| Invalid Fast c1 negative control | One active c1 residue changed on a cloned Fast P0 ciphertext; ordinary Standard relin key present in keyset | Not applicable: nonzero Standard c1 is legitimate | Expected rejection passed | Fast relin-key lookup delta 0; input and pre-populated output encodings unchanged. |
| Rescale / RescaleTo | Level-3 product; logical q3 divisor; RescaleTo(minScale=2^50) | RMSE/max 5.963338919e-14 / 1.237105912e-13 | 5.508744900e-15 / 1.249024960e-14 | Both consumed q3 once, Level 3→2, Scale divided by actual q3. P0 had four active input rows and three output rows; no dormant row was needed. |
| Add → MulRelin → Rescale → Rotate | Same input messages and public frontend; oracle checked at every link | Final Rotate RMSE/max 1.258718129e-13 / 2.593902953e-13 | 8.004069326e-15 / 2.394245269e-14 | Standard made native relin/Galois lookups. Fast made neither; final Rotate used shared fastcore. Level/Scale transitions were checked at each checkpoint. |

Standard input ciphertexts had nonzero c1; Fast inputs and successful outputs had zero c1. At Level 3, each component had exactly four materialized Q rows (q0..q3); at Level 2, three (q0..q2). Both backends reported identical effective Q/P and workload hashes. No conversion to compact C0 was attempted.

## Remaining boundary / next review question

This milestone only covers P0 public ciphertexts with fully materialized active Q rows. It leaves the explicit C0 compact-Q public interoperability gap untouched. The results also do not show optimized Fast dispatch for generic Add/Sub, scalar/vector multiplication, or Rescale, and make no speed or security claim. Web review should decide the next bounded architecture priority for unchanged-public-API integration across P0 and C0; this report does not choose a new adapter, conversion, or kernel design.

## Validation and cost

Passed under both dependency builds:

    GOCACHE=/private/tmp/fast-dropin-public-primitives-013-gocache GOWORK=off go test ./tools/fast-dropin-public-primitives-batch-013 -count=1
    GOCACHE=/private/tmp/fast-dropin-public-primitives-013-gocache GOWORK=off go vet ./tools/fast-dropin-public-primitives-batch-013
    GOCACHE=/private/tmp/fast-dropin-public-primitives-013-gocache GOWORK=/private/tmp/fast-dropin-api-audit-standard.work go test ./tools/fast-dropin-public-primitives-batch-013 -count=1
    GOCACHE=/private/tmp/fast-dropin-public-primitives-013-gocache GOWORK=/private/tmp/fast-dropin-api-audit-standard.work go vet ./tools/fast-dropin-public-primitives-batch-013

The final paired runner invocations were:

    GOCACHE=/private/tmp/fast-dropin-public-primitives-013-gocache GOWORK=off go run ./tools/fast-dropin-public-primitives-batch-013 -out /private/tmp/fast-dropin-public-primitives-013-fast-final.json -primary-commit e7dbe218c53b7f92df26da80f9a099c8b24e434a -primary-dirty=true -backend-commit 00ac70ba136d190fa31bbb26c2f51d003a221634 -backend-ref fast-qprefix -backend-dirty=false
    GOCACHE=/private/tmp/fast-dropin-public-primitives-013-gocache GOWORK=/private/tmp/fast-dropin-api-audit-standard.work go run ./tools/fast-dropin-public-primitives-batch-013 -out /private/tmp/fast-dropin-public-primitives-013-standard-final.json -primary-commit e7dbe218c53b7f92df26da80f9a099c8b24e434a -primary-dirty=true -backend-commit 5dbffbdea05394de2ca3a432ed5318aa832e3f40 -backend-ref pinned-standard-5dbffbdea05394de2ca3a432ed5318aa832e3f40 -backend-dirty=false
    GOCACHE=/private/tmp/fast-dropin-public-primitives-013-gocache GOWORK=off go run ./tools/fast-dropin-public-primitives-batch-013 -combine-standard /private/tmp/fast-dropin-public-primitives-013-standard-final.json -combine-fast /private/tmp/fast-dropin-public-primitives-013-fast-final.json -out results/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013-evidence.json

Each go test executes the shared arithmetic suite; go vet checks the same new package. Full-repository and Lattigo-wide test suites were not run: the charter calls for the bounded package tests and explicitly caps Bootstrap calls at zero. The final paired runner made zero Bootstrap calls, zero benchmarks, and no LogN16 runs. git diff --check is part of final pre-commit validation.

Evidence: results/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013-evidence.json. Execution journal: results/FAST-DROPIN-PUBLIC-PRIMITIVES-AUTONOMOUS-BATCH-013-journal.md.
