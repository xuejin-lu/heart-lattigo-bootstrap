# Independent Web Review — 015R / Final Batch 015 Numerical Acceptance

**Date:** 2026-10-10 (Asia/Taipei)
**Decision:** **ACCEPT Batch 015 + 015R AS A BOUNDED PUBLIC CKKS FAST Q-PREFIX PRIMITIVE-CHAIN INTEGRATION MILESTONE.**
**Primary submission:** `main@16c16ed9f54e8b6f18d37e261f99fe09d74d7770`; verified GitHub remote HEAD at review.
**Secondary unchanged:** `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`; verified GitHub remote HEAD.
**Genuine Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (separate original baseline, locally clean as reported by Codex; Web cannot inspect local worktrees).

## Independent inspection

Reviewed committed 015R `summary.md`, `journal.md`, compact `evidence.json`, full `tools/fast-dropin-compact-consumers-batch-015/main.go`, its added `main_test.go`, Secondary 015 source integration, frozen public goal/Secondary Q-prefix spec, and earlier 012/014/015 reviews. Web inspected source/repository evidence but **did not personally execute Go tests or both backend runs**. Thus numeric claims are verified against committed evidence and source-backed method, not independently regenerated raw per-slot values.

The 015R commit modifies **only Primary evidence runner and tests** plus three NEW 015R artifacts. It leaves previous original 015 report/journal/evidence intact and does not modify Fast or Standard dependency implementation.

## Root cause fixed

- Previous `combine()` summarized checkpoint slices before direct comparison. Its `summaryOf()` loop mutated the original underlying `Checkpoints` elements and cleared their decoded vectors. `compare([],[])` then wrongly returned finite zero.
- New `combine()` first validates both backend runs, checks exact eight checkpoint IDs, state, 16 nonempty finite decoded samples per checkpoint, frozen source/Q/P/input/profile hashes, expected backend commits and metadata; **then** computes direct pairwise metrics from real vectors. Only afterwards does it construct compact summaries.
- New `summaryOf()` copies the checkpoint slice before clearing `Decoded`, and `compare()` rejects empty/mismatched/nonfinite inputs rather than returning a passing zero or indexing past the slice. Regression tests cover source immutability, positive known differences, legitimate equal-vector zero, empty and malformed vectors, checkpoint-count mismatch and provenance mismatch.
- The eight historical 015 `Fast-vs-Standard RMSE=0/max=0` columns are INVALID; the new **015R** aggregate supersedes **only those pair columns**, not historical 015 evidence in place.

## Numerical and architecture outcome

- Same unchanged public CKKS frontend with `AddNew → MulRelinNew → Rescale → RotateNew`, in genuine Standard and intentionally insecure Fast zero-secret builds; fixed LogN13, Q bit sizes [55,39,40,39], P [60], scale 2^45, 16 slots, input Levels 1 and 3.
- All eight fresh paired checks have `16` decoded complex samples per backend, a finite nonzero direct Fast-vs-Standard RMSE and max. Largest paired RMSE = **7.788804037185317e-13** (Add L3); largest max difference = **1.41288837464095e-12** (Add L3). Each backend independently satisfies original plaintext gates, with matched Level/Scale/degree/NTT/Q-row progression. The new direct RMSE results are mathematically compatible with distinct Standard/Fast plaintext RMSE; the previous impossible zero values are gone.
- Fast supported pairwise MulRelin/Rotate performs zero direct `GetRelinearizationKey` / `GetGaloisKey` lookups. `GetGaloisKeysList` is enumerated once at evaluator construction; do not describe this as *no key metadata access ever*.
- Public Fast operations reuse the already existing Fast algorithm in import-neutral shared core, with fixed Q-prefix row policy `w_Q(level)=min(level+1,4)`, keyless zero-secret c1, and no evidenced native full-RNS fallback on dormant rows. Confirmed in earlier 015 source review and unchanged in this 015R.
- Level-5 six-Q-prime compact composition is **structural only**: q4/q5 stay dormant. It is NOT an independent decoded high-level mathematical/numerical proof. No Bootstrap, benchmark or LogN16 run was conducted here. Bootstrap ephemeral-secret parameter E is N/A to this primitive-only chain; this does not validate E32 Bootstrap or a full frontend swap for every API overload.

## Acceptance boundaries / next required science

This milestone establishes public-API **bounded primitive-chain correctness** and real Fast core dispatch; it does **not** establish (1) numerical correctness of actual logical Levels above the cap q3, (2) a complete compact-Q-to-public-Bootstrap workflow, (3) unrestricted CKKS public API support, (4) measured acceleration, or (5) any cryptographic security equivalence.

**Next authorized charter:** `FAST-DROPIN-ABOVE-CAP-NUMERICAL-AUTONOMOUS-BATCH-016`. First obtain an independently capacity-certified **Level>3** numerical proof with dormant q4/q5 and identical public frontend, without running Bootstrap or performance measurements. A separate later stage will connect an already validated compact primitive chain to existing public Fast Bootstrap, with a workload-specific E and proper native Standard comparison.
