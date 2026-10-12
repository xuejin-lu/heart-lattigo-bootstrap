# BTS-FAST-PUBLIC-API-COVERAGE-004 — Batch journal

- Status: `BATCH_IN_PROGRESS`; charter/pointer commit pending initial Primary push.
- Active checkpoint: A (source-backed BTS public API coverage audit).
- Task 003 remains complete and immutable: its one Fast Bootstrap call panicked in C2S; no attempt/result/journal reused or edited.
- Batch 004 actual Bootstrap calls: **0 / 2**; no reservation exists, and no Bootstrap has been run for this Batch.
- Primary at startup: `main` `2bb79d9c8a9d69fa91f76c263eec24f0b3912a4f`, clean and equal to fetched `origin/main`.
- Original Secondary: `fast-qprefix` `0b922dc48315128a27db811a5f237e0a6c9813d6`, five known dirty CKKS files preserved; current tracked diff SHA-256 equals the external safety copy `a052b913ee6e575d88b3cff5416ac21477e563bbbcd3536ad7014ab2f4f64a60`.
- Independent Secondary: `/private/tmp/lattigo-bts-003-working.HnjCjy/repo`, branch `fast-qprefix`, commit and fetched `origin/fast-qprefix` both `049ffa782729d049e08f973b5852ee1e832a55f3`, clean; official remote is `git@github-xuejin:xuejin-lu/lattigo.git`.
- Known Task 003 finding to verify/repair: ordinary `ckks.Evaluator.Mul(ctImag, -1i, ctImag)` entered the generic full-Q scalar routine and panicked on an unbacked Q row in `ring.Ring.MulDoubleRNSScalar`. Existing explicit Fast Q-prefix complex-scalar kernel is in `schemes/ckks/fast/evaluator_ntt.go`; implementation architecture remains subject to Checkpoint A dependency/representation review.
- Next action: push this Primary charter/pointer/journal by normal fast-forward, then continue A–C in the clean independent Secondary checkout. No Bootstrap until all gates pass and Secondary repair is pushed.
