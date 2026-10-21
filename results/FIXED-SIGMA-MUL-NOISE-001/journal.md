# FIXED-SIGMA-MUL-NOISE-001 — Durable Experiment Journal

**建立日：** 2026-10-21 (Asia/Taipei)
**狀態：** `PLANNED_NOT_RUN`
**授權來源：** 使用者確認「每次完整 ct×ct＋Relin＋Rescale 新增誤差」，全體乘法共用一個固定 σ，要求 GitHub 版本管理；本 GitHub 提交僅為**規格與任務建置**。

## 版本控制及不可變歷史

- Primary 初始規格基底：`4476da0fad330205d7333d25b23e3ebccebfe2bb` (`main`)。
- Secondary 初始參考：`ee308c9743000b5f281a84a871e7c4bd59b176c4` (`fast-qprefix`)。
- 原始 Standard 參考：`5dbffbdea05394de2ca3a432ed5318aa832e3f40`（只讀、不可修改）。
- 最新實際 Git commit、原始碼 SHAs、binary SHA 與 dirty/clean 狀態：**執行後填寫**。
- 歷史 BTS Batch008 結果不可修改，不能重算或借用其 Bootstrap 配額。

## 研究決策

1. 舊方案「依被清零之 a 大小直接加回 Noise」：**SUPERSEDED / REJECTED**。
2. 被採用：在相同參數與 Scale 條件下量測每次 `Mul + Relin + Rescale` **新增**誤差；以主導項理論與 Standard 實測校準。
3. 被採用：只在 Fast (c_0) 注入，單一 coefficient-domain σ 在此參數組的**所有** ct×ct 完整鏈共用，逐次獨立抽樣。
4. 暫未證實：Standard vs Fast 增量誤差的差值可由此類獨立 Gaussian noise 表示；待數值實驗驗證。
5. 不聲稱新的安全性，不把 `DIAGNOSTIC_ONLY` 誤當 formal Standard baseline。

## 執行記錄（新增資料時保留既有紀錄）

| Checkpoint | 狀態 | 備註 |
|---|---|---|
| CP-A Source/maths audit | NOT_STARTED | source map／reuse-first |
| CP-B Standard/Fast single-op | NOT_STARTED | Bootstrap 0 calls |
| CP-C Opt-in Fast helper | NOT_STARTED | 只有 A/B 成功且無未決數學問題才能實作 |
| CP-D Summary & review handoff | NOT_STARTED | 記錄 commits、證據與結果 |

## Expensive call budget

| 類型 | 授權 | 已用 | 保留 |
|---|---:|---:|---:|
| Standard Bootstrap | 0 | 0 | 0 |
| Fast Bootstrap | 0 | 0 | 0 |

尚未實作、尚未量測、沒有校準出 σ。**不可填造結果。**

## 中斷／續跑原則

新的 Codex session 先檢查 Primary `CURRENT_TASK.md`、spec、此 journal 與本機所有既有 checkpoint；依序恢復，不能因 session 更新而重跑高成本工作、丟棄工作樹變更或重置預算。數學／架構問題寫清阻塞與建議，交給 GPT Web。
