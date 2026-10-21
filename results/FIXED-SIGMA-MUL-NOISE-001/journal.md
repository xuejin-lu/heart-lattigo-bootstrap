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
| CP-A Source/maths audit | COMPLETE (isolated operation only) | source map／theory／Q-prefix and Scale limits recorded below; no Bootstrap-wide boundary claim |
| CP-B Standard/Fast single-op | IN_PROGRESS | Bootstrap 0 calls; E=32, LogN13 |
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

## 2026-10-22 Codex continuation — CP-A source map and math preflight

**同步基線：** Primary `main` clean at `61495387be17673767412b3b4fa8ecffb295fa6c`; `origin/main` same SHA. Secondary `fast-qprefix` clean at `ee308c9743000b5f281a84a871e7c4bd59b176c4`; `origin/fast-qprefix` same SHA. E is explicitly 32 per user choice. Genuine Standard source pin remains `5dbffbdea05394de2ca3a432ed5318aa832e3f40`. No source/build edits or Bootstrap calls were made during CP-A.

**Reuse inventory and call chain (Secondary at `ee308c9743000b5f281a84a871e7c4bd59b176c4`):**

- Ordinary public `ckks.Evaluator.MulRelinNew` / `MulRelin` in `schemes/ckks/evaluator.go` detects the Fast parameter capability and dispatches degree-one ct×ct to `mulRelinFastCKKSZeroSecret` in `schemes/ckks/evaluator_fast_zero_secret.go`, then to the shared `schemes/ckks/internal/fastcore.MulCore.ApplyRows`. The supported Fast input is NTT, non-Montgomery, degree one, c1=0; only `w_Q(level)=min(level+1,4)` rows are authoritative.
- Ordinary public `ckks.Evaluator.Rescale` in `schemes/ckks/evaluator.go` sends the ordinary compact Fast case to `RescaleCore.ApplyRows`; the explicit `schemes/ckks/fast` evaluator uses the same shared core. The rescale kernel reconstructs and rounds with its declared source Q-prefix, divides by the logical top-q divisor, preflights centered capacity, transforms each target residue back to NTT, and preserves Scale/NTT/Montgomery metadata. The high-level lazy-QP compatibility route is separately armed and requires full rows; it is outside this isolated workload.
- At the selected Level 5 → Level 4 operation, Fast keeps q0..q3 as input and output residues while q5 is the logical rescale divisor. The shared core enforces the centered-prefix capacity condition (strict `2B < q0*q1*q2*q3`). This is a representation bound, not a claim that arbitrary high-Level ciphertexts are decryptable from the prefix.
- `schemes/ckks/fast/evaluator_ntt.go` exposes explicit Fast `MulRelin`; `schemes/ckks/fast/rescale.go` routes `Rescale` through the same Q-prefix policy. These are existing kernels, not candidates for duplication.
- `circuits/ckks/polynomial/fast.go` contains Q-prefix polynomial-power operations such as `MulRelinElementQPrefixRows`; their other operands are polynomial elements/plaintexts and are outside this ct×ct target. Bootstrapping contains standalone `Rescale`/`RescaleTo`, scalar/polynomial operations, and deferred/lazy sequences; no generic source boundary proves that every MulRelin is paired one-to-one with a following Rescale. CP-C, if reached, must own an explicitly isolated `MulRelin → Rescale → inject` helper and must not instrument arbitrary `Rescale` or claim Bootstrap coverage.
- The initial candidate workload is two native encryptions at Scale `2^45`, ct×ct `MulRelin`, then one Rescale at input Level 5 using q5, producing Level 4 with Scale `2^90/q5`. E=32 is recorded from the accepted public profile; no Bootstrap is invoked. Both backends must verify actual constructed parameters and output metadata rather than infer them from config labels.

**Theory and observables:** Let decoded encrypted inputs be `x̃=x+e_x`, `ỹ=y+e_y`. For output `z̃` and independently decoded input product, `r_op=z̃-(x̃⊙ỹ)` is the operation's decoded incremental residual. The final error against the original message decomposes as `z̃-x⊙y = x⊙e_y + y⊙e_x + e_x⊙e_y + r_op`; dropping the second-order term requires `|e_x⊙e_y|` to be small at the reported tolerance. In `r_op`, the input-error propagation is removed by the measured decoded-input product; its remaining contributors can include Standard relinearization/key-switch error, Rescale rounding, scale/encoding effects, and numerical decode effects. Fast's current zero-secret `c1=0` path omits Standard's key-switch contribution, so CP-B must not label the entire residual as “relinearization noise.” Rescale rounding is deterministic conditional on ciphertext coefficients and need not be zero-mean or independent across slots/trials. Therefore a variance subtraction is only a candidate estimator after its assumptions are checked; negative difference, material bias/correlation, or unmodelled distribution stops calibration.

