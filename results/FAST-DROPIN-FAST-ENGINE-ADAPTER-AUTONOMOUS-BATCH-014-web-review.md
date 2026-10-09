# Independent Web Scientific Review — FAST-DROPIN-FAST-ENGINE-ADAPTER-AUTONOMOUS-BATCH-014

**Date:** 2026-10-10 (Asia/Taipei).
**Decision:** **ACCEPT CHECKPOINT B AS A REAL FAST ADD/SUB BRIDGE; KEEP BATCH 014 BLOCKED AT THE FIRST COMPACT DOWNSTREAM CONSUMER.** This is not full CKKS drop-in, complete compact-Q composition, or speedup evidence.

## Verified repository provenance

- GitHub remote `xuejin-lu/heart-lattigo-bootstrap main` pointed to `1df4b085702b1c0b99f4fbaaec80da6333f1125a` during independent review; the commit adds the dedicated runner/test and three Batch 014 artifacts.
- GitHub remote `xuejin-lu/lattigo fast-qprefix` pointed to `0c92e7ed071a1be6eb258474811db470e4463df7`; its diff is limited to CKKS public Add/Sub dispatch, existing explicit Fast Add/Sub wrappers, shared `fastcore/addsub.go` and focused tests/allocator helpers. No genuine Standard baseline source was modified. Genuine Standard pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40` is externally referenced; remote source exists, but Web has not verified developer-local worktree cleanliness.
- Reviewed committed `014-summary.md`, `014-journal.md`, aggregate `014-evidence.json`, and relevant Secondary production diffs plus existing `MulRelin`, `Rescale`, Rotate and explicit Fast primitive sources. Web **did not rerun Go tests, vet or numeric experiments**. Passing test commands/worktree cleanliness are Codex-reported; actual source and committed artifact content independently inspected.

## What 014 proves

1. Ordinary public `ckks.Evaluator.Add/Sub` ciphertext/ciphertext Fast zero-secret dispatch now invokes the extracted `schemes/ckks/internal/fastcore.AddSubCore`. Existing explicit `fast.Evaluator.AddQPrefixRows/SubQPrefixRows` delegates to the **same** core, rather than retaining a second copied algorithm. No Go parent/child import cycle was introduced.
2. Public `AddNew/SubNew` can allocate compact output with logical Level >= Q-prefix cap, retaining only `min(Level+1,4)` backed Q rows, with no frontend Fast constructor. Level-5 structural test proves q0..q3 are backed and q4/q5 deliberately dormant; it is **not** a Level-5 decoded arithmetic-oracle proof.
3. Matching identical public frontend/profile/input hashes are recorded for genuine Standard and Fast paired runs. Six Level-1/3 `AddNew`, `SubNew` and `Add→Sub` checkpoints pass their reported complex plaintext oracles. Fast-vs-Standard RMSE ranges approx 5.12e-13 to 8.18e-13, maximum difference <=1.61e-12. Neither benchmark nor Bootstrap nor LogN16 was run.
4. At Level 5, first attempted downstream public `MulRelin` enters `mulRelinFastCKKSZeroSecret` and fails its *full-active-Q* validator on absent component-0 q4; separate `Rescale` and `Rotate` checks fail their own full-Q validation. This is a correct fail-closed stop, **not evidence that the existing explicit Fast algorithms mathematically cannot consume compact-Q input**.

## Architecture interpretation

Authoritative Secondary `docs/FAST_QPREFIX_SPEC.md` says `w_Q(ell)=min(ell+1,4)`, dormant higher rows are intentionally non-authoritative, Rescale uses logical top-`q_ell` via bounded centered CRT, and compact Fast work must never secretly fall back to full-RNS. At Level 5, demanding materialized q4/q5 in every internal primitive defeats this contract and is **not** the approved repair.

The next engineering step is **integration dispatch/adapter migration**: existing `fast.Evaluator.MulRelinElementQPrefixRows` delegates to `mulElementRows` (source `schemes/ckks/fast/evaluator.go`, `fast/evaluator_ntt.go`), while `fast.Evaluator.RescaleQPrefixRows` and `AutomorphismQPrefixRows` exist. Adapt or move their existing kernels into import-neutral internal code so unmodified public calls use the **same** compact operations. Do not duplicate or replace the previously proved CRT, keyless multiplication, or rotation algorithms. Enforce the capacity/Level/Scale/row-provenance rules already prescribed in constitution. A concrete newly uncovered capacity failure must be escalated with exact stage/values, not assumed.

**Additional incomplete-contract note:** The new `fastCKKSZeroSecretAddSubEligible` selector permits some full-backed nonzero-c1 or unsupported cases to take the ordinary generic path. For a strict transparent zero-secret Fast backend this must be inventoried; never claim that all public overloads are now optimized Fast, or silently route invalid/dormant Fast state through generic full-RNS. Rejection or verified explicit allowed handling is preferred over a hidden fallback. No unrelated broad refactor is approved here.

## Decision and next charter

Accept the isolated 014 Add/Sub implementation as a bounded integration milestone; retain its reports as immutable evidence. **Do not classify Batch 014 as BATCH_COMPLETE.** Approve the bounded next work `FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015`, focusing first on public compact-Q `MulRelin`, then previously implemented Fast Rescale and Rotate consumers, ending in a small unchanged-frontend composition if arithmetic/representation guards pass. Stop for a **new actual** mathematical or provenance conflict. No Bootstrap, benchmark, or LogN16.
