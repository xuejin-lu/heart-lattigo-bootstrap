# Independent Web Review — FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015

**Date:** 2026-10-10 Asia/Taipei
**Disposition:** **INTEGRATION ACCEPTED PROVISIONALLY; FINAL NUMERICAL EVIDENCE NOT ACCEPTED — PRIMARY HARNESS BUG.**
**Primary submitted:** `6ffea5c6ab87a4888ac947170327449b2732cb95` on `main` (observed remote HEAD).
**Fast Secondary submitted:** `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` on `fast-qprefix` (observed remote HEAD).
**Genuine Standard comparison:** pinned `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (reported clean separate checkout; Web did not inspect local worktree).

## What remains valid

GitHub source confirms the Secondary submission touches the approved CKKS Fast public/explicit adapters and shared `schemes/ckks/internal/fastcore` multiplication/rescale kernels. The committed 015 report and runner describe/implement ordinary CKKS `AddNew → MulRelinNew → Rescale → RotateNew` with identical frontend source and original Lattigo Standard pin vs intentionally insecure Fast zero-secret dependency. Public and explicit Fast share underlying import-neutral cores; Fast uses Q-prefix row policy `w_Q(Level)=min(Level+1,4)`, zero-c1 and keyless supported operations, with no demonstrated full-Q fallback on dormant prefix rows.

The report's separate **plaintext-oracle** values are recorded for eight checkpoint states at start Levels 1 and 3. The reported state/scale/Level and focused tests/vet are useful bounded evidence **as reported**, not independently re-executed by Web. Level-5 checks are structural-only; they do NOT prove >cap centered numeric correctness. No Bootstrap, benchmark or LogN16 was executed.

The pre-submit P1 issues (narrow sourceRows `RescaleTo` no-op authority and malformed `MulRelinNew` panic) have corresponding code changes and negative tests in Secondary; do not redo this accepted integration solely to repair the external evidence pipeline.

## Independently reproduced logical defect in committed Primary source

File: `tools/fast-dropin-compact-consumers-batch-015/main.go`, commit `6ffea5c...`.

`combine()` currently calls `summaryOf(standard)` and `summaryOf(fast)` **before** comparing their decoded samples. `summaryOf` takes `runEvidence` by value but `Checkpoints` is a slice, so its loop `result.Checkpoints[i].Decoded = nil` mutates the same underlying checkpoint elements as the original `standard` / `fast` inputs. The subsequent `compare(valuesFromEvidence(sc.Decoded), valuesFromEvidence(fc.Decoded))` therefore compares two empty arrays.

`compare()` sets `Finite = len(want)==len(got)`, does not reject `len==0`, and returns a finite zero RMSE/max for empty arrays. With unequal lengths, it may also panic when indexing `got[i]`. Consequently the committed `FAST-DROPIN-COMPACT-CONSUMERS-AUTONOMOUS-BATCH-015-evidence.json` values `fast_vs_standard_rmse = 0` and `fast_vs_standard_max_complex_error = 0` for all eight checkpoints are **INVALID**, not proof of exact agreement.

Proof of contradiction in existing aggregate numbers (shared deterministic target): at Add Level 1, Standard-vs-plaintext RMSE is 8.570052293144194e-13 and Fast-vs-plaintext RMSE is 6.366054749149383e-14, implying by reverse triangle inequality a Standard-vs-Fast RMSE of at least approximately 7.933446818229255e-13, **not zero**. This is evidence/reporting corruption; there is currently **no basis to infer an algorithm failure** from this defect.

## Review decision and authorization

Do **not** approve full Batch 015 numerical provenance until a narrowly-scoped Primary-only `FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R` task:

1. fixes slice aliasing (compare before any lossy compact summary, and ensure summary conversion cannot mutate its source);
2. rejects empty/mismatched/nonfinite decoded comparisons instead of accepting zero or panicking;
3. adds regression tests that *fail* on the 015 bug, including nonzero known difference, zero-length, mismatch and summary immutability;
4. runs the exact same unchanged public frontend/profile/input against genuine original Standard and current Fast once each to produce **fresh** 8-point decoded paired metrics; validates CKKS plaintext gates and backend identity, no threshold gaming;
5. writes a distinct 015R corrected report/evidence, explicitly supersedes ONLY the invalid direct-paired columns of historical 015 aggregate, and leaves original 015 artifacts intact.

**Execution scope:** Primary runner/tests/new 015R evidence only; Secondary and pinned Standard read-only. Zero Bootstrap, benchmark, LogN16. Web acceptance review after repair; no new Fast primitive design or architectural changes in 015R.
