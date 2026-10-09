# FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R Journal

## Checkpoint A — combiner repair

- Start: Primary `main@2794bcc55d67c158a81fe704a173c4a03d5bed73`, clean after mandatory sync. Secondary `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` and pinned Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40` were clean and read-only.
- Added regression coverage for `summaryOf` input immutability, nonempty comparisons, empty/mismatched/nonfinite inputs, valid exact-zero pairs, eight-checkpoint shape, exact backend/input provenance and nonzero pair results.
- Before the repair, the focused test reproduced: `summaryOf` cleared the source decoded slice; two empty vectors returned finite zero metrics; unequal lengths panicked during indexing.
- Repair: summary conversion now copies checkpoint elements before stripping `Decoded`; `compare` rejects empty/unequal/nonfinite samples; combine validates exact pins, frozen profile/Q/P/input/runner/lifecycle/runtime metadata, all eight checkpoint identities/states and 16 finite samples, plus plaintext gates, before emitting compact paired metrics.
- No workload values, operation sequence, backend code, key semantics, or tolerances were changed.

## Checkpoint B — fresh matched evidence

- Standard module resolution verified as `/private/tmp/fast-standard-rebaseline-002/lattigo-standard`, clean at the pinned commit; Fast resolution verified as `../lattigo`, clean at the pinned `fast-qprefix` commit.
- Exactly one runner execution per backend: genuine Standard native keygen/encrypt/decrypt lifecycle and Fast zero-secret lifecycle. Primary run identity was `2794bcc55d67c158a81fe704a173c4a03d5bed73` with task changes present (`primary_dirty=true`).
- Both complete raw outputs were combined while their decoded arrays were still present. All 8 comparisons used 16 samples; all backend plaintext gates, state/row checks, and Fast direct key-lookup checks passed.
- Compact evidence: `results/FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R-evidence.json` (13,471 bytes; decoded vectors excluded). Raw output JSON stayed in `/private/tmp` and is not part of the repository diff.
- Profile SHA-256 `f1d7944ea3febd9731cfd846af71c8596089893d014b729e489fc694d27f7ebc`; effective Q/P SHA-256 `41c103fb3338c8fca82ab0c5b2e4dfd17cf1affc6d7295e082a6ba4a6e372324`; deterministic input SHA-256 `296447e0602fcaec7839fd302ec783dd0994418aad18757c20cfe9d38323cd24`; runner SHA-256 `a65af37101c255f55252d94e94ff017d51e34ed618d2efb368ee5dd70496612a`.
- Bootstrap calls: 0. Benchmark runs: 0. LogN16 runs: 0. Bootstrap-specific ephemeral-secret `E` is N/A for this primitive-only workload.

## Checkpoint C — self-review and submission

- Focused Primary test: `GOWORK=off GOCACHE=/private/tmp/fast-dropin-015r-gocache go test ./tools/fast-dropin-compact-consumers-batch-015` — passed.
- Focused vet: `GOWORK=off GOCACHE=/private/tmp/fast-dropin-015r-gocache go vet ./tools/fast-dropin-compact-consumers-batch-015` — passed.
- `git diff --check`: passed. Final authorized file set is the runner, its regression tests, and the three 015R report artifacts only.
- Self-review found no remaining task defect: pair comparison precedes lossy summary serialization; malformed/empty/nonfinite samples fail closed; backend, workload, source, shape, and runtime provenance are checked; original gates and Q-prefix rows remain fixed.
- Commit and normal fast-forward push are completed under the Primary repository's standing safe-push authorization; final SHA and remote state are in the task handoff.
- Secondary and pinned Standard remained read-only and clean; no next batch was started.
