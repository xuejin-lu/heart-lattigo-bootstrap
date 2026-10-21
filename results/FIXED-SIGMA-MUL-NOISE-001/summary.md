# FIXED-SIGMA-MUL-NOISE-001 — 實驗摘要

**狀態：`BATCH_COMPLETE_READY_FOR_WEB_REVIEW`；CP-C 結果分類：`RMSE_ONLY`。** 全批 Standard/Fast Bootstrap 呼叫 `0` 次；未執行 LogN16、大型 benchmark 或 BTS。

- 正式規格：[`specs/FIXED-SIGMA-MUL-NOISE-001.md`](../../specs/FIXED-SIGMA-MUL-NOISE-001.md)
- 執行紀錄及 CP-A source map／理論限制：[`journal.md`](journal.md)
- 完整 JSON/vector 檔案大小與 SHA-256：[`manifest.json`](manifest.json)
- Genuine 原始 Standard：`5dbffbdea05394de2ca3a432ed5318aa832e3f40`，未修改。
- CP-B 原始 Fast baseline：`ee308c9743000b5f281a84a871e7c4bd59b176c4`。
- CP-C Fast fixed-σ helper：Secondary `fast-qprefix@9ef0315feaf59dbee03bf0852b72c76fdeb6b800`，已推送。
- CP-C Primary harness：`a47e402a31e9181ac94dfb430e1cc337fbd8b3a0`；pair-01 證據 checkpoint `e543cf8a9aac5199ebc849e044328cbfac835d3d`。所有 CP-C runs 的 measurement source SHA-256 相同：`79b13e0978030f575458230d7068f572976ae29b3ebeed13f3a9fbc66fa8d8e5`。
- CP-C Standard/Fast+Noise binary SHA-256：`7521b4e33dfeeb80c266d5526d82845f1cf54c619b61dfe151bcd73af050ea76` / `8e6eb3934ed5bb545524344654314cdf7fbff67928fdd5268b18d75ad39153d8`。
- 執行環境：Go `go1.26.4`、`darwin/arm64`；E=32、LogN13、4096 slots、N=8192、Level5→4、輸入 Scale `2^45×q5`、輸出 Scale `2^45`、SK cohort。PK unsupported，沒有將 SK 冒充 PK。

## 誤差定義與 CP-A

`r_op = Decode(Rescale(Relin(Mul(ct_x,ct_y)))) − Decode(ct_x) ⊙ Decode(ct_y)`。以同一輸入的 decoded product 扣除既有輸入誤差傳播；對原始明文 oracle 的總誤差另存，避免用最終 RMSE 取代單次操作新增誤差。

CP-A 確認 ordinary `ckks.Evaluator.MulRelinNew` 會 dispatch 到 Fast `MulCore.ApplyRows`，ordinary `Rescale` 會 dispatch 到共用 `RescaleCore.ApplyRows`。本次只在隔離的完整 MulRelin+Rescale 後由 perfprobe Fast build 呼叫一次 helper；沒有任意攔截 Rescale，也沒有聲稱涵蓋 Bootstrap、lazy/deferred 路徑或乘法鏈全部位置。q0..q3 是 Level4 維護的四列，q5 為 Rescale 除數；注入保留原 Scale，且須對實際 signed coefficient lift 滿足 centered Q-prefix capacity。理論主導項、Scale、Q-prefix／NTT 表示限制詳見 journal。

## CP-B 校準與未見基線

三組 genuine Standard 與未修改 Fast 的單次鏈皆為零 Bootstrap。候選只由 CP-B Pair 01 校準；Pairs 02/03 作 held-out。依 CP-A 的實際 encoder/decoder 正規化，若係數誤差獨立且零均值，`Var(slot)=N·Var(eta)/Scale²`，因此得到並凍結 `sigma_candidate=4.004280012097432`。CP-B held-out 倒推為 `3.9806088657467544` 與 `4.04591114352061`，沒有拿來調參。完整三組結果見 journal 與 manifest。

## CP-C 三欄比較

CP-C Fast baseline 數字是同一個 Fast+Noise run 的**注入前快照**；故三欄用同一 Primary source SHA、Fast helper binary 及工作負載比較。每個值是單一 4096-slot vector 的 moments；不是把 slot 當獨立重複實驗。

