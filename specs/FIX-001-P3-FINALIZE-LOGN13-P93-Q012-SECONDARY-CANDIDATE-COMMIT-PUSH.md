# FIX-001-P3-FINALIZE-LOGN13-P93-Q012-SECONDARY-CANDIDATE-COMMIT-PUSH

## Purpose

The production repair task at Primary commit

`91a255ba0366e8c83fbc88e70a7c76f23b9f1508`

is accepted.

Classification:

`P93_POSTPRODUCT_REPAIR_1E2_SYSTEM_PASS`

The repaired LogN13/P93/Q012 production candidate now satisfies the complete current diagnostic milestone.

Accepted final controls:

### Generated powers

Required:

[
T_2,T_3,T_4,T_6,T_8,T_{16}
]

all align with the exact Chebyshev oracle at approximately `1e-15` or better and remain centered-Q012 safe.

### Polynomial

Fast vs Genuine Standard:

- real `1.8570138426987626e-8`
- imag `1.6219059983946238e-8`

Both pass:

[
E_{poly}le3.716228023823462e-8.
]

### EvalMod stable gates

DA2 post-Rescale:
- real `1.5967815852787189e-7`
- imag `1.3046858460163482e-7`

Pass:

[
E_{DA2}le3.0517578125e-7.
]

Internal final:
- real `1.635104343335875e-4`
- imag `1.3359983063118935e-4`

Pass:

[
E_{internal}le3.125e-4.
]

Public EvalMod:
- real `0.0052323338986748`
- imag `0.004275194580198059`

Pass:

[
E_{public}le1e-2.
]

### Exact full production E2E

[
E_{max}=0.009705381393898434<1e-2.
]

Genuine Standard control:

[
5.830057349387463e-8.
]

The current `1e-2` diagnostic/system milestone is therefore passed end-to-end.

This task performs **no new algorithmic work**. It must:

1. audit the entire accumulated Secondary dirty candidate;
2. verify that the final dirty state is exactly the validated candidate;
3. run final regression tests;
4. commit the entire validated Secondary candidate;
5. push `fast-ckks` by safe fast-forward only;
6. record the resulting Secondary commit in Primary.

---

## Important milestone scope

Passing `1e-2` is the current system milestone only.

It is **not** the final research correctness target.

Do not claim that Fast now matches Genuine Standard precision.

Future precision tightening toward approximately `1e-7` is a separate later phase.

---

## Primary

Repository:
`xuejin-lu/heart-lattigo-bootstrap`

Branch:
`main`

Follow `AGENTS.md`.

Read:

- `AGENTS.md`
- `CURRENT_TASK.md`
- `docs/CODEX_HANDOFF.md`
- this spec
- accepted repair summary:
  `results/FIX-001-P3-PROD-LOGN13-Q012-GENERATED-POWER-POSTPRODUCT-SCHEDULE-REPAIR-summary.json`

Primary repair-evidence commit:

`91a255ba0366e8c83fbc88e70a7c76f23b9f1508`

---

## Secondary starting state

Repository:
`xuejin-lu/lattigo`

Branch:
`fast-ckks`

Committed HEAD before final commit:

`7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`

Validated final dirty fingerprint:

`4d2567717bb0a88024d8330cf57db6a4a3b93a7faea2374ada6ac5161ec84764`

Require exact match before doing anything.

Expected dirty diff stat:

```
circuits/ckks/bootstrapping/fast_bootstrap.go |  5 +-
circuits/ckks/bootstrapping/fast_modup.go     | 26 ++++++--
circuits/ckks/bootstrapping/fast_packing.go   | 34 ++++++-----
circuits/ckks/dft/fast.go                     |  5 +-
circuits/ckks/mod1/fast.go                    | 48 ++++++++++-----
circuits/ckks/polynomial/fast.go              | 30 ++++++++--
circuits/ckks/polynomial/fast_test.go         | 54 +++++++++++++++++
schemes/ckks/fast/automorphism.go             |  6 +-
schemes/ckks/fast/ciphertext.go               | 22 ++++---
schemes/ckks/fast/evaluator.go                | 12 ++--
schemes/ckks/fast/evaluator_ntt.go            | 62 ++++++++++++-------
schemes/ckks/fast/fast_add.go                 | 16 ++++-
schemes/ckks/fast/fast_mul.go                 | 41 ++++++++-----
schemes/ckks/fast/linear_transform.go         | 44 ++++++++------
schemes/ckks/fast/partial_ntt.go              | 29 ++++++---
schemes/ckks/fast/rescale.go                  | 86 +++++++++++++++++++++++++--
schemes/ckks/fast/ring_degree.go              |  5 +-
schemes/ckks/fast/trace.go                    |  7 +--
schemes/ckks/fast/truncate.go                 | 32 ++++++++--
19 files changed, 418 insertions(+), 146 deletions(-)
```

If branch, HEAD, dirty fingerprint, or file set differs, stop:

`SECONDARY_FINAL_STATE_MISMATCH`.

Never reset, stash, clean, discard, checkout-overwrite, or reconstruct.

---

# R0 — complete Secondary diff audit

Inspect the full accumulated diff:

[
7d05f1f3... ightarrow 	ext{current dirty worktree}
]

not just the last two repaired files.

Audit all 19 expected files.

For each file classify its changes under one of the intended Fast-CKKS candidate areas:

- zero-key / maintained arithmetic
- q012 maintained representation
- NTT / partial NTT
- addition / multiplication / linear transform
- automorphism / trace / ring-degree / truncate
- ModUp / packing / bootstrap
- DFT
- EvalMod
- polynomial generated powers / PS
- tests

Verify:

