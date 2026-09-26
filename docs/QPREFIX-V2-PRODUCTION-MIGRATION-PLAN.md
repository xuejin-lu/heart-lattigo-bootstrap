# Q-prefix v2 Production Migration Plan

## 1. Architecture map

基準為同步後 Primary `main@ee56cb8`、Secondary `fast-qprefix@32b99ed6`；兩份 `CURRENT_TASK.md` 均指向 `QPREFIX-ARCH-PLAN-001`。Secondary 從 pre-F branch point `40532b4d`（`fast-ckks: validate LogN13 P93 Q012 bootstrap candidate`）出發；自 QPREFIX-AUDIT-002 證據 commit `fa46986` 起只改了 `CURRENT_TASK.md`，所以其後的 production source 未變。架構以 `docs/FAST_QPREFIX_SPEC.md` 為準：維持原 CKKS Q chain、Level、Scale 與 API，令 `w_Q(ell)=min(ell+1,4)`；不引入 F。

| Subsystem | Current role | Q-prefix v2 change | Depends on | Risk |
|---|---|---|---|---|
| Prefix policy、bounds、capacity (`schemes/ckks/fast/q012.go`) | 以 q0 形狀選 Q01/Q012，並提供 Q012 固定寬 CRT；參數與 Ring 路徑各有 policy 判斷。 | 建立唯一的 Level→prefix policy；以實際 q 值計算 prefix product，統一 per-component bound、strict `2B<S_Q` 與容量失敗語意。 | 無；所有 subsystem 的共同基礎。 | High |
| Ciphertext lifecycle (`ciphertext.go`、evaluator scratch) | constructors 依現有 policy 配 Q01/Q012；`Resize` 由既有 backing row 推斷寬度。 | allocation、resize、copy、degree growth、scratch 與 output shape 一律依同一 policy，保留 logical Level header，不將 dormant rows 誤當有效資料。 | Prefix policy。 | High |
| Evaluator arithmetic/domain (`fast_add.go`、`evaluator_ntt.go`、`fast_mul.go`、`automorphism.go`、`partial_ntt.go`、`truncate.go`) | 部分迴圈已讀 maintained count；另有 Q01 專用驗證、累加、旋轉與 helper。 | 將所有產生輸出的 primitive 一致擴至最多四個真實 q rows；保留 NTT/Montgomery、alias 與 evaluator-owned scratch 路徑。 | Policy、ciphertext lifecycle、bound contract。 | High |
| Rescale (`rescale.go`) | 有 Q01 與 Q012 分支；以 maintained lift 除 logical top `q_ell`，但沒有 Q0123 authoritative path。 | 對任意 Level 從目前 prefix 重建已證明的 centered lift，仍除 `q_ell`；驗證結果界、Level/Scale，容量不成立時不得提交半成品。 | Policy、bounds、arithmetic、lifecycle。 | High |
| 非 Rescale Level crossing (`Resize` callers、`fast_scaledown.go`) | `Resize` 可縮短 row slice；其本身是結構操作，不能決定 integer representative。 | 明確指定每個 DropLevel 邊界是 same-lift contraction 或 canonical logical contraction；把語意放在 level-transition owner，不藏在 resize。 | Policy、bounds、lifecycle。 | High |
| ScaleDown、ModUp、Trace (`fast_scaledown.go`、`fast_modup.go`、`trace.go`) | Level-0 ModUp 目前 materialize q1，profile 啟用時再 materialize q2；Trace 的 primitive 可依 maintained count 處理 rows。 | 以 q0 canonical representative（含 midpoint 規則）materialize 至 `q0..q_min(L,3)`；Trace、ScaleDown 與 domain conversion 共用 prefix policy。 | Policy、lifecycle、arithmetic、Level-0 contract。 | High |
| DFT / LinearTransform (`circuits/ckks/dft/fast.go`、`schemes/ckks/fast/linear_transform.go`) | C2S/S2C 共用 logical-Q plaintext matrices；Fast LinearTransform 與部分 diagonal accumulation 明確只處理 Q01。 | 保留單一 plaintext representation，直接消費 ciphertext 的 maintained prefix；一般化 BSGS/direct path、scratch 與 diagonal L1 capacity gate。 | Policy、lifecycle、arithmetic、Rescale；C2S 另依賴 ModUp。 | High |
| EvalMod / polynomial (`fast_evalmod.go`、`mod1/fast.go`、`polynomial/fast.go`) | 通用 PS planner 與特定 LogN13 Chebyshev/Q012 schedule 並存；DoubleAngle 逐輪乘、加、Rescale。 | generated powers、baby/giant steps、PS accumulation、Rescale boundary 與 DoubleAngle 共用 bounds/prefix；去除 Q012-only profile gate，但不在此計畫另定數學 schedule。 | Policy、lifecycle、arithmetic、Rescale、capacity oracle。 | High |
| Bootstrap public boundary (`fast_bootstrap.go`、`fast_packing.go`、ring-degree conversion) | 串接 ScaleDown→ModUp→C2S→EvalMod→S2C；packing、N1/N2 switch、finalization 有獨立 row/domain 邊界。 | 在不改 Standard frontend/API 的前提下，明確內部 compact rows 與 public structural contract；N1/N2、batch、finalization 不丟失有效 prefix，也不讀 dormant rows。 | 全部前述 subsystem。 | High |
| Acceptance / performance | 有 Q-prefix C2S raw/Rescale 證據與 public q0=56 control；完整目標路徑尚未由同一 current-branch stage map 證明。 | 每個 milestone 使用 semantic oracle、capacity/domain invariant、精簡 regression；整合後才比較固定 profile 的速度與 allocation，不以完整 Q fallback 遮掩失敗。 | 對應被驗收的 subsystem。 | Medium |

