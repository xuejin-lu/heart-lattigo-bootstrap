# Current Task

Task: FAST-INTEGRATION-002
Status: NEEDS_BENCHMARK_EVIDENCE

Candidate Secondary implementation:
`57ffb88744c82778c0a9392ecab394e19f712a3d`

Specification:
`specs/FAST-INTEGRATION-002-FUSED-MODUP-BRIDGE-OPTIMIZATION.md`

Independent Web code review:
- fused production kernel is present;
- production `modUpBasis` no longer materializes transient private-F state;
- standalone private-F Import/ModUp/compact-export APIs remain as semantic oracle;
- canonical q0 midpoint semantics are preserved;
- only maintained logical rows are materialized;
- input is non-mutating;
- no blocking correctness defect found in source review.

Acceptance is blocked only on the mandatory performance evidence.

Codex must, on the current candidate commit, run the same LogN13 benchmark pair used for FAST-INTEGRATION-001 at least three times each:
- `BenchmarkFastModUpBasisLogN13`
- `BenchmarkStandardModUpBasisLogN13`

Report for every run:
- ns/op
- B/op
- allocs/op

Acceptance gates:
- fused Fast allocs/op <= 873;
- B/op materially below the Integration-001 ~3.35 MB/op baseline;
- ns/op below the Integration-001 ~4.50 ms/op baseline.

Do not modify code unless a gate fails and the existing task permits one bounded repair pass.
If all gates pass, report `READY_FOR_WEB_REVIEW`.
If any gate fails after the bounded repair allowance, report `NEEDS_WEB_REVIEW` with the measurements.
