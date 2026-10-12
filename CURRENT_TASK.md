# CURRENT TASK — BTS No-Trace Warm and Numerical Batch 008

Task: `BTS-NOTRACE-WARM-AND-NUMERICAL-008`
Status: `READY_FOR_WEB_REVIEW`
Authoritative spec: `specs/BTS-NOTRACE-WARM-AND-NUMERICAL-008.md`
Primary approved baseline: `f3eb665e35f15b91d5f290626fbb7a3bb360a95c`; the task-spec/pointer commits are authorized additions.

User objective: finish the **teammate's original BTS** (LogN16, 32,768 slots) **No-Trace Cold/Warm performance table**, then use the exact same measured executions' decoded Standard/Fast vectors for numerical comparison. First complete the existing-platform BTS measurement adapter and zero-Bootstrap tests, then run Standard first and Fast second. Resource telemetry (unused RAM, free Swap, `memory_pressure`, and the prior Task007 memorystatus event) must be recorded but is advisory and cannot alone cause `RESOURCE_BLOCKED`. Real process failures still stop the batch and consume their reservations; do not bypass macOS protection or alter parameters. One cold plus five subsequent warm Bootstrap calls per backend, max 6 per backend and one measurement process per backend, no hidden test/warm-up calls. Both series must be environment-comparable to calculate a speedup. The existing public-native Notion baseline is a different workload and cannot be substituted.

Read `AGENTS.md`, `docs/MEASUREMENT_PLATFORM.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md`, and the spec after mandatory safe fetch/fast-forward. Reuse maintained measurement tools, preserve external/BTS source and pinned pristine Standard and Fast math. The old Secondary five-dirty-files note is historical: the user confirmed those changes were upstreamed and cleared; current `fast-qprefix` is clean at `ee308c9743000b5f281a84a871e7c4bd59b176c4`. All earlier Task007 evidence/call counts remain immutable. Record resource telemetry without treating it as an automatic gate, and do not infer OOM solely from exit 137.

Batch 008 is complete: Standard and Fast each used one process for exactly one cold + five warm calls; see `results/BTS-NOTRACE-WARM-AND-NUMERICAL-008/summary.md` and its run/evidence artifacts. No additional Bootstrap call is authorized by this completed batch.

Perform the spec's enumerated autonomous checks sequentially; follow mandatory coding self-review and only hand off at completed batch boundary or a genuine STOP escalation. User-facing updates in Traditional Chinese.