| 指標 | Genuine Standard（CP-C） | Fast 原始輸出（注入前） | Fast + fixed σ（CP-C） |
|---|---:|---:|---:|
| `r_op` variance | `1.0544e-22`–`1.0692e-22` | `5.5313e-25`（三次相同） | `1.0567e-22`–`1.0882e-22` |
| `r_op` RMS | `1.0270e-11`–`1.0341e-11` | `7.4373e-13` | `1.0279e-11`–`1.0434e-11` |
| `r_op` Max | `4.5395e-11`–`5.5430e-11` | `2.2054e-12` | `2.9713e-11`–`3.3719e-11` |
| 明文 oracle total-error RMS | `1.0267e-11`–`1.0342e-11` | `7.4421e-13`（CP-B baseline） | `1.0279e-11`–`1.0435e-11` |
| 明文 oracle total-error Max | `4.5456e-11`–`5.5362e-11` | `2.1986e-12`（CP-B baseline） | `2.9697e-11`–`3.3712e-11` |
| 係數域 σ / empirical RMS | 不適用 | 不適用 | 固定 `4.004280012097432` / `3.9887`–`4.0413` |
| 注入數 / 注入 overhead | `0` / 不適用 | `0` / 不適用 | 每次恰好 `1` / `1.305`–`2.716 ms` |

| Pair / seed | Standard variance − Fast+Noise variance | RMS 相對 Standard | 注入 delta variance vs `N·Var(eta)/2^90` | 注入 overhead / Mul+Rescale 時間 |
|---|---:|---:|---:|---:|
| 01 / `20261022`（transfer check） | `−2.2137e-24`（negative） | `+1.047%` | `1.070440e-22` vs `1.070641e-22`（−0.019%） | `1.305 / 2.269 ms` |
| 02 / `20261023`（held out） | `+1.2545e-24` | `−0.598%` | `1.052830e-22` vs `1.052771e-22`（+0.006%） | `2.063 / 3.532 ms` |
| 03 / `20261024`（held out） | `−3.0089e-24`（negative） | `+1.423%` | `1.080246e-22` vs `1.080080e-22`（+0.015%） | `2.716 / 5.365 ms` |

實際注入前後 decoded delta 與係數域抽樣映射相符；q0..q3 四列的 8192 個 signed coefficients 都通過 centered-prefix strict-fit。Standard−Fast+Noise 的 paired variance delta 有兩負一正，因此整體分類是 `RMSE_ONLY`：觀察到增量 RMS 接近（平均 Fast+Noise `1.03635e-11`，Standard `1.02995e-11`，約高 `0.62%`），但方差並未逐組一致靠攏。不能宣稱一般獨立 Gaussian 模型或分布等價，也沒有變更 σ。各 pair 的跨 backend correlation magnitude 為 `0.00983`–`0.01586`。單次注入 overhead 約為 MulRelin+Rescale 時間的 `50.6%`–`58.4%`；這是診斷樣本，並非 benchmark campaign。

## 驗證與限制

- 固定 σ opt-in 預設關閉，僅限 LogN13、E32、零 Bootstrap、指定 Fast helper SHA；σ=0 測試為真正 no-op。相同 signed integer polynomial 逐列映射到 q0..q3，NTT 後只加到 c0；c1、metadata、Level、Scale、dormant limbs 維持不變。helper 候選提交為 Secondary `9ef0315...`，其實作細節與單元測試列在 CP-C journal。
- Standard/Fast 專項 `perfprobe` tests 通過；完整 Fast `cmd/perfprobe` suite 通過；Secondary fixed-sigma 與 `internal/fastcore` tests 通過。完整 Standard `cmd/perfprobe` suite 仍有已記錄且與本任務無關的歷史 A=0 測試失敗，以及 Standard modfile 下的相對 build-info path 測試失敗；本任務 targeted Standard tests 通過。task worktree policy 的精確路徑檢查及兩項單元測試通過。
- 全批 Bootstrap budget `0/0`；未執行 LogN16、大型 benchmark、BTS；multiparty 路徑未修改。舊 `experiment/standard-a-zero-noise-preserved` 未用作 Standard 或本次證據。
- 所有結果為 `DIAGNOSTIC_ONLY`，Fast zero-secret 路徑不具 RLWE 安全性；沒有恢復 `a`、沒有改 Standard arithmetic，也不宣稱任何 Bootstrap 覆蓋。
