# FIXED-SIGMA-DIRECT-MUL-004 — Direct ct×ct multiplication and frozen-σ transfer test

**Status:** `AUTHORIZED_AUTONOMOUS_BATCH`, **PLANNED_NOT_RUN**. Codex executes only on the user's next `開始` after the required Git safe startup.
**Primary:** `xuejin-lu/heart-lattigo-bootstrap/main`.
**Secondary:** `xuejin-lu/lattigo/fast-qprefix`; initial pin `1e7828293c36f9845e30723ba4e9500d84bb50d2`.
**Genuine Standard:** original clean `xuejin-lu/lattigo@5dbffbdea05394de2ca3a432ed5318aa832e3f40` (isolated checkout, read-only).
**Scientific handoff baseline:** Primary `574a1aee5ce5b0e329096095de3eda1b2df7bd1b`. Reviewed prior interpretation: `results/FIXED-SIGMA-PROPAGATION-003/web-review.md`.
**Frozen coefficient noise σ:** EXACTLY `4.004280012097432`. The numeric value is an empirical fit to the prior **Add-containing** fixture and is *not* universal or re-fit here.
**Explicit compute budget:** at most **2 genuine Standard two-stage chains**, at most **16 Fast two-stage chains total**, **zero Standard/Fast Bootstrap**, zero LogN16 workload/chain, zero BTS/multiparty, no broad benchmark or implicit Bootstrap tests. Counters persist across interrupted sessions. Existing completed batch budgets/results cannot be reused or overwritten.

## 1. Why this new experiment exists

Tasks 001–003 inherited a Batch019 public lifecycle whose earlier semantics deliberately included `AddNew(A,B)` before `MulRelinNew(...,C)`. That Add was a valuable *historical integration fixture*, not a mathematical prerequisite for fixed-sigma noise modeling.

**New fixed target:** one pure ct×ct `MulRelin→Rescale` on separately encrypted **A and B**, followed by another real ct×ct `MulRelin→Rescale` multiplying that *actual output* by separately encrypted C. No `AddNew`, `Rotate`, ct×pt surrogate, direct-c0 fake Standard input, or Bootstrap in this workload. Both backends must compile the **same application-facing CKKS call sequence**; only isolated dependency build selects genuine Standard vs intentionally insecure zero-secret Fast.

```text
A_ct: EncryptNew(A, Level5, Scale Δ=2^45)
B_ct: EncryptNew(B, Level5, Scale q5)
C_ct: EncryptNew(C, Level4, Scale q4)

P1 = MulRelinNew(A_ct, B_ct)
X  = Rescale(P1)     # Level5→4, scale Δ
    [Fast-only diagnostic opt-in c0 noise η1 after completed operation]
P2 = MulRelinNew(X, C_ct)
Y  = Rescale(P2)     # Level4→3, scale Δ
    [Fast-only diagnostic opt-in c0 noise η2 after completed operation]
```

`q5` and `q4` MUST come from **actual constructed Q chain**, not literals; record actual levels and scales, reject any mismatch. Same signed Gaussian coefficient polynomial per injection maps consistently to all authoritative Q-prefix residues, in correct coefficient/NTT representation. Noise is new per *enabled* stage, with deterministic reproducible distinct seeds; never change Standard or original Fast arithmetic. `c1` must stay zero in Fast; no security restoration claim.

**Independent workload ID:** `direct-ctct-v1`, with its own config/version/workload/input hashes and stage schedule. **Do not re-use or overwrite** old canonical Add fixture hash, `PublicFixedSigmaChainFingerprint` or accepted 001–003 artifacts; share the maintained platform and code rather than duplicating it.

## 2. Inputs frozen before observing any new results

Use existing `perfmeasure.DeterministicInput(4096)` as the numerical A values and the current maintained `PublicMulRescaleWorkload(4096)`'s B values, but **do not invoke its Add operation**. Introduce a new stable C message **different from A** to avoid the hidden historical `C=A` alias:

