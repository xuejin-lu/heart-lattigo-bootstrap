# Fast-CKKS Codex Handoff

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`
- branch `main`

Secondary:
- `xuejin-lu/lattigo`
- branch `fast-ckks`
- committed base before finalization: `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Finalized Secondary candidate:
- committed SHA: `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- `origin/fast-ckks`: `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- local worktree: clean
- finalized commit contains 21 files total:
  - 19 previously tracked files modified from the starting base
  - 2 intentional new Q012 files:
    - `schemes/ckks/fast/q012.go`
    - `schemes/ckks/fast/q012_test.go`

Do not rewrite or discard the finalized candidate history.

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

## Finalized Secondary provenance

Starting dirty fingerprint before finalization:

`4d2567717bb0a88024d8330cf57db6a4a3b93a7faea2374ada6ac5161ec84764`

Starting state:
- branch `fast-ckks`
- committed HEAD `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`
- 19 previously tracked files modified
- 2 intentional untracked Q012 source/test files later included in the finalized commit.

Final state:
- commit `40532b4dce5c7eeae2db5b0b6f21be64801ce923`
- ordinary fast-forward push
- local/remote synchronized
- worktree clean.

## Finalization status

The LogN13/P93/Q012 candidate was audited, committed, and pushed by ordinary fast-forward.

Primary finalization classification:
`LOGN13_P93_Q012_MILESTONE_1E2_FINALIZED`

The `1e-2` system milestone is complete; the final precision target remains a later phase.

## Current task

None.

Primary `CURRENT_TASK.md` status:

`MILESTONE_1E2_COMPLETE`

No further Codex action is pending for this milestone.

## Prohibitions for this completed milestone

- no retroactive parameter tuning
- no threshold relaxation
- no force push/history rewrite
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign as part of this completed milestone
