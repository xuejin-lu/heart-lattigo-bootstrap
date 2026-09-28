# DIAG-FRAMEWORK-001-R1 結果摘要

**結果：`DIAGNOSTIC_FRAMEWORK_REPAIR_READY`**

**交接：`READY_FOR_WEB_REVIEW`**

## 提交

- Primary closure regression test: `255baf3872562db98d4d62d4f2e08ff76468aaa4`
- Secondary Rescale event hierarchy and shape test: `4783c641cee2df5f504b8033905a62a82971da28`
- Primary 本摘要另以結果 commit 加入；最終 Primary HEAD 由交接回報提供。

## 修正與 closure 證據

Secondary 不再同時輸出同父層的 `reconstruct_center_round_capacity` span 與累積 record。現在每個 component/pass 只有一個 `coefficient_loop` span；累積 reconstruction 與 residue-materialization records 是其子節點。Rescale arithmetic、preflight、transactional failure-before-mutation、Q-prefix policy 與 power schedule 均未改動。

Tagged p93-q55 trace 對每一筆 Rescale 驗證：root 恰有一個 `preflight` 與一個 `materialization`；各 pass 按 `Count` 對每個 component 恰有一個 `prefix_to_coefficient` 及一個 `coefficient_loop`；materialization 另有每 component 一個 `ntt_montgomery_restore`。每個 loop 恰有一個 reconstruction child；materialization loop 再恰有一個 `residue_materialization` child，沒有同語義 sibling duplicates。

代表性事件樹（該 Rescale 有兩個 components）：

```text
rescale
├── preflight
│   ├── prefix_to_coefficient (c0, c1)
│   └── coefficient_loop (c0, c1)
│       └── reconstruct_center_round_capacity (各一筆)
└── materialization
    ├── prefix_to_coefficient (c0, c1)
    ├── coefficient_loop (c0, c1)
    │   ├── reconstruct_center_round_capacity (各一筆)
    │   └── residue_materialization (各一筆)
    └── ntt_montgomery_restore (c0, c1)
```

Primary synthetic nested-event test 驗證 closure 僅扣直接 children：root delta/children delta `100/100`；preflight `50/50`；materialization `50/50`；兩個 coefficient loop 分別 `45/45` 與 `40/40`。各層 residual 均為 0；孫節點沒有被重複加到祖先 closure。

## Normal-build 與診斷 overhead

環境為 Apple M4、darwin/arm64、Go 1.26.4、`GOMAXPROCS=10`。accepted framework baseline 的普通 build median 為 117.664 ms、15,952 allocs/op。R1 candidate 以 `-benchtime=1s -count=3` 測得 `117.605454 / 116.969000 / 117.364532 ms`，median `117.364532 ms`（約 −0.25%）；allocs/op 三次皆為 15,951，未見新增 diagnostic allocation。普通 build binary symbol inspection 未發現診斷 hook/collector symbols；僅有 benchmark fixture 與無關 `fastDiagonal` symbol。診斷計時呼叫仍位於 compile-time-disabled branch 內。

一次 all-scope wrapper trace（warmup 1、5 次）Bootstrap median `219.199 ms`；以同 profile ordinary-build median `117.365 ms` 為參照，啟用 overhead 約 `+86.8%`。5/5 runs `numerical_match=true`，每次 1,001 events。這是診斷啟用成本的校正值；本修正未嘗試最佳化該成本。

## 驗證

- Primary：`go test ./cmd/fastdiag` — pass（包含 nested closure test）。
- Secondary normal build：`go test ./internal/fastdiag ./schemes/ckks/fast ./circuits/ckks/polynomial ./circuits/ckks/bootstrapping` — pass。
- Secondary tagged trace：`FASTDIAG_TRACE=stage,power,rescale go test -tags fastdiag ./circuits/ckks/bootstrapping -run '^(TestScopeParserAndUnknownScope|TestNestedEventsHaveStableParentAndSequence|TestUnselectedScopeEmitsNoEvent|TestFastDiagP93Q55Trace)$' -count=1` — pass。
- Wrapper smoke trace：`./scripts/fastdiag trace --profile p93-q55 --trace stage,power,rescale --warmup 1 --repetitions 5` — pass，並以 sample event tree 核對 closure 階層。
- Ordinary-build gate：`go test ./circuits/ckks/bootstrapping -run '^$' -bench '^BenchmarkFastDiagP93Q55Count1$' -benchmem -benchtime=1s -count=3` — pass，未見 >2% latency regression。
- `git diff --check` — pass。
