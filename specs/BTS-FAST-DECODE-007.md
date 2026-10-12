# BTS-FAST-DECODE-007 — Finish Fast BTS and Compare Numerical Outputs

Status: `IN_PROGRESS`
Class: bounded backend implementation and experiment (I/E)
Supersedes: the unresolved Fast decrypt/decode blocker from Batch 006 only. Batch 006's three Fast attempts, evidence, status, and journal remain immutable.

## Objective and fixed sources

Make the unchanged BTS LogN16 workload complete the ordinary public Fast decrypt/decode lifecycle using an explicitly experimental Q-prefix plaintext projection, then compare all 32,768 decoded slots with a genuine Standard BTS output.

- Primary baseline: `6abd11de9214aba524b7ababf52980ffde915510`
- Fast Secondary baseline: `e05283690fcb0262540a2f979b570dc979f49682`
- Genuine Standard: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Original BTS: `external/BTS/`, unchanged; LogN16, SlotLimbs=15, BTSModePlain, seed=0, amp=1, 32,768 slots.

Preserve the original Secondary dirty worktree and use an independent clean checkout based on the verified Fast source. Do not modify BTS Go source or module files, Standard, CKKS workload/parameters/order, or Q-prefix arithmetic. Fast remains intentionally insecure zero-secret simulation; this task makes no security-equivalence claim and does not restart generic capacity-certificate research.

## Approved Fast Decrypt/Decode contract

At `rlwe.NewDecryptor` construction, detect the existing Fast CKKS zero-secret parameter capability by reusing the mode-detection pattern already used by `NewEncryptor`. For a ciphertext, use the projection only after validating the Fast compact representation, metadata/domain, required effective rows, and that every non-c0 component is zero over the authoritative prefix. Use `w_Q(Level)=min(Level+1,4)` and never access dormant rows.

For that validated path, project c0's effective residues into a plaintext at `min(ciphertext.Level,3)`, preserving Scale and required NTT/Montgomery and batching metadata. Keep the input ciphertext's logical Level unchanged; report both Levels. Do not synthesize absent high Q residues, claim full-high-Level decryption, or extend this behavior to Standard, other schemes, nonzero-secret modes, or malformed/unsupported representations. Reject nonzero c1, missing authoritative rows, bad metadata, and unsupported domains. Preserve the existing generic decryption path for complete-Q ciphertexts.

## Zero-Bootstrap validation

Add focused existing-package tests for LogN16 compact logical Level 18 (`q0..q3`, c1=0) using ordinary `DecryptNew → Decode`; assert plaintext Level 3 versus ciphertext logical Level 18, finite decoded slots, and an independent coefficient-domain prefix oracle. Cover Levels 0–3 and negative cases for nonzero c1, missing required rows, invalid metadata, and unsupported domain. Confirm genuine Standard behavior is unchanged. Test compact ciphertext snapshot serialize/restore and decoded equality without Bootstrap. No full BTS, EvalMod, C2S, or S2C in tests.

Use maintained measurement owners (`cmd/perfprobe`, `cmd/fastdiag`, `internal/perfmeasure`, `internal/numericalmetrics`) and existing regression suites; do not create a duplicate runner, metrics implementation, CRT oracle, or capacity tracker. Record the four independent claims separately: API modular behavior, numeric oracle, Fast dispatch/projection path, and `CAPACITY_UNPROVEN_EXPERIMENTAL` status.

## Output checkpoint and bounded experiments

The next Fast call is permitted only after implementation, focused tests, snapshot round-trip/decode tests, source review, and safe Secondary commit/push. Use a minimal measurement-only overlay outside `external/BTS/` to record immediately after Bootstrap returns and before the original `DecryptNew`:

- ciphertext logical Level, Scale, Degree, NTT/Montgomery, batching, and effective row count;
- a versioned compact snapshot of only actually backed authoritative component rows, with metadata and hashes; never fabricate dormant rows;
- complete original execution log and decoded vector/hash if public decode succeeds.

Fix the binary SHA, linked backend commit, original BTS source manifest, command, and exclusive reservation in the durable Task 007 journal before invocation. Fast budget: at most **1** complete `bts.Bootstrap()` call. A failed decode may be repaired/replayed offline from the saved snapshot; it never authorizes another Fast call.

After a valid Fast vector exists, inspect prior Standard artifacts for a complete compatible 32,768-slot vector. If none exists, perform at most **1** independently compiled BTS run against pristine genuine Standard; do not reuse a Standard-looking binary linked to Fast. No warmup, trace, retry, or extra Bootstrap.

## Numerical comparison and report

For same-input aligned vectors `X` (original plaintext), `S` (genuine Standard), and `F` (Fast projection), use the shared numerical-metrics implementation to calculate:

- `MaxAbsDiff=max|F[i]-S[i]|`, `RMSE=sqrt(mean(|F[i]-S[i]|^2))`, and `MeanAbsDiff=mean(|F[i]-S[i]|)`;
- `StandardMaxError=max|S[i]-X[i]|` and `FastMaxError=max|F[i]-X[i]|`.

Require exactly 32,768 finite slots and valid expected metadata. Save canonical input, Standard/Fast vector, source manifest, and binary hashes. Every nonzero metric includes scientific notation, `log10`, and `10^floor(log10)` order; exact zero is reported as zero without logarithm. Do not add a `1e-6` gate or hide large errors. Label Fast plaintext Level 3 versus Standard native full-Q plaintext Level 18. The comparison is an experimental prefix-projection observation, not proof of full ciphertext semantic equivalence.

## Verification, Git, and terminal conditions

Run the focused decrypt/decode/snapshot tests, relevant Fast/core tests, affected-package `go vet`, and `git diff --check`; check call graphs so tests do not execute Bootstrap. Commit/push the Fast repair normally to `origin/fast-qprefix`, using an isolated clean checkout and preserving the original five dirty Secondary paths. Commit/push the Primary spec, pointer, journal, compact report, and required task evidence normally to `origin/main`. No force push, reset, rebase, stash, or history rewrite.

Success labels: `FAST_BTS_DECODED`, then `STANDARD_FAST_OUTPUT_COMPARED`, then `READY_FOR_WEB_REVIEW`. If Fast cannot be decoded or comparison is invalid, keep the one-call budget consumed as applicable, preserve the checkpoint/log, report the exact boundary and first suspected arithmetic stage, and do not rerun Bootstrap. Report both repository commits, tests, hashes, Levels/Scale, metrics with decimal orders, and remaining issues.
