# Task006 LogN16/E32 offline recursive trace revalidation

- Classification: `OFFLINE_RECURSIVE_TRACE_REVALIDATED`
- Original classification/status: `TRACE_COLLECTED_UNVERIFIED` / `TRACE_UNVERIFIED` (unchanged)
- New Bootstrap calls: **0**
- Event count: **539**; independent event roots: **38**
- Generated-power branch roots: `[{real 59 1.60157292e+08 1.60155917e+08 1375} {imag 279 1.6358475e+08 1.63583292e+08 1458}]`
- Raw SHA-256: `9bd5ba81abce75b8fd7c06ac544410246fb385bafb1585a646d7d71e2f8e872d`
- Failure-sidecar SHA-256: `f598b481b776448dcc78ff80083610e9376cee12451716ddb88fd414abfc756b`
- Task006 ledger SHA-256: `e752dbaf4aee28229e14cbcbc85c6c8c9f4b4467c0da3b5dbf29577b4e28672d`
- Captured Fast production / diagnostic adapter commits: `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` / `649da85757a97dc8c60b5f9cc031634f39ad4a40`
- Adapter patch digest matches committed patch: **true**
- Historical adapter commit object currently available: **false**

- Bootstrap root timing: inclusive 1060517791 ns; direct children 1060469666 ns; signed residual 48125 ns (0.00%)
- Existing analyzer detail counts: event-type/root records 539; positive-exclusive groups 38; Amdahl scenarios 9

## Nonoverlapping direct Bootstrap stages

The table contains only direct child stage spans from the Bootstrap root. Nested power and Rescale instrumentation roots are reported separately below and are not added to this wall-time table.

| Stage | Count | Inclusive (ns) | Inclusive (ms) | Share of Bootstrap root |
|---|---:|---:|---:|---:|
| coeffs_to_slots | 1 | 152656958 | 152.656958 | 14.39% |
| evalmod_imag | 1 | 382586417 | 382.586417 | 36.08% |
| evalmod_real | 1 | 375138083 | 375.138083 | 35.37% |
| mod_up_trace | 1 | 4629292 | 4.629292 | 0.44% |
| pack_n1_to_n2 | 1 | 85291 | 0.085291 | 0.01% |
| public_finalization | 1 | 73000 | 0.073000 | 0.01% |
| scale_down | 1 | 62750 | 0.062750 | 0.01% |
| slots_to_coeffs | 1 | 145065667 | 145.065667 | 13.68% |
| unpack_n2_to_n1 | 1 | 172208 | 0.172208 | 0.02% |

## Independent generated-power and Rescale diagnostic roots

These are separate instrumentation trees that overlap Bootstrap wall-time measurement; their durations must not be summed with each other or with direct stage spans.

Generated-power branches:

| Branch | Root sequence | Inclusive (ns) | Inclusive (ms) | Direct children (ns) | Signed residual (ns) |
|---|---:|---:|---:|---:|---:|
| real | 59 | 160157292 | 160.157292 | 160155917 | 1375 |
| imag | 279 | 163584750 | 163.584750 | 163583292 | 1458 |

Rescale diagnostic roots:

| Root sequence | Events | Inclusive (ns) | Inclusive (ms) | Direct children (ns) | Signed residual (ns) |
|---:|---:|---:|---:|---:|---:|
| 6 | 13 | 25697333 | 25.697333 | 25563000 | 134333 |
| 19 | 13 | 25699459 | 25.699459 | 25555750 | 143709 |
| 32 | 13 | 25637750 | 25.637750 | 25506084 | 131666 |
| 45 | 13 | 25624291 | 25.624291 | 25491125 | 133166 |
| 69 | 13 | 25620000 | 25.620000 | 25488625 | 131375 |
| 87 | 13 | 26226583 | 26.226583 | 26095083 | 131500 |
| 105 | 13 | 25576041 | 25.576041 | 25438541 | 137500 |
| 123 | 13 | 25634750 | 25.634750 | 25500126 | 134624 |
| 143 | 13 | 25834083 | 25.834083 | 25694875 | 139208 |
| 161 | 13 | 25603917 | 25.603917 | 25472375 | 131542 |
| 174 | 13 | 25624083 | 25.624083 | 25491959 | 132124 |
| 187 | 13 | 24061250 | 24.061250 | 23928209 | 133041 |
| 200 | 13 | 24263167 | 24.263167 | 24128125 | 135042 |
| 213 | 13 | 25594625 | 25.594625 | 25462958 | 131667 |
| 226 | 13 | 25621833 | 25.621833 | 25487333 | 134500 |
| 239 | 13 | 25562750 | 25.562750 | 25431583 | 131167 |
| 252 | 13 | 25569000 | 25.569000 | 25437749 | 131251 |
| 265 | 13 | 25621208 | 25.621208 | 25490375 | 130833 |
| 289 | 13 | 25711625 | 25.711625 | 25580208 | 131417 |
| 307 | 13 | 27033583 | 27.033583 | 26894583 | 139000 |
| 325 | 13 | 26161166 | 26.161166 | 26029750 | 131416 |
| 343 | 13 | 26095917 | 26.095917 | 25956167 | 139750 |
| 363 | 13 | 25856833 | 25.856833 | 25722667 | 134166 |
| 381 | 13 | 27091375 | 27.091375 | 26953250 | 138125 |
| 394 | 13 | 26302166 | 26.302166 | 26169208 | 132958 |
| 407 | 13 | 24624583 | 24.624583 | 24491625 | 132958 |
| 420 | 13 | 25194584 | 25.194584 | 25054541 | 140043 |
| 433 | 13 | 26215500 | 26.215500 | 26075875 | 139625 |
| 446 | 13 | 25790041 | 25.790041 | 25654750 | 135291 |
| 459 | 13 | 25695250 | 25.695250 | 25549917 | 145333 |
| 472 | 13 | 25812042 | 25.812042 | 25677334 | 134708 |
| 485 | 13 | 25708333 | 25.708333 | 25572542 | 135791 |
| 499 | 13 | 25312209 | 25.312209 | 25165000 | 147209 |
| 512 | 13 | 24749750 | 24.749750 | 24599291 | 150459 |
| 525 | 13 | 20162458 | 20.162458 | 20061291 | 101167 |

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
