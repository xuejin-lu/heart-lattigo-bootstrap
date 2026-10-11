# Task006 LogN16/E32 offline recursive trace revalidation

- Classification: `OFFLINE_RECURSIVE_TRACE_REVALIDATED`
- Original classification/status: `TRACE_COLLECTED_UNVERIFIED` / `TRACE_UNVERIFIED` (unchanged)
- New Bootstrap calls: **0**
- Event count: **539**; independent event roots: **38**
- Generated-power branch roots: `[{real 59} {imag 279}]`
- Raw SHA-256: `9bd5ba81abce75b8fd7c06ac544410246fb385bafb1585a646d7d71e2f8e872d`
- Failure-sidecar SHA-256: `f598b481b776448dcc78ff80083610e9376cee12451716ddb88fd414abfc756b`
- Task006 ledger SHA-256: `e752dbaf4aee28229e14cbcbc85c6c8c9f4b4467c0da3b5dbf29577b4e28672d`
- Captured Fast production / diagnostic adapter commits: `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` / `649da85757a97dc8c60b5f9cc031634f39ad4a40`
- Adapter patch digest matches committed patch: **true**
- Historical adapter commit object currently available: **false**

- Bootstrap root timing: inclusive 1060517791 ns; direct children 1060469666 ns; signed residual 48125 ns (0.00%)
- Existing analyzer detail counts: event-type/root records 539; positive-exclusive groups 38; Amdahl scenarios 9

## Recorded numerical outputs (metadata only; no decrypt repeated)

| Call | Level | Scale log2 | Degree | Q-prefix rows | Oracle RMSE | Oracle max | Task005 Fast-reference max | Gate |
|---|---:|---:|---:|---:|---:|---:|---:|---|
| first_cold_bootstrap | 1 | 45 | 1 | 2 | 5.82126708e-09 | 4.66946557e-07 | 0 | true |
| warm_bootstrap_01_traced | 1 | 45 | 1 | 2 | 5.82126708e-09 | 4.66946557e-07 | 0 | true |

## Recursive power ancestry

The corrected shared validator accepts the captured recurrence chain `generated_powers#59 → T16#60 → T8#61 → T4#62 → T2#63`; the old direct-child-only rejection is not reinstated. Stage timings and event-type summaries are partitioned by independent roots, so nested inclusive times are not treated as disjoint work.

## Limitations
- Analysis reuses the maintained Batch025 recursive event validator and root-partitioned analyzer; it does not rerun Bootstrap, decrypt, or change the immutable raw trace.
- The historical diagnostic adapter commit object is absent from the current Secondary clone. Its recorded commit and patch digest match Task006 metadata, and the committed patch digest is independently verified; the event attribution is certified from captured trace data, not a fresh source-tree rebuild.
- Task005 reference vector/result hashes are recorded provenance; the historical temporary vector payloads are not embedded in this repository and are not reconstructed here.
