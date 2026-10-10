# FAST-DROPIN-MEASUREMENT-PLATFORM-REPAIR-AUTONOMOUS-BATCH-021

**Status:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW` — public-native E32 pre-Bootstrap gate passed. No Bootstrap, benchmark, LogN16, or Secondary production code was run/changed.

## Reuse and repair

The existing `cmd/perfprobe`, `internal/perfmeasure`, and `internal/numericalmetrics` platform was extended in place. The formal mode is explicitly selected with `--mode public-native --ephemeral-secret-weight 32`; the default remains the labeled `legacy-diagnostic` E0 path. Historical direct-c0 input and `NewFastEvaluator` behavior remain on the legacy path and are not promoted to formal evidence. The legacy input smoke passed for Standard and Fast at E0 under a zero-call budget; Fast retained observed c1=0 semantics. Only the exact historical Fast pin and the Batch021-approved Fast pin are accepted by that legacy adapter.

P1 reuse map, compatibility boundaries, remaining `fastdiag` gap, and the status matrix are recorded in [`docs/MEASUREMENT_PLATFORM.md`](../docs/MEASUREMENT_PLATFORM.md). The E32 workload helper delegates to shared profile/fingerprint logic and reproduces the accepted Batch019 workload fingerprint.

## Clean paired provenance

| Item | Value |
|---|---|
| Primary source commit | `981502b1b91dd2a0b50d69b0b32fe18bab9c8d88` |
| Primary measurement-source SHA-256 | `804feceeb6f8ccbb4ce7be20111ec78e3336f23d05bf0183b2630a1d6d9d1865` |
| Standard Lattigo | `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (clean detached checkout) |
| Fast Lattigo | `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` (clean detached checkout) |
| Profile / E / slots | LogN13 / 32 / 4096 |
| Config / QP / input / workload SHA-256 | `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98` / `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b` / `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285` / `00b70a2e77887c7d6a41db1859246f6513f2c984915de648734e5dd8c584cf74` |
| Public operation chain | `EncryptNew(A/B/C) → AddNew → MulRelinNew → Rescale(q5) → RotateNew(1) → DropLevelNew(4)` |
| Actual Bootstrap calls / budget | `0 / 0` for Standard and Fast |

The same committed Primary source was built using isolated temporary modfiles with `GOWORK=off`; the committed Primary `go.mod` and both pinned Lattigo repositories were unchanged. Both backends used native `EncryptNew`. Standard used ordinary key generation, `GenEvaluationKeys`, `bootstrapping.NewEvaluator`, and native `DecryptNew`. Fast used `GenEvaluationKeys → bootstrapping.NewEvaluator` dispatch; its c1=0 property was observed, not injected. Fast high-level checkpoints use the zero-secret c0 authoritative-Q-prefix measurement adapter; the terminal Level0 checkpoint uses native `DecryptNew`.

## Numerical and capacity gates

All eight checkpoints had matched Level/Scale/Degree and passed the fixed per-backend and pairwise `1e-6` maximum-complex-difference gate.

| Checkpoint | Level | Scale | Fast-vs-Standard complex RMSE | Max complex difference | Result |
|---|---:|---|---:|---:|---|
| `encrypt_a_level5` | 5 | `2^45` | `8.251111503934595e-12` | `2.533903731767853e-11` | PASS |
| `encrypt_b_level5` | 5 | `2^45` | `8.295098813825173e-12` | `2.410526822419659e-11` | PASS |
| `encrypt_c_q5_level5` | 5 | `q5` | `2.557443368663639e-16` | `9.112802267583663e-16` | PASS |
| `add_level5` | 5 | `2^45` | `1.1632120518068362e-11` | `3.1844983748742575e-11` | PASS |
| `mulrelin_level5` | 5 | `2^45·q5` | `2.3921236572666304e-13` | `9.246278832549843e-13` | PASS |
| `rescale_q5_level4` | 4 | `2^45` | `1.0505119033759885e-11` | `5.611470433440537e-11` | PASS |
| `rotate_level4` | 4 | `2^45` | `1.474977190390055e-11` | `5.564771587596444e-11` | PASS |
| `drop_level0` | 0 | `2^45` | `1.474977190390055e-11` | `5.564771587596444e-11` | PASS |

The maximum paired RMSE was `1.474977190390055e-11`; the maximum paired complex difference was `5.611470433440537e-11`. Fast encrypted input objects retain their historical full physical rows; starting at compact `AddNew`, q4/q5 are zero-length dormant rows at Level5, q4 is dormant at Level4, and Level0 contains only q0. The exact active row counts and hashes are in the compact runner evidence.

The capacity observer passed every strict Q0123 bound for the inputs and public operations, and the strict terminal bound: `B_rescale=2271135713118062`, so `2B=4542271426236124 < q0=36028797018652673`; Q0123 product was `5986308565615587353347023386369277282933624144412673`. The C input scale was exactly q5, and the post-Rescale scale returned to `2^45`.

## Budget, diagnostics, and remaining acceptance

The hard Bootstrap budget reserves each call in an exclusive journal before invocation, counts errors as consumed attempts, and cannot be reused after restart. Unit tests cover zero-, one-, and two-attempt policies plus over-budget refusal using stubs only. No actual Bootstrap was called. Evaluator construction was measured; first/cold and later/warm Bootstrap timings are recorded as unavailable, not zero.

The historical `cmd/fastdiag` / `internal/fastdiag` P93/E0 traces are not labeled as E32 Standard-comparable. Batch021 made no Secondary edits and did not invent Standard internal traces. The remaining Web decision is a separately authorized, exactly one-call Fast E32 public acceptance smoke on the terminal Level0 output through the verified ordinary `NewEvaluator` dispatch, followed by output metadata, Q-prefix, and native low-Level decode checks. Only after that evidence is accepted may a separate task measure first/cold and later/warm Bootstrap phases.

## Verification

- `GOCACHE=/private/tmp/heart-lattigo-batch021-gocache go test ./internal/perfmeasure ./internal/numericalmetrics ./cmd/perfprobe` — PASS.
- `GOCACHE=/private/tmp/heart-lattigo-batch021-gocache go test -run '^$' -tags=perf_standard ./cmd/perfprobe` — PASS.
- `GOCACHE=/private/tmp/heart-lattigo-batch021-gocache go test -run '^$' -tags=perf_fast ./cmd/perfprobe` — PASS.
- `GOCACHE=/private/tmp/heart-lattigo-batch021-gocache go vet ./internal/perfmeasure ./internal/numericalmetrics ./cmd/perfprobe` — PASS.
- `git diff --check` — PASS.
- Standard and Fast public-native LogN13/E32 dual-checkout preflights — PASS, 8 checkpoints each, 0 Bootstrap calls; `compare-public` — PASS.
- Legacy Standard/Fast E0 input-only smoke — PASS, 0 Bootstrap calls.

See [`journal`](FAST-DROPIN-MEASUREMENT-PLATFORM-REPAIR-AUTONOMOUS-BATCH-021-journal.md) for the source audit and bounded self-review; [`evidence.json`](FAST-DROPIN-MEASUREMENT-PLATFORM-REPAIR-AUTONOMOUS-BATCH-021-evidence.json) contains compact machine-readable provenance and aggregates.
