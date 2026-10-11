# Experiment repository guidance

## User-facing language

All user-facing progress updates, warnings, questions, error explanations, and final reports must be written in **Traditional Chinese (Taiwan usage)** unless the user explicitly asks for another language.

Keep source code, shell commands, file paths, Git refs, API/type/function names, and other technical identifiers in their original form. Do not convert technical identifiers merely to satisfy the language rule.

## Mandatory startup preflight

This section has highest priority for every new Codex run in this project.

When the user says `開始`, `start`, `continue`, or otherwise asks to begin the current task, **do not trust the local `CURRENT_TASK.md` or any remembered task status yet**. A local task marked `COMPLETE` may be stale because the orchestrator can update the remote repository between Codex runs.

Before deciding whether there is work to do:

1. Go to the primary repository `heart-lattigo-bootstrap`.
2. Check the current branch, `git status --short`, and `git rev-parse HEAD`. If the branch is not `main`, stop and report it. Do not infer task status from local `CURRENT_TASK.md` yet.
3. Inventory local changes before synchronization. A dirty worktree is **not automatically unsafe**: expected task outputs or implementation files may be legitimate checkpoints from the current task. Do not reset, stash, overwrite, discard, or switch away from any local work.
4. After recording that inventory, run `git fetch origin` even if there are untracked files or reviewed task changes. Fetch updates Git metadata and remote-tracking refs; it does not check out or overwrite worktree files. If fetch fails, stop and report the exact error.
5. After a successful fetch, read the remote-authoritative instructions and task pointer without checking files out:
   - `git show origin/main:AGENTS.md`
   - `git show origin/main:CURRENT_TASK.md`
   - the referenced spec from `origin/main` (for example, `git show origin/main:specs/<name>.md`)
   Inspect local-only commits and local changes as well; preserve a verified in-progress task checkpoint when it is authorized by the active task.
6. Classify every local change against the remote task and spec. Inspect tracked diffs and, for untracked results, their exact paths, sizes, hashes, timestamps, compact required fields, provenance, and any attempt journal/budget. The path checker is only a scope guard, not proof that file contents are valid:
   - `python3 scripts/task_worktree_policy.py --repo . --allow-task-path <verified-task-path> ...`
   - `python3 -m unittest discover -s scripts -p 'test_task_worktree_policy.py' -v`
   Continue with task-owned changes only when their content and provenance are verified and every changed path is explicitly expected by the active task. If any path or modification is unknown, unrelated, unsafe, or cannot be verified, stop and report it without changing it.
7. Compare local `HEAD` and fetched `origin/main` ancestry before any merge:
   - If `origin/main` is a descendant of local `HEAD`, first compare the paths changed by `HEAD..origin/main` with every local changed path. Treat case-only path differences as overlaps on case-insensitive filesystems. If any path overlaps, stop; otherwise run `git merge --ff-only origin/main`. This check is mandatory when untracked files are present so a fast-forward cannot overwrite them.
   - If `HEAD` equals `origin/main`, no merge is needed.
   - If `origin/main` is an ancestor of local `HEAD`, inspect every local-only commit. Continue only when those commits are verified as authorized current-task work; do not rewrite or discard them.
   - If neither commit is an ancestor of the other, stop and report the local-only and remote-only commits. Do not merge, rebase, reset, or force push.
8. Verify synchronization after the decision: when no verified local-only commits remain, confirm `git rev-parse HEAD` equals `git rev-parse origin/main`; when authorized task commits are ahead, confirm `origin/main` is an ancestor of `HEAD` and report the ahead count. In all cases, confirm every reviewed task file remains present and unchanged across any fast-forward.
9. **Only after fetch and safe synchronization/ancestry checks**, re-read the effective `AGENTS.md`, `CURRENT_TASK.md`, and authoritative spec. A remote task update supersedes stale local instructions; verified authorized local task commits remain part of the resumable checkpoint.

Never use `git pull`; after a successful fetch, use the fetched `origin/main` for any fast-forward. Never use `sudo`, arbitrary permission changes, reset, rebase, stash, force push, or history rewriting to get past a failed or unsafe synchronization.

A previous local `CURRENT_TASK.md` saying `COMPLETE` is never sufficient reason to skip the startup fetch. If the synchronized task requires work in the secondary `lattigo` repository, perform that repository's own safe synchronization procedure before reading or editing task-relevant files.

### Process-launch failure recovery for routine fast-forward sync

A process-launch error such as `No such file or directory` while invoking the **permission-escalation wrapper** is not evidence that `git merge` ran or failed. Do not classify it as Git divergence or treat it as a CKKS task blocker requiring mathematical review.