**Injection representation contract already fixed by the spec:** The only candidate boundary is after the isolated complete chain. A single signed coefficient polynomial `η(X)` must be reduced modulo each maintained q0..q3, transformed independently with that row's NTT, and added to c0 in the output's existing NTT/non-Montgomery representation. It must not sample a different polynomial per prime or write coefficient samples directly into NTT rows. The operation must retain output Level/Scale/metadata and must not touch dormant limbs. Capacity must be checked against the actual post-injection centered Q-prefix coefficients; no capacity waiver or scale adjustment is authorized. This defines a bounded candidate for implementation but does not authorize extending it to other operations.

**Primary platform reuse/gap:** Reuse `cmd/perfprobe` for the matched executable/provenance/result path, `internal/perfmeasure.ParametersFromConfigWithE` and `PublicMulRescaleWorkload`/fingerprint for E32 effective parameters and deterministic inputs, and `internal/numericalmetrics` for decoded-domain checks. Existing `numericalmetrics.Compare` supplies complex SNR and `CompareRealVectors` supplies real-vector absolute mean/RMSE/max; neither provides signed complex mean and variance required here. If CP-B proceeds, add the minimum reusable complex-statistics primitive to `internal/numericalmetrics` rather than implementing duplicate aggregation in a runner. Existing public Mul/Rescale fixture has an intervening Add and a third q5-scaled operand, so it is reusable infrastructure/provenance prior art but not the exact two-ciphertext `r_op` fixture.

**CP-A outcome:** `PASS_SCOPED` for one isolated ct×ct `MulRelin + Rescale` boundary, with the conditions above. No general Bootstrap injection boundary or completeness claim is established. CP-B is authorized to proceed; CP-C remains gated on measured evidence and implementation feasibility. The old `experiment/standard-a-zero-noise-preserved` worktree and historical A=0 outputs were not opened or used as a Standard source.

## 2026-10-22 CP-B adapter implementation and pre-measurement review

**Platform reuse:** CP-B extends only the maintained `cmd/perfprobe public-native` and `compare-public` path plus the shared `internal/numericalmetrics` complex-vector moments/covariance helper. The briefly prototyped separate `mulrescale` runner and its task-only workload/fingerprint helpers were removed during self-review before measurement; no retired CLI or second statistics path remains.

**Actual operation boundary:** To keep the accepted public workload and application operation sequence unchanged, the measured MulRelin input `x` is the decoded `add_level5` ciphertext (`AddNew(EncryptNew(A), EncryptNew(B))`) and `y` is the decoded `encrypt_c_q5_level5` ciphertext. Exactly one `MulRelinNew(x,y)` followed by `Rescale(q5)` is the operation under review; its output is the decoded `rescale_q5_level4` checkpoint. This uses three native SK-encrypted operands in the surrounding public workload; the prior Add is not attributed to the target operation because `r_op = z̃ - x̃*ỹ` uses the actual decoded operation inputs. `decoded_total_error = z̃ - (A+B)*C` is stored separately against the independent cleartext oracle. The operation's combined elapsed time is the sum of its existing single-call MulRelin and Rescale phase timers.

**Artifact contract:** New nested schema `fixed-sigma-mul-relin-rescale-residual.v1` stores DIAGNOSTIC_ONLY class, SK cohort, PK unsupported reason, one ct×ct multiplication, zero injections, no sigma/seed, incremental residual mean/variance/RMS/max, total-error moments, decoded input/output fingerprints, and binary SHA-256. The existing vector sidecar carries the two residual/error vectors; the comparator re-derives both from checkpoint vectors and cleartext oracle before reporting them. The pair output stores Standard/Fast incremental and total-error statistics, slot variance difference, and cross-backend covariance. A single pair is explicitly insufficient to prove zero-mean independent noise or identify coefficient-domain σ. The exact task Fast pin `ee308c9743000b5f281a84a871e7c4bd59b176c4` is accepted only for LogN13; original Standard remains pinned to `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.

**Pre-measurement tests:** Targeted `perfprobe` residual/provenance/pin tests and all `internal/numericalmetrics` / `internal/perfmeasure` tests passed with both isolated source modfiles and their respective `perf_standard` / `perf_fast` tags. An initial unfiltered `cmd/perfprobe` package run also selected historical A=0-specific tests and one test requiring the default Secondary path; those failed under the genuine Standard modfile (`TestStandardAZeroBootstrapKeyAuditChecksQAndP`, `TestStandardAZeroDiagnosticUsesNativeEncryptNewInput`, and `TestVerifyCompiledBackendSourceResolvesRelativeReplacePath`). They are not part of this task and were not edited or used as measurement evidence; subsequent runs are restricted to task-relevant tests.

**Measurement status:** No CP-B workload has been executed yet; Bootstrap calls remain exactly zero. No coefficient σ has been estimated. The implementation source checkpoint must be committed and clean before the platform's provenance gate will permit either backend run.
