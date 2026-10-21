# FIXED-SIGMA-MUL-NOISE-001 — 實驗摘要

**目前狀態：`CP_B_COMPLETE_CP_C_IN_PROGRESS`。** 真實 Bootstrap 呼叫 `0`；LogN16、大型 benchmark 與 BTS 均未執行。

- 正式規格：[`specs/FIXED-SIGMA-MUL-NOISE-001.md`](../../specs/FIXED-SIGMA-MUL-NOISE-001.md)
- 執行紀錄與 CP-A source map：[`journal.md`](journal.md)
- 完整 JSON/vector 檔案的大小與 SHA-256：[`manifest.json`](manifest.json)
- CP-A—CP-B Primary source commit：`b99f63710718ac4373e704afef3023d62fc17f3d`
- CP-B genuine Standard：`5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- CP-B original Fast：`ee308c9743000b5f281a84a871e7c4bd59b176c4`
- 有效設定：E=32、LogN13、4096 slots、N=8192、Level5→4、Scale `2^45`；SK 對照，PK 不支援。
- Input SHA-256：`d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`
- Workload SHA-256：`00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74`
- Q/P SHA-256：`1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`

`r_op = Decode(Rescale(Relin(Mul(ct_x,ct_y)))) − Decode(ct_x) ⊙ Decode(ct_y)`。輸出相對明文 oracle 的總誤差另行記錄；表中標準差與 RMS 均指單次操作的 decoded incremental residual，而非總 RMSE。

| CP-B 統計（3 組） | Genuine Standard | Original Fast | Fast + fixed-σ |
|---|---:|---:|---:|
| Incremental mean (real, imag) | 每組約 `1e-13` | `(1.84e-17, 9.12e-17)` | 尚未執行 |
| Incremental variance | `1.0541e-22`—`1.0888e-22` | `5.5313e-25` | 尚未執行 |
| Incremental RMS | `1.0268e-11`—`1.0435e-11` | `7.4373e-13` | 尚未執行 |
| Incremental Max | `4.6052e-11`—`8.2184e-11` | `2.2054e-12` | 尚未執行 |
| Total oracle-error RMS | `1.0274e-11`—`1.0443e-11` | `7.4421e-13` | 尚未執行 |
| Total oracle-error Max | `4.5994e-11`—`8.2343e-11` | `2.1986e-12` | 尚未執行 |
| Coefficient σ / empirical RMS | 不適用 | 不適用 | 尚未執行 |
| Injection count / overhead | `0 / 不適用` | `0 / 不適用` | 尚未執行 |

校準只使用 Pair 01 的非負 slot-variance delta，得到候選並凍結 `sigma=4.004280012097432`。依 CKKS slot IFFT/解碼正規化，獨立零均值係數 Noise 的一階預期為 `Var(slot)=N·Var(eta)/Scale²`，其中 `N=8192`、輸出 Scale=`2^45`。Pair 02/03 是 held-out：各自倒推 `sigma=3.9806088657467544` 與 `4.04591114352061`，未用於調參。此推導假設獨立、零均值且同分布；實際 rounded sampler 統計與解碼注入前後 delta 必須在 CP-C 直接核對。

CP-B 的 `compare-public` 三組皆通過既有 1e-6 明文 oracle checkpoint gate，且比較器從 sidecar 重算了增量殘差。這個較寬的 workload gate 不代表 Noise 假設已成立；Standard residual 仍可能混合 Relin/KeySwitch、Rescale rounding 與 encoding/decode 效應。Fast 零秘密 `c1=0` 不代表安全性，也不等同 Standard arithmetic。結果僅屬 `DIAGNOSTIC_ONLY`。

## CP-C 狀態

Secondary helper 與回歸測試已提交至 `fast-qprefix` commit `9ef0315feaf59dbee03bf0852b72c76fdeb6b800` 並已推送。Helper 預設不執行，只接受 LogN13、Level4、Scale `2^45`、四列 Q-prefix、NTT 非 Montgomery 且 `c1=0`；同一個 signed Gaussian 整數向量逐列約化、各自 NTT，對 c0 候選加法後以既有 centered-Q-prefix lift 加 signed vector 檢查容量並核對四列，通過才提交 c0。σ=0 為不觸碰 ciphertext 或 sampler 的真 no-op。

CP-C 的 matched Standard/Fast+Noise 測量、實際 decoded delta、三組 held-out 結果、注入 overhead、binary/source SHA 與最終分類仍待完成。若經驗 delta 不支持固定 σ 模型，需如實標記 `UNMODELLED` 或 `RMSE_ONLY`，不能改 σ。
