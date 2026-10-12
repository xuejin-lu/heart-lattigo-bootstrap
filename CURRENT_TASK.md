# CURRENT TASK — BTS Fast Decrypt/Decode 007

Task: `BTS-FAST-DECODE-007`
Status: `NEEDS_WEB_REVIEW`
Authoritative spec: `specs/BTS-FAST-DECODE-007.md`

The remaining paragraph records the immutable Batch 006 historical state only. It does not constrain Task 007's separate one-Fast-call budget.

This user-authorized Batch prioritizes completing the unchanged LogN16 BTS on the intentionally insecure Fast zero-secret backend, then comparing decoded output with genuine Standard. The user's experimental capacity-proof waiver supersedes only Batch 005's capacity-certificate stop for this Fast zero-secret mode; Batch 005 history/evidence remains unchanged. Record missing prefix-bound proof as `CAPACITY_UNPROVEN_EXPERIMENTAL`, not as a correctness proof. Task 007 Fast budget: 1/1 Bootstrap entered and completed; decoded output and snapshot were validated. One Genuine Standard process was launched but terminated with exit code 137 after `NewBTS()` and before `bts.Bootstrap()`; Standard Bootstrap entered: 0/1 and no Standard vector exists. No retry is being made in this Batch. Standard/Fast numerical comparison is therefore incomplete; see `results/BTS-FAST-DECODE-007/summary.md` and `results/BTS-FAST-DECODE-007/journal.json`. Preserve the original Secondary dirty worktree and use the independent clean checkout.
