# CURRENT TASK — BTS Fast public API coverage Batch 004

Task: `BTS-FAST-PUBLIC-API-COVERAGE-004`
Status: `NEEDS_WEB_REVIEW`
Authoritative spec: `specs/BTS-FAST-PUBLIC-API-COVERAGE-004.md`

Stop condition: both authorized Fast BTS Bootstrap calls are consumed. Attempt 02 reached EvalSine `T[6]` and failed in ordinary ciphertext `Sub` because component 0 lacked `q4`; do not run another attempt or change representation/scale behavior until Web review resolves the boundary.

Execute the authorized sequential checkpoints A–D in the spec. Preserve Task 003 and Task 002 results/journals unchanged. This Batch has a separate maximum of two actual Fast `bts.Bootstrap` calls, each requiring a durable reservation first. `external/BTS/` remains unchanged and must never be committed or pushed. Preserve the original Secondary dirty worktree; use only a clean independent checkout synced to the official GitHub origin.
