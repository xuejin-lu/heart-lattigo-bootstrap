# FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012 Journal

**Batch state:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`

**Current checkpoint:** D — read-only coverage inventory `PASSED`
**Bootstrap budget:** Standard `1/1 completed`; Fast `1/1 completed`

## Resume preflight — 2026-10-10

- Primary `main` safely synchronized; `HEAD == origin/main == 2321df7bf761585a05b9f116abe2c390a29904b5`, clean before this checkpoint's reporting repair.
- Secondary `fast-qprefix` safely synchronized; `HEAD == origin/fast-qprefix == 00ac70ba136d190fa31bbb26c2f51d003a221634`, clean. No Secondary files changed.
- Re-read the active Primary batch charter, key-plan spec, research workflow §4B, and Secondary `AGENTS.md`, `CURRENT_TASK.md`, and `docs/FAST_QPREFIX_SPEC.md`.
- Current task remains `FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012`; checkpoint C is authorized because both B gates passed.

## Pinned startup state

- Primary `main`: startup sync began at `92253f315b28826e6c3d188ca80a5724f9a9055f`; no task files were modified before the synchronized charter was read.
- Secondary `fast-qprefix`: `00ac70ba136d190fa31bbb26c2f51d003a221634`, clean; no Secondary production changes authorized or made.
- Genuine Standard checkout: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean.
- Frozen config SHA-256: `919a2d9409b8ddeb720458d3112855f87d7a69e0cc39aad825b0ddf794769c98`.
- Frozen deterministic input SHA-256: `d9151964e398ae9fb77248394b4b28f84c9e5737c621cf5e5b0570e343dcc285`.

## Checkpoint A — corrected key-plan evaluator

**Status:** `PASSED_IMPLEMENTATION_AND_COMPILE`

**Commit:** `69868b2` (`test: add Bootstrap keyplan-compatible Rotate runner`)
**Frontend:** `tools/fast-dropin-rotate-bootstrap-keyplan-011/main.go`; identical source selected by Standard/Fast GOWORK.

- Public pre-Bootstrap Rotate now constructs `ckks.NewEvaluator(btpParams.BootstrappingParameters, keys.MemEvaluationKeySet)`; the Level-0 input still comes from the unchanged residual parameters and is not raised or padded.
- Before Rotate, the runner asserts same N, Standard ring type, exact ordered residual-Q prefix and q0 equality; it verifies the Bootstrap plan contains the requested Galois key, checks Standard layout/P level against evaluator `MaxLevelP`, and compares the returned Bootstrap secret's q0 residues to the original secret without serializing secret data.
- Source basis for these assertions: Standard `circuits/ckks/bootstrapping/parameters.go` copies all residual Q primes into the beginning of Bootstrap Q; Standard `bootstrapping/keys.go` and Fast `bootstrapping/fast_keys.go` extend the same-N residual secret into Bootstrap Q/P. Fast's `GenEvaluationKeys` dispatches to `GenFastEvaluationKeys` for this profile.
- Runner-boundary panic recovery serializes failure evidence with frontend SHA captured before setup/operations; it does not continue after recovery.
- Verification passed under both pinned GOWORKs: `go test ./tools/fast-dropin-rotate-bootstrap-keyplan-011 -count=1`; `go vet ./tools/fast-dropin-rotate-bootstrap-keyplan-011`.
- The identical frontend was committed before the paired preflights. The frontend SHA used by B was `14435c478ee9af0ba2ae0e3548850ed3c121b69141c4e0918d8a05bee4650e7f`.
- A consumed **Bootstrap** call count: Standard `0`, Fast `0`.

## Checkpoint B — paired public Rotate preflights

**Status:** `PASSED`

- Attempt counts: Standard `1/1`; Fast `1/1`. Both used the same frontend SHA above, canonical config/input hashes, and matching actual residual/bootstrap Q and Bootstrap P identities. No preflight was repeated during resume.
- Standard raw: `/private/tmp/fast-dropin-keyplan-011-standard-preflight.json` (960,166 bytes); Fast raw: `/private/tmp/fast-dropin-keyplan-011-fast-preflight.json` (960,074 bytes). Raw files remain outside the repository; committed evidence is compact and excludes decoded arrays.
- Both report `preflight_passed`, clean pinned provenance, exact residual-Q prefix and q0, same N=8192 / Standard ring / Level0 / Scale 2^45 / q0 backing, and 30 planned rotation keys.
- Genuine Standard path: native EncryptNew c1 is nonzero (8192/8192); Rotate does one `GetGaloisKey`, zero list calls; oracle RMSE `1.342087606921074e-11`, max complex error `5.379396273494723e-11`, SNR `183.67157992669885 dB`.
- Fast path: c1 remains zero before and after Rotate; Rotate makes zero Galois-key and list calls; oracle RMSE `7.44997342660252e-13`, max complex error `2.1019564094557047e-12`, SNR `208.78410277402207 dB`.
- Direct decoded Fast-vs-Standard Rotate delta: RMSE `1.3396909694030365e-11`, max complex error `5.3631037691780624e-11`, SNR `183.68710464953378 dB`.
- Compact schema self-review found the nested `evaluation_keys.rotation_galois_element` field had been serialized as zero despite a captured top-level element `5` and a present matching key. The runner now records it directly; comparison of the immutable B raw captures transparently derives the missing nested field from their captured top-level value. It also validates pinned provenance, expected dispatch/key lookups, c1 contracts, phase status/call budget, and deterministic repair ordering. B was **not rerun**.
- Repaired comparison artifact: `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-evidence.json` (10,349 bytes). Its `evidence_repairs` field explicitly records the two derived nested fields. The B run-source hash remains the original SHA above; the repaired shared frontend SHA for checkpoint C is `a19841d94311d95a6f0cbab9a3301e49d5820fb5660885f1037a088e1a9204ad`.
- Reporting repair tests and vet passed under both pinned workspaces. At the B checkpoint boundary, Bootstrap budget was Standard `0/1`; Fast `0/1`.

## Checkpoint C — single bounded public Bootstrap composition

**Status:** `PASSED` — one Standard and one Fast public Bootstrap completed exactly once each.

- Standard attempt count `1/1`; one call completed successfully with `bootstrap_completed`; input was the pre-rotated ciphertext. Provenance: Primary `30a312996b5916de97b1e14b4af4438d95660ebe`, shared frontend SHA `a19841d94311d95a6f0cbab9a3301e49d5820fb5660885f1037a088e1a9204ad`, clean pinned Standard `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- Standard output: Level 1, Scale log2 `45`, degree 1, 4096 slots, full active Q rows, NTT/non-Montgomery; native Standard c1 is nonzero across 16,384 active coefficients.
- Standard Bootstrap-vs-rotated-cleartext oracle: RMSE `4.885854924910455e-9`, max complex error `1.5351218791077742e-8`, SNR `132.44838589971747 dB`; all recorded metrics finite.
- Standard raw evidence: `/private/tmp/fast-dropin-keyplan-011-standard-bootstrap.json` (1,309,325 bytes), kept outside the repository; it contains decoded vectors and is not printed or committed.
- Fast attempt count `1/1`; one call completed successfully with `bootstrap_completed`; input was the pre-rotated ciphertext. Provenance: Primary `a682e8616e441f0c9fba15b3bf888187fba276d2`, same shared frontend SHA/config/input, clean pinned Fast `00ac70ba136d190fa31bbb26c2f51d003a221634`.
- Fast output: Level 1, Scale log2 `45`, degree 1, 4096 slots, full active Q rows, NTT/non-Montgomery; zero c1 across 16,384 active coefficients. Fast dispatch selected; Rotate made zero Galois-key lookups.
- Fast Bootstrap-vs-rotated-cleartext oracle: RMSE `1.1814120448423705e-9`, max complex error `5.771241651805313e-8`, SNR `144.77916936116213 dB`; all recorded metrics finite.
- Direct decoded Fast-vs-Standard Bootstrap delta: RMSE `4.954605783029611e-9`, max complex error `4.3240068697009917e-8`, SNR `132.3270151591381 dB`; finite.
- Fast raw evidence: `/private/tmp/fast-dropin-keyplan-011-fast-bootstrap.json` (1,309,603 bytes), kept outside the repository; it contains decoded vectors and is not printed or committed.
- Compact C comparison artifact: `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-bootstrap-evidence.json` (12,861 bytes), records each run's actual Primary commit separately while confirming the same measured frontend SHA `a19841d94311d95a6f0cbab9a3301e49d5820fb5660885f1037a088e1a9204ad`, fixed config/input, pinned backend identities, and exactly one call each.
- A reporting-only follow-up now permits distinct clean Primary commits when source/config/input hashes match and retains both commits in the comparison. It does not weaken experiment gates or alter the already-executed frontend. Tests/vet passed under both pinned workspaces. Current comparison-runner source SHA after that follow-up: `2ae57c525fa39fb8c3ec35a646a92da00e364b441a8a076aec62f854404c2ba7`.

