# Experiment repository guidance

## User-facing language

All user-facing progress updates, warnings, questions, error explanations, and final reports must be written in **Traditional Chinese (Taiwan usage)** unless the user explicitly asks for another language.

Keep source code, shell commands, file paths, Git refs, API/type/function names, and other technical identifiers in their original form. Do not convert technical identifiers merely to satisfy the language rule.

## Mandatory startup preflight

This section has highest priority for every new Codex run in this project.

When the user says `開始`, `start`, `continue`, or otherwise asks to begin the current task, **do not trust the local `CURRENT_TASK.md` or any remembered task status yet**. A local task marked `COMPLETE` may be stale because the orchestrator can update the remote repository between Codex runs.

Before deciding whether there is work to do:

1. Go to the primary repository `heart-lattigo-bootstrap`.
2. Check the current branch and `git status --short`.
3. If the worktree is dirty, the branch is not `main`, or another unsafe local condition exists, stop and report it. Never reset, stash, overwrite, discard, or switch away from user work automatically.
4. If the worktree is clean and the branch is `main`, run:
   - `git fetch origin`
   - `git pull --ff-only origin main`
5. If fetch or fast-forward pull fails, stop and report the exact failure. Do not continue from stale local instructions.
6. **Only after the remote sync succeeds**, re-read the freshly synchronized:
   - `AGENTS.md`
   - `CURRENT_TASK.md`
   - the specification referenced by `CURRENT_TASK.md`
7. Only now may you decide that the current task is `READY`, `COMPLETE`, superseded, or otherwise actionable.

A previous local `CURRENT_TASK.md` saying `COMPLETE` is never sufficient reason to skip the startup fetch.

If the synchronized task requires work in the secondary `lattigo` repository, then perform the secondary repository's own safe synchronization procedure before reading or editing its task-relevant files.

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

## Task execution after startup sync

After the mandatory startup preflight has completed:

1. Read the synchronized `CURRENT_TASK.md`.
2. Read the referenced specification in `specs/`.
3. If the task requires modifying the secondary `lattigo` repository, read that repository's `AGENTS.md` and `docs/FAST_CKKS_SPEC.md` after safely synchronizing the required secondary branch.
4. Inspect current source before editing. Repository evidence overrides assumptions from previous work.
5. Execute the current task with the bounded autonomous cycle below.
6. Report commits, tests, benchmark evidence, self-review findings, and any unresolved compatibility gap.

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
- the task implementation and required validation have completed successfully;
- the local branch is `main`;
- the worktree is clean after committing the task result;
- `origin/main` is an ancestor of local `HEAD`, so the push is a normal fast-forward;
- the commits being pushed contain only the current authorized task/workflow changes;
- no force push, ref rewrite, rebase, amend, reset, history rewrite, branch deletion, or unrelated remote mutation is required.

Immediately stop and report instead of pushing if any of those conditions fail, if `git push` would be non-fast-forward, if remote history changed unexpectedly, if unknown/unrelated commits are present, or if the exact destination ref is ambiguous.

This standing authorization is intentionally narrow. It does not authorize force push, destructive Git operations, publishing secrets, pushing unrelated work, changing repository visibility/settings, creating releases, merging pull requests, or any other external side effect beyond the normal fast-forward task-completion push.

If the execution environment itself enforces an interactive approval that cannot be satisfied by repository instructions, report that as an environment-level restriction; do not misrepresent it as a project requirement.

For a secondary-repository task, the same principle may be used only when the active Primary specification explicitly requires a secondary commit/push and all secondary repository safety rules are satisfied. Otherwise do not modify or push Secondary.

## Permanent architecture rules

- The primary repository must not know whether the active backend is Standard or Fast.
- Do not add runtime backend selectors such as `--fast`, `--normal`, `HW_MODE`, `isFastMode`, or equivalent switches.
- Do not move Fast implementation details into this repository. In particular, do not add q0/q1 residue policy, zero-key construction, Fast ciphertext storage, Fast evaluator internals, or Fast noise behavior here.
- If the same public frontend cannot run against both Standard and Fast because of a backend API/behavior mismatch, fix the compatibility at the Lattigo backend boundary whenever practical instead of branching the experiment frontend.
- A valid Standard-vs-Fast comparison must use the same frontend source, same experiment config, same workload, same measurement code, and same command. The Lattigo implementation/commit is the intended variable.
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