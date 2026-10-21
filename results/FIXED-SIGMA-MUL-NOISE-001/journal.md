# FIXED-SIGMA-MUL-NOISE-001 — Durable Experiment Journal

**建立日：** 2026-10-21 (Asia/Taipei)
**狀態：** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`（CP-C 分類為 `RMSE_ONLY`）
**授權來源：** 使用者確認「每次完整 ct×ct＋Relin＋Rescale 新增誤差」，全體乘法共用一個固定 σ，要求 GitHub 版本管理；本 GitHub 提交僅為**規格與任務建置**。

## 版本控制及不可變歷史

- Primary 初始規格基底：`4476da0fad330205d7333d25b23e3ebccebfe2bb` (`main`)。
- Secondary 初始參考：`ee308c9743000b5f281a84a871e7c4bd59b176c4` (`fast-qprefix`)。
- 原始 Standard 參考：`5dbffbdea05394de2ca3a432ed5318aa832e3f40`（只讀、不可修改）。
- CP-A/CP-B Primary commits：`b99f63710718ac4373e704afef3023d62fc17f3d` 與 `8cf9f4c693d67d54ca246ec53c3478828bf42f4f`；CP-C harness commit：`a47e402a31e9181ac94dfb430e1cc337fbd8b3a0`。
- CP-C 證據 checkpoint：`e543cf8a9aac5199ebc849e044328cbfac835d3d`；CP-C 執行期間 Primary source hash 固定為 `79b13e0978030f575458230d7068f572976ae29b3ebeed13f3a9fbc66fa8d8e5`。
- Genuine Standard：`5dbffbdea05394de2ca3a432ed5318aa832e3f40`；CP-C helper Fast：`9ef0315feaf59dbee03bf0852b72c76fdeb6b800`，位於 `fast-qprefix`。
- CP-C Standard/Fast+Noise binary SHA-256：`7521b4e33dfeeb80c266d5526d82845f1cf54c619b61dfe151bcd73af050ea76` / `8e6eb3934ed5bb545524344654314cdf7fbff67928fdd5268b18d75ad39153d8`。
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
| CP-B Standard/Fast single-op | COMPLETE | genuine Standard 與原始 Fast；Bootstrap 0 calls；候選 σ 僅由 Pair 01 校準 |
| CP-C Opt-in Fast helper | COMPLETE — `RMSE_ONLY` | 單一孤立 MulRelin+Rescale 邊界；固定 σ；三組 matched、每組恰注入一次；未建立一般 Noise 分布模型 |
| CP-D Summary & review handoff | READY_FOR_WEB_REVIEW | journal、summary、manifest、source/test/JSON/vector 證據及 commits 已整理；等待 Web 獨立審查 |

## Expensive call budget

| 類型 | 授權 | 已用 | 保留 |
|---|---:|---:|---:|
| Standard Bootstrap | 0 | 0 | 0 |
| Fast Bootstrap | 0 | 0 | 0 |

累計 Standard/Fast Bootstrap 呼叫均為 `0`；LogN16、大型 benchmark 與 BTS 均未執行。CP-B 凍結候選 `σ=4.004280012097432`；CP-C 不曾調整此值。

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
- The preliminary source-map workload description is refined by CP-B evidence below: the measured `ct_x` is `AddNew(EncryptNew(A), EncryptNew(B))` at Scale `2^45`; `ct_y=EncryptNew(C)` is at q5 Scale. Their product is Relin'ed once at Level 5, then Rescale divides by q5 and yields Level 4 at Scale `2^45`. E=32 is recorded from the accepted public profile; no Bootstrap is invoked. Both backends verify actual constructed parameters and output metadata rather than infer them from config labels.

**Theory and observables:** Let decoded encrypted inputs be `x̃=x+e_x`, `ỹ=y+e_y`. For output `z̃` and independently decoded input product, `r_op=z̃-(x̃⊙ỹ)` is the operation's decoded incremental residual. The final error against the original message decomposes as `z̃-x⊙y = x⊙e_y + y⊙e_x + e_x⊙e_y + r_op`; dropping the second-order term requires `|e_x⊙e_y|` to be small at the reported tolerance. In `r_op`, the input-error propagation is removed by the measured decoded-input product; its remaining contributors can include Standard relinearization/key-switch error, Rescale rounding, scale/encoding effects, and numerical decode effects. Fast's current zero-secret `c1=0` path omits Standard's key-switch contribution, so CP-B must not label the entire residual as “relinearization noise.” Rescale rounding is deterministic conditional on ciphertext coefficients and need not be zero-mean or independent across slots/trials. Therefore a variance subtraction is only a candidate estimator after its assumptions are checked; negative difference, material bias/correlation, or unmodelled distribution stops calibration.

**Injection representation contract already fixed by the spec:** The only candidate boundary is after the isolated complete chain. A single signed coefficient polynomial `η(X)` must be reduced modulo each maintained q0..q3, transformed independently with that row's NTT, and added to c0 in the output's existing NTT/non-Montgomery representation. It must not sample a different polynomial per prime or write coefficient samples directly into NTT rows. The operation must retain output Level/Scale/metadata and must not touch dormant limbs. Capacity must be checked against the actual post-injection centered Q-prefix coefficients; no capacity waiver or scale adjustment is authorized. This defines a bounded candidate for implementation but does not authorize extending it to other operations.

**Primary platform reuse/gap:** Reuse `cmd/perfprobe` for the matched executable/provenance/result path, `internal/perfmeasure.ParametersFromConfigWithE` and `PublicMulRescaleWorkload`/fingerprint for E32 effective parameters and deterministic inputs, and `internal/numericalmetrics` for decoded-domain checks. Existing `numericalmetrics.Compare` supplies complex SNR and `CompareRealVectors` supplies real-vector absolute mean/RMSE/max; neither provides signed complex mean and variance required here. If CP-B proceeds, add the minimum reusable complex-statistics primitive to `internal/numericalmetrics` rather than implementing duplicate aggregation in a runner. Existing public Mul/Rescale fixture has an intervening Add and a third q5-scaled operand, so it is reusable infrastructure/provenance prior art but not the exact two-ciphertext `r_op` fixture.

**CP-A outcome:** `PASS_SCOPED` for one isolated ct×ct `MulRelin + Rescale` boundary, with the conditions above. No general Bootstrap injection boundary or completeness claim is established. CP-B is authorized to proceed; CP-C remains gated on measured evidence and implementation feasibility. The old `experiment/standard-a-zero-noise-preserved` worktree and historical A=0 outputs were not opened or used as a Standard source.

## 2026-10-22 CP-B adapter implementation and pre-measurement review

**Platform reuse:** CP-B extends only the maintained `cmd/perfprobe public-native` and `compare-public` path plus the shared `internal/numericalmetrics` complex-vector moments/covariance helper. The briefly prototyped separate `mulrescale` runner and its task-only workload/fingerprint helpers were removed during self-review before measurement; no retired CLI or second statistics path remains.

**Actual operation boundary:** To keep the accepted public workload and application operation sequence unchanged, the measured MulRelin input `x` is the decoded `add_level5` ciphertext (`AddNew(EncryptNew(A), EncryptNew(B))`) and `y` is the decoded `encrypt_c_q5_level5` ciphertext. Exactly one `MulRelinNew(x,y)` followed by `Rescale(q5)` is the operation under review; its output is the decoded `rescale_q5_level4` checkpoint. This uses three native SK-encrypted operands in the surrounding public workload; the prior Add is not attributed to the target operation because `r_op = z̃ - x̃*ỹ` uses the actual decoded operation inputs. `decoded_total_error = z̃ - (A+B)*C` is stored separately against the independent cleartext oracle. The operation's combined elapsed time is the sum of its existing single-call MulRelin and Rescale phase timers.

**Artifact contract:** New nested schema `fixed-sigma-mul-relin-rescale-residual.v1` stores DIAGNOSTIC_ONLY class, SK cohort, PK unsupported reason, one ct×ct multiplication, zero injections, no sigma/seed, incremental residual mean/variance/RMS/max, total-error moments, decoded input/output fingerprints, and binary SHA-256. The existing vector sidecar carries the two residual/error vectors; the comparator re-derives both from checkpoint vectors and cleartext oracle before reporting them. The pair output stores Standard/Fast incremental and total-error statistics, slot variance difference, and cross-backend covariance. A single pair is explicitly insufficient to prove zero-mean independent noise or identify coefficient-domain σ. The exact task Fast pin `ee308c9743000b5f281a84a871e7c4bd59b176c4` is accepted only for LogN13; original Standard remains pinned to `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.

