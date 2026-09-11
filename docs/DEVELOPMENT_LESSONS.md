# Fast-CKKS Development Lessons

這份文件整理本專案到目前為止踩過的坑、錯誤模式，以及之後開發時應採用的工作原則。目的不是檢討誰做錯，而是把失敗轉成可重複使用的工程規範。

---

## 1. 不要把「第一個看得到的錯誤」當成 root cause

早期某些 stage 只能在特定邊界直接比較，因此曾出現「unpack/switch 看起來第一個 diverge」的情況。

後續更細 diagnostics 證明真正更早的 divergence 在 EvalMod / polynomial path。

### 經驗

- `first directly comparable checkpoint` ≠ `root cause`。
- 每次結論必須清楚標示：
  - first observed
  - first directly supported
  - proven cause
- 如果前一段 representation 無法直接比較，不可把下一個能比較的 stage 說成根因。

---

## 2. Oracle 本身也必須被驗證

曾經 Primary 的 Chebyshev plaintext oracle 把 `T0` 錯當成 0，導致 T2 / power-generation 被錯誤分類成失敗。

後來修正 `T0=1` 後，T2 實際上通過，first failing power 變成 T3。

### 經驗

- diagnostic oracle 不是天然正確。
- 新 oracle 必須至少有：
  - trivial identity/control cases；
  - 與歷史已知 vector 的 reproduction；
  - source-backed recurrence / coefficient definition；
  - cross-check actual production intermediate where possible。
- 一旦 oracle bug 被發現，必須在文件中明確標記舊結論 `superseded`，避免之後 AI 又引用。

---

## 3. 不要只看 residue 是否在 centered interval；要看 intended integer 是否唯一

block-2 constant diagnostic 顯示：

- 實際 q0/q1 centered residue 當然會落在 `[-Q01/2,Q01/2)`；
- 但 nominal intended encoded integer 已經是 `Q01/2` 的數百萬倍。

如果只看 `outside_half_count == 0`，會誤以為安全。

### 經驗

對 scalar injection / pre-Rescale value，容量判斷要分兩種：

1. **stored wrapped residue**：總會 modulo 回合法區間；
2. **intended centered integer**：才決定 q0/q1 是否能唯一表示原值。

真正的安全條件是 intended integer 满足：

`|x_intended| < Q01/2`

而不是 residue 看起來「沒有超界」。

---

## 4. Standard operation order 不能直接照抄到 two-limb Fast

Standard T2 的順序：

`Mul -> double -> Add(-1) -> Rescale`

在 full-RNS 下合法；在只維護 q0/q1 時，`-1` 於高 Scale 編碼後會先超過 centered uniqueness，Rescale 前已經 alias。

把 correction 移到 Rescale 後，T2 才恢復正確。

### 經驗

Fast backend 不是只把 Standard implementation 的 loop 縮短。

對每個 Standard pattern，都要問：

- 中間 intended integer 的 magnitude？
- 何時第一次需要 centered reconstruction？
- scalar 在哪個 Scale 被 encode？
- full-RNS 容量是否被 Standard 隱含使用？

---

## 5. 修一個例子之前，先判斷是不是整個 pattern 都有問題

block-2 constant 是 PS 第一個 failure，但不能直接特判 `block 2 / coefficient 0`。

因為整個 baby accumulator 都在約 `2^120` scale，後續 non-integer coefficient term 也會遇到同一 representation issue。

### 經驗

看到單一 failing input 時，先問：

> 這是 input-specific bug，還是 shared representation rule 的第一個 manifestation？

repair spec 應優先修 rule，不是修 index。

---

## 6. 不要用 heuristic 蓋掉設計問題

P2 one-sided pre-Rescale 原始想法會把 formal operand scale 從 ~`2^60` 降到 ~1，精度直接崩掉。

Codex 為了讓 tests 過，加入了 `Scale > q*2^10` 類 gate。結果 formal T3 乾脆不走新路徑，bug 完全沒修到。

### 經驗

當 implementation 自行加入 heuristic 時，要立刻問：

- 這個 heuristic 是 spec 中的 design，還是在掩蓋 design 不成立？
- formal failing case 到底有沒有走到 repair branch？

