# CURRENT TASK — BTS Fast Decrypt/Decode 007

Task: `BTS-FAST-DECODE-007`
Status: `IN_PROGRESS`
Authoritative spec: `specs/BTS-FAST-DECODE-007.md`

The remaining paragraph records the immutable Batch 006 historical state only. It does not constrain Task 007's separate one-Fast-call budget; Task 007 permits one new Fast call only after its gates pass.

This user-authorized Batch prioritizes completing the unchanged LogN16 BTS on the intentionally insecure Fast zero-secret backend, then comparing decoded output with genuine Standard. The user's experimental capacity-proof waiver supersedes only Batch 005's capacity-certificate stop for this Fast zero-secret mode; Batch 005 history/evidence remains unchanged. Record missing prefix-bound proof as `CAPACITY_UNPROVEN_EXPERIMENTAL`, not as a correctness proof. Fast budget: 3/3 entered (1 Bootstrap returned a ciphertext; original decrypt panicked), so no further Fast call is authorized. Conditional new Standard budget: 0/1, not used because no Fast decoded vector exists. Full BTS and Standard/Fast numerical comparison are incomplete; see the task summary and journal. Preserve the original Secondary dirty worktree and use an independent clean checkout.