**Pre-measurement tests:** Targeted `perfprobe` residual/provenance/pin tests and all `internal/numericalmetrics` / `internal/perfmeasure` tests passed with both isolated source modfiles and their respective `perf_standard` / `perf_fast` tags. An initial unfiltered `cmd/perfprobe` package run also selected historical A=0-specific tests and one test requiring the default Secondary path; those failed under the genuine Standard modfile (`TestStandardAZeroBootstrapKeyAuditChecksQAndP`, `TestStandardAZeroDiagnosticUsesNativeEncryptNewInput`, and `TestVerifyCompiledBackendSourceResolvesRelativeReplacePath`). They are not part of this task and were not edited or used as measurement evidence; subsequent runs are restricted to task-relevant tests.

**Measurement status at the pre-measurement checkpoint:** CP-B had not yet run. The following continuation supersedes that status after the recorded small matched runs.

## 2026-10-22 CP-B completed — matched zero-Bootstrap single-operation evidence

**Execution contract:** Same Primary source commit `b99f63710718ac4373e704afef3023d62fc17f3d` and measurement source SHA-256 `6f8a9237ebf2152c369abc2fe797c4f168793f056954a753d706126fda38d3ad`; genuine Standard source `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; unmodified Fast `ee308c9743000b5f281a84a871e7c4bd59b176c4`; E=32, LogN13, 4096 slots, 8192 coefficients, q0..q5 input Level5 and Scale `2^45 × q5`, Level4 output Scale `2^45`. All runs used SK lifecycle, the same original public workload and zero Bootstrap budget/calls. Standard was rebuilt from its isolated original source; no A=0 source or result was used.

**Build and environment provenance:** `go1.26.4`, `darwin/arm64`; Standard binary SHA-256 `081041b3971117c262c1a0759d3e2923efe168f30fd55f84503e337e0c95eb84`; original Fast binary SHA-256 `b988c99c78da11c263c930f7e90b1723839717f1a707759e49e7629a21b9f877`. These three Fast outputs and their vector sidecars are byte-identical because this zero-secret path is deterministic for the fixed workload; all three Standard runs are retained independently. The sidecars hold the bounded decoded vectors used by `compare-public`; compact JSON summaries, hashes, exact provenance, and pair statistics are in this directory. The artifact manifest's size/SHA entries were checked against all 13 artifacts.

| Pair | Standard `r_op` variance | Standard RMS | Standard Max | Fast baseline variance | Fast RMS | Fast Max | Standard−Fast variance | `sigma` implied by this pair |
|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| 01 (calibration only) | `1.0665895235026625e-22` | `1.0327794731812117e-11` | `4.6051551029409136e-11` | `5.531307679675344e-25` | `7.43727622606826e-13` | `2.2054128106146474e-12` | `1.061058215822987e-22` | `4.004280012097432` |
| 02 (held out) | `1.0540817934067598e-22` | `1.0268142314295168e-11` | `5.2315488998066924e-11` | `5.531307679675344e-25` | `7.43727622606826e-13` | `2.2054128106146474e-12` | `1.0485504857270844e-22` | `3.9806088657467544` |
| 03 (held out) | `1.088767133446543e-22` | `1.0435049051696197e-11` | `8.218360133970715e-11` | `5.531307679675344e-25` | `7.43727622606826e-13` | `2.2054128106146474e-12` | `1.0832358257668675e-22` | `4.04591114352061` |

The calibration-only candidate is frozen at coefficient-domain `sigma=4.004280012097432`; no run 02/03 value was used to tune it. The mapping follows the encoder's special slot IFFT normalization: with `S=4096` slots, a real coefficient perturbation contributes two real coefficient lanes per complex slot, so for independent zero-mean coefficient errors `Var(slot) = 2S·Var(eta)/Scale² = N·Var(eta)/Scale²` with `N=8192`. Thus `sigma_candidate = sqrt(variance_delta × Scale² / N)`. This is a first-order variance model to be checked against the actual decoded injection delta; the integer sampler's empirical variance and the injected slot delta remain required evidence. The three pairwise implied values are all nonnegative and within about 0.6% of the frozen candidate, while cross-backend correlation magnitudes are `0.00396`, `0.02160`, and `0.02470`. This supports a bounded CP-C trial, not a general noise model or proof of independent trials: each row is one vector of correlated CKKS slots.

All three existing `compare-public` pairs passed the shared 1e-6 public checkpoint gate and their incremental residual vectors were rederived from decoded input/output vectors. This gate is only the harness's broad plaintext-oracle check; residual differences are around `1e-11`, so it does not by itself establish noise-model success. Standard operation elapsed times were `4.561001`, `5.022542`, `5.711041 ms`; Fast baseline times were `2.102584`, `2.074166`, `2.036751 ms`. These small samples are descriptive only.

**CP-B outcome:** `PASS_SCOPED_CANDIDATE_IDENTIFIED`. It is reasonable to perform the already-authorized, isolated CP-C injection trial at this one LogN13/E32/Scale profile. No claim is made that Standard's whole `r_op` is key-switch noise; its contributors also include rescale rounding and encoding/decode effects. PK remains unsupported because the Fast zero-secret MulRelin contract requires `c1=0`; no SK result is relabeled PK.

**CP-B artifacts:** `standard-run-01..03.json`, their three `*-vectors.json` sidecars, `fast-run-01..03.json`, `fast-vectors-shared.json`, and `paired-run-01..03.json`; all exact sizes and SHA-256 values are in `manifest.json`. No Bootstrap, LogN16, or BTS calls were made.

## 2026-10-22 CP-C completed — one post-Rescale fixed-σ injection

**Source and build provenance:** Primary source was `a47e402a31e9181ac94dfb430e1cc337fbd8b3a0` for pair 01 and `e543cf8a9aac5199ebc849e044328cbfac835d3d` for pairs 02/03. The latter checkpoint adds pair-01 evidence only; all three binaries report the identical measurement source SHA-256 `79b13e0978030f575458230d7068f572976ae29b3ebeed13f3a9fbc66fa8d8e5`. Genuine Standard stayed at pristine `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Fast helper source is `fast-qprefix@9ef0315feaf59dbee03bf0852b72c76fdeb6b800`. Standard and Fast+Noise binary SHA-256 are respectively `7521b4e33dfeeb80c266d5526d82845f1cf54c619b61dfe151bcd73af050ea76` and `8e6eb3934ed5bb545524344654314cdf7fbff67928fdd5268b18d75ad39153d8`. Environment was Go `go1.26.4`, `darwin/arm64`; E=32, LogN13, 4096 slots, N=8192, input Level5 / Scale `2^45×q5`, output Level4 / Scale `2^45`, SK cohort, one ct×ct MulRelin and one Rescale. All three pairs passed `compare-public`; each run reported zero Bootstrap calls and exactly one Fast injection. Same public workload, config, Q/P, and decoded-operation definition were used; Standard arithmetic and source were untouched.

