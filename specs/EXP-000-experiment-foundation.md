# EXP-000 — Experiment Foundation

## Goal

Transform `heart-lattigo-bootstrap` from the legacy hardware golden-model project into a clean CKKS bootstrapping experiment harness whose frontend, workload, parameters, and measurement logic stay fixed while the underlying Lattigo implementation/commit changes.

This task is infrastructure only. Do not perform Fast-CKKS performance optimization and do not begin noise experiments.

## Repositories

Primary:
- `xuejin-lu/heart-lattigo-bootstrap`

Secondary:
- `xuejin-lu/lattigo`

The primary repository is the control plane and experiment harness. The secondary repository is the cryptographic backend.

## Core invariant

A valid Standard-vs-Fast experiment must be runnable with:

- the same primary source commit,
- the same experiment config,
- the same workload,
- the same measurement code,
- the same command,

while changing only the Lattigo backend implementation/commit.

The primary frontend must not contain a Standard/Fast selector.

## Phase A — Preserve legacy hardware work

Before deleting or restructuring hardware-model code, verify that the legacy hardware implementation is preserved at commit:

`b7ef984fb4699355f00b16bf143fc4d7b0dcdda5`

Required preserved reference:
- branch `legacy-hardware-model`

If tooling permits a non-destructive annotated or lightweight tag, also create:
- `legacy-hardware-model-v1`

Do not move the legacy branch forward during this task.

Two temporary setup branches may exist from control-plane preparation:
- `experiment-foundation`
- `exp-000-control-plane`

If they still point only to the legacy base and contain no unique work, delete them from the remote during cleanup. Do not delete `legacy-hardware-model`.

## Phase B — Convert primary `main` into experiment harness

Inspect the current tree and remove active-mainline hardware simulator implementation that is not needed by the experiment frontend, including legacy hardware evaluator/functional-unit infrastructure and hardware-only tests/reports.

Preserve or adapt only code that is genuinely useful for the new experiment harness, especially parameter/config definitions and generic input/measurement helpers.

Expected direction, not a mandatory exact tree:

```text
heart-lattigo-bootstrap/
├── AGENTS.md
├── CURRENT_TASK.md
├── README.md
├── bootstrap_config.go
├── configs/
├── experiments/
├── results/
├── specs/
├── go.mod
└── go.sum
```

Do not create unnecessary abstraction layers or config formats merely for aesthetics. Reuse the existing Go `BootstrapConfig` model where practical.

Rewrite README so the active project is described as a CKKS bootstrapping experiment harness, not a hardware golden model.

## Phase C — Backend compatibility gate

This phase decides whether `lattigo` must be modified.

### Required experiment-facing contract

The same primary frontend source must be able to build and run against both:

1. a Standard Lattigo implementation/commit, and
2. the current Fast-CKKS implementation on `xuejin-lu/lattigo`.

The primary code must not branch on backend identity and must not import Fast-only packages or constructors.

Forbidden frontend mechanisms include:
- `--fast`
- `--normal`
- `HW_MODE`
- `isFastMode`
- build tags whose only purpose is selecting Standard vs Fast
- q0/q1 Fast residue policy in primary
- zero-key/Fast key construction in primary
- direct use of `NewFastEvaluator` from primary

### First inspect, then modify

Do not assume the exact compatibility solution in advance.

First compare the public construction/execution path required by the primary harness against:
- Standard Lattigo behavior/API,
- Fast-CKKS behavior/API.

Known current gaps that must be explicitly checked include:
- Standard bootstrap construction uses the ordinary bootstrap evaluator path while Fast tests currently use a Fast-specific constructor.
- Fast key generation currently has Fast-specific key semantics.
- the legacy `bootstrap_config.go` builder currently uses a full residual parameter set and `DecodeThenModUp`, while current Fast Stage-A has more restrictive bootstrap parameter requirements.

