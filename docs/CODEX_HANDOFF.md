# Fast-CKKS Codex Handoff

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed base before finalization: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- local worktree intentionally dirty with the fully validated LogN13/P93/Q012 candidate.

Finalized Secondary candidate:
- committed SHA: `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- `origin/fast-ckks`: `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- local worktree: clean

Never reset, stash, clean, discard, checkout-overwrite, rebase, pull across, reconstruct, or otherwise alter the validated dirty candidate.

## Startup

When the user says `開始`, follow `AGENTS.md` preflight, synchronize Primary, then read synchronized `CURRENT_TASK.md` and its exact spec.

## Milestone evidence

Primary repair commit:
`91a255ba0366e8c83fbc88e70a7c76f23b9f1508`

Classification:
`P93_POSTPRODUCT_REPAIR_1E2_SYSTEM_PASS`

Accepted:

- generated powers `T2/T3/T4/T6/T8/T16` all align with exact Chebyshev oracle at about `1e-15` or better and are Q012-safe;
- polynomial:
  - real `1.8570138426987626e-8`
  - imag `1.6219059983946238e-8`
  - both pass `3.716228023823462e-8`;
- DA2:
  - real `1.5967815852787189e-7`
  - imag `1.3046858460163482e-7`
  - both pass `3.0517578125e-7`;
- internal final:
  - real `1.635104343335875e-4`
  - imag `1.3359983063118935e-4`
  - both pass `3.125e-4`;
- public EvalMod:
  - real `0.0052323338986748`
  - imag `0.004275194580198059`
  - both pass `1e-2`;
- exact full Fast production E2E:
  `0.009705381393898434`
  - passes `1e-2`;
- Genuine Standard control:
  `5.830057349387463e-8`.

Therefore the current `1e-2` system milestone is passed end-to-end.

Important: this is not the final research precision target. Tightening toward ~`1e-7` is a later phase.

## Validated Secondary final dirty state

Required starting fingerprint for finalization:

`4d2567717bb0a88024d8330cf57db6a4a3b93a7faea2374ada6ac5161ec84764`

Expected:
- branch `fast-ckks`
- committed HEAD `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- 19-file accumulated Fast-CKKS candidate
- 418 insertions / 146 deletions.

## Finalization status

The LogN13/P93/Q012 candidate was audited, committed, and pushed by ordinary
fast-forward. The `1e-2` system milestone is complete; the final precision
target remains a later phase.

## Current task

`specs/FIX-001-P3-FINALIZE-LOGN13-P93-Q012-SECONDARY-CANDIDATE-COMMIT-PUSH.md`

Completed with Primary `CURRENT_TASK.md` status `MILESTONE_1E2_COMPLETE`.

This task must:

1. audit the complete 19-file Secondary dirty diff;
2. reject any unexpected/unrelated change;
3. safely inspect remote `origin/fast-ckks` without pulling across the dirty tree;
4. run final Secondary regression tests and authoritative Primary exact-E2E confirmation;
5. if all pass, commit the entire validated Secondary candidate;
6. push `fast-ckks` by fast-forward only;
7. verify remote/local synchronization and clean Secondary worktree;
8. record the new Secondary commit SHA in Primary;
9. set `CURRENT_TASK.md` to `MILESTONE_1E2_COMPLETE`.

No new algorithmic work is authorized.

## Two-word workflow

1. Codex: `開始`
2. Codex audits, validates, commits/pushes Secondary safely, then updates/pushes Primary evidence/state.
3. ChatGPT Web: `review`
4. Orchestrator confirms milestone finalization.

## Prohibitions

- no new algorithmic changes
- no parameter tuning
- no threshold relaxation
- no force push
- no reset/stash/clean/discard
- no history rewrite
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign
