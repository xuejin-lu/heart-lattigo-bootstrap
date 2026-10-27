# Montgomery Noise Representation Repair 001

Date: 2026-10-27  
Task: `FAST-BOOTSTRAP-NOISE-FIDELITY-001`  
Web ruling: `MONTGOMERY_NOISE_REPRESENTATION_REPAIR_REQUIRED`  
Repair status: `READY_FOR_WEB_REVIEW`

## Scope and budget

This is a no-Bootstrap repair of the existing opt-in fixed-sigma noise hook. The original CP-C budget remains exhausted at 3/3 calls: genuine Standard completed once, Fast completed once, and the single Fast+Noise call failed. The Standard and Fast captures, FN failure record, reservations, and attempt ledger are unchanged. No Bootstrap, LogN16 workload, BTS, or binary rebuild was run for this repair. No new attempt was reserved.

## Finding

The retained FN event identifies the failure in Fast Mod1 Chebyshev power 2 at Level 12 and Scale 2^60. The multiplication path requires NTT plus Montgomery representation. The Bootstrap diagnostic helper accepted NTT non-Montgomery inputs only, so it rejected the ciphertext before sampling or injecting any noise. The event evidence records zero injected polynomials. This resolves the representation mismatch; it does not produce or imply a successful FN Bootstrap result.

## Minimal repair

Secondary commit `49b65dd927780cec0d7e05772a3df8554ef58005` on `experiment/fast-bootstrap-noise-fidelity` changes only the existing helper and targeted tests. The helper still constructs the same signed Gaussian coefficient residues and applies per-row NTT. When the target ciphertext is Montgomery, it now applies that row's `MForm` to the transformed noise before adding it to c0. Non-Montgomery behavior is unchanged. The existing strict `fastcore.CheckQPrefixAdditionCapacity` check remains in force; the candidate is committed to the target only after it passes.

The Fast hook itself is unchanged: it copies both operands before injection, derives independent left/right seeds, and then calls the existing multiplication core. Thus the source operands, including a square input, are not mutated by the injection. The repair does not alter multiplication, Rescale, c1, metadata, Scale, Level, dormant Q limbs, or ct×pt/scalar/rotation paths.

## Verification

All commands below ran against Secondary commit `49b65dd927780cec0d7e05772a3df8554ef58005`, with a task-local Go build cache. Each passed:

```text
GOCACHE=/private/tmp/lattigo-fast-bootstrap-noise-gocache go test ./schemes/ckks -run '^(TestFixedSigmaBootstrapMulNoise|TestFixedSigmaCoefficientNoise)' -count=1
ok github.com/tuneinsight/lattigo/v6/schemes/ckks

GOCACHE=/private/tmp/lattigo-fast-bootstrap-noise-gocache go test ./schemes/ckks/fast -run '^TestFixedSigmaBootstrapNoise' -count=1
ok github.com/tuneinsight/lattigo/v6/schemes/ckks/fast

GOCACHE=/private/tmp/lattigo-fast-bootstrap-noise-gocache go test ./circuits/ckks/polynomial -run '^TestFastChebyshevPower2AndPower4WithMontgomeryFixedSigmaNoise$' -count=1
ok github.com/tuneinsight/lattigo/v6/circuits/ckks/polynomial
```

The tests cover a valid LogN13 ring ciphertext at Level 12, Scale 2^60, NTT+Montgomery; they convert copies back with INTT+IMForm and verify every coefficient in each active Q-prefix row equals the original signed coefficient plus the fixed-seed sampled Gaussian residue. They also verify Level/Scale/metadata/c1/dormant limbs, strict capacity failure without mutation, existing non-Montgomery behavior, both independently injected operands, and small Fast Chebyshev power 2 and power 4 multiplication chains without calling full Bootstrap. The power tests assert completion and strict capacity evidence at the observed multiplication levels.

During test authoring, the first fixture used Q/P bit sizes not constructible by the LogN13 standard-modulus generator (61-bit Q and 62-bit P); those fixture values were corrected to valid generated moduli without changing the required Level, Scale, or representation. The initial default Go-cache access was denied by the environment; rerunning with the task-local cache succeeded. These were test-environment/fixture issues, not CKKS arithmetic failures. No other targeted test failure remains.

The four modified Secondary source/test file SHA-256 values are:

| File | SHA-256 |
| --- | --- |
| `schemes/ckks/evaluator_fast_fixed_sigma_noise.go` | `494e179cafc90f2b610a7a6867afa5f7f8c78c4ecdb51d46116a347c3a8ea4e8` |
| `schemes/ckks/evaluator_fast_fixed_sigma_noise_test.go` | `5554fc55c07bef0e640f6f4c7ba8df1b6226227f7b422db2ae38bee6f7c9b824` |
| `schemes/ckks/fast/fixed_sigma_bootstrap_noise_test.go` | `207d0a56e5828d82419d2f2dfb7ebd7b88bc5cdf0a559f1d7c83ddba0129faee` |
| `circuits/ckks/polynomial/fast_test.go` | `b1aff83ecf71be78e90c59de8fdc2f366efcc5b2d1859b6b1eceb1da10bde1fe` |

Secondary commit was pushed normally to `origin/experiment/fast-bootstrap-noise-fidelity`; `fast-qprefix` was not modified. No Bootstrap binary was built, so this repair has no new binary SHA.

## Review boundary

This repair only makes the opt-in hook representation-correct for the observed Montgomery boundary and passes the focused no-Bootstrap tests. It does not establish that fixed sigma improves final Bootstrap precision. The original triplet remains incomplete because FN has no output and its 3/3 call budget is spent. Any new Standard/Fast/Fast+Noise comparison requires a separate task and budget. Requested next action: independent Web review; do not resume Bootstrap automatically.