## Checkpoint D — read-only coverage inventory

**Status:** `PASSED`.

- Created `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-coverage-audit.md` with source-backed operation coverage for EncryptNew, Add, MulRelin, Rescale, Rotate, Bootstrap, and DecryptNew; it distinguishes numerical evidence, selected Fast boundary paths, unproven per-primitive dispatch, P0/C0 limits, security semantics, and runtime speed evidence.
- The audit recommends four bounded follow-up I/E task areas (Add/Sub, MulRelin operands, public Rescale, explicit compact-Q boundary contract); none was started or implemented.
- No additional Bootstrap, benchmark, or Secondary edit occurred. Secondary remains clean at `00ac70ba136d190fa31bbb26c2f51d003a221634`.
- Final runner test/vet commands passed in both pinned workspaces after the last code change; `git diff --check` passed. All report/evidence changes are confined to Primary `results/`.

## Batch result

`BATCH_COMPLETE_READY_FOR_WEB_REVIEW` / `READY_FOR_WEB_REVIEW`.
Final compact deliverables: this journal, `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-summary.md`, B evidence `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-evidence.json`, C evidence `results/FAST-DROPIN-INTEGRATION-AUTONOMOUS-BATCH-012-bootstrap-evidence.json`, and the D coverage audit above. No further Bootstrap work is authorized or required in this batch.