## 2. Migration order

Q-prefix v2 已是確定的 production 方向，不再要求先完成一個全路徑 readiness audit 才能開始 implementation。現有 QPREFIX-AUDIT-001/002 保留作 provenance 與 C2S correctness 證據；尚未補齊的 EvalMod/PS/DoubleAngle、S2C、finalization/packing capacity 證據，改由各自 implementation milestone 的 acceptance gate 同步完成。若某個 milestone 實際遇到 capacity failure，再針對該具體 stage 停止並回報；不要為了證明「不需要 F」而預先重跑整條 pipeline audit。

### QPREFIX-IMPL-001 — Prefix and capacity contract

- **Goal / scope:** 建立單一 `w_Q(ell)` 與 prefix product/bound/capacity contract，定義可辨識的 transactional capacity failure。
- **Why now:** 各 primitive 必須先共用同一 policy，避免把 Q01/Q012 分支擴成更多互不一致特例。
- **Acceptance gate:** Level 0/1/2/3/高 Level 的寬度表正確；`2B<S_Q` 嚴格邊界可測；拒絕時輸出 ciphertext 不變。
- **Out of scope:** 不改任何現有 primitive 的有效 row 數、不加入 F、不選自適應寬度。

### QPREFIX-IMPL-002 — Compact ciphertext lifecycle

- **Goal / scope:** 將 constructor、resize、copy、scratch、degree/output allocation 接到 policy，保持 dormant rows 不可讀。
- **Why now:** 算術遷移前先固定所有輸入/輸出 buffer 的形狀與所有權。
- **Acceptance gate:** 各 Level 的實際 backing rows 恰為 policy 寬度；alias、copy、縮放 Level 與 degree growth 的有效 rows 正確，logical Level/metadata 不變。
- **Out of scope:** 不擴充算術、不改公共參數或 ciphertext API。

### QPREFIX-IMPL-003 — Prefix-complete evaluator primitives

- **Goal / scope:** 一致擴充 Add/Sub、scalar/integer、Mul/Relin、automorphism/rotation/conjugation、partial NTT 與 truncate 等底層 producer/consumer。
- **Why now:** 後續 Rescale、ModUp 與 circuits 必須能依賴每個維持中的 residue row 都正確。
- **Acceptance gate:** 以逐 q row modular oracle 比對；檢查 degree、alias、NTT/Montgomery 與 dormant-row poison regression；不觸碰 Standard semantics。
- **Out of scope:** 不定義 PS/DoubleAngle schedule，不做完整 Bootstrap 整合。

### QPREFIX-IMPL-004 — Rescale and explicit Level transitions

- **Goal / scope:** 擴充 authoritative lift 到 Q0123；Rescale 永遠使用 logical `q_ell`、輸出 Level/Scale 按 Standard 規則更新；指定各 DropLevel 的 same-lift 或 canonical owner。
- **Why now:** 這是高 Level q_l 不在 maintained prefix、以及 3→2/2→1/1→0 縮窄時的共同語意關口。
- **Acceptance gate:** 與 Standard centered-rounding oracle 比較；涵蓋 Level>4、4→3、3→2、2→1、1→0；DropLevel 分別驗證 same-lift 與 canonical contraction；新 Level 的 bound 嚴格成立；錯誤具 transactional 性。
- **Out of scope:** 不以 dormant q rows 或 Standard full-RNS fallback 修補 capacity failure。

