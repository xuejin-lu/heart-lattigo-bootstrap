# Current Task

Task: QPREFIX-PERF-OPT-003
Status: READY_FOR_CODEX

Specification:
`specs/QPREFIX-PERF-OPT-003-SOURCE-INTT-BATCHING.md`

Task class:
`P — Performance Repair`

Accepted parent:
- `results/QPREFIX-PERF-DIAG-007-summary.md`
- classification `POST_OPT_RESCALE_MIXED`
- current production implementation `6930cf6cb3c71ce139a1eb42eede7be335b7174c`

Accepted current evidence:
- full rows4 Rescale about `1.130242 ms`;
- source INTT phase `273.364 us` (22.21%);
- commit NTT phase `242.367 us` (19.69%);
- transforms combined 41.89%;
- source INTT stack about 23.39% cumulative.

Goal:
Change only the source-domain preflight transform order from component-major to row-major across ciphertext components, reusing evaluator-owned staged polys.

Do not:
- change Q-prefix width/policy;
- change CRT/round/capacity semantics;
- weaken transactionality;
- add goroutines, unsafe, assembly, or a new NTT algorithm;
- change P93/generated-power schedules.

Required classification:
- `SOURCE_INTT_BATCHING_READY`
- `SOURCE_INTT_BATCHING_CORRECT_BUT_NO_WIN`
- `SOURCE_INTT_BATCHING_BLOCKED`

Write:
`results/QPREFIX-PERF-OPT-003-summary.md`

Then report `READY_FOR_WEB_REVIEW`.