If `git fetch origin` succeeded, the branch is `main`, all local changes have been verified as task-owned, and `origin/main` is a descendant of local `HEAD`, first confirm the remote-changed paths do not overlap reviewed local changes, then execute `git merge --ff-only origin/main` once using the normal available shell execution channel and the repository's real absolute working directory. Do not request escalated permissions unless a normal Git command **actually returns a relevant permission error**. A previous wrapper process-launch failure does not itself require escalation.

After an ordinary Git metadata permission failure, use the formal escalation mechanism once when available. If that wrapper itself cannot launch a process, do not repeatedly invoke the same broken wrapper. Verify paths and tools once. If the ordinary channel is also unable to launch or Git still cannot write metadata, stop with `ENVIRONMENT_SYNC_BLOCKED`, report the exact attempted command/error, and provide the smallest safe manual command for a user terminal; never pretend synchronization occurred.

No `sudo`, arbitrary permission changes, `pull`, reset, rebase, stash, force push, history rewriting, or modifications to unreviewed local work. Never skip the startup fetch merely to get past tool failure. Do not repeat completed expensive measurements: resume from the verified task artifact/journal and enforce the task-wide attempt budget across sessions. Bootstrap reservations and actual calls are durable task-scoped totals; a new Codex session never resets them, and the journal/results must be reconciled before another expensive call.

## Mandatory measurement-platform reuse (all chats and Codex sessions)

**This project already owns a mature measurement and diagnostic platform. DO NOT rebuild it for every Batch.** Before ANY performance, correctness-diagnostics, Standard/Fast comparison, Bootstrap profile, experiment-runner or tooling task, read `docs/MEASUREMENT_PLATFORM.md` **after startup sync and before implementation/spec interpretation**. This rule is durable across Web-chat/model/Codex handoffs, overrides a task's silence about reuse, and applies even when `CURRENT_TASK.md` points to a new Batch.

Default and only maintained Primary measurement platform: `cmd/perfprobe/` (matched Standard/Fast experiments and timings), `cmd/fastdiag/` plus `scripts/fastdiag` (stage/power/rescale attribution), `internal/perfmeasure/` (shared effective parameters/config/input/provenance), and `internal/numericalmetrics/` (shared decoded metrics). Secondary `internal/fastdiag/` hooks are library instrumentation, not a parallel Primary platform. Root `fast_measurement*.go`, old root commands, and historical `tools/` runners are prior art only and may be retired after dependency and regression-test review. Read relevant historical DIAG/OPT evidence as prior art. **Inventory first, identify unsupported capability second, extend shared tools third; never build another runner, statistics tool, or event-tree validator.** Do not duplicate stage timing, RMSE/SNR, profile hashing, warmup/statistics or report aggregation as convenience.

**Obsolete legacy assumptions must be fixed in the platform itself:** `perfmeasure` hard-coded E=0, `perfprobe` historical Fast direct-encoded c0 and explicit `NewFastEvaluator`, pinned old refs and forced warmup/repetition minima do not automatically represent the accepted public E32 workflow. Repair these with explicit versioned diagnostic-vs-formal-public modes and regression tests; do not silently discard legacy experiments or fabricate a formal Standard baseline. The intentionally insecure current Fast zero-secret **c1=0 remains a mode-specific invariant**; do not confuse that with the obsolete universal E=0 requirement or manual direct-c0 construction.

If shared measurement infrastructure cannot yet run the accepted public CKKS workload correctly, **platform repair takes precedence over a new benchmarking Batch**, and Web must authorize any required new math/backend representation semantics. Handoff reports must state tool reuse status and remaining unsupported mode(s).

### Effective-parameter generality and historical-tool retirement

- Derive recorded `LogN`, `LogSlots`, slot count, scale, exact Q/P chains, and E from the constructed effective CKKS and bootstrapping parameters. Do not make shared measurement algorithms depend on a LogN13/LogN16 allowlist. Frozen historical profiles may retain explicit hashes and pins, but that fixture identity must remain separate from generic platform capability.
- Exercise the same shared logic for LogN14 and LogN15 where effective parameters can be constructed. If parameter construction, capacity, or a backend rejects a profile, return an explicit unsupported/not-proven classification and the exact reason; never relabel the profile as supported or substitute another backend.
- For trace validation and event aggregation, reuse the existing `cmd/fastdiag` recursive event-tree analysis and closure utilities. Do not add a second event-tree validator or duplicate the Task025 recursive generated-power fix in another adapter.
- Historical runner removal is allowed when its implementation is superseded by the maintained platform and every independent mathematical/security regression it uniquely protects is retained in an existing maintained package. Preserve research specs, configs, result artifacts, fingerprints, and Git history; do not add compatibility wrappers solely to keep retired CLIs compiling.
- Bootstrap limits are scoped to the active authorized task/charter, not permanent repository-wide bans. Historical reservations remain immutable evidence and are never reusable within the task that spent them. A later task may authorize a new bounded capture with its own journal and stop conditions; never repeat completed performance campaigns merely to refresh a number.
- The application CKKS operation source remains identical across backends and uses ordinary public APIs. Standard/Fast selection is only through isolated compilation against the respective dependency; no frontend backend-specific arithmetic branch is permitted.

