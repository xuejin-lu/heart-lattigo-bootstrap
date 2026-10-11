# CURRENT TASK — Offline recover already-captured LogN16 Fast stage trace

Task: FAST-LOGN16-TRACE-OFFLINE-RECOVERY-007
Status: READY_FOR_CODEX
Authoritative spec: `specs/FAST-LOGN16-TRACE-OFFLINE-RECOVERY-007.md`

Task006 captured **539 real Fast LogN16/E32 trace events** and passed two native terminal numerical oracles, but its test-only validator repeated the **historical Batch024 mistaken direct-child assumption** for recursive generated powers. Event #61 T8 correctly has parent #60 T16, whose parent is #59 generated_powers. Task005 exhausted 16/16 Bootstrap calls, Task006 exhausted 2/2; **NO NEW BOOTSTRAP CALLS ARE AUTHORIZED FOR TASK007**.

Read existing Task006 raw trace/sidecar and compact report, verify raw path+SHA, 005/006 ledgers and provenance, then reuse the **Batch025 accepted offline recursive generated-power / event-tree analysis** to validate the *already captured* LogN16 trace. Diagnose/fix the minimal duplicated validator error in Primary task-owned test adapter without altering Secondary production math. Do NOT regenerate, rerun or recollect the trace. If raw temp files are absent, STOP with RAW_TRACE_UNAVAILABLE; do not run a bootstrap.

Output new compact Task007 Markdown/JSON with offline status and meaningful in-circuit Bootstrap stage, generated-power and Rescale breakdowns if actually verified; parent/child stage shares must be nonoverlapping while independent power/Rescale roots are analyzed separately. Combine descriptively with old 005 uninstrumented LogN16 Standard 6425.276 ms, Fast 663.878 ms and old 006 traced warm 1060.783 ms; do not claim comparable Standard in-circuit stage traces or stage RMSE without an oracle. Preserve Task005/006 artifacts and call ledgers, Secondary stays clean. Use existing perfprobe/fastdiag/perfmeasure/numericalmetrics, not a new harness. Safe commit/push Primary, READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW. Codex trigger remains `開始`.