1. no unrelated application/demo/debug code;
2. no temporary instrumentation left enabled;
3. no hardcoded test-only values in production paths;
4. no accidental parameter change;
5. no commented-out safety check that should remain;
6. no generated artifacts or binaries;
7. no changes outside the expected 19-file set.

Pay special attention to the final repair in:

- `circuits/ckks/polynomial/fast.go`
- `circuits/ckks/polynomial/fast_test.go`

Verify the repaired Q012 generated-power ordering is actually:

[
oxed{
	ext{Mul/MulRelin}
ightarrow
	ext{full Chebyshev recurrence}
ightarrow
	ext{single Rescale}
}
]

inside the proven predicate, with balanced fallback retained outside it.

If the audit finds an unrelated or unsafe change, stop without committing.

---

# R1 — remote fast-ckks safety check

Do not pull across the dirty worktree.

Use read-only remote inspection/fetch/ls-remote as safe.

Determine the current remote `origin/fast-ckks` commit.

Safe commit/push requires remote state to be compatible with the local base.

Acceptable:

- remote branch absent; or
- remote `fast-ckks` exactly at `7d05f1f3c6f8a14dea2bcb2d3fa05246d322a797`; or
- another state that is provably an ancestor and permits a clean fast-forward without rewriting another person's work.

If remote has unexpected commits not already incorporated, stop:

`SECONDARY_REMOTE_DIVERGED`.

No force push.

---

# R2 — final validation before commit

Run against the exact final dirty candidate.

## Secondary

Required:

- `git diff --check`
- directly affected Fast polynomial tests
- `go test ./circuits/ckks/polynomial/...`
- `go test ./schemes/ckks/fast/...`
- `go test ./circuits/ckks/mod1/...`
- `go test ./circuits/ckks/bootstrapping/...`

Also run:

- `go test ./...`

If full `go test ./...` fails only for a clearly documented pre-existing/environment-only reason, record it exactly and do **not** misrepresent it as pass. Any failure plausibly related to the candidate blocks the commit/push.

## Primary system confirmation

Re-run the final production repair validation or the minimum authoritative test that reproduces:

[
E_{max}approx0.009705381393898434.
]

Require:

[
E_{max}le1e-2.
]

Also require polynomial and public EvalMod gates remain passed.

No parameter changes.

---

# R3 — commit Secondary candidate

Only if R0–R2 pass.

Commit **the entire validated current Secondary dirty candidate** on branch `fast-ckks`.

This is intentional: the validated candidate consists of the accumulated 19-file Fast-CKKS implementation, not only the last polynomial repair.

Suggested commit message:

`fast-ckks: validate LogN13 P93 Q012 bootstrap candidate`

Before commit:

- show `git status --short`;
- verify only the expected 19 files are included;
- verify no untracked accidental artifact will be committed.

After commit:

- record commit SHA;
- require worktree clean;
- record `git status --short`.

Do not amend or squash unrelated remote history.

---

# R4 — push Secondary safely

Only after a successful local commit and final remote safety recheck.

Push:

`fast-ckks -> origin/fast-ckks`

Fast-forward only.

No force push.

After push:

- verify remote branch points to the new Secondary commit;
- verify local branch and remote branch are synchronized;
- verify worktree remains clean.

If push fails or remote changed between checks, stop and report without rewriting history.

---

# R5 — final milestone verification

After Secondary push, run or confirm the authoritative Primary validation against the committed Secondary candidate.

Record:

- Secondary commit SHA;
- branch;
- exact E2E;
- polynomial real/imag;
- public EvalMod real/imag;
- Genuine Standard control;
- test status.

Required milestone:

[
E_{max}le1e-2.
]

---

# R6 — update Primary project state

Create compact artifact:

`results/FIX-001-P3-FINALIZE-LOGN13-P93-Q012-SECONDARY-CANDIDATE-COMMIT-PUSH-summary.json`

Include:

- starting Secondary HEAD/fingerprint
- audited file list
- audit pass/fail
- remote pre-push state
- validation results
- new Secondary commit SHA
- remote post-push state
- clean-worktree confirmation
- exact final E2E
- current milestone status
- explicit note:
  `1e-2 system milestone passed; final precision target remains a later phase`

Then update:

- `docs/CODEX_HANDOFF.md`

to make the new Secondary committed SHA authoritative.

Also update:

- `CURRENT_TASK.md`

to:

```
# Current Task

Task: None
Status: MILESTONE_1E2_COMPLETE
```

unless a failure prevents finalization.

Commit and push these Primary state updates according to `AGENTS.md`.

---

# Decision classification

Choose exactly one:

## A — `SECONDARY_FINAL_STATE_MISMATCH`

Starting dirty candidate differs from the validated fingerprint/file set.

## B — `SECONDARY_FINAL_DIFF_AUDIT_FAIL`

Unexpected/unsafe diff found.

## C — `SECONDARY_REMOTE_DIVERGED`

Remote branch cannot be safely fast-forwarded.

## D — `SECONDARY_FINAL_VALIDATION_FAIL`

Candidate fails relevant regression/system validation.

## E — `SECONDARY_COMMIT_OR_PUSH_FAIL`

Audit/validation pass, but commit or push cannot be completed safely.

## F — `LOGN13_P93_Q012_MILESTONE_1E2_FINALIZED`

All hold:

1. full 19-file diff audited;
2. validation passes;
3. exact E2E <= `1e-2`;
4. Secondary candidate committed;
5. Secondary pushed by safe fast-forward;
6. remote/local synchronized;
7. Secondary clean;
8. Primary handoff updated to the new committed SHA.

---

## Prohibitions

- no new algorithmic modification
- no parameter tuning
- no threshold relaxation
- no force push
- no reset/stash/clean/discard
- no history rewrite
- no LogN16
- no Gate4/5
- no EXP-003
- no benchmark campaign

This task is only final audit, commit, push, and milestone-state publication.