The first attempt to start pairs 02/03 while pair-01 files were uncommitted was rejected by the runner's clean-Primary provenance guard before measurement. The six outputs were then generated under `/private/tmp` while Primary remained clean and copied unchanged into the task result directory. The rejected attempts consumed no Bootstrap budget and produced no measurement artifacts.

The frozen value remained `sigma=4.004280012097432` for all three seeds. Pair 01 is the calibration-transfer check (seed `20261022`); pairs 02/03 are held out (seeds `20261023`, `20261024`). No pair was used to change σ. Summary statistics are per single 4096-slot vector; slots are not independent experimental repetitions.

| Pair / role | Standard `r_op` variance / RMS / Max | Fast before injection: variance / RMS / Max | Fast + fixed σ: variance / RMS / Max | Standard − Fast+Noise variance | Correlation magnitude |
|---|---:|---:|---:|---:|---:|
| 01 / transfer check | `1.054413e-22` / `1.026963e-11` / `4.701360e-11` | `5.531308e-25` / `7.437276e-13` / `2.205413e-12` | `1.076549e-22` / `1.037713e-11` / `3.371901e-11` | `-2.213650e-24` (`UNMODELLED_NEGATIVE_SLOT_VARIANCE_DELTA`) | `0.01436` |
| 02 / held out | `1.069198e-22` / `1.034121e-11` / `4.539547e-11` | `5.531308e-25` / `7.437276e-13` / `2.205413e-12` | `1.056653e-22` / `1.027940e-11` / `2.971252e-11` | `+1.254525e-24` (`NONNEGATIVE_SLOT_VARIANCE_DELTA_ONLY`) | `0.00983` |
| 03 / held out | `1.058062e-22` / `1.028754e-11` / `5.542996e-11` | `5.531308e-25` / `7.437276e-13` / `2.205413e-12` | `1.088151e-22` / `1.043396e-11` / `3.070904e-11` | `-3.008898e-24` (`UNMODELLED_NEGATIVE_SLOT_VARIANCE_DELTA`) | `0.01586` |