## Durable research workflow authority

Before executing or reviewing nontrivial research work, use:

`docs/RESEARCH_ENGINEERING_WORKFLOW.md`

That document defines the permanent Web/Codex division of labor, M/I/E task classes, scientific-versus-coding review boundaries, escalation rules, and the minimal `開始` / `review` user workflow.

If this file and an active task spec are silent about who should make a mathematical/scientific decision, defer to that workflow document rather than improvising authority.

## Project role

This repository is the primary project and experiment harness for CKKS bootstrapping experiments.

Primary repository:
- `xuejin-lu/heart-lattigo-bootstrap`
- Owns experiment configuration, workloads, measurement, result metadata, and experiment documentation.

Secondary repository:
- `xuejin-lu/lattigo`
- Owns CKKS implementation details, including the Fast-CKKS backend.

The core experiment contract is:

> Keep the experiment frontend, workload, parameters, and measurement method fixed; change only the Lattigo implementation/commit being tested.

### Formal Standard baseline contract

For any result claimed as a formal Standard numerical-quality or performance baseline:

- Standard must run from an explicitly pinned Standard Lattigo implementation/commit whose production arithmetic has not been modified for Fast compatibility or for the experiment.
- Standard input must be produced through the native Standard lifecycle appropriate to the experiment: ordinary Standard key generation, encryption, the measured operation (for example Bootstrap), and decryption/decoding. Do not hand-construct or mutate Standard ciphertext components (for example forcing `c1 = 0`) to make them resemble Fast state.
- Standard and Fast must use the same original message/workload and the same effective CKKS parameter configuration within a profile. They are **not** required to have identical ciphertext coefficients, key material, randomness, or security/noise semantics.
- A plaintext-like, zero-a, manually constructed, or otherwise simplified ciphertext may be used only when an explicit diagnostic task defines that specialized experiment. Such a run must be labeled diagnostic-only and must not be promoted to a formal Standard baseline, formal Standard-vs-Fast numerical result, or correctness-preserving performance result.
- If the formal Standard lifecycle cannot be run under the required matched profile, report the experiment as blocked/non-comparable instead of substituting a simplified Standard input path.

## Task execution after startup sync

After the mandatory startup preflight has completed:

1. Read the synchronized `CURRENT_TASK.md`.
2. Read the referenced specification in `specs/`.
3. If the task requires work on the secondary `lattigo` `fast-qprefix` branch, safely synchronize that branch and read its `AGENTS.md` and authoritative `docs/FAST_QPREFIX_SPEC.md`. Do not read the historical `docs/FAST_CKKS_SPEC.md` by default; consult it only when the active task explicitly requires historical background. For a pinned Standard baseline, follow the active task's specified Standard ref without treating Q-prefix or historical Fast guidance as Standard rules.
4. Inspect current source before editing. Repository evidence overrides assumptions from previous work.
5. Execute the current task with the bounded autonomous cycle below.
6. Report commits, tests, benchmark evidence, self-review findings, and any unresolved compatibility gap.

### Milestone-batch exception (explicitly authorized only)

When synchronized `CURRENT_TASK.md` points to an **approved autonomous batch charter** in `specs/`, apply `docs/RESEARCH_ENGINEERING_WORKFLOW.md` §4B: execute its **sequential authorized checkpoints**, using the coding self-review/repair cycle after each checkpoint, and automatically advance after a passing gate **without waiting for Web**. The batch charter replaces the usual single-task `READY_FOR_WEB_REVIEW` handoff **only for its enumerated I/E scope**; the GPT Web review is mandatory at the completed batch boundary or at a STOP escalation. Preserve a compact `results/` journal and expensive-experiment budget for safe resume across Codex sessions. A new mathematical/representation choice, unexpected failure or forbidden scope change **always** requires immediate `NEEDS_WEB_REVIEW`; it is not something Codex may self-authorize. Without an explicitly approved batch, the usual single-task cycle below remains in force.

