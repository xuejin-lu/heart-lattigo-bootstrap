# CURRENT TASK — Independent Standard/Fast numerical revalidation

Task: FAST-CURRENT-DUAL-BACKEND-NUMERICAL-REVALIDATION-002
Status: READY_FOR_CODEX
Class: PRIMARY MEASUREMENT-BACKEND ISOLATION + FRESH BOUNDED LOGN13/LOGN16 NUMERICAL VALIDATION

Authoritative spec: `specs/FAST-CURRENT-DUAL-BACKEND-NUMERICAL-REVALIDATION-002.md`

## Review decision

Prior task FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001 finished as LOGN16_CURRENT_REVALIDATION_BLOCKED; its results and consumed 2 Fast + 1 in-tree Standard API LogN16 Bootstrap attempts remain unchanged.

Confirmed root cause: 'Standard' evaluator compiled from Fast checkout inherits FastCKKSZeroSecretSimulation and wrong Fast Q-prefix Rescale, so C2S matrix 1 sees empty active q4 and panics. This is not genuine Standard. DO NOT alter Secondary math or mode dispatch to work around bad experimental separation.

Mandatory: safe Primary/Secondary startup sync; both AGENTS.md, docs/MEASUREMENT_PLATFORM.md; use two clean independent pinned Lattigo checkouts with temporary modfiles, genuine Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` and current Fast `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`. Prove backend source, input/Q/P provenance and full Standard C2S limb backing with ZERO Bootstrap calls first. Reuse shared platform; no new one-off framework or old Q01/P93 tests.

Only after zero-call gates PASS, the new task authorizes at most **ONE fresh genuine Standard + ONE fresh Fast full Bootstrap per profile**, LogN13 and LogN16 (four fresh calls total; old calls remain spent). No hidden warmups/retries or performance runs. Record all comparable public-stage RMSE, terminal Fast-vs-Standard and input-oracle RMSE/SNR, accurately classify missing genuine Standard internal traces. Fail closed on source confusion. Report newly measured values, commit/push only permitted Primary artifacts, READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW on new blockers.

No performance optimization authorized.
