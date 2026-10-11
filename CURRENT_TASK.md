# CURRENT TASK — Original-input LogN16/E32 native oracle + performance

Task: FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005
Status: READY_FOR_CODEX
Authoritative spec: `specs/FAST-LOGN16-PUBLIC-NATIVE-CAPACITY-SAFE-PERFORMANCE-005.md`

## Scientific correction (2026-10-11)

The earlier 005 proposal to scale workload C by 15/16 **IS CANCELLED**. The user correctly questioned whether a conservative coefficient capacity bound should dictate a modified CNN-like plaintext. The original configs use `log_default_scale=45`, LogN16 effective `q0=36028797019488257`, and canonical tiny plaintext values. The old sufficient-coefficient `2B<q0` gate failed by `309374390400729`, but this means its pessimistic proof is unavailable, **not that an actual plaintext coefficient overflow or CKKS error was observed**.

Maintain the same unmodified LogN16/E32 A/B/C workload, original input/workload/config/Q/P hashes, q0, Scale, frontend operation chain and genuine Standard / current Fast backend pins. Keep exact old bound and warning. FIRST explain the mathematical bound assumptions and establish whether, for this specified profile, the complete eight-stage native public-operation output against independent plaintext oracle (including actual Level0 after DropLevelNew) can safely serve as an observed-workload gate. Never interpret centered residues alone as proof of no wraparound; preserve legacy LogN13 strict behavior. Do not implement a generic unsafe `--skip-capacity`.

Zero Bootstrap phase: source/static audit and matched genuine Standard/Fast formal E32, full 32768-slot preflight with numerical/metadata/oracle/provenance gates, plus real LogN16 Fast kernel-path evidence. If anything fails, do not invoke Bootstrap.

Performance phase only if zero-call gates pass: **TRANSFER original 004 allowance** (0/16 consumed so far), not an additional allowance: LogN16 Standard/Fast each 1 cold+5 warm; frozen LogN13 Standard/Fast each 1 cold+1 warm. Total ceiling remains 16; no implicit calls, repeats, profiler calls or unauthorized math/Secondary changes. Use shared perfprobe/fastdiag/perfmeasure/numericalmetrics, track timings/allocations and matched error; report true Fast dispatch and honest limitations.

Original 002/003/004 artifacts immutable. Output updated task-005 summary/JSON plus journal, safe commit/push and READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW. Continue normal Codex `開始` → Web `review` workflow.