### Bounded autonomous implementation/review cycle

A normal `開始` run should not stop after the first implementation pass merely to ask the user for a review. Codex owns one bounded local review-and-repair cycle before handing the result back to the independent GPT Web orchestrator.

Default cycle:

```text
Pass 1: implement
        ↓
Self-review: inspect spec compliance + diff + tests + edge cases
        ↓
Pass 2: repair only findings supported by that self-review
        ↓
Final validation
        ↓
Commit + normal push
        ↓
READY_FOR_WEB_REVIEW
```

#### Pass 1 — implementation

- Implement only the synchronized task specification.
- Run the task's required targeted tests/benchmarks.
- Do not expand scope merely because adjacent cleanup is convenient.

#### Self-review — mandatory

Before committing, review the actual local result as if reviewing another developer's change:

- compare every acceptance criterion against the implementation;
- inspect the complete task diff for unintended changes;
- check architecture/invariant compliance;
- inspect error handling, boundary cases, and stale assumptions;
- run or re-run the most relevant targeted tests;
- classify every discovered issue as either a concrete task defect, unrelated pre-existing debt, or an unresolved design question.

Do not treat self-review as independent scientific validation. It is a coding-quality gate only.

#### Pass 2 — bounded repair

- Fix concrete defects found by the self-review.
- Re-run affected tests.
- Do not start a third design/implementation direction.
- Do not broaden the task specification.
- If self-review finds no concrete defect, Pass 2 may be a no-op.

After Pass 2, perform final validation, review the final diff, commit, and push under the standing safe-push rules.

Intermediate commits/pushes between Pass 1 and Pass 2 are not required unless the active task explicitly requests them. Prefer one clean final task commit (or the minimum coherent commits naturally required by the task) rather than creating bookkeeping commits solely for the self-review cycle.

#### Mandatory escalation to GPT Web

Stop the autonomous cycle and report `NEEDS_WEB_REVIEW` instead of continuing to improvise when any of the following occurs:

- the specification conflicts with durable architecture or current source evidence;
- a mathematical/cryptographic semantic decision is required rather than a coding decision;
- the same unexplained acceptance/test failure remains after Pass 2;
- the supported repair would require expanding the authorized task scope;
- a Primary-only task appears to require Secondary changes, or vice versa;
- safe Git synchronization/push conditions fail;
- evidence is insufficient to distinguish a real defect from stale diagnostic/provenance debt;
- completing the task would require relaxing a threshold, changing parameters, or weakening an invariant not authorized by the specification.

Do not loop indefinitely. The default autonomous budget is exactly one implementation pass, one self-review, and one repair pass.

#### Handoff status

On successful completion, report:

`READY_FOR_WEB_REVIEW`

The independent GPT Web orchestrator remains responsible for accepting/rejecting the result and deciding the next task. Codex must not invent the next task or treat its own self-review as final project acceptance unless the synchronized specification explicitly delegates that authority.

## Standing safe-push authorization

The user has provided standing authorization for the ordinary task-completion push described by this repository workflow. **Do not stop solely to ask for a separate push confirmation** when all conditions below are true.

For the primary repository, `git push origin main` is pre-authorized after task completion only when:

- startup preflight previously succeeded for the current task;
- the task implementation and required task-specific validation have completed successfully;
- if the repository-wide suite still contains a pre-existing documented failure, that failure may be non-blocking only when all of the following are true: the active task did not modify its code path, the exact failing test/classification matches durable repository documentation, the task-specific required tests pass, and the completion report explicitly records the debt; any new, changed, unexplained, or task-related failure remains blocking;
- the local branch is `main`;
- the worktree is clean after committing the task result;
- `origin/main` is an ancestor of local `HEAD`, so the push is a normal fast-forward;
- the commits being pushed contain only the current authorized task/workflow changes;
- no force push, ref rewrite, rebase, amend, reset, history rewrite, branch deletion, or unrelated remote mutation is required.

Immediately stop and report instead of pushing if any of those conditions fail, if `git push` would be non-fast-forward, if remote history changed unexpectedly, if unknown/unrelated commits are present, or if the exact destination ref is ambiguous.

This standing authorization is intentionally narrow. It does not authorize force push, destructive Git operations, publishing secrets, pushing unrelated work, changing repository visibility/settings, creating releases, merging pull requests, or any other external side effect beyond the normal fast-forward task-completion push.

If the execution environment itself enforces an interactive approval that cannot be satisfied by repository instructions, report that as an environment-level restriction; do not misrepresent it as a project requirement.

