# BTS-FAST-DECODE-007 — result

**Disposition: `NEEDS_WEB_REVIEW`**
**Fast status: `FAST_BTS_DECODED`**
**Standard/Fast status: not compared; no `STANDARD_FAST_OUTPUT_COMPARED` claim.**

The unchanged LogN16 BTS completed once on the Fast zero-secret backend and produced a finite 32,768-slot decoded vector. The one Genuine Standard process was terminated with exit code 137 after `NewBTS()` and before the `[enc]` line and `bts.Bootstrap()`; its captured log has no error explanation. No Standard Bootstrap or decoded vector was produced. This Batch does not relaunch that Standard process, and the output-difference metrics cannot be calculated from the available evidence.

## Fast result

| Item | Result |
|---|---:|
| Profile | LogN16, SlotLimbs=15, `BTSModePlain`, seed=0, amp=1, 32,768 real slots |
| `NewBTS()` initialization | 45.084 s |
| First complete Fast `bts.Bootstrap()` | 7.369 s, success |
| Input ciphertext | Level 3, scale `2^32` |
| Returned ciphertext | Logical Level 18, scale `2^32`, degree 1, NTT=true, Montgomery=false |
| Fast projected plaintext | Level 3, scale `2^32`; returned ciphertext remains Level 18 |
| Effective prefix / actual backing | q0–q3; 4 backed rows in each component; q4+ remained absent |
| Decoded output | 32,768 slots; all finite |
| Input vector SHA-256 | `5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373` |
| Fast decoded vector SHA-256 | `577c0ff802da8036ce2557cc37cac04416f0b185d5393e1de617083b35e17f0b` |
| Snapshot SHA-256 | `0b03f7df93e53f544071ae6ee77d33e1015d41b42c509914ed5b773a960ea277` |
| Fast binary SHA-256 | `af13b685b1cfb16a2cdb927f88ba55109408df8f6876816a03d5040817dfebc8` |
| Fast backend | `ee308c9743000b5f281a84a871e7c4bd59b176c4` |

The BTS log reports Fast-vs-original-input max absolute error `2.525e-05` at slot 18078 and mean absolute error `4.534e-06`. From those logged rounded values, their approximate decimal logs/orders are `log10(max) ≈ -4.597739` (`10^-5`) and `log10(mean) ≈ -5.343518` (`10^-6`). These are Fast-vs-input errors, **not** Fast-vs-Standard differences.

The real post-Bootstrap ciphertext snapshot was restored offline with the same Fast commit and effective BTS parameters; ordinary `DecryptNew → Decode` reproduced every saved float64 output slot bit-for-bit. The verification command was:

```text
go test -modfile=/private/tmp/bts-fast-e2e-first-006.A1qAGy/adapter/fast.mod -run '^TestTask007ReturnedCiphertextSnapshotRestoresAndDecodes$' -count=1 -v
```

It did not call Bootstrap. The test also confirmed logical Level 18, plaintext Level 3, the four backed prefix rows, and absent dormant q4+ rows.

## Genuine Standard attempt and comparison limitation

- Pinned Standard source: `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean detached checkout.
- Independently built Standard binary SHA-256: `946ac52d456f2fcc58850f8aa4cb3d75310ea8a184810f103353d747e22a9a01`; `go version -m` confirms the module replacement points to that Standard checkout, not Fast.
- Its single process invocation exited 137 after logging `NewBTS()` = 1m30.373s and before `[enc]` / `bts.Bootstrap()`. The captured stdout/stderr contains no cause; this report does not infer OOM or another cause.
- It did generate the same canonical input vector before termination; its SHA-256 matches the Fast input (`5ca151755bd8232e0b581939fb1b2c079fa4451b454c0e1b0ad9f2b1a6198373`).
- Standard Bootstrap calls entered: 0/1. No Standard output ciphertext or decoded vector exists. The historical Task 006 Standard max/mean error summary has no slots and is not a substitute for a paired vector.
- Consequently `MaxAbsDiff`, `RMSE`, `MeanAbsDiff`, and new paired Standard/Fast max errors are unavailable. No speedup is calculated.

## Provenance, tests, and limitations

- Original BTS source/config manifest SHA-256: `53205b37abd1fb57a960394256b958f5d5e5b1a66f527bfb7cbadabfe3cf0cda`; `external/BTS/` was not modified.
- Fast measurement-only overlay `main.go` SHA-256: `b788a13b5985c89a70f02b64c2b2bcd067e76fbe9b67c4981e673ead4cc515bd`.
- Passing checks: focused projection/decode/snapshot tests; `go test ./core/rlwe -count=1`; `go test ./schemes/ckks/fast -count=1`; `go vet ./core/rlwe ./schemes/ckks ./schemes/ckks/fast`; actual snapshot restore/decode test; `git diff --check`.
- `go test ./schemes/ckks -count=1` is not clean on this Fast checkout. On pristine `e05283690fcb0262540a2f979b570dc979f49682`, the focused `Mul/Ct/Ct/Degree0` test reproduces the same four Fast MulRelin error-message mismatches. The pristine full suite also panics in the ConjugateInvariant RescaleTo test when generic decrypt reads absent q4/c1 backing. The current full-suite run continues to fail in legacy Fast Rescale/precision and unsupported compact ConjugateInvariant paths; those are not claimed as passing or all individually proven baseline-equivalent here.
- No full test sweep that could invoke Bootstrap, no additional EvalMod/C2S/S2C, no trace, no warmup, and no Fast retry was run.
- Fast zero-secret mode remains intentionally insecure. The Level-3 plaintext is an experimental Q-prefix projection for numeric observation, not a proof of full Level-18 decryption or secure CKKS equivalence. Capacity status remains `CAPACITY_UNPROVEN_EXPERIMENTAL`.

The durable attempt journal, complete Fast and partial Standard logs, ciphertext checkpoint, vector files, metadata, and human-readable Fast vector are stored alongside this summary. The original Secondary dirty worktree was not modified.
