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
5. Implement only the requested scope, run the required tests/benchmarks, review the diff, commit, and push.
6. Report commits, tests, benchmark evidence, and any unresolved compatibility gap.

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