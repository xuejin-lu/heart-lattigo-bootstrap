# FAST-DROPIN-HIGHLEVEL-TO-BOOTSTRAP-AUTONOMOUS-BATCH-018

**Result:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`
**Date:** 2026-10-10
**Runner commit:** `86af97092ba666a4a16ab2e7c4dde861ee669a96` (Primary clean during both accepted runs)

## Scope and provenance

This bounded LogN13 experiment composed public high-Level arithmetic with the already-frozen canonical Batch 017 E32 Bootstrap profile. Standard used the genuine, unmodified Lattigo pin `5dbffbdea05394de2ca3a432ed5318aa832e3f40`; Fast used `xuejin-lu/lattigo` `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`. Both trees were clean and read-only during the accepted runs. No Secondary code was changed.

Both builds used the same Primary frontend and deterministic original A/B workload, generated full-profile keys through public `GenEvaluationKeys`, and executed:

`full-profile Level-5 public EncryptNew(A,B) → AddNew → RotateNew(1) → DropLevelNew(5) → residual Level-0 native DecryptNew/Decode → public E32 Bootstrap → residual native DecryptNew/Decode`.

The frozen config SHA-256 is `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`; Q/P SHA-256 is `1f045e603a856968779d62e045a037274bba08cbfce8b1dd3dec2828f1f6a46b`; canonical input SHA-256 is `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`. The effective profile was LogN13, 4096 slots, E32, residual max Level 1, full Bootstrap max Level 16, and Scale `2^45`.

## High-Level Q-prefix and capacity evidence

Fast full-profile encryption inputs at logical Level 5 contain six allocated Q rows. The public Fast Add and Rotate outputs remained Level 5 / Scale `2^45`, authoritative width 4, and had q4+ row backing lengths zero. `DropLevelNew(5)` produced Level 0 at the same Scale with q0 as the sole authoritative row. Fast c1 remained zero. The pinned public evaluator dispatches eligible Add through the shared Fast Add/Sub core and Rotate through the shared Fast automorphism core; the compact output layout and existing Fast capacity observer corroborated the Q-prefix execution. Standard used its native public evaluator and ordinary randomized ciphertext lifecycle.

For Fast's zero-secret encoded c0 representation, the independent bounds were `B_A=2348557866436`, `B_B=1804822278899`, and `B_Add/Rotate=4153380145335`. At the target Level 0, `2B=8306760290670 < q0=36028797018652673`; at Level 5, the q0..q3 product was `5986308565615587353347023386369277282933624144412673`. The existing observer passed at both inputs, Add, Rotate, and DropLevel. This bound is not asserted for Standard's randomized RLWE components; Standard was checked by its native cleartext oracle.

## Numerical results

All cleartext and paired max-error gates remained fixed at `1e-6`.

| Checkpoint | Standard | Fast | Fast vs Standard |
|---|---:|---:|---:|
| Before Bootstrap, native Level-0 oracle max error | `7.1795923428e-11` | `3.3770666929e-12` | RMSE `1.5656982926e-11`; max `7.1377417263e-11` — PASS |
| After Bootstrap, native cleartext oracle max error | `1.8458745857e-8` | `5.7649062389e-8` | RMSE `4.9639272689e-9`; max `3.9338641341e-8` — PASS |
| Post-Bootstrap SNR | `134.4577446217 dB` | `146.7678416952 dB` | — |

Each backend executed exactly one public Bootstrap, on the same backend-specific B-preflight ciphertext retained in memory; there were no retries. Both returned 4096 decoded slots at Level 1 / Scale `2^45`. The preflight and post-Bootstrap native decryption/decoding gates passed.

This result demonstrates the bounded public composition for this frozen profile and workload; it makes no performance claim and does not establish broader numerical or parameter generality.

## Validation

- Fast build: `GOCACHE=/private/tmp/batch018-gocache go test ./tools/fast-dropin-highlevel-bootstrap-batch-018 -count=1` — PASS
- Standard build: `GOWORK=/private/tmp/batch016-standard-work/go.work GOCACHE=/private/tmp/batch018-gocache go test -tags lattigo_standard ./tools/fast-dropin-highlevel-bootstrap-batch-018 -count=1` — PASS
- Fast vet: `GOCACHE=/private/tmp/batch018-gocache go vet ./tools/fast-dropin-highlevel-bootstrap-batch-018` — PASS
- Standard vet: `GOWORK=/private/tmp/batch016-standard-work/go.work GOCACHE=/private/tmp/batch018-gocache go vet -tags lattigo_standard ./tools/fast-dropin-highlevel-bootstrap-batch-018` — PASS
- `git diff --check` — PASS

See the companion journal for the bounded checkpoint log and the compact machine-readable evidence JSON for provenance, row hashes, capacity observations, and exact aggregate metrics.
