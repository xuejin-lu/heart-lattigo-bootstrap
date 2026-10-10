# Batch021 source audit and implementation journal

## P1 — Existing platform audit

- `cmd/perfprobe/main.go`: existing public Bootstrap timer, warmup/repetition loop, stage replay, provenance document, and backend adapter interface. Historically, actual calls included warmups, timing repetitions, and numerical trials. Reused this execution path and added one shared process-wide budget wrapper, durable exclusive pre-call reservations, and phase labels; no standalone stopwatch was added.
- `cmd/perfprobe/input.go`: existing `prepareAndValidateInput` / `validateInputCiphertext` input provenance and pre-decode quality checks. Legacy Fast still encodes c0 and zeros c1 directly. Formal public mode bypasses that constructor and uses `rlwe.NewEncryptor(...).EncryptNew`; it does not patch ciphertext components.
- `cmd/perfprobe/backend_fast.go`: legacy `NewFastEvaluator` retained and labeled. Formal mode uses a build-tag adapter for ordinary `bootstrapping.NewEvaluator` dispatch. Its embedded evaluator shape is checked against each pinned backend.
- `cmd/perfprobe/compare.go`: existing pair validation, metrics and report renderer reused for legacy comparison. v2 results remain readable; v3 is accepted only with `Mode=legacy-diagnostic`. The formal public comparison shares `compareVectors` and rejects empty/non-finite vectors, unapproved pins, mismatched provenance, absent checkpoints, state mismatch, and threshold failures.
- `internal/perfmeasure/profile.go`: old `ParametersFromConfig` stays E0-compatible; new `ParametersFromConfigWithE` validates and records explicit E. Fast acceptance is checked against actual pinned capability: E=0 or E=32, not arbitrary E.
- `internal/perfmeasure/public_workload.go`: the only workload addition is the frozen Batch019 A/B/C operation fixture plus the exact Q0123 and q0 capacity observer. Unit tests pin the accepted A/input and full workload fingerprints.
- `internal/numericalmetrics` remains the existing numerical metric implementation; `cmd/fastdiag`, `scripts/fastdiag`, Secondary `internal/fastdiag`, `fast_measurement*.go`, and `stage_runner.go` were not copied or replaced.
- `fastdiag`'s P93/E0 checkpoint assumptions do not express this E32 public pre-Bootstrap lifecycle. In-circuit internals are therefore explicitly unavailable/non-comparable; no Standard internal hook was invented.

## P2/P3 — Repair details

- CLI now separates `legacy-diagnostic` (default, historical E0) from `public-native` (requires explicit E and, for this fixture, E32). Legacy input-only mode remains usable, with the old Fast pin plus only the Batch021-approved Fast pin; arbitrary commits are still refused.
- Formal public mode performs native input construction, public `GenEvaluationKeys`, `NewEvaluator` dispatch, original Batch019 operation ordering, per-checkpoint native Standard decryption, Fast c0 authoritative-prefix observation and terminal native Level0 decryption. Fast c1=0 is observed on input and every checkpoint.
- Per-checkpoint evidence records Level, Scale, Degree, physical row lengths, active row hashes, and oracle metrics. The Fast prefix compaction check begins at `AddNew`: accepted Batch019 evidence shows `EncryptNew` initially owns full rows (`fast_compact_prefix_layout=false`), while `AddNew` and later outputs are compact. The output gate checks dormant q4/q5 as appropriate for each Level.
- `bootstrapBudget.invoke` creates an exclusive per-attempt JSON reservation before the underlying method runs. Failed calls consume budget, over-budget calls do not execute, and a restarted run cannot reuse an output/journal base. Tests exercise only stub calls. The current public E32 gate fixes budget/calls to 0/0.
- Evaluator construction has individual wall/allocation samples. The report distinguishes first/cold and later/warm Bootstrap phases, but marks both unavailable because no Bootstrap method was invoked.

## P4 — Paired evidence provenance

- Primary main: `981502b1b91dd2a0b50d69b0b32fe18bab9c8d88`, clean at both runs.
- Standard checkout: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, detached and clean.
- Fast checkout: `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`, detached and clean.
- Separate temporary modfiles selected the dependencies; `GOWORK=off`. `go list -m -json` identified the exact replacement directories. Primary `go.mod` and both Lattigo code checkouts were untouched.
- Frozen config/input/workload/QP fingerprints all matched Batch019; E=32 and 4096 slots. Each backend emitted all eight required checkpoints. All per-backend oracle errors and all paired differences were below `1e-6`; all state tuples matched. Actual Bootstrap calls were zero.
- Fast encrypted inputs retained full physical rows as the pinned `EncryptNew` path and Batch019 record; at the first compact output, `AddNew`, q4/q5 were empty dormant rows. Compactness persisted through `MulRelinNew`, `Rescale`, `RotateNew`, and `DropLevelNew`.
- Capacity: `B_rescale=2271135713118062`, `2B=4542271426236124`, `q0=36028797018652673`; Q0123 product `5986308565615587353347023386369277282933624144412673`. Strict `2B<q0` passed.
- The separate legacy E0 input-only smoke passed on Standard and Fast with no Bootstrap call. It remains diagnostic evidence only and is not compared/promoted as the formal E32 result.

## Bounded self-review findings

1. The initial row gate incorrectly required Fast `EncryptNew` to be compact. The Batch019 source evidence records full physical rows at all three encrypted inputs and compact rows starting at `AddNew`; the gate was corrected to match that established boundary without changing production code.
2. Generic RLWE `DecryptNew` cannot consume a Fast high-level compact ciphertext whose non-authoritative q4/q5 coefficient rows are empty. A first run exposed a panic in that measurement adapter before any Bootstrap. The final Fast observer reads only the authoritative c0 Q-prefix for high-level checkpoints and runs genuine `DecryptNew` at the terminal Level0 checkpoint, which the spec explicitly requires. Standard uses native decryption at every checkpoint. Decode paths are recorded and checked by the paired comparator.
3. Legacy input smoke revealed that its historical Fast pin check rejected the expressly approved Batch021 Fast commit. The adapter now preserves the old pin and additionally admits exactly the approved pin; a regression test confirms arbitrary pins remain rejected.
4. Pair validation originally accepted two artifacts with the same incomplete checkpoint set. It now requires the exact eight canonical stages, approved backend pins, clean provenance, per-backend oracle passes and expected decode/compact declarations.

These repairs are limited to the existing Primary measurement platform. No Bootstrap was called, no numerical threshold or cryptographic arithmetic was changed, no Secondary code was modified, and no LogN16/benchmark task was started.
