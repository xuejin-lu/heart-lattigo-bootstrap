# CURRENT TASK — FIXED-SIGMA-CHAIN-002

Task: `FIXED-SIGMA-CHAIN-002`
Status: `AUTHORIZED_AUTONOMOUS_BATCH` — **NOT_RUN**; execute only on a future user `開始` in Codex.
Authoritative spec: `specs/FIXED-SIGMA-CHAIN-002.md`
Journal: `results/FIXED-SIGMA-CHAIN-002/journal.md`
Summary: `results/FIXED-SIGMA-CHAIN-002/summary.md`

**Previously completed:** `FIXED-SIGMA-MUL-NOISE-001` is independently reviewed as `ACCEPTED_SCOPED_RMSE_ONLY`; see `results/FIXED-SIGMA-MUL-NOISE-001/web-review.md`. DO NOT rerun it, re-fit sigma, or silently promote it to general Gaussian/noise equivalence.

**Next task:** With the same **frozen coefficient σ = 4.004280012097432**, use genuine Standard and current Fast to validate two consecutive complete ct×ct MulRelin+Rescale operations (first Level5→4, then Level4→3 only if verified), with Fast-only explicit opt-in injection after each complete operation. Record per-operation incremental residual and propagated total error; attribute the current helper's observed ~50–58% one-shot cost. **No Bootstrap** (0 authorized), no LogN16/BTS, no arbitrary global Rescale hook, no source/parameter/security shortcuts.

Follow Primary `AGENTS.md` startup fetch/safe ff-only/worktree preflight, then Secondary `fast-qprefix` instructions and the spec; reuse existing measurement tools and only make bounded, mathematically authorized changes. STOP `NEEDS_WEB_REVIEW` at any unsupported Scale/Level, capacity, API/injection boundary or new math decision. Existing completed 001 data and logs remain immutable.