### QPREFIX-IMPL-005 — Level-0 ModUp and Trace

- **Goal / scope:** 以 Level-0 q0 canonical lift materialize 最高四個有效 q rows，並使 Trace/ScaleDown 遵循統一 policy。
- **Why now:** 建立 Bootstrap execution spine 的入口表示，供 C2S 使用。
- **Acceptance gate:** q0 midpoint、負/正邊界及每個物化 row 均有 oracle；Trace 前後 residue/domain/Level/Scale 符合 contract。
- **Out of scope:** 不修改參數、bootstrap schedule 或 Standard frontend。

### QPREFIX-IMPL-006 — C2S/S2C and LinearTransform

- **Goal / scope:** generalize direct/BSGS LinearTransform 與 C2S/S2C group chain；plaintext 繼續使用既有 logical-Q matrices。
- **Why now:** 依賴正確 arithmetic、Rescale 與 ModUp，且現有 QPREFIX-AUDIT-002 已提供 C2S recurrence 的可重用 oracle/evidence 形狀。
- **Acceptance gate:** 所有 maintained rows 對照 Standard/reference；diagonal L1 bound、strict capacity、Level/Scale/domain 與完整 helper equivalence 通過。
- **Out of scope:** 不複製 private-F plaintext mirror，不調整 DFT factorization。

### QPREFIX-IMPL-007 — EvalMod, PS, and DoubleAngle

- **Goal / scope:** 將 generated powers、PS baby/giant accumulation、polynomial Rescale boundary 與逐輪 DoubleAngle 納入相同 prefix/bound contract。
- **Why now:** 這是高增長中間量與舊 Q012 特例最集中的 circuit subsystem。
- **Acceptance gate:** 固定已接受的 profile/workload，逐必要 checkpoint 證明界與容量；對照 Standard/public correctness oracle；失敗需原子拒絕並指出 checkpoint。
- **Out of scope:** 不降低 precision threshold、不調參，也不在本 milestone 改變數學 polynomial 或 schedule。

### QPREFIX-IMPL-008 — Public Bootstrap boundary integration

- **Goal / scope:** 整合 packing/unpacking、N1/N2 ring-degree conversion、BootstrapMany 與 finalization 的 compact/public representation 邊界。
- **Why now:** 只有全鏈接合才能確認外部 API 和批次路徑未依賴 stale/full-Q rows。
- **Acceptance gate:** 同一 Standard/Fast frontend、config、workload、command；輸入/輸出結構與 metadata contract、generated/zero-secret correctness 及無 dormant-row read 通過。
- **Out of scope:** 不更動 Standard Lattigo、公開參數鏈或 security/noise model。

### QPREFIX-IMPL-009 — Performance and release gate

- **Goal / scope:** 以固定 LogN13/P93 workload 對 pre-F Q-prefix baseline 驗證 cap、配置量、hot-path regressions 與速度；只把 private-F 結果作外部比較資料。
- **Why now:** 正確性與 public boundary 穩定後才判斷 Q-prefix production 目標是否達成。
- **Acceptance gate:** 全路徑 semantic/capacity gate 通過；最多四個實際 q rows、無 full-RNS fallback；同條件效能結果與既定門檻一起交獨立 review。
- **Out of scope:** 不在本 architecture-planning 任務執行 benchmark，不重寫 `fast-ckks`。

## 3. Dependency graph

```text
001 policy/bounds → 002 lifecycle → 003 primitives
                                                    ├→ 004 Rescale/Level ─┐
                                                    └→ 005 ModUp/Trace ───┴→ 006 DFT/C2S/S2C
                                           003 + 004 ───────────────────────→ 007 EvalMod/PS/DA
                                      005 + 006 + 007 → 008 public Bootstrap → 009 performance gate
```

004 與 005 可在 003 的介面凍結後平行；006 與 007 可在共用 arithmetic/Rescale contract 穩定後平行。008 需等待兩條 circuit path 與 packing 邊界都完成。

## 4. Reuse map

### 可直接沿用 pre-F / current `fast-qprefix`

