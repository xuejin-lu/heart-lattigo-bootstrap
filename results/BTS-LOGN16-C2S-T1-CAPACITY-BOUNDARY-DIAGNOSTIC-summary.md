# LogN16 BTS C2S/T1 Capacity Boundary Diagnostic

**Classification: `T1_SOURCE_EVIDENCE_MISSING`**

The single authorized S2C-through-T1 diagnostic completed. It did not call `bts.Bootstrap()`, EvalMod, T2/T3, or Standard BTS. No concrete capacity counterexample was established. The centered T1 observation is below the requested threshold, but it is not a proof: the conservative no-wrap chain first fails during S2C stage 3, and the true T1 has only Q0123 backing after the C2S real split.

## Provenance and execution budget

- Profile: LogN16, LogSlots 15, 32768 slots, SlotLimbs 15, `BTSModePlain`, seed 0, amp 1; S2C starts at Level 3 and C2S at Level 29.
- Fast backend: `d1b6dc9963b0df95a7c149306fb62660e1020f7b` (tree `3e8804cde94200e67506d1a0dafe8186c5218998` is recorded in the journal; archive SHA-256 `79744e1094ab09253d23c3e1d3284d69222ca365d2d98062c6ef473002f3650e`).
- Final diagnostic binary SHA-256: `e52da31805b79b316153e9583acb00f8c521124992b0d5543d3e4170b8de8a99` (`go1.26.4`, darwin/arm64, isolated `-modfile` replacement to the pinned backend source).
- Original BTS Go/module source manifest: 12 files, SHA-256 `6c1e081968b0ac8b3897a69e62efba02c8774dbf0bea5d32b87f29c10ad310f0`. Original full-file manifest: `030cc0bbff410599b1d2dfa65a7582ae6249cb7b9dd05b0db9d2c49852241c46`; original `btstest` SHA-256: `0908513a6df542922e5355964af8695bf109fe70611663e89d2f9a5eddbae813`. `external/BTS/` is unchanged.
- One test-only source-IFFT setup invocation failed **before S2C** because it passed arbitrary-precision values to a double-precision encoder; its log is preserved separately. The isolated test was corrected to mirror the actual default `complex128` encoder path, rebuilt, and the zero-computation preflight passed. Exactly one S2C calculation then ran. Total: 1 S2C calculation, 0 Bootstrap, 0 EvalMod/T2/T3, 0 Standard BTS.

## Capacity evidence

The seeded input was matched exactly against the original encoder output; its unmodded coefficient bound is `39,477,590` (worst coefficient 34,951), with zero `c1`. S2C stages 1 and 2 pass the propagated strict full-Q and maintained-prefix bounds, and their observed rows agree with the complete-Q representatives.

The first proof gap is S2C stage 3, raw pre-Rescale at Level 1:

| Quantity | Value |
|---|---:|
| Conservative raw coefficient bound | `232076387351619387697099559` |
| `q0*q1` | `18884847496112318709761` |
| Strict centered capacity `(q0*q1-1)/2` | `9442423748056159354880` |
| `2*bound - q0*q1` | `464133889855742663075489357` |
| Bound / strict capacity | about `24578.05×` |

The observed centered full-Q maximum at that checkpoint is `18442236704021432251`, and the post-Rescale centered observation is `4294966417`. These centered observations do not reveal whether an earlier integer value wrapped; therefore they are not a no-wrap proof or a counterexample. The post-Rescale conservative bound is `54047689868142566`, which also does not strictly fit q0.

ScaleDown and ModUp both used integer multiplier `1`; the row-by-row centered q0 ModUp lift verified. C2S ran using the actual BTS stage sequence, but its propagated capacity bounds are unavailable because the S2C proof chain had already failed. The C2S stage outputs were physically full-Q; the real split output and T1 were Q0123-only, so no independent full-Q T1 observation exists.

`Step2_Offset` applied exactly `-1/64`. The measured coefficient-domain delta was exactly `-68719476736` at coefficient 0, with one changed coefficient; the Q-prefix row hashes are unchanged by `Step1_Relabel`. The observed centered T1 Q-prefix maximum is `68717802446`, below the requested `354567853947127`, but `T1ProvenUpperBound` is absent and the upstream no-wrap chain is incomplete. It must not be reported as `T1_BOUND_PROVEN`.

## Artifacts

- [Persistent attempt journal](BTS-LOGN16-C2S-T1-CAPACITY-BOUNDARY-DIAGNOSTIC-journal.json)
- [Compact evidence and checkpoint hashes](BTS-LOGN16-C2S-T1-CAPACITY-BOUNDARY-DIAGNOSTIC-evidence.json)
- [Pre-S2C setup failure log](BTS-LOGN16-C2S-T1-CAPACITY-BOUNDARY-DIAGNOSTIC-attempt-001-raw.log)
- [Single S2C/T1 run log](BTS-LOGN16-C2S-T1-CAPACITY-BOUNDARY-DIAGNOSTIC-attempt-001-s2c-raw.log)

The next decision is for Web review. No production code, BTS source, parameters, task pointer, or Secondary worktree was modified.
