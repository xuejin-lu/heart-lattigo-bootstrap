# FAST-DROPIN-EVALUATOR-REUSE-FEASIBILITY-008 — feasibility summary

**Classification:** `REUSE_REQUIRES_REDESIGN`
**Handoff:** `NEEDS_WEB_REVIEW` / architecture decision required
**Primary source:** `heart-lattigo-bootstrap` `main@57c988f6fe69792edb8d20d88afcd50dd324cac8` (clean at investigation start)
**Fast source:** `xuejin-lu/lattigo` `fast-qprefix@93ab7ecc5a864fd9bbacea55f0af730941552691` (equals `origin/fast-qprefix`; clean)
**Genuine Standard reference:** `5dbffbdea05394de2ca3a432ed5318aa832e3f40` (pinned by the task)
**Activity:** source-only feasibility audit. No source implementation, tests, Bootstrap, benchmark, or numerical experiment was run.

## End goal and finding

The end goal remains identical frontend CKKS source, public calls, and parameter values, with only the Lattigo dependency changed. Fast zero-secret semantics must not silently invoke key switching or consume unmaterialized/stale Q rows. This report does not claim numerical correctness or that a shared Fast kernel has been invoked through the public evaluator.

The existing Fast Rotate/Automorphism kernel is a real extraction candidate: `fast/automorphism.go:33-83` implements the coefficient/NTT permutation using `ring` types, and its evaluator-owned wrapper adds a per-evaluator NTT index cache and scratch (`fast/automorphism.go:85-142`). A neutral internal package can therefore reuse the existing algorithm without importing either `schemes/ckks` or `schemes/ckks/fast`.

That kernel-level result is not sufficient to integrate public `ckks.Evaluator.Rotate`. The public evaluator accepts full-active-Q ciphertexts and its current Automorphism key-switches then transforms all active rows; Fast's explicit API transforms only its selected Q-prefix rows, and `fast.NewCiphertext` can contain nil rows above that prefix. The Fast row selector is narrower still for the current legacy producer. No inspected source contract makes those untouched rows valid output for a later full-RNS public operation. Thus the code-sharing seam is feasible, while the end-to-end API/representation contract requires redesign and Web review before implementation.

## Source-level call and dependency map

```text
ordinary public frontend
  ckks.NewEvaluator(params, evk) -> *ckks.Evaluator
  ckks.Evaluator.Rotate(ct, k, out)
    -> embedded rlwe.Evaluator.Automorphism(ct, params.GaloisElement(k), out)
       -> CheckAndGetGaloisKey -> GadgetProduct / key switch
       -> ring automorphism across the active ring level

explicit Fast frontend
  fast.NewEvaluator(params) -> *fast.Evaluator
  fast.Evaluator.Rotate(ct, out, k)
    -> Fast.Evaluator.Automorphism(ct, out, GaloisElementForRotation(k))
       -> fastAutomorphism (legacy row policy)
       -> fastAutomorphismRows (row validation, NTT index cache, scratch)
       -> FastAutomorphismRows / ring.AutomorphismNTTIndex
```

The public method argument order is `Rotate(ct, k, out)`; the explicit Fast order is `Rotate(ct, out, k)`. Both map rotation through the CKKS parameter Galois-element mapping (`Parameters.GaloisElementForRotation(k)` delegates to `Parameters.GaloisElement(k)`). The public `rlwe.Evaluator.Automorphism` rejects non-degree-one inputs, obtains a Galois key, performs `GadgetProduct`, then applies the ring automorphism to both components (`core/rlwe/evaluator_automorphism.go:13-53`). Fast's `Automorphism` handles both components without key lookup, copies input metadata, permits input/output aliasing through the low-level temporary, and has no zero-`c1` guard (`fast/evaluator.go:243-284`, `fast/automorphism.go:33-83`). A zero-secret public adapter must explicitly validate the zero-`c1` contract; the primitive itself does not establish it.

The relevant concrete signatures are:

