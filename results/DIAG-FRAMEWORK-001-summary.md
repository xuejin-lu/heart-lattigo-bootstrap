# DIAG-FRAMEWORK-001 結果摘要

**結果：`REUSABLE_DIAGNOSTIC_FRAMEWORK_READY`**
**交接：`READY_FOR_WEB_REVIEW`**

## 提交

- Primary implementation：`9662cd15d1a8c54e3ee618731f6f397fd45ca213`
- Secondary instrumentation：`6312e8b9a982a405709a5124041eec569b0ae980`
- Primary 本摘要另以結果提交加入；最終 Primary HEAD 由交接回報提供。

## 實作與範圍

Secondary 新增 compile-time `fastdiag` build tag 與 `internal/fastdiag` event/span collector，並在實際 Fast Bootstrap、Chebyshev power generation、Q-prefix Rescale 路徑加入 `stage`、`power`、`rescale` scopes。一般 build 的 `Enabled=false`；不讀診斷環境變數、不計時、不收集事件。普通 build test binary 的 symbol 檢查未發現診斷 collector/hook symbols（命中項僅測試 fixture 與無關的 `fastDiagonal`）。

Primary 新增 `scripts/fastdiag` trace/compare wrapper、bounded `p93-q55` runner、median/delta/closure aggregation、安全檢查、測試與使用文件。支援 refs 不含 hooks 時回報 `DIAGNOSTIC_HOOKS_UNAVAILABLE_AT_REF`，不 patch 歷史原始碼。Self-review 另修正此情況下 summary 將缺失 Rescale 資料誤顯示為零的問題，並加入 regression test。

未改 Q-prefix 算術、Rescale transactional 行為、P93 schedule、參數、power DAG 或 evaluator/backend 選擇語義；變更只增加診斷 metadata/timing callsites 與測試。

## Performance 與 trace evidence

環境：Apple M4、darwin/arm64、Go 1.26.4、`GOMAXPROCS=10`。普通 build q0=55 Count-1 Bootstrap，pre-task 與 candidate 皆為 1s × 5 benchmark runs：

| 模式 | Median latency | Median B/op | Median allocs/op |
|---|---:|---:|---:|
| pre-task | 117.663546 ms | 7,615,616 | 15,952 |
| candidate, diagnostics off | 117.712699 ms | 7,615,617 | 15,952 |

latency 差 `+0.042%`，低於 `2%` gate；allocs/op 中位數相同，沒有新增 persistent per-operation diagnostic allocation。B/op 中位數差 1 byte。

同一 `p93-q55` workload（4096 個輸入值，SHA-256 `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`），warmup 1、重複 5 次：

| Trace mode | Bootstrap median | 相對 diagnostics-off candidate |
|---|---:|---:|
| diagnostics off benchmark | 117.713 ms | — |
| `stage` | 117.338 ms | −0.32%（落在測量雜訊範圍） |
| `stage,power,rescale` | 210.841 ms | +79.11% |

All-scope trace 每次含 1,001 個事件；5/5 次 `numerical_match=true`。Stage-only trace 每次含 10 個事件。全 scope 的高 overhead 主要來自逐係數 Rescale 診斷計時，解讀細粒度 timing 時應將其納入校正。

Compare 範例以同一 Secondary commit 作 baseline/candidate，兩側皆 `READY`、產生 1,001 個 event deltas；Bootstrap delta 為 `+0.125 ms`，Rescale totals 為 `196.163 ms` / `196.158 ms`。此 self-comparison 用於驗證 compare/aggregation 管線，不代表版本效能結論。以 pre-hook `e3f232ecd8ecb9c5c8682269eb7bfd63df326372` 比對新 commit 時，baseline 正確回報 `DIAGNOSTIC_HOOKS_UNAVAILABLE_AT_REF`，未動態修改舊 source。

## 驗證

通過：

- Secondary focused tests：`go test ./internal/fastdiag ./schemes/ckks/fast ./circuits/ckks/polynomial ./circuits/ckks/bootstrapping`
- Secondary tagged trace tests：`FASTDIAG_TRACE=stage,power,rescale go test -tags fastdiag ./circuits/ckks/bootstrapping -run '^(TestScopeParserAndUnknownScope|TestNestedEventsHaveStableParentAndSequence|TestUnselectedScopeEmitsNoEvent|TestFastDiagP93Q55Trace)$' -count=1`
- Primary CLI/tests：`go test ./cmd/fastdiag`
- Ordinary-build performance gate：`go test ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkFastDiagP93Q55Count1$' -benchmem -benchtime=1s -count=5`（candidate 與 pre-task worktree 配對執行）
- Wrapper trace：`./scripts/fastdiag trace --profile p93-q55 --trace stage --warmup 1 --repetitions 5` 及 `--trace stage,power,rescale`
- Wrapper compare：`./scripts/fastdiag compare --profile p93-q55 --baseline <ref> --candidate <ref> --trace stage,power,rescale --warmup 1 --repetitions 5`
- `git diff --check`

Primary `go test ./...` 曾只因既有、已記錄且與本任務無關的 `TestFIX001P3GenuineStandardPublicVsStagedConsistency` 失敗，classification 為 `P93_GENUINE_STANDARD_BASELINE_REPLAY_CONFLICT`（見 `docs/CODEX_HANDOFF.md` 與 FIX-001 P3 audit 文件）。本任務未修改該路徑；新 Primary tests 與 Secondary task-specific tests 通過。
