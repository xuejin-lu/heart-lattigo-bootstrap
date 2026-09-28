# Fast-CKKS diagnostics

`scripts/fastdiag` collects bounded stage, generated-power, and Q-prefix
Rescale timings for the named `p93-q55` workload. It is diagnostic evidence,
not a performance repair or a causal verdict.

Normal builds remain diagnostics-off. The Secondary call sites are guarded by
the `fastdiag.Enabled` compile-time constant, which is false unless the
dedicated `fastdiag` build tag is selected. No runtime backend selector or
third-party tracing dependency is used. Diagnostic output contains event
metadata and timing only; it does not serialize ciphertexts, keys, or
polynomial arrays.

## Trace scopes

- `stage`: public Bootstrap parent plus pack, ScaleDown, ModUp/Trace, C2S,
  real/imaginary EvalMod, S2C, unpack, and public finalization.
- `power`: generated Chebyshev powers and the observed workspace, scaling,
  Relinearize, Mul/MulRelin, Rescale, doubling, and recurrence-correction
  categories.
- `rescale`: whole Q-prefix Rescale, preflight, materialization, and bounded
  subphase timings.

Select one or more scopes; `all` selects all three:

```sh
./scripts/fastdiag trace --profile p93-q55 --trace stage
./scripts/fastdiag trace --profile p93-q55 --trace stage,power,rescale \
  --warmup 1 --repetitions 5 --out /tmp/fastdiag.json
```

The command runs the actual Secondary Bootstrap path with the `fastdiag` build
tag, checks output metadata and decoded slots against a diagnostics-off replay,
and writes versioned JSON (`fastdiag.trace.v1`) plus a concise Markdown
summary. Repetitions are capped at 10 to bound raw artifact size. Without `--out`, both files are placed under a newly created
temporary directory, whose path is printed. The current-checkout trace records
dirty state rather than changing the authoritative checkout.

## Compare Secondary refs

```sh
./scripts/fastdiag compare --profile p93-q55 \
  --baseline <sha-or-ref> --candidate <sha-or-ref> \
  --trace stage,power,rescale --warmup 1 --repetitions 5 \
  --out /tmp/fastdiag-compare.json
```

Comparison requires a clean authoritative Secondary checkout on
`fast-qprefix`. Each ref is measured sequentially in a detached temporary
worktree with its own `GOCACHE`; the authoritative checkout is not moved.
Task-created worktrees and caches are removed on normal completion. On an
unexpected trace failure, the command reports and preserves the relevant
temporary paths for inspection. No reset, stash, force, or source overlay is
used. Refs without committed diagnostic hooks are reported as
`DIAGNOSTIC_HOOKS_UNAVAILABLE_AT_REF` rather than being patched dynamically.

The compare JSON (`fastdiag.compare.v1`) reports event medians, ratios, signed
deltas, parent contributions, and parent/child closure. Zero denominators have
no ratio; negative deltas are retained. Stage closure is the E2E delta minus
the sum of stage deltas. These measurements do not by themselves establish a
cause or show that an optimization is safe.

The only supported profile in this framework is `p93-q55`: LogN=13,
LogSlots=12, degree-30 Chebyshev/P93, DoubleAngle=3, and the deterministic
4096-value input used for the accepted q0=55 diagnostic workload. Adding
profiles requires an explicit future task; this interface is not a profile
DSL.