production repair 必須有 branch-selection evidence，不能只看 tests pass。

---

## 7. 一邊 pre-Rescale 不行，不代表 pre-Rescale 思路完全錯

one-sided pre-Rescale 失敗後，真正可行的是 balanced two-sided：

`left' = Rescale(m1*left)`

`right' = Rescale(m2*right)`

with `m1*m2 ≈ qL`.

formal case 找到 `m1=m2=2^30`，T3 誤差下降到 ~1e-7。

### 經驗

- repair 失敗時不要只在原 design 上調 threshold。
- 回到 algebra，重新維持原本 target Scale / Level identity。
- 優先推導 invariants，再寫 code。

---

## 8. Scale metadata 與 represented value 要分開看

compressed PS 的第一個 giant step中：

- Rescale semantic oracle PASS；
- Mul(T8) semantic oracle PASS；
- 只是 `b.Scale` 跟 `a.Scale` 差約 1e-12 relative；
- production `Scale.InDelta` 要求約 2^-116 closeness，因此拒絕。

### 經驗

CKKS debugging 要同時追：

1. coefficients / decoded represented value；
2. metadata Scale；
3. Level；
4. capacity。

`Scale mismatch` 不自動等於 arithmetic wrong。

相反地，metadata-only normalization 也不能直接假定安全，必須做 end-to-end semantic diagnostic。

---

## 9. 不要放寬 tolerance 來「讓它過」

當 Scale drift 只有 ~2^-39，而 production gate 是 ~2^-116，最簡單的誘惑是直接把 `InDelta` tolerance 改寬。

這不應直接做。

### 經驗

正確順序：

1. 先 diagnostic：metadata normalization 是否完整 PS 都正確？
2. 測後續 giant steps、final Rescale、target-scale restoration。
3. 若整條通過，再決定 production policy：
   - normalize metadata；
   - align via safe integer multiplier；
   - 或調整 planner scale rule。
4. 最後才考慮 tolerance，且要有 bounded rationale。

---

## 10. 每個 production repair 都要有「formal failing case actually takes repair path」證據

P2 最大教訓：code 雖然存在，但 formal case根本沒有走到它。

### 經驗

production gate 至少要記錄：

- branch selected；
- factors / scale schedule；
- input Level/Scale；
- output Level/Scale；
- formal case 確實走新 branch。

這些最好放 result artifact，不要只靠 debug print。

---

## 11. 每次只解一層，不要同時修 powers、PS、EvalMod、Bootstrap

目前有效的進展幾乎都來自 narrow diagnostics：

- Q01 stage
- EvalMod
- polynomial
- T2 primitive
- T3 capacity
- PS baby step
- compressed PS
- scale normalization

### 經驗

當 full Bootstrap fail 時，不要直接改 Bootstrap。

應使用：

`full -> stage -> substage -> primitive -> operation -> representation invariant`

直到 first supported cause 足夠窄，再做 production repair。

---

## 12. Repair 前先做 control；repair 後保留 regression

T2 diagnostics 中很有價值的 controls：

- `MulRelin -> Rescale` pass
- `MulRelin -> double -> Rescale` pass
- correction-before-rescale fail
- reordered correction-after-rescale pass
- low-scale control pass

這些讓因果關係變得很強。

### 經驗

一個好的 diagnostic 不只要有 failing case，還應有：

- one-operation-removed control；
- reordered control；
- low-scale / safe-range control；
- full-Q or Standard reference where appropriate。

---

## 13. 不要讓 diagnostic helper 汙染 production branch

多次 diagnostics 都需要 package-internal state。

### 經驗

允許：

- temporary build-tagged helper；
- test-only helper；
- `/tmp` raw output。

但任務結束前必須：

- helper 移除；
- Secondary 回到 exact expected commit（diagnostic-only task時）；
- worktree clean；
- only Primary result artifacts commit。

---

## 14. Codex 的「完成報告」不是證據本身

Codex 可能：

- 加 spec 沒要求的 heuristic；
- test pass 但 formal branch 沒走到；
- report 某個 first cause，但 artifact/oracle 有 bug；
- local worktree clean，ChatGPT 其實沒辦法直接看到 local 狀態。

