# BTS Fast QP 透明化整合：階段結果

**狀態：`NEEDS_WEB_REVIEW`**。一次獲准的 Fast No-Trace 呼叫已進入 `bts.Bootstrap`，但在 C2S conjugation 失敗。依任務規則，已開始的失敗呼叫不可重跑，因此新修正沒有端到端 Bootstrap 驗證，也不能宣稱完成或計算效能倍率。

## 執行結果

| 項目 | 結果 |
|---|---|
| 前端 | `external/BTS/` 原始 `main.go`，未修改；LogN16、SlotLimbs=15、`BTSModePlain`、amp=1、seed=0 |
| 實際執行 backend | `fast-qprefix` commit `ac497c82adf60733ef652cfcd39acf741f6cb1d8` |
| `NewBTS()` | 成功，46.41 秒 |
| 輸入 | Level 3、Scale `2^32`、32,768 slots |
| Bootstrap | 已進入；在 C2S `Conjugate` 的 Galois key lookup 失敗，沒有輸出 ciphertext 或 Bootstrap 完成時間 |
| 精度／延遲比較 | 無 Fast 輸出，不計算誤差或倍率 |
| 重試 | 不執行；本輪唯一已開始呼叫已消耗，第二次呼叫不符合「僅首呼叫未進入 Bootstrap 才可用」條件 |

錯誤原文：

```text
Bootstrap: C2S: Conjugate: cannot Conjugate: cannot apply Automorphism: unsupported: cannot use Fast evaluation key with Standard execution: key for galEl 131071 = 5^{16384} key is missing
```

原因已由呼叫鏈確認：未修改的 BTS `NewC2SEvaluatorDH` 使用普通 `ckks.NewEvaluator`，然後呼叫 `ckks.Evaluator.Conjugate`。原本 CKKS evaluator 沒有覆寫 public `Automorphism`，呼叫因此落到 Standard RLWE evaluator，再由 `MemEvaluationKeySet.GetGaloisKey` 拒絕 `KeyLayoutFast`。這是 Fast public API dispatch 缺口，不是 Q-prefix 容量錯誤。

## 有界修正與驗證

在失敗呼叫之後，已加入並推送 Secondary 修正 `0b922dc48315128a27db811a5f237e0a6c9813d6`。Fast public `Automorphism` 現在：

- 每次檢查輸入 authoritative Q-prefix 上的 `c1=0`；
- compact ciphertext 重用既有共享 Fast Q-prefix automorphism core；
- 完整高層 Q 表示只有通過既有 centered-lift／容量驗證才保留全列；
- 真正執行 Galois permutation，只省略已證明零輸入所對應的 key switch；
- 驗證失敗時不修改輸出。

此修正通過非 Bootstrap 測試：

- `go test ./schemes/ckks -run 'TestGenericFastCKKSKeyGeneratorProducesZeroSecretEvaluationKeys|TestFastQPCompatibility|TestPublicFastAutomorphismUsesPrefixCoreAndVerifiedFullQPath' -count=1`
- `go test ./schemes/ckks/fast -count=1`
- bootstrapping API focused tests（未呼叫 Bootstrap）
- `go vet ./core/rlwe ./schemes/ckks ./schemes/ckks/fast ./circuits/ckks/bootstrapping`
- `git diff --check`

修正後的 `go test ./schemes/ckks -count=1` **未通過**：包含 Fast 模式下舊的錯誤字串斷言、期待 Standard-layout relinearization key 的測試，以及 `RescaleTo/Single` 在空 `c1` row 上的歷史 panic。這些沒有在本任務中擴大修正。最終 Secondary commit 的 BTS binary 僅完成 compile-only 驗證，未執行。

## 操作路徑分類

1. **數學上省略**：普通 Fast key generation 產生 `a=0` evaluation keys；只在 API 邊界證明 `c1`／decomposition 為零後才省略 Decompose、GadgetProduct 或 key-switch 的實際乘加。Galois permutation 不省略。
2. **重用 Fast 核心**：compact Q-prefix 公開 automorphism 重用共享 Fast automorphism core，Q-prefix Rescale 仍走既有 Fast 路徑；BTS 原有 BSGS 程式保持不變，未以完整 Standard key switching 取代。
3. **QP 表示相容**：lazy QP automorphism 保留 P-scaled Q 結果及零 P；ModDown 先驗證 P-scaled 表示後移除 P 因子，再確認 Q-prefix centered-lift 條件。高層完整 Q Rescale 只在完整輸入及輸出表示驗證後使用。
4. **有條件 full-Q 相容路徑**：完整 Q 的 public Galois permutation／Rescale 需通過既有表示驗證；不執行 Standard GadgetProduct 或完整 Key Switching。具體到本次失敗呼叫，C2S 走到後處理 `Conjugate`，但失敗發生在 Galois lookup；修復後的這條路徑未經 Bootstrap 再驗證。

## Provenance 與歷史 Standard 對照

- BTS 原始 `.go`、`go.mod`、`go.sum` 的 SHA-256 manifest、build/run 命令、完整 console transcript 與 attempt journal：見 [journal](BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002-journal.md) 及 [attempt log](BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002-attempt-01.log)。
- 第一次實際執行 binary SHA-256：`482c854e23125085692aba3c302954177a7263a566b8f3daa245d92968ddbe0b`，綁定 commit `ac497c82...`。
- 修復後 compile-only binary SHA-256：`72a4d9d38e9478d4d244de03c06bda1dc8fe60a8cacd43c4024f7a4acfb5167b`，綁定 commit `0b922dc4...`；未執行。
- 原有 Standard baseline（不重跑）：Bootstrap 139.653 秒、max absolute error `1.708e-3`、mean absolute error `1.172e-4`。Fast 沒有成功輸出，因此不提供 speedup 或數值比較。
- `external/BTS/` 原始碼及正式 module files 未修改，未加入 Git；所有 BTS binaries 均未提交。