The Fast+Noise incremental RMS was respectively `+1.047%`, `−0.598%`, and `+1.423%` relative to its paired Standard result. Across the three descriptive vectors, mean RMS was `1.036350e-11` versus Standard `1.029946e-11` (about `0.62%` higher); mean variance was `1.073784e-22` versus `1.060558e-22` (about `1.25%` higher). Two of three paired variance deltas were negative, so the variance model is not consistently confirmed. Max errors were lower in all three Fast+Noise pairs than their paired Standard observations. Cross-backend correlation magnitudes were below `0.016`; this small per-vector result does not prove independent sampling.

**Coefficient-to-slot check:** For seeds 20261022/23/24, coefficient empirical variance was `16.179063` / `15.909019` / `16.321714`, empirical RMS `4.022504` / `3.988738` / `4.041306`, and sample count 8192 each. Observed decoded injection-delta variance was `1.070440e-22` / `1.052830e-22` / `1.080246e-22`; `N·Var(eta)/2^90` predicts `1.070641e-22` / `1.052771e-22` / `1.080080e-22`, within `0.019%` in each sample. This confirms the isolated coefficient-to-slot mapping at the tested Scale; it does not explain the Standard residual distribution or turn the variance subtraction into a general model. Q-prefix capacity was strict-fit for all 8192 coefficients across q0..q3 on all three calls; the sampler hash, capacity evidence, exact delta vectors, fingerprints and full provenance are in each run JSON and vector sidecar, inventoried in `manifest.json`.

