# CURRENT TASK — BTS Fast No-Trace validation 003

Task: BTS-FAST-NOTRACE-VALIDATION-003
Status: READY_FOR_CODEX
Authoritative spec: `specs/BTS-FAST-NOTRACE-VALIDATION-003.md`

This is an independent, one-call follow-up. Preserve task 002 and its attempt journal unchanged; do not reuse its failed call. The new task allows at most one actual Fast `bts.Bootstrap` invocation, after code validation and Secondary push, with its own durable reservation journal. Do not rerun Standard or perform warmup/Trace/benchmark calls.

`FAST-Q0123-RESCALE-EXACT-DIVISION-PILOT-006` remains suspended. BTS source under `external/BTS/` must remain unchanged and must never be committed or pushed.
