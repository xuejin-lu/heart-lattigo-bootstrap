# FAST-QPREFIX-CHEBYSHEV-ORDER-002 結果摘要

分類：`QPREFIX_CHEBYSHEV_ORDER_RESTORED`

## 修復與範圍

Secondary `fast-qprefix` 從基線 `5117fc57949647182f476dc5952099c706b9f869` 修復並 push 至 `75ef5dbe7bbf7d3947fb2b9fb232c4a56f05c948`。修改僅限：

- `circuits/ckks/polynomial/fast.go`
- `circuits/ckks/polynomial/fast_test.go`

Chebyshev generated powers 現在不依 LogN、q0 bit-size 或 Q-prefix 寬度分流，統一執行：原 operands 在共同 logical level 相乘 → 於 product scale 執行 doubling / recurrence correction → 檢查 pre-Rescale strict capacity → 對 corrected result Rescale 一次。舊 LogN13 schedule selector 已移除；Chebyshev 路徑不再 operand-side integer scaling 或分別預先 Rescale。`QPrefixWidth(Level)=min(Level+1,4)` 未變，Level ≥ 3 才使用 Q0123，較低 level 自然收縮。非 Chebyshev 通用路徑未改。

Standard production arithmetic、Fast key/error-retaining semantics、參數、Q/P chain 與 workload 均未改。Primary `cmd/fastdiag/numerical.go` 另有一行 pinned Secondary SHA 更新，使既有 canonical harness 能驗證本次 commit；其 clean provenance commit 為 `19c29880e6eee6859788f8ab3b6ad83c1ccd1113`。

## Canonical Fast-vs-genuine-Standard 結果

兩 profile 使用相同的 Primary harness、配置與輸入；每次各執行 2 次 Fast，並用 3 組新 Standard keys/evaluation keys。Final RMSE 為 direct decoded Fast-vs-Standard complex RMSE。

| Profile | Baseline RMSE | 本次 RMSE | Fast SNR：baseline → 本次 | Standard SNR | `1e-2` Fast-vs-Standard 座標超限 |
|---|---:|---:|---:|---:|---:|
| LogN13 `p93-q55` | `1.44680895e-10` | `1.44680895e-10` | `144.710750 → 144.710750 dB` | `144.713390 dB` | `0 / 24,576` |
| LogN16 `logn16-q55` | `2.0485498e-2` | `0` | `-0.000462 → 130.947666 dB` | `130.947666 dB` | `0 / 196,608` |

LogN16 改善後，Fast 與 Standard 在本 harness 的 public decoded output 逐座標相同；Fast-vs-original RMSE 為 `5.80819335e-9`，與 Standard 相同。LogN13 regression 維持先前精度。

## generated powers 與內部 checkpoints

下表中的 Level / Scale 是 Fast 與 genuine Standard 的共同 metadata；Scale 為 `log2(Scale)`。power RMSE/max 為 real / imag branch 各自的 Fast-vs-Standard complex 指標。PS 額外需要的 T3、T6 也一併列入。

| Checkpoint | Level / Scale₂ | LogN13 RMSE real / imag | LogN13 max real / imag | LogN16 RMSE / max（兩 branch） |
|---|---:|---:|---:|---:|
| T1 | 12 / 60 | `2.367e-17 / 2.388e-17` | `9.194e-17 / 9.541e-17` | `0 / 0` |
| T2 | 11 / ≈60 | `2.819e-17 / 3.208e-17` | `2.220e-16 / 2.220e-16` | `0 / 0` |
| T3 | 10 / ≈60 | `7.054e-17 / 7.124e-17` | `2.776e-16 / 2.706e-16` | `0 / 0` |
| T4 | 10 / ≈60 | `5.979e-17 / 6.505e-17` | `3.331e-16 / 3.331e-16` | `0 / 0` |
| T6 | 9 / ≈60 | `5.780e-17 / 5.939e-17` | `2.220e-16 / 3.331e-16` | `0 / 0` |
| T8 | 9 / ≈60 | `1.160e-16 / 1.244e-16` | `3.331e-16 / 4.441e-16` | `0 / 0` |
| T16 | 8 / ≈60 | `3.460e-16 / 3.567e-16` | `1.110e-15 / 1.110e-15` | `0 / 0` |

其餘必要 checkpoints（格式為 LogN13 RMSE/max，real / imag；LogN16 各項皆 `0 / 0`）：

