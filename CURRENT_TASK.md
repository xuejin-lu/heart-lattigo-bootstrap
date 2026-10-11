# CURRENT TASK — Zero-Bootstrap audit of LogN16 exact-zero evidence

Task: FAST-DUAL-LOGN16-EXACT-ZERO-EVIDENCE-AUDIT-003
Status: READY_FOR_CODEX
Class: READ-ONLY-FIRST INDEPENDENT ARTIFACT-INTEGRITY AUDIT; ZERO BOOTSTRAP CALLS

Authoritative spec: `specs/FAST-DUAL-LOGN16-EXACT-ZERO-EVIDENCE-AUDIT-003.md`

Previous task FAST-CURRENT-DUAL-BACKEND-NUMERICAL-REVALIDATION-002 completed as `DUAL_BACKEND_NUMERICAL_PARTIAL` at Primary origin/main 027929a491b7aa3dcec9135553d69806929c777a. Both independent binaries and public replay/final RMSE gates were reported passing, with LogN13 paired terminal RMSE 1.4468089544e-10 and LogN16 paired terminal RMSE **exactly zero** at all eight replay stages and final Bootstrap. LogN16 outputs vs original input both have RMSE 5.8081933491e-9 and max 4.708804915e-7. The zero result warrants a lightweight independent raw-artifact/hash check. Old Bootstrap budgets consumed: do NOT run any new Bootstrap, warmup, retries or full-suite tests implicitly invoking Bootstrap.

Scope: safe sync and read permanent rules; inspect original separately serialized Standard/Fast vector artifacts, SHA-256, original source/binary provenance, per-stage metrics, precise LogN16 bit equality, and missing-data limitations. If temporary raw files are gone, report NOT_INDEPENDENTLY_AUDITABLE; do not repeat Bootstrap. Add only narrowly scoped reusable comparator hash tests/fixes if demonstrably missing. No Secondary changes, no frontend changes, no performance work, no Standard/Fast internal trace fabrication, no upgrades to native-encrypted Standard claims (prior experiment used diagnostic manually encoded c0 and c1=0 on both backends). Preserve original reports and PARTIAL classification. Commit/push only authorized Primary audit artifacts and small integrity tests/fixes; READY_FOR_WEB_REVIEW or NEEDS_WEB_REVIEW according to AGENTS.md.
