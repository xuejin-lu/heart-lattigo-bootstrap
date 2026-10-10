# CURRENT TASK — LogN16 current numerical revalidation

Task: FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001
Status: READY_FOR_CODEX
Task class: FRESH NUMERICAL CORRECTNESS VALIDATION ONLY — NO PERFORMANCE BENCHMARK OR ARITHMETIC MODIFICATION

**Authoritative spec:** `specs/FAST-LOGN16-CURRENT-NUMERICAL-REVALIDATION-001.md`
**Permanent rules:** both AGENTS.md; Primary docs/MEASUREMENT_PLATFORM.md; existing numerical/provenance measurement tools.

User requires fresh execution with latest Secondary Fast on LogN16 and LogN13 control. Do not present historical LogN16 RMSE 2.569e-3 (EvalMod) or 2.0485e-2 (final) as if still current. Latest fast source has removed the prior LogN13-only Chebyshev post-product selection; whether it fixes LogN16 requires an actual new numerical run.

On START/開始: safe sync both repos; read this updated current task/spec; repair old hardcoded Fast SHA in shared Primary numerical tool via an explicit current profile; use same canonical deterministic inputs and original effective Q/P primes for LogN16/13; run bounded fresh matched Fast/Standard numerical test, verify all stage and EvalMod internal RMSE and original-message reference, then commit/push reports and say READY_FOR_WEB_REVIEW.

**Prohibited:** speculative scaling/normalization workaround, Secondary production math changes, runtime/performance benchmark, reusing expired previous batch timing budgets, accepting metadata-only smoke test as numerical validation.

## Archived previous task (completed)

The former Batch027 completed as HOTSPOT_NOT_RESOLVED with artifacts at results/FAST-E32-RESCALE-CAUSAL-HOTSPOT-AUTONOMOUS-BATCH-027-summary.md. Its source-preserved fixture and results remain unchanged. It is superseded by this user-requested new correctness check.