Injection elapsed times were `1.305`, `2.063`, and `2.716 ms`; the measured MulRelin+Rescale operation phases excluding injection were `2.269`, `3.532`, and `5.365 ms`. These single-run timings imply diagnostic overhead of roughly `50.6%–58.4%` of the operation phase and are descriptive, not a performance campaign. Post-injection total oracle-error RMS was `1.037713e-11`, `1.027922e-11`, `1.043452e-11`; maximum total errors were `3.370794e-11`, `2.969736e-11`, `3.071152e-11`.

**CP-C outcome:** `RMSE_ONLY`. One unchanged σ brought the Fast incremental RMS close to the three genuine Standard observations and the decoded injection delta followed the coefficient-domain mapping. However, paired variance deltas were negative twice and positive once; CP-C does not establish distribution equivalence, a universal Gaussian model, or Bootstrap behavior. The fixed σ was not tuned. Results remain `DIAGNOSTIC_ONLY`; no security claim is made.

**Verification:** Both the full Fast `cmd/perfprobe` package and task-focused Standard/Fast tests passed; the focused tests include exact Fast-pin scoping, opt-in fail-closed rules, residual evidence recomputation and injected-delta tamper rejection. Secondary fixed-sigma/fastcore tests passed. Full Standard `cmd/perfprobe` tests reproduce the already-recorded unrelated historical A=0 failures and the build-info-relative-path test under the isolated Standard modfile. The task path policy accepted the exact generated artifact paths and its two unit tests passed. The helper is committed and pushed as Secondary `9ef0315feaf59dbee03bf0852b72c76fdeb6b800`; the CP-C harness/evidence checkpoints are Primary `a47e402...` and `e543cf8...`; the complete CP-D evidence, manifest and summary were committed and pushed as Primary `ba9c54ba9cea89e6e2a2278c28a08e87fd640dbe`. Bootstrap budget remains 0; LogN16, large benchmarks and BTS were not run.
