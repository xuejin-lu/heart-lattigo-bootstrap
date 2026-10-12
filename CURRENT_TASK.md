# CURRENT TASK — BTS low-level Fast transparency integration

Task: BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002
Status: NEEDS_WEB_REVIEW
Authoritative spec: `specs/BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002.md`

Handoff: the single authorized Fast BTS invocation entered `bts.Bootstrap` and failed at C2S conjugation. The attempt is consumed; do not retry. See `results/BTS-LOWLEVEL-FAST-TRANSPARENCY-INTEGRATION-002-summary.md` and the durable journal.

The prior `BTS-LOGN16-NOTRACE-BASELINE-001` is historical and preserved; do not rerun its Standard baseline or alter its results/logs. Its failed Fast attempt is not reusable as an attempt in this separately authorized task. This task permits one target No-Trace Fast BTS run and a hard maximum of two actual Fast `bts.Bootstrap` invocations, with a durable task-scoped journal and fail-closed reservation.

`FAST-Q0123-RESCALE-EXACT-DIVISION-PILOT-006` remains suspended. BTS source under `external/BTS/` must remain unchanged and must never be committed or pushed.
