# CURRENT TASK — Q0123 exact-division microkernel pilot

Task: FAST-Q0123-RESCALE-EXACT-DIVISION-PILOT-006
Status: READY_FOR_CODEX
Authoritative spec: `specs/FAST-Q0123-RESCALE-EXACT-DIVISION-PILOT-006.md`

Prior FAST-Q0123-RESCALE-REALISM-OPT-PILOT-005 finished at `e9661df0f33b2650a5f1d7641d7690c267833d75` with scientific decision REJECT for its one invariant-precomputation candidate. Historical Task005 source state remained READY_FOR_CODEX despite the delivered result; that bookkeeping mismatch is superseded by this pointer. Do not rerun or iterate that rejected prototype.

When the user says `開始`, safe fetch/status/sync as described in AGENTS.md, then execute this **single bounded, source-justified** Task006: investigate exact 128/192-bit fixed-divisor reduction helpers for Q0123 Rescale, reuse existing Task004 benchmark patch/Task005 zero-c1 vs nonzero-c1 tests, attempt no more than one safe bit-exact implementation-only optimization in an isolated Secondary clone, and reject if the oracle or repeated uninstrumented benchmark gate fails. Do not create new tooling, perform full Bootstrap, modify/push real Secondary, or falsely claim secure Fast FHE. Finish with concise Primary evidence, normal fast-forward push, and update both task/spec status to READY_FOR_WEB_REVIEW. If no justified safe candidate, finish NO_SAFE_DIVISION_CANDIDATE without code churn. Only stop for demonstrated math/security/Git/provenance blockers.
