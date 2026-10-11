# CURRENT TASK — Current Fast LogN16 execution path and performance

Task: FAST-LOGN16-BACKEND-PATH-AND-PERFORMANCE-004
Status: READY_FOR_CODEX
Class: FAST PATH PROOF, MEASUREMENT PLATFORM REUSE, MATCHED STANDARD/FAST TIMING — NO ARITHMETIC OPTIMIZATION

**Authoritative specification:** `specs/FAST-LOGN16-BACKEND-PATH-AND-PERFORMANCE-004.md`

## Why this task

Task 002 measured good diagnostic RMSE, but LogN16's 32,768 complex values were bitwise identical at every public replay stage and final output; Task 003 audited source/vector independence, so this equality is real in the archived numeric evidence. **Equality does not prove optimized Fast execution.** The user requests a deeper test using our existing speed/profiling/allocations/stage measurement platform rather than more RMSE-only checks.

Source `circuits/ckks/bootstrapping/fast_c2s_compression.go` has a LogN13-only C2S compression gate. It **does not** imply that the rest of LogN16 Fast Bootstrap is skipped. Verify the actual ordinary `bootstrapping.NewEvaluator` dispatch, Fast kernel call paths and data-dependent eligibility at LogN16 before any new full Bootstrap.

## Mandatory staged gate

A. Zero Bootstrap: source-backed Fast dispatch and actual kernel-path evidence; shared platform inventory. Identify exactly what is LogN13-only versus generally Fast at LogN16. No Secondary modifications.

B. Zero Bootstrap: use shared `cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, `internal/numericalmetrics`, isolated genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and Fast `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f` builds. Preserve same ordinary CKKS frontend without `NewFastEvaluator` calls. Extend existing formal native-E32 public mode for LogN16 where source-supported, without changing frozen LogN13 fixture. Confirm matched input, Q/P, metadata, proper Standard encryption and Fast mode.

C. Only after A and B PASS: a **new separate, strictly bounded performance budget** of at most 16 full Bootstrap calls: LogN16 each backend one cold and five warm (12 total); LogN13 each backend one cold and one warm (4 total). All reserved persistently before each call; no retries, hidden warmups, duplicate older experiments, or unbudgeted full Bootstrap from tests/tracers. If a preflight cannot pass, run ZERO calls. Capture wall-clock, allocations, cold/warm, source/runtime path proof, shared stage timing where semantically justified, complex RMSE/max/SNR, and exact limitations. Genuine original Standard source is immutable; no Fast production math changes, workaround, noise/security equivalence claims or performance optimizations.

Output: new results/FAST-LOGN16-BACKEND-PATH-AND-PERFORMANCE-004.md/.json, and bounded-attempt journal; safe fast-forward commit/push and READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW. Keep the earlier 002 PARTIAL and 003 exact-zero audit results untouched. Follow both AGENTS.md and docs/BACKEND_TRANSPARENCY_CONTRACT.md / docs/MEASUREMENT_PLATFORM.md. 
