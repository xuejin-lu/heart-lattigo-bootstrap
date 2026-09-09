# CKKS bootstrapping experiment harness

`heart-lattigo-bootstrap` is the control plane for reproducible CKKS
bootstrapping experiments. The primary source, configuration, workload, and
measurement method stay fixed while the Lattigo checkout at `../lattigo` is
changed between Standard and Fast implementations.

The old hardware golden model is preserved at the `legacy-hardware-model`
branch and is not part of the active `main` experiment path.

## Run the harness

Requirements: Go 1.25 or newer, with the Lattigo checkout available at
`../lattigo` (or set `LATTIGO_REPO` for metadata collection).

```text
go test ./...
go run . -config configs/bootstrap_config.example.json -out results/exp-000.json
```

The runner performs the configured warm-up count, then measures complete
bootstrap calls for the configured repetition count. It writes JSON containing
per-bootstrap elapsed time, allocation metrics, effective parameters, and
reproducibility metadata for both repositories and the host.

Use `-repetitions` and `-warmup` to override the JSON configuration without
editing it:

```text
go run . -config configs/bootstrap_config.example.json -repetitions 3 -warmup 1
```

## Standard/Fast compatibility check

The `go.mod` replacement points to the sibling Lattigo checkout. Keep the
primary repository and command unchanged, and change only that checkout:

```text
# Run A: Standard
cd ../lattigo && git checkout main
cd ../heart-lattigo-bootstrap
go run . -config configs/bootstrap_config.example.json -repetitions 3 -warmup 1 -out /tmp/standard.json

# Run B: Fast
cd ../lattigo && git checkout fast-ckks
cd ../heart-lattigo-bootstrap
go run . -config configs/bootstrap_config.example.json -repetitions 3 -warmup 1 -out /tmp/fast.json
```

The frontend calls the ordinary bootstrapping construction API in both runs;
it has no Standard/Fast selector and does not import Fast-only packages. The
Fast branch adapts that public construction path inside Lattigo and keeps the
Fast evaluator behind the backend boundary.

## Parameter adaptation

The historical configuration values remain the experiment input. Because the
current Fast Stage-A contract requires a Standard ring with a residual level of
at most one, `q0` and the first `q_slots_to_coeffs` prime form the residual
ciphertext. Lattigo's bootstrapping parameter builder then creates the full
circuit chain from the configured DFT factorization and EvalMod settings. The
effective generated parameters are recorded in each result.

The legacy `q_circuit_slots`, `q_eval_mod`, and `log_bsgs_ratio` fields are
not active knobs in this compatibility path and are removed from the JSON
schema. The loader rejects them as unknown fields so an old configuration
cannot silently claim to change the workload.

## Continuous integration

GitHub Actions checks out this repository and `xuejin-lu/lattigo` as sibling
directories, then runs the same `go test ./...` command against both the
Standard `main` backend and Fast `fast-ckks` backend. The full end-to-end
runner remains the documented local acceptance command because CI is limited
to compile/unit tests.

## Repository layout

```text
.
├── AGENTS.md
├── CURRENT_TASK.md
├── bootstrap_config.go
├── configs/
├── main.go
├── results/
├── runner.go
└── specs/
```
