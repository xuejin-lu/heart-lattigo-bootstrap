# BTS-FAST-NOISE-FIDELITY-009

## 核心任務

請在我們已經成功執行的**組員原始 BTS Bootstrapping 專案**上，完成 Standard、Fast OFF、Fast＋Noise ON 三組數值精度與效能比較。

**最高原則：不要重新實作或修復已經成功的 BTS／FAST 整合。只補上目前缺少的 Noise Hook 接線，然後重用既有量測平台完成實驗。**

請自主完成必要的程式修改、定向測試、零 Bootstrap 預檢、正式實驗與 GitHub 證據版本化。只有遇到真正的新數學／容量／表示問題才停止交由 Web Review。

## 一、既有研究成果：不要重新造輪子

Primary：
`xuejin-lu/heart-lattigo-bootstrap`

組員原始 BTS：
`external/BTS/`

原始來源 manifest：
`53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`

舊 BTS 成功基線：
`BTS-NOTRACE-WARM-AND-NUMERICAL-008`

舊版 FAST：
`ee308c9743000b5f281a84a871e7c4bd59b176c4`

目前通用 Noise FAST：
`experiment/fast-bootstrap-noise-fidelity@325e501709c42ab5995bf7f36fb8d19621d34131`

Genuine Standard：
`5dbffbdea05394de2ca3a432ed5318aa832e3f40`

GitHub 已確認目前 Noise FAST 是舊 BTS 成功版的直接後代，比舊版前進 8 個 commits、沒有落後 commits，而且 public `schemes/ckks/evaluator.go` 在兩版本間沒有修改。

因此 BTS 曾經需要的 Fast public API、CoeffToSlots、Q-prefix、Mul、MulRelin、Rescale 等相容性修復都已經存在。

**不要重新開發這些已經完成的功能。**

請先確認 Git 狀態及來源 pin，沿用歷史 BTS-008 的量測接入，不修改原始 `external/BTS/`，不修改 `fast-qprefix`。

## 二、第一階段：最小 Noise Hook 接線

已知問題：

- 組員 BTS 使用 `ckks.NewEvaluator` 建構的 public CKKS evaluator。
- 目前自動 Noise Hook 主要掛在 `schemes/ckks/fast.Evaluator`。
- BTS 現有的 Fast ct×ct 乘法已經能正確執行，但不會自動觸發上述 Noise Hook。

請只修正這個接線問題。

### 實作要求

1. 在現有 Fast zero-secret public CKKS ct×ct 乘法路徑上接入可明確啟用／關閉的 Noise Hook。
2. 必須支援 `Mul`、`MulNew`、`MulRelin`、`MulRelinNew`，但不得讓委派路徑造成重複注入。
3. 保留非 relinearized `Mul` 的 degree、Scale、Level、輸出與後續運算語意。
4. 重用已存在的 Gaussian sampler、NTT/Montgomery 轉換、signed coefficient addition、Q-prefix capacity checker、seed/SHA 和事件記錄。
5. 每個有效 ct×ct 乘法，對左右獨立 operand 副本分別注入 Noise；原始 operand 不可被污染。
6. Noise OFF 必須保持原本 Fast 的計算語意，且不得產生 Noise 事件。
7. ct×pt、scalar、rotation 等非目標操作不得注入。
8. 保留 strict capacity check；任何不安全或不支援的情況都必須明確拒絕，不可靜默略過。

注意 Go package 的相依方向：`ckks` 不應為了呼叫 Hook 而錯誤地反向依賴 `ckks/fast`。請優先共用現有 parent CKKS helper 或底層運算元件，避免複製 Noise 演算法。

不要重新寫乘法核心。

### 通用性要求

不要再次硬編碼 LogN=16、MaxLevel=29 或 Scale=2^32 到 Library Hook。

Hook 只依賴合法 CKKS 參數、Fast zero-secret capability、密文表示、degree、Q-prefix backing 和嚴格容量條件。

實驗專屬參數留在 Primary 的量測控制層。

## 三、第二階段前置驗證：零 Bootstrap

正式實驗採用組員 BTS 原始設定：

| 參數 | 數值 |
|---|---|
| LogN | 16 |
| N | 65536 |
| Slots | 32768 |
| 最高 Level | 29 |
| Q primes | 30 |
| P primes | 5 |
| q0 bits | 42 |
| 輸入 Level / Scale | L3 / 2^32 |
| Mod1 起始 Level / Scale | L26 / 2^42 |
| Mod1 Chebyshev degree | 31 |
| PS depth | 5 |
| Double-angle | 3 |
| 輸出 Level / Scale | L18 / 2^32 |

完整 Q/P 值以原始 BTS 及 BTS-008 保存的有效參數為準，不可重新生成另一套參數。

請使用原始 BTS 真實 Q/P、Level 和 Scale 做定向測試。

重點是確認：

