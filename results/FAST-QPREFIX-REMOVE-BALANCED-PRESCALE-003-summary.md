# FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003

## Outcome

`BALANCED_PRESCALE_REMOVED`. The historical balanced generated-power workaround is deleted from the active Fast polynomial implementation. Every generated power now uses one direct Q-prefix source-level multiplication. Chebyshev applies its recurrence correction at product scale, checks the existing strict Q-prefix capacity observer before Rescale, then performs one result Rescale. The internal Monomial path performs one post-product Rescale and preserves its logical Level/Scale progression. No Standard arithmetic, parameters, key semantics, PS planner, or unrelated code changed.

The cleanup deletes the balanced factor/schedule helpers and constants, operand-side integer scaling and Rescales, balanced workspace fields/helper, and tests requiring the obsolete algorithm. Generic `copyQPrefixAtLevel` remains because it still serves the ordinary Q-prefix copy helper and its generic test. Secondary diff scope: `circuits/ckks/polynomial/fast.go` and `circuits/ckks/polynomial/fast_test.go` only.

## Provenance and workload

- Secondary `fast-qprefix`: `5feb44917fca40c93abec6def6f26bc81a82c536`, parent `75ef5dbe7bbf7d3947fb2b9fb232c4a56f05c948`, clean during measurements.
- Primary harness pin: `5feb44917fca40c93abec6def6f26bc81a82c536`.
- Measurement Primary code: `b0c58609f7f9022a509c18b716ad31b4bf791daf` plus the sole local change, the required Secondary SHA pin (`dirty=true`). The numerical frontend/config/input were otherwise unchanged.
- Each profile used 2 Fast executions and 3 fresh genuine Standard key/evaluation-key trials, the same encoded input, and the existing `1e-2` coordinate threshold. Fast repeats were deterministic. The full raw JSON and generated detailed Markdown remain in `/private/tmp`, not in the repository.
- LogN13 input SHA-256: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` (`configs/bootstrap_config.logN13.json`). LogN16 input SHA-256: `659fc59c899341a8253887270f1ae3478528fd864d66aca14bba918550ce8ccf` (`configs/bootstrap_config.logN16.json`).

## Numerical results

| Profile | Classification | Final Fast-vs-Standard complex RMSE / max | Fast Bootstrap SNR | Standard Bootstrap SNR | Fast-vs-Standard coordinates > `1e-2` | EvalMod capacity checkpoints |
|---|---|---:|---:|---:|---:|---:|
| LogN13 / q0=55 | `FAST_STANDARD_NUMERICAL_CLOSE` | `1.446809e-10` / `4.675261e-10` | `144.710750 dB` | `144.713390 dB` | `0 / 24,576` | `80/80` strict-fit |
| LogN16 / q0 target=55 (effective q0=56 bits) | `FAST_STANDARD_NUMERICAL_CLOSE` | `0` / `0` | `130.947666 dB` | `130.947666 dB` | `0 / 196,608` | `80/80` strict-fit |

Both profiles had zero capacity failures, no first observable/material divergence, and passed the public metadata contract. All generated-power Level/Scale comparisons matched genuine Standard. LogN13's largest generated-power complex RMSE was `3.567378e-16` and largest max difference was `1.110224e-15`; LogN16 values were zero.

## Generated Chebyshev powers

Each cell is `complex RMSE / max complex difference`; the paired real/imag values are separate parallel branches. Scales are `log2(Scale)`. Every row had a genuine Standard `PowerBasis.GenPower` reference and matching Level/Scale.

| Profile | Power | Level / Scale₂ | Real RMSE / max | Imag RMSE / max |
|---|---:|---:|---:|---:|
| LogN13 | T1 (polynomial input) | 12 / 60.000000 | `2.367438e-17 / 9.194034e-17` | `2.387932e-17 / 9.540979e-17` |
| LogN13 | T2 | 11 / 60.000000 | `2.818592e-17 / 2.220446e-16` | `3.208066e-17 / 2.220446e-16` |
| LogN13 | T3 (PS) | 10 / 60.000000 | `7.054116e-17 / 2.775558e-16` | `7.123580e-17 / 2.706169e-16` |
| LogN13 | T4 | 10 / 60.000000 | `5.979137e-17 / 3.330669e-16` | `6.504635e-17 / 3.330669e-16` |
| LogN13 | T6 (PS) | 9 / 60.000000 | `5.779520e-17 / 2.220446e-16` | `5.938737e-17 / 3.330669e-16` |
| LogN13 | T8 | 9 / 60.000000 | `1.160321e-16 / 3.330669e-16` | `1.243931e-16 / 4.440892e-16` |
| LogN13 | T16 | 8 / 60.000000 | `3.459806e-16 / 1.110223e-15` | `3.567378e-16 / 1.110223e-15` |
| LogN16 | T1 (polynomial input) | 12 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | T2 | 11 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | T3 (PS) | 10 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | T4 | 10 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | T6 (PS) | 9 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | T8 | 9 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | T16 | 8 / 60.000000 | `0 / 0` | `0 / 0` |

## Polynomial, DoubleAngle, and EvalMod landmarks

Each cell is `complex RMSE / max complex difference`, comparing Fast with genuine Standard at the same checkpoint. Scale is `log2(Scale)`.

| Profile | Checkpoint | Level / Scale₂ | Real RMSE / max | Imag RMSE / max |
|---|---|---:|---:|---:|
| LogN13 | polynomial output before DoubleAngle | 7 / 60.000000 | `2.074137e-16 / 6.661338e-16` | `2.141523e-16 / 7.771562e-16` |
| LogN13 | DoubleAngle round 0 after Rescale | 6 / 60.000000 | `5.992258e-16 / 1.998401e-15` | `6.041472e-16 / 2.109424e-15` |
| LogN13 | DoubleAngle round 1 after Rescale | 5 / 60.000000 | `1.377895e-15 / 4.329870e-15` | `1.391663e-15 / 4.773959e-15` |
| LogN13 | DoubleAngle round 2 after Rescale | 4 / 60.000000 | `1.554912e-15 / 4.891915e-15` | `1.567266e-15 / 5.353314e-15` |
| LogN13 | EvalMod output after public Scale reset | 4 / 45.000000 | `5.095135e-11 / 1.602983e-10` | `5.135617e-11 / 1.754174e-10` |
| LogN16 | polynomial output before DoubleAngle | 7 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | DoubleAngle round 0 after Rescale | 6 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | DoubleAngle round 1 after Rescale | 5 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | DoubleAngle round 2 after Rescale | 4 / 60.000000 | `0 / 0` | `0 / 0` |
| LogN16 | EvalMod output after public Scale reset | 4 / 45.000000 | `0 / 0` | `0 / 0` |

## Bootstrap checkpoints

Each row reports complex RMSE / maximum complex difference against genuine Standard.

| Profile | Checkpoint | Level / Scale₂ | RMSE / max |
|---|---|---:|---:|
| LogN13 | input | 0 / 45.000000 | `0 / 0` |
| LogN13 | ScaleDown | 0 / 45.000000 | `0 / 0` |
| LogN13 | ModUp | 16 / 50.000000 | `0 / 0` |
| LogN13 | C2S real | 12 / 50.000000 | `2.396436e-14 / 9.185075e-14` |
| LogN13 | C2S imag | 12 / 50.000000 | `2.409568e-14 / 9.768351e-14` |
| LogN13 | EvalMod real | 4 / 45.000000 | `5.095135e-11 / 1.602983e-10` |
| LogN13 | EvalMod imag | 4 / 45.000000 | `5.135617e-11 / 1.754174e-10` |
| LogN13 | S2C | 1 / 45.000000 | `1.446809e-10 / 4.675261e-10` |
| LogN13 | final public output | 1 / 45.000000 | `1.446809e-10 / 4.675261e-10` |
| LogN16 | input | 0 / 45.000000 | `0 / 0` |
| LogN16 | ScaleDown | 0 / 45.000000 | `0 / 0` |
| LogN16 | ModUp | 16 / 50.000000 | `0 / 0` |
| LogN16 | C2S real | 12 / 50.000000 | `0 / 0` |
| LogN16 | C2S imag | 12 / 50.000000 | `0 / 0` |
| LogN16 | EvalMod real | 4 / 45.000000 | `0 / 0` |
| LogN16 | EvalMod imag | 4 / 45.000000 | `0 / 0` |
| LogN16 | S2C | 1 / 45.000000 | `0 / 0` |
| LogN16 | final public output | 1 / 45.000000 | `0 / 0` |

## Tests and commands

Passed:

- `go test ./circuits/ckks/polynomial -run '^(TestFastPolynomialNonPowerOfTwoChebyshevOracle|TestFastPolynomialChebyshevLazyMetadata|TestChebyshevGeneratedPowerOrderIndependentOfLogNAndQPrefix|TestFastMonomialGeneratedPowerPostProductRescaleAndSourceImmutability|TestFastGeneratedPowerIgnoresDormantResidues)$' -count=1`
- `go test ./circuits/ckks/polynomial ./circuits/ckks/mod1 ./circuits/ckks/bootstrapping ./schemes/ckks/fast -count=1`
- `go test ./cmd/fastdiag ./internal/numericalmetrics -count=1`
- `git diff --check` (Primary and Secondary)
- `go run ./cmd/fastdiag numerical --profile p93-q55 --standard-trials 3 --out /private/tmp/FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003-logn13.json`
- `go run ./cmd/fastdiag numerical --profile logn16-q55 --standard-trials 3 --out /private/tmp/FAST-QPREFIX-REMOVE-BALANCED-PRESCALE-003-logn16.json`

No timing campaign, Standard source change, parameter tuning, fallback, or numerical-gate relaxation was performed.