### 經驗

ChatGPT reviewer 每次都要獨立：

- fetch commit；
- inspect diff；
- inspect actual source；
- inspect result JSON；
- verify exact Secondary SHA；
- compare spec methodology。

Final response要分清：

- `GitHub verified`
- `Codex reports local ...`

若 tool 無法驗 local clean，不要假裝驗過。

---

## 15. Git workflow 必須保守

多次工作跨兩個 repo，最怕 stale task / dirty tree / wrong branch。

### 經驗

每次 Codex `開始` 都應：

Primary：

1. branch / status
2. clean main 才 fetch + `pull --ff-only`
3. re-read AGENTS / CURRENT_TASK / spec

Secondary：

1. branch / status
2. clean `fast-ckks` 才 fetch + `pull --ff-only`
3. re-read AGENTS / FAST_CKKS_SPEC

dirty / wrong branch / non-FF：停止，不要自動 stash/reset/discard。

---

## 16. CURRENT_TASK 只當 pointer，不要塞歷史

如果 CURRENT_TASK 同時塞大量 context，容易 stale，也會讓 Codex 讀錯版本。

### 經驗

分層：

- `AGENTS.md`：永久 workflow / rules
- `docs/FAST_CKKS_SPEC.md`：Secondary durable architecture
- `docs/ORCHESTRATOR_HANDOFF.md`：研究始末 / current scientific state
- `docs/DEVELOPMENT_LESSONS.md`：踩坑與方法論
- `specs/...`：單一 task contract
- `CURRENT_TASK.md`：只指向 active spec

---

## 17. 每個 scientific conclusion 要帶 provenance

同一個數字如果來自不同 commit，語意可能不同。

### 經驗

重要結論要至少記：

- Primary commit
- Secondary commit
- formal config
- oracle version / corrected T0 status
- threshold
- first failing gate

不要只寫「T3 fail」。

---

## 18. 一旦結論被 supersede，要主動清理認知債

本專案已經有多個被推翻的中間結論。

### 經驗

handoff / spec / review 時應明確用：

- `superseded`
- `ruled out`
- `still valid independent evidence`

避免新 AI 把歷史 diagnostic 當現況。

---

## 19. 先證明 feasibility，再 productionize

目前最成功的 repair 都走：

`diagnostic prototype -> source-backed proof -> production spec -> formal gates`

而不是直接 production coding。

### 經驗

推薦固定模式：

1. isolate cause
2. derive math repair
3. diagnostic prototype
4. prove formal failing case
5. productionize generally
6. regression
7. formal end-to-end gate
8. only then performance / larger LogN

---

## 20. 不要太早跑 LogN16 / benchmark

當 LogN13 correctness 未通過時，跑 LogN16 或 performance 只會製造更多資料但不增加因果理解。

### 經驗

正式順序：

`LogN13 correctness -> review -> LogN16 correctness -> performance -> EXP-003`

除非 active spec 明確改變順序。

---

# Recommended checklist for every new task

開始前：

- [ ] sync Primary safely
- [ ] read AGENTS
- [ ] read CURRENT_TASK + spec
- [ ] sync/read Secondary if needed
- [ ] verify exact expected commits
- [ ] restate first supported cause, not historical superseded cause

實作／診斷中：

- [ ] one bounded question only
- [ ] formal case takes intended path
- [ ] local semantic oracle uses actual pre-operation inputs
- [ ] track Level + Scale + Degree + q0/q1 capacity
- [ ] distinguish nominal intended integer from wrapped residue
- [ ] add controls when causal claim needs them

結束前：

- [ ] stop at first failing gate if spec says so
- [ ] remove temp helpers
- [ ] run required tests
- [ ] inspect diff
- [ ] commit/push normally
- [ ] preserve exact result artifacts
- [ ] mark superseded conclusions explicitly
- [ ] reviewer independently verifies GitHub before next task

---

# One-sentence engineering principle

> 在 two-limb Fast CKKS 中，不能只問「這個 Standard CKKS 操作數學上對不對」；必須同時問「它的每一個中間 representation，在 q0/q1 容量、Scale、Level 與下一個 centered reconstruction 邊界上，是否仍然可唯一且精確地表示」。