```go
func NewEvaluator(parameters Parameters, evk rlwe.EvaluationKeySet) *ckks.Evaluator
func (eval Evaluator) Rotate(op0 *rlwe.Ciphertext, k int, opOut *rlwe.Ciphertext) error
func (eval Evaluator) Automorphism(ctIn *rlwe.Ciphertext, galEl uint64, opOut *rlwe.Ciphertext) error

func NewEvaluator(params ckks.Parameters) *fast.Evaluator
func (eval *fast.Evaluator) Rotate(ctIn, ctOut *rlwe.Ciphertext, k int) error
func (eval *fast.Evaluator) Automorphism(ctIn, ctOut *rlwe.Ciphertext, galEl uint64) error
func (eval *fast.Evaluator) AutomorphismQPrefixRows(ctIn, ctOut *rlwe.Ciphertext, galEl uint64, rows int) error
```

`schemes/ckks/evaluator.go` imports `core/rlwe`, `ring`, `ring/ringqp`, and utilities, and `ckks.Evaluator` embeds `*rlwe.Evaluator`; `NewEvaluator` keeps returning the concrete `*ckks.Evaluator` (`schemes/ckks/evaluator.go:1-30`). The Fast evaluator imports parent `schemes/ckks` for `ckks.Parameters`, as well as `core/rlwe` and `ring` (`schemes/ckks/fast/evaluator.go:1-12`). Therefore a direct `schemes/ckks -> schemes/ckks/fast` import creates a Go import cycle. Replacing the public constructor's return type also breaks its established API. The explicit Fast evaluator deliberately does not implement the generic scheme evaluator interface because that interface entails full-Q/QP and key-switch operations (`schemes/ckks/fast/evaluator_ntt.go:10-20`). Do not widen it as a dispatch shortcut.

### Reuse seam and concrete candidate

`FastAutomorphismRows` itself takes `*ring.Ring`, `ring.Poly` input/output, `galEl`, domain flag, and explicit row count; its arithmetic imports only `ring` (`fast/automorphism.go:33-83`). The cached NTT path also needs only `ring.AutomorphismNTTIndex`, a `map[uint64][]uint64`, and per-row temporary storage (`fast/automorphism.go:99-142`). The Fast wrapper's current row-validation helper is in the Fast package (`fast/prefix_kernels.go:9-40`), so an extraction must move or re-home the generic validation and prefix-width constant too; it cannot import that helper from the child package.

A plausible, **unimplemented** shape is `schemes/ckks/internal/fastcore` with ring-only types and an explicit-row entrypoint, for example:

```go
type AutomorphismWorkspace struct { /* NTT index cache and row scratch */ }
func (w *AutomorphismWorkspace) ApplyRows(
    ringQ *ring.Ring, in, out ring.Poly, galEl uint64, isNTT bool, rows int,
) error
```

The package must import neither `schemes/ckks` nor `schemes/ckks/fast`; Go's `internal` visibility permits both packages beneath `schemes/ckks` to import it. The explicit Fast evaluator would retain its public wrappers and legacy row-selection behavior while delegating the shared kernel. The public CKKS evaluator could own a workspace without changing `NewEvaluator`'s public return type. These are dependency facts, not authorization to create the package or choose the public row/representation contract.

Other operations confirm that the API boundary is broader than a method-name match. Public Rescale is `func (eval Evaluator) Rescale(op0, opOut *rlwe.Ciphertext) error`; Fast exposes `func (eval *Evaluator) Rescale(op0, opOut *rlwe.Ciphertext) error` and `func (eval *Evaluator) RescaleQPrefixRows(op0 *rlwe.Ciphertext, rows int, opOut *rlwe.Ciphertext) error` (`fast/rescale.go:143-188`), while the public path is full-RNS. Fast exposes `func (eval *Evaluator) MulRelinElementQPrefixRows(op0 *rlwe.Ciphertext, op1 *rlwe.Element[ring.Poly], rows int, opOut *rlwe.Ciphertext) error`; 006A public zero-secret MulRelin instead validates fully materialized active-Q rows and uses its own full-Q kernel (`schemes/ckks/evaluator_fast_zero_secret.go:21-26`, `schemes/ckks/evaluator.go:750-778`). Neither is proof that the public path reuses the existing Q-prefix kernel. Preserve that 006A fail-closed behavior; do not fall through to native key switching for unsupported Fast inputs.

## Q-prefix versus full-Q representation contract

