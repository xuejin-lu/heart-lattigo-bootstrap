# FIXED-SIGMA-DIRECT-MUL-004 — Direct ct×ct multiplication and frozen-σ transfer test

**Status:** `AUTHORIZED_AUTONOMOUS_BATCH`, resumed 2026-10-22 after the user's F5-R1 and F5-R2 Web-review rulings below.
**Primary:** `xuejin-lu/heart-lattigo-bootstrap/main`.
**Secondary:** `xuejin-lu/lattigo/fast-qprefix`; initial pin `1e7828293c36f9845e30723ba4e9500d84bb50d2`.
**Genuine Standard:** original clean `xuejin-lu/lattigo@5dbffbdea05394de2ca3a432ed5318aa832e3f40` (isolated checkout, read-only).
**Scientific handoff baseline:** Primary `574a1aee5ce5b0e329096095de3eda1b2df7bd1b`. Reviewed prior interpretation: `results/FIXED-SIGMA-PROPAGATION-003/web-review.md`.
**Frozen coefficient noise σ:** EXACTLY `4.004280012097432`. The numeric value is an empirical fit to the prior **Add-containing** fixture and is *not* universal or re-fit here.
**Effective compute budget:** at most **2 genuine Standard two-stage chains** and **16 actual Fast two-stage chain executions**; at most **18 Fast launch attempts** under the two narrowly linked F5 pre-chain recovery authorizations below. This is a launch-only allowance: the actual Fast chain ceiling remains 16, with 5 completed and at most 11 still possible at the F5-R2 ruling. Standard/Fast Bootstrap, LogN16 workload/chain, BTS/multiparty, broad benchmarks and implicit Bootstrap tests remain at zero. Counters persist across interrupted sessions. Existing completed batch budgets/results cannot be reused or overwritten.

## Web ruling addendum — one pre-chain replacement only (2026-10-22)

The user relayed the completed independent Web review decision `AUTHORIZED_ONE_PRECHAIN_REPLACEMENT`. The original F5 launch remains permanently recorded as a CLI flag-parse failure; it is excluded from **actual chain executions** because the runner returned before any ciphertext operation. This one-time ruling supersedes only the original attempt-accounting treatment of that known F5 pre-chain failure.

- Preserve F5 and authorize exactly one linked launch slot `F5-R1`, using `direct-base`, `OFF/OFF`, stage seeds `20261103` and `20261104`, frozen σ `4.004280012097432`, original parameters, Fast commit `1e7828293c36f9845e30723ba4e9500d84bb50d2`, and the unchanged direct-mul implementation source. F5-R1 is the same deterministic sample; it does not select new seeds or tune the experiment.
- The total Fast launch-attempt ceiling is 17, with exactly one known pre-chain CLI failure allowed (the original F5). The actual Fast chain-execution ceiling remains 16. The ledger must separately count launch attempts and actual chain executions; it must not rewrite the original F5 as unlaunched or erase it from launch history.
- At authorization, 5 Fast chains have completed and 6 Fast launches are recorded. F5-R1 plus the 10 remaining original Fast slots may produce at most 11 additional actual Fast chains. Genuine Standard remains spent at 2/2. Bootstrap, LogN16 workload, BTS and multiparty remain 0.
- F5-R1 must be the only replacement record. Any new unexpected failure, including another pre-chain error, stops the task immediately; this ruling authorizes no further retry or replacement and does not relax any capacity, oracle, dispatch, provenance or representation gate.
- Before F5-R1, run a CLI parse/validation check with an intentionally invalid direct-mul C-variant so the binary rejects before loading CKKS parameters or running any homomorphic operation. Then run the exact corrected flags with no positional `mul-rescale-chain` argument.

## Web ruling addendum — one F5-R2 pre-chain recovery (2026-10-22)

The user relayed independent Web approval to resume after both F5 and F5-R1 stopped before any ciphertext chain. This addendum supersedes the earlier ruling **only** for the number of authorized F5 recovery launches and the Fast launch-attempt ceiling; it does not change the actual-chain budget, workload, seeds, σ, source pins, or any acceptance gate.

- Preserve the original F5 CLI parse failure and consumed F5-R1 provenance failure as separate immutable launch records. Both remain zero actual chain executions. Do not rerun any of the five successful Fast chains or either successful Standard chain.
- Authorize exactly one further launch, `F5-R2`, linked to F5: `direct-base`, `OFF/OFF`, stage seeds `20261103`/`20261104`, frozen σ `4.004280012097432`, E32 LogN13 parameters, Primary measurement source at `977d853541cf59b37e42cdfc1e635bc0cab7862d`, and Fast source `1e7828293c36f9845e30723ba4e9500d84bb50d2`. It restores the missing F5 sample; it does not authorize another chain beyond the original 16-slot matrix.
- The cumulative Fast launch-attempt ceiling is 18 (the 16 preregistered slots plus the two specifically authorized F5 recovery launches). At this ruling there are 7 launch attempts and 5 actual Fast chains; no more than 11 additional actual chains can run. Genuine Standard stays 2/2. Bootstrap, LogN16, BTS, multiparty, and broad benchmarks stay at zero. Any later unexpected failure has no retry authorization and stops the matrix.
- Before F5-R2, verify a clean source checkout at Primary `977d853...`, Fast source/ref/clean state, Go `vcs.revision` and `vcs.modified`, linked module source, binary SHA, and the exact CLI flags. Never bypass provenance checks. The new binary must be reused unchanged for all remaining Fast captures in the same frozen seed pairs; if pairwise source/binary consistency cannot be maintained, stop.
- The previously captured successful Fast binary SHA `a7083bc12aa1d3da062c0b094b99f5c4fbcf3ceb8d86ea92f981ad73ab9db2ff` was not present in the searched task build and temporary build locations. A clean independent Primary clone at `977d853...`, using the repository's ordinary `go.mod` replacement and `-buildvcs=true`, produced candidate Fast binary SHA `bde01f61c36f70dc70939eb12d72f3a42ce07a925942904939b54e516cafae17`; `go version -m` reports `vcs.revision=977d853541cf59b37e42cdfc1e635bc0cab7862d`, `vcs.modified=false`, and `-tags=perf_fast`. Its `../lattigo` module replacement resolves to the clean Secondary checkout at the pinned Fast commit. The candidate has not run an experiment chain.
- Before the authorized F5-R2 launch, run the no-compute CLI validation with an intentionally invalid `--fixed-sigma-c-variant`, using the corrected flag layout and no positional `mul-rescale-chain` token. Require the expected variant-validation error and no result/vector files. Then launch F5-R2 once with the exact frozen valid arguments. A failed or provenance-rejected F5-R2 is consumed; do not retry.

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
