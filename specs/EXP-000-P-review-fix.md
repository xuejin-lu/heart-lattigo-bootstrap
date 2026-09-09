# EXP-000-P — Review Fix

## Goal

Fix the remaining experiment-foundation issues found during independent review of:

- Primary: `cfe579dabb4f1f7d6b0ab9c46c4d714d974c8f64`
- Secondary: `d5429b1c368a6828ef1bb8a229e3f57458ad6491`

Do not redesign the harness and do not begin EXP-001. Keep the successful Standard/Fast drop-in compatibility mechanism unless a fix below requires a bounded adjustment.

## Finding 1 — Active-looking config fields are currently dead or only historical

`BootstrapConfig` still exposes historical fields such as:

- `q_circuit_slots`
- `q_eval_mod`
- `log_bsgs_ratio`

but the active `NewBootstrapParametersFromConfig` path does not directly use all of them to construct the effective bootstrapping parameters.

This is dangerous for future experiments because a user may edit a visible config knob and assume the workload changed when it did not.

### Required fix

Make the active experiment contract explicit and mechanically safe.

For every field exposed in the active JSON config, do one of the following:

1. **Active field** — wire it into effective parameter construction and add a test proving that changing it changes the intended effective parameter, or
2. **Compatibility-derived / constrained field** — validate it against the derived effective parameter and fail loudly when the requested value cannot be represented, or
3. **Legacy-only field** — remove it from the active experiment config surface, or rename/structure it so it is unmistakably metadata-only and cannot be mistaken for an active knob.

Do not silently ignore an active-looking field.

Specific requirements:

- `log_bsgs_ratio`: if the current Lattigo builder hardcodes `1`, either apply the configured value to the generated DFT matrix literals after construction in a correct way, or reject unsupported non-1 values with a clear error. A user setting another value must never silently get `1`.
- `q_eval_mod`: if the current public builder derives the EvalMod chain from `mod1_log_scale`, polynomial degree, double-angle depth, etc., then either validate the historical list against the derived chain or remove it from the active knob surface. Editing `q_eval_mod` must not silently do nothing.
- `q_circuit_slots`: document and enforce how the old `DecodeThenModUp` circuit-slot prime maps—or does not map—to the new `ModUpThenEncode` compatibility path. If it is not an active parameter in the new path, it must be clearly legacy-only rather than presented as an active experiment knob.

Add targeted tests around this behavior.

## Finding 2 — GitHub Actions is incompatible with the checked-in local replace

Primary `go.mod` intentionally contains:

```text
replace github.com/tuneinsight/lattigo/v6 => ../lattigo
```

but the current workflow checks out only the primary repository at the workspace root and then runs `go test ./...`. In that layout `../lattigo` does not exist, so the checked-in workflow is not a valid reproducibility path.

### Required fix

Update `.github/workflows/go.yml` so CI checks out both repositories in sibling directories that satisfy the same local linkage contract used by developers.

Preferred structure:

```text
<workspace>/heart-lattigo-bootstrap
<workspace>/lattigo
```

Run the primary tests from `<workspace>/heart-lattigo-bootstrap`.

Use a matrix or equivalent coverage for at least:

- Standard backend: `xuejin-lu/lattigo` `main`
- Fast backend: `xuejin-lu/lattigo` `fast-ckks`

The same primary source and test command must be used in both matrix entries.

Do not add a frontend `--fast` / `--normal` selector to make CI work.

If GitHub Actions cannot run the full bootstrap smoke test within practical CI limits, it is acceptable for CI to run compile/unit tests for both backends while the documented local command remains the end-to-end acceptance path. State this clearly in README.

## Finding 3 — stale Go module identity

Primary `go.mod` still declares:

```text
module github.com/kenny0915/heart-lattigo-hw-bootstrap
```

The active repository is `xuejin-lu/heart-lattigo-bootstrap`.

### Required fix

Change the module declaration to:

```text
module github.com/xuejin-lu/heart-lattigo-bootstrap
```

Update any affected imports or metadata if necessary.

## Preserve the successful EXP-000 contract

The following behavior from EXP-000 is correct and must remain true:

- Primary source contains no Standard/Fast selector.
- Primary imports no Fast-only package.
- Primary calls ordinary `bootstrapping.GenEvaluationKeys` / `bootstrapping.NewEvaluator` path.
- Fast selection remains internal to the Fast Lattigo backend.
- Standard and Fast use the same primary command/config/source.
- Legacy branch and `legacy-hardware-model-v1` tag remain unchanged.
- Do not start speed optimization or noise experiments.

## Validation

Primary:

```text
go test ./...
```

Add tests that prove exposed config knobs cannot be silently ignored.

Backend compatibility:

- same primary source/config/command still succeeds with Standard backend
- same primary source/config/command still succeeds with Fast backend

CI:

- workflow syntax is valid
- Standard matrix entry resolves the sibling Lattigo checkout correctly
- Fast matrix entry resolves the sibling Lattigo checkout correctly

Secondary:

No secondary change is required unless the primary-side parameter validation cannot be implemented correctly without a bounded backend change. If secondary is changed, obey `lattigo/AGENTS.md` and run the relevant Fast/Standard regression tests.

## Deliverable

Report:

- primary commit and push result
- secondary commit and push result, if any
- exact treatment of `q_circuit_slots`, `q_eval_mod`, and `log_bsgs_ratio`
- tests added
- CI matrix layout
- confirmation that the same Standard/Fast command still works

## Stop condition

EXP-000-P passes when the experiment harness has no misleading dead config knobs, the checked-in CI matches the sibling-backend linkage model, and the repository identity is correct, while the EXP-000 drop-in backend contract remains intact.