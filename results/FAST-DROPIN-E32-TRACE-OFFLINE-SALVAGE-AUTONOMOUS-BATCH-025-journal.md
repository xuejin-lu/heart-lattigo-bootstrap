# Batch025 offline salvage journal

- Status: `OFFLINE_REVALIDATED_TRACE`
- Generated: `2026-10-10T17:44:28.166696Z`
- New Standard/Fast Bootstrap calls: **0**
- Original raw SHA-256: `f922587fa944da7ef07dcb94a9497505018f2fee2bc67e7f87af6d35bc3d1c74`

## Checkpoints

| Checkpoint | Result | Evidence boundary |
|---|---|---|
| O1 immutable evidence/source audit | PASS | Raw/sidecar/manifest/vectors hashes, source pins, attempt journals and source identity recorded in evidence JSON. |
| O2 offline replay and full-tree validation | PASS | 539 events; original raw remains `TRACE_UNVERIFIED`; no trace/runner/Bootstrap call. |
| O3 timing and closure analysis | PASS | 38 independent roots; separate-root Pareto and signed closures in evidence JSON. |
| O4 artifacts | PASS | Summary, journal, evidence and platform documentation are task outputs; output files are create-only. |

## Attempt accounting

- Attempt 01 `irrevocably_reserved_before_call`: `2026-10-10T17:09:08.231611Z`, budget=2, phase=`diagnostic_fast_cold`, journal SHA-256 `88866058b84fb0c0e1fecb531cba9494453fdcda88139813723179448184c9bf`.
- Attempt 02 `irrevocably_reserved_before_call`: `2026-10-10T17:09:08.232464Z`, budget=2, phase=`diagnostic_fast_traced_warm`, journal SHA-256 `ab23db35f3ed89830af21f2b619caba1efd390a7fbee790cfa99c219c95572ce`.

Captured budget/outputs: 2/2 calls; newly invoked calls: 0.

No offline validation anomalies. The raw capture is immutable and still unverified as a live trace artifact; only this separate offline replay receives its classification.