`QPrefixWidth(L)=min(L+1,4)` (`fast/qprefix.go:9-24`). Ordinary `ckks.NewCiphertext` uses RLWE full backing through the logical level (`core/rlwe/ciphertext.go:15-20`); `fast.NewCiphertext` preserves the logical level but allocates backing only for rows within the capped prefix, leaving higher rows nil (`fast/ciphertext.go:8-24`). The existing `validatePrefixRows` rejects requests beyond `min(L+1,4)` and checks physical backing, not semantic provenance (`fast/prefix_kernels.go:9-40`). `Resize` can allocate the whole prefix even where a narrower producer did not write all those rows; allocation alone does not make a row authoritative.

For the legacy Fast automorphism producer, the default is q0/q1. It uses q0/q1/q2 only for the identified q012 modulus shape (`bits.Len64(q0)==56`, q1 at most 39 bits, q2 at most 40 bits); the canonical q0=55 profile remains two-row (`fast/q012.go:10-34,49-64`). The explicit `AutomorphismQPrefixRows` is a separate API that accepts a caller-selected prefix, including one row at Level 0; ordinary `Automorphism` requires Level >= 1 and q0/q1 (`fast/evaluator.go:243-320`).

| Logical Level | Q-prefix width | Public CKKS backing | Fast compact backing | Legacy automorphism rows | Consequence for a non-identity Fast Rotate result |
|---|---:|---|---|---:|---|
| 0 | 1 | q0 | q0 | 1 by row policy, but ordinary Fast `Rotate` rejects Level 0; explicit rows API can process 1 | Only the explicit one-row API is available; a public adapter needs an explicit Level-0 contract. |
| 1 | 2 | q0–q1 | q0–q1 | 2 | All active rows fit the legacy path, assuming both rows are authoritative and input/output metadata/domain checks pass. |
| 2 | 3 | q0–q2 | q0–q2 | 2 for canonical q0=55; 3 for q012 profile | Canonical output q2 is not rotated; it is unsafe for a full-Q consumer. q012 is complete only for this level. |
| 3 | 4 | q0–q3 | q0–q3 | 2 for canonical profile; 3 for q012 profile | Canonical output q2–q3, or q012 output q3, is not rotated and cannot be treated as full-Q output. |
| >=4 | 4 active prefix; full logical Q has L+1 rows | q0–qL | q0–q3; q4+ nil | 2 canonical or 3 q012 | Even q012 leaves q3 and all higher active full-Q rows untouched; compact output cannot enter native/full-RNS APIs. |

The table describes row coverage, not a guarantee that every source row is mathematically authoritative. At Levels where producer rows are fewer than prefix width, higher allocated prefix rows may be stale. A public/full-Q boundary must fail closed unless every row a downstream full-RNS operation can read has valid output provenance. It must not infer authority from slice length, allocated backing, or `Level` alone. Likewise, a compact ciphertext can be consumed only by operations that explicitly honor its row contract; sending it to native ring/full-RNS code can read nil or stale rows.

The explicit Fast automorphism requires degree one, Standard ring, matching logical Levels, dimensions, NTT flag, and Montgomery representation; it accepts coefficient and NTT domains and copies metadata/Scale (`fast/evaluator.go:243-284`). The rows API additionally checks requested backing for both components. Low-level aliasing is handled with temporary storage. These checks do not validate `c1==0`; that belongs in a zero-secret public adapter. Current general Fast arithmetic elsewhere commonly requires NTT, Level >=1, and matching Montgomery flags (`fast/evaluator_ntt.go:592-623`). A new adapter must define, not assume, its allowed domain and Montgomery states.

## Ranked architecture options

1. **Conditional shared-kernel extraction (`schemes/ckks/internal/fastcore`) — best reuse seam, not yet an end-to-end answer.** Move the actual Fast row kernel, its reusable workspace, and neutral validation into the internal package. Preserve `fast.Evaluator` wrappers and make both explicit and public routes call the same instrumentable function. Before coding, Web must choose whether the public path is compact-Q with an explicitly bounded dispatch surface, or full-Q with a proof that all active rows are produced. Current four-row validation and the existing producer policies cannot be silently widened to solve this. No fallback to full-RNS key switching is permitted.
2. **Neutral package-layering redesign — viable but larger.** Remove `fast`'s dependency on parent `ckks` through a neutral parameter/provider boundary, then let public `ckks` hold an internal Fast adapter while preserving its concrete public evaluator type. Current `fast.Evaluator` stores `ckks.Parameters` and uses CKKS-level parameter methods, so this requires a source-backed inventory of every such dependency plus lifecycle, storage, and dispatch changes. It is broader than extracting the ring-only kernel and does not by itself resolve stale/full-Q rows.
3. **Native ring primitive reuse — safe only under a separately chosen full-row contract; not reuse of the existing Fast kernel.** Public RLWE already calls `ring.AutomorphismNTTWithIndex` or `ring.Automorphism` after key switching (`core/rlwe/evaluator_automorphism.go:35-50`); ring's NTT primitive processes every row through the Ring level and documents non-in-place output (`ring/automorphism.go:47-60`). Calling that native primitive without the key-switch stage could cover full-Q rows only when all active rows are valid and the output is full-backed. It is a different algorithm/path from the Fast explicit-row kernel, does not handle compact rows safely by itself, and would not satisfy the same-kernel reuse proof.