```text
for i in 0..4095:
    C[i] = complex( ((7*i+3)%23 - 11)/512, ((11*i+5)%29 - 14)/1024 )
```

Define two predeclared workload variants before running:
- `direct-base`: A, B, C as above.
- `direct-c4`: identical A and B, C' = 4*C; unchanged q4 Scale and parameters.

The C formula is a definition **not yet run or validated**. Capacity and oracle constraints override the desire to run both: if `direct-c4` exceeds strict Q-prefix capacity, document it as `UNSUPPORTED_CAPACITY`, never quietly invent an alternate C or lower acceptance thresholds. Record exact messages and SHA-256 input/workload identities. C×4 tests sensitivity to second-step input amplitude while respecting the prior two-workload idea; it does not test arbitrary transferability.

## 3. Metrics and falsification

For each operation `i`, retain actual decoded input messages `x_i,y_i`, decoded output `z_i` and separate plaintext oracle:
```text
r_i = z_i - x_i*y_i                  # new incremental operation residual
t_i = z_i - plaintext_oracle_i      # total error, including preceding errors
```
For arms `OFF/OFF`, `ON/OFF`, `OFF/ON`, `ON/ON`, measure stagewise complex signed mean, variance, RMS, max, p95/p99/p99.9 tails, and honest correlation/covariance. At Stage2 isolate the propagation:
```text
P1 = Y_ON/OFF - Y_OFF/OFF
d1 = X_ON - X_OFF
r2a = Y_ON/OFF - X_ON*C_decoded
r2b = Y_OFF/OFF - X_OFF*C_decoded
P1 = d1*C_decoded + (r2a-r2b)       # exact within numerical representation
```
Use the actually decoded **same** C_ct input in both matched Fast arms, and validate exact provenance/row hashes and input equality before labeling this an equality. Recompute the decomposition against vectors and report decomposition residual error/tolerance explicitly, without attributing `r2a-r2b` to a specific CKKS primitive. Do not equate `r_i` with pure Relinearization noise.

Compare genuine Standard versus Fast baseline versus Fast+frozen-σ. Record RMS **relative difference** at both stages for both workload variants. Before any captures, predeclare a **descriptive screening band** of ±10% relative Stagewise Standard RMS; passing this is only `SCOPED_RMS_TRANSFER_SCREEN_PASS`, not proof of distribution/tail/Gaussian equivalence. Any failures, nonzero biases, tail gaps, new unexpected capacity losses or scale changes are real results and must remain visible; never tune sigma to meet the band.

## 4. Bounded checkpoints

### CP-A — Source and math/provenance preflight; no substantive arithmetic edits

1. Startup: Follow current Primary `AGENTS.md`, safe `git status` → `git fetch origin` → read `origin/main` AGENTS/CURRENT_TASK/spec → safe ff-only **only if no local path conflicts**; check actual local worktree checkpoints. Follow Secondary's `fast-qprefix` safe sync, `AGENTS.md` and `docs/FAST_QPREFIX_SPEC.md`. Preserve verified dirty work, don't reset/stash/clean/rebase/force-push.
2. Reuse existing `cmd/perfprobe` and `internal/perfmeasure`/`numericalmetrics`. Read `docs/MEASUREMENT_PLATFORM.md`, `docs/BACKEND_TRANSPARENCY_CONTRACT.md`, `docs/RESEARCH_ENGINEERING_WORKFLOW.md`. No new runner, duplicate statistics library, or backend-specific CKKS arithmetic branching. Inspect file *size* before reading large JSON/log files.
3. Independently show Stage1 `(Δ*q5)/q5=Δ`, Stage2 `(Δ*q4)/q4=Δ`, exact Q-prefix active rows, and strict 2B<capacity checks on the actual A/B/C values for both variants; do **not** reuse the old Add-containing capacity plan for A×B. The existing helper currently accepts LogN13, Level4/3, Scale 2^45, four active Q rows; if any assumptions fail, record `BATCH_BLOCKED_NEEDS_WEB_REVIEW`, no experimental captures.
4. Freeze workloads, seed policy, acceptance band, oracle and comparative measures in journal *before* numerical captures. Build-task code may be checkpointed with normal safe Git commit.

