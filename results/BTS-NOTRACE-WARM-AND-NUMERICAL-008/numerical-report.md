# BTS-NOTRACE-WARM-AND-NUMERICAL-008 — measured results

Status: `TABLE_COMPLETE_AND_OUTPUTS_COMPARED`

| 環維度 | CKKS slots | 實作 | Cold (ms) | Warm 中位數 (ms) | Warm 倍率 |
|---|---:|---|---:|---:|---:|
| LogN16 (N=65,536) | 32,768 | Standard | 103327.916 | 154536.049 | 1× |
| LogN16 (N=65,536) | 32,768 | Fast | 5889.669 | 5044.745 | 30.6331× |

NewBTS initialization (separate from Bootstrap): Standard 72.000814 s; Fast 43.319619 s.

Input SHA-256: `5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373`. Warm speedup uses Standard warm median / Fast warm median.

## Decoded numerical comparison

All six same-index sample comparisons are included below; index pairing is descriptive because backend key/encryption randomness is independent. Metrics are over 32,768 real slots per sample.

| Attempt | Phase | Fast vs Standard MaxAbs / RMSE / MeanAbs | Standard vs X MaxAbs | Fast vs X MaxAbs |
|---:|---|---|---|---|
| 1 | first_cold_bootstrap | 1.15945733e-03 / 1.62793260e-04 / 1.16704845e-04 | 1.15430948e-03 | 2.52521213e-05 |
| 2 | warm_bootstrap_01 | 1.15945733e-03 / 1.62793260e-04 / 1.16704845e-04 | 1.15430948e-03 | 2.52521213e-05 |
| 3 | warm_bootstrap_02 | 1.15945733e-03 / 1.62793260e-04 / 1.16704845e-04 | 1.15430948e-03 | 2.52521213e-05 |
| 4 | warm_bootstrap_03 | 1.15945733e-03 / 1.62793260e-04 / 1.16704845e-04 | 1.15430948e-03 | 2.52521213e-05 |
| 5 | warm_bootstrap_04 | 1.15945733e-03 / 1.62793260e-04 / 1.16704845e-04 | 1.15430948e-03 | 2.52521213e-05 |
| 6 | warm_bootstrap_05 | 1.15945733e-03 / 1.62793260e-04 / 1.16704845e-04 | 1.15430948e-03 | 2.52521213e-05 |

### Aggregate over the six recorded sample vectors

- Fast vs Standard: MaxAbs 1.15945733e-03 (log10 -2.935745; 10^-3), RMSE 1.62793260e-04 (log10 -3.788364; 10^-4), MeanAbs 1.16704845e-04 (log10 -3.932911; 10^-4), worst slot 27650 of the concatenated six-vector set.
- Standard vs X: MaxAbs 1.15430948e-03 (log10 -2.937678; 10^-3), RMSE 1.62755050e-04 (log10 -3.788466; 10^-4), MeanAbs 1.16583762e-04 (log10 -3.933362; 10^-4).
- Fast vs X: MaxAbs 2.52521213e-05 (log10 -4.597702; 10^-5), RMSE 5.70847687e-06 (log10 -5.243480; 10^-6), MeanAbs 4.53363389e-06 (log10 -5.343554; 10^-6).

Fast plaintext Level 3 is a Q-prefix projection of the logical Level 18 output; Standard plaintext is native Level 18. Fast remains intentionally insecure zero-secret simulation; numerical similarity is not security equivalence.
