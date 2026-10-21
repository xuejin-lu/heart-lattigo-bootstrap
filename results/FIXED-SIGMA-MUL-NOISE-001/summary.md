# FIXED-SIGMA-MUL-NOISE-001 — 實驗報告

**狀態：`PLANNED_NOT_RUN`。本文件是版本化結果入口，**不是**完成的實驗結果。**

- 正式研究規格：[`specs/FIXED-SIGMA-MUL-NOISE-001.md`](../../specs/FIXED-SIGMA-MUL-NOISE-001.md)
- 任務指標：[`CURRENT_TASK.md`](../../CURRENT_TASK.md)
- 耐久 journal：[`journal.md`](journal.md)
- 編譯／測試／數值結果：**尚未執行；NOT_RUN**
- 候選 σ／凍結 σ：**NOT_CALIBRATED**
- 樣本數與種子：**NOT_RUN**
- 真實 Bootstrap：0 / 0 授權；目前 0 次。
- 安全性定位：`DIAGNOSTIC_ONLY`，不恢復 RLWE 安全性。

## 預定比較表（不得把佔位符當測量值）

| 指標 | Genuine Standard | Original Fast | Fast + fixed-σ |
|---|---|---|---|
| Input/workload fingerprint | NOT_RUN | NOT_RUN | NOT_RUN |
| Effective LogN, Level, Scale | NOT_RUN | NOT_RUN | NOT_RUN |
| 增量 mean | NOT_RUN | NOT_RUN | NOT_RUN |
| 增量 variance | NOT_RUN | NOT_RUN | NOT_RUN |
| 增量 RMSE | NOT_RUN | NOT_RUN | NOT_RUN |
| 增量 Max Error | NOT_RUN | NOT_RUN | NOT_RUN |
| 輸出總 RMSE | NOT_RUN | NOT_RUN | NOT_RUN |
| 輸出總 Max Error | NOT_RUN | NOT_RUN | NOT_RUN |
| 係數 σ、實際 Noise RMS | 不適用 | 不適用 | NOT_CALIBRATED |
| 明確 Opt-in／真正 Fast dispatch | 不適用 | NOT_RUN | NOT_RUN |
| 與未見樣本一致性 | NOT_RUN | NOT_RUN | NOT_RUN |

Codex 實際工作時於本檔填入結果、關鍵 provenance、失敗／否證與限制；詳細原始證據請放此目錄的界限內，不建立另一套 runner。
