# FAST-DROPIN-ABOVE-CAP-NUMERICAL-AUTONOMOUS-BATCH-016

**Status:** READY_FOR_CODEX
**Research class:** M-frozen, bounded I/E, four checkpoints; one Web review or genuine evidence/mathematical blocker.
**Accepted predecessor:** `results/FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R-web-review.md`
**Standard genuine pin:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, independent unmodified checkout.
**Fast starting pin:** `xuejin-lu/lattigo fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` (read-only for first three checkpoints).
**Mission:** Independently verify **numerical** correctness of the already-integrated ordinary public CKKS Fast compact-Q arithmetic chain at logical Level **above the four-row cap**, unlike the prior structural-only Level 5 test.

## Frozen architecture

The exact same frontend source/public methods and effective parameters must run in genuine Standard and Fast by swapping only the Lattigo checkout. Fast is intentionally insecure zero-secret; no secret/security equivalence claims. CKKS Level/Scale must follow Standard semantics; Fast maintains `w_Q(Level)=min(Level+1,4)` real logical q residues. At Level 5, only q0..q3 are authoritative and q4/q5 are deliberately dormant. `docs/FAST_QPREFIX_SPEC.md` centered uniqueness `2B<S_Q(Level)` and operation-specific result/intermediate capacity checks are mandatory. **Never** reconstruct from stale q4/q5, fabricate full-Q values or silently fall back to Standard. Do not redefine q parameters, change E or modify public calls to force an outcome. Reuse the already-shared `fastcore` Add/MulRelin/Rescale/Automorphism implementations; no new math or duplicated kernels.

## Checkpoint A — ground a valid Level 5 numerical oracle design

1. Safe-sync Primary/Secondary, inspect `AGENTS.md`, Batch 015R review/runner, authoritative Secondary Q-prefix spec and source, and approved existing Fast capacity observer/CRT oracle tests.
2. Choose and freeze ONE small 6-Q-prime LogN13 Level-5 profile and small deterministic 16-slot input values, retaining Standard original frontend and actual Q/P identity across builds. Keep original Batch 015 Levels 1/3 metrics as frozen history, not silently change them.
3. Explain and instrument how decoded numeric output is independently obtained from a **truly compact** high-Level Fast ciphertext without normal full-active-Q decryptor reading dormant q4/q5. If a proven bounded centered reconstruction/materialization is required solely at the measurement boundary, use the already existing contract and clearly isolate/label that boundary; do not alter production arithmetic, add a hidden materialization fallback, or claim this proves general unrestricted ciphertext decoding.
4. Establish a **source-backed** independent `2B<S_Q` gate for all required input/intermediate/output states and logical transitions. Just checking the value returned by ordinary Decoding on full-backed ciphertext is insufficient; zero-filled dormant rows may silently corrupt a claimed oracle.
5. **STOP** if no mathematically supported independent high-Level decode/centered oracle exists within existing contracts; output an exact feasibility map, do not invent new semantics or reroute to full Q. This gate is more important than a numeric green status.

## Checkpoint B — identical public arithmetic chain at high Level

1. Implement a new Primary-only deterministic runner based on prior accepted 015R infrastructure. Run `AddNew → MulRelinNew → Rescale → RotateNew` with starting Level 5 in both original Standard and Fast dependency workspaces, the same frozen public CKKS frontend, genuine Standard keygen/encrypt/decrypt, and intentional Fast zero-secret public lifecycle.
2. Prove real q4/q5 dormancy and fixed-prefix authority through each Fast public operation; validate Level 5→4 and Scale under real q5 divisor, q0..q3 authoritative, zero c1 and no per-op KeySwitch/Galois-key lookup. Distinguish a source-controlled **measurement-only** bounded materialization/decoder from the evaluation path; no external Fast-specific call in the common frontend. If measurement requires specialized backend-specific hooks, keep them out of the claimed byte-identical frontend and label evidence accurately.
3. Compare 16 nonempty finite complex decoded Standard/Fast samples and independent cleartext oracle at every operation. Use strict, fixed justified tolerances and paired RMSE/max. If capacity checks require unachievable profile/values within this fixed scope, STOP with exact bound evidence, do not change original profile values to fit a desirable result or weaken `2B<S_Q`.
4. Include malformed-state fail-closed tests and a shared-kernel/row read dispatch audit. No Benchmark or Bootstrap.

## Checkpoint C — verify and report constraints

1. Check that the capacity observer is actually used (not merely declared) and tests cover near-boundary invalid inputs. Explain any gap between representational Level 5 and actual unique centered lift.
2. Validate no stale/dormant row reads, no P0 full-Q fallback, and no hidden normal evaluator arithmetic. Record original vs Fast exact backend SHAs, frontend/config/input hashes, key counts, NTT/Montgomery, Scale/Level progression, authoritative maintained row counts, per-operation capacity gates and per-checkpoint numerical statistics. No new algorithm change.
3. Run one Standard and one Fast compact deterministic profile; focused tests/vet on allowed source; no unlimited reruns or parameter sweeps.
4. If any computational correctness issue in existing Secondary kernels is *clearly evidenced* and repairable without changing math/contract, record exact minimal source-backed cause and **STOP for Web review rather than edit Secondary**. First aim is experimental proof, not another broad integration refactor.

## Checkpoint D — documentation and handoff

1. New Primary outputs: `results/FAST-DROPIN-ABOVE-CAP-NUMERICAL-AUTONOMOUS-BATCH-016-{summary.md,journal.md,evidence.json}` and a dedicated runner/test under `tools/`. Keep existing prior artifacts intact.
2. Codex self-review, focused tests/vet, `git diff --check`, safe commit/push of **Primary only**. Genuine Standard and Fast Secondary stay clean/pinned. Do not commit raw per-slot vectors, large CSVs, keys or huge logs; keep compact metrics plus provenance.
3. Terminal status `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` only if the high-level mathematical capacity and both independent numeric gates genuinely pass. `BATCH_BLOCKED_NEEDS_WEB_REVIEW` for a first true discrepancy or mathematical/source contract gap with exact evidence.
4. Do NOT autonomously start 017. No Bootstrap, benchmarks, LogN16 or performance sweeps this batch.

**Scientific boundary:** Batch 016 does not assert broad public bootstrap compatibility or Fast speedup. A later separate batch can connect the now-validated compact primitive chain with public Bootstrap at E32 using matched genuine Standard under a bounded expensive-run budget.
