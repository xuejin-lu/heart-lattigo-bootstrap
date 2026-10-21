# FIXED-SIGMA-MUL-NOISE-001 — 固定 σ 的乘法增量 Noise 補償研究

**狀態：** `AUTHORIZED_AUTONOMOUS_BATCH`（由使用者 2026-10-21 授權建立 GitHub 任務與研究紀錄；Codex 需使用者另行輸入「開始」才執行）
**類別：** M（理論與架構契約，GPT Web 已定義）＋ E（量測）＋限縮 I（實驗性實作）
**Primary：** `xuejin-lu/heart-lattigo-bootstrap` / `main`
**Secondary：** `xuejin-lu/lattigo` / `fast-qprefix`
**本次規格初始依據：** Primary `4476da0fad330205d7333d25b23e3ebccebfe2bb`；Secondary `ee308c9743000b5f281a84a871e7c4bd59b176c4`；原始 Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`。實際執行時另記同步後的正確 SHA。禁止以後續變更的 SHA 假裝是上述初始版本。

## 1. 使用者決策與不可竄改的原始假設

本研究追求 **數值誤差模擬**，不是恢復被歸零的 `a`、還原標準 RLWE 安全性或修改 Standard 密碼學。使用者明確否定「按原始 a 大小補 Noise」的舊假設。

**已確定：**
- Q1=C：量測「每一次 ct×ct → Relinearization → Rescale **新引入**的數值誤差」，不是把運算結束的總誤差當新誤差。
- Q2=A：在一套固定有效 CKKS 參數下，整個乘法鏈共用**一個固定的係數域 σ**；每次邏輯操作重新取樣。不按層數、Level、Scale 或輸入另行調參。跨不同參數組不得暗示同一 σ 已經證實普適。
- 理論先以一階主導項估計數值量級，Standard 實測供校準與驗證；不要求完整安全 Noise bound。
- 第一版在**完整 Mul + Relin + Rescale 結束後**，將抽樣的多項式加入 Fast `c0`；不改 `c1`、Secret Key、Standard arithmetic。
- 僅限 ct×ct；排除 ct×pt、純常數乘法、加法、Rotate，以及不涉及目標乘法鏈的單獨 Rescale。
- 所有補償只存在於 **Fast experimental opt-in**，預設關閉；正式 application/frontend 仍使用相同公開 CKKS API 和運算順序，禁止 Fast/Standard runtime selector。
- 原始 Standard、原始 Fast baseline、Fast+Noise 必須有可核對的獨立 source/build provenance。Fast 模式不具 RLWE 安全性；所有結果標記 `DIAGNOSTIC_ONLY`。

## 2. 理論與可執行觀測定義

對同一 backend 的兩個輸入密文 `ct_x, ct_y` 和一次完整乘法結果 `ct_z`，以同一個 native decrypt/decode 方法定義：

```text
x_tilde = Decode(Decrypt(ct_x))
y_tilde = Decode(Decrypt(ct_y))
z_tilde = Decode(Decrypt(Rescale(Relin(Mul(ct_x, ct_y)))))
r_op    = z_tilde - (x_tilde ⊙ y_tilde)
```

`r_op` 是這段操作的**輸出增量殘差（decoded slot domain）**，不是 KeySwitch 自身的純 noise。必須保留獨立明文 oracle 的輸出總誤差作為交叉檢查，避免既有輸入誤差混入「本次新誤差」定義。

理論工作：記錄主導項（兩個既有輸入誤差的一階傳播、Relin/KeySwitch、Rescale rounding）；忽略二階項時必須說明條件。**量測沒有證明其中哪一項占主導前，不能宣稱 `r_op` 完全來自 Relinearization。**

固定 σ 定義：`eta_i = round(N(0,σ²))` 的整數係數（明確記錄實際離散化與 sampling variance），同一次注入組成 `eta(X)`，加到 **Rescale 後** 的 `c0`。此 σ 是**coefficient-domain 密文整數單位**，不是 decoded slot RMSE，也不是 Lattigo 原始 `Xe=3.2`。

必須將**同一個整數多項式**一致映射到所有實際維護的 Q-prefix residues，再依 ciphertext 的 coefficient/NTT/Montgomery 表示正確轉換，保留 Level、Scale、NTT metadata、容量條件及 dormant limb 不可讀規範。禁止為不同 `q_i` 獨立抽樣，禁止直接把 `sigma` 塞入 NTT residue 當作 coefficient noise，禁止自行假設 slot RMSE = σ / Scale。若 Scale 差異造成固定 σ 失效，紀錄為模型的否證，不暗中改成動態 σ。

校準時先評估 `r_op,Standard` 和 `r_op,Fast` 的均值、變異數、RMS、Max 和分布；以 **實際注入前後的 decoded delta** 校驗係數 σ 到 decoded slot 增量的映射。僅在零均值／獨立等假設可檢查時，使用增量殘差的變異數差估計 σ；負差、相關性或顯著偏差都應回報 **UNMODELLED**，不能以平方根負數、偏置或 tuning 偽造成功。校準後 freeze 一個 σ，用未見資料測試，不得逐乘法微調。

## 3. 重用既有環境與 Git 操作

讀 Primary `AGENTS.md`、`docs/RESEARCH_ENGINEERING_WORKFLOW.md`、`docs/MEASUREMENT_PLATFORM.md`、`docs/BACKEND_TRANSPARENCY_CONTRACT.md`，完成 Primary `main` 安全 fetch／條件式 ff-only；再按 Secondary `fast-qprefix` `AGENTS.md` 安全同步，讀 `docs/FAST_QPREFIX_SPEC.md`。Codex 的現有 Primary/Secondary 工作樹不可 reset、stash、重建、切成舊 `fast-ckks` 或覆蓋未知 dirty 變更。使用指定隔離的 Standard／Fast checkout 與 Go `-modfile`（`GOWORK=off`），同一 Primary frontend 兩次獨立編譯；不要篡改已提交的 `go.mod`。

量測只可擴充 Primary `cmd/perfprobe/`、`cmd/fastdiag/`、`internal/perfmeasure/`、`internal/numericalmetrics/`；Secondary 只負責 Fast library 的真實實驗性 injection/helper、低層回歸測試、必要 hooks。**不得再造另一套 runner、統計函式、trace validator、timing framework。** 大檔先看 size，讀 bounded range。

## 4. 授權的循序 checkpoint（不需要使用者逐步搬運提示詞）

### CP-A — Read-only source audit + math preflight

1. 查核 `ckks.Evaluator.MulRelin`、`Rescale`、`RescaleTo`、Fast Q-prefix evaluator 和 `circuits/ckks/polynomial/fast.go` / bootstrapping 的真實呼叫鏈；列出哪些邏輯 ct×ct 有一對一的 Rescale，哪些是 lazy/deferred relin/rescale。不可假定有通用 `MulRelinRescale()` 出口，亦不可對每個單獨 Rescale 盲目注入。
2. 核對各路徑 c0 domain、Scale、Q-prefix、NTT/Montgomery、clean source、native SK 與 PK input 生命周期。明確指出最小 library-only opt-in 的候選注入位置和 dispatch 證據。
3. 清點現有 Primary 增量誤差／數值量測 API、單位、可復用 oracle；整理上節主導項近似及未證實假設。**此 checkpoint 不改 CKKS production code。**
4. 把簡短 source map、假設、可行性與 STOP 障礙寫入 `results/FIXED-SIGMA-MUL-NOISE-001/journal.md`。如有新的表示／數學決策或無法保證正確注入邊界，標記 `BATCH_BLOCKED_NEEDS_WEB_REVIEW` 後停止。

### CP-B — Small matched single-op baseline (ZERO Bootstrap)

在已有量測平台上加入必要的最小測試／量測 adapter，不變更 Standard 或 Fast 原本算術。優先 LogN13、固定一套有效 Q/P/Scale/ct×ct 工作負載，先執行 zero-Bootstrap oracle，再執行少量單次鏈。Standard 真正原始 Lattigo / Fast 非補償版本，採相同明文、參數及操作，記錄每次 `r_op`、總誤差、Level/Scale，輸入與結果 fingerprint、commit/binary identity。SK／PK 明確分組；如果 PK 所需的公鑰 lifecycle 不支持所選 backend，必須標記 unsupported，**不得把 SK 冒充 PK**。

統計估計一個 `sigma_candidate`，寫明係數 sigma / slot RMSE 映射未定或其實測標定方式；用小樣本確認是否存在非負、可解釋的補償目標。不能用驗證集參數反覆優化。若無法從單次差值合理辨識，停止並說明原因，不自創任意 noise 量級。

### CP-C — Minimum opt-in Fast library injection + zero-Bootstrap validation

只有 CP-A/B 沒有遇到數學或架構未決問題才執行。Secondary 實作**最小的一個實驗性 Noise helper 與必要的呼叫／測試**；除 opt-in 外原版 Fast 不變，禁止更動 Standard、Q-prefix 合法性、cryptographic key generation、其他 CKKS primitive。除非對應已經證明，這階段只支援**isolated MulRelin + Rescale**，**不許假裝整個 Bootstrap 的每個 ct×ct 已被覆蓋**。

對固定 `sigma_candidate`：確保一次完整鏈恰好注入一次、`sigma=0` 為真正 no-op、隨機 seed 可重現、相同 sampled signed coefficient 映射各 maintained residues 且表示正確；證實實際 Fast optimized kernel dispatch，且無 dormant-Q/Standard fallback。校準樣本與未見驗證樣本分離；比較 Baseline Standard、Baseline Fast、Fast+Noise 的增量與總 RMSE、mean、variance、Max Error，並量測注入 overhead。禁止安全性聲稱。

若需要做新 Q-prefix capacity、scale 或 arithmetic semantics 的設計選擇，立即停止 `NEEDS_WEB_REVIEW`，不可讓 Codex 自行發明公式或把驗證門檻改低。

### CP-D — Resume-ready Git evidence + milestone review

使用 Git 正常 commit / push 完成已授權修改，Secondary 若有庫改動則提交至 `fast-qprefix`，Primary 提交測試/結果至 `main`。每次提交前要有正確 source/test provenance、工作樹核對。將所有進度與結果記入 `results/FIXED-SIGMA-MUL-NOISE-001/journal.md`，寫出 `results/FIXED-SIGMA-MUL-NOISE-001/summary.md` 的實測表與 limit（若沒執行，保持 `NOT_RUN`）；若中斷可從 checkpoint 接續、不重跑已完成量測。

Codex 完成時回傳 `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` 或 `BATCH_BLOCKED_NEEDS_WEB_REVIEW`，GPT Web 再獨立審查證據；**未授權執行 Bootstrap**。

## 5. 實驗記錄與成功準則

成功首先是：同一 fixed σ，在未見輸入與多次操作的簡單非 Bootstrap chain 上，使 Fast 的增量誤差統計向 Standard 靠近，同時無嚴重偏差或不合理 Max；若僅 RMSE 接近而分布／原因不相符，只能標記 `RMSE_ONLY`。**不能預先宣稱一定成功。**

每筆量測需包含：
```text
task_id, checkpoint, status, timestamp_utc, backend={standard|fast|fast_noise},
primary_commit, secondary_commit_or_original_standard_pin, built_binary_sha256,
go_version, os_arch, workload_hash, parameters_hash, effective_logN,
Q_chain_or_hash, P_chain_or_hash, slots, level_in, level_out,
scale_in, scale_out, encryption_mode={sk|pk}, logical_ct_ct_mul_count,
logical_noise_injections, sampling_seed, sigma_coefficient,
coefficient_noise_empirical_rms, decoded_incremental_mean,
decoded_incremental_variance, decoded_incremental_rmse,
decoded_incremental_max_error, decoded_total_rmse, decoded_total_max_error,
time_ms, oracle_status, dispatch_evidence, limitations
```

`summary.md` 至少包含 Standard / Fast / Fast+Noise 三欄比較、候選與最終 σ、理論預期、校準／未見驗證區分、結果類別，以及否證條件。每次行為與結果及真實 commit SHA 必須留在 GitHub；**commit message 不是唯一或完整的實驗記錄**。

## 6. 預算及禁止項目

- **真實 Bootstrap：0 次（Standard 0、Fast 0）。** 任何首次完整 Bootstrap、LogN16、高成本重跑或組員 BTS 介入都必須先提交新授權規格與預算；不能挪用完成的 Batch007/008 配額。
- 只允許廉價 LogN13 單次 ct×ct/最小連乘操作，先 smoke / test，校準與未見驗證各自保存 seed、repetitions、hash；不進行無界 sweep。
- 禁止修改原始 Standard、原始 teammate `external/BTS`、已接受的 baseline 結果，禁止 runtime application Fast selector、Standard fallback、擅自降低 gate、假造 sigma 量測或 claim restored security。
- `CURRENT_TASK.md` 僅是指標，完整科學假設、原始決策與結果只存在本規格及 `results/FIXED-SIGMA-MUL-NOISE-001/`；不可只靠 Codex prompt 或 git commit message 交接。
