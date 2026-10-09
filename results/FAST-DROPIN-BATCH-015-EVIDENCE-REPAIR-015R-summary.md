# FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R

**Status:** `BATCH_COMPLETE_READY_FOR_WEB_REVIEW`

**Scope:** Primary evidence combiner repair and one fresh matched public Add → MulRelin → Rescale → Rotate run per pinned backend. No Bootstrap, benchmark, or LogN16.

## Result

The historical Batch 015 direct Fast-vs-Standard pairwise columns (`RMSE=0`, `max=0`) are invalid because `summaryOf` mutated the original checkpoint slice before comparison and `compare` accepted two empty vectors. Only those paired columns are superseded here. The original 015 summary, journal, and evidence remain unchanged as forensic history; its separate plaintext-oracle evidence is not reclassified by this repair.

The repaired combiner retained and compared all **16 decoded samples at each of eight matched checkpoints** before making compact summaries. All pair metrics were finite and nonempty. Standard and Fast each passed the original per-checkpoint plaintext gates: Add max error ≤ `1e-6`; MulRelin, Rescale, and Rotate max error ≤ `1e-4`. State checks passed at every checkpoint: matching Level, Degree, Scale, NTT/non-Montgomery representation and expected row counts. Fast retained zero `c1`; direct relinearization and Galois-key retrieval counts were zero (key-list enumeration count 1 at evaluator construction). Bootstrap calls: 0.

| Checkpoint | Level | log2 Scale | Fast-vs-Standard complex RMSE | Fast-vs-Standard max complex error |
|---|---:|---:|---:|---:|
| `add-l1` | 1 | 45 | 6.329845366478272e-13 | 1.0105182222775388e-12 |
| `mul-relin-l1` | 1 | 90 | 1.1020884790199848e-13 | 2.2117492161824084e-13 |
| `rescale-l1` | 0 | 51.00000021497571 | 1.1167344098924615e-13 | 2.1813066076952567e-13 |
| `rotate-l0` | 0 | 51.00000021497571 | 1.0908691599533871e-13 | 2.15991017672228e-13 |
| `add-l3` | 3 | 45 | 7.788804037185317e-13 | 1.41288837464095e-12 |
| `mul-relin-l3` | 3 | 90 | 1.174081903812359e-13 | 3.2225890565381054e-13 |
| `rescale-l3` | 2 | 50.999999441053866 | 1.1603436858949153e-13 | 3.1621121862998285e-13 |
| `rotate-l2` | 2 | 50.999999441053866 | 1.1600555709010704e-13 | 3.0785875691030975e-13 |

The complete compact evidence, including each backend's plaintext RMSE/max, checkpoint states, hashes and key counters, is in [`FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R-evidence.json`](FAST-DROPIN-BATCH-015-EVIDENCE-REPAIR-015R-evidence.json). It contains no decoded vectors.

## Frozen profile and provenance

- Primary during both runs: `2794bcc55d67c158a81fe704a173c4a03d5bed73`, dirty (`true`) because this task's runner and regression tests were present. The artifact records this honestly.
- Shared runner SHA-256: `a65af37101c255f55252d94e94ff017d51e34ed618d2efb368ee5dd70496612a`.
- Genuine Standard: clean detached `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, resolved by Go to the separate Standard checkout.
- Fast: clean `fast-qprefix@2d6145d7e1db0ca7351eb47a03e1b352fc4ef9ac`.
- Go/runtime: `go1.26.4`, `darwin/arm64`.
- Profile: LogN13; Q bit sizes `[55,39,40,39]`; P `[60]`; scale `2^45`; `LogSlots=4`; Levels `[1,3]`; ternary secret weight `H=192`.
- Profile SHA-256: `f1d7944ea3febd9731cfd846af71c8596089893d014b729e489fc694d27f7ebc`.
- Effective Q/P SHA-256: `41c103fb3338c8fca82ab0c5b2e4dfd17cf1affc6d7295e082a6ba4a6e372324`.
- Deterministic input SHA-256: `296447e0602fcaec7839fd302ec783dd0994418aad18757c20cfe9d38323cd24`.
- Bootstrap-specific ephemeral-secret weight `E`: not applicable to this primitive-only chain; no Bootstrap parameter object or call was used. No E value was fabricated or changed.

## Validation

- Regression-first test run against the old combiner reproduced source-slice mutation, empty-vector acceptance as zero error, and a length-mismatch panic.
- `GOWORK=off GOCACHE=/private/tmp/fast-dropin-015r-gocache go test ./tools/fast-dropin-compact-consumers-batch-015` — passed.
- `GOWORK=off GOCACHE=/private/tmp/fast-dropin-015r-gocache go vet ./tools/fast-dropin-compact-consumers-batch-015` — passed.
- One Standard run and one Fast run used identical runner/profile/input. Raw runner outputs stayed in `/private/tmp`; only compact paired evidence is committed.
- `git diff --check` and final Git submission state are recorded in the task journal / handoff.
