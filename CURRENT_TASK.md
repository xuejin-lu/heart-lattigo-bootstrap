# Current Task

Task: FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020
Status: READY_FOR_CODEX
Mode: FOUR CHECKPOINTS, BOUNDED STANDARD/FAST PERFORMANCE QUALIFICATION
Web review: once at milestone completion or immediately on fair-measurement blocker

**Runnable spec:** `specs/FAST-DROPIN-PUBLIC-CHAIN-PERFORMANCE-AUTONOMOUS-BATCH-020.md`
**Accepted previous milestone:** `results/FAST-DROPIN-HIGHLEVEL-MUL-RESCALE-BOOTSTRAP-AUTONOMOUS-BATCH-019-web-review.md`
**End goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`
**Workflow:** both AGENTS.md, `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4A/4B and Secondary FAST_QPREFIX_SPEC.md

## Accepted 019 and the next decision

Batch019 accepted bounded identical-public-source true Standard vs zero-secret Fast LogN13/E32 full-Q chain: Level5 `Add→MulRelin→Rescale(logical q5)→Rotate→DropLevel(4)` using four physical authoritative prefix q rows in Fast with q4+ dormant, then Level0 native decrypt/Scale2^45 and public E32 Bootstrap/Level1 native decode. Input C uses actual Scale=q5 in both builds, while A/B use default 2^45 and all canonical Q/P, E and default Scale unchanged. Post-Bootstrap pair RMSE `4.937529503163558e-9`, max `5.038821901648196e-8`; max-gate 1e-6; each backend one Bootstrap, no benchmark. This is workload-specific numeric proof, NOT an execution time comparison.

## Batch020 A→D

A: source/provenance-backed fairness design, controlled same-source stage timing including keygen/encryption/arithmetic/Bootstrap/decrypt, excluding build time and diagnostic observer instrumentation. Freeze repetitions/cold-warm and RSS/allocation methods before running.
B: small bounded matched Standard/Fast non-Bootstrap timing with identical deterministic input/profile, preserving plaintext/Level/Scale/q-authority gates; show raw cheap sample dispersion.
C: only after A/B pass, MAX one real Standard + one Fast E32 Bootstrap, record timed one-shot wall + allocations/RSS; never claim statistical speedup from a single call. No retry/warmup Bootstraps.
D: compact numerical + timing evidence, explicit limitations, focused tests/vet, clean ordinary Primary commit/push; independent Web review. STOP on apples-to-oranges setup or mathematical/provenance blocker rather than fabricating performance.

**Pins:** genuine original Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Fast `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, both read-only. Primary only may add one 020 runner/test/results.
**Budget:** maximum one Standard + one Fast Bootstrap, zero backend algorithm changes, zero LogN16, no sweeps, no unapproved subsequent task.