| Checkpoint | Branch | Level / Scale₂ | LogN13 RMSE / max | LogN16 RMSE / max |
|---|---|---:|---:|---:|
| C2S real | real | 12 / 50 | `2.396e-14 / 9.185e-14` | `0 / 0` |
| C2S imag | imag | 12 / 50 | `2.410e-14 / 9.768e-14` | `0 / 0` |
| Polynomial output | real | 7 / ≈60 | `2.074e-16 / 6.661e-16` | `0 / 0` |
| Polynomial output | imag | 7 / ≈60 | `2.142e-16 / 7.772e-16` | `0 / 0` |
| DoubleAngle round 0 after Rescale | real | 6 / ≈60 | `5.992e-16 / 1.998e-15` | `0 / 0` |
| DoubleAngle round 0 after Rescale | imag | 6 / ≈60 | `6.041e-16 / 2.109e-15` | `0 / 0` |
| DoubleAngle round 1 after Rescale | real | 5 / ≈60 | `1.378e-15 / 4.330e-15` | `0 / 0` |
| DoubleAngle round 1 after Rescale | imag | 5 / ≈60 | `1.392e-15 / 4.774e-15` | `0 / 0` |
| DoubleAngle round 2 after Rescale | real | 4 / 60 | `1.555e-15 / 4.892e-15` | `0 / 0` |
| DoubleAngle round 2 after Rescale | imag | 4 / 60 | `1.567e-15 / 5.353e-15` | `0 / 0` |
| EvalMod output after public Scale reset | real | 4 / 45 | `5.095e-11 / 1.603e-10` | `0 / 0` |
| EvalMod output after public Scale reset | imag | 4 / 45 | `5.136e-11 / 1.754e-10` | `0 / 0` |
| S2C, combined branches | real + imag | 1 / 45 | `1.447e-10 / 4.675e-10` | `0 / 0` |
| Final public output | decoded vector | 1 / 45 | `1.447e-10 / 4.675e-10` | `0 / 0` |

Fast / Standard 的 checkpoint Level、Scale 與 public metadata 均符合 harness 比較契約。LogN16 的 T2、T4、T8、T16、polynomial output、各 DoubleAngle round、EvalMod 與 final public output RMSE/max diff 全為 0。

## Q-prefix capacity 與驗證

每個 profile 的 EvalMod audit 都有 80 個 checkpoint，strict `2B < S_Q` 失敗數均為 0。首個 T2 pre-Rescale checkpoint 均在 Level 12、4 rows（Q0123）；逐項 exact evidence：

| Profile | `max_abs_per_component`（c0, c1）| Exact `S_Q` | Strict fit |
|---|---|---|---|
| LogN13 | `[1328578958637108797142441602194643596, 0]` | `5986308565615587353347023386369277282933624144412673` | true |
| LogN16 | `[1328578958671072390198909220938889744, 0]` | `5986249335185303576695453761753997162624003383361537` | true |

Focused unit tests 另確認 Level 2 使用 3 rows、Level 3 使用 4 rows；post-Rescale 後 row count 按 logical Level 自然收縮。沒有容量 workaround 或 fallback。

執行並通過：

- Secondary focused generated-power tests，含 LogN13 / LogN16、Q012 / Q0123、不同 source Level / Scale，以及 T2/T3/T5 recurrence oracle。
- `go test ./schemes/ckks/fast ./circuits/ckks/polynomial ./circuits/ckks/bootstrapping -count=1`
- `go test ./circuits/ckks/mod1 -count=1`
- Primary `go test ./cmd/fastdiag ./internal/numericalmetrics -count=1`
- `git diff --check`（兩 repository）
- 既有 Primary canonical harness：
  - `go run ./cmd/fastdiag numerical --profile p93-q55 --standard-trials 3 --out /private/tmp/FAST-QPREFIX-CHEBYSHEV-ORDER-002-logn13.json`
  - `go run ./cmd/fastdiag numerical --profile logn16-q55 --standard-trials 3 --out /private/tmp/FAST-QPREFIX-CHEBYSHEV-ORDER-002-logn16.json`

Numerical provenance：Primary `19c29880e6eee6859788f8ab3b6ad83c1ccd1113` (`main`, clean)；Secondary `75ef5dbe7bbf7d3947fb2b9fb232c4a56f05c948` (`fast-qprefix`, clean)。兩次 profile 均分類 `FAST_STANDARD_NUMERICAL_CLOSE`。未執行 timing campaign，未更動 Standard、Bootstrap workload、參數或歷史報告。
