# Experiment repository guidance

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

## Start here

Before starting a task:

1. Safely synchronize the primary repository branch requested by the task. Never reset, stash, overwrite, or discard dirty work automatically.
2. Read this `AGENTS.md`.
3. Read `CURRENT_TASK.md`.
4. Read the referenced specification in `specs/`.
5. If the task requires modifying the secondary `lattigo` repository, then read that repository's `AGENTS.md` and `docs/FAST_CKKS_SPEC.md` before editing it.
6. Inspect current source before editing. Repository evidence overrides assumptions from previous work.
7. Implement only the requested scope, run the required tests/benchmarks, review the diff, commit, and push.
8. Report commits, tests, benchmark evidence, and any unresolved compatibility gap.

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

`CURRENT_TASK.md` is only the active work pointer. Task-specific requirements belong in `specs/`. Durable project rules belong here.