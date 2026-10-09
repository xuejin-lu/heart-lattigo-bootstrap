# FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009

Classification: `FAST_SHARED_AUTOMORPHISM_BRIDGE_COMPLETE_PENDING_WEB_REVIEW`
Handoff: `READY_FOR_WEB_REVIEW`

## Result

Extracted the existing Fast CKKS coefficient/NTT automorphism permutation into `schemes/ckks/internal/fastcore`. Both the explicit `fast.Evaluator` C0 path and the ordinary public `ckks.Evaluator.Rotate` P0 path delegate to this same core. The internal package imports only `ring`, so it introduces no import cycle. No Standard repository, frontend, parameter, Bootstrap, or unrelated production arithmetic was changed.

P0 validates degree-one/full-active-Q storage, matching representation flags, metadata, canonical c0 residues, and zero c1 before invoking the core. It passes `rows=Level+1`, preserves level/scale/metadata and c1, handles in-place and separate output, and does not retrieve a Galois key or call native key switching. The coefficient-domain result canonicalizes the ring primitive's equivalent `q` representation of negative zero back to zero, keeping zero-secret c1 exactly zero.

C0 keeps the existing Fast wrapper row policies and Q-prefix bounds. The legacy evaluator path still selects q0/q1 in the exercised case; the explicit three-row path transforms exactly q0..q2. The shared core validates every requested pair before any writes and does not read or write rows outside the requested prefix.

## Correctness evidence

- Shared-core instrumentation verifies the ordinary public entrypoint calls the core once for c0+c1 with `rows=Level+1` (including Level 5 / six active rows), and the explicit Fast entrypoints call the same core with their existing row counts.
- Public Fast `EncryptNew` → `Rotate` / `RotateNew` tests cover k=0,+1,-1 at Levels 1 and 5, custom `LogSlots=4`, separate and aliased outputs, and a full-slot coefficient-domain case. Results are compared to independent plaintext rotation oracles; Scale, Level, degree, dimensions/metadata, domain flags, full active-row backing, and zero c1 are checked.
- Keyless behavior is tested with nil evaluation keys and with a tracking keyset containing a real Standard-layout Galois key. Both successful rotation and rejection of nonzero c1 perform zero `GetGaloisKey` calls.
- Transactional rejection cases cover nonzero c1, compact Level 5 storage, a missing active input row, missing output row, noncanonical c0 residue, mismatched domain, invalid Montgomery coefficient representation, nil input/output metadata, degree two, nil input, and conjugate-invariant ring.
- Core-level coefficient and NTT results match the ring oracle at six active rows; aliasing and validation-before-write behavior are tested.

## Canonical unchanged-frontend regression

The fixed frontend was unchanged: SHA-256 `62e1a069d2ce3c0106e34594cb799388d9d6fbebbb72dd55b0ef5989f02d224e`. The Fast run used Primary `d0a3d1bfb1adcea96450c32364349e3961d490e5` (clean), Go `go1.26.4`, `darwin/arm64`, and the existing LogN13 profile (Q bits `[55,39,40,39]`, P bits `[60]`, input Level 3, Scale `2^45`, 16 slots, Hamming weight 192, rotate-left 1). Profile SHA-256 `8a2b9ee73d1beee7759171b0f8ea08a554c07392b666baed7f0212a3664fa91a`; deterministic input SHA-256 `b46752705bcda9226eefbf6f19aa5d75e03e74226f029f7e6db5694050d257a4`.

The run was performed once against Secondary `fast-qprefix` at base `93ab7ecc5a864fd9bbacea55f0af730941552691`, with `backend_dirty=true`. The production source in that working tree is the source committed as `00ac70ba136d190fa31bbb26c2f51d003a221634`; subsequent changes before commit were confined to tests. Focused and required package tests were rerun after those test-only edits. The raw run JSON remains local-only at `/private/tmp/FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009-fast.json`.

The runner completed all public `ckks.NewEvaluator` operations. It reports `fast_evaluator_constructed=false` because the explicit `fast.Evaluator` API was not constructed; the public Fast zero-secret Rotate path was the target of this bridge. All decoded results were finite, output c1 remained zero, and Bootstrap calls were 0.

The previously accepted genuine Standard result is reused: pinned Standard Lattigo `5dbffbdea05394de2ca3a432ed5318aa832e3f40`, clean, with the same frontend, profile, and input hashes. Standard run Primary SHA was `3a42531068129dadce7ba1b5bcdcc52123764426`; the frontend SHA and workload/profile hashes match the current Fast run.

| Checkpoint | Level / log2(Scale) | Standard RMSE / max complex error | Fast RMSE / max complex error |
|---|---:|---:|---:|
| Add | 3 / 45 | `7.130044578116784e-13` / `1.1382388697121386e-12` | `6.366054749149383e-14` / `9.928313431575474e-14` |
| MulRelin | 3 / 90 | `5.643830896935627e-14` / `9.705823679407266e-14` | `5.320587702687035e-15` / `1.2117592599745943e-14` |
| Rescale | 2 / `50.999999441053866` | `5.909656072886534e-14` / `9.542571743484449e-14` | `5.508744899042129e-15` / `1.2490249597131585e-14` |
| Rotate left 1 | 2 / `50.999999441053866` | `5.81639608533112e-14` / `9.16706303498959e-14` | `5.508361443991879e-15` / `1.2490141058066392e-14` |

These are per-backend decoded complex errors against the same cleartext oracle; they are not presented as a direct ciphertext-coefficient or cross-run randomness comparison. The Fast run used ordinary Standard-layout evaluation keys from `ckks.NewKeyGenerator`; the Rotate bridge itself does not access a Galois key.

## Validation and limitations

Passing focused tests:

```text
GOWORK=off GOCACHE=/private/tmp/fast-shared-automorphism-bridge-009-gocache go test ./schemes/ckks/internal/fastcore ./schemes/ckks ./schemes/ckks/fast -run 'TestAutomorphismWorkspace|TestFastAutomorphismWrappersUseSharedCoreAndKeepRowPolicies|TestPublicFastZeroSecretRotate' -count=1
```

Passing required Secondary suite:

```text
GOWORK=off GOCACHE=/private/tmp/fast-shared-automorphism-bridge-009-gocache go test ./schemes/ckks ./schemes/ckks/fast ./core/rlwe -count=1
git diff --check
```

P0 is a zero-secret numerical simulation, not secure HE: nonzero c1 is rejected and no key switching occurs. Runtime validation can establish full row backing and zero c1, but cannot prove arbitrary ciphertext-row freshness or provenance. Callers must keep this public path bounded to source-proven full-Q Fast inputs; this change adds no implicit conversion between P0 and C0. Full-active-Q processing is not a compact-Q speedup.

Secondary `fast-qprefix` commit `00ac70ba136d190fa31bbb26c2f51d003a221634` was normally pushed and verified at `origin/fast-qprefix`; Secondary worktree is clean. Primary was clean at `d0a3d1bfb1adcea96450c32364349e3961d490e5` before adding this summary. This task does not run Bootstrap, benchmarks, or other experiments.
