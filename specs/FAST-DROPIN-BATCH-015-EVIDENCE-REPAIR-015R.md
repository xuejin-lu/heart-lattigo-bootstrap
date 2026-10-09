# FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R

**Status:** READY_FOR_CODEX
**Task class:** Primary-only bounded evidence/harness repair, 3 checkpoints; one final Web review.
**Scientific review:** `results/FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-web-review.md`.
**Genuine Standard (read-only):** Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
**Fast Secondary (read-only):** `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
**Original submitted Primary:** `6ffea5c6ab87a4888ac947170327449b2732cb95`; repository head may also contain the Web review/this spec/task pointer.
**Resource cap:** 0 Bootstrap, 0 benchmark, 0 LogN16; two matched small runs (one per dependency) plus narrowly focused regression tests.

## Nonnegotiable result

Repair the factual integrity of **Fast-vs-genuinely-encrypted-Standard decoded numerical comparisons** in Batch 015. Retain exactly the original public CKKS frontend, parameters, actual Q/P primes, 16-slot deterministic inputs, operation order, Levels 1/3, and tolerances. Neither Fast nor Standard Lattigo source, APIs, library pins, keys semantics, encryption path, or mathematical algorithms may be modified. Do not rerun Bootstrap.

Current Primary `tools/fast-dropin-compact-consumers-batch-015/main.go` has two coupled defects:
- `combine()` calls `summaryOf(standard/fast)` before comparing `sc.Decoded` and `fc.Decoded`; `summaryOf` mutates underlying `Checkpoints` slice elements, so pairwise comparisons see **empty** arrays.
- `compare()` accepts two empty arrays as finite, zero-RMSE comparison; unequal lengths risk panic.

The stored eight `fast_vs_standard_rmse=0` and max=0 values in historical `...015-evidence.json` are **INVALID**. This does NOT imply Fast arithmetic failed. Do not change historical 015 result, journal, or summary in place.

## Checkpoint A — regression-first source repair

1. Safe sync Primary, re-read `AGENTS.md`, this task, 015 Web review and current runner. Check exact current refs and that Secondary/Standard remain pinned, clean and untouched before runs.
2. Add focused `*_test.go` for the combined evidence functions, first demonstrating the old failure condition via fixtures: distinct *nonempty* decoded arrays produce strictly positive RMSE and max; pairwise result must be calculated with unchanged decoded inputs; `summaryOf` must not mutate its input; two empty inputs are **invalid**; mismatched lengths are **invalid without panic**; NaN/Inf are invalid; equal-length identical nonempty values are valid zero; wrong checkpoint shape/count and provenance are rejected.
3. Fix `summaryOf` to produce a deep-independent checkpoint slice (or equivalently separate lossy summary from raw provenance; not merely reorder one call). Calculate paired metric while decoded values still exist. Harden `compare` to avoid indexing mismatched inputs, reject zero samples and nonfinite values. `combine()` must fail closed on any pair with absent/empty/mismatched decoded vectors, not just on nonfinite result.
4. Make no changes to arithmetic, workload, source values or numerical tolerances. Run just the new focused test once after implementation; self-review aliasing, empty arrays, panic paths.

## Checkpoint B — fresh matched evidence

1. Independently use the existing original genuine Standard checkout and Fast Secondary at the pinned commits with read-only validation; no library writes. Run **the same patched public frontend runner** against each backend with the same frozen LogN13 profile, genuine Standard keygen/EncryptNew/DecryptNew and Fast zero-secret flow, original Q/P, E and input hash. Do not use the old raw pair data, which were lossy in committed aggregate.
2. Run patched combiner on the two complete new raw outputs **before** stripping decoded arrays. Report all eight actual pairwise complex RMSE/max; ensure values are finite, nonempty and gate-matched, plus Standard/Fast individually pass original plaintext oracle tolerances.
3. Check scale/Level/degree/zero-c1/row authority status and Fast key-lookup counts remain as intended. Raw decoded arrays should not be committed; compact results only.
4. Publish `results/FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R-summary.md`, `...-journal.md`, `...-evidence.json`, with current source and runtime provenance hashes. Clearly label invalid historical 015 *pairwise columns* as superseded. Do NOT quietly rewrite old `015-evidence.json`; preserve forensic audit trail.
5. If the corrected pairwise comparison unexpectedly fails, STOP with the actual metric values and first reproducible issue, do not loosen gates or modify library.

## Checkpoint C — self-review and submission

1. Run focused Primary tests and vet relevant package(s), `git diff --check`; inspect changed files for accidental scope, floating-point sample count and aliasing regressions.
2. Ensure all emitted evidence labels are truthful: no claim of exact Standard/Fast equality without numeric basis; Level 5 remains structural-only; zero Bootstrap, benchmark, LogN16.
3. Commit and normal fast-forward push Primary `main` under standing safe-push rules; do not modify/push Secondary or Standard. Return `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` on A+B+C pass. Otherwise `BATCH_BLOCKED_NEEDS_WEB_REVIEW` with precise evidence.
4. Do not autonomously begin 016 or performance experiments. Leave worktrees clean if the authorized work completes.

**STOP:** unsafe Git state; mismatch to pinned backend/profiles/source hashes; unknown mathematical difference or oracle failure; inability to produce nonempty matched raw vectors; scope expansion into Fast/Standard code; more than one bounded repair pass for a persistent defect.