If these prevent the same frontend from running both backends, implement the smallest backend-side compatibility change in `xuejin-lu/lattigo` that preserves Fast architecture and Standard behavior.

When modifying the secondary repository:
- obey `lattigo/AGENTS.md`,
- obey `docs/FAST_CKKS_SPEC.md`,
- do not weaken Standard behavior,
- do not route Fast ciphertexts through Standard full-RNS code as a hidden fallback,
- keep the current speed-first Stage-A constraints unless compatibility genuinely requires a bounded extension.

### Acceptance test

Provide one primary command/workload that can be executed twice without editing the primary source or config:

```text
Run A: Lattigo Standard implementation/commit -> succeeds
Run B: Lattigo Fast implementation/commit     -> succeeds
```

Only the backend checkout/version may differ.

If exact commit switching is constrained by Go module mechanics, establish a reproducible local-development linkage that still preserves the invariant that the primary source/config/command do not change between A and B. Document exactly how the backend commit is selected.

## Phase D — Establish first speed-runner skeleton

Create the minimum experiment runner needed for the next task (`EXP-001`). Do not overbuild.

It should be able to:
- load/use a baseline bootstrap config,
- construct a deterministic or reproducible input workload,
- execute complete bootstrapping,
- perform warm-up before timed measurement,
- run a configurable repetition count without changing code,
- emit machine-readable results (JSON or CSV; JSON preferred if only one is needed),
- report at least elapsed time per bootstrap and enough metadata for reproducibility.

Where practical also expose Go benchmark-compatible memory metrics (`B/op`, `allocs/op`) or equivalent measurements, but do not distort the end-to-end timing methodology merely to obtain them.

Do not add stage breakdown yet unless required to make the runner correct.

## Result metadata

Experiment output should automatically record, when available:
- primary repository commit
- Lattigo repository commit
- Lattigo branch/ref
- dirty/clean repository state
- Go version
- OS
- architecture
- CPU/model if reliably obtainable
- full effective bootstrap config
- repetition count
- timestamp

Do not require the user to manually label a result as Standard or Fast. Backend identity should come from repository/version metadata, not from a frontend mode flag.

## Baseline parameter intent

The historical `DefaultBootstrapConfig()` values are the initial experimental parameter source of truth unless an incompatibility requires a documented compatibility adaptation.

Important: preserve the numerical experiment intent, but do not force the legacy parameter-builder implementation if its construction semantics conflict with the backend compatibility contract.

Any adaptation needed solely because current Fast Stage-A has bounded public parameter requirements must be explicit and documented. Do not silently compare materially different workloads and call them the same experiment.

## Validation

At minimum:

Primary:
- `go test ./...`
- the new experiment command/run path succeeds

Backend compatibility:
- same primary source/config/command succeeds with Standard backend
- same primary source/config/command succeeds with Fast backend

Secondary, if modified:
- relevant targeted Fast tests
- relevant Standard regression tests
- `go test` for directly affected packages

Review the final diffs in both repositories and remove unrelated changes.

## Commit / push policy

Primary changes belong on `heart-lattigo-bootstrap/main` unless a safe local condition requires stopping.

Secondary changes, if required, belong on `lattigo/fast-ckks` unless the architecture evidence shows a different existing branch is explicitly intended.

Never reset, stash, overwrite, or discard dirty work automatically. If either repo is dirty, on the wrong branch, or cannot fast-forward safely, stop and report the exact condition.

## Deliverable / report

At completion report:
- primary commit and push result
- secondary commit and push result, if any
- files removed/retained in the primary cleanup
- exact backend compatibility mechanism
- exact command used for the same-source Standard and Fast runs
- validation results
- any remaining blocker before `EXP-001`

## Stop condition

EXP-000 is complete only when the primary repository is a clean experiment harness and the same primary frontend can run a complete bootstrap against both Standard and Fast by changing only the Lattigo backend implementation/version.

Do not start speed optimization, stage profiling, or noise experiments in this task.