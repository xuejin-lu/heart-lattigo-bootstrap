# Batch 022 — public-native E32 cold/warm Bootstrap

**Status:** `PASS_PUBLIC_BOOTSTRAP_COLD_WARM` / `READY_FOR_WEB_REVIEW`

**Primary source:** `37002ff3bbd97889cf0ec8108ff2cd527ab8d3d5`

**Standard:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (clean detached checkout)

**Fast:** `2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac` (clean detached checkout)

## Scope and integrity

Used the existing `cmd/perfprobe` public-native path and shared numerical/timing helpers. The frozen fixture is LogN13/E32, Q/P and deterministic 4096-slot workload, with A/B Scale `2^45`, C Scale `q5`, and the public `EncryptNew → AddNew → MulRelinNew → Rescale(q5) → RotateNew → DropLevelNew(4) → Bootstrap` lifecycle. Standard used genuine native key generation/encryption/decryption. Fast used the pinned intentionally insecure zero-secret mode. Both called the ordinary public `bootstrapping.NewEvaluator(...).Bootstrap` interface; no production arithmetic or Secondary source was changed.

The separate zero-call preflight passed all eight canonical checkpoints, matched environment and provenance, and the `compare-public` pair reported `PASS` at the fixed `1e-6` gate. Its worst pre-Bootstrap Fast-vs-Standard maximum complex difference was `7.781760773684476e-11`. Actual calls were exactly two per backend (four total), each on a fresh copy of that backend's same held Level0 ciphertext; the held-input fingerprints matched cold/warm within each backend. No retries or additional acceptance calls occurred.

## Timing and allocations

All times are one sample; allocated bytes and Go allocation counts are reported separately. “Evaluator+cold” is specifically `public_bootstrapping_evaluator_construction + first_cold_bootstrap`, not key generation or evaluation-key generation.

| Backend / phase | Time | Allocated bytes | Allocations |
|---|---:|---:|---:|
| Standard Bootstrap evaluator construction | 350.424 ms | 710,669,552 | 14,232,894 |
| Standard first/cold Bootstrap | 312.426 ms | 144,032,536 | 36,694 |
| Standard later/warm Bootstrap | 299.841 ms | 91,806,536 | 35,019 |
| Standard evaluator+cold | 662.850 ms | 854,702,088 | 14,269,588 |
| Fast Bootstrap evaluator construction | 2.443 ms | 8,853,664 | 32,085 |
| Fast first/cold Bootstrap | 536.207 ms | 960,789,168 | 19,282,477 |
| Fast later/warm Bootstrap | 70.399 ms | 7,505,112 | 8,085 |
| Fast evaluator+cold | 538.650 ms | 969,642,832 | 19,314,562 |

Key material setup was also measured separately: Standard secret-key generation `0.174 ms` / `198,448 B` / `208` allocations and `GenEvaluationKeys` `397.630 ms` / `395,320,432 B` / `94,257`; Fast secret-key generation `0.164 ms` / `198,448 B` / `210` and `GenEvaluationKeys` `819.350 ms` / `578,831,992 B` / `421,110`.

Fast's much shorter evaluator-construction phase and larger first-call phase are consistent with the pinned implementation's lazy Fast circuit initialization; the first call is not steady-state latency. Standard builds its circuit data in `NewEvaluator`. One cold and one warm sample are descriptive only and do not establish a stable speedup.

## Numerical result

Both backends produced native Level1 outputs at Scale `2^45`, degree 1, with two authoritative Q-prefix rows; both native decrypt/decode and per-backend pre-Bootstrap plaintext-oracle gates passed. Output vectors each contained exactly 4096 finite slots.

| Phase | Standard self RMSE / max error | Fast self RMSE / max error | Fast-vs-Standard RMSE / max complex difference | Gate |
|---|---:|---:|---:|---|
| First/cold | `4.627356909363751e-9` / `1.4141817696739812e-8` | `1.1795785978778852e-9` / `5.738790672737443e-8` | `4.7262101606249815e-9` / `4.489543437647087e-8` | PASS |
| Later/warm | `4.627356909363751e-9` / `1.4141817696739812e-8` | `1.1795785978778852e-9` / `5.738790672737443e-8` | `4.7262101606249815e-9` / `4.489543437647087e-8` | PASS |

The same fixed `1e-6` maximum-complex threshold applied to all pre-Bootstrap checkpoints and Bootstrap outputs. This result does not claim security equivalence. Formal Fast E32 internal stage/power tracing remains unavailable; no GPU/CNN performance claim is made.

## Validation

Passed:

- `GOCACHE=/private/tmp/fast-dropin-batch022-gocache go test ./internal/perfmeasure ./internal/numericalmetrics ./cmd/perfprobe -count=1`
- `GOCACHE=/private/tmp/fast-dropin-batch022-gocache go test -run '^$' -tags=perf_standard ./cmd/perfprobe`
- `GOCACHE=/private/tmp/fast-dropin-batch022-gocache go test -run '^$' -tags=perf_fast ./cmd/perfprobe`
- `GOCACHE=/private/tmp/fast-dropin-batch022-gocache go vet ./internal/perfmeasure ./internal/numericalmetrics ./cmd/perfprobe`
- `git diff --check`

`go test ./...` was not run: Batch022 requires focused tests and explicitly guards against accidentally invoking unbudgeted cryptographic Bootstrap tests. Unit tests use fake calls only. Full provenance, aggregate metrics and hashes are in the adjacent compact `*-evidence.json`; operation history is in `*-journal.md`. Raw decoded vectors and isolated temporary modfiles remain outside the repository.