Directly importing the Fast child package from public CKKS, changing `NewEvaluator` to return `*fast.Evaluator`, init/blank-import registration, and copying another Rotate implementation are rejected: they respectively create an import cycle, break the concrete API, rely on hidden linkage, or duplicate the already-existing kernel without resolving row authority.

## First implementation task after Web decision

Do not implement until Web selects one representation contract. The first bounded slice should then be **Rotate/Automorphism only**:

1. Keep the public `ckks.NewEvaluator` signature and the ordinary Standard lifecycle unchanged; dispatch only at the existing Fast zero-secret capability boundary, without a frontend selector.
2. Extract the current Fast automorphism kernel/workspace once into an import-neutral package; retain the explicit Fast wrapper and verify both public and explicit entrypoints call the same function with instrumentation.
3. Define accepted input provenance and output storage, including logical Level/Scale, exact authoritative rows, NTT/Montgomery state, Standard-ring restriction, degree one, zero-`c1`, alias behavior, and fail-closed errors. Never expose untouched rows as valid full-Q output.
4. Prove no Galois-key lookup, `GadgetProduct`, native KeySwitch fallback, or dormant/stale row read occurs on the Fast zero-secret path. Leave ordinary Standard behavior unchanged.

The architecture decision is specifically: **does public Fast Rotate preserve a compact Q-prefix representation and require a complete bounded Fast dispatch boundary, or must its public output remain fully materialized full-Q (which the current four-row Fast kernel cannot guarantee at arbitrary Level)?** If neither contract can be made explicit and source-valid, stop before implementation rather than widening the kernel or adding a conversion workaround.

## Future proof plan (not executed)

- Use unchanged frontend source and profile; compare against the pinned genuine Standard native key-generation/encrypt/operation/decrypt lifecycle and a cleartext oracle. Record Level, Scale, slots, and numerical metrics.
- Instrument the shared kernel to prove it is invoked from both `ckks.Evaluator.Rotate` in the Fast zero-secret build and existing `fast.Evaluator.Rotate`; separately instrument/assert no Galois-key lookup or key-switch helper on the Fast route.
- Positive cases: supported authoritative full/compact rows, zero `c1`, coefficient and NTT where declared, supported Montgomery state, alias/non-alias, unchanged Level/Scale and metadata.
- Negative cases: nonzero `c1`, missing/stale required rows, unsupported ring/degree/domain/Level, and any compact ciphertext reaching an operation that assumes full-Q. Assert deterministic fail-closed behavior and no fallback.
- Build both consumers of the internal package to establish legal import direction/no cycle. Test ordinary Standard regression and public Bootstrap only as a later gate after the first slice is accepted. Do not benchmark until correctness and dispatch are both proven.

## Validation and scope

Required source provenance was confirmed: Primary `main` was synchronized at `57c988f6fe69792edb8d20d88afcd50dd324cac8`; Secondary `fast-qprefix` and `origin/fast-qprefix` both equal `93ab7ecc5a864fd9bbacea55f0af730941552691`, with a clean worktree. This audit did not execute tests, benchmarks, Bootstrap, or numerical experiments, and changed no Secondary files. The only authorized repository change for this task is this report. `FAST-STANDARD-PERF-REBASELINE-003` remains blocked.

**Decision:** shared ring-kernel extraction is technically feasible; public drop-in integration is **`REUSE_REQUIRES_REDESIGN`** because the import/API and full-Q/compact-Q contracts are unresolved. Stop here for **`NEEDS_WEB_REVIEW`**; do not start suspended task 007 or any implementation.