- 普通 CKKS parameter chain、`rlwe.Ciphertext` Level/Scale/API 與 Standard oracle；不建立第二種 frontend。
- NTT/Montgomery representation、evaluator-owned scratch、現有 Q01/Q012 row kernels 與 `uint192` 算術原語；只在已證範圍內沿用，不能把 Q012 標籤等同 Q0123 已支援。
- 共用 logical-Q DFT plaintext matrices、現有 Bootstrap stage orchestration，以及 QPREFIX-AUDIT-002 的四組 C2S raw/Rescale、restore、helper-equivalence 證據形式。
- 既有 C2S public q0=56 control 作 regression baseline；不把 Primary 舊 Audit-001 的其他 backend provenance 當作 current-branch capacity proof。

### 從 private-F 研究帶回概念，不帶 code

- per-component coefficient bounds、strict centered uniqueness、canonical ModUp midpoint、exact logical Rescale recurrence。
- transactional capacity errors、明確的 representation/domain boundaries、LinearTransform/plaintext L1 bounds。
- 區分數學最低容量與固定四-row engineering cap，以及 correctness/performance 分離的驗收方式。

不得帶回 private F primes、`FastCiphertext` 獨立 storage、Q↔F bridge、private-F plaintext mirror 或 F-resident circuit。

### 必須 generalize/reimplement 的部分

- Q01/Q012 profile selector、重複的 params/Ring width 判斷與由 backing row 推寬度的 lifecycle；改為唯一 `w_Q` policy。
- Q012 centered reconstruction/Rescale path 到最多四個 logical-Q rows，含實際 q 值、logical `q_ell`、rounding、縮窄與 overflow/capacity error。
- 仍寫死 Q01 的 arithmetic、automorphism/partial NTT、LinearTransform/DFT 累加，以及只 materialize 至 q2 的 ModUp。
- Q012 專用 PS/LogN13 shortcut 與 DoubleAngle 對 prefix/capacity contract 的接合；N1/N2 conversions 與 public packing/finalization 的 row contract。

## 5. Top architectural risks

1. **Rescale 與 prefix contraction 的 lift 語意（High）。** `q_ell` 可能高於 q3；新 Level 的 prefix 又可能變窄。Rescale、same-lift contraction 與 canonical DropLevel 是不同語意，必須各有 owner、界與 transactional gate。
2. **跨 primitive 的 row 完整性（High）。** policy 若先擴寬但某個 Add/Mul/rotation/DFT/scratch 仍只寫 Q01，會把 stale q2/q3 暴露成有效資料；每個 producer/consumer 必須一起納入驗收。
3. **EvalMod/PS/DoubleAngle 的成長界（High）。** 現有 P93 Q012 schedule 是 profile-specific；Q012 capacity evidence不能直接推論 q0123 全路徑正確或效能可接受。
4. **ModUp midpoint 合約衝突（High）— `NEEDS_LOCAL_IMPLEMENTATION_INVESTIGATION`.** 規格要求奇數 q0 的 `r=q0>>1` 使用正代表；目前 `circuits/ckks/bootstrapping/fast_modup.go` 的比較式是 `coeff >= q0>>1`。未做新實驗或修碼前，需由具體邊界 oracle 確認實作是否符合 frozen contract。
5. **Public structural boundary — `NEEDS_LOCAL_IMPLEMENTATION_INVESTIGATION`.** `schemes/ckks/fast/ring_degree.go` 的 conversion 驗證 maintained width，但目前 coefficient mapping 有 Q01 限制；需確認 N1/N2 packing 和 Bootstrap 外部消費者何處允許 compact rows、何處需要 materialization，避免暗中 full-RNS fallback。

尚未補齊的 EvalMod/PS/DA、S2C 與 finalization stage-map/capacity 證據，不再視為開始 implementation 的前置阻塞；它們分別由 QPREFIX-IMPL-007、QPREFIX-IMPL-006/008 的 acceptance gate 在實作時補齊。

## 6. First Codex task

### QPREFIX-IMPL-001 — Prefix and capacity contract

- 在 `fast-qprefix` 建立唯一的 `w_Q(ell)=min(ell+1,4)` policy 與 prefix-product / per-component bound contract。
- 將 strict centered-capacity predicate 與可辨識的 transactional capacity failure 形式固定為小型、可測的核心介面。
- 測試 Level 0、1、2、3、>=4，實際 q 值、`2B<S_Q` strict 邊界，以及失敗時 output 保持不變。
- 暫不把新寬度接到 constructors、arithmetic 或 circuits；避免下游尚未寫入 q2/q3 時先宣稱其有效。
- 不改參數、schedule、Standard semantics、`fast-ckks` 或 Secondary task pointer；無 benchmark。
- 後續 storage/primitive milestones 逐步切換 consumers，每一步都以同一 policy 和 oracle 驗收。
