# CURRENT TASK — FIXED-SIGMA-MUL-NOISE-001

Task: `FIXED-SIGMA-MUL-NOISE-001`
Status: `AUTHORIZED_AUTONOMOUS_BATCH`（**尚未執行**；等待使用者對 Codex 發出 `開始`）
Authoritative spec: `specs/FIXED-SIGMA-MUL-NOISE-001.md`
Durable journal: `results/FIXED-SIGMA-MUL-NOISE-001/journal.md`
Planned result summary: `results/FIXED-SIGMA-MUL-NOISE-001/summary.md`

研究目標：對真正的 Standard vs Fast 的單次 `ct×ct → Mul → Relin → Rescale`「新增」數值誤差，建立有主導項理論依據、由實測校準的固定係數 σ 模型；在 Fast-only opt-in 診斷路徑的 `c0` 加入一致的 coefficient-domain Noise，所有乘法共用同一 σ。**不是**把被歸零的 `a` 加回來、不是恢復安全性，也不是把輸出總誤差直接當成一次新誤差。

Codex 啟動必須先遵守 Primary `AGENTS.md` 的 fetch／safe ff-only／本機變更 preflight，讀 `docs/RESEARCH_ENGINEERING_WORKFLOW.md`、`docs/MEASUREMENT_PLATFORM.md`、`docs/BACKEND_TRANSPARENCY_CONTRACT.md` 和上方 spec；涉及 Secondary 時遵守 `fast-qprefix` 的 `AGENTS.md` 與 `docs/FAST_QPREFIX_SPEC.md`。工作環境是**既有** Primary／Secondary，勿 clone 重建、勿清理未知本機變更。

本次僅授權 spec 的 CP-A～CP-D 及廉價 LogN13 非 Bootstrap 測試。**Bootstrap 預算 = 0 次**；完整 Bootstrap 注入覆蓋、LogN16、大型 benchmark、組員 BTS 不在此批次。遇到數學／表示法／實際呼叫鏈未決問題，保存 checkpoint 後 `NEEDS_WEB_REVIEW`。本次任務完成前不得指向下一個任務；勿重新執行先前已完成的 BTS Batch008。

歷史 Batch008 的 spec 與結果仍保存在 GitHub；本次新任務不會修改其來源、統計或預算。GPT Web 將在里程碑結束時獨立審查。