### CP-B — Minimal direct-mul workload and matched oracle baseline

1. Implement a **task-scoped extension** inside maintained Primary `cmd/perfprobe`, reusing `fixed_sigma_chain.go` / `fixed_sigma_ablation.go` as appropriate; do not change/remove old fixtures and do not create separate executable or statistics framework. A and B should be generated as separate genuine ciphertexts; C separately. **No AddNew in the measured direct-mul fixture**. Public CKKS API call order identical across Standard/Fast builds. Reject unsupported paths rather than silent fallback.
2. Run **one genuinely native Standard reference per workload variant**, max 2/2 total; 0 Bootstrap. Before real captures, focused `go test -run 'TestFixedSigmaDirectMul|TestFixedSigmaChain'` (or task-scoped names confirmed from source) for both isolated `perf_standard` and `perf_fast`, excluding package-wide LogN16/bootstrap tests. Check Binary `go version -m` vcs revision and modified flag plus source/build/config/workload SHA; do not misidentify genuine Standard as Fast-linked.
3. Execute `OFF/OFF` original Fast baseline and ensure native plaintext oracle and full Q-prefix strict-fit for both stages; proceed only if both backend oracles/level/scale/dispatch match. All actual Fast calls count toward the <=16 ceiling.

### CP-C — Frozen σ ablation, no retuning

Per workload `direct-base`, `direct-c4`, predeclare **two stage1 seed pairs** `(20261101,20261102)`, `(20261103,20261104)` for `direct-base`; `(20261105,20261106)`, `(20261107,20261108)` for `direct-c4`. Each pair runs four Fast arms, for **max 16 Fast chain executions across the task including any earlier attempts**. Stage1 and Stage2 seeds must remain unchanged within the corresponding four arms; OFF stages never sample. Do not consume extra hidden warmup/diagnostic chain executions beyond budget.

Each enabled stage uses exactly the frozen σ, one injection after complete MulRelin+Rescale. Track independent stage seeds, actual sample hash, strict transactional capacity evidence, unchanged c1/dormant/meta, true Fast dispatch, total/error/propagation/tail statistics, and attempts. If any matched arm or budget gate fails, stop with evidence; do not rerun/adaptively replace that sample without a new authorized budget.

### CP-D — Source-backed review and durable delivery

Recompute/validate paired numerical analysis using the existing comparators and shared metrics; explicitly write whether removing Add changed the fixed σ RMS transfer, the propagation decomposition, and peak/tail behavior. **No extra Bootstrap source audit or checker optimization** in this task; those are separate issues and must not distract from this scientific question.

Save durable `results/FIXED-SIGMA-DIRECT-MUL-004/journal.md`, `summary.md` and compact hash-listed evidence. Preserve prior `results/FIXED-SIGMA-{MUL-NOISE-001,CHAIN-002,PROPAGATION-003}/` unchanged. Run targeted regression/tests without selecting LogN16 constructors, Bootstrap, or broad package tests. Normal safe commit+push Primary; Secondary unchanged unless specifically blocked on a new scientific interface choice (then STOP for Web, don't self-invent). End `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` or `BATCH_BLOCKED_NEEDS_WEB_REVIEW`.

## Hard constraints

**No Bootstrap calls (0), LogN16 workload/chain runs (0), BTS (0), multiparty changes, large benchmarks, Standard source edits, zero-a historical A=0 comparator, silent drop of capacity checking, general automatic Rescale hook or Bootstrap-wide injection.** Results are `DIAGNOSTIC_ONLY` and intentionally insecure Fast; accepting numeric accuracy is not accepting RLWE security or equal statistical tails. Historical 003 budget 16/2 is *already spent* and is not available to rerun; 004 has its own bounded accounting.
