# BTS Fast No-Trace Validation 003 — attempt 01

**Classification:** `FAST_BTS_BOOTSTRAP_PANIC_BEFORE_OUTPUT`\
**Mode:** No Trace; original BTS LogN16 / SlotLimbs=15 / `BTSModePlain` / amp=1 / seed=0.\
**Outcome:** the single authorized `bts.Bootstrap` invocation was entered and panicked during C2S. The attempt budget is consumed; no retry was made or is permitted.

## Provenance

- Primary source tree at build: `8ee8f7392f17a4d03b4c218f8a2c5b2dc85e38da`; BTS source manifest is recorded in the attempt journal and matches the unchanged `external/BTS/` project.
- Fast backend: Secondary `fast-qprefix` commit `049ffa782729d049e08f973b5852ee1e832a55f3`, pushed and verified against `origin/fast-qprefix`.
- Go/runtime: `go1.26.4 darwin/arm64`, Darwin 25.6.0 arm64.
- Exact build/run commands, alternate modfile replacement, module provenance, and binary SHA-256 are recorded in the journal.
- Historical Standard baseline was not rerun: 139.653 s; max absolute error `1.708e-3`; mean absolute error `1.172e-4`.

## Result

| Measurement | Fast result |
|---|---:|
| `NewBTS()` initialization | 45.048 s, successful |
| Input ciphertext | Level 3, Scale `2^32` |
| `bts.Bootstrap()` | Entered; panic during C2S; no completed latency |
| Output Level / Scale | Unavailable |
| Max / mean absolute error, max-error slot | Unavailable; no output/decryption |

Panic: `slice bounds out of range [:32768] with capacity 0` at `ring.Ring.MulDoubleRNSScalar` (`ring/operations.go:253`), reached from `MyC2S_DHManual` (`external/BTS/myc2s.go:317`) through CKKS `Mul`, within `(*BTS).Bootstrap` (`external/BTS/bts.go:262`). This identifies the failing call path; the run provides no successful output or numerical evidence and makes no performance or correctness claim.

Complete stdout/stderr: [attempt-01 raw log](BTS-FAST-NOTRACE-VALIDATION-003-attempt-01.log). Durable one-call accounting: [attempt journal](BTS-FAST-NOTRACE-VALIDATION-003-journal.md). Task 002 records were not changed.

## Pre-call gates and implementation scope

The focused compatibility/zero-secret automorphism tests, `go test ./schemes/ckks/fast -count=1`, bootstrapping interface/dispatch tests (without Bootstrap), affected-package `go vet`, and `git diff --check` passed on the final pushed code. A new Fast Bridge panic found during regression review was changed to fail closed with an error before output mutation; Standard behavior is unchanged. Full `go test ./schemes/ckks -count=1` remains non-green, with the same failures on pristine baseline `2cb8bc799ccf5ec2e8e011b84fb9e0dabd77886f`: `TestCKKS/Evaluator/RescaleTo/Single` Standard and ConjugateInvariant cases panic in decryptor `ring.addvec` on dormant rows; `TestCKKS/Evaluator/RescaleTo/Many` retains its baseline precision failures at q385 and capacity rejection at q680 (`Q-prefix product through q3 exceeds 192 bits`); and `TestCKKS/Evaluator/Mul/Ct/Ct/Degree0` has the baseline expected-error-text mismatch. No unrelated baseline failures were repaired.

The BTS path reuses Fast BSGS/rotation/Rescale/Mul kernels and retains required Galois automorphisms. Zero-operand key-switch arithmetic is skipped only after validating zero components. This is **not** Q0123-only execution: full-Q/QP diagonal accumulation, QP automorphisms, ModDown conversions, and full-Q basis/NTT operations remain on the BTS path. No Standard fallback or BTS source change was introduced.
