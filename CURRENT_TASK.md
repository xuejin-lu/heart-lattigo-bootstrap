# Current Task

Task: FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004
Status: READY_FOR_CODEX
Task class: M-approved invariant + bounded Secondary implementation (I) and conditional single-pair experiment (E)

**Executable spec:** `specs/FAST-DROPIN-ZERO-SECRET-ENCRYPT-BRIDGE-004.md`
**Approved scientific assessment:** `results/FAST-DROPIN-API-LIFECYCLE-AUDIT-003-web-review.md`
**Long-term goal:** `docs/FAST_DROPIN_ZERO_SECRET_GOAL.md`

## Previous audit accepted as partial

`FAST-DROPIN-API-LIFECYCLE-AUDIT-003` produced the **same unchanged source**, successfully compiled and preflighted against genuine Standard and Fast dependency checkouts. But both versions of ordinary native `EncryptNew` still created nonzero c1 under a nonzero secret. The Fast `GenEvaluationKeys` and `NewEvaluator` already dispatch to intentional zero-secret bootstrap, which cannot faithfully decode a native ciphertext by ignoring nonzero `c1*s`. Thus stopping before Bootstrap was appropriate. The audit is accepted as PARTIAL, not as drop-in success. Primary report `441ee378690de9ca8567520c8fedf9ad00aaa1af`.

Two AGENTS Git-sync corrections are verified committed and pushed: Primary `49d2a9ece038c0f58cfa77d0e7addfa056eb0a20` and Secondary `68c3da8ec18222f985ebb5c43f6a8495b853cfc2`.

## Active bounded repair

Implement in the Fast fork only an **internally CKKS-scoped** zero-secret input encryption bridge: the ordinary public `rlwe.NewEncryptor(ckksParams, sk).EncryptNew(pt)` yields logically `(encoded pt, 0)`, with identical signatures and frontend params; other schemes remain unaffected. Maintain metadata and native DecryptNew numeric interpretation. **Do not clear c1 from a normally encrypted ciphertext**. If an operation has nonzero c1, do not silently treat it as zero-secret. Respect Q-prefix support and fail clearly if needed.

Add an unmodified-source Standard/Fast integration frontend under Primary and preflight both builds. Only after correct preflight, at most **one Standard E32** plus **one Fast E32** public Bootstrap; compare decoded outputs and original. Report exact results; no benchmarking, no parameter tuning, no claims of cryptographic security or universal API compatibility.

Use safely synced repos and their AGENTS, tests, self-review, and permitted fast-forward pushes. Return `FAST_DROPIN_ENCRYPT_BRIDGE_COMPLETE_PENDING_WEB_REVIEW`, `FAST_DROPIN_ENCRYPT_BRIDGE_PARTIAL` or `FAST_DROPIN_ENCRYPT_BRIDGE_BLOCKED`, and `READY_FOR_WEB_REVIEW` or `NEEDS_WEB_REVIEW`.

`FAST-STANDARD-PERF-REBASELINE-003` remains BLOCKED.
