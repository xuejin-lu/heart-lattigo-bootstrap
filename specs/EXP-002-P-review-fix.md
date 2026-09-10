# EXP-002-P — Review Evidence Fix

## Goal

Close two evidence-completeness gaps found in the independent review of EXP-002 without changing the experiment, backend, runner, or raw measurements.

EXP-002 timing data itself is accepted. This task is summary/evidence repair only.

## Fixed evidence

Use the already-archived raw results exactly as recorded:

- `results/EXP-002-logN13-standard.json`
- `results/EXP-002-logN13-fast.json`
- `results/EXP-002-logN16-standard.json`
- `results/EXP-002-logN16-fast.json`

Do **not** rerun profiling and do **not** modify these four raw files.

Pinned identities remain:

- Primary experiment commit: `8186f50e7b591b7f76b39fb89b47c32ad1cc1410`
- Standard backend: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`
- Fast backend: `ce79b861c9b4ecb45f7a42ca5de2e98dbbdb9ef2`

## Gap 1 — missing `eval_mod_total`

The EXP-002 spec explicitly requires the useful aggregate:

```text
eval_mod_total = eval_mod_real + eval_mod_imag
```

Add an `eval_mod_total` aggregate to both pair summaries:

- `results/EXP-002-logN13-summary.json`
- `results/EXP-002-logN16-summary.json`

Derive it from the raw per-repetition measurements, not by inventing new timings.

For each repetition/backend:

```text
eval_mod_total.elapsed_ns   = eval_mod_real.elapsed_ns + eval_mod_imag.elapsed_ns
eval_mod_total.alloc_bytes  = eval_mod_real.alloc_bytes + eval_mod_imag.alloc_bytes
eval_mod_total.allocs       = eval_mod_real.allocs + eval_mod_imag.allocs
```

Then report the same summary statistics used elsewhere:

- count
- median elapsed time
- mean elapsed time
- min/max elapsed time
- median allocation bytes
- median allocation count
- median percentage of full-bootstrap time
- Standard/Fast median speedup
- median saved time

Keep `eval_mod_total` clearly marked as an **aggregate** so it is not double-counted as an additional physical stage.

Do not insert it into the existing top-stage/top-saved rankings unless the ranking is explicitly labeled as including aggregates. Prefer leaving the existing physical-stage rankings unchanged.

## Gap 2 — missing active/trivial stage context

The EXP-002 spec requires the results to explicitly record whether packing, Trace, and ring switching are effectively active or trivial for the current full-slot Standard-ring profiles.

Add a compact `stage_context` section to each LogN pair summary and an equivalent profile-level record in `results/EXP-002-matrix-summary.json`.

At minimum record, for each profile:

- packing boundary call: present
- logical packing work: active or trivial
- N1→N2 / N2→N1 ring-degree switching: active or trivial
- Trace inside `ModUp`: active or trivial
- short machine-readable reason/evidence for each classification

Derive the classification from the fixed profile/effective parameters and the actual public bootstrap path. For these current profiles, explicitly verify rather than assume:

- whether `Residual LogN == Bootstrap LogN`
- whether `LogSlots == LogMaxSlots`
- whether one input ciphertext/full-slot packing implies no real packing fan-in/fan-out
- whether the Trace loop has any effective work at full packing

The public boundary calls may still allocate/copy even when the underlying logical packing/ring-switch work is trivial; preserve that distinction.

## Scope

Allowed changes:

- `results/EXP-002-logN13-summary.json`
- `results/EXP-002-logN16-summary.json`
- `results/EXP-002-matrix-summary.json`
- optional small deterministic summary-generation/helper script only if needed to prevent arithmetic mistakes
- this spec / task pointer status bookkeeping

Do not change:

- the four raw EXP-002 result files
- `stage_runner.go`
- experiment configs
- Standard Lattigo
- Fast Lattigo
- any EXP-001/EXP-001-P/EXP-001P2 evidence

No new profiling run is required.

## Validation

Before completion:

1. Recompute `eval_mod_total` directly from all 10 raw repetitions for both backends and both profiles.
2. Verify the aggregate statistics in the summaries against those raw values.
3. Verify each `stage_context` classification against the recorded effective parameters and current bootstrap source path.
4. Preserve all existing identity checks and existing stage/full-bootstrap numbers unchanged.
5. Run `go test ./...` if any Go helper/source is added or modified. If only JSON/Markdown changes are made, no Go rerun is required.

## Stop condition

EXP-002-P passes when the archived summaries explicitly contain the missing `eval_mod_total` aggregate and active/trivial stage context, without altering or rerunning the accepted raw measurements.

Do not begin EXP-003 automatically.