For a secondary-repository task, the same principle may be used only when the active Primary specification explicitly requires a secondary commit/push and all secondary repository safety rules are satisfied. Otherwise do not modify or push Secondary.

## Authoritative current backend-transparent CKKS status (2026-10-11)

**MUST READ:** `docs/BACKEND_TRANSPARENCY_CONTRACT.md` before any task involving CKKS frontend, evaluator dispatch, Standard/Fast measurement, or backend selection. The current supported frontend workflow already achieves backend-transparent drop-in through **ordinary** `ckks.NewEvaluator`, `bootstrapping.NewEvaluator`, public CKKS operation calls, and library-internal Fast dispatch. Do not demand, recommend, or reintroduce explicit `NewFastEvaluator` in current frontend code. That constructor is allowed only inside the Fast library or clearly historical diagnostics.

Use **the same Primary application source built twice** against separate pinned, clean Lattigo source checkouts: original genuine Standard and current Fast. A purported Standard evaluator **compiled against Fast checkout is not genuine Standard**, irrespective of its constructor name. Enforce binary/module source provenance but do not create an unnecessary Standard algorithm-revalidation project. Preserve historical data and clearly separate achieved API transparency from pending current LogN16 numerical acceptance. Older 'known gap' discussions are historical, not current status.

## Permanent architecture rules

- Production/application code must remain backend-transparent: application CKKS usage must not branch on Standard versus Fast.
- Development and research harnesses may select Standard versus Fast automatically through build tags, build configuration, or orchestration scripts. This selection is allowed only at the harness/backend-adapter boundary; it must not change the shared workload, effective CKKS parameters, measurement definitions, or application-facing CKKS algorithm being measured.
- Do not add application/runtime Fast selectors such as `--fast`, `--normal`, `HW_MODE`, `isFastMode`, or equivalent switches to production/application code. Backend choice for research automation belongs in the development harness/build layer.
- Do not move Fast implementation details into this repository. In particular, do not add q0/q1 residue policy, zero-key construction, Fast ciphertext storage, Fast evaluator internals, or Fast noise behavior here.
- If the same public application/frontend cannot run against both Standard and Fast because of a backend API/behavior mismatch, fix the compatibility at the Lattigo backend boundary whenever practical instead of branching the application/frontend.
- A valid Standard-vs-Fast comparison must use the same shared harness source, experiment config, original workload/message, effective CKKS parameters, measurement code, and measurement procedure. Build-time/backend-adapter selection may differ in order to bind the harness to the pinned Standard or Fast implementation; that selection itself must not alter the measured algorithm or metric definition.
- Automatically record backend identity and environment metadata in experiment outputs whenever the runner is implemented: primary commit, Lattigo commit, Lattigo branch/ref when available, dirty/clean state when available, Go version, OS/architecture, config, repetitions, and timestamp.
- Current research priority is speed. Do not add new noise-fidelity mechanisms unless a later task explicitly starts the noise phase.

## Repository roles

Keep experiment-facing code here: parameter definitions, workload construction, measurement, result serialization, experiment orchestration, and analysis inputs.

Keep cryptographic backend implementation in `lattigo`: evaluator behavior, key semantics, Fast arithmetic, compact residue storage, bootstrapping internals, and compatibility shims that belong to the library.

## Legacy hardware model

The previous hardware golden-model implementation is historical work, not the active mainline architecture. Preserve it through the dedicated legacy branch/tag specified by the active task before removing hardware-model code from `main`.

## Result artifact read discipline

Large diagnostic result artifacts can exhaust tool time or conversation context even when their filename contains `summary`.

- Never full-fetch `results/*.json` unless its size is already known to be small and human-reviewable.
- A filename containing `summary` does **not** imply the artifact is compact.
- For unknown or large result artifacts, first inspect metadata/size when available, then use targeted search for specific keys/values or bounded excerpts.
- Do not load full per-slot vectors, repeated index arrays, operation traces, coefficient dumps, or large hash collections into an orchestrator chat merely to review a result.
- Prefer compact evidence: provenance, classification, checkpoint status, aggregate metrics, mismatch count, first mismatch, and worst error.
- If an artifact turns out to be unexpectedly huge, stop reading it in full and report that the artifact format needs refinement.
- New `summary.json` artifacts must remain compact and human-reviewable; detailed evidence belongs in the raw artifact, and even raw artifacts should avoid redundant multi-thousand-entry arrays when hashes/counts/first-worst evidence are sufficient.

`CURRENT_TASK.md` is only the active work pointer. Task-specific requirements belong in `specs/`. Durable project rules belong here.
