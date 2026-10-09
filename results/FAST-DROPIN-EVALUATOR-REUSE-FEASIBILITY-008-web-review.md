# Independent Web Review — FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008

**Disposition: ACCEPT the source investigation.** Codex classification `REUSE_REQUIRES_REDESIGN` is justified. The architecture choice below is Web-approved for the next bounded implementation, not a claim that it is already coded or generally proven.

## Grounded evidence

- Primary 008 report commit `15c3d532da721ef7a45d2a92635adcf5c2b4fda2`; Fast Secondary remains `93ab7ecc5a864fd9bbacea55f0af730941552691`. Genuine unmodified Standard pinned `5dbffbdea05394de2ca3a432ed5318aa832e3f40`.
- `schemes/ckks/evaluator.go` public `NewEvaluator` returns concrete `*ckks.Evaluator`, embedding native `*rlwe.Evaluator`; public Rotate uses native Automorphism/GaloisKey/GadgetProduct.
- `schemes/ckks/fast/evaluator.go` and `fast/automorphism.go` already implement a Fast rotation/automorphism with no Galois-key lookup; direct parent-to-child import cycles because child Fast imports parent CKKS.
- `fast/automorphism.go` contains a reusable ring-level NTT permutation/coefficient automorphism, including evaluator-local index cache/scratch; its existing prefix validator `fast/prefix_kernels.go` imposes `MaxQPrefixWidth=4`. The explicit Fast default operation uses an even narrower legacy producer selector: usually two rows, or three rows for a particular 56-bit q0 profile (`fast/q012.go`).
- `fast/ciphertext.go` can hold compact Q backing with nil dormant high rows. In contrast `ckks.NewCiphertext` public source creates full active Q backing. Logical Level alone and physically allocated backing **do not prove authoritative residues**. `FAST_QPREFIX_SPEC.md` explicitly forbids reading dormant/stale Q limbs.

Codex did not execute experiments; Web independently inspected the relevant current sources, not new numerical measurements.

## Architecture decision: full-Q public baseline, explicit compact-Q internal authority

Freeze two distinct **representation contracts**, not two duplicated arithmetic algorithms:

**P0: ordinary public CKKS Fast zero-secret numeric lifecycle (FIRST INTEGRATION TARGET).** A supported, provenance-known public Fast ciphertext produced by secret-key `EncryptNew` and the previously checked public full-Q operations has degree one, effective `s=0`, active `c1=0`, and all `level+1` logical Q rows authoritative. A successful public `ckks.Evaluator.Rotate/RotateNew` must preserve fully authoritative active rows, Level, Scale, dimensions and NTT representation, with no Galois key or native KeySwitch. The shared core is invoked with `rows=level+1` and full-backed source/output. This may be >4 rows when Level>=4. **The shared low-level permutation must therefore have validation/workspace capable of this explicit row count**, while the existing Fast wrappers retain their original Q-prefix <=4/legacy row policy. This is reuse of the moved ring algorithm, not permission for native full-Q RLWE KeySwitch or a second Rotate implementation. Full-Q processing at high Levels is not yet a Q-prefix computational speedup.

**C0: explicit compact-Q Fast evaluator contract (PRESERVED, not silently promoted to public).** The original `fast.Evaluator` continues to select only its proven producer-authoritative row count, e.g. existing q0/q1 or q012 path, and retains its own existing callers, scratch policy and Q-prefix capacity/bound rules. It delegates to the same moved low-level kernel for those rows. On compact output, high logical Q rows are not authoritative and **cannot** be fed into any native/public full-Q consumer. Existing `2B<S_Q` capacity and materialization rules remain in force.

**No implicit P0↔C0 conversion.** An ordinary `ckks.Evaluator.Rotate` on a compact ciphertext must reject missing authoritative backing before mutation, not widen it and not invoke native Standard full-RNS. Conversely C0 should never claim to have updated full-Q rows. A physically present row can still be stale: do not claim that mere slice-length checks solve provenance in every possible arbitrary ciphertext. Restrict initial acceptance to source-traced public full-Q input chains; report how future provenance detection could be enforced before claiming arbitrary mixed-operation transparency.

This architectural decision is **transitional**: the long-term requirement remains unchanged frontend with faster Fast evaluation. A later comprehensive compact-Q drop-in path requires all ordinary CKKS operations in the workload to use correct compact authority, or formally proven explicit boundaries, with no stale native fallback. Do not claim that the P0 path already achieves Q-prefix acceleration or complete CNN compatibility.

## Package direction

Preferred seam `schemes/ckks/internal/fastcore` is acceptable because packages `schemes/ckks` and `schemes/ckks/fast` can both import it. Shared core should depend on `ring` and simple import-neutral types only, **not** parent CKKS or explicit Fast package. Move the *existing* automorphism row permutation/index reuse into that seam, retaining original public Fast wrappers and ABI; adapt parameter and workspace ownership internally while preserving algorithm identity. Do not invent a new standalone rotation algorithm copied into `ckks`.

Rejected: child import into parent, constructor-return-type changes, blank-import registration, blanket Fast `schemes.Evaluator` interface widening, and duplicate public Fast Rotate kernel.

## Acceptance for next implementation

One **bounded** shared-core extraction + one public zero-secret Rotate/RotateNew adapter, under unchanged public signatures, only in Fast Secondary. Same Fast underlying kernel must be called by explicit and public entrypoints, demonstrated by source and instrumentation/testing; no new frontend requirements. Full active Q rows must be maintained for P0 even at Level>=4; explicit Fast wrapper must retain original Q-prefix policy, including no dormant read. Proof of zero Galois key lookups with real available Standard-layout keys, of no native GadgetProduct/KeySwitch, and of transactional fail-closed invalid input. No mandatory Bootstrap/benchmark. Test positive two or more levels including Level>=4 with full Q, negative compact Level>=4, aliasing, +1/-1/0 slots, metadata and matched numerical oracles. If actual source makes full-Q high-Level extraction unsafe, **STOP rather than reducing scope or inventing conversion**.

This authorizes a carefully bounded implementation task `FAST-DROPIN-SHARED-AUTOMORPHISM-BRIDGE-009`. It does not yet authorize broad Add/MulRelin/Rescale refactoring, changing Q-prefix policy or wholesale rewrites.
