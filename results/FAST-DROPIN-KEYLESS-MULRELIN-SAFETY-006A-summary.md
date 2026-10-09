# FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A

Classification: `FAST_DROPIN_KEYLESS_MULRELIN_SAFETY_COMPLETE_PENDING_WEB_REVIEW`

Handoff: `READY_FOR_WEB_REVIEW`

## Repair

Secondary `xuejin-lu/lattigo` commit `93ab7ecc5a864fd9bbacea55f0af730941552691` (`fast-qprefix`) makes Fast-marked public CKKS ciphertext/ciphertext `MulRelin` and `MulRelinNew` fail closed. Nonzero active `c1`, unsupported degree, malformed/missing active-Q backing, and unsupported representations return errors before output mutation; they cannot fall through to native full-Q KeySwitch/GadgetProduct. `MulRelinNew` also returns an error for nil ciphertext operands instead of panicking.

The supported fully materialized degree-one zero-secret product remains unchanged: it uses the existing CKKS product kernel with relinearization disabled, preserves the keyless `(a0*b0, 0)` result and metadata behavior, and works with either a nil evaluation-key set or an available Standard-layout relin key. Ordinary plaintext/scalar overloads and Standard source were not changed.

Tests generate an actual Standard-layout relin key through `ckks.NewKeyGenerator`. With that key present, nonzero-`c1` and higher-degree inputs are rejected, input/output serialization (including metadata) remains byte-identical, and an instrumented key set records zero relin-key lookups on rejection. Nil-key rejection and the accepted keyless product are also covered. Existing CKKS tests that had expected native ciphertext fallback or degree-zero ciphertext coercion now assert the frozen Fast fail-closed contract.

## Validation

Passed on the final Secondary commit:

```text
git diff --check
go test ./schemes/ckks ./schemes/ckks/fast ./core/rlwe -count=1
```

All three packages passed. `go test ./...` was not required by this bounded task and was not run. No Bootstrap, benchmark, or Standard computation was run.

## Single Fast numerical regression

The unchanged frontend `tools/fast-dropin-ckks-primitive-api-audit-005/main.go` was verified at SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`. One Fast dependency run completed all eight public-operation checkpoints; `BootstrapCalls=0`, and the temporary raw evidence is `/private/tmp/FAST-DROPIN-KEYLESS-MULRELIN-SAFETY-006A-fast.json`.

| Checkpoint | Level / log2(Scale) | Complex RMSE | Max complex error |
|---|---:|---:|---:|
| EncryptNew, Add operand A | 3 / 45 | `4.3074465359754706e-14` | `8.564136368436897e-14` |
| EncryptNew, Add operand B | 3 / 45 | `4.8063139929870046e-14` | `9.499856654154729e-14` |
| Add | 3 / 45 | `6.366054749149383e-14` | `9.928313431575474e-14` |
| EncryptNew, MulRelin operand A | 3 / 45 | `4.3074465359754706e-14` | `8.564136368436897e-14` |
| EncryptNew, MulRelin operand B | 3 / 45 | `4.8063139929870046e-14` | `9.499856654154729e-14` |
| MulRelin | 3 / 90 | `5.320587702687035e-15` | `1.2117592599745943e-14` |
| Rescale | 2 / `50.999999441053866` | `5.508744899042129e-15` | `1.2490249597131585e-14` |
| Rotate | 2 / `50.999999441053866` | `5.508361443991879e-15` | `1.2490141058066392e-14` |

All outputs had zero `c1`; all stages passed. The numerical run used Primary `71229074dc1b62ad9287ea0435dbc55812e35a6d` clean and Secondary base commit `6ce15cf8b949c8a89f610bcca5ab506dd3f560f8` with the fail-closed implementation working tree dirty. The subsequent self-review change added only the Fast `MulRelinNew` nil-input pre-allocation guard and its test; the multiplication/arithmetic path was unchanged, and the final source passed the focused suite above. The pinned Standard result from the previous review was not recomputed.

## Final state

- Secondary `origin/fast-qprefix`: `93ab7ecc5a864fd9bbacea55f0af730941552691`; pushed normally and clean.
- This summary is the only Primary change for this task.