- Chebyshev power basis 的 ct×ct `MulRelinNew`。
- PS giant-step 的非 relinearized ct×ct `Mul`。
- Double-angle 的 ct×ct `MulRelin`。
- 實際 BTS 所用的其他 ct×ct 路徑。

以上都必須正確觸發 Hook，且事件數與實際數學乘法次數一致。

請執行 N13～N16 相關回歸，確認目前已完成的通用化能力沒有退化。

若無法在不執行完整 Bootstrap 的前提下驗證某些事實，請明確區分「靜態可證明」、「定向測試已通過」及「仍需正式 Bootstrap 才能確認」。

**零呼叫預檢通過、程式提交並固定 Binary SHA 之前，禁止開始正式 Bootstrap。**

## 四、第三階段：正式三組 Bootstrap 實驗

上述全部通過後，啟動新的獨立實驗預算：

**最多 18 次完整 Bootstrap。**

| 模式 | Cold | Warm | 合計 |
|---|---:|---:|---:|
| Standard | 1 | 5 | 6 |
| Fast OFF | 1 | 5 | 6 |
| Fast＋Noise | 1 | 5 | 6 |
| 總計 | 3 | 15 | 18 |

此為新任務 009 的獨立額度，不得借用或修改 BTS-008、002、003 已完成的預算。

每組一次 Cold、五次 Warm，使用同一 Evaluator 生命週期，保留完整呼叫順序與逐次證據。

不得隱藏 warmup、重試或額外執行 Bootstrap。

Fast OFF 與 Fast＋Noise 必須由**同一個 Fast Binary** 執行，僅以 Noise ON/OFF 設定區分。

Noise 先使用已經在 002／003 實驗採用的：

- σ=`4.004280012097432`
- root seed=`20261027`

**不要預先調整 σ。**

如果 BTS 高 Level 的 Noise 容量檢查失敗，立即保存完整證據並停止。不能為了得到漂亮的 Precision bits 而降低輸入、改 Scale、改 σ、修改 Q/P 或放寬檢查。

如果某個 arm 失敗，不得宣稱三組比較已完成；不能用歷史 BTS-008 的數據冒充本次新量測。

## 五、量測與正式報告

重用：

- `cmd/perfprobe`
- 既有 BTS measurement overlay
- `internal/perfmeasure`
- `internal/numericalmetrics`
- 現有 Cold/Warm repeatability、attempt ledger 和 comparator

不要創造第二套量測平台。

每個 arm 記錄：

1. 最高 Level、Q/P 鏈、輸入／輸出 Level 和 Scale。
2. Cold 時間與五次 Warm 原始時間。
3. Warm 平均、中位、樣本標準差。
4. Cold／Warm allocated bytes 和 allocation counts。
5. 與相同原始 BTS plaintext 的 RMSE。
6. Global Precision bits。
7. MaxAbs、p99、p99.9。
8. Noise 事件數、各 Level、seed、SHA、capacity。
9. Binary SHA、Git commits、input/config fingerprint。

Precision bits 必須使用：

\[
\mathrm{RMSE}=\sqrt{\frac{1}{S}\sum_{i=0}^{S-1}|\widehat m_i-m_i|^2}
\]

\[
\mathrm{PrecisionBits}=-\log_2(\mathrm{RMSE})
\]

請明確區分全域 RMSE precision 和逐 slot precision median。

最終輸出 Standard、Fast、Fast＋Noise 三組比較表，另外與 BTS-008 歷史結果進行版本標記明確的比較。

BTS-008 歷史 Warm 中位：

- Standard：154536.049 ms
- Fast：5044.745 ms

這只是歷史基線，不能冒充本次結果。

## 六、版本控制與停止條件

請先建立獨立 009 spec、ledger、journal，然後在既有 Secondary 專用實驗分支上做最小修正。

整合成功後重新 pin Fast commit，固定 binaries，完成零 Bootstrap 預檢並提交 checkpoint，再開始正式 18 次量測。

不修改原始 BTS、不覆寫舊實驗、不修改 `fast-qprefix`。

所有巨大 CSV／JSON／slot vectors 必須先檢查大小，只選擇必要欄位讀取；優先使用精簡 summary、manifest 與 SHA。

正常 Commit／Push 到核准分支，並確認工作樹乾淨、遠端一致。

若完成全部實驗與 Comparator：

`BATCH_COMPLETE_READY_FOR_WEB_REVIEW`

若遇到真正需要新數學／表示／容量設計才能解決的問題：

`BATCH_BLOCKED_NEEDS_WEB_REVIEW`

**請自主執行，不必在一般編譯、測試或版本化步驟反覆詢問我；但不得擴大實驗預算或修改已完成的研究歷史。**

最重要的是：**FAST 跑完整個 BTS 的能力早已驗證成功。這次任務只是在成熟的 FAST BTS 基礎上接入既有 Noise Hook，然後量測三組數值精度與速度。**