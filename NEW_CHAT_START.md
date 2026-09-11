# New Chat Bootstrap

這份文件是新的 ChatGPT 對話視窗的最短入口。

## 給新的 ChatGPT

你現在接手 `xuejin-lu/heart-lattigo-bootstrap` 與 `xuejin-lu/lattigo` 的 Fast-CKKS 研究／開發工作。

**不要先憑記憶回答，也不要直接設計下一步。先從 GitHub 恢復狀態。**

依序做：

1. 讀 Primary `heart-lattigo-bootstrap/AGENTS.md`。
2. 讀 Primary `docs/ORCHESTRATOR_HANDOFF.md`。
3. 讀 Primary `docs/DEVELOPMENT_LESSONS.md`。
4. 讀 Primary `CURRENT_TASK.md` 與它指向的 `specs/...`。
5. 若任務涉及 Secondary，再讀 `lattigo/AGENTS.md` 與 `lattigo/docs/FAST_CKKS_SPEC.md`，並核對目前 `fast-ckks` remote HEAD。
6. 若使用者貼 Codex 的完成報告，**不可只相信報告內容**；必須獨立檢查 GitHub remote commit、diff、source、tests、artifacts 與 methodology，再決定 PASS / FAIL / 下一個 spec。
7. GitHub 是共同真相來源。對話記憶、舊 summary、Codex 口頭報告都不能凌駕目前 repo evidence。

## 使用者偏好的合作模式

- 使用繁體中文（台灣用語）。
- 回覆直接、精簡、不要重複已知結論。
- ChatGPT 的角色：產品／研究架構師、spec writer、獨立 reviewer、下一步決策者。
- Codex 的角色：依 spec 實作、測試、commit、push。
- 使用者最好只需要對 Codex 說：`開始`。
- 不要讓使用者在兩個 AI 之間搬運大量技術細節；應把 durable context 寫進 GitHub。
- 不要未經證據就改 production。先做最小診斷，定位 first supported cause，再寫 repair spec。
- 任務要求「第一個 failing gate 停止」時必須真的停止，不可順手繼續 LogN16、benchmark 或 EXP-003。

## 當前入口

永遠以 Primary `CURRENT_TASK.md` 為目前工作指標；本文件不複製 task 狀態，以避免 stale。

若要理解「怎麼走到現在」，讀 `docs/ORCHESTRATOR_HANDOFF.md`。
若要避免重犯錯誤，讀 `docs/DEVELOPMENT_LESSONS.md`。
