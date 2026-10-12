# CURRENT TASK — BTS No-Trace Warm and Numerical Batch 008

Task: `BTS-NOTRACE-WARM-AND-NUMERICAL-008`
Status: `AUTHORIZED_AUTONOMOUS_BATCH`
Authoritative spec: `specs/BTS-NOTRACE-WARM-AND-NUMERICAL-008.md`
Primary approved baseline: `f3eb665e35f15b91d5f290626fbb7a3bb360a95c`; the task-spec/pointer commits are authorized additions.

User objective: finish the **teammate's original BTS** (LogN16, 32,768 slots) **No-Trace Cold/Warm performance table**, then use the exact same measured executions' decoded Standard/Fast vectors for numerical comparison. Standard first, after bounded resource preflight for prior exit 137. One cold plus five subsequent warm Bootstrap calls per backend, max 6 per backend and one measurement process per backend, no hidden test/warm-up calls. Both series must be environment-comparable to calculate a speedup. The existing public-native Notion baseline is a different workload and cannot be substituted.

Read `AGENTS.md`, `docs/MEASUREMENT_PLATFORM.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md`, and the spec after mandatory safe fetch/fast-forward. Reuse maintained measurement tools, preserve external/BTS source and pinned pristine Standard and Fast math. Original Secondary checkout with five dirty files remains untouched. All earlier Task007 evidence/call counts remain immutable. Log and preserve any environment resource issue without asserting exit 137 means OOM.

Perform the spec's enumerated autonomous checks sequentially; follow mandatory coding self-review and only hand off at completed batch boundary or a genuine STOP escalation. User-facing updates in Traditional Chinese